package aws

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type S3MetadataConcurrencyTestSuite struct {
	suite.Suite
}

// countingS3Client wraps FakeS3Client to record how many HeadObject calls are
// made and how many run at once.
type countingS3Client struct {
	*FakeS3Client
	calls       atomic.Int64
	inFlight    atomic.Int64
	maxInFlight atomic.Int64
	block       chan struct{} // when non-nil, HeadObject waits for a send
}

func (c *countingS3Client) HeadObject(ctx context.Context, params *s3.HeadObjectInput,
	optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	c.calls.Add(1)
	inFlight := c.inFlight.Add(1)
	for {
		max := c.maxInFlight.Load()
		if inFlight <= max || c.maxInFlight.CompareAndSwap(max, inFlight) {
			break
		}
	}
	if c.block != nil {
		<-c.block
	}
	defer c.inFlight.Add(-1)
	return c.FakeS3Client.HeadObject(ctx, params, optFns...)
}

// manyObjects builds a bucket of n checksum-bearing objects.
func manyObjects(n int) (map[string][]byte, map[string]FakeS3Checksum) {
	objects := map[string][]byte{}
	for i := 0; i < n; i++ {
		objects[fmt.Sprintf("object-%03d.txt", i)] = []byte(fmt.Sprintf("content %d\n", i))
	}
	return objects, fullObjectChecksums(objects)
}

func (suite *S3MetadataConcurrencyTestSuite) newClient(n int) *countingS3Client {
	objects, checksums := manyObjects(n)
	return &countingS3Client{
		FakeS3Client: &FakeS3Client{
			Bucket:    fakeS3TestBucketName,
			Objects:   objects,
			Checksums: checksums,
		},
	}
}

func (suite *S3MetadataConcurrencyTestSuite) TestMakesExactlyOneCallPerObject() {
	client := suite.newClient(50)

	_, err := getS3DataFromMetadataClient(client, fakeS3TestBucketName, nil, nil, nil, nil,
		logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), int64(50), client.calls.Load())
}

func (suite *S3MetadataConcurrencyTestSuite) TestRespectsTheConcurrencyBound() {
	client := suite.newClient(200)

	_, err := getS3DataFromMetadataClient(client, fakeS3TestBucketName, nil, nil, nil, nil,
		logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	require.LessOrEqual(suite.T(), client.maxInFlight.Load(), int64(defaultS3MetadataConcurrency),
		"more requests were in flight than the concurrency bound allows")
}

// TestFingerprintIsIndependentOfCompletionOrder pins that concurrency cannot
// reorder the digests: the same bucket must fingerprint identically whether the
// requests finish in listing order or not.
func (suite *S3MetadataConcurrencyTestSuite) TestFingerprintIsIndependentOfCompletionOrder() {
	first, err := getS3DataFromMetadataClient(suite.newClient(30), fakeS3TestBucketName,
		nil, nil, nil, nil, logger.NewStandardLogger())
	require.NoError(suite.T(), err)

	for i := 0; i < 5; i++ {
		again, err := getS3DataFromMetadataClient(suite.newClient(30), fakeS3TestBucketName,
			nil, nil, nil, nil, logger.NewStandardLogger())
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), first, again)
	}
}

func (suite *S3MetadataConcurrencyTestSuite) TestAnAPIErrorStopsRemainingWork() {
	client := suite.newClient(500)
	client.HeadObjectErr = errInjected

	_, err := getS3DataFromMetadataClient(client, fakeS3TestBucketName, nil, nil, nil, nil,
		logger.NewStandardLogger())
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "injected error")
	require.Less(suite.T(), client.calls.Load(), int64(500),
		"an API error should cancel the remaining requests rather than running all of them")
}

// TestReportsEveryUnusableObject checks the error names all the offending keys
// rather than only the first, so a bucket-wide migration is not a
// guess-and-retry loop.
func (suite *S3MetadataConcurrencyTestSuite) TestReportsEveryUnusableObject() {
	objects, checksums := manyObjects(6)
	// Three objects have no stored checksum.
	for _, key := range []string{"object-000.txt", "object-002.txt", "object-004.txt"} {
		delete(checksums, key)
	}
	client := &FakeS3Client{
		Bucket:    fakeS3TestBucketName,
		Objects:   objects,
		Checksums: checksums,
	}

	_, err := getS3DataFromMetadataClient(client, fakeS3TestBucketName, nil, nil, nil, nil,
		logger.NewStandardLogger())
	require.Error(suite.T(), err)
	for _, key := range []string{"object-000.txt", "object-002.txt", "object-004.txt"} {
		require.Contains(suite.T(), err.Error(), key)
	}
	require.NotContains(suite.T(), err.Error(), "object-001.txt",
		"objects that are fine should not be named")
}

// TestCapsTheListOfUnusableObjects keeps the error readable when a whole bucket
// is unusable.
func (suite *S3MetadataConcurrencyTestSuite) TestCapsTheListOfUnusableObjects() {
	objects, _ := manyObjects(40)
	client := &FakeS3Client{
		Bucket:  fakeS3TestBucketName,
		Objects: objects,
	}

	_, err := getS3DataFromMetadataClient(client, fakeS3TestBucketName, nil, nil, nil, nil,
		logger.NewStandardLogger())
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "and 30 more")
	// Each message names the key once in quotes and once in the fix command, so
	// count the quoted form to count messages.
	require.Equal(suite.T(), maxReportedUnusableObjects, strings.Count(err.Error(), `"object-`),
		"the error should list exactly the capped number of keys")
}

// TestMixedFailuresReportChecksumProblems checks that when objects are unusable
// for different reasons, both reasons reach the user.
func (suite *S3MetadataConcurrencyTestSuite) TestMixedFailuresReportChecksumProblems() {
	objects, checksums := manyObjects(4)
	delete(checksums, "object-000.txt")
	checksums["object-001.txt"] = FakeS3Checksum{
		SHA256: base64Sha256(objects["object-001.txt"]) + "-3",
		Type:   s3Types.ChecksumTypeComposite,
	}
	client := &FakeS3Client{
		Bucket:    fakeS3TestBucketName,
		Objects:   objects,
		Checksums: checksums,
	}

	_, err := getS3DataFromMetadataClient(client, fakeS3TestBucketName, nil, nil, nil, nil,
		logger.NewStandardLogger())
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "no SHA256 checksum")
	require.Contains(suite.T(), err.Error(), "multipart (composite)")
}

func TestS3MetadataConcurrencyTestSuite(t *testing.T) {
	suite.Run(t, new(S3MetadataConcurrencyTestSuite))
}
