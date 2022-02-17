package aws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/merkely-development/reporter/internal/digest"
	"github.com/merkely-development/reporter/internal/utils"
	"github.com/sirupsen/logrus"
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

// S3EnvRequest represents the PUT request body to be sent to merkely from a server
type S3EnvRequest struct {
	Artifacts []*S3Data `json:"artifacts"`
	Type      string    `json:"type"`
	Id        string    `json:"id"`
}

// S3Data represents the harvested server artifacts data
type S3Data struct {
	Digests               map[string]string `json:"digests"`
	LastModifiedTimestamp int64             `json:"creationTimestamp"`
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

func AWSCredentials(id, secret string) *credentials.Credentials {
	creds := credentials.NewEnvCredentials()
	if _, err := creds.Get(); err != nil {
		creds = credentials.NewStaticCredentials(id, secret, "")
	}
	return creds
}

// GetS3Data returns a digest of the S3 bucket content
func GetS3Data(bucket string, creds *credentials.Credentials, region string) ([]*S3Data, error) {
	s3Data := []*S3Data{}
	awsConfig := &aws.Config{Credentials: creds, Region: aws.String(region)}
	s3Session, err := session.NewSession(awsConfig)
	if err != nil {
		return s3Data, err
	}

	svc := s3.New(s3Session)
	input := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	}

	result, err := svc.ListObjects(input)
	if err != nil {
		return s3Data, err
	}
	tempDirName, err := os.MkdirTemp(".", "bucketContent")
	if err != nil {
		return s3Data, err
	}
	defer os.RemoveAll(tempDirName)

	downloaderSession, err := session.NewSession(awsConfig)
	if err != nil {
		return s3Data, err
	}
	downloader := s3manager.NewDownloader(downloaderSession)
	lastModifiedTime := result.Contents[0].LastModified
	for _, object := range result.Contents {
		err := downloadFileFromBucket(downloader, tempDirName, *object.Key, bucket)
		if object.LastModified.After(*lastModifiedTime) {
			lastModifiedTime = object.LastModified
		}
		if err != nil {
			return s3Data, err
		}
	}

	sha256, err := digest.DirSha256(tempDirName, logrus.New())
	if err != nil {
		return s3Data, err
	}
	s3Data = append(s3Data, &S3Data{Digests: map[string]string{bucket: sha256}, LastModifiedTimestamp: lastModifiedTime.Unix()})

	return s3Data, nil
}

func downloadFileFromBucket(downloader *s3manager.Downloader, dirName, key, bucket string) error {
	file, err := utils.CreateFile(filepath.Join(dirName, key))
	if err != nil {
		return err
	}
	defer file.Close()

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		return err
	}
	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

	return nil
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
