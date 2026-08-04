package aws

import (
	"testing"
	"time"

	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type S3MetadataTestSuite struct {
	suite.Suite
}

// fullObjectChecksums builds the checksum map S3 would report for objects
// uploaded single-part with --checksum-algorithm SHA256.
func fullObjectChecksums(objects map[string][]byte) map[string]FakeS3Checksum {
	checksums := map[string]FakeS3Checksum{}
	for key, content := range objects {
		checksums[key] = FakeS3Checksum{
			SHA256: base64Sha256(content),
			Type:   s3Types.ChecksumTypeFullObject,
		}
	}
	return checksums
}

func (suite *S3MetadataTestSuite) TestGetS3DataFromMetadataClient() {
	earlier := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	later := time.Date(2024, 3, 20, 8, 0, 0, 0, time.UTC)

	readme := []byte(fakeReadmeBody)
	for _, t := range []struct {
		name             string
		objects          map[string][]byte
		checksums        map[string]FakeS3Checksum
		lastModified     map[string]time.Time
		includePaths     []string
		excludePaths     []string
		listErr          error
		headErr          error
		wantArtifactName string
		wantFingerprint  string
		wantLastModified int64
		wantErr          bool
		wantErrMsg       string
	}{
		{
			name:             "a single object is fingerprinted from its stored checksum",
			objects:          map[string][]byte{"README.md": readme},
			checksums:        fullObjectChecksums(map[string][]byte{"README.md": readme}),
			wantArtifactName: "README.md",
			wantFingerprint:  fakeReadmeSha256,
		},
		{
			name:             "a single nested object keeps its base name",
			objects:          map[string][]byte{"dummy/dummy_2/template.yml": []byte(fakeTemplateBody)},
			checksums:        fullObjectChecksums(map[string][]byte{"dummy/dummy_2/template.yml": []byte(fakeTemplateBody)}),
			wantArtifactName: "template.yml",
			wantFingerprint:  fakeTemplateSha256,
		},
		{
			name:             "the newest matched object sets the timestamp",
			objects:          map[string][]byte{"README.md": readme},
			checksums:        fullObjectChecksums(map[string][]byte{"README.md": readme}),
			lastModified:     map[string]time.Time{"README.md": later},
			wantArtifactName: "README.md",
			wantLastModified: later.Unix(),
		},
		{
			name:    "folder markers are skipped",
			objects: map[string][]byte{"dummy/": nil, "README.md": readme},
			checksums: fullObjectChecksums(map[string][]byte{
				"README.md": readme,
			}),
			lastModified:     map[string]time.Time{"README.md": earlier},
			wantArtifactName: "README.md",
			wantFingerprint:  fakeReadmeSha256,
		},
		{
			name:             "filters select the object to fingerprint",
			objects:          map[string][]byte{"README.md": readme, "notes.txt": []byte(fakeNotesBody)},
			checksums:        fullObjectChecksums(map[string][]byte{"README.md": readme, "notes.txt": []byte(fakeNotesBody)}),
			includePaths:     []string{"README.md"},
			wantArtifactName: "README.md",
			wantFingerprint:  fakeReadmeSha256,
		},
		{
			name:      "an object with no stored checksum is an error",
			objects:   map[string][]byte{"README.md": readme},
			checksums: nil,
			wantErr:   true,
			// The message has to say how to fix it, not just what is wrong.
			wantErrMsg: "has no SHA256 checksum",
		},
		{
			name:    "a composite multipart checksum is an error",
			objects: map[string][]byte{"README.md": readme},
			checksums: map[string]FakeS3Checksum{
				"README.md": {SHA256: base64Sha256(readme) + "-4", Type: s3Types.ChecksumTypeComposite},
			},
			wantErr:    true,
			wantErrMsg: "multipart (composite) SHA256 checksum",
		},
		{
			name:    "a composite checksum is rejected on its type even without a suffix",
			objects: map[string][]byte{"README.md": readme},
			checksums: map[string]FakeS3Checksum{
				"README.md": {SHA256: base64Sha256(readme), Type: s3Types.ChecksumTypeComposite},
			},
			wantErr:    true,
			wantErrMsg: "multipart (composite) SHA256 checksum",
		},
		{
			name:    "a checksum that is not valid Base64 is an error",
			objects: map[string][]byte{"README.md": readme},
			checksums: map[string]FakeS3Checksum{
				"README.md": {SHA256: "not base64!", Type: s3Types.ChecksumTypeFullObject},
			},
			wantErr:    true,
			wantErrMsg: "README.md",
		},
		{
			name:             "several objects are fingerprinted as a directory named after the bucket",
			objects:          map[string][]byte{"README.md": readme, "notes.txt": []byte(fakeNotesBody)},
			checksums:        fullObjectChecksums(map[string][]byte{"README.md": readme, "notes.txt": []byte(fakeNotesBody)}),
			lastModified:     map[string]time.Time{"README.md": earlier, "notes.txt": later},
			wantArtifactName: fakeS3TestBucketName,
			wantLastModified: later.Unix(),
		},
		{
			// Reading it would need the object's content, so metadata mode
			// cannot apply its rules and must not silently ignore them.
			name: "a root .kosli_ignore is an error",
			objects: map[string][]byte{
				"README.md":     readme,
				".kosli_ignore": []byte("notes.txt\n"),
				"notes.txt":     []byte(fakeNotesBody),
			},
			checksums: fullObjectChecksums(map[string][]byte{
				"README.md":     readme,
				".kosli_ignore": []byte("notes.txt\n"),
				"notes.txt":     []byte(fakeNotesBody),
			}),
			wantErr:    true,
			wantErrMsg: ".kosli_ignore",
		},
		{
			name: "a nested .kosli_ignore is an ordinary object",
			objects: map[string][]byte{
				"README.md":           readme,
				"dummy/.kosli_ignore": []byte("README.md\n"),
			},
			checksums: fullObjectChecksums(map[string][]byte{
				"README.md":           readme,
				"dummy/.kosli_ignore": []byte("README.md\n"),
			}),
			wantArtifactName: fakeS3TestBucketName,
		},
		{
			name:      "an object key that is also a directory prefix is an error",
			objects:   map[string][]byte{"a": readme, "a/b": []byte(fakeNotesBody)},
			checksums: fullObjectChecksums(map[string][]byte{"a": readme, "a/b": []byte(fakeNotesBody)}),
			wantErr:   true,
			// digest rejects it; the message must name the clashing path.
			wantErrMsg: "both a file and a directory",
		},
		{
			// S3 allows these keys, but they cannot be mapped onto a directory
			// tree unambiguously: content mode silently collapses "a//b" onto
			// "a/b" via filepath.Join, which can even collide with a real
			// "a/b" object. Refusing beats a quietly wrong fingerprint, and
			// the help text says so.
			name:      "an object key with an empty path segment is an error",
			objects:   map[string][]byte{"a//b": readme, "c.txt": []byte(fakeNotesBody)},
			checksums: fullObjectChecksums(map[string][]byte{"a//b": readme, "c.txt": []byte(fakeNotesBody)}),
			wantErr:   true,
			// The wrapper must name the bucket and keep the underlying reason.
			wantErrMsg: "not a clean relative path",
		},
		{
			name:       "an absolute object key is an error",
			objects:    map[string][]byte{"/foo": readme, "c.txt": []byte(fakeNotesBody)},
			checksums:  fullObjectChecksums(map[string][]byte{"/foo": readme, "c.txt": []byte(fakeNotesBody)}),
			wantErr:    true,
			wantErrMsg: "not a clean relative path",
		},
		{
			name:       "an empty bucket is an error",
			objects:    map[string][]byte{"dummy/": nil},
			wantErr:    true,
			wantErrMsg: "no matching file or dirs in bucket: [" + fakeS3TestBucketName + "]",
		},
		{
			name:         "filtering everything out is an error",
			objects:      map[string][]byte{"README.md": readme},
			checksums:    fullObjectChecksums(map[string][]byte{"README.md": readme}),
			includePaths: []string{"non-existing.md"},
			wantErr:      true,
			wantErrMsg:   "no matching file or dirs in bucket: [" + fakeS3TestBucketName + "]",
		},
		{
			name:       "a listing error propagates",
			objects:    map[string][]byte{"README.md": readme},
			checksums:  fullObjectChecksums(map[string][]byte{"README.md": readme}),
			listErr:    errInjected,
			wantErr:    true,
			wantErrMsg: "injected error",
		},
		{
			name:       "a metadata error propagates",
			objects:    map[string][]byte{"README.md": readme},
			checksums:  fullObjectChecksums(map[string][]byte{"README.md": readme}),
			headErr:    errInjected,
			wantErr:    true,
			wantErrMsg: "injected error",
		},
	} {
		suite.Run(t.name, func() {
			client := &FakeS3Client{
				Bucket:           fakeS3TestBucketName,
				Objects:          t.objects,
				Checksums:        t.checksums,
				LastModified:     t.lastModified,
				ListObjectsV2Err: t.listErr,
				HeadObjectErr:    t.headErr,
			}

			data, err := getS3DataFromMetadataClient(client, fakeS3TestBucketName, t.includePaths,
				nil, t.excludePaths, nil, logger.NewStandardLogger())

			if t.wantErr {
				require.Error(suite.T(), err)
				if t.wantErrMsg != "" {
					require.Contains(suite.T(), err.Error(), t.wantErrMsg)
				}
				return
			}
			require.NoError(suite.T(), err)
			require.Len(suite.T(), data, 1)

			wantArtifactName := t.wantArtifactName
			if wantArtifactName == "" {
				wantArtifactName = fakeS3TestBucketName
			}
			require.Contains(suite.T(), data[0].Digests, wantArtifactName)
			if t.wantFingerprint != "" {
				require.Equal(suite.T(), t.wantFingerprint, data[0].Digests[wantArtifactName])
			}
			if t.wantLastModified != 0 {
				require.Equal(suite.T(), t.wantLastModified, data[0].LastModifiedTimestamp)
			}
		})
	}
}

// TestMetadataMatchesDownloadFingerprint is the test the whole feature rests on:
// for the same bucket, fingerprinting from stored checksums must produce exactly
// what downloading and hashing produces. If these ever diverge, a snapshot stops
// matching the artifact that was attested.
func (suite *S3MetadataTestSuite) TestMetadataMatchesDownloadFingerprint() {
	for _, t := range []struct {
		name         string
		objects      map[string][]byte
		excludePaths []string
	}{
		{
			name:    "a single object",
			objects: map[string][]byte{"README.md": []byte(fakeReadmeBody)},
		},
		{
			name: "a single nested object",
			objects: map[string][]byte{
				"dummy/dummy_2/template.yml": []byte(fakeTemplateBody),
			},
		},
		{
			name: "two objects at the bucket root",
			objects: map[string][]byte{
				"README.md": []byte(fakeReadmeBody),
				"notes.txt": []byte(fakeNotesBody),
			},
		},
		{
			name: "objects nested under prefixes",
			objects: map[string][]byte{
				"README.md":                  []byte(fakeReadmeBody),
				"dummy/dummy_2/template.yml": []byte(fakeTemplateBody),
				"dummy/notes.txt":            []byte(fakeNotesBody),
			},
		},
		{
			// A lone .kosli_ignore is fingerprinted as a file, so its rules
			// never apply and metadata mode can handle it like any object.
			name:    "a lone .kosli_ignore object",
			objects: map[string][]byte{".kosli_ignore": []byte("notes.txt\n")},
		},
		{
			// '.' sorts before '/', so a flat key sort would order these
			// differently from the directory walk the download path uses.
			name: "a prefix sharing a name prefix with a sibling object",
			objects: map[string][]byte{
				"a.txt":   []byte(fakeReadmeBody),
				"a/z.txt": []byte(fakeNotesBody),
				"b.txt":   []byte(fakeTemplateBody),
			},
		},
		{
			// Excluding a root .kosli_ignore makes the two sources agree again:
			// content mode never downloads it, so DirSha256 finds no ignore file
			// and applies no rules, which is what metadata mode does too. The
			// error message for an un-excluded .kosli_ignore says as much, so
			// pin it here rather than leaving the claim untested.
			name: "an excluded root .kosli_ignore",
			objects: map[string][]byte{
				".kosli_ignore": []byte("notes.txt\n"),
				"README.md":     []byte(fakeReadmeBody),
				"notes.txt":     []byte(fakeNotesBody),
			},
			excludePaths: []string{".kosli_ignore"},
		},
	} {
		suite.Run(t.name, func() {
			newClient := func() *FakeS3Client {
				return &FakeS3Client{
					Bucket:    fakeS3TestBucketName,
					Objects:   t.objects,
					Checksums: fullObjectChecksums(t.objects),
				}
			}

			downloaded, err := getS3DataFromClient(newClient(), fakeS3TestBucketName,
				nil, nil, t.excludePaths, nil, logger.NewStandardLogger())
			require.NoError(suite.T(), err)

			fromMetadata, err := getS3DataFromMetadataClient(newClient(), fakeS3TestBucketName,
				nil, nil, t.excludePaths, nil, logger.NewStandardLogger())
			require.NoError(suite.T(), err)

			require.Equal(suite.T(), downloaded, fromMetadata,
				"metadata mode must produce the same artifacts as content mode")
		})
	}
}

func TestS3MetadataTestSuite(t *testing.T) {
	suite.Run(t, new(S3MetadataTestSuite))
}
