package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go/aws"
)

// PodData represents the harvested ECS service data
type EcsServiceData struct {
	Name              string            `json:"name"`
	Images            map[string]string `json:"images"`
	CreationTimestamp string            `json:"creationTimestamp"`
}

// NewEcsServiceData creates a NewEcsServiceData object from an ECS service
func NewEcsServiceData() *EcsServiceData {
	images := make(map[string]string)

	return &EcsServiceData{
		Name:              "",
		Images:            images,
		CreationTimestamp: "",
	}
}

// NewAWSClient creates an AWS client from:
// 1) Environment Variables
// 2) Shared Configuration
// 3) Shared Credentials files.
func NewAWSClient() (*ecs.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return ecs.NewFromConfig(cfg), nil
}

func ListEcsTasks(client *ecs.Client, cluster string, family string, serviceName string) error {
	var input *ecs.ListTasksInput
	if serviceName != "" {
		input = &ecs.ListTasksInput{
			ServiceName: aws.String(serviceName),
		}
	} else {
		input = &ecs.ListTasksInput{
			Cluster: aws.String(cluster),
			Family:  aws.String(family),
		}
	}

	list, err := client.ListTasks(context.Background(), input)
	if err != nil {
		return err
	}
	tasks := list.TaskArns

	if len(tasks) > 0 {
		result, err := client.DescribeTasks(context.Background(), &ecs.DescribeTasksInput{
			Tasks: tasks,
		})
		if err != nil {
			return err
		}

		for _, taskDesc := range result.Tasks {
			fmt.Println(*taskDesc.Containers[0].Image)
			fmt.Println(taskDesc.Containers[0].ImageDigest)
		}
	}

	return nil
}
