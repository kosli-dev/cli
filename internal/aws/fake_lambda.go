package aws

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// FakeLambdaClient is an in-memory implementation of LambdaAPI for testing.
// It simulates marker-based pagination and returns errors for missing functions.
type FakeLambdaClient struct {
	Functions []types.FunctionConfiguration
	// PageSize controls how many functions are returned per ListFunctions call.
	// Defaults to 50 (matching the AWS default) if zero.
	PageSize int
}

func (f *FakeLambdaClient) pageSize() int {
	if f.PageSize > 0 {
		return f.PageSize
	}
	return 50
}

func (f *FakeLambdaClient) ListFunctions(_ context.Context, params *lambda.ListFunctionsInput, _ ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error) {
	pageSize := f.pageSize()
	if params.MaxItems != nil && int(*params.MaxItems) < pageSize {
		pageSize = int(*params.MaxItems)
	}

	start := 0
	if params.Marker != nil {
		parsed, err := strconv.Atoi(*params.Marker)
		if err != nil {
			return nil, fmt.Errorf("invalid marker: %s", *params.Marker)
		}
		start = parsed
	}

	end := start + pageSize
	if end > len(f.Functions) {
		end = len(f.Functions)
	}

	out := &lambda.ListFunctionsOutput{
		Functions: f.Functions[start:end],
	}
	if end < len(f.Functions) {
		marker := strconv.Itoa(end)
		out.NextMarker = &marker
	}

	return out, nil
}

func (f *FakeLambdaClient) GetFunctionConfiguration(_ context.Context, params *lambda.GetFunctionConfigurationInput, _ ...func(*lambda.Options)) (*lambda.GetFunctionConfigurationOutput, error) {
	if params.FunctionName == nil {
		return nil, fmt.Errorf("FunctionName is required")
	}
	for _, fn := range f.Functions {
		if fn.FunctionName != nil && *fn.FunctionName == *params.FunctionName {
			return &lambda.GetFunctionConfigurationOutput{
				FunctionName: fn.FunctionName,
				CodeSha256:   fn.CodeSha256,
				LastModified: fn.LastModified,
				PackageType:  fn.PackageType,
			}, nil
		}
	}
	return nil, fmt.Errorf("function not found: %s", *params.FunctionName)
}
