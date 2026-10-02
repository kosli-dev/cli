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
	payload     any
	attachments []string
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
	return err
}
