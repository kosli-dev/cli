package aws

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/merkely-development/reporter/internal/digest"
)

// EcsEnvRequest represents the PUT request body to be sent to merkely from ECS
type EcsEnvRequest struct {
	Artifacts []*EcsTaskData `json:"artifacts"`
	Type      string         `json:"type"`
	Id        string         `json:"id"`
}

// EcsTaskData represents the harvested ECS task data
type EcsTaskData struct {
	TaskArn   string            `json:"taskArn"`
	Digests   map[string]string `json:"digests"`
	StartedAt int64             `json:"creationTimestamp"`
}

// NewEcsTaskData creates a NewEcsTaskData object from an ECS task
func NewEcsTaskData(taskArn string, digests map[string]string, startedAt time.Time) *EcsTaskData {

	return &EcsTaskData{
		TaskArn:   taskArn,
		Digests:   digests,
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

// GetS3Digest returns a digest of the S3 bucket content
func GetS3Digest(client *ecs.Client, bucket string) (string, error) {
	os.Setenv("AWS_ACCESS_KEY", "")
	os.Setenv("AWS_SECRET_KEY", "")

	item := "Chart.yaml"

	file, err := os.Create(item)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	sess, _ := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	downloader := s3manager.NewDownloader(sess)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String("k8s-reporter/" + item),
		})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

	sha256, err := digest.FileSha256(item)
	if err != nil {
		return "", err
	} else {
		return sha256, nil
	}
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
			digests := make(map[string]string)
			if *taskDesc.LastStatus == "RUNNING" {
				for _, container := range taskDesc.Containers {
					if container.ImageDigest != nil {
						digests[*container.Image] = strings.TrimPrefix(*container.ImageDigest, "sha256:")
					} else if strings.Contains(*container.Image, "@sha256:") {
						digests[*container.Image] = strings.Split(*container.Image, "@sha256:")[1]
					} else {
						digests[*container.Image] = ""
					}
				}
				data := NewEcsTaskData(*taskDesc.TaskArn, digests, *taskDesc.StartedAt)
				tasksData = append(tasksData, data)
			}
		}
	}

	return tasksData, nil
}
