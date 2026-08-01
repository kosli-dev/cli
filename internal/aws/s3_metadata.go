package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/logger"
)

// GetS3DataFromMetadata returns a digest and metadata of the S3 bucket content,
// fingerprinting it from the SHA256 checksums S3 stores for each object instead
// of downloading the objects and hashing them.
//
// The fingerprint is identical to the one GetS3Data produces, so a snapshot
// still matches the artifact that was attested. This saves the download, the
// local disk and the local hashing, but not permissions: reading an object's
// checksum needs s3:GetObject, the same permission that downloading it needs.
//
// Every matching object must carry a full-object SHA256 checksum, which S3 only
// stores when the upload asked for one.
func (staticCreds *AWSStaticCreds) GetS3DataFromMetadata(bucket string, includePaths, includeRegex,
	excludePaths, excludeRegex []string, logger *logger.Logger) ([]*S3Data, error) {
	client, err := NewS3ClientFunc(staticCreds)
	if err != nil {
		return []*S3Data{}, err
	}
	return getS3DataFromMetadataClient(client, bucket, includePaths, includeRegex,
		excludePaths, excludeRegex, logger)
}

// getS3DataFromMetadataClient harvests bucket content using the provided client.
// It takes S3MetadataAPI rather than S3API so that it cannot read object content.
func getS3DataFromMetadataClient(client S3MetadataAPI, bucket string, includePaths, includeRegex,
	excludePaths, excludeRegex []string, logger *logger.Logger) ([]*S3Data, error) {
	s3Data := []*S3Data{}

	includeRegexCompiled, err := compilePathRegex(includeRegex)
	if err != nil {
		return s3Data, err
	}
	excludeRegexCompiled, err := compilePathRegex(excludeRegex)
	if err != nil {
		return s3Data, err
	}

	objects, err := listMatchingS3Objects(client, bucket, includePaths, includeRegexCompiled,
		excludePaths, excludeRegexCompiled)
	if err != nil {
		return s3Data, err
	}

	if err := rejectKosliIgnore(objects, bucket); err != nil {
		return s3Data, err
	}

	files := make([]digest.VirtualFile, 0, len(objects))
	newest := objects[0].lastModified
	for _, object := range objects {
		sha256, err := objectSha256FromMetadata(client, bucket, object.key, logger)
		if err != nil {
			return s3Data, err
		}
		files = append(files, digest.VirtualFile{Path: object.key, Sha256: sha256})
		if object.lastModified.After(newest) {
			newest = object.lastModified
		}
	}

	// One object is fingerprinted as that file and named after it; several are
	// fingerprinted as a directory named after the bucket. This mirrors what
	// containsSingleFile decides once the objects are on disk.
	artifactName := bucket
	var sha256 string
	if file, ok := digest.SingleVirtualFile(files); ok {
		artifactName = file.Name()
		sha256 = file.Sha256
	} else {
		sha256, err = digest.VirtualDirSha256(files, logger)
		if err != nil {
			return s3Data, fmt.Errorf("failed to fingerprint bucket [%s] from object metadata: %w", bucket, err)
		}
	}

	s3Data = append(s3Data, &S3Data{
		Digests:               map[string]string{artifactName: sha256},
		LastModifiedTimestamp: newest.Unix(),
	})
	return s3Data, nil
}

// kosliIgnoreFile is read from the root of a directory artifact by
// digest.DirSha256, and its rules change the fingerprint.
const kosliIgnoreFile = ".kosli_ignore"

// rejectKosliIgnore fails when the selection contains a bucket-root
// .kosli_ignore. Applying its rules needs the object's content, which metadata
// mode does not read, and ignoring them would silently produce a fingerprint
// that differs from the downloaded one.
//
// Only a root .kosli_ignore matters: DirSha256 reads exactly one, at the root of
// the artifact, and treats any nested one as an ordinary file.
func rejectKosliIgnore(objects []s3Object, bucket string) error {
	for _, object := range objects {
		if object.key != kosliIgnoreFile {
			continue
		}
		if len(objects) == 1 {
			// The only object: it is fingerprinted as a file, and DirSha256's
			// ignore handling never comes into play.
			return nil
		}
		return fmt.Errorf("bucket [%s] has a %s object at its root, and its exclusion rules change the "+
			"fingerprint. Fingerprinting from S3 metadata cannot apply them, because reading the file "+
			"would mean downloading it. Fingerprint by downloading the objects instead. Excluding the "+
			"file with --exclude would not help: the fingerprint would still differ, because the rules "+
			"inside it would go unapplied", bucket, kosliIgnoreFile)
	}
	return nil
}

// objectSha256FromMetadata reads one object's stored SHA256 and returns it as hex.
func objectSha256FromMetadata(client S3HeadAPI, bucket, key string, logger *logger.Logger) (string, error) {
	// S3 only returns a stored checksum when the request asks for it.
	out, err := client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(key),
		ChecksumMode: s3Types.ChecksumModeEnabled,
	})
	if err != nil {
		return "", fmt.Errorf("failed to read checksum metadata for object %q in bucket [%s]: %w. "+
			"This requires the s3:GetObject permission (the same permission needed to fingerprint by "+
			"downloading); an SSE-KMS object also needs kms:GenerateDataKey and kms:Decrypt", key, bucket, err)
	}

	sha256, err := objectChecksumSha256(bucket, key, out)
	if err != nil {
		return "", err
	}
	logger.Debug("object %s -- checksum digest: %s", key, sha256)
	return sha256, nil
}

// objectChecksumSha256 converts one HeadObject result into a hex SHA256 of the
// object's content, or explains why it cannot.
func objectChecksumSha256(bucket, key string, out *s3.HeadObjectOutput) (string, error) {
	if out.ChecksumSHA256 == nil || *out.ChecksumSHA256 == "" {
		return "", fmt.Errorf("object %q in bucket [%s] has no SHA256 checksum, so its fingerprint "+
			"cannot be derived from S3 metadata. Re-upload it with a checksum: "+
			"aws s3api put-object --bucket %s --key %s --body <file> --checksum-algorithm SHA256. "+
			"Or fingerprint by downloading the objects instead", key, bucket, bucket, key)
	}

	// A composite checksum hashes the part checksums rather than the object, so
	// it is not the object's digest. S3 reports it two ways -- an explicit
	// COMPOSITE type, and a "-N" part-count suffix on the value -- and the SDK
	// itself treats a "-" as the marker when deciding whether a checksum can be
	// validated. Check both, so neither a missing type nor a missing suffix
	// lets a composite checksum through.
	checksum := *out.ChecksumSHA256
	if out.ChecksumType == s3Types.ChecksumTypeComposite || strings.Contains(checksum, "-") {
		return "", fmt.Errorf("object %q in bucket [%s] has a multipart (composite) SHA256 checksum "+
			"%q, which hashes the part checksums rather than the object content. Re-upload it as a "+
			"single part with an SHA256 checksum, or copy it in place to collapse the parts: "+
			"aws s3api copy-object --checksum-algorithm SHA256 --copy-source %s/%s --bucket %s --key %s. "+
			"Or fingerprint by downloading the objects instead",
			key, bucket, checksum, bucket, key, bucket, key)
	}

	sha256, err := decodeBase64Sha256(checksum)
	if err != nil {
		return "", fmt.Errorf("failed to decode the SHA256 checksum %q of object %q in bucket [%s]: %w",
			checksum, key, bucket, err)
	}
	return sha256, nil
}
