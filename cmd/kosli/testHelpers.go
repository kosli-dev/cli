package main

import (
	"fmt"
	"testing"

	"github.com/kosli-dev/cli/internal/utils"
	"github.com/stretchr/testify/require"
)

const ImageName = "library/alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5"

func PullExampleImage(t *testing.T) {
	err := utils.PullDockerImage(ImageName)
	require.NoError(t, err, fmt.Sprintf("pulling example image %s should work without error", ImageName))
}

// CreatePipeline creates a pipeline on the server
func CreatePipeline(pipelineName string, t *testing.T) {
	o := &pipelineDeclareOptions{
		payload: PipelinePayload{
			Name:        pipelineName,
			Description: "test pipeline",
			Visibility:  "private",
		},
	}

	err := o.run()
	require.NoError(t, err, "pipeline should be created without error")
}

// CreateArtifact creates an artifact on the server
func CreateArtifact(pipelineName, artifactFingerprint, artifactName string, t *testing.T) {
	o := &artifactCreationOptions{
		srcRepoRoot:  "../..",
		pipelineName: pipelineName,
		payload: ArtifactPayload{
			Sha256:    artifactFingerprint,
			GitCommit: "6ef6fc37c373922eecd4e823cf2633326790cfe8",
			BuildUrl:  "www.yr.no",
			CommitUrl: " www.nrk.no",
		},
	}

	err := o.run([]string{artifactName})
	require.NoError(t, err, "artifact should be created without error")
}
