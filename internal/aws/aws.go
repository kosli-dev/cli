package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go/aws"
)

// EcsTaskData represents the harvested ECS task data
type EcsTaskData struct {
	TaskArn   string            `json:"taskArn"`
	Images    map[string]string `json:"images"`
	StartedAt time.Time         `json:"startedAt"`
}

// NewEcsTaskData creates a NewEcsTaskData object from an ECS task
func NewEcsTaskData(taskArn string, images map[string]string, startedAt time.Time) *EcsTaskData {

	return &EcsTaskData{
		TaskArn:   taskArn,
		Images:    images,
		StartedAt: startedAt,
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

func ListEcsTasks(client *ecs.Client, cluster string, family string, serviceName string) ([]*EcsTaskData, error) {
	var input *ecs.ListTasksInput
	tasksData := []*EcsTaskData{}
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
		return tasksData, err
	}
	tasks := list.TaskArns

	if len(tasks) > 0 {
		result, err := client.DescribeTasks(context.Background(), &ecs.DescribeTasksInput{
			Tasks: tasks,
		})
		if err != nil {
			return tasksData, err
		}

		for _, taskDesc := range result.Tasks {
			images := make(map[string]string)
			for _, container := range taskDesc.Containers {
				if container.ImageDigest != nil {
					images[*container.Image] = *container.ImageDigest
				} else {
					images[*container.Image] = ""
				}
			}
			data := NewEcsTaskData(*taskDesc.TaskArn, images, *taskDesc.StartedAt)
			tasksData = append(tasksData, data)
		}
	}

	return tasksData, nil
}
