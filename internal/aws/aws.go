package aws

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/utils"
)

// EcsEnvRequest represents the PUT request body to be sent to kosli from ECS
type EcsEnvRequest struct {
	Artifacts []*EcsTaskData `json:"artifacts"`
}

// EcsTaskData represents the harvested ECS task data
type EcsTaskData struct {
	TaskArn   string            `json:"taskArn"`
	Digests   map[string]string `json:"digests"`
	StartedAt int64             `json:"creationTimestamp"`
}

// S3EnvRequest represents the PUT request body to be sent to kosli from a server
type S3EnvRequest struct {
	Artifacts []*S3Data `json:"artifacts"`
}

// LambdaEnvRequest represents the PUT request body to be sent to kosli from a server
type LambdaEnvRequest struct {
	Artifacts []*LambdaData `json:"artifacts"`
}

// S3Data represents the harvested S3 artifacts data
type S3Data struct {
	Digests               map[string]string `json:"digests"`
	LastModifiedTimestamp int64             `json:"creationTimestamp"`
}

// LambdaData represents the harvested Lambda artifacts data
type LambdaData struct {
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

// AWSStaticCreds represents static creds provided by user
type AWSStaticCreds struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

// GetConfigOptFns returns a slice of config loading options functions based on
// user-provided static creds
func (s *AWSStaticCreds) GetConfigOptFns() []func(*config.LoadOptions) error {
	optFns := []func(*config.LoadOptions) error{}
	if s.Region != "" {
		optFns = append(optFns, config.WithRegion(s.Region))
	}
	if s.AccessKeyID != "" && s.SecretAccessKey != "" {
		optFns = append(optFns, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s.AccessKeyID, s.SecretAccessKey, "")))
	}
	return optFns
}

// NewAWSConfigFromEnvOrFlags returns an AWS config that can be used to construct
// AWS service clients.
// Credentials for config can be sourced from multiple sources, in this order:
// 1) static credentials (from CLI flags or KOSLI env vars), if provided
// 2) AWS Environment variables
// 3) Shared AWS Configuration/Credentials files (see https://docs.aws.amazon.com/sdkref/latest/guide/file-format.html)
// more details can be found here: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials
func (staticCreds *AWSStaticCreds) NewAWSConfigFromEnvOrFlags() (aws.Config, error) {
	return config.LoadDefaultConfig(context.TODO(), staticCreds.GetConfigOptFns()...)
}

// NewS3Client returns a new S3 API client
func (staticCreds *AWSStaticCreds) NewS3Client() (*s3.Client, error) {
	cfg, err := staticCreds.NewAWSConfigFromEnvOrFlags()
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(cfg), nil
}

// NewLambdaClient returns a new Lambda API client
func (staticCreds *AWSStaticCreds) NewLambdaClient() (*lambda.Client, error) {
	cfg, err := staticCreds.NewAWSConfigFromEnvOrFlags()
	if err != nil {
		return nil, err
	}
	return lambda.NewFromConfig(cfg), nil
}

// NewECSClient returns a new ECS API client
func (staticCreds *AWSStaticCreds) NewECSClient() (*ecs.Client, error) {
	cfg, err := staticCreds.NewAWSConfigFromEnvOrFlags()
	if err != nil {
		return nil, err
	}
	return ecs.NewFromConfig(cfg), nil
}

// getAllLambdaFuncs fetches all lambda functions recursively (50 at a time) and returns a list of FunctionConfiguration
func getAllLambdaFuncs(client *lambda.Client, nextMarker *string, allFunctions *[]types.FunctionConfiguration) (*[]types.FunctionConfiguration, error) {
	params := &lambda.ListFunctionsInput{}
	if nextMarker != nil {
		params.Marker = nextMarker
	}

	listFunctionsOutput, err := client.ListFunctions(context.TODO(), params)
	if err != nil {
		return allFunctions, err
	}

	*allFunctions = append(*allFunctions, listFunctionsOutput.Functions...)
	if listFunctionsOutput.NextMarker != nil {
		_, err := getAllLambdaFuncs(client, listFunctionsOutput.NextMarker, allFunctions)
		if err != nil {
			return allFunctions, err
		}
	}
	return allFunctions, nil
}

// GetLambdaPackageData returns a digest and metadata of a Lambda function package
func (staticCreds *AWSStaticCreds) GetLambdaPackageData(functionNames []string) ([]*LambdaData, error) {
	lambdaData := []*LambdaData{}
	client, err := staticCreds.NewLambdaClient()
	if err != nil {
		return lambdaData, err
	}

	if len(functionNames) == 0 {
		allFunctions, err := getAllLambdaFuncs(client, nil, &[]types.FunctionConfiguration{})
		if err != nil {
			return lambdaData, err
		}

		for _, function := range *allFunctions {
			oneLambdaData, err := processOneLambdaFunc(*function.LastModified, *function.CodeSha256, *function.FunctionName, string(function.PackageType))
			if err != nil {
				return lambdaData, err
			}
			lambdaData = append(lambdaData, oneLambdaData)
		}

	} else {
		var (
			wg    sync.WaitGroup
			mutex = &sync.Mutex{}
		)

		// run concurrently
		errs := make(chan error, 1) // Buffered only for the first error
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel() // Make sure it's called to release resources even if no errors

		for _, functionName := range functionNames {
			wg.Add(1)
			go func(functionName string) {
				defer wg.Done()
				// Check if any error occurred in any other gorouties:
				select {
				case <-ctx.Done():
					return // Error somewhere, terminate
				default: // Default is a must to avoid blocking
				}
				oneLambdaData, err := getAndProcessOneLambdaFunc(client, functionName)
				if err != nil {
					// Non-blocking send of error
					select {
					case errs <- err:
					default:
					}
					cancel() // send cancel signal to goroutines
					return
				}

				mutex.Lock()
				lambdaData = append(lambdaData, oneLambdaData)
				mutex.Unlock()

			}(functionName)

		}

		wg.Wait()
		// Return (first) error, if any:
		if ctx.Err() != nil {
			return lambdaData, <-errs
		}
	}

	return lambdaData, nil
}

// getAndProcessOneLambdaFunc get a lambda function by its name and return a LambdaData object from it
func getAndProcessOneLambdaFunc(client *lambda.Client, functionName string) (*LambdaData, error) {
	params := &lambda.GetFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
	}

	function, err := client.GetFunctionConfiguration(context.TODO(), params)
	if err != nil {
		return &LambdaData{}, err
	}

	lambdaData, err := processOneLambdaFunc(*function.LastModified, *function.CodeSha256, *function.FunctionName, string(function.PackageType))
	if err != nil {
		return lambdaData, err
	}

	return lambdaData, nil
}

// processOneLambdaFunc returns LambdaData object from lambda function attributes
func processOneLambdaFunc(lastModified, codeSha256, functionName, packageType string) (*LambdaData, error) {
	lambdaData := &LambdaData{}
	lastModifiedTimestamp, err := formatLambdaLastModified(lastModified)
	if err != nil {
		return lambdaData, err
	}
	lambdaData.LastModifiedTimestamp = lastModifiedTimestamp.Unix()
	lambdaData.Digests = map[string]string{functionName: codeSha256}

	if packageType == "Zip" {
		lambdaData.Digests[functionName], err = decodeLambdaFingerprint(codeSha256)
		if err != nil {
			return lambdaData, err
		}
	}

	return lambdaData, nil
}

// formatLambdaLastModified converts string lastModified to time object
func formatLambdaLastModified(lastModified string) (time.Time, error) {
	layout := "2006-01-02T15:04:05.000+0000"
	return time.Parse(layout, lastModified)
}

// decodeLambdaFingerprint decodes a base64 lambda function fingerprint
func decodeLambdaFingerprint(fingerprint string) (string, error) {
	sha256base64, err := base64.StdEncoding.DecodeString(fingerprint)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sha256base64), nil
}

// GetS3Data returns a digest and metadata of the S3 bucket content
func (staticCreds *AWSStaticCreds) GetS3Data(bucket string, logger *logger.Logger) ([]*S3Data, error) {
	s3Data := []*S3Data{}
	client, err := staticCreds.NewS3Client()
	if err != nil {
		return s3Data, err
	}

	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	}

	objects, err := client.ListObjects(context.TODO(), params)
	if err != nil {
		return s3Data, err
	}
	tempDirName, err := os.MkdirTemp("", "bucketContent")
	if err != nil {
		return s3Data, err
	}
	defer os.RemoveAll(tempDirName)

	downloader := s3manager.NewDownloader(client)
	var lastModifiedTime *time.Time
	for _, object := range objects.Contents {
		err := downloadFileFromBucket(downloader, tempDirName, *object.Key, bucket, logger)
		if lastModifiedTime == nil || object.LastModified.After(*lastModifiedTime) {
			lastModifiedTime = object.LastModified
		}

		if err != nil {
			return s3Data, err
		}
	}

	sha256, err := digest.DirSha256(tempDirName, []string{}, logger)
	if err != nil {
		return s3Data, err
	}
	s3Data = append(s3Data, &S3Data{Digests: map[string]string{bucket: sha256}, LastModifiedTimestamp: lastModifiedTime.Unix()})

	return s3Data, nil
}

func downloadFileFromBucket(downloader *s3manager.Downloader, dirName, key, bucket string, logger *logger.Logger) error {
	file, err := utils.CreateFile(filepath.Join(dirName, key))
	if err != nil {
		return err
	}
	defer file.Close()

	numBytes, err := downloader.Download(context.TODO(), file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		return err
	}
	logger.Debug("downloaded", file.Name(), numBytes, "bytes")

	return nil
}

// GetEcsTasksData returns a list of tasks data for an ECS cluster or service
func (staticCreds *AWSStaticCreds) GetEcsTasksData(cluster string, serviceName string) ([]*EcsTaskData, error) {
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

	client, err := staticCreds.NewECSClient()
	if err != nil {
		return tasksData, err
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
