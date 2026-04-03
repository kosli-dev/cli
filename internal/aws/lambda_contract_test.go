package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
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
		// Request one function per page to force pagination
		maxItems := int32(1)
		out, err := client.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{
			MaxItems: &maxItems,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		require.LessOrEqual(t, len(out.Functions), 1)

		if out.NextMarker != nil {
			// Follow the marker to prove pagination works
			out2, err := client.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{
				MaxItems: &maxItems,
				Marker:   out.NextMarker,
			})
			require.NoError(t, err)
			require.NotNil(t, out2)
			require.LessOrEqual(t, len(out2.Functions), 1)
		}
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

func TestLambdaContract_RealAWS(t *testing.T) {
	testHelpers.SkipIfEnvVarUnset(t, []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"})

	creds := &AWSStaticCreds{Region: "eu-central-1"}
	client, err := creds.NewLambdaClient()
	require.NoError(t, err)

	runLambdaContractTests(t, client, "cli-tests")
}
