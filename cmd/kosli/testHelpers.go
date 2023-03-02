package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// CreateFlow creates a flow on the server
func CreateFlow(flowName string, t *testing.T) {
	o := &createFlowOptions{
		payload: FlowPayload{
			Name:        flowName,
			Description: "test flow",
			Visibility:  "private",
		},
	}

	err := o.run([]string{flowName})
	require.NoError(t, err, "flow should be created without error")
}

// CreateArtifact creates an artifact on the server
func CreateArtifact(flowName, artifactFingerprint, artifactName string, t *testing.T) {
	o := &reportArtifactOptions{
		srcRepoRoot: "../..",
		flowName:    flowName,
		payload: ArtifactPayload{
			Fingerprint: artifactFingerprint,
			GitCommit:   "0fc1ba9876f91b215679f3649b8668085d820ab5",
			BuildUrl:    "www.yr.no",
			CommitUrl:   " www.nrk.no",
		},
	}

	err := o.run([]string{artifactName})
	require.NoError(t, err, "artifact should be created without error")
}

// CreateApproval creates an approval for an artifact in a flow
func CreateApproval(flowName, fingerprint string, t *testing.T) {
	o := &reportApprovalOptions{
		payload: ApprovalPayload{
			ArtifactFingerprint: fingerprint,
			Description:         "some description",
		},
		flowName:        flowName,
		oldestSrcCommit: "HEAD~1",
		newestSrcCommit: "HEAD",
		srcRepoRoot:     "../..",
	}

	err := o.run([]string{"filename"}, false)
	require.NoError(t, err, "approval should be created without error")
}

// ExpectDeployment reports a deployment expectation of a given artifact to the server
func ExpectDeployment(flowName, fingerprint, envName string, t *testing.T) {
	o := &expectDeploymentOptions{
		flowName: flowName,
		payload: ExpectDeploymentPayload{
			Fingerprint: fingerprint,
			Environment: envName,
			BuildUrl:    "https://example.com",
		},
	}
	err := o.run([]string{})
	require.NoError(t, err, "deployment should be expected without error")
}

// CreateEnv creates an env on the server
func CreateEnv(owner, envName, envType string, t *testing.T) {
	o := &createEnvOptions{
		payload: CreateEnvironmentPayload{
			Owner:       owner,
			Name:        envName,
			Type:        envType,
			Description: "test env",
		},
	}

	err := o.run([]string{envName})
	require.NoError(t, err, "env should be created without error")
}

// ReportServerArtifactToEnv reports files/dirs in paths as server env artifacts
func ReportServerArtifactToEnv(paths []string, envName string, t *testing.T) {
	o := &snapshotServerOptions{
		paths: paths,
	}
	err := o.run([]string{envName})
	require.NoError(t, err, "server env should be reported without error")
}
