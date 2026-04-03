package aws

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/kosli-dev/cli/internal/filters"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AWSTestSuite struct {
	suite.Suite
}

func (suite *AWSTestSuite) TestFormatLambdaLastModified() {
	for _, t := range []struct {
		name         string
		lastModified string
		want         time.Time
		wantErr      bool
	}{
		{
			name:         "valid last modified get converted",
			lastModified: "2023-01-22T15:04:05.000+0000",
		},
		{
			name:         "invalid format causes an error",
			lastModified: "2023-01-22",
			wantErr:      true,
		},
	} {
		suite.Run(t.name, func() {
			got, err := formatLambdaLastModified(t.lastModified)
			require.False(suite.T(), (err != nil) != t.wantErr,
				"formatLambdaLastModified() error = %v, wantErr %v", err, t.wantErr)
			if !t.wantErr {
				date, err := time.Parse("2006-01-02T15:04:05.000+0000", t.lastModified)
				require.NoError(suite.T(), err)
				require.Equal(suite.T(), date, got)
			}
		})
	}
}

func (suite *AWSTestSuite) TestDecodeLambdaFingerprint() {
	for _, t := range []struct {
		name              string
		base64Fingerprint string
		wantFingerprint   string
		wantErr           bool
	}{
		{
			name:              "valid base64 fingerprint gets decoded and converted",
			base64Fingerprint: "16ikLdccyKitxEizXiYBnXQUOkf2Y49MagwOKmTykdg=",
			wantFingerprint:   "d7a8a42dd71cc8a8adc448b35e26019d74143a47f6638f4c6a0c0e2a64f291d8",
		},
		{
			name:              "invalid base64 string causes an error",
			base64Fingerprint: "2023-01-22",
			wantErr:           true,
		},
	} {
		suite.Run(t.name, func() {
			got, err := decodeLambdaFingerprint(t.base64Fingerprint)
			require.False(suite.T(), (err != nil) != t.wantErr,
				"decodeLambdaFingerprint() error = %v, wantErr %v", err, t.wantErr)
			if !t.wantErr {
				require.Equal(suite.T(), t.wantFingerprint, got)
			}
		})
	}
}

func (suite *AWSTestSuite) TestNewEcsTaskData() {
	taskARN := ""
	cluster := "foo"
	service := "bar"
	digests := map[string]string{}
	time := time.Now()
	expected := &EcsTaskData{
		TaskArn:   taskARN,
		Cluster:   cluster,
		Service:   service,
		Digests:   digests,
		StartedAt: time.Unix(),
	}
	got := NewEcsTaskData(taskARN, cluster, service, digests, time)
	require.Equal(suite.T(), expected, got)
}

func (suite *AWSTestSuite) TestGetConfigOptFns() {
	for _, t := range []struct {
		name         string
		creds        *AWSStaticCreds
		wantedLength int
	}{
		{
			name:         "no creds provided results in an empty list of OptFns",
			creds:        &AWSStaticCreds{},
			wantedLength: 0,
		},
		{
			name: "specifying the region results in one OptFns",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			wantedLength: 1,
		},
		{
			name: "specifying the region and only one auth value results in one OptFns",
			creds: &AWSStaticCreds{
				Region:      "eu-central-1",
				AccessKeyID: "ssss",
			},
			wantedLength: 1,
		},
		{
			name: "specifying the region and both auth value results in two OptFns",
			creds: &AWSStaticCreds{
				Region:          "eu-central-1",
				AccessKeyID:     "ssss",
				SecretAccessKey: "sssss",
			},
			wantedLength: 2,
		},
		{
			name: "specifying the both auth value results in one OptFns",
			creds: &AWSStaticCreds{
				AccessKeyID:     "ssss",
				SecretAccessKey: "sssss",
			},
			wantedLength: 1,
		},
	} {
		suite.Run(t.name, func() {
			got := t.creds.GetConfigOptFns()
			require.Len(suite.T(), got, t.wantedLength)
		})
	}
}

func (suite *AWSTestSuite) TestNewAWSConfigFromEnvOrFlags() {
	for _, t := range []struct {
		name        string
		creds       *AWSStaticCreds
		checkRegion bool
		checkAuth   bool
		wantErr     bool
	}{
		{
			name:  "not providing creds still produces a config",
			creds: &AWSStaticCreds{},
		},
		{
			name: "a provided region is configured in the returned config",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			checkRegion: true,
		},
		{
			name: "a provided region and auth are configured in the returned config",
			creds: &AWSStaticCreds{
				Region:          "eu-central-1",
				AccessKeyID:     "ssss",
				SecretAccessKey: "ssss",
			},
			checkRegion: true,
			checkAuth:   true,
		},
	} {
		suite.Run(t.name, func() {
			config, err := t.creds.NewAWSConfigFromEnvOrFlags()
			require.False(suite.T(), (err != nil) != t.wantErr,
				"NewAWSConfigFromEnvOrFlags() error = %v, wantErr %v", err, t.wantErr)
			if !t.wantErr {
				require.NotNil(suite.T(), config)
				if t.checkRegion {
					require.Equal(suite.T(), config.Region, t.creds.Region)
				}
				if t.checkAuth {
					c, err := config.Credentials.Retrieve(context.TODO())
					require.NoError(suite.T(), err)
					require.Equal(suite.T(), c.AccessKeyID, t.creds.AccessKeyID)
					require.Equal(suite.T(), c.SecretAccessKey, t.creds.SecretAccessKey)
				}
			}
		})
	}
}

func (suite *AWSTestSuite) TestAWSClients() {
	for _, t := range []struct {
		name    string
		creds   *AWSStaticCreds
		wantErr bool
	}{
		{
			name:  "not providing creds still produces valid clients",
			creds: &AWSStaticCreds{},
		},
		{
			name: "a provided region can produce valid clients",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
		},
		{
			name: "a provided region and auth can produce clients",
			creds: &AWSStaticCreds{
				Region:          "eu-central-1",
				AccessKeyID:     "ssss",
				SecretAccessKey: "ssss",
			},
		},
	} {
		suite.Run(t.name, func() {
			s3Client, err := t.creds.NewS3Client()
			require.False(suite.T(), (err != nil) != t.wantErr,
				"NewS3Client() error = %v, wantErr %v", err, t.wantErr)
			if !t.wantErr {
				require.NotNil(suite.T(), s3Client)
			}

			lambdaClient, err := t.creds.NewLambdaClient()
			require.False(suite.T(), (err != nil) != t.wantErr,
				"NewLambdaClient() error = %v, wantErr %v", err, t.wantErr)
			if !t.wantErr {
				require.NotNil(suite.T(), lambdaClient)
			}

			ecsClient, err := t.creds.NewECSClient()
			require.False(suite.T(), (err != nil) != t.wantErr,
				"NewECSClient() error = %v, wantErr %v", err, t.wantErr)
			if !t.wantErr {
				require.NotNil(suite.T(), ecsClient)
			}
		})
	}
}

// The tests below make actual calls to AWS API.
// Some test cases test failing the requests and others test passing them
// The passing cases require AWS creds to be exported in the env, otherwise,
// they are skipped
// All cases will run in CI

// TestGetLambdaPackageData is a smoke test for the real AWS integration.
// Filtering logic and fingerprint processing are covered by fake-backed tests
// (TestGetFilteredLambdaFuncs, TestGetLambdaPackageDataFromClient).
// These cases prove the real SDK wiring works.
func (suite *AWSTestSuite) TestGetLambdaPackageData() {
	type expectedFunction struct {
		name        string
		fingerprint string
	}
	for _, t := range []struct {
		name              string
		requireEnvVars    bool
		creds             *AWSStaticCreds
		filter            *filters.ResourceFilterOptions
		expectedFunctions []expectedFunction
		wantErr           bool
	}{
		{
			name: "invalid credentials causes an error",
			creds: &AWSStaticCreds{
				Region:          "eu-central-1",
				AccessKeyID:     "ssss",
				SecretAccessKey: "ssss",
			},
			filter:  &filters.ResourceFilterOptions{IncludeNames: []string{"cli-tests"}},
			wantErr: true,
		},
		{
			name: "can get zip package lambda function data",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			filter: &filters.ResourceFilterOptions{IncludeNames: []string{"cli-tests"}},
			expectedFunctions: []expectedFunction{{name: "cli-tests",
				fingerprint: "321e3c38e91262e5c72df4bd405e9b177b6f4d750e1af0b78ca2e2b85d6f91b4"}},
			requireEnvVars: true,
		},
		{
			name: "can get image package lambda function data",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			filter: &filters.ResourceFilterOptions{IncludeNames: []string{"cli-tests-docker"}},
			expectedFunctions: []expectedFunction{{name: "cli-tests-docker",
				fingerprint: "e908950659e56bb886acbb0ecf9b8f38bf6e0382ede71095e166269ee4db601e"}},
			requireEnvVars: true,
		},
	} {
		suite.Run(t.name, func() {
			skipOrSetCreds(suite.T(), t.requireEnvVars, t.creds)
			data, err := t.creds.GetLambdaPackageData(t.filter)
			require.False(suite.T(), (err != nil) != t.wantErr,
				"GetLambdaPackageData() error = %v, wantErr %v", err, t.wantErr)
			if !t.wantErr {
				require.Len(suite.T(), data, len(t.expectedFunctions))
				for _, expected := range t.expectedFunctions {
					found := false
					for _, item := range data {
						if fingerprint, ok := item.Digests[expected.name]; ok && fingerprint == expected.fingerprint {
							found = true
							break
						}
					}
					require.True(suite.T(), found, "expected function %s with fingerprint %s not found", expected.name, expected.fingerprint)
				}
			}
		})
	}
}

func (suite *AWSTestSuite) TestGetS3Data() {
	for _, t := range []struct {
		name             string
		requireEnvVars   bool // indicates that a test case needs real credentials from env vars
		creds            *AWSStaticCreds
		bucketName       string
		includePaths     []string
		excludePaths     []string
		wantFingerprint  string
		wantArtifactName string
		wantErr          bool
	}{
		{
			name: "invalid credentials causes an error",
			creds: &AWSStaticCreds{
				Region:          "eu-central-1",
				AccessKeyID:     "ssss",
				SecretAccessKey: "ssss",
			},
			bucketName: "kosli-cli-public",
			wantErr:    true,
		},
		{
			name: "providing wrong region causes an error",
			creds: &AWSStaticCreds{
				Region: "ap-south-1",
			},
			bucketName:     "kosli-cli-public",
			requireEnvVars: true,
			wantErr:        true,
		},
		{
			name: "can get S3 bucket data from entire bucket",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			bucketName:     "kosli-cli-public",
			requireEnvVars: true,
		},
		{
			name: "can get S3 bucket data. includePaths is a sub-directory",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			bucketName:      "kosli-cli-public",
			includePaths:    []string{"dummy"},
			requireEnvVars:  true,
			wantFingerprint: "1b7888b437ba378a9884a937552cb1f945f420c3f4201437b42e690f102ff698",
		},
		{
			name: "when includePaths is not an absolute or relative paths",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			bucketName:     "kosli-cli-public",
			includePaths:   []string{"dummy_2"},
			requireEnvVars: true,
			wantErr:        true,
		},
		{
			name: "can get S3 bucket data. includePaths is a nested sub-directory",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			bucketName:       "kosli-cli-public",
			includePaths:     []string{"dummy/dummy_2"},
			requireEnvVars:   true,
			wantFingerprint:  "02eb06f5778c69431b4b00489074b76f05814d8170949f965ebe13a211bf682a",
			wantArtifactName: "template.yml",
		},
		{
			name: "includePaths is a nested sub-directory starting with slash",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			bucketName:       "kosli-cli-public",
			includePaths:     []string{"/dummy/dummy_2"},
			requireEnvVars:   true,
			wantFingerprint:  "02eb06f5778c69431b4b00489074b76f05814d8170949f965ebe13a211bf682a",
			wantArtifactName: "template.yml",
		},
		{
			name: "can get S3 bucket data. includePaths is a file",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			bucketName:       "kosli-cli-public",
			includePaths:     []string{"README.md"},
			requireEnvVars:   true,
			wantFingerprint:  "77b1b4df1eb620e05ce365e9e84d37a7e04fde8a66251c121773d013dfba0ee6",
			wantArtifactName: "README.md",
		},
		{
			name: "can get S3 bucket data. excludePaths is a file",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			bucketName:      "kosli-cli-public",
			excludePaths:    []string{"README.md"},
			requireEnvVars:  true,
			wantFingerprint: "1b7888b437ba378a9884a937552cb1f945f420c3f4201437b42e690f102ff698",
		},
		{
			name: "can get S3 bucket data. excludePaths is a dir",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			bucketName:       "kosli-cli-public",
			excludePaths:     []string{"dummy"},
			requireEnvVars:   true,
			wantFingerprint:  "77b1b4df1eb620e05ce365e9e84d37a7e04fde8a66251c121773d013dfba0ee6",
			wantArtifactName: "README.md",
		},
	} {
		suite.Run(t.name, func() {
			skipOrSetCreds(suite.T(), t.requireEnvVars, t.creds)
			data, err := t.creds.GetS3Data(t.bucketName, t.includePaths, t.excludePaths, logger.NewStandardLogger())
			require.False(suite.T(), (err != nil) != t.wantErr,
				"GetS3Data() error = %v, wantErr %v", err, t.wantErr)
			if !t.wantErr {
				if t.wantArtifactName == "" {
					t.wantArtifactName = t.bucketName
				}
				if t.wantFingerprint == "" {
					require.Contains(suite.T(), data[0].Digests, t.wantArtifactName)
				} else {
					require.Equal(suite.T(), t.wantFingerprint, data[0].Digests[t.wantArtifactName])
				}
			}
		})
	}
}

func (suite *AWSTestSuite) TestGetEcsTasksData() {
	for _, t := range []struct {
		name                 string
		requireEnvVars       bool // indicates that a test case needs real credentials from env vars
		creds                *AWSStaticCreds
		clusterFilter        *filters.ResourceFilterOptions
		serviceFilter        *filters.ResourceFilterOptions
		minNumberOfArtifacts int
		wantErr              bool
	}{
		{
			name: "invalid credentials causes an error",
			creds: &AWSStaticCreds{
				Region:          "eu-central-1",
				AccessKeyID:     "ssss",
				SecretAccessKey: "ssss",
			},
			clusterFilter: &filters.ResourceFilterOptions{IncludeNames: []string{"merkely"}},
			wantErr:       true,
		},
		{
			name: "can get ECS data without any filters",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			minNumberOfArtifacts: 1,
			requireEnvVars:       true,
		},
		{
			name: "can get ECS data with cluster name alone",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			clusterFilter:        &filters.ResourceFilterOptions{IncludeNames: []string{"merkely"}},
			minNumberOfArtifacts: 1,
			requireEnvVars:       true,
		},
		{
			name: "providing the wrong region finds 0 artifacts",
			creds: &AWSStaticCreds{
				Region: "ap-south-1",
			},
			clusterFilter:        &filters.ResourceFilterOptions{IncludeNames: []string{"merkely"}},
			minNumberOfArtifacts: 0,
			requireEnvVars:       true,
		},
		{
			name: "can get ECS data with exclude names",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			clusterFilter:        &filters.ResourceFilterOptions{ExcludeNames: []string{"slackapp"}},
			minNumberOfArtifacts: 1,
			requireEnvVars:       true,
		},
		{
			name: "can get ECS data with exclude names regex",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			clusterFilter:        &filters.ResourceFilterOptions{ExcludeNamesRegex: []string{"slack.*"}},
			minNumberOfArtifacts: 1,
			requireEnvVars:       true,
		},
		{
			name: "can get ECS data with cluster names regex",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			clusterFilter:        &filters.ResourceFilterOptions{IncludeNamesRegex: []string{"^merk.*"}},
			minNumberOfArtifacts: 1,
			requireEnvVars:       true,
		},
		{
			name: "can get ECS data with service names",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			clusterFilter:        &filters.ResourceFilterOptions{IncludeNames: []string{"merkely"}},
			serviceFilter:        &filters.ResourceFilterOptions{IncludeNames: []string{"kosli"}},
			minNumberOfArtifacts: 1,
			requireEnvVars:       true,
		},
		{
			name: "can get ECS data with service names regex",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			clusterFilter:        &filters.ResourceFilterOptions{IncludeNames: []string{"merkely"}},
			serviceFilter:        &filters.ResourceFilterOptions{IncludeNamesRegex: []string{"^kos.*"}},
			minNumberOfArtifacts: 1,
			requireEnvVars:       true,
		},
		{
			name: "can get ECS data with service names regex",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			clusterFilter:        &filters.ResourceFilterOptions{IncludeNames: []string{"merkely"}},
			serviceFilter:        &filters.ResourceFilterOptions{IncludeNamesRegex: []string{"^kos.*"}},
			minNumberOfArtifacts: 1,
			requireEnvVars:       true,
		},
		{
			name: "can get ECS data with exclude service names",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			clusterFilter:        &filters.ResourceFilterOptions{IncludeNames: []string{"merkely"}},
			serviceFilter:        &filters.ResourceFilterOptions{ExcludeNames: []string{"kosli"}},
			minNumberOfArtifacts: 0,
			requireEnvVars:       true,
		},
		{
			name: "can get ECS data with exclude service names regex",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			clusterFilter:        &filters.ResourceFilterOptions{IncludeNames: []string{"merkely"}},
			serviceFilter:        &filters.ResourceFilterOptions{ExcludeNamesRegex: []string{"^kos.*"}},
			minNumberOfArtifacts: 0,
			requireEnvVars:       true,
		},
	} {
		suite.Run(t.name, func() {
			skipOrSetCreds(suite.T(), t.requireEnvVars, t.creds)
			if t.clusterFilter == nil {
				t.clusterFilter = new(filters.ResourceFilterOptions)
			}
			if t.serviceFilter == nil {
				t.serviceFilter = new(filters.ResourceFilterOptions)
			}
			data, err := t.creds.GetEcsTasksData(t.clusterFilter, t.serviceFilter, logger.NewStandardLogger())
			require.False(suite.T(), (err != nil) != t.wantErr,
				"GetEcsTasksData() error = %v, wantErr %v", err, t.wantErr)
			if !t.wantErr {
				require.GreaterOrEqual(suite.T(), len(data), t.minNumberOfArtifacts)
			}
		})
	}
}

// helper to build a FakeLambdaClient with named functions for testing
func fakeLambdaClientWithFunctions(names ...string) *FakeLambdaClient {
	fns := make([]types.FunctionConfiguration, len(names))
	lastModified := "2024-01-15T10:30:00.000+0000"
	codeSha256 := "abc123"
	for i, name := range names {
		n := name
		fns[i] = types.FunctionConfiguration{
			FunctionName: &n,
			CodeSha256:   &codeSha256,
			LastModified: &lastModified,
			PackageType:  types.PackageTypeZip,
		}
	}
	return &FakeLambdaClient{Functions: fns}
}

func functionNames(result *[]types.FunctionConfiguration) []string {
	names := make([]string, len(*result))
	for i, f := range *result {
		names[i] = *f.FunctionName
	}
	return names
}

func (suite *AWSTestSuite) TestGetFilteredLambdaFuncs() {
	for _, t := range []struct {
		name          string
		functions     []string
		filter        *filters.ResourceFilterOptions
		pageSize      int
		expectedNames []string
		wantErr       bool
	}{
		{
			name:          "no filter returns all functions",
			functions:     []string{"alpha", "beta", "gamma"},
			filter:        &filters.ResourceFilterOptions{},
			expectedNames: []string{"alpha", "beta", "gamma"},
		},
		{
			name:          "empty function list returns empty result",
			functions:     []string{},
			filter:        &filters.ResourceFilterOptions{},
			expectedNames: []string{},
		},
		{
			name:          "IncludeNames filters to matching functions",
			functions:     []string{"alpha", "beta", "gamma"},
			filter:        &filters.ResourceFilterOptions{IncludeNames: []string{"beta"}},
			expectedNames: []string{"beta"},
		},
		{
			name:          "IncludeNames with multiple names",
			functions:     []string{"alpha", "beta", "gamma"},
			filter:        &filters.ResourceFilterOptions{IncludeNames: []string{"alpha", "gamma"}},
			expectedNames: []string{"alpha", "gamma"},
		},
		{
			name:          "IncludeNamesRegex filters by pattern",
			functions:     []string{"alpha", "beta", "gamma"},
			filter:        &filters.ResourceFilterOptions{IncludeNamesRegex: []string{"^a.*"}},
			expectedNames: []string{"alpha"},
		},
		{
			name:          "ExcludeNames removes matching functions",
			functions:     []string{"alpha", "beta", "gamma"},
			filter:        &filters.ResourceFilterOptions{ExcludeNames: []string{"beta"}},
			expectedNames: []string{"alpha", "gamma"},
		},
		{
			name:          "ExcludeNamesRegex removes matching pattern",
			functions:     []string{"alpha", "beta", "gamma"},
			filter:        &filters.ResourceFilterOptions{ExcludeNamesRegex: []string{"^[bg].*"}},
			expectedNames: []string{"alpha"},
		},
		{
			name:      "combined ExcludeNames and ExcludeNamesRegex",
			functions: []string{"alpha", "beta", "gamma", "delta"},
			filter: &filters.ResourceFilterOptions{
				ExcludeNames:      []string{"alpha"},
				ExcludeNamesRegex: []string{"^d.*"},
			},
			expectedNames: []string{"beta", "gamma"},
		},
		{
			name:          "multi-page results with filtering across pages",
			functions:     []string{"alpha", "beta", "gamma", "delta"},
			filter:        &filters.ResourceFilterOptions{IncludeNamesRegex: []string{"^[ag].*"}},
			pageSize:      2,
			expectedNames: []string{"alpha", "gamma"},
		},
		{
			name:          "multi-page results without filtering",
			functions:     []string{"alpha", "beta", "gamma"},
			filter:        &filters.ResourceFilterOptions{},
			pageSize:      1,
			expectedNames: []string{"alpha", "beta", "gamma"},
		},
		{
			name:      "invalid regex causes an error",
			functions: []string{"alpha"},
			filter:    &filters.ResourceFilterOptions{IncludeNamesRegex: []string{"invalid["}},
			wantErr:   true,
		},
	} {
		suite.Run(t.name, func() {
			client := fakeLambdaClientWithFunctions(t.functions...)
			if t.pageSize > 0 {
				client.PageSize = t.pageSize
			}
			result, err := getFilteredLambdaFuncs(client, nil, &[]types.FunctionConfiguration{}, t.filter)
			if t.wantErr {
				require.Error(suite.T(), err)
				return
			}
			require.NoError(suite.T(), err)
			require.ElementsMatch(suite.T(), t.expectedNames, functionNames(result))
		})
	}
}

func (suite *AWSTestSuite) TestGetLambdaPackageDataFromClient() {
	// base64-encoded SHA256 that decodes to a known hex fingerprint
	zipCodeSha256 := "Mh48OOkSYuXHLfS9QF6bF3tvTXUOGvC3jKLiuF1vkbQ="
	expectedZipFingerprint := "321e3c38e91262e5c72df4bd405e9b177b6f4d750e1af0b78ca2e2b85d6f91b4"
	// Image package types use the raw CodeSha256 (not base64-decoded)
	imageCodeSha256 := "e908950659e56bb886acbb0ecf9b8f38bf6e0382ede71095e166269ee4db601e"
	lastModified := "2024-01-15T10:30:00.000+0000"

	for _, t := range []struct {
		name                string
		client              *FakeLambdaClient
		filter              *filters.ResourceFilterOptions
		expectedDigests     map[string]string // functionName -> fingerprint
		wantErr             bool
		wantErrMsgSubstring string
	}{
		{
			name: "single Zip function returns decoded fingerprint",
			client: func() *FakeLambdaClient {
				fnName := "zip-func"
				return &FakeLambdaClient{Functions: []types.FunctionConfiguration{
					{FunctionName: &fnName, CodeSha256: &zipCodeSha256, LastModified: &lastModified, PackageType: types.PackageTypeZip},
				}}
			}(),
			filter:          &filters.ResourceFilterOptions{},
			expectedDigests: map[string]string{"zip-func": expectedZipFingerprint},
		},
		{
			name: "single Image function returns raw CodeSha256",
			client: func() *FakeLambdaClient {
				fnName := "image-func"
				return &FakeLambdaClient{Functions: []types.FunctionConfiguration{
					{FunctionName: &fnName, CodeSha256: &imageCodeSha256, LastModified: &lastModified, PackageType: types.PackageTypeImage},
				}}
			}(),
			filter:          &filters.ResourceFilterOptions{},
			expectedDigests: map[string]string{"image-func": imageCodeSha256},
		},
		{
			name: "multiple functions processed concurrently",
			client: func() *FakeLambdaClient {
				fn1 := "zip-func"
				fn2 := "image-func"
				return &FakeLambdaClient{Functions: []types.FunctionConfiguration{
					{FunctionName: &fn1, CodeSha256: &zipCodeSha256, LastModified: &lastModified, PackageType: types.PackageTypeZip},
					{FunctionName: &fn2, CodeSha256: &imageCodeSha256, LastModified: &lastModified, PackageType: types.PackageTypeImage},
				}}
			}(),
			filter: &filters.ResourceFilterOptions{},
			expectedDigests: map[string]string{
				"zip-func":   expectedZipFingerprint,
				"image-func": imageCodeSha256,
			},
		},
		{
			name: "empty function list returns empty result",
			client: &FakeLambdaClient{
				Functions: []types.FunctionConfiguration{},
			},
			filter:          &filters.ResourceFilterOptions{},
			expectedDigests: map[string]string{},
		},
		{
			name: "GetFunctionConfiguration error propagates",
			client: func() *FakeLambdaClient {
				fnName := "will-fail"
				return &FakeLambdaClient{
					Functions: []types.FunctionConfiguration{
						{FunctionName: &fnName, CodeSha256: &zipCodeSha256, LastModified: &lastModified, PackageType: types.PackageTypeZip},
					},
					GetFunctionConfigurationErr: fmt.Errorf("simulated AWS error"),
				}
			}(),
			filter:  &filters.ResourceFilterOptions{},
			wantErr: true,
		},
	} {
		suite.Run(t.name, func() {
			data, err := getLambdaPackageDataFromClient(t.client, t.filter)
			if t.wantErr {
				require.Error(suite.T(), err)
				return
			}
			require.NoError(suite.T(), err)

			gotDigests := map[string]string{}
			for _, d := range data {
				for name, fp := range d.Digests {
					gotDigests[name] = fp
				}
			}
			require.Equal(suite.T(), t.expectedDigests, gotDigests)
		})
	}
}

func skipOrSetCreds(T *testing.T, requireEnvVars bool, creds *AWSStaticCreds) {
	if requireEnvVars {
		// skips the test case if it requires env vars and they are not set
		testHelpers.SkipIfEnvVarUnset(T, []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAWSTestSuite(t *testing.T) {
	suite.Run(t, new(AWSTestSuite))
}
