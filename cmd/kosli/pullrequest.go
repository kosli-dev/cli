package main

import (
	"fmt"
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
	pullRequestsEvidence, err = o.getRetriever().PREvidenceForCommitV2(o.payload.Commit.Sha1)
	if err != nil {
		return err
	}

	o.payload.PullRequests = pullRequestsEvidence

	form, cleanupNeeded, evidencePath, err := prepareAttestationForm(o.payload, o.attachments)
	if err != nil {
		return err
	}
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer func() {
			if err := os.Remove(evidencePath); err != nil {
				logger.Warn("failed to remove evidence file %s: %v", evidencePath, err)
			}
		}()
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
		errString := ""
		if err != nil {
			errString = fmt.Sprintf("%s\nError: ", err.Error())
		}
		err = fmt.Errorf("%sassert failed: no %s found for the given commit: %s", errString, label, o.payload.Commit.Sha1)
	}

	return wrapAttestationError(err)
}

func getGitProviderAndLabel(retriever any) (string, string) {
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
