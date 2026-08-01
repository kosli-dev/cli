package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"

	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/kosli-dev/cli/internal/aws"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type SnapshotS3TestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
	bucketName            string
}

func (suite *SnapshotS3TestSuite) SetupTest() {
	suite.envName = "snapshot-s3-env"
	suite.bucketName = "kosli-cli-public"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateEnv(global.Org, suite.envName, "S3", suite.T())

	// Inject a fake S3 client so tests run without AWS credentials.
	// The fake is seeded with the objects the test cases filter on.
	bucketName := suite.bucketName
	objects := map[string][]byte{
		"README.md":                  []byte("# kosli cli public\n"),
		"dummy/dummy_2/template.yml": []byte("key: value\n"),
	}
	// Only README.md carries a stored checksum, so the metadata cases cover both
	// an object that can be fingerprinted from metadata and one that cannot.
	readmeSum := sha256.Sum256(objects["README.md"])
	aws.NewS3ClientFunc = func(_ *aws.AWSStaticCreds) (aws.S3API, error) {
		return &aws.FakeS3Client{
			Bucket:  bucketName,
			Objects: objects,
			Checksums: map[string]aws.FakeS3Checksum{
				"README.md": {
					SHA256: base64.StdEncoding.EncodeToString(readmeSum[:]),
					Type:   s3Types.ChecksumTypeFullObject,
				},
			},
		}, nil
	}
}

func (suite *SnapshotS3TestSuite) TearDownTest() {
	aws.ResetS3ClientFactory()
}

func (suite *SnapshotS3TestSuite) TestSnapshotS3Cmd() {
	tests := []cmdTestCase{
		{
			name:   "snapshot s3 works with --bucket",
			cmd:    fmt.Sprintf(`snapshot s3 %s %s --bucket %s`, suite.envName, suite.defaultKosliArguments, suite.bucketName),
			golden: "bucket " + suite.bucketName + " was reported to environment " + suite.envName + "\n",
		},
		{
			wantError: true,
			name:      "snapshot s3 fails without --bucket",
			cmd:       fmt.Sprintf(`snapshot s3 %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"bucket\" not set\n",
		},
		{
			wantError: true,
			name:      "snapshot s3 fails if no args are set",
			cmd:       fmt.Sprintf(`snapshot s3 %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "snapshot s3 fails two args are set",
			cmd:       fmt.Sprintf(`snapshot s3 %s xxx %s --bucket %s`, suite.envName, suite.defaultKosliArguments, suite.bucketName),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "snapshot s3 fails if --include and --exclude are set",
			cmd:       fmt.Sprintf(`snapshot s3 %s %s --bucket %s --include foo --exclude bar`, suite.envName, suite.defaultKosliArguments, suite.bucketName),
			golden:    "Error: only one of --include, --exclude is allowed\n",
		},
		{
			wantError: true,
			name:      "snapshot s3 fails if --include and --exclude-regex are set",
			cmd:       fmt.Sprintf(`snapshot s3 %s %s --bucket %s --include foo --exclude-regex bar`, suite.envName, suite.defaultKosliArguments, suite.bucketName),
			golden:    "Error: only one of --include, --exclude-regex is allowed\n",
		},
		{
			wantError: true,
			name:      "snapshot s3 fails if --include-regex and --exclude are set",
			cmd:       fmt.Sprintf(`snapshot s3 %s %s --bucket %s --include-regex foo --exclude bar`, suite.envName, suite.defaultKosliArguments, suite.bucketName),
			golden:    "Error: only one of --include-regex, --exclude is allowed\n",
		},
		{
			wantError: true,
			name:      "snapshot s3 fails if --include-regex and --exclude-regex are set",
			cmd:       fmt.Sprintf(`snapshot s3 %s %s --bucket %s --include-regex foo --exclude-regex bar`, suite.envName, suite.defaultKosliArguments, suite.bucketName),
			golden:    "Error: only one of --include-regex, --exclude-regex is allowed\n",
		},
		{
			name:   "can snapshot a subset of files/dirs using --include",
			cmd:    fmt.Sprintf(`snapshot s3 %s %s --bucket %s --include README.md`, suite.envName, suite.defaultKosliArguments, suite.bucketName),
			golden: "bucket kosli-cli-public was reported to environment snapshot-s3-env\n",
		},
		{
			wantError: true,
			name:      "fails when --include does not match any file or dir",
			cmd:       fmt.Sprintf(`snapshot s3 %s %s --bucket %s --include non-existing.md`, suite.envName, suite.defaultKosliArguments, suite.bucketName),
			golden:    "Error: no matching file or dirs in bucket: [kosli-cli-public]\n",
		},
		{
			name:   "can snapshot entire bucket except a subset of files/dirs using --exclude",
			cmd:    fmt.Sprintf(`snapshot s3 %s %s --bucket %s --exclude dummy`, suite.envName, suite.defaultKosliArguments, suite.bucketName),
			golden: "bucket kosli-cli-public was reported to environment snapshot-s3-env\n",
		},
		{
			name:   "--fingerprint-source metadata fingerprints from the stored checksum",
			cmd:    fmt.Sprintf(`snapshot s3 %s %s --bucket %s --include README.md --fingerprint-source metadata`, suite.envName, suite.defaultKosliArguments, suite.bucketName),
			golden: "bucket kosli-cli-public was reported to environment snapshot-s3-env\n",
		},
		{
			name:   "--fingerprint-source content is the default behaviour",
			cmd:    fmt.Sprintf(`snapshot s3 %s %s --bucket %s --fingerprint-source content`, suite.envName, suite.defaultKosliArguments, suite.bucketName),
			golden: "bucket kosli-cli-public was reported to environment snapshot-s3-env\n",
		},
		{
			wantError: true,
			name:      "--fingerprint-source rejects an unknown value",
			cmd:       fmt.Sprintf(`snapshot s3 %s %s --bucket %s --fingerprint-source etag`, suite.envName, suite.defaultKosliArguments, suite.bucketName),
			golden:    "Error: etag is not a valid fingerprint source. Valid sources are: [content, metadata]\nUsage: kosli snapshot s3 ENVIRONMENT-NAME [flags]\n",
		},
		{
			wantError: true,
			name:      "--fingerprint-source metadata fails on an object with no stored checksum",
			cmd:       fmt.Sprintf(`snapshot s3 %s %s --bucket %s --include dummy --fingerprint-source metadata`, suite.envName, suite.defaultKosliArguments, suite.bucketName),
			golden:    "Error: object \"dummy/dummy_2/template.yml\" in bucket [kosli-cli-public] has no SHA256 checksum, so its fingerprint cannot be derived from S3 metadata. Re-upload it with a checksum: aws s3api put-object --bucket kosli-cli-public --key dummy/dummy_2/template.yml --body <file> --checksum-algorithm SHA256. Or fingerprint by downloading the objects instead\n",
		},
	}

	for _, t := range tests {
		runTestCmd(suite.T(), []cmdTestCase{t})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSnapshotS3TestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotS3TestSuite))
}
