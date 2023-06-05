package testHelpers

import (
	"os"
	"testing"
)

func SkipIfEnvVarUnset(T *testing.T, requiredEnvVars []string) {
	for _, envVar := range requiredEnvVars {
		_, ok := os.LookupEnv(envVar)
		if !ok {
			T.Logf("skipping %s as %s is unset in environment", T.Name(), envVar)
			T.Skipf("requires %s", envVar)
		}
	}
}

// Originally we had commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6
// This worked for a while, but now the PR is no longer found by the github api
// for reasons we cannot fathom
// We are now using an even older commit, which currently works.
func GithubCommitWithPR() string {
	return "e21a8afff429e0c87ee523d683f2438113f0a105"
}
