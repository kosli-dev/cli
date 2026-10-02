package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/utils"
	cp "github.com/otiai10/copy"
)

// attestationSubmitter records Attestations on Trails.
type attestationSubmitter struct {
	client *requests.Client
	host   string
	org    string
	token  string
	dryRun bool
	logger *log.Logger
}

func newAttestationSubmitter() *attestationSubmitter {
	return &attestationSubmitter{
		client: kosliClient,
		host:   global.Host,
		org:    global.Org,
		token:  global.ApiToken,
		dryRun: global.DryRun,
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
	// evidence is an attachment already read into memory; it replaces
	// attachments.
	evidence *requests.FileBytes
	// assertFailures fail the command after the Attestation is recorded, so
	// the evidence is kept even when --assert rejects it.
	assertFailures []error
}

// attestationPayload is the body of any Attestation type; every one embeds
// CommonAttestationPayload, which provides the name.
type attestationPayload interface {
	attestationName() string
}

func (s *attestationSubmitter) submit(a attestationSubmission) error {
	url, err := url.JoinPath(s.host, "api/v2/attestations", s.org, a.flow, "trail", a.trail, a.slug)
	if err != nil {
		return err
	}

	form, cleanup, err := s.form(a)
	if err != nil {
		return err
	}
	defer cleanup()

	_, err = s.client.Do(&requests.RequestParams{
		Method: http.MethodPost,
		URL:    url,
		Form:   form,
		DryRun: s.dryRun,
		Token:  s.token,
	})
	if err == nil && !s.dryRun {
		s.logger.Info("%s attestation '%s' is reported to trail: %s", a.label, a.payload.attestationName(), a.trail)
	}
	if !s.dryRun {
		for _, failure := range a.assertFailures {
			if err != nil {
				err = fmt.Errorf("%s\nError: %s", err.Error(), failure.Error())
			} else {
				err = failure
			}
		}
	}
	return wrapAttestationError(err)
}

// form returns the multipart form for a, and a cleanup that removes any
// evidence tarball the form refers to.
func (s *attestationSubmitter) form(a attestationSubmission) ([]requests.FormItem, func(), error) {
	form := []requests.FormItem{
		{Type: "field", FieldName: "data_json", Content: a.payload},
	}
	noCleanup := func() {}

	if a.evidence != nil {
		return append(form, requests.FormItem{Type: "file-bytes", FieldName: "attachment_file", Content: *a.evidence}), noCleanup, nil
	}
	if len(a.attachments) == 0 {
		return form, noCleanup, nil
	}

	evidencePath, cleanupNeeded, err := getPathOfEvidenceFileToUpload(a.attachments)
	if err != nil {
		return nil, noCleanup, err
	}
	s.logger.Debug("evidence file %s will be uploaded", evidencePath)
	form = append(form, requests.FormItem{Type: "file", FieldName: "attachment_file", Content: evidencePath})
	if !cleanupNeeded {
		return form, noCleanup, nil
	}
	return form, func() {
		if err := os.Remove(evidencePath); err != nil {
			s.logger.Warn("failed to remove evidence file: %v", err)
		}
	}, nil
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
func getPathOfEvidenceFileToUpload(evidencePaths []string) (string, bool, error) {
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
		defer func() {
			if err := os.RemoveAll(tmpDir); err != nil {
				logger.Warn("failed to remove temporary directory %s: %v", tmpDir, err)
			}
		}()
	}

	// tar the required dir and return the path of the tar file
	tarFilePath, err := utils.Tar(dirToTar, "evidence.tgz")
	if err != nil {
		return "", cleanupNeeded, err
	}
	cleanupNeeded = true
	return tarFilePath, cleanupNeeded, nil
}
