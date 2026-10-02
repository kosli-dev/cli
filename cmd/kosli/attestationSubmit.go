package main

import (
	"net/http"
	"net/url"
	"os"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
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

	form, cleanupNeeded, evidencePath, err := newAttestationForm(a.payload, a.attachments)
	if err != nil {
		return err
	}
	if cleanupNeeded {
		defer func() {
			if err := os.Remove(evidencePath); err != nil {
				s.logger.Warn("failed to remove evidence file: %v", err)
			}
		}()
	}

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
	return wrapAttestationError(err)
}
