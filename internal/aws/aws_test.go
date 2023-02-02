package aws

import (
	"context"
	"testing"
	"time"

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

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAWSTestSuite(t *testing.T) {
	suite.Run(t, new(AWSTestSuite))
}
