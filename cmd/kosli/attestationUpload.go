package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/kosli-dev/cli/internal/requests"
)

// dryRunUploadID stands in for the id a real mint would return, so a dry run
// can still show the attestation payload the id goes in.
const dryRunUploadID = "<upload id from the mint request>"

type uploadRequest struct {
	Filename  string `json:"filename"`
	SizeBytes int64  `json:"size_bytes"`
	Sha256    string `json:"sha256"`
}

type uploadResponse struct {
	UploadID string            `json:"upload_id"`
	URL      string            `json:"url"`
	Method   string            `json:"method"`
	Headers  map[string]string `json:"headers"`
}

// uploadAttachment mints an upload for the file at path, sends the file to the
// landing bucket and returns the upload id the attestation attaches it by.
// The server never carries the bytes: S3 checks them against the SHA-256 the
// mint declared.
func (s *attestationSubmitter) uploadAttachment(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to read attachment %s: %w", path, err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	mintURL, err := url.JoinPath(s.host, "api/v2/attestations", s.org, "uploads")
	if err != nil {
		return "", err
	}
	resp, err := s.client.Do(&requests.RequestParams{
		Method: http.MethodPost,
		URL:    mintURL,
		Payload: uploadRequest{
			Filename:  filepath.Base(path),
			SizeBytes: info.Size(),
			Sha256:    fmt.Sprintf("%x", hasher.Sum(nil)),
		},
		DryRun: s.dryRun,
		Token:  s.token,
	})
	if err != nil {
		return "", err
	}
	if s.dryRun {
		return dryRunUploadID, nil
	}
	var minted uploadResponse
	if err := json.Unmarshal([]byte(resp.Body), &minted); err != nil {
		return "", fmt.Errorf("failed to parse the upload response: %w", err)
	}

	s.logger.Debug("uploading %s (%d bytes) to %s", path, info.Size(), minted.URL)
	err = s.client.Upload(&requests.UploadParams{
		Method:  minted.Method,
		URL:     minted.URL,
		Headers: minted.Headers,
		Body:    file,
		Size:    info.Size(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload attachment %s: %w", path, err)
	}
	return minted.UploadID, nil
}
