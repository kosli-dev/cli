package intoto

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const (
	githubActionsBuildType = "https://actions.github.io/buildtypes/workflow/v1"
	slsaGitHubBuildType    = "https://slsa-framework.github.io/github-actions-buildtypes/workflow/v1"
	slsaGitHubGeneratorV02 = "https://github.com/slsa-framework/slsa-github-generator/"
	buildKitBuildTypeV1    = "https://github.com/moby/buildkit/blob/master/docs/attestations/slsa-definitions.md"
	buildKitBuildTypeV02   = "https://mobyproject.org/buildkit@v1"
	buildKitMetadataV02    = "https://mobyproject.org/buildkit@v1#metadata"
)

// Provenance is what a provenance statement says about the build that made an
// artifact, in the terms Kosli records an artifact by. It is filled from the
// statement alone, never from the environment reading it, because that may be
// a different job or workflow from the one that ran the build.
type Provenance struct {
	// Filename is the subject's name for the artifact.
	Filename  string
	BuilderID string
	BuildType string
	// BuildURL is the run that built the artifact.
	BuildURL string
	// Commit is the SHA of the git commit the artifact was built from.
	Commit     string
	Branch     string
	Repository Repository
	CommitURL  string
	// Workflow is the CI workflow file that ran the build, when the build
	// type names one.
	Workflow string
	// Missing names what a caller would want but the statement did not say,
	// as Missing* values.
	Missing []string
}

// The values of Provenance.Missing.
const (
	MissingBuildURL   = "build URL"
	MissingCommit     = "commit"
	MissingCommitURL  = "commit URL"
	MissingRepository = "repository"
)

// Repository is the source repository a provenance statement says was built.
type Repository struct {
	// URL is the repository's web address, such as https://github.com/acme/widget.
	URL string
	// Name is the repository's path on its host, such as acme/widget.
	Name string
	// ID is the host's identifier for the repository, when the statement
	// carries one.
	ID   string
	Host string
	// Provider is the Kosli repo provider for the host: github, gitlab or
	// bitbucket_cloud, or empty when the host is none of those.
	Provider string
}

// ProvenanceFrom maps a provenance attestation about the artifact with
// SHA-256 digest. repoProvider names the provider for a repository host it
// does not recognise, such as a self-hosted GitHub Enterprise.
func ProvenanceFrom(a *Attestation, digest, repoProvider string) (*Provenance, error) {
	if !KindProvenance.Matches(a.Statement.GetPredicateType()) {
		return nil, fmt.Errorf("%s is %s, not SLSA provenance", a.Describe(), a.Statement.GetPredicateType())
	}
	var predicate map[string]any
	if err := json.Unmarshal(a.Predicate, &predicate); err != nil {
		return nil, fmt.Errorf("%s: the provenance predicate is not a JSON object: %w", a.Describe(), err)
	}

	p := &Provenance{Filename: subjectName(a, digest)}
	if a.Statement.GetPredicateType() == slsaProvenanceV1 {
		p.fromV1(predicate)
	} else {
		p.fromV02(predicate)
	}
	p.finish(repoProvider)
	return p, nil
}

func subjectName(a *Attestation, digest string) string {
	digest = normaliseDigest(digest)
	for _, subject := range a.Statement.GetSubject() {
		if strings.EqualFold(subject.GetDigest()["sha256"], digest) {
			return subject.GetName()
		}
	}
	return ""
}

func (p *Provenance) fromV1(predicate map[string]any) {
	p.BuildType = str(predicate, "buildDefinition", "buildType")
	p.BuilderID = str(predicate, "runDetails", "builder", "id")
	// BuildKit writes invocationID, against the SLSA spec's invocationId.
	invocation := firstOf(str(predicate, "runDetails", "metadata", "invocationId"), str(predicate, "runDetails", "metadata", "invocationID"))
	dependencies := list(predicate, "buildDefinition", "resolvedDependencies")

	switch p.BuildType {
	case githubActionsBuildType, slsaGitHubBuildType:
		workflow := obj(predicate, "buildDefinition", "externalParameters", "workflow")
		p.BuildURL = webURL(invocation)
		p.Workflow = str(workflow, "path")
		p.Repository.URL = str(workflow, "repository")
		p.Repository.ID = str(predicate, "buildDefinition", "internalParameters", "github", "repository_id")
		p.Branch = branchOf(str(workflow, "ref"))
		p.Commit = sourceCommit(dependencies, p.Repository.URL)
	case buildKitBuildTypeV1:
		p.fromBuildKit(obj(predicate, "runDetails", "metadata", "buildkit_metadata", "vcs"),
			obj(predicate, "buildDefinition", "externalParameters", "configSource"), invocation, dependencies)
	default:
		p.BuildURL = webURL(invocation)
		p.Commit = sourceCommit(dependencies, "")
	}
	p.fillFromDependency(dependencies)
}

func (p *Provenance) fromV02(predicate map[string]any) {
	p.BuildType = str(predicate, "buildType")
	p.BuilderID = str(predicate, "builder", "id")
	// BuildKit writes buildInvocationID, against the SLSA spec's buildInvocationId.
	invocation := firstOf(str(predicate, "metadata", "buildInvocationId"), str(predicate, "metadata", "buildInvocationID"))
	configSource := obj(predicate, "invocation", "configSource")
	materials := list(predicate, "materials")

	switch {
	case strings.HasPrefix(p.BuildType, slsaGitHubGeneratorV02):
		p.BuildURL = webURL(invocation)
		p.Workflow = str(configSource, "entryPoint")
		p.Repository.ID = str(predicate, "invocation", "environment", "github_repository_id")
		p.Repository.URL, p.Branch = repoAndBranch(str(configSource, "uri"))
		p.Commit = firstOf(str(configSource, "digest", "sha1"), sourceCommit(materials, p.Repository.URL))
	case p.BuildType == buildKitBuildTypeV02:
		p.fromBuildKit(obj(predicate, "metadata", buildKitMetadataV02, "vcs"), configSource, invocation, materials)
	default:
		p.BuildURL = webURL(invocation)
		p.Commit = firstOf(str(configSource, "digest", "sha1"), sourceCommit(materials, ""))
		p.Repository.URL, p.Branch = repoAndBranch(str(configSource, "uri"))
	}
	p.fillFromDependency(materials)
}

// fromBuildKit reads BuildKit's provenance. Its invocation ID is a build
// reference, not a URL; docker/build-push-action puts the run's URL in the
// builder ID instead.
func (p *Provenance) fromBuildKit(vcs, configSource map[string]any, invocation string, dependencies []any) {
	p.BuildURL = firstOf(webURL(p.BuilderID), webURL(invocation))
	p.Repository.URL = str(vcs, "source")
	repo, branch := repoAndBranch(str(configSource, "uri"))
	p.Repository.URL = firstOf(p.Repository.URL, repo)
	p.Branch = branch
	p.Commit = firstOf(str(vcs, "revision"), str(configSource, "digest", "sha1"), sourceCommit(dependencies, p.Repository.URL))
}

// fillFromDependency takes the repository and branch from the git
// dependency that supplied the commit, when the build type did not say.
func (p *Provenance) fillFromDependency(dependencies []any) {
	for _, d := range dependencies {
		dependency, _ := d.(map[string]any)
		if p.Commit == "" || dependencyCommit(dependency) != p.Commit {
			continue
		}
		repo, branch := repoAndBranch(str(dependency, "uri"))
		p.Repository.URL = firstOf(p.Repository.URL, repo)
		p.Branch = firstOf(p.Branch, branch)
		return
	}
}

func (p *Provenance) finish(repoProvider string) {
	if repo, err := url.Parse(p.Repository.URL); err == nil && repo.Host != "" {
		p.Repository.URL = strings.TrimSuffix(strings.TrimSuffix(p.Repository.URL, "/"), ".git")
		p.Repository.Host = repo.Host
		p.Repository.Name = strings.TrimSuffix(strings.Trim(repo.Path, "/"), ".git")
		p.Repository.Provider = providerFor(repo.Host, repoProvider)
	} else {
		p.Repository = Repository{}
	}
	if p.Commit != "" && p.Repository.URL != "" {
		p.CommitURL = commitURL(p.Repository.Provider, p.Repository.URL, p.Commit)
	}

	for _, field := range []struct{ name, value string }{
		{MissingBuildURL, p.BuildURL},
		{MissingCommit, p.Commit},
		{MissingCommitURL, p.CommitURL},
		{MissingRepository, p.Repository.URL},
	} {
		if field.value == "" {
			p.Missing = append(p.Missing, field.name)
		}
	}
}

func providerFor(host, fallback string) string {
	switch strings.ToLower(host) {
	case "github.com":
		return "github"
	case "gitlab.com":
		return "gitlab"
	case "bitbucket.org":
		return "bitbucket_cloud"
	}
	return fallback
}

func commitURL(provider, repoURL, commit string) string {
	switch provider {
	case "github":
		return repoURL + "/commit/" + commit
	case "gitlab":
		return repoURL + "/-/commit/" + commit
	case "bitbucket", "bitbucket_cloud":
		return repoURL + "/commits/" + commit
	}
	return ""
}

// sourceCommit returns the commit of the first git dependency, preferring
// one from repoURL when that is known.
func sourceCommit(dependencies []any, repoURL string) string {
	fallback := ""
	for _, d := range dependencies {
		dependency, _ := d.(map[string]any)
		commit := dependencyCommit(dependency)
		if commit == "" {
			continue
		}
		repo, _ := repoAndBranch(str(dependency, "uri"))
		if repoURL == "" || sameRepo(repo, repoURL) {
			return commit
		}
		fallback = firstOf(fallback, commit)
	}
	return fallback
}

func dependencyCommit(dependency map[string]any) string {
	commit := firstOf(str(dependency, "digest", "gitCommit"), str(dependency, "digest", "sha1"))
	if commit == "" || !strings.HasPrefix(str(dependency, "uri"), "git+") && !strings.Contains(str(dependency, "uri"), ".git") {
		return ""
	}
	return commit
}

func sameRepo(a, b string) bool {
	normalise := func(s string) string {
		return strings.ToLower(strings.TrimSuffix(strings.TrimSuffix(s, "/"), ".git"))
	}
	return normalise(a) == normalise(b)
}

// repoAndBranch splits a source URI, in the forms
// git+https://host/owner/repo@refs/heads/main and
// https://host/owner/repo.git#refs/heads/main, into the repository's web
// address and the branch. The branch is empty for a tag or any other ref.
func repoAndBranch(uri string) (string, string) {
	uri = strings.TrimPrefix(uri, "git+")
	repo, ref, found := strings.Cut(uri, "#")
	if !found {
		if parsed, err := url.Parse(uri); err == nil && parsed.Host != "" {
			if path, pathRef, hasRef := strings.Cut(parsed.Path, "@"); hasRef {
				repo, ref = parsed.Scheme+"://"+parsed.Host+path, pathRef
			}
		}
	}
	if webURL(repo) == "" {
		return "", ""
	}
	return strings.TrimSuffix(repo, ".git"), branchOf(ref)
}

func branchOf(ref string) string {
	if branch, ok := strings.CutPrefix(ref, "refs/heads/"); ok {
		return branch
	}
	return ""
}

// webURL returns s when it is an http or https URL, and "" otherwise.
func webURL(s string) string {
	parsed, err := url.Parse(s)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return ""
	}
	return s
}

func firstOf(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func obj(m map[string]any, path ...string) map[string]any {
	current := m
	for _, key := range path {
		next, ok := current[key].(map[string]any)
		if !ok {
			return nil
		}
		current = next
	}
	return current
}

func str(m map[string]any, path ...string) string {
	if len(path) == 0 {
		return ""
	}
	value, _ := obj(m, path[:len(path)-1]...)[path[len(path)-1]].(string)
	return value
}

func list(m map[string]any, path ...string) []any {
	if len(path) == 0 {
		return nil
	}
	value, _ := obj(m, path[:len(path)-1]...)[path[len(path)-1]].([]any)
	return value
}
