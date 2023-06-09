package aws

import (
	"context"
	"testing"
	"time"

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
	digests := map[string]string{}
	time := time.Now()
	expected := &EcsTaskData{
		TaskArn:   taskARN,
		Digests:   digests,
		StartedAt: time.Unix(),
	}
	got := NewEcsTaskData(taskARN, digests, time)
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

func (suite *AWSTestSuite) TestGetLambdaPackageData() {
	for _, t := range []struct {
		name             string
		requireEnvVars   bool // indicates that a test case needs real credentials from env vars
		creds            *AWSStaticCreds
		functionNames    []string
		wantFingerprints []string
		wantErr          bool
	}{
		{
			name: "invalid credentials causes an error",
			creds: &AWSStaticCreds{
				Region:          "eu-central-1",
				AccessKeyID:     "ssss",
				SecretAccessKey: "ssss",
			},
			functionNames: []string{"ewelina-test"},
			wantErr:       true,
		},
		{
			name: "providing the wrong region causes a failure",
			creds: &AWSStaticCreds{
				Region: "ap-south-1",
			},
			functionNames:  []string{"ewelina-test"},
			requireEnvVars: true,
			wantErr:        true,
		},
		{
			name: "can get zip package lambda function data from name",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			functionNames:    []string{"ewelina-test"},
			wantFingerprints: []string{"61db6a4ac396a7af4c48bc927c9d02e5a093ddc2a4c51b50d2194be436452592"},
			requireEnvVars:   true,
		},
		{
			name: "can get image package lambda function data from name",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			functionNames:    []string{"lambda-docker-test"},
			wantFingerprints: []string{"e3e2e565788902e24d1b17c0c6bdb7dfc1cdfe6193b762482bbe982bd83a9876"},
			requireEnvVars:   true,
		},
		{
			name: "can get a list of lambda functions data from names",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			functionNames: []string{"lambda-docker-test", "ewelina-test"},
			wantFingerprints: []string{"e3e2e565788902e24d1b17c0c6bdb7dfc1cdfe6193b762482bbe982bd83a9876",
				"61db6a4ac396a7af4c48bc927c9d02e5a093ddc2a4c51b50d2194be436452592"},
			requireEnvVars: true,
		},
	} {
		suite.Run(t.name, func() {
			skipOrSetCreds(suite.T(), t.requireEnvVars, t.creds)
			data, err := t.creds.GetLambdaPackageData(t.functionNames)
			require.False(suite.T(), (err != nil) != t.wantErr,
				"GetLambdaPackageData() error = %v, wantErr %v", err, t.wantErr)
			if !t.wantErr {
				matchFound := false
			loop1:
				for index, name := range t.functionNames {
					for _, item := range data {
						if fingerprint, ok := item.Digests[name]; ok {
							if t.wantFingerprints[index] == fingerprint {
								matchFound = true
								break loop1
							} else {
								suite.T().Logf("fingerprint did not match: GOT %s -- WANT %s", fingerprint, t.wantFingerprints[index])
							}
						}
					}
				}
				require.True(suite.T(), matchFound)
			}
		})
	}
}

func (suite *AWSTestSuite) TestGetS3Data() {
	for _, t := range []struct {
		name            string
		requireEnvVars  bool // indicates that a test case needs real credentials from env vars
		creds           *AWSStaticCreds
		bucketName      string
		wantFingerprint string
		wantErr         bool
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
			name: "can get S3 bucket data",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			bucketName:     "kosli-cli-public",
			requireEnvVars: true,
		},
	} {
		suite.Run(t.name, func() {
			skipOrSetCreds(suite.T(), t.requireEnvVars, t.creds)
			data, err := t.creds.GetS3Data(t.bucketName, logger.NewStandardLogger())
			require.False(suite.T(), (err != nil) != t.wantErr,
				"GetS3Data() error = %v, wantErr %v", err, t.wantErr)
			if !t.wantErr {
				if t.wantFingerprint == "" {
					require.Contains(suite.T(), data[0].Digests, t.bucketName)
				} else {
					require.Equal(suite.T(), t.wantFingerprint, data[0].Digests[t.bucketName])
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
		clusterName          string
		serviceName          string
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
			clusterName: "merkely",
			wantErr:     true,
		},
		{
			name: "can get ECS data with cluster name alone",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			clusterName:          "merkely",
			minNumberOfArtifacts: 2,
			requireEnvVars:       true,
		},
		{
			name: "providing the wrong region causes an error",
			creds: &AWSStaticCreds{
				Region: "ap-south-1",
			},
			clusterName:          "merkely",
			minNumberOfArtifacts: 2,
			requireEnvVars:       true,
			wantErr:              true,
		},
		{
			name: "can get ECS data with cluster name and service name",
			creds: &AWSStaticCreds{
				Region: "eu-central-1",
			},
			clusterName:          "merkely",
			serviceName:          "merkely",
			minNumberOfArtifacts: 2,
			requireEnvVars:       true,
		},
	} {
		suite.Run(t.name, func() {
			skipOrSetCreds(suite.T(), t.requireEnvVars, t.creds)
			data, err := t.creds.GetEcsTasksData(t.clusterName, t.serviceName)
			require.False(suite.T(), (err != nil) != t.wantErr,
				"GetEcsTasksData() error = %v, wantErr %v", err, t.wantErr)
			if !t.wantErr {
				require.GreaterOrEqual(suite.T(), len(data), t.minNumberOfArtifacts)
			}
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
