package aws

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/require"
)

// errInjected is the error tests inject into FakeS3Client to exercise error paths.
var errInjected = errors.New("injected error")

// runS3ContractTests exercises the S3API contract. It verifies the behaviours
// we depend on — object listing, continuation-token pagination, object
// download, and error responses for missing buckets and keys.
//
// Any implementation that passes this suite is a valid stand-in for the real
// AWS S3 API as far as this codebase is concerned.
//
// bucket must name a bucket the client can see, holding at least two objects.
// existingKey must name an object in that bucket with a non-empty body.
func runS3ContractTests(t *testing.T, client S3API, bucket, existingKey string) {
	t.Helper()

	t.Run("ListObjectsV2 returns objects with keys and modification times", func(t *testing.T) {
		out, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket: aws.String(bucket),
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		require.NotEmpty(t, out.Contents)
		for _, object := range out.Contents {
			require.NotNil(t, object.Key, "Key should be present")
			require.NotEmpty(t, *object.Key)
			require.NotNil(t, object.LastModified, "LastModified should be present")
		}
	})

	t.Run("ListObjectsV2 with MaxKeys paginates via ContinuationToken", func(t *testing.T) {
		// Request one object per page to force pagination. The paginator only
		// follows a page when IsTruncated is true AND NextContinuationToken is
		// non-empty, so both must be set together.
		out, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket:  aws.String(bucket),
			MaxKeys: aws.Int32(1),
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		require.Len(t, out.Contents, 1)

		if out.IsTruncated == nil || !*out.IsTruncated {
			t.Skip("only 1 object in bucket; pagination not exercisable")
		}
		require.NotNil(t, out.NextContinuationToken, "a truncated result must carry a continuation token")
		require.NotEmpty(t, *out.NextContinuationToken)

		out2, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			MaxKeys:           aws.Int32(1),
			ContinuationToken: out.NextContinuationToken,
		})
		require.NoError(t, err)
		require.NotNil(t, out2)
		require.Len(t, out2.Contents, 1)
		require.NotEqual(t, *out.Contents[0].Key, *out2.Contents[0].Key,
			"the second page should return a different object")
	})

	t.Run("ListObjectsV2 errors for a nonexistent bucket", func(t *testing.T) {
		_, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket: aws.String("kosli-nonexistent-bucket-that-should-not-exist"),
		})
		require.Error(t, err)
	})

	t.Run("DownloadObject writes the object body to the WriterAt", func(t *testing.T) {
		file, err := os.Create(filepath.Join(t.TempDir(), "downloaded"))
		require.NoError(t, err)
		defer file.Close() //nolint:errcheck

		out, err := client.DownloadObject(context.TODO(), &transfermanager.DownloadObjectInput{
			Bucket:   aws.String(bucket),
			Key:      aws.String(existingKey),
			WriterAt: file,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		require.NotNil(t, out.ContentLength, "ContentLength should be present")

		content, err := os.ReadFile(file.Name())
		require.NoError(t, err)
		require.NotEmpty(t, content, "the object body should have been written")
		require.Equal(t, *out.ContentLength, int64(len(content)),
			"ContentLength should match the bytes written")
	})

	t.Run("DownloadObject errors for a missing key", func(t *testing.T) {
		file, err := os.Create(filepath.Join(t.TempDir(), "downloaded"))
		require.NoError(t, err)
		defer file.Close() //nolint:errcheck

		_, err = client.DownloadObject(context.TODO(), &transfermanager.DownloadObjectInput{
			Bucket:   aws.String(bucket),
			Key:      aws.String("nonexistent-key-that-should-not-exist"),
			WriterAt: file,
		})
		require.Error(t, err)
	})
}

func TestS3Contract_Fake(t *testing.T) {
	bucket := "kosli-cli-public"
	client := &FakeS3Client{
		Bucket: bucket,
		Objects: map[string][]byte{
			"README.md":                  []byte("# readme\n"),
			"dummy/dummy_2/template.yml": []byte("key: value\n"),
		},
		// One object per page so the pagination contract is genuinely exercised.
		PageSize: 1,
	}

	runS3ContractTests(t, client, bucket, "README.md")

	// Error injection is a fake-specific mechanism with no real-API equivalent.
	// These tests verify the fake itself, not the contract.
	t.Run("ListObjectsV2 returns error when ListObjectsV2Err is injected", func(t *testing.T) {
		client.ListObjectsV2Err = errInjected
		defer func() { client.ListObjectsV2Err = nil }()
		_, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
		require.Error(t, err)
	})

	t.Run("DownloadObject returns error when DownloadObjectErr is injected", func(t *testing.T) {
		client.DownloadObjectErr = errInjected
		defer func() { client.DownloadObjectErr = nil }()
		file, err := os.Create(filepath.Join(t.TempDir(), "downloaded"))
		require.NoError(t, err)
		defer file.Close() //nolint:errcheck

		_, err = client.DownloadObject(context.TODO(), &transfermanager.DownloadObjectInput{
			Bucket:   aws.String(bucket),
			Key:      aws.String("README.md"),
			WriterAt: file,
		})
		require.Error(t, err)
	})

	// The fake rejects listing inputs outside the contract. Real S3 accepts
	// these without erroring, so they cannot live in runS3ContractTests — they
	// exist so a future caller fails loudly instead of hanging or panicking.
	t.Run("ListObjectsV2 rejects MaxKeys below 1", func(t *testing.T) {
		// MaxKeys 0 would otherwise yield empty-but-truncated pages, spinning
		// s3.ListObjectsV2Paginator forever.
		_, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket:  aws.String(bucket),
			MaxKeys: aws.Int32(0),
		})
		require.Error(t, err)
	})

	t.Run("ListObjectsV2 rejects a negative continuation token", func(t *testing.T) {
		// A negative token parses as a number but would index the key slice
		// out of range.
		_, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			ContinuationToken: aws.String("-1"),
		})
		require.Error(t, err)
	})
}

func TestS3Contract_RealAWS(t *testing.T) {
	testHelpers.SkipIfEnvVarUnset(t, []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"})

	creds := &AWSStaticCreds{Region: "eu-central-1"}
	client, err := defaultNewS3Client(creds)
	require.NoError(t, err)

	runS3ContractTests(t, client, "kosli-cli-public", "README.md")
}
