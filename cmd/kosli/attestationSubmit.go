package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/kosli-dev/cli/internal/kosli"
	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/utils"
	cp "github.com/otiai10/copy"
)

// attestationSubmitter records Attestations on Trails.
type attestationSubmitter struct {
	kosli  *kosli.Kosli
	org    string
	logger *log.Logger
}

func newAttestationSubmitter() *attestationSubmitter {
	return &attestationSubmitter{
		kosli:  kosliAPI(),
		org:    global.Org,
		logger: logger,
	}
}

// attestationSubmission is one Attestation to record on a Trail.
type attestationSubmission struct {
	flow  string
	trail string
	// slug is the attestation type's segment in the attestations URL.
	slug string
	// label names the attestation type in the success message.
	label       string
	payload     attestationPayload
	attachments []string
	// evidence is an attachment already read into memory; it cannot be
	// combined with attachments.
	evidence *requests.FileBytes
	// assertFailures fail the command after the Attestation is recorded, so
	// the evidence is kept even when --assert rejects it.
	assertFailures []error
}

// attestationPayload is the body of any Attestation type; every one embeds
// CommonAttestationPayload, which provides the name.
type attestationPayload interface {
	attestationName() string
	setAttachmentUploadID(id string)
}

// submit packages the attachments once, then records the same Attestation on
// every host concurrently, each with its own upload of them.
func (s *attestationSubmitter) submit(a attestationSubmission) error {
	if a.evidence != nil && len(a.attachments) > 0 {
		return fmt.Errorf("an attestation carries either in-memory evidence or attachments, not both")
	}
	evidencePath := ""
	if len(a.attachments) > 0 {
		path, cleanupNeeded, err := getPathOfEvidenceFileToUpload(a.attachments, s.logger)
		if err != nil {
			return err
		}
		if cleanupNeeded {
			defer func() {
				if err := os.Remove(path); err != nil {
					s.logger.Warn("failed to remove evidence file: %v", err)
				}
			}()
		}
		evidencePath = path
	}

	err := s.kosli.Send(func(h kosli.Host) error {
		err := s.submitTo(h, a, evidencePath)
		if err == nil && !h.DryRun {
			h.Info("%s attestation '%s' is reported to trail: %s", a.label, a.payload.attestationName(), a.trail)
		}
		if !h.DryRun {
			for _, failure := range a.assertFailures {
				if err != nil {
					err = fmt.Errorf("%s\nError: %s", err.Error(), failure.Error())
				} else {
					err = failure
				}
			}
		}
		return err
	})
	return wrapAttestationError(err)
}

func (s *attestationSubmitter) submitTo(h kosli.Host, a attestationSubmission, evidencePath string) error {
	endpoint, err := url.JoinPath(h.URL, "api/v2/attestations", s.org, a.flow, "trail", a.trail, a.slug)
	if err != nil {
		return err
	}

	payload := a.payload
	uploadID := ""
	switch {
	case a.evidence != nil:
		uploadID, err = s.uploadBytes(h, *a.evidence)
	case evidencePath != "":
		uploadID, err = s.uploadAttachment(h, evidencePath)
	}
	if err != nil {
		return err
	}
	if uploadID != "" {
		payload = withAttachmentUploadID(payload, uploadID)
	}

	_, err = h.Do(&requests.RequestParams{
		Method: http.MethodPost,
		URL:    endpoint,
		Form:   []requests.FormItem{{Type: "field", FieldName: "data_json", Content: payload}},
		DryRun: h.DryRun,
		Token:  h.Token,
	})
	return err
}

// withAttachmentUploadID returns a copy of p naming the upload id. Every host
// mints its own id, so the payload the hosts share is never changed.
func withAttachmentUploadID(p attestationPayload, id string) attestationPayload {
	v := reflect.ValueOf(p)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	clone := reflect.New(v.Type())
	clone.Elem().Set(v)
	common := clone.Elem().FieldByName("CommonAttestationPayload")
	if common.IsValid() && common.Kind() == reflect.Pointer && !common.IsNil() {
		copied := *common.Interface().(*CommonAttestationPayload)
		common.Set(reflect.ValueOf(&copied))
	}
	cloned := clone.Interface().(attestationPayload)
	cloned.setAttachmentUploadID(id)
	if reflect.ValueOf(p).Kind() == reflect.Pointer {
		return cloned
	}
	return clone.Elem().Interface().(attestationPayload)
}

// wrapAttestationError turns the server's binding error into one naming the
// flags that satisfy it.
func wrapAttestationError(err error) error {
	if err != nil {
		return fmt.Errorf("%s", strings.Replace(err.Error(), "requires at least one of: artifact_fingerprint or git_commit_info.",
			"requires at least one of: specifying the fingerprint (either by calculating it using the artifact name/path and --artifact-type, or by providing it using --fingerprint) or providing --commit (requires an available git repo to access commit details)", 1))
	}
	return err
}

// getPathOfEvidenceFileToUpload returns the path of an evidence file to upload based
// on the provided evidencePaths.
// - if one path is provided and it is a file, that path is returned as it
// - if one path is provided and it is a directory, the directory is tarred and the
// path of the generated tar file is returned
// - if multiple paths are provided, they are packaged into a tar file and the
// path of the generated tar file is returned
func getPathOfEvidenceFileToUpload(evidencePaths []string, logger *log.Logger) (string, bool, error) {
	cleanupNeeded := false
	if len(evidencePaths) == 0 {
		return "", cleanupNeeded, fmt.Errorf("no evidence paths provided")
	}
	dirToTar := ""
	if len(evidencePaths) == 1 {
		ok, err := utils.IsFile(evidencePaths[0])
		if err != nil {
			return "", cleanupNeeded, err
		}
		if ok {
			logger.Debug("file %s is provided as evidence", evidencePaths[0])
			return evidencePaths[0], cleanupNeeded, nil
		}

		ok, err = utils.IsDir(evidencePaths[0])
		if err != nil {
			return "", cleanupNeeded, err
		}
		if ok {
			logger.Debug("dir %s is provided as evidence. It will be tarred", evidencePaths[0])
			dirToTar = evidencePaths[0]
		}

	} else { // there are multiple paths
		// copy all paths to a new temp dir
		tmpDir, err := os.MkdirTemp("", "")
		if err != nil {
			return "", cleanupNeeded, err
		}
		defer func() {
			if err := os.RemoveAll(tmpDir); err != nil {
				logger.Warn("failed to remove temporary directory %s: %v", tmpDir, err)
			}
		}()

		logger.Debug("[%d] paths are provided as evidence. They will be tarred from %s", len(evidencePaths), tmpDir)

		for _, path := range evidencePaths {
			volume := filepath.VolumeName(path)
			volumeWithoutColon := strings.Replace(volume, ":", "", 1)
			pathWithoutVolume := path[len(volume):]
			pathWithoutColon := volumeWithoutColon + pathWithoutVolume

			// this copy is only a staging area for the tarball built below, and
			// nothing downstream reads uid/gid. Preserving ownership would Lchown
			// to a foreign uid, which fails for any non-root process handed
			// evidence written by a docker container.
			err := cp.Copy(path, filepath.Join(tmpDir, pathWithoutColon), cp.Options{
				PreserveTimes: true,
			})
			if err != nil {
				return "", cleanupNeeded, fmt.Errorf("failed to package attachment %s: %w", path, err)
			}
		}
		dirToTar = tmpDir
	}

	// tar the required dir and return the path of the tar file
	tarFilePath, err := utils.Tar(dirToTar, "evidence.tgz")
	if err != nil {
		return "", cleanupNeeded, err
	}
	cleanupNeeded = true
	return tarFilePath, cleanupNeeded, nil
}
