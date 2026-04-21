package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/require"
)

// runLambdaContractTests exercises the LambdaAPI contract. It verifies the
// behaviours we depend on — pagination, function config retrieval, and error
// responses for missing functions.
//
// Any implementation that passes this suite is a valid stand-in for the real
// AWS Lambda API as far as this codebase is concerned.
//
// existingFunctionName must name a function that the client can see.
func runLambdaContractTests(t *testing.T, client LambdaAPI, existingFunctionName string) {
	t.Helper()

	t.Run("ListFunctions returns without error", func(t *testing.T) {
		out, err := client.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{})
		require.NoError(t, err)
		require.NotNil(t, out)
	})

	t.Run("ListFunctions with MaxItems paginates via Marker", func(t *testing.T) {
		// Request one function per page to force pagination.
		// The test setup must seed at least 2 functions for this to exercise
		// the marker path — if there's only one, NextMarker will be nil and
		// the test degrades to a no-op.
		maxItems := int32(1)
		out, err := client.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{
			MaxItems: &maxItems,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		require.LessOrEqual(t, len(out.Functions), 1)

		// Follow the marker to prove pagination works
		if out.NextMarker == nil {
			t.Skip("only 1 function in account; pagination not exercisable")
		}
		out2, err := client.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{
			MaxItems: &maxItems,
			Marker:   out.NextMarker,
		})
		require.NoError(t, err)
		require.NotNil(t, out2)
		require.LessOrEqual(t, len(out2.Functions), 1)
		require.NotEmpty(t, out2.Functions, "second page should return at least one function")
	})

	t.Run("GetFunctionConfiguration returns config for existing function", func(t *testing.T) {
		out, err := client.GetFunctionConfiguration(context.TODO(), &lambda.GetFunctionConfigurationInput{
			FunctionName: &existingFunctionName,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		require.NotNil(t, out.FunctionName)
		require.Equal(t, existingFunctionName, *out.FunctionName)
		require.NotNil(t, out.CodeSha256, "CodeSha256 should be present")
		require.NotNil(t, out.LastModified, "LastModified should be present")
	})

	t.Run("GetFunctionConfiguration errors for missing function", func(t *testing.T) {
		missingName := "nonexistent-function-that-should-not-exist-" + t.Name()
		_, err := client.GetFunctionConfiguration(context.TODO(), &lambda.GetFunctionConfigurationInput{
			FunctionName: &missingName,
		})
		require.Error(t, err)
	})
}

func TestLambdaContract_Fake(t *testing.T) {
	fnName1 := "my-function"
	fnName2 := "other-function"
	lastModified := "2024-01-15T10:30:00.000+0000"
	codeSha256 := "abc123"
	client := &FakeLambdaClient{
		Functions: []types.FunctionConfiguration{
			{
				FunctionName: &fnName1,
				CodeSha256:   &codeSha256,
				LastModified: &lastModified,
				PackageType:  types.PackageTypeZip,
			},
			{
				FunctionName: &fnName2,
				CodeSha256:   &codeSha256,
				LastModified: &lastModified,
				PackageType:  types.PackageTypeZip,
			},
		},
		PageSize: 1,
	}
	runLambdaContractTests(t, client, fnName1)
}

func TestLambdaContract_RealAWS(t *testing.T) {
	testHelpers.SkipIfEnvVarUnset(t, []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"})

	creds := &AWSStaticCreds{Region: "eu-central-1"}
	client, err := creds.NewLambdaClient()
	require.NoError(t, err)

	runLambdaContractTests(t, client, "cli-tests")
}
