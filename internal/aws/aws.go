package aws

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/filters"
	"github.com/kosli-dev/cli/internal/logger"
	"golang.org/x/sync/semaphore"
)

// EcsEnvRequest represents the PUT request body to be sent to kosli from ECS
type EcsEnvRequest struct {
	Artifacts []*EcsTaskData `json:"artifacts"`
}

// EcsTaskData represents the harvested ECS task data
type EcsTaskData struct {
	TaskArn   string            `json:"taskArn"`
	Cluster   string            `json:"cluster_name,omitempty"`
	Service   string            `json:"service_name,omitempty"`
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
func NewEcsTaskData(taskArn, clusterName, serviceName string, digests map[string]string, startedAt time.Time) *EcsTaskData {
	return &EcsTaskData{
		TaskArn:   taskArn,
		Cluster:   clusterName,
		Service:   serviceName,
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
//
// Credentials are sourced in this order:
// 1) static credentials (from CLI flags or KOSLI env vars), if provided
// 2) AWS environment variables
// 3) shared AWS configuration/credentials files (see https://docs.aws.amazon.com/sdkref/latest/guide/file-format.html)
//
// Retry: uses adaptive mode (up to MaxAttempts attempts) with a shared in-memory token
// bucket. Commands like "kosli snapshot lambda" fetch functions concurrently
// (see GetLambdaPackageData), which can trigger AWS rate limits (HTTP 429).
// The shared token bucket slows down the entire batch of goroutines when
// throttling is detected, rather than each goroutine retrying independently.
//
// More details: https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry
func (staticCreds *AWSStaticCreds) NewAWSConfigFromEnvOrFlags() (aws.Config, error) {
	optFns := staticCreds.GetConfigOptFns()
	optFns = append(optFns, config.WithRetryer(func() aws.Retryer {
		return retry.NewAdaptiveMode(func(o *retry.AdaptiveModeOptions) {
			o.StandardOptions = append(o.StandardOptions, func(so *retry.StandardOptions) {
				so.MaxAttempts = 10
			})
		})
	}))
	return config.LoadDefaultConfig(context.TODO(), optFns...)
}

// NewS3Client returns a new S3 API client
func (staticCreds *AWSStaticCreds) NewS3Client() (*s3.Client, error) {
	cfg, err := staticCreds.NewAWSConfigFromEnvOrFlags()
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(cfg), nil
}

// S3ListAPI is the S3 listing operation used by this package. The real
// *s3.Client satisfies this implicitly, and the interface also satisfies
// s3.ListObjectsV2APIClient so it can drive the SDK paginator.
type S3ListAPI interface {
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

// S3DownloadAPI downloads a single object from a bucket. The real
// *transfermanager.Client satisfies this implicitly.
//
// This is the transfer manager's own operation rather than a raw S3 API call.
// Faking at this level means a fake writes object bytes straight to the
// WriterAt instead of having to reimplement the transfer manager's ranged
// GetObject/HeadObject machinery.
type S3DownloadAPI interface {
	DownloadObject(ctx context.Context, params *transfermanager.DownloadObjectInput, optFns ...func(*transfermanager.Options)) (*transfermanager.DownloadObjectOutput, error)
}

// S3HeadAPI reads an object's metadata without reading the object itself,
// including the checksum S3 stores for it. The real *s3.Client satisfies this
// implicitly.
//
// The stored checksum is only returned when the request sets ChecksumMode to
// ChecksumModeEnabled.
type S3HeadAPI interface {
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}

// S3API is the combined S3 surface that GetS3Data depends on.
type S3API interface {
	S3ListAPI
	S3DownloadAPI
	S3HeadAPI
}

// s3Client combines the real AWS clients that back S3API: *s3.Client for
// listing and metadata, and *transfermanager.Client for downloading.
type s3Client struct {
	S3ListAPI
	S3DownloadAPI
	S3HeadAPI
}

// defaultNewS3Client creates a real S3 client from credentials.
func defaultNewS3Client(creds *AWSStaticCreds) (S3API, error) {
	client, err := creds.NewS3Client()
	if err != nil {
		return nil, err
	}
	// Five parts per object is the SDK's default, pinned so the connection count,
	// objects in flight times parts, cannot move with an SDK upgrade.
	return &s3Client{
		S3ListAPI: client,
		S3DownloadAPI: transfermanager.New(client, func(o *transfermanager.Options) {
			o.Concurrency = 5
		}),
		S3HeadAPI: client,
	}, nil
}

// NewS3ClientFunc is the factory used by GetS3Data to create an S3API client.
// Tests can replace this to inject a FakeS3Client.
var NewS3ClientFunc = defaultNewS3Client

// ResetS3ClientFactory restores the default (real AWS) client factory.
func ResetS3ClientFactory() {
	NewS3ClientFunc = defaultNewS3Client
}

// NewLambdaClient returns a new Lambda API client
func (staticCreds *AWSStaticCreds) NewLambdaClient() (*lambda.Client, error) {
	cfg, err := staticCreds.NewAWSConfigFromEnvOrFlags()
	if err != nil {
		return nil, err
	}
	return lambda.NewFromConfig(cfg), nil
}

// LambdaAPI is the interface for AWS Lambda operations used by this package.
// The real *lambda.Client satisfies this implicitly.
type LambdaAPI interface {
	ListFunctions(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error)
	GetFunctionConfiguration(ctx context.Context, params *lambda.GetFunctionConfigurationInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionConfigurationOutput, error)
}

// defaultNewLambdaClient creates a real Lambda client from credentials.
func defaultNewLambdaClient(creds *AWSStaticCreds) (LambdaAPI, error) {
	return creds.NewLambdaClient()
}

// NewLambdaClientFunc is the factory used by GetLambdaPackageData to create a
// LambdaAPI client. Tests can replace this to inject a FakeLambdaClient.
var NewLambdaClientFunc = defaultNewLambdaClient

// ResetLambdaClientFactory restores the default (real AWS) client factory.
func ResetLambdaClientFactory() {
	NewLambdaClientFunc = defaultNewLambdaClient
}

// ECSServicesAPI is the subset of AWS ECS operations used when listing and
// describing services in a cluster. The real *ecs.Client satisfies this
// implicitly.
type ECSServicesAPI interface {
	ListServices(ctx context.Context, params *ecs.ListServicesInput, optFns ...func(*ecs.Options)) (*ecs.ListServicesOutput, error)
	DescribeServices(ctx context.Context, params *ecs.DescribeServicesInput, optFns ...func(*ecs.Options)) (*ecs.DescribeServicesOutput, error)
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
// filter is pre-compiled by the caller, so its regex patterns are not re-compiled per function or per page
func getFilteredLambdaFuncs(client LambdaAPI, nextMarker *string, allFunctions *[]types.FunctionConfiguration,
	filter *filters.CompiledResourceFilter) (*[]types.FunctionConfiguration, error) {
	params := &lambda.ListFunctionsInput{}
	if nextMarker != nil {
		params.Marker = nextMarker
	}

	listFunctionsOutput, err := client.ListFunctions(context.TODO(), params)
	if err != nil {
		return allFunctions, err
	}

	if !filter.IsSet() {
		*allFunctions = append(*allFunctions, listFunctionsOutput.Functions...)
	} else {
		for _, f := range listFunctionsOutput.Functions {
			included, err := filter.ShouldInclude(*f.FunctionName)
			if err != nil {
				return allFunctions, err
			}
			if included {
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
	client, err := NewLambdaClientFunc(staticCreds)
	if err != nil {
		return []*LambdaData{}, err
	}
	return getLambdaPackageDataFromClient(client, filter)
}

// getLambdaPackageDataFromClient fetches Lambda function data using the provided LambdaAPI client.
func getLambdaPackageDataFromClient(client LambdaAPI, filter *filters.ResourceFilterOptions) ([]*LambdaData, error) {
	lambdaData := []*LambdaData{}

	// compile the filter patterns once, instead of once per function and per page
	compiledFilter := filter.Compile()

	filteredFunctions, err := getFilteredLambdaFuncs(client, nil, &[]types.FunctionConfiguration{}, compiledFilter)
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
func getAndProcessOneLambdaFunc(client LambdaAPI, functionName string) (*LambdaData, error) {
	params := &lambda.GetFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
	}

	function, err := client.GetFunctionConfiguration(context.TODO(), params)
	if err != nil {
		return &LambdaData{}, err
	}

	lambdaData, err := processOneLambdaFunc(aws.ToString(function.LastModified), aws.ToString(function.CodeSha256), aws.ToString(function.FunctionName), function.PackageType)
	if err != nil {
		return lambdaData, err
	}

	return lambdaData, nil
}

// processOneLambdaFunc returns LambdaData object from lambda function attributes
func processOneLambdaFunc(lastModified, codeSha256, functionName string, packageType types.PackageType) (*LambdaData, error) {
	lambdaData := &LambdaData{}
	lastModifiedTimestamp, err := formatLambdaLastModified(lastModified)
	if err != nil {
		return lambdaData, err
	}
	lambdaData.LastModifiedTimestamp = lastModifiedTimestamp.Unix()
	lambdaData.Digests = map[string]string{functionName: codeSha256}

	if packageType == types.PackageTypeZip {
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

// shouldExcludePath checks if a bucket object should be excluded.
// Paths in includedPaths/excludedPaths match by literal prefix.
// includedRegex/excludedRegex are pre-compiled regular expressions
// matched against the full object key.
func shouldExcludePath(key string, includedPaths []string, includedRegex []*regexp.Regexp, excludedPaths []string, excludedRegex []*regexp.Regexp) bool {
	if len(includedPaths) > 0 || len(includedRegex) > 0 {
		return !objectMatchesFilter(key, includedPaths, includedRegex)
	}
	if len(excludedPaths) > 0 || len(excludedRegex) > 0 {
		return objectMatchesFilter(key, excludedPaths, excludedRegex)
	}
	return false
}

// compilePathRegex pre-compiles a list of path regex patterns so the result
// can be reused across many object keys without re-compiling per iteration.
func compilePathRegex(patterns []string) ([]*regexp.Regexp, error) {
	if len(patterns) == 0 {
		return nil, nil
	}
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid path regex pattern %q: %v", pattern, err)
		}
		compiled = append(compiled, re)
	}
	return compiled, nil
}

// objectMatchesFilter reports whether key matches any of the filter entries.
// A key matches when it is prefixed by one of paths (literal prefix match)
// or when one of patterns matches the full key.
func objectMatchesFilter(key string, paths []string, patterns []*regexp.Regexp) bool {
	for _, prefix := range paths {
		prefix = strings.TrimLeft(prefix, "/")
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	for _, re := range patterns {
		if re.MatchString(key) {
			return true
		}
	}
	return false
}

// GetS3Data returns a digest and metadata of the S3 bucket content.
// includePaths / excludePaths match object keys by literal prefix.
// includeRegex / excludeRegex match object keys by Go regular expression.
// Include and exclude filters are mutually exclusive (callers enforce this).
func (staticCreds *AWSStaticCreds) GetS3Data(bucket string, includePaths, includeRegex, excludePaths, excludeRegex []string, limits DownloadLimits, logger *logger.Logger) ([]*S3Data, error) {
	client, err := NewS3ClientFunc(staticCreds)
	if err != nil {
		return []*S3Data{}, err
	}
	return getS3DataFromClient(client, bucket, includePaths, includeRegex, excludePaths, excludeRegex, limits, logger)
}

// getS3DataFromClient harvests bucket content using the provided S3API client.
func getS3DataFromClient(client S3API, bucket string, includePaths, includeRegex, excludePaths, excludeRegex []string, limits DownloadLimits, logger *logger.Logger) ([]*S3Data, error) {
	s3Data := []*S3Data{}

	includeRegexCompiled, err := compilePathRegex(includeRegex)
	if err != nil {
		return s3Data, err
	}
	excludeRegexCompiled, err := compilePathRegex(excludeRegex)
	if err != nil {
		return s3Data, err
	}

	objects, err := listMatchingS3Objects(client, bucket, includePaths, includeRegexCompiled, excludePaths, excludeRegexCompiled)
	if err != nil {
		return s3Data, err
	}
	if len(objects) == 0 {
		return s3Data, fmt.Errorf("no matching file or dirs in bucket: [%s]", bucket)
	}

	newest := objects[0].lastModified
	for _, object := range objects {
		if object.lastModified.After(newest) {
			newest = object.lastModified
		}
	}
	if newest.IsZero() {
		return s3Data, fmt.Errorf("bucket [%s] reported no modification time for any matching object", bucket)
	}

	artifactName, sha256, err := fingerprintS3Objects(client, bucket, objects, limits, logger)
	if err != nil {
		return s3Data, err
	}

	s3Data = append(s3Data, &S3Data{Digests: map[string]string{artifactName: sha256}, LastModifiedTimestamp: newest.Unix()})

	return s3Data, nil
}

// s3Object is one listed object that survived the include and exclude filters.
type s3Object struct {
	key          string
	lastModified time.Time
	size         int64
}

// DownloadLimits bounds the object downloads in flight at once when
// fingerprinting a bucket.
type DownloadLimits struct {
	// Concurrency is the number of objects downloading at the same time. Each
	// one may buffer up to five 8 MiB parts in memory while it writes, so memory
	// rises with this figure independently of BytesInFlight.
	Concurrency int
	// BytesInFlight caps the sum of the listed sizes of the objects downloading
	// at the same time, and so the temp disk they occupy. An object larger than
	// the whole budget downloads alone.
	BytesInFlight int64
}

// DefaultDownloadLimits keeps peak temp disk near half a gigabyte, which fits
// Lambda's default /tmp, and part buffers near 320 MiB of memory.
var DefaultDownloadLimits = DownloadLimits{Concurrency: 8, BytesInFlight: 512 << 20}

// listMatchingS3Objects lists the bucket, dropping folder markers and keys the
// filters exclude, in the order S3 returns them.
func listMatchingS3Objects(client S3ListAPI, bucket string, includePaths []string, includeRegex []*regexp.Regexp,
	excludePaths []string, excludeRegex []*regexp.Regexp) ([]s3Object, error) {
	objects := []s3Object{}
	seen := map[string]bool{}
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		for _, object := range page.Contents {
			// Real S3 always sets both fields; S3-compatible stores may not.
			// Dropping an entry with no key would lose an object silently.
			if object.Key == nil {
				return nil, fmt.Errorf("bucket [%s] listed an object with no key", bucket)
			}
			if strings.HasSuffix(*object.Key, "/") { // skip folders
				continue
			}
			if shouldExcludePath(*object.Key, includePaths, includeRegex, excludePaths, excludeRegex) {
				continue
			}
			// A key listed twice is a listing fault, not a collision between two
			// keys, and the collision report relies on keys being distinct. Only
			// keys that reach the fingerprint are checked, so memory stays bounded
			// by the filtered set rather than the whole bucket.
			if seen[*object.Key] {
				return nil, fmt.Errorf("bucket [%s] listed object key [%s] more than once", bucket, *object.Key)
			}
			seen[*object.Key] = true
			// An object without a timestamp stays in the fingerprint and out of
			// the snapshot timestamp, as it was before.
			listed := s3Object{key: *object.Key}
			if object.LastModified != nil {
				listed.lastModified = *object.LastModified
			}
			if object.Size != nil {
				listed.size = *object.Size
			}
			objects = append(objects, listed)
		}
	}
	return objects, nil
}

// s3DigestSource is where the fingerprint pipeline gets each object's content
// sha256 once the tree is known. Content mode downloads the object into the
// pipeline's temp dir and hashes it; a source that reads S3's stored checksum
// never touches the disk. Everything else -- key rule, .kosli_ignore, the tree
// walk -- is shared, so the two sources cannot fingerprint the same bucket
// differently.
type s3DigestSource struct {
	// sha256 returns the hex digest of one object's content. tempDir is scratch
	// space the pipeline owns and removes when it is done.
	sha256 func(ctx context.Context, tempDir string, object s3Object) (string, error)
	// usesDisk reports whether an object's listed size occupies temp disk while
	// sha256 runs, and so counts against DownloadLimits.BytesInFlight.
	usesDisk bool
}

// downloadDigests is the content-mode source: download, hash, remove.
func downloadDigests(downloader S3DownloadAPI, bucket string, logger *logger.Logger) s3DigestSource {
	return s3DigestSource{
		sha256: func(ctx context.Context, tempDir string, object s3Object) (string, error) {
			return downloadAndHashS3Object(ctx, downloader, tempDir, bucket, object.key, nil, logger)
		},
		usesDisk: true,
	}
}

// fingerprintS3Objects fingerprints the objects as the directory their keys
// describe, downloading each one to an anonymous temp file, hashing it and
// removing it. See fingerprintS3Tree for the pipeline.
func fingerprintS3Objects(downloader S3DownloadAPI, bucket string, objects []s3Object, limits DownloadLimits, logger *logger.Logger) (string, string, error) {
	return fingerprintS3Tree(downloader, downloadDigests(downloader, bucket, logger), bucket, objects, limits, logger)
}

// fingerprintS3Tree fingerprints the objects as the directory their keys
// describe, without ever using a key as a local file name. Each object's
// content sha256 comes from source; the fingerprint is then computed from the
// (key, sha256) pairs by digest.VirtualDirSha256, which reproduces what
// digest.DirSha256 gives the same tree on disk. A single object is
// fingerprinted as that file and named after it, as before.
//
// A root .kosli_ignore is always downloaded first, whatever the source, because
// its rules decide which other objects take part; objects the rules exclude are
// not fetched at all. The remaining objects are fetched in parallel within
// limits, and the first failure cancels the rest.
func fingerprintS3Tree(downloader S3DownloadAPI, source s3DigestSource, bucket string, objects []s3Object, limits DownloadLimits, logger *logger.Logger) (string, string, error) {
	keys := make([]string, len(objects))
	for i, object := range objects {
		keys[i] = object.key
	}
	paths, err := virtualPathsForS3Keys(keys)
	if err != nil {
		return "", "", err
	}

	tempDir, err := os.MkdirTemp("", "bucketContent")
	if err != nil {
		return "", "", err
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			logger.Warn("failed to remove temp dir %s: %v", tempDir, err)
		}
	}()

	// The manifest starts as paths only; digests are filled in by index below.
	files := make([]digest.VirtualFile, len(objects))
	for i, object := range objects {
		files[i].Path = paths[object.key]
	}

	// One object is fingerprinted as that file and named after it, as it was
	// when the objects were laid out on disk.
	if file, ok := digest.SingleVirtualFile(files); ok {
		sha256, err := source.sha256(context.TODO(), tempDir, objects[0])
		if err != nil {
			return "", "", err
		}
		return file.Name(), sha256, nil
	}

	var rules []string
	contentSha256 := map[string]string{}
	for _, key := range keys {
		if paths[key] != digest.IgnoreFileName {
			continue
		}
		sha256, err := downloadAndHashS3Object(context.TODO(), downloader, tempDir, bucket, key, func(file *os.File) error {
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				return err
			}
			parsed, err := digest.ParseIgnoreRules(file)
			if err != nil {
				return err
			}
			rules = parsed
			return nil
		}, logger)
		if err != nil {
			return "", "", err
		}
		contentSha256[key] = sha256
		logger.Debug("object key [%s] is the bucket's %s -- excluding paths: %s", key, digest.IgnoreFileName, rules)
	}

	needed, err := digest.FilesNeedingContent(files, rules)
	if err != nil {
		return "", "", ignoreRuleError(err)
	}

	var toDownload []int
	for i, object := range objects {
		sha256, downloaded := contentSha256[object.key]
		switch {
		case downloaded: // the ignore-file pass already hashed this object
		case needed[files[i].Path]:
			toDownload = append(toDownload, i)
		default:
			logger.Debug("object key [%s] is excluded by %s and is not downloaded", object.key, digest.IgnoreFileName)
		}
		files[i].Sha256 = sha256
	}

	// Each fetch writes its own slot, so the manifest stays in listing order
	// however the fetches interleave.
	if err := fetchS3DigestsInParallel(source, tempDir, objects, toDownload, files, limits, logger); err != nil {
		return "", "", err
	}

	sha256, err := digest.VirtualDirSha256(files, rules, logger)
	if err != nil {
		return "", "", ignoreRuleError(err)
	}
	return bucket, sha256, nil
}

// ignoreRuleError names the bucket's ignore file when one of its rules cannot
// be applied, and leaves any other failure as it is.
func ignoreRuleError(err error) error {
	if errors.Is(err, path.ErrBadPattern) {
		return fmt.Errorf("the bucket's %s holds a rule that cannot be applied: %w", digest.IgnoreFileName, err)
	}
	return err
}

// fetchS3DigestsInParallel reads the digests of the objects at indexes from
// source and writes each into files at the same index. A fixed worker pool
// bounds fetches and goroutines alike, a weighted semaphore bounds the listed
// bytes of sources that use the disk, and the first error cancels the context
// so nothing further starts.
func fetchS3DigestsInParallel(source s3DigestSource, tempDir string, objects []s3Object, indexes []int,
	files []digest.VirtualFile, limits DownloadLimits, logger *logger.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	budget := semaphore.NewWeighted(max(limits.BytesInFlight, 1))
	firstErr := make(chan error, 1)
	fail := func(err error) {
		select {
		case firstErr <- err:
		default: // an earlier failure is already recorded
		}
		cancel()
	}

	work := make(chan int)
	var wg sync.WaitGroup
	for range min(max(limits.Concurrency, 1), len(indexes)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range work {
				object := objects[i]
				// An object larger than the budget takes all of it and so runs alone.
				// A source that never touches the disk owes the budget nothing.
				var weight int64
				if source.usesDisk {
					weight = max(min(object.size, limits.BytesInFlight), 1)
					if err := budget.Acquire(ctx, weight); err != nil {
						return // cancelled while waiting
					}
				}
				sha256, err := source.sha256(ctx, tempDir, object)
				budget.Release(weight)
				if err != nil {
					fail(err)
					return
				}
				files[i].Sha256 = sha256
			}
		}()
	}

feed:
	for _, i := range indexes {
		select {
		case work <- i:
		case <-ctx.Done():
			break feed
		}
	}
	close(work)
	wg.Wait()

	select {
	case err := <-firstErr:
		return err
	default:
		return nil
	}
}

// downloadAndHashS3Object fetches one object into a fresh temp file, lets
// inspect read it when given, returns the sha256 of its content and removes the
// file. The file's name comes from the OS, so nothing about the key reaches the
// filesystem.
func downloadAndHashS3Object(ctx context.Context, downloader S3DownloadAPI, tempDir, bucket, key string, inspect func(*os.File) error, logger *logger.Logger) (string, error) {
	file, err := os.CreateTemp(tempDir, "object-*")
	if err != nil {
		return "", fmt.Errorf("object key [%s]: %w", key, err)
	}
	defer func() {
		// Close before remove: Windows will not delete an open file.
		if err := file.Close(); err != nil {
			logger.Warn("failed to close temp file for object key [%s]: %v", key, err)
		}
		if err := os.Remove(file.Name()); err != nil {
			logger.Warn("failed to remove temp file for object key [%s]: %v", key, err)
		}
	}()

	result, err := downloader.DownloadObject(ctx, &transfermanager.DownloadObjectInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		WriterAt: file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to download object key [%s]: %w", key, err)
	}
	if result.ContentLength != nil {
		logger.Debug("downloaded object key [%s]: %d bytes", key, *result.ContentLength)
	}

	if inspect != nil {
		if err := inspect(file); err != nil {
			return "", fmt.Errorf("object key [%s]: %w", key, err)
		}
	}
	sha256, err := digest.FileSha256(file.Name(), logger)
	if err != nil {
		return "", fmt.Errorf("failed to hash object key [%s]: %w", key, err)
	}
	return sha256, nil
}

// getFilteredECSClusters fetches a filtered set of ECS clusters recursively (50 at a time) and returns a list of ecs Clusters
// clusterFilter is pre-compiled by the caller, so its regex patterns are not re-compiled per cluster or per page
func getFilteredECSClusters(client *ecs.Client, allClusters *[]ecsTypes.Cluster,
	clusterFilter *filters.CompiledResourceFilter, nextToken *string, logger *logger.Logger) (*[]ecsTypes.Cluster, error) {
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
		logger.Info("all ECS clusters in the AWS account will be scanned")
		*allClusters = append(*allClusters, describeClustersOutput.Clusters...)
	} else {
		for _, c := range describeClustersOutput.Clusters {
			included, err := clusterFilter.ShouldInclude(*c.ClusterName)
			if err != nil {
				return allClusters, err
			}
			if included {
				*allClusters = append(*allClusters, c)
			}
		}
	}

	if listClustersOutput.NextToken != nil {
		_, err := getFilteredECSClusters(client, allClusters, clusterFilter, listClustersOutput.NextToken, logger)
		if err != nil {
			return allClusters, err
		}
	}
	if clusterFilter.IsSet() {
		clusterNames := make([]string, len(*allClusters))
		for i, cluster := range *allClusters {
			clusterNames[i] = *cluster.ClusterName
		}
		logger.Info("the following ECS clusters will be scanned: %v", clusterNames)
	}
	return allClusters, nil
}

// GetEcsTasksData returns a list of tasks data for an ECS cluster or service
func (staticCreds *AWSStaticCreds) GetEcsTasksData(clusterFilter, serviceFilter *filters.ResourceFilterOptions, logger *logger.Logger) ([]*EcsTaskData, error) {
	allTasksData := []*EcsTaskData{}
	client, err := staticCreds.NewECSClient()
	if err != nil {
		return allTasksData, fmt.Errorf("failed to create ECS client: %w", err)
	}

	// compile the filter patterns once, instead of once per cluster and per service
	compiledClusterFilter := clusterFilter.Compile()
	compiledServiceFilter := serviceFilter.Compile()

	filteredClusters, err := getFilteredECSClusters(client, &[]ecsTypes.Cluster{}, compiledClusterFilter, nil, logger)
	if err != nil {
		return allTasksData, fmt.Errorf("failed to filter ECS clusters: %w", err)
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

			filteredServices, err := getFilteredECSServicesInCluster(client, clusterName, &[]ecsTypes.Service{}, compiledServiceFilter, nil, logger)
			if err != nil {
				errChan <- fmt.Errorf("failed to filter ECS services in cluster %s: %w", clusterName, err)
				return
			}

			tasksData, err := getTasksDataInClusterService(client, clusterName, filteredServices, nil, logger)
			if err != nil {
				errChan <- fmt.Errorf("failed to get tasks data in cluster %s: %w", clusterName, err)
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

// getFilteredECSServicesInCluster fetches a filtered set of ECS services recursively (10 at a time) and returns a list of ecs Services
// serviceFilter is pre-compiled by the caller, so its regex patterns are not re-compiled per service, per page or per cluster
func getFilteredECSServicesInCluster(client ECSServicesAPI, cluster string, allServices *[]ecsTypes.Service, serviceFilter *filters.CompiledResourceFilter,
	nextToken *string, logger *logger.Logger) (*[]ecsTypes.Service, error) {
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

	// A cluster with no services yields an empty ServiceArns list. The AWS ECS
	// DescribeServices API rejects an empty Services list with
	// "InvalidParameterException: Services cannot be empty", so skip the call
	// and let the cluster contribute zero services.
	if len(listServicesOutput.ServiceArns) == 0 {
		return allServices, nil
	}

	describeServicesOutput, err := client.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{Cluster: aws.String(cluster), Services: listServicesOutput.ServiceArns})
	if err != nil {
		return allServices, err
	}

	if !serviceFilter.IsSet() {
		logger.Info("all ECS services in cluster [%s] will be scanned", cluster)
		*allServices = append(*allServices, describeServicesOutput.Services...)
	} else {
		for _, s := range describeServicesOutput.Services {
			included, err := serviceFilter.ShouldInclude(*s.ServiceName)
			if err != nil {
				return allServices, err
			}
			if included {
				*allServices = append(*allServices, s)
			}
		}
	}

	if listServicesOutput.NextToken != nil {
		_, err := getFilteredECSServicesInCluster(client, cluster, allServices, serviceFilter, listServicesOutput.NextToken, logger)
		if err != nil {
			return allServices, err
		}
	}
	if serviceFilter.IsSet() {
		serviceNames := make([]string, len(*allServices))
		for i, service := range *allServices {
			serviceNames[i] = *service.ServiceName
		}
		logger.Info("the following ECS services in cluster [%s] will be scanned: %v", cluster, serviceNames)
	}
	return allServices, nil
}

// getTasksDataInClusterService fetches a filtered set of ECS tasks recursively (100 at a time) and returns a list of ecs Tasks
func getTasksDataInClusterService(client *ecs.Client, clusterName string, filteredServices *[]ecsTypes.Service, nextToken *string, logger *logger.Logger) ([]*EcsTaskData, error) {
	tasksData := []*EcsTaskData{}
	var mutex sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(*filteredServices))

	for _, service := range *filteredServices {
		wg.Add(1)
		go func(svc ecsTypes.Service) {
			defer wg.Done()

			logger.Debug("scanning ECS tasks in service [%s] in cluster [%s]", *svc.ServiceName, clusterName)
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
					if aws.ToString(taskDesc.LastStatus) == "RUNNING" {
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
						data := NewEcsTaskData(*taskDesc.TaskArn, clusterName, *svc.ServiceName, digests, *taskDesc.StartedAt)
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
				additionalTasksData, err := getTasksDataInClusterService(client, clusterName, &[]ecsTypes.Service{svc}, listTasksOutput.NextToken, logger)
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
