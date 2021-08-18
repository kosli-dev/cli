package aws

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go/aws"
)

// EcsTaskData represents the harvested ECS task data
type EcsTaskData struct {
	TaskArn   string            `json:"taskArn"`
	Images    map[string]string `json:"images"`
	StartedAt int64             `json:"creationTimestamp"`
}

// NewEcsTaskData creates a NewEcsTaskData object from an ECS task
func NewEcsTaskData(taskArn string, images map[string]string, startedAt time.Time) *EcsTaskData {

	return &EcsTaskData{
		TaskArn:   taskArn,
		Images:    images,
		StartedAt: startedAt.Unix(),
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

// GetEcsTasksData returns a list of tasks data for an ECS cluster or service
func GetEcsTasksData(client *ecs.Client, cluster string, serviceName string) ([]*EcsTaskData, error) {
	listInput := &ecs.ListTasksInput{}
	descriptionInput := &ecs.DescribeTasksInput{}
	tasksData := []*EcsTaskData{}
	if serviceName != "" {
		listInput.ServiceName = aws.String(serviceName)
	}
	if cluster != "" {
		listInput.Cluster = aws.String(cluster)
		descriptionInput.Cluster = aws.String(cluster)
	}

	list, err := client.ListTasks(context.Background(), listInput)
	if err != nil {
		return tasksData, err
	}
	tasks := list.TaskArns

	if len(tasks) > 0 {
		descriptionInput.Tasks = tasks
		result, err := client.DescribeTasks(context.Background(), descriptionInput)
		if err != nil {
			return tasksData, err
		}

		for _, taskDesc := range result.Tasks {
			images := make(map[string]string)
			if *taskDesc.LastStatus == "RUNNING" {
				for _, container := range taskDesc.Containers {
					if container.ImageDigest != nil {
						images[*container.Image] = *container.ImageDigest
					} else if strings.Contains(*container.Image, "@sha256:") {
						images[*container.Image] = strings.Split(*container.Image, "@sha256:")[1]
					} else {
						images[*container.Image] = ""
					}
				}
				data := NewEcsTaskData(*taskDesc.TaskArn, images, *taskDesc.StartedAt)
				tasksData = append(tasksData, data)
			}
		}
	}

	return tasksData, nil
}
