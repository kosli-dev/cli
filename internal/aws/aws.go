package aws

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
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
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/filters"
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
	Cluster   string            `json:"cluster,omitempty"`
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
func NewEcsTaskData(taskArn, cluster string, digests map[string]string, startedAt time.Time) *EcsTaskData {
	return &EcsTaskData{
		TaskArn: taskArn,
		// Cluster:   cluster,
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

// getFilteredLambdaFuncs fetches a filtered set of lambda functions recursively (50 at a time) and returns a list of FunctionConfiguration
func getFilteredLambdaFuncs(client *lambda.Client, nextMarker *string, allFunctions *[]types.FunctionConfiguration,
	filter *filters.ResourceFilterOptions) (*[]types.FunctionConfiguration, error) {
	params := &lambda.ListFunctionsInput{}
	if nextMarker != nil {
		params.Marker = nextMarker
	}

	listFunctionsOutput, err := client.ListFunctions(context.TODO(), params)
	if err != nil {
		return allFunctions, err
	}

	if len(filter.IncludeNames) == 0 && len(filter.IncludeNamesRegex) == 0 &&
		len(filter.ExcludeNames) == 0 && len(filter.ExcludeNamesRegex) == 0 {
		*allFunctions = append(*allFunctions, listFunctionsOutput.Functions...)
	} else {
		for _, f := range listFunctionsOutput.Functions {
			include, err := filter.ShouldInclude(*f.FunctionName)
			if err != nil {
				return allFunctions, err
			}
			if include {
				*allFunctions = append(*allFunctions, f)
			}
		}
	}

	if listFunctionsOutput.NextMarker != nil {
		_, err := getFilteredLambdaFuncs(client, listFunctionsOutput.NextMarker, allFunctions, filter)
		if err != nil {
			return allFunctions, err
		}
	}
	return allFunctions, nil
}

// GetLambdaPackageData returns a digest and metadata of a Lambda function package
func (staticCreds *AWSStaticCreds) GetLambdaPackageData(filter *filters.ResourceFilterOptions) ([]*LambdaData, error) {
	lambdaData := []*LambdaData{}
	client, err := staticCreds.NewLambdaClient()
	if err != nil {
		return lambdaData, err
	}

	filteredFunctions, err := getFilteredLambdaFuncs(client, nil, &[]types.FunctionConfiguration{}, filter)
	if err != nil {
		return lambdaData, err
	}

	var (
		wg    sync.WaitGroup
		mutex = &sync.Mutex{}
	)

	// run concurrently
	errs := make(chan error, 1) // Buffered only for the first error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Make sure it's called to release resources even if no errors

	for _, function := range *filteredFunctions {
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

		}(*function.FunctionName)

	}

	wg.Wait()
	// Return (first) error, if any:
	if ctx.Err() != nil {
		return lambdaData, <-errs
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

// shouldExcludePath checks if a bucket object should be excluded
func shouldExcludePath(key string, includedPaths, excludedPaths []string) bool {
	if len(includedPaths) > 0 {
		return !objectInPaths(key, includedPaths)
	} else if len(excludedPaths) > 0 {
		return objectInPaths(key, excludedPaths)
	}
	return false
}

// containsSingleFile checks if a path contains only a single file
func containsSingleFile(directoryPath string) (bool, string, error) {
	files, err := os.ReadDir(directoryPath)
	if err != nil {
		return false, "", err
	}

	if len(files) == 1 {
		fileInfo := files[0]

		if fileInfo.IsDir() {
			// If it's a directory, recursively check inside
			subDir := filepath.Join(directoryPath, fileInfo.Name())
			return containsSingleFile(subDir)
		}

		// If it's a file, return information about it
		path := filepath.Join(directoryPath, fileInfo.Name())
		return true, path, nil
	}

	return false, "", nil
}

func objectInPaths(key string, paths []string) bool {
	for _, path := range paths {
		path = strings.TrimLeft(path, "/")
		if strings.HasPrefix(key, path) {
			return true
		}
	}
	return false
}

// GetS3Data returns a digest and metadata of the S3 bucket content
func (staticCreds *AWSStaticCreds) GetS3Data(bucket string, includePaths, excludePaths []string, logger *logger.Logger) ([]*S3Data, error) {
	s3Data := []*S3Data{}

	tempDirName, err := os.MkdirTemp("", "bucketContent")
	if err != nil {
		return s3Data, err
	}
	defer os.RemoveAll(tempDirName)

	client, err := staticCreds.NewS3Client()
	if err != nil {
		return s3Data, err
	}

	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}

	downloader := s3manager.NewDownloader(client)
	var lastModifiedTime *time.Time
	paginator := s3.NewListObjectsV2Paginator(client, params)
	for paginator.HasMorePages() {
		objects, err := paginator.NextPage(context.TODO())
		if err != nil {
			return s3Data, err
		}

		for _, object := range objects.Contents {
			if strings.HasSuffix(*object.Key, "/") { // skip folders
				continue
			}
			if shouldExcludePath(*object.Key, includePaths, excludePaths) { // decide if we should skip
				continue
			}
			err := downloadFileFromBucket(downloader, tempDirName, *object.Key, bucket, logger)
			if err != nil {
				return s3Data, err
			}

			if lastModifiedTime == nil || object.LastModified.After(*lastModifiedTime) {
				lastModifiedTime = object.LastModified
			}
		}
	}

	if lastModifiedTime == nil {
		return s3Data, fmt.Errorf("no matching file or dirs in bucket: [%s]", bucket)
	}

	fileSnapshot, artifactPath, err := containsSingleFile(tempDirName)
	if err != nil {
		return s3Data, err
	}
	var sha256 string
	artifactName := bucket
	if fileSnapshot {
		sha256, err = digest.FileSha256(artifactPath)
		if err != nil {
			return s3Data, err
		}
		artifactName = filepath.Base(artifactPath)
	} else {
		sha256, err = digest.DirSha256(tempDirName, []string{}, logger)
		if err != nil {
			return s3Data, err
		}
	}

	s3Data = append(s3Data, &S3Data{Digests: map[string]string{artifactName: sha256}, LastModifiedTimestamp: lastModifiedTime.Unix()})

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

// getFilteredECSClusters fetches a filtered set of ECS clusters recursively (50 at a time) and returns a list of ecs Clusters
func getFilteredECSClusters(client *ecs.Client, allClusters *[]ecsTypes.Cluster,
	clusterFilter *filters.ResourceFilterOptions, nextToken *string) (*[]ecsTypes.Cluster, error) {
	params := &ecs.ListClustersInput{}
	if nextToken != nil {
		params.NextToken = nextToken
	}

	listClustersOutput, err := client.ListClusters(context.TODO(), params)
	if err != nil {
		return allClusters, err
	}

	describeClustersOutput, err := client.DescribeClusters(context.TODO(), &ecs.DescribeClustersInput{Clusters: listClustersOutput.ClusterArns})
	if err != nil {
		return allClusters, err
	}

	if !clusterFilter.IsSet() {
		*allClusters = append(*allClusters, describeClustersOutput.Clusters...)
	} else {
		for _, c := range describeClustersOutput.Clusters {
			include, err := clusterFilter.ShouldInclude(*c.ClusterName)
			if err != nil {
				return allClusters, err
			}
			if include {
				*allClusters = append(*allClusters, c)
			}
		}
	}

	if listClustersOutput.NextToken != nil {
		_, err := getFilteredECSClusters(client, allClusters, clusterFilter, listClustersOutput.NextToken)
		if err != nil {
			return allClusters, err
		}
	}
	return allClusters, nil
}

// GetEcsTasksData returns a list of tasks data for an ECS cluster or service
func (staticCreds *AWSStaticCreds) GetEcsTasksData(clusterFilter, serviceFilter *filters.ResourceFilterOptions) ([]*EcsTaskData, error) {
	allTasksData := []*EcsTaskData{}
	client, err := staticCreds.NewECSClient()
	if err != nil {
		return allTasksData, err
	}

	filteredClusters, err := getFilteredECSClusters(client, &[]ecsTypes.Cluster{}, clusterFilter, nil)
	if err != nil {
		return allTasksData, err
	}

	var (
		wg    sync.WaitGroup
		mutex sync.Mutex
	)
	errChan := make(chan error, len(*filteredClusters))

	for _, cluster := range *filteredClusters {
		wg.Add(1)
		go func(clusterName string) {
			defer wg.Done()

			filteredServices, err := getFilteredECSServicesInCluster(client, clusterName, &[]ecsTypes.Service{}, serviceFilter, nil)
			if err != nil {
				errChan <- err
				return
			}

			tasksData, err := getTasksDataInClusterService(client, clusterName, filteredServices, nil)
			if err != nil {
				errChan <- err
				return
			}

			// Safely append to shared allTasksData
			mutex.Lock()
			allTasksData = append(allTasksData, tasksData...)
			mutex.Unlock()
		}(*cluster.ClusterName)
	}

	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
		if err != nil {
			return allTasksData, err
		}
	}

	return allTasksData, nil
}

// getFilteredECSServicesInCluster fetches a filtered set of ECS services recursively (100 at a time) and returns a list of ecs Services
func getFilteredECSServicesInCluster(client *ecs.Client, cluster string, allServices *[]ecsTypes.Service, serviceFilter *filters.ResourceFilterOptions, nextToken *string) (*[]ecsTypes.Service, error) {
	listInput := &ecs.ListServicesInput{
		Cluster: aws.String(cluster),
	}
	if nextToken != nil {
		listInput.NextToken = nextToken
	}
	listServicesOutput, err := client.ListServices(context.TODO(), listInput)
	if err != nil {
		return allServices, err
	}

	fmt.Println("listServicesOutput.ServiceArns", listServicesOutput.ServiceArns)

	describeServicesOutput, err := client.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{Cluster: aws.String(cluster), Services: listServicesOutput.ServiceArns})
	if err != nil {
		return allServices, err
	}

	if !serviceFilter.IsSet() {
		*allServices = append(*allServices, describeServicesOutput.Services...)
	} else {
		for _, s := range describeServicesOutput.Services {
			include, err := serviceFilter.ShouldInclude(*s.ServiceName)
			if err != nil {
				return allServices, err
			}
			if include {
				*allServices = append(*allServices, s)
			}
		}
	}

	if listServicesOutput.NextToken != nil {
		_, err := getFilteredECSServicesInCluster(client, cluster, allServices, serviceFilter, listServicesOutput.NextToken)
		if err != nil {
			return allServices, err
		}
	}
	return allServices, nil
}

// getTasksDataInClusterService fetches a filtered set of ECS tasks recursively (100 at a time) and returns a list of ecs Tasks
func getTasksDataInClusterService(client *ecs.Client, clusterName string, filteredServices *[]ecsTypes.Service, nextToken *string) ([]*EcsTaskData, error) {
	tasksData := []*EcsTaskData{}
	var mutex sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(*filteredServices))

	for _, service := range *filteredServices {
		wg.Add(1)
		go func(svc ecsTypes.Service) {
			defer wg.Done()

			listInput := &ecs.ListTasksInput{
				Cluster:     aws.String(clusterName),
				ServiceName: svc.ServiceName,
			}
			if nextToken != nil {
				listInput.NextToken = nextToken
			}
			descriptionInput := &ecs.DescribeTasksInput{
				Cluster: aws.String(clusterName),
			}

			listTasksOutput, err := client.ListTasks(context.Background(), listInput)
			if err != nil {
				errChan <- err
				return
			}
			tasks := listTasksOutput.TaskArns

			if len(tasks) > 0 {
				descriptionInput.Tasks = tasks
				result, err := client.DescribeTasks(context.Background(), descriptionInput)
				if err != nil {
					errChan <- err
					return
				}

				serviceTasksData := []*EcsTaskData{}
				for _, taskDesc := range result.Tasks {
					digests := make(map[string]string)
					if *taskDesc.LastStatus == "RUNNING" {
						for _, container := range taskDesc.Containers {
							imageName := container.Image
							if imageName == nil {
								// some images like AWS Guard Duty don't get an image name from AWS
								// so we default to the container name
								imageName = container.Name
							}
							if container.ImageDigest != nil {
								digests[*imageName] = strings.TrimPrefix(*container.ImageDigest, "sha256:")
							} else if strings.Contains(*imageName, "@sha256:") {
								digests[*imageName] = strings.Split(*imageName, "@sha256:")[1]
							} else {
								digests[*imageName] = ""
							}
						}
						data := NewEcsTaskData(*taskDesc.TaskArn, clusterName, digests, *taskDesc.StartedAt)
						serviceTasksData = append(serviceTasksData, data)
					}
				}

				// Safely append to shared tasksData
				mutex.Lock()
				tasksData = append(tasksData, serviceTasksData...)
				mutex.Unlock()
			}

			// Handle pagination for this service's tasks
			if listTasksOutput.NextToken != nil {
				additionalTasksData, err := getTasksDataInClusterService(client, clusterName, &[]ecsTypes.Service{svc}, listTasksOutput.NextToken)
				if err != nil {
					errChan <- err
					return
				}
				mutex.Lock()
				tasksData = append(tasksData, additionalTasksData...)
				mutex.Unlock()
			}
		}(service)
	}

	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
		if err != nil {
			return tasksData, err
		}
	}

	return tasksData, nil
}
