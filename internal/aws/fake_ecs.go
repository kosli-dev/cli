package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// FakeECSClient is an in-memory implementation of ECSClustersAPI and ECSServicesAPI
// for testing. It returns the configured clusters and services and reproduces the
// real AWS contract where DescribeServices rejects an empty Services list.
type FakeECSClient struct {
	// Clusters is the list of clusters returned by DescribeClusters, in one page.
	Clusters []ecsTypes.Cluster
	// ServiceArns is the list of service ARNs returned by ListServices.
	ServiceArns []string
	// Services is the list of services returned by DescribeServices.
	Services []ecsTypes.Service
}

func (f *FakeECSClient) ListClusters(_ context.Context, _ *ecs.ListClustersInput, _ ...func(*ecs.Options)) (*ecs.ListClustersOutput, error) {
	arns := make([]string, len(f.Clusters))
	for i, c := range f.Clusters {
		arns[i] = "arn:aws:ecs:eu-central-1:123:cluster/" + *c.ClusterName
	}
	return &ecs.ListClustersOutput{ClusterArns: arns}, nil
}

func (f *FakeECSClient) DescribeClusters(_ context.Context, _ *ecs.DescribeClustersInput, _ ...func(*ecs.Options)) (*ecs.DescribeClustersOutput, error) {
	return &ecs.DescribeClustersOutput{Clusters: f.Clusters}, nil
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
