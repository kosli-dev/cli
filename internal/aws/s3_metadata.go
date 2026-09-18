package aws

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/kosli-dev/cli/internal/logger"
)

// GetS3DataFromMetadata returns a digest and metadata of the S3 bucket content,
// taking each object's SHA256 from the checksum S3 stores for it instead of
// downloading the object and hashing it.
//
// The fingerprint is identical to the one GetS3Data produces: both run the
// same pipeline over the same keys and .kosli_ignore rules, and differ only in
// where an object's digest comes from. What this saves is the download, the
// temp disk and the hashing -- not permissions: AWS requires s3:GetObject to
// read an object's checksum, the same permission downloading it needs.
//
// Every object that contributes to the fingerprint must carry a full-object
// SHA256 checksum, which S3 only stores when the upload asked for one. A root
// .kosli_ignore is still downloaded, because its rules decide which objects
// contribute; objects the rules exclude are never fetched, so they need no
// checksum.
func (staticCreds *AWSStaticCreds) GetS3DataFromMetadata(bucket string, includePaths, includeRegex, excludePaths, excludeRegex []string, limits DownloadLimits, logger *logger.Logger) ([]*S3Data, error) {
	client, err := NewS3ClientFunc(staticCreds)
	if err != nil {
		return []*S3Data{}, err
	}
	return getS3DataFromMetadataClient(client, bucket, includePaths, includeRegex, excludePaths, excludeRegex, limits, logger)
}

// getS3DataFromMetadataClient harvests bucket content using the provided client,
// reading digests from stored checksums.
func getS3DataFromMetadataClient(client S3API, bucket string, includePaths, includeRegex, excludePaths, excludeRegex []string, limits DownloadLimits, logger *logger.Logger) ([]*S3Data, error) {
	return getS3DataWithSource(client, metadataDigests(client, bucket, logger), bucket, includePaths, includeRegex, excludePaths, excludeRegex, limits, logger)
}

// metadataDigests is the source that reads each object's stored SHA256 with a
// HeadObject and never touches the disk.
func metadataDigests(client S3HeadAPI, bucket string, logger *logger.Logger) s3DigestSource {
	return s3DigestSource{
		sha256: func(ctx context.Context, _ string, object s3Object) (string, error) {
			// S3 only returns a stored checksum when the request asks for it.
			out, err := client.HeadObject(ctx, &s3.HeadObjectInput{
				Bucket:       aws.String(bucket),
				Key:          aws.String(object.key),
				ChecksumMode: s3Types.ChecksumModeEnabled,
			})
			if err != nil {
				return "", fmt.Errorf("failed to read the checksum of object key [%s]: %w. This needs the "+
					"s3:GetObject permission, the same one downloading the object needs; an SSE-KMS object "+
					"also needs kms:GenerateDataKey and kms:Decrypt", object.key, err)
			}
			sha256, err := objectChecksumSha256(bucket, object.key, out)
			if err != nil {
				return "", err
			}
			logger.Debug("object key [%s] -- stored checksum digest: %s", object.key, sha256)
			return sha256, nil
		},
	}
}

// unusableChecksumError says why one object's stored checksum cannot stand in
// for its content digest. It describes the object, not the connection, so the
// fan-out keeps going and reports every such object at once rather than
// stopping at the first: a bucket-wide migration is then one run, not a
// guess-and-retry loop.
type unusableChecksumError struct {
	msg string
}

func (e unusableChecksumError) Error() string { return e.msg }

// objectChecksumSha256 converts one HeadObject result into the hex SHA256 of
// the object's content, or explains why it cannot.
func objectChecksumSha256(bucket, key string, out *s3.HeadObjectOutput) (string, error) {
	if out.ChecksumSHA256 == nil || *out.ChecksumSHA256 == "" {
		return "", unusableChecksumError{fmt.Sprintf("object key [%s] has no SHA256 checksum, so its fingerprint "+
			"cannot be read from S3 metadata. Upload it with one: aws s3api put-object --bucket %s --key %s "+
			"--body <file> --checksum-algorithm SHA256; or fingerprint by downloading the objects instead",
			key, bucket, key)}
	}

	// A composite checksum hashes the part checksums rather than the object, so
	// it is not the object's digest. S3 reports it two ways -- an explicit
	// COMPOSITE type, and a "-N" part-count suffix on the value -- and the SDK's
	// own response validation keys off the "-". Check both, so neither a missing
	// type nor a missing suffix lets a composite through.
	checksum := *out.ChecksumSHA256
	if out.ChecksumType == s3Types.ChecksumTypeComposite || strings.Contains(checksum, "-") {
		return "", unusableChecksumError{fmt.Sprintf("object key [%s] has a multipart (composite) SHA256 checksum "+
			"[%s], which hashes the part checksums rather than the object content. Collapse it into a single "+
			"part in place: aws s3api copy-object --checksum-algorithm SHA256 --copy-source %s/%s --bucket %s "+
			"--key %s; or fingerprint by downloading the objects instead",
			key, checksum, bucket, key, bucket, key)}
	}

	sha256, err := decodeBase64Sha256(checksum)
	if err != nil {
		return "", unusableChecksumError{fmt.Sprintf("object key [%s] has an SHA256 checksum that cannot be decoded [%s]: %v",
			key, checksum, err)}
	}
	return sha256, nil
}

// combineUnusableChecksumErrors reports every object whose checksum cannot be
// used, capped like key problems are so a whole-bucket problem stays readable.
func combineUnusableChecksumErrors(errs []error) error {
	// The workers finish in any order; sorting keeps the message stable.
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		messages = append(messages, err.Error())
	}
	sort.Strings(messages)

	if len(messages) == 1 {
		return errors.New(messages[0])
	}
	shown, suffix := messages, ""
	if len(shown) > maxReportedS3KeyProblems {
		shown = shown[:maxReportedS3KeyProblems]
		suffix = fmt.Sprintf("\n(and %d more)", len(messages)-maxReportedS3KeyProblems)
	}
	return fmt.Errorf("%d objects cannot be fingerprinted from S3 metadata:\n%s%s",
		len(messages), strings.Join(shown, "\n"), suffix)
}
