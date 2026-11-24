package main

import (
	"testing"

	"github.com/kosli-dev/cli/internal/aws"
	"github.com/stretchr/testify/require"
)

const harvestTestBucket = "harvest-test-bucket"

func harvestFrom(t *testing.T, objects map[string][]byte, o *snapshotS3Options) ([]*aws.S3Data, error) {
	t.Helper()
	aws.NewS3ClientFunc = func(_ *aws.AWSStaticCreds) (aws.S3API, error) {
		return &aws.FakeS3Client{Bucket: harvestTestBucket, Objects: objects}, nil
	}
	t.Cleanup(aws.ResetS3ClientFactory)
	o.bucket = harvestTestBucket
	o.awsStaticCreds = new(aws.AWSStaticCreds)
	o.downloadLimits = aws.DefaultDownloadLimits
	return o.harvest()
}

func TestAllowEmptyBucketReportsAnEmptyBucketAsAnEmptySnapshot(t *testing.T) {
	for _, source := range []string{fingerprintSourceContent, fingerprintSourceMetadata} {
		t.Run(source, func(t *testing.T) {
			data, err := harvestFrom(t, map[string][]byte{"dir/": nil},
				&snapshotS3Options{fingerprintSource: source, allowEmptyBucket: true})

			require.NoError(t, err)
			require.Empty(t, data)
			require.NotNil(t, data, "the payload must carry an empty list, not null")
		})
	}
}

func TestAnEmptyBucketFailsWithoutAllowEmptyBucket(t *testing.T) {
	for _, source := range []string{fingerprintSourceContent, fingerprintSourceMetadata} {
		t.Run(source, func(t *testing.T) {
			_, err := harvestFrom(t, map[string][]byte{}, &snapshotS3Options{fingerprintSource: source})

			require.EqualError(t, err, "no matching file or dirs in bucket: ["+harvestTestBucket+"]")
		})
	}
}

// The flag is about the bucket, not the filters: a filter that leaves nothing
// is more likely mistyped than the bucket is empty.
func TestAllowEmptyBucketDoesNotExcuseAFilterThatMatchesNothing(t *testing.T) {
	_, err := harvestFrom(t, map[string][]byte{"README.md": []byte("# readme\n")},
		&snapshotS3Options{fingerprintSource: fingerprintSourceContent, allowEmptyBucket: true, includePaths: []string{"non-existing.md"}})

	require.EqualError(t, err, "no matching file or dirs in bucket: ["+harvestTestBucket+"]")
}

func TestAllowEmptyBucketLeavesOtherFailuresAlone(t *testing.T) {
	_, err := harvestFrom(t, map[string][]byte{"README.md": []byte("# readme\n")},
		&snapshotS3Options{fingerprintSource: fingerprintSourceMetadata, allowEmptyBucket: true})

	require.ErrorContains(t, err, "object key [README.md] has no SHA256 checksum")
}
