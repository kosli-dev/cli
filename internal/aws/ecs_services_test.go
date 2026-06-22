package aws

import (
	"testing"

	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/kosli-dev/cli/internal/filters"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
)

// TestGetFilteredECSServicesInCluster_EmptyCluster reproduces the regression
// where a cluster with no services caused DescribeServices to be called with an
// empty Services list, which the real AWS API rejects with
// "InvalidParameterException: Services cannot be empty".
//
// An empty cluster should simply contribute zero services without error.
func TestGetFilteredECSServicesInCluster_EmptyCluster(t *testing.T) {
	client := &FakeECSClient{
		ServiceArns: []string{}, // cluster has no services
		Services:    []ecsTypes.Service{},
	}

	allServices, err := getFilteredECSServicesInCluster(
		client,
		"empty-cluster",
		&[]ecsTypes.Service{},
		&filters.ResourceFilterOptions{},
		nil,
		logger.NewStandardLogger(),
	)

	require.NoError(t, err)
	require.Empty(t, *allServices)
}

// TestGetFilteredECSServicesInCluster_WithServices verifies the happy path is
// unchanged by the empty-cluster guard: a cluster with services returns them.
func TestGetFilteredECSServicesInCluster_WithServices(t *testing.T) {
	svcName := "my-service"
	client := &FakeECSClient{
		ServiceArns: []string{"arn:aws:ecs:eu-central-1:123:service/cluster/my-service"},
		Services:    []ecsTypes.Service{{ServiceName: &svcName}},
	}

	allServices, err := getFilteredECSServicesInCluster(
		client,
		"cluster",
		&[]ecsTypes.Service{},
		&filters.ResourceFilterOptions{},
		nil,
		logger.NewStandardLogger(),
	)

	require.NoError(t, err)
	require.Len(t, *allServices, 1)
	require.Equal(t, svcName, *(*allServices)[0].ServiceName)
}
