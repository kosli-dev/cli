package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"

	azUtils "github.com/kosli-dev/cli/internal/azure"
	bbUtils "github.com/kosli-dev/cli/internal/bitbucket"
	ghUtils "github.com/kosli-dev/cli/internal/github"
	gitlabUtils "github.com/kosli-dev/cli/internal/gitlab"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/types"
)

type PullRequestEvidencePayload struct {
	TypedEvidencePayload
	GitProvider  string              `json:"git_provider"`
	PullRequests []*types.PREvidence `json:"pull_requests"`
}

type pullRequestOptions struct {
	payload          PullRequestEvidencePayload
	retriever        interface{}
	userDataFilePath string
	assert           bool
}

type pullRequestArtifactOptions struct {
	fingerprintOptions *fingerprintOptions
	flowName           string
	commit             string
	pullRequestOptions
}

func (o *pullRequestArtifactOptions) getRetriever() types.PRRetriever {
	return o.retriever.(types.PRRetriever)
}

func (o *pullRequestArtifactOptions) run(out io.Writer, args []string) error {
	var err error
	if o.payload.ArtifactFingerprint == "" {
		o.payload.ArtifactFingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	label := ""
	o.payload.GitProvider, label = getGitProviderAndLabel(o.retriever)

	url := fmt.Sprintf("%s/api/v2/evidence/%s/artifact/%s/pull_request", global.Host, global.Org, o.flowName)

	var pullRequestsEvidence []*types.PREvidence
	pullRequestsEvidence, err = o.getRetriever().PREvidenceForCommitV1(o.commit)
	if err != nil {
		return err
	}

	o.payload.PullRequests = pullRequestsEvidence
	o.payload.UserData, err = LoadJsonData(o.userDataFilePath)
	if err != nil {
		return err
	}

	// PR evidence does not have files to upload
	form, cleanupNeeded, evidencePath, err := newEvidenceForm(o.payload, []string{})
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer os.Remove(evidencePath)
	}

	if err != nil {
		return err
	}

	logger.Info("found %d %s(s) for commit: %s", len(pullRequestsEvidence), label, o.commit)

	reqParams := &requests.RequestParams{
		Method: http.MethodPost,
		URL:    url,
		Form:   form,
		DryRun: global.DryRun,
		Token:  global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("%s %s evidence is reported to artifact: %s", o.payload.GitProvider, label, o.payload.ArtifactFingerprint)
	}

	if len(pullRequestsEvidence) == 0 && o.pullRequestOptions.assert && !global.DryRun {
		return fmt.Errorf("assert failed: no %s found for the given commit: %s", label, o.commit)
	}
	return err
}

type PRAttestationPayload struct {
	*CommonAttestationPayload
	GitProvider  string              `json:"git_provider"`
	PullRequests []*types.PREvidence `json:"pull_requests"`
}

type attestPROptions struct {
	*CommonAttestationOptions
	retriever any
	assert    bool
	payload   PRAttestationPayload
}

func (o *attestPROptions) getRetriever() types.PRRetriever {
	return o.retriever.(types.PRRetriever)
}

func (o *attestPROptions) run(args []string) error {
	url := fmt.Sprintf("%s/api/v2/attestations/%s/%s/trail/%s/pull_request", global.Host, global.Org, o.flowName, o.trailName)

	err := o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)
	if err != nil {
		return err
	}

	label := ""
	o.payload.GitProvider, label = getGitProviderAndLabel(o.retriever)

	var pullRequestsEvidence []*types.PREvidence
	if o.payload.GitProvider == "github" || o.payload.GitProvider == "gitlab" {
		pullRequestsEvidence, err = o.getRetriever().PREvidenceForCommitV2(o.payload.Commit.Sha1)
	} else {
		pullRequestsEvidence, err = o.getRetriever().PREvidenceForCommitV1(o.payload.Commit.Sha1)
	}

	o.payload.PullRequests = pullRequestsEvidence

	form, cleanupNeeded, evidencePath, err := prepareAttestationForm(o.payload, o.attachments)
	if err != nil {
		return err
	}
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer os.Remove(evidencePath)
	}

	logger.Info("found %d %s(s) for commit: %s", len(pullRequestsEvidence), label, o.payload.Commit.Sha1)

	reqParams := &requests.RequestParams{
		Method: http.MethodPost,
		URL:    url,
		Form:   form,
		DryRun: global.DryRun,
		Token:  global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("%s %s attestation '%s' is reported to trail: %s", o.payload.GitProvider, label, o.payload.AttestationName, o.trailName)
	}

	if len(pullRequestsEvidence) == 0 && o.assert && !global.DryRun {
		return fmt.Errorf("assert failed: no %s found for the given commit: %s", label, o.payload.Commit.Sha1)
	}
	return wrapAttestationError(err)
}

type pullRequestCommitOptions struct {
	pullRequestOptions
}

func (o *pullRequestCommitOptions) getRetriever() types.PRRetriever {
	return o.retriever.(types.PRRetriever)
}

func (o *pullRequestCommitOptions) run(args []string) error {
	url := fmt.Sprintf("%s/api/v2/evidence/%s/commit/pull_request", global.Host, global.Org)

	var err error
	o.payload.UserData, err = LoadJsonData(o.userDataFilePath)
	if err != nil {
		return err
	}

	label := ""
	o.payload.GitProvider, label = getGitProviderAndLabel(o.retriever)

	var pullRequestsEvidence []*types.PREvidence
	pullRequestsEvidence, err = o.getRetriever().PREvidenceForCommitV1(o.payload.CommitSHA)
	if err != nil {
		return err
	}

	o.payload.PullRequests = pullRequestsEvidence

	// PR evidence does not have files to upload
	form, cleanupNeeded, evidencePath, err := newEvidenceForm(o.payload, []string{})
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer os.Remove(evidencePath)
	}

	if err != nil {
		return err
	}
	logger.Info("found %d %s(s) for commit: %s", len(pullRequestsEvidence), label, o.payload.CommitSHA)

	reqParams := &requests.RequestParams{
		Method: http.MethodPost,
		URL:    url,
		Form:   form,
		DryRun: global.DryRun,
		Token:  global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("%s %s evidence is reported to commit: %s", o.payload.GitProvider, label, o.payload.CommitSHA)
	}

	if len(pullRequestsEvidence) == 0 && o.pullRequestOptions.assert && !global.DryRun {
		return fmt.Errorf("assert failed: no %s found for the given commit: %s", label, o.payload.CommitSHA)
	}
	return err
}

func getGitProviderAndLabel(retriever interface{}) (string, string) {
	label := "pull request"
	provider := ""
	t := reflect.TypeOf(retriever)
	switch t {
	case reflect.TypeOf(&gitlabUtils.GitlabConfig{}):
		provider = "gitlab"
		label = "merge request"
	case reflect.TypeOf(&ghUtils.GithubConfig{}):
		provider = "github"
	case reflect.TypeOf(&azUtils.AzureConfig{}):
		provider = "azure"
	case reflect.TypeOf(&bbUtils.Config{}):
		provider = "bitbucket"
	}
	return provider, label
}
