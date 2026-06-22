package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// FakeECSClient is an in-memory implementation of ECSServicesAPI for testing.
// It returns the configured services for a cluster and reproduces the real AWS
// contract where DescribeServices rejects an empty Services list.
type FakeECSClient struct {
	// ServiceArns is the list of service ARNs returned by ListServices.
	ServiceArns []string
	// Services is the list of services returned by DescribeServices.
	Services []ecsTypes.Service
}

func (f *FakeECSClient) ListServices(_ context.Context, _ *ecs.ListServicesInput, _ ...func(*ecs.Options)) (*ecs.ListServicesOutput, error) {
	return &ecs.ListServicesOutput{ServiceArns: f.ServiceArns}, nil
}

func (f *FakeECSClient) DescribeServices(_ context.Context, params *ecs.DescribeServicesInput, _ ...func(*ecs.Options)) (*ecs.DescribeServicesOutput, error) {
	// Mirror the real AWS ECS API, which rejects an empty Services list with
	// InvalidParameterException: Services cannot be empty.
	if len(params.Services) == 0 {
		return nil, fmt.Errorf("InvalidParameterException: Services cannot be empty")
	}
	return &ecs.DescribeServicesOutput{Services: f.Services}, nil
}
