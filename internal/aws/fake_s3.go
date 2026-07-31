package aws

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// fakeS3LastModified is the modification time reported for objects that have no
// entry in FakeS3Client.LastModified.
var fakeS3LastModified = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

// FakeS3Client is an in-memory implementation of S3API for testing.
// It simulates continuation-token pagination and returns errors for unknown
// buckets and missing objects.
type FakeS3Client struct {
	// Bucket is the bucket this fake serves. Requests for any other bucket
	// return an error, as the real API does for a bucket that does not exist.
	Bucket string
	// Objects maps object key to object content. Keys ending in "/" represent
	// the folder markers that real S3 returns for explicitly created folders.
	Objects map[string][]byte
	// LastModified maps object key to modification time. Keys without an entry
	// report fakeS3LastModified.
	LastModified map[string]time.Time
	// PageSize controls how many objects are returned per ListObjectsV2 call.
	// Defaults to 1000 (matching the AWS default) if zero.
	PageSize int
	// ListObjectsV2Err, if set, is returned by ListObjectsV2 for any request.
	// Useful for testing error propagation.
	ListObjectsV2Err error
	// DownloadObjectErr, if set, is returned by DownloadObject for any object.
	// Useful for testing error propagation.
	DownloadObjectErr error
}

func (f *FakeS3Client) pageSize() int {
	if f.PageSize > 0 {
		return f.PageSize
	}
	return 1000
}

// sortedKeys returns the object keys in lexicographic order, which is the order
// the real ListObjectsV2 returns them in.
func (f *FakeS3Client) sortedKeys() []string {
	keys := make([]string, 0, len(f.Objects))
	for key := range f.Objects {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (f *FakeS3Client) lastModified(key string) time.Time {
	if t, ok := f.LastModified[key]; ok {
		return t
	}
	return fakeS3LastModified
}

func (f *FakeS3Client) ListObjectsV2(_ context.Context, params *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	if params.Bucket == nil {
		return nil, fmt.Errorf("Bucket is required")
	}
	if *params.Bucket != f.Bucket {
		// Real S3 returns *types.NoSuchBucket.
		return nil, fmt.Errorf("bucket not found: %s", *params.Bucket)
	}
	if f.ListObjectsV2Err != nil {
		return nil, f.ListObjectsV2Err
	}

	pageSize := f.pageSize()
	if params.MaxKeys != nil {
		if maxKeys := int(*params.MaxKeys); maxKeys < pageSize {
			pageSize = maxKeys
		}
	}

	start := 0
	if params.ContinuationToken != nil {
		parsed, err := strconv.Atoi(*params.ContinuationToken)
		if err != nil {
			return nil, fmt.Errorf("invalid continuation token: %s", *params.ContinuationToken)
		}
		start = parsed
	}

	keys := f.sortedKeys()
	if start > len(keys) {
		start = len(keys)
	}
	end := min(start+pageSize, len(keys))

	contents := make([]s3Types.Object, 0, end-start)
	for _, key := range keys[start:end] {
		contents = append(contents, s3Types.Object{
			Key:          aws.String(key),
			LastModified: aws.Time(f.lastModified(key)),
			Size:         aws.Int64(int64(len(f.Objects[key]))),
		})
	}

	out := &s3.ListObjectsV2Output{
		Contents:    contents,
		KeyCount:    aws.Int32(int32(len(contents))),
		IsTruncated: aws.Bool(end < len(keys)),
	}
	// The paginator only follows a page when IsTruncated is true and the token
	// is non-empty, so the two are always set together.
	if end < len(keys) {
		out.NextContinuationToken = aws.String(strconv.Itoa(end))
	}

	return out, nil
}

func (f *FakeS3Client) DownloadObject(_ context.Context, params *transfermanager.DownloadObjectInput, _ ...func(*transfermanager.Options)) (*transfermanager.DownloadObjectOutput, error) {
	if params.Bucket == nil || params.Key == nil {
		return nil, fmt.Errorf("Bucket and Key are required")
	}
	if *params.Bucket != f.Bucket {
		// Real S3 returns *types.NoSuchBucket.
		return nil, fmt.Errorf("bucket not found: %s", *params.Bucket)
	}
	if f.DownloadObjectErr != nil {
		return nil, f.DownloadObjectErr
	}
	content, ok := f.Objects[*params.Key]
	if !ok {
		// Real S3 returns *types.NoSuchKey.
		return nil, fmt.Errorf("object not found: %s", *params.Key)
	}
	if params.WriterAt == nil {
		return nil, fmt.Errorf("WriterAt is required")
	}
	written, err := params.WriterAt.WriteAt(content, 0)
	if err != nil {
		return nil, err
	}

	return &transfermanager.DownloadObjectOutput{
		ContentLength: aws.Int64(int64(written)),
		LastModified:  aws.Time(f.lastModified(*params.Key)),
	}, nil
}
