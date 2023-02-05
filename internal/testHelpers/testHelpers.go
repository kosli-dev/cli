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
