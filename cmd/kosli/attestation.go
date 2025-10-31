package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/kosli-dev/cli/internal/requests"
)

const commitDescription = `You can optionally associate the attestation to a git commit using ^--commit^ (requires access to a git repo).
You can optionally redact some of the git commit data sent to Kosli using ^--redact-commit-info^.
Note that when the attestation is reported for an artifact that does not yet exist in Kosli, ^--commit^ is required to facilitate
binding the attestation to the right artifact.`

type URLInfo struct {
	Href        string `json:"href"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

type CommonAttestationPayload struct {
	ArtifactFingerprint string                   `json:"artifact_fingerprint,omitempty"`
	Commit              *gitview.BasicCommitInfo `json:"git_commit_info,omitempty"`
	GitRepoInfo         *gitview.GitRepoInfo     `json:"repo_info,omitempty"`
	AttestationName     string                   `json:"attestation_name"`
	TargetArtifacts     []string                 `json:"target_artifacts,omitempty"`
	ExternalURLs        map[string]*URLInfo      `json:"external_urls,omitempty"`
	OriginURL           string                   `json:"origin_url,omitempty"`
	UserData            interface{}              `json:"user_data,omitempty"`
	Description         string                   `json:"description,omitempty"`
	Annotations         map[string]string        `json:"annotations,omitempty"`
}

type CommonAttestationOptions struct {
	fingerprintOptions      *fingerprintOptions
	attestationNameTemplate string
	flowName                string
	trailName               string
	userDataFilePath        string
	attachments             []string
	commitSHA               string
	redactedCommitInfo      []string
	srcRepoRoot             string
	externalURLs            map[string]string
	externalFingerprints    map[string]string
	annotations             map[string]string
}

func (o *CommonAttestationOptions) run(args []string, payload *CommonAttestationPayload) error {
	var err error

	p1, p2, err := parseAttestationNameTemplate(o.attestationNameTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse attestation name: %s", err)
	}
	if p1 != "" && p2 != "" {
		payload.TargetArtifacts = []string{p1}
		payload.AttestationName = p2
	} else {
		payload.AttestationName = p1
	}

	if o.fingerprintOptions.artifactType != "" {
		payload.ArtifactFingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return fmt.Errorf("failed to calculate artifact fingerprint: %s", err)
		}
	}

	if o.commitSHA != "" {
		gv, err := gitview.New(o.srcRepoRoot)
		if err != nil {
			return fmt.Errorf("failed to get commit info. %s", err)
		}
		commitInfo, err := gv.GetCommitInfoFromCommitSHA(o.commitSHA, false, o.redactedCommitInfo)
		if err != nil {
			return fmt.Errorf("failed to get commit info. %s", err)
		}
		payload.Commit = &commitInfo.BasicCommitInfo
	}

	payload.GitRepoInfo, err = getGitRepoInfoFromEnvironment()
	if err != nil {
		logger.Warn("failed to get git repo info. %s", err.Error())
	}

	payload.UserData, err = LoadJsonData(o.userDataFilePath)
	if err != nil {
		return fmt.Errorf("failed to load user data. %s", err)
	}

	// process external urls
	payload.ExternalURLs, err = processExternalURLs(o.externalURLs, o.externalFingerprints)
	if err != nil {
		return err
	}

	// process annotations
	payload.Annotations, err = processAnnotations(o.annotations)
	return err
}

func processAnnotations(annotations map[string]string) (map[string]string, error) {
	for label := range annotations {
		if !regexp.MustCompile(`^[A-Za-z0-9_]+$`).MatchString(label) {
			return nil, fmt.Errorf("--annotate flag should be in the format key=value. Invalid key: '%s'. Key can only contain [A-Za-z0-9_].", label)
		}
	}
	return annotations, nil
}

func processExternalURLs(externalURLs, externalFingerprints map[string]string) (map[string]*URLInfo, error) {
	processedExternalURLs := make(map[string]*URLInfo)
	if len(externalFingerprints) > len(externalURLs) {
		return processedExternalURLs, fmt.Errorf("--external-fingerprints have labels that don't have a URL in --external-url")
	}

	for label, url := range externalURLs {
		processedExternalURLs[label] = &URLInfo{Href: url}
	}
	for label, fingerprint := range externalFingerprints {
		if urlInfo, exists := processedExternalURLs[label]; exists {
			urlInfo.Fingerprint = fingerprint
		} else {
			return processedExternalURLs, fmt.Errorf("%s in --external-fingerprint does not match any labels in --external-url", label)
		}
	}
	return processedExternalURLs, nil
}

func prepareAttestationForm(payload interface{}, evidencePaths []string) ([]requests.FormItem, bool, string, error) {
	form, cleanupNeeded, evidencePath, err := newAttestationForm(payload, evidencePaths)
	if err != nil {
		return []requests.FormItem{}, cleanupNeeded, evidencePath, err
	}
	return form, cleanupNeeded, evidencePath, nil
}

func parseAttestationNameTemplate(template string) (string, string, error) {
	parts := strings.Split(template, ".")
	if len(parts) == 1 {
		return parts[0], "", nil
	} else if len(parts) == 2 {
		return parts[0], parts[1], nil
	} else {
		return "", "", fmt.Errorf("invalid attestation name format: %s", template)
	}
}

// newAttestationForm constructs a list of FormItems for an attestation
// form submission.
func newAttestationForm(payload interface{}, attachments []string) (
	[]requests.FormItem, bool, string, error,
) {
	form := []requests.FormItem{
		{Type: "field", FieldName: "data_json", Content: payload},
	}

	var evidencePath string
	var cleanupNeeded bool
	var err error

	if len(attachments) > 0 {
		evidencePath, cleanupNeeded, err = getPathOfEvidenceFileToUpload(attachments)
		if err != nil {
			return form, cleanupNeeded, evidencePath, err
		}
		form = append(form, requests.FormItem{Type: "file", FieldName: "attachment_file", Content: evidencePath})
		logger.Debug("evidence file %s will be uploaded", evidencePath)
	}

	return form, cleanupNeeded, evidencePath, nil
}

func wrapAttestationError(err error) error {
	if err != nil {
		return fmt.Errorf("%s", strings.Replace(err.Error(), "requires at least one of: artifact_fingerprint or git_commit_info.",
			"requires at least one of: specifying the fingerprint (either by calculating it using the artifact name/path and --artifact-type, or by providing it using --fingerprint) or providing --commit (requires an available git repo to access commit details)", 1))
	}
	return err
}

func getGitRepoInfoFromEnvironment() (*gitview.GitRepoInfo, error) {
	ci := WhichCI()
	switch ci {
	case github:
		return getGitRepoInfoFromGitHub(), nil
	case gitlab:
		return getGitRepoInfoFromGitLab(), nil
	case bitbucket:
		return getGitRepoInfoFromBitbucket(), nil
	case azureDevops:
		return getGitRepoInfoFromAzureDevops(), nil
	case circleci:
		return getGitRepoInfoFromCircleci(), nil
	case codeBuild:
		return getGitRepoInfoFromCodeBuild(), nil
	}
	return nil, fmt.Errorf("unsupported CI: %s", ci)
}

func getGitRepoInfoFromGitHub() *gitview.GitRepoInfo {
	return &gitview.GitRepoInfo{
		URL:  fmt.Sprintf("%s/%s", os.Getenv("GITHUB_SERVER_URL"), os.Getenv("GITHUB_REPOSITORY")),
		Name: os.Getenv("GITHUB_REPOSITORY"),
		ID:   os.Getenv("GITHUB_REPOSITORY_ID"),
	}
}

func getGitRepoInfoFromGitLab() *gitview.GitRepoInfo {
	return &gitview.GitRepoInfo{
		URL:         os.Getenv("CI_PROJECT_URL"),
		Name:        os.Getenv("CI_PROJECT_PATH"),
		ID:          os.Getenv("CI_PROJECT_ID"),
		Description: os.Getenv("CI_PROJECT_DESCRIPTION"),
	}
}

func getGitRepoInfoFromBitbucket() *gitview.GitRepoInfo {
	return &gitview.GitRepoInfo{
		URL:  os.Getenv("BITBUCKET_GIT_HTTP_ORIGIN"),
		Name: os.Getenv("BITBUCKET_REPO_FULL_NAME"),
		ID:   os.Getenv("BITBUCKET_REPO_UUID"),
	}
}

func getGitRepoInfoFromAzureDevops() *gitview.GitRepoInfo {
	return &gitview.GitRepoInfo{
		URL:  os.Getenv("BUILD_REPOSITORY_URI"),
		Name: os.Getenv("BUILD_REPOSITORY_NAME"),
		ID:   os.Getenv("BUILD_REPOSITORY_ID"),
	}
}

func getGitRepoInfoFromCircleci() *gitview.GitRepoInfo {
	return &gitview.GitRepoInfo{
		URL:  os.Getenv("CIRCLE_REPOSITORY_URL"),
		Name: os.Getenv("CIRCLE_PROJECT_REPONAME"),
	}
}

func getGitRepoInfoFromCodeBuild() *gitview.GitRepoInfo {
	return &gitview.GitRepoInfo{
		URL: os.Getenv("CODEBUILD_SOURCE_REPO_URL"),
	}
}
