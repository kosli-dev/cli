package main

import (
	"fmt"

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
	o.commitRequiredFor = "find pull requests"
	err := o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)
	if err != nil {
		return err
	}

	label := ""
	o.payload.GitProvider, label = o.getRetriever().ProviderAndLabel()

	var pullRequestsEvidence []*types.PREvidence
	pullRequestsEvidence, err = o.getRetriever().PREvidenceForCommitHybrid(o.payload.Commit.Sha1)
	if err != nil {
		return err
	}

	o.payload.PullRequests = pullRequestsEvidence

	logger.Info("found %d %s(s) for commit: %s", len(pullRequestsEvidence), label, o.payload.Commit.Sha1)

	var assertFailures []error
	if o.assert && len(pullRequestsEvidence) == 0 {
		assertFailures = append(assertFailures, fmt.Errorf("assert failed: no %s found for the given commit: %s", label, o.payload.Commit.Sha1))
	}

	return newAttestationSubmitter().submit(attestationSubmission{
		flow:           o.flowName,
		trail:          o.trailName,
		slug:           "pull_request",
		label:          o.payload.GitProvider + " " + label,
		payload:        o.payload,
		attachments:    o.attachments,
		assertFailures: assertFailures,
	})
}
