package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/kosli-dev/cli/internal/requests"
)

const commitDescription = `You can optionally associate the attestation to a git commit using ^--commit^ (requires access to a git repo).
You can optionally redact some of the git commit data sent to Kosli using ^--redact-commit-info^.
Note that when the attestation is reported for an artifact that does not yet exist in Kosli, ^--commit^ is required to facilitate
binding the attestation to the right artifact.
To record repository information, all three of ^--repo-id^, ^--repo-url^, and ^--repository^ must be set together.
These are automatically set in GitHub Actions, GitLab CI, Bitbucket Pipelines, and Azure DevOps.
In other CI systems, set them explicitly to capture repository metadata.`

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
	repoID                  string
	repoName                string
	repoURL                 string
	repoProvider            string
	repoURLExplicit         bool
	repoNameExplicit        bool
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

	if err := validateRepoFlags(o.repoURL, o.repoProvider, o.repoURLExplicit); err != nil {
		return err
	}
	payload.GitRepoInfo = mergeGitRepoInfo(payload.GitRepoInfo, o.repoID, o.repoName, o.repoURL, o.repoProvider, o.repoNameExplicit)

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

// mergeGitRepoInfo applies flag overrides onto base (which may be nil) and
// returns nil if ID, Name, or URL is still empty after merging, so that the
// field is omitted from the JSON payload.
//
// --repository only overrides the CI-detected name when set explicitly (or when
// base has none), so its short default doesn't clobber the fuller CI value
// (e.g. GitLab's CI_PROJECT_PATH).
func mergeGitRepoInfo(base *gitview.GitRepoInfo, repoID, repoName, repoURL, repoProvider string, repoNameExplicit bool) *gitview.GitRepoInfo {
	if base == nil {
		base = &gitview.GitRepoInfo{}
	}
	if repoID != "" {
		base.ID = repoID
	}
	if repoName != "" && (repoNameExplicit || base.Name == "") {
		base.Name = repoName
	}
	if repoURL != "" {
		base.URL = repoURL
	}
	if repoProvider != "" {
		base.Provider = repoProvider
	}
	if base.ID == "" || base.Name == "" || base.URL == "" {
		logger.Warn("Repo information will not be reported as ID, Name and URL are required.")
		return nil
	}
	return base
}

var allowedRepoProviders = map[string]struct{}{
	"github": {}, "gitlab": {},
	"bitbucket": {}, "bitbucket_cloud": {}, "bitbucket_dc": {},
	"azure-devops": {}, "azure_devops_services": {}, "azure_devops_server": {},
}

func validateRepoFlags(repoURL, repoProvider string, validateURL bool) error {
	if repoURL != "" && validateURL {
		parsed, err := url.Parse(repoURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("--repo-url '%s' is not a valid URL", repoURL)
		}
	}
	if repoProvider != "" {
		if _, ok := allowedRepoProviders[repoProvider]; !ok {
			return fmt.Errorf("--repo-provider '%s' is not allowed. Must be one of: github, gitlab, "+
				"bitbucket, bitbucket_cloud, bitbucket_dc, azure-devops, azure_devops_services, azure_devops_server", repoProvider)
		}
	}
	return nil
}

func processAnnotations(annotations map[string]string) (map[string]string, error) {
	for label := range annotations {
		if !regexp.MustCompile(`^[A-Za-z0-9_]+$`).MatchString(label) {
			return nil, fmt.Errorf("--annotate flag should be in the format key=value. Invalid key: '%s'. Key can only contain [A-Za-z0-9_]", label)
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
	p1, p2, found := strings.Cut(template, ".")
	// No dot: treat the whole string as the attestation name
	if !found {
		return template, "", nil
	}
	// Reject empty sides (e.g. ".foo", "foo.") or multiple dots (e.g. "foo.bar.baz")
	if p1 == "" || p2 == "" || strings.Contains(p2, ".") {
		return "", "", fmt.Errorf("invalid attestation name format: %s", template)
	}
	return p1, p2, nil
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
	// case codeBuild:
	// 	return getGitRepoInfoFromCodeBuild(), nil
	case unknown:
		return nil, nil
	}
	return nil, fmt.Errorf("unsupported CI: %s", ci)
}

func getGitRepoInfoFromGitHub() *gitview.GitRepoInfo {
	return &gitview.GitRepoInfo{
		URL:      fmt.Sprintf("%s/%s", os.Getenv("GITHUB_SERVER_URL"), os.Getenv("GITHUB_REPOSITORY")),
		Name:     os.Getenv("GITHUB_REPOSITORY"),
		ID:       os.Getenv("GITHUB_REPOSITORY_ID"),
		Provider: "github",
	}
}

func getGitRepoInfoFromGitLab() *gitview.GitRepoInfo {
	return &gitview.GitRepoInfo{
		URL:           os.Getenv("CI_PROJECT_URL"),
		Name:          os.Getenv("CI_PROJECT_PATH"),
		ID:            os.Getenv("CI_PROJECT_ID"),
		Description:   os.Getenv("CI_PROJECT_DESCRIPTION"),
		Provider:      "gitlab",
		NamespacePath: splitNonEmpty(os.Getenv("CI_PROJECT_NAMESPACE")),
	}
}

func getGitRepoInfoFromBitbucket() *gitview.GitRepoInfo {
	repoFullName := os.Getenv("BITBUCKET_REPO_FULL_NAME")
	workspace, _, _ := strings.Cut(repoFullName, "/")

	var additionalInfo map[string]interface{}
	if projectKey := os.Getenv("BITBUCKET_PROJECT_KEY"); projectKey != "" {
		additionalInfo = map[string]interface{}{"project_key": projectKey}
	}

	return &gitview.GitRepoInfo{
		URL:  os.Getenv("BITBUCKET_GIT_HTTP_ORIGIN"),
		Name: repoFullName,
		ID:   os.Getenv("BITBUCKET_REPO_UUID"),
		// Bitbucket Pipelines (the only CI this WhichCI() branch detects, via
		// BITBUCKET_BUILD_NUMBER) exists for Bitbucket Cloud only, so this is
		// a known fact rather than a heuristic. Self-hosted Data Center users
		// run a different CI and must pass --repo-provider bitbucket_dc themselves.
		Provider: "bitbucket_cloud",
		// Bitbucket's workspace is the only path segment ahead of the repo itself
		// (the project key is a separate, non-hierarchical grouping within a
		// workspace, so it goes in AdditionalInfo rather than NamespacePath).
		NamespacePath:  splitNonEmpty(workspace),
		AdditionalInfo: additionalInfo,
	}
}

func getGitRepoInfoFromAzureDevops() *gitview.GitRepoInfo {
	info := &gitview.GitRepoInfo{
		URL:      os.Getenv("BUILD_REPOSITORY_URI"),
		Name:     os.Getenv("BUILD_REPOSITORY_NAME"),
		ID:       os.Getenv("BUILD_REPOSITORY_ID"),
		Provider: "azure-devops",
	}

	// Path composition and provider refinement only make sense for genuine
	// Azure Repos Git repos (TfsGit); for any other source (GitHub, Bitbucket,
	// generic Git, TFVC, ...) they'd be wrong. Empty ⇒ older agent, assume TfsGit.
	if provider := os.Getenv("BUILD_REPOSITORY_PROVIDER"); provider == "TfsGit" || provider == "" {
		info.Name = azureFullPathRepoName()
		info.Provider = azureDevopsProvider()
		info.NamespacePath = azureNamespacePath()
	}

	return info
}

// splitNonEmpty splits s on "/", returning nil for an empty string so the
// NamespacePath field is omitted from the JSON payload rather than sent as [""].
func splitNonEmpty(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "/")
}

// azureDevopsProvider refines the coarse "azure-devops" provider to
// azure_devops_services or azure_devops_server based on SYSTEM_COLLECTIONURI's
// host (dev.azure.com / *.visualstudio.com are cloud-hosted Services;
// anything else is on-prem Server). Falls back to the coarse value when the
// collection URI is missing or has no host, so the server's own URL-based
// derivation fallback still applies.
func azureDevopsProvider() string {
	parsed, err := url.Parse(os.Getenv("SYSTEM_COLLECTIONURI"))
	if err != nil || parsed.Host == "" {
		return "azure-devops"
	}
	host := strings.ToLower(parsed.Host)
	if host == "dev.azure.com" || strings.HasSuffix(host, ".visualstudio.com") {
		return "azure_devops_services"
	}
	return "azure_devops_server"
}

// azureFullPathRepoName composes the full namespace path for an Azure DevOps
// repo (`Collection/Project/repo` on-prem, `Org/Project/repo` in Services)
// from SYSTEM_COLLECTIONURI and SYSTEM_TEAMPROJECT, since BUILD_REPOSITORY_NAME
// alone is just the bare repo name. Falls back to the bare name if either
// piece is unavailable, so unupgraded setups degrade gracefully.
func azureFullPathRepoName() string {
	repoName := os.Getenv("BUILD_REPOSITORY_NAME")
	collection, teamProject, ok := azureCollectionAndProject()
	if !ok {
		return repoName
	}
	return fmt.Sprintf("%s/%s/%s", collection, teamProject, repoName)
}

// azureNamespacePath returns the [collection, project] path ahead of the repo
// name itself, or nil when either piece is unavailable.
func azureNamespacePath() []string {
	collection, teamProject, ok := azureCollectionAndProject()
	if !ok {
		return nil
	}
	return []string{collection, teamProject}
}

// azureCollectionAndProject extracts the collection/org name and
// SYSTEM_TEAMPROJECT. ok is false if either piece is missing or
// SYSTEM_COLLECTIONURI can't be parsed into a non-empty collection.
//
// On *.visualstudio.com hosts (the legacy Services URL format) the org name
// is the subdomain, not a path segment - e.g. https://fabrikam.visualstudio.com/
// has an empty path. Everywhere else (dev.azure.com/MyOrg, on-prem Server
// collection URIs) the collection is the last path segment.
func azureCollectionAndProject() (collection, teamProject string, ok bool) {
	teamProject = os.Getenv("SYSTEM_TEAMPROJECT")
	collectionURI := os.Getenv("SYSTEM_COLLECTIONURI")
	if teamProject == "" || collectionURI == "" {
		return "", "", false
	}

	parsed, err := url.Parse(collectionURI)
	if err != nil || parsed.Host == "" {
		return "", "", false
	}

	if host := strings.ToLower(parsed.Host); strings.HasSuffix(host, ".visualstudio.com") {
		collection = strings.TrimSuffix(host, ".visualstudio.com")
		return collection, teamProject, collection != ""
	}

	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	collection = segments[len(segments)-1]
	if collection == "" {
		return "", "", false
	}

	return collection, teamProject, true
}

func getGitRepoInfoFromCircleci() *gitview.GitRepoInfo {
	return &gitview.GitRepoInfo{
		URL:      os.Getenv("CIRCLE_REPOSITORY_URL"),
		Name:     os.Getenv("CIRCLE_PROJECT_REPONAME"),
		Provider: "circleci",
	}
}

// func getGitRepoInfoFromCodeBuild() *gitview.GitRepoInfo {
// 	return &gitview.GitRepoInfo{
// 		URL: os.Getenv("CODEBUILD_SOURCE_REPO_URL"),
// 	}
// }
