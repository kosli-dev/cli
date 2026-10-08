package aws

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type S3MetadataTestSuite struct {
	suite.Suite
}

// fullObjectChecksums is what S3 stores for objects uploaded single-part with
// --checksum-algorithm SHA256.
func fullObjectChecksums(objects map[string][]byte) map[string]FakeS3Checksum {
	checksums := map[string]FakeS3Checksum{}
	for key, content := range objects {
		if strings.HasSuffix(key, "/") {
			continue // folder markers carry no checksum
		}
		checksums[key] = FakeS3Checksum{SHA256: base64Sha256(content), Type: s3Types.ChecksumTypeFullObject}
	}
	return checksums
}

// checksummedBucket is a fake whose every object carries a full-object SHA256.
func checksummedBucket(objects map[string][]byte) *FakeS3Client {
	return &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: objects, Checksums: fullObjectChecksums(objects)}
}

func snapshotMetadata(t *testing.T, client S3API, excludePaths []string) ([]*S3Data, error) {
	t.Helper()
	return getS3DataFromMetadataClient(client, fakeS3TestBucketName, nil, nil, excludePaths, nil,
		DefaultDownloadLimits, logger.NewStandardLogger())
}

// TestMatchesContentMode is the property the feature rests on: for the same
// bucket, the stored checksums must fingerprint to exactly what downloading and
// hashing produces -- digest, artifact name and timestamp. Both sources run the
// shared pipeline, so the only thing that can differ is the per-object digest,
// and the checksum of a full-object upload is that digest.
func (suite *S3MetadataTestSuite) TestMatchesContentMode() {
	for _, t := range []struct {
		name         string
		objects      map[string][]byte
		excludePaths []string
	}{
		{name: "a single object", objects: map[string][]byte{"README.md": []byte(fakeReadmeBody)}},
		{name: "a single nested object", objects: map[string][]byte{"dummy/dummy_2/template.yml": []byte(fakeTemplateBody)}},
		{name: "two objects at the root", objects: map[string][]byte{"README.md": []byte(fakeReadmeBody), "notes.txt": []byte(fakeNotesBody)}},
		{
			name: "objects nested under prefixes with folder markers",
			objects: map[string][]byte{
				"dir/": nil, "dir/sub/": nil, "dir/sub/x.yml": []byte("x"), "dir/y.txt": []byte("y"), "README.md": []byte("r"),
			},
		},
		{
			// '.' sorts before '/', so a flat key sort would order these
			// differently from the directory walk the download path uses.
			name:    "a prefix sharing a name prefix with a sibling object",
			objects: map[string][]byte{"a.txt": []byte("1"), "a/z": []byte("2"), "a/b/c": []byte("3"), "b": []byte("4")},
		},
		{
			// The key rule is shared, so keys fold the same way in both modes.
			name:    "unusual key shapes",
			objects: map[string][]byte{"/lead.txt": []byte("u\n"), "a//b": []byte("o\n"), "./c.txt": []byte("t\n"), `d\e.txt`: []byte("n\n")},
		},
		{
			name: "a root .kosli_ignore whose rules exclude objects",
			objects: map[string][]byte{
				".kosli_ignore": []byte("logs\n*.tmp\n"), "app.js": []byte("app"), "lib/util.js": []byte("util"),
				"logs/a.log": []byte("a"), "logs/deep/b.log": []byte("b"), "scratch.tmp": []byte("tmp"),
			},
		},
		{name: "a lone .kosli_ignore", objects: map[string][]byte{".kosli_ignore": []byte("logs\n")}},
		{name: "a nested .kosli_ignore is an ordinary object", objects: map[string][]byte{"README.md": []byte("r"), "vendor/.kosli_ignore": []byte("README.md\n")}},
		{
			name:         "filters apply before either source",
			objects:      map[string][]byte{"README.md": []byte("r"), "filtered/out.txt": []byte("o"), "keep/in.txt": []byte("i")},
			excludePaths: []string{"filtered/"},
		},
	} {
		suite.Run(t.name, func() {
			content, err := getS3DataFromClient(checksummedBucket(t.objects), fakeS3TestBucketName, nil, nil, t.excludePaths, nil,
				DefaultDownloadLimits, logger.NewStandardLogger())
			require.NoError(suite.T(), err)

			metadata, err := snapshotMetadata(suite.T(), checksummedBucket(t.objects), t.excludePaths)
			require.NoError(suite.T(), err)

			require.Equal(suite.T(), content, metadata, "metadata mode must produce exactly what content mode produces")
		})
	}
}

// TestPinnedFingerprints re-derives, from checksums alone, the fingerprints
// recorded before content mode stopped writing objects under their keys. They
// are the values already on the server, so metadata mode must hit them too.
func (suite *S3MetadataTestSuite) TestPinnedFingerprints() {
	for _, t := range []struct {
		name             string
		objects          map[string][]byte
		wantArtifactName string
		wantFingerprint  string
	}{
		{
			name:             "unusual key shapes fold as filepath.Join folded them",
			objects:          map[string][]byte{"/lead.txt": []byte("u\n"), "a//b": []byte("o\n"), "./c.txt": []byte("t\n"), `d\e.txt`: []byte("n\n")},
			wantArtifactName: fakeS3TestBucketName,
			wantFingerprint:  "27e2d8aa07677b7818b8cf101b45c928aa8fbf7d17a8c8efc84469e24a106ec3",
		},
		{
			name:             "a dot sorts before a slash",
			objects:          map[string][]byte{"a.txt": []byte("1"), "a/z": []byte("2"), "a/b/c": []byte("3"), "b": []byte("4")},
			wantArtifactName: fakeS3TestBucketName,
			wantFingerprint:  "aaddb8f3e299e12316d42fa9ac36d4ed7d38c34e7ec002239269c86a033bb9fd",
		},
		{
			name:             "nested prefixes with folder markers",
			objects:          map[string][]byte{"dir/": nil, "dir/sub/": nil, "dir/sub/x.yml": []byte("x"), "dir/y.txt": []byte("y"), "README.md": []byte("r")},
			wantArtifactName: fakeS3TestBucketName,
			wantFingerprint:  "c26910cdb177dde3c6493d18ba1e916b04715cdfc535e4552b5f9831590e933a",
		},
		{
			name:             "a single object is the file, named by its base name",
			objects:          map[string][]byte{"only/one/file.bin": []byte("solo")},
			wantArtifactName: "file.bin",
			wantFingerprint:  "5364f2f2fc4f54e9d47ad29cfb08ef430c8153394bf2a0dff5cbe77a0ffef861",
		},
	} {
		suite.Run(t.name, func() {
			data, err := snapshotMetadata(suite.T(), checksummedBucket(t.objects), nil)
			require.NoError(suite.T(), err)
			require.Len(suite.T(), data, 1)
			require.Equal(suite.T(), map[string]string{t.wantArtifactName: t.wantFingerprint}, data[0].Digests)
		})
	}
}

// countingHeader records HeadObject calls and their peak overlap, and can delay
// each one so the overlap is real.
type countingHeader struct {
	S3API
	delay time.Duration

	mu          sync.Mutex
	calls       map[string]int
	inFlight    int
	maxInFlight int
	// checksumModeMissing counts requests that forgot to ask for the checksum.
	checksumModeMissing int
}

func (h *countingHeader) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	h.mu.Lock()
	if h.calls == nil {
		h.calls = map[string]int{}
	}
	h.calls[*params.Key]++
	if params.ChecksumMode != s3Types.ChecksumModeEnabled {
		h.checksumModeMissing++
	}
	h.inFlight++
	h.maxInFlight = max(h.maxInFlight, h.inFlight)
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		h.inFlight--
		h.mu.Unlock()
	}()
	time.Sleep(h.delay)
	return h.S3API.HeadObject(ctx, params, optFns...)
}

// TestReadsMetadataNotContent pins what leaves the bucket: one HeadObject per
// object that contributes to the fingerprint, and a download of nothing but the
// root .kosli_ignore, whose rules decide what contributes.
func (suite *S3MetadataTestSuite) TestReadsMetadataNotContent() {
	objects := map[string][]byte{
		".kosli_ignore": []byte("logs\n"), "app.js": []byte("app"), "lib/util.js": []byte("util"),
		"logs/a.log": []byte("a"), "lib/": nil, "filtered/out.txt": []byte("out"),
	}
	checksums := fullObjectChecksums(objects)
	// Neither the ignore file (downloaded) nor an excluded object (never
	// fetched) needs a stored checksum.
	delete(checksums, ".kosli_ignore")
	delete(checksums, "logs/a.log")
	delete(checksums, "filtered/out.txt")
	downloads := &recordingDownloader{S3API: &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: objects, Checksums: checksums}}
	client := &countingHeader{S3API: downloads}

	data, err := getS3DataFromMetadataClient(client, fakeS3TestBucketName, nil, nil, []string{"filtered/"}, nil,
		DefaultDownloadLimits, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	require.Len(suite.T(), data, 1)
	require.Equal(suite.T(), []string{".kosli_ignore"}, downloads.downloadedKeys(), "only the ignore file may be downloaded")
	require.Equal(suite.T(), map[string]int{"app.js": 1, "lib/util.js": 1}, client.calls, "exactly one HeadObject per contributing object")
	require.Zero(suite.T(), client.checksumModeMissing, "every HeadObject must ask for the stored checksum")
}

// A source that reads metadata occupies no temp disk, so the byte budget must
// not throttle it: with a one-byte budget the HEADs still overlap up to the
// worker count.
func (suite *S3MetadataTestSuite) TestHeadsAreBoundedByConcurrencyNotBytes() {
	objects := map[string][]byte{}
	for i := 0; i < 40; i++ {
		objects[fmt.Sprintf("dir%d/object-%03d.bin", i%4, i)] = []byte(fmt.Sprintf("%08d", i))
	}
	client := &countingHeader{S3API: checksummedBucket(objects), delay: 5 * time.Millisecond}

	_, err := getS3DataFromMetadataClient(client, fakeS3TestBucketName, nil, nil, nil, nil,
		DownloadLimits{Concurrency: 4, BytesInFlight: 1}, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	require.Len(suite.T(), client.calls, 40)
	require.LessOrEqual(suite.T(), client.maxInFlight, 4)
	require.GreaterOrEqual(suite.T(), client.maxInFlight, 2, "HEADs must actually overlap for the bound to be tested")
}

func (suite *S3MetadataTestSuite) TestErrors() {
	readme := []byte(fakeReadmeBody)
	for _, t := range []struct {
		name       string
		objects    map[string][]byte
		checksums  map[string]FakeS3Checksum
		listErr    error
		headErr    error
		wantErrMsg []string
	}{
		{
			// Pinned in full: the command-level golden for this case mirrors it.
			name:    "an object with no stored checksum",
			objects: map[string][]byte{"README.md": readme},
			wantErrMsg: []string{"object key [README.md] has no SHA256 checksum, so its fingerprint cannot be read from S3 " +
				"metadata. Upload it with one: aws s3api put-object --bucket " + fakeS3TestBucketName + " --key README.md " +
				"--body <file> --checksum-algorithm SHA256; or fingerprint by downloading the objects instead"},
		},
		{
			name:       "a composite multipart checksum",
			objects:    map[string][]byte{"README.md": readme},
			checksums:  map[string]FakeS3Checksum{"README.md": {SHA256: base64Sha256(readme) + "-4", Type: s3Types.ChecksumTypeComposite}},
			wantErrMsg: []string{"multipart (composite) SHA256 checksum", "aws s3api copy-object --checksum-algorithm SHA256"},
		},
		{
			name:       "a composite checksum is rejected on its type even without a suffix",
			objects:    map[string][]byte{"README.md": readme},
			checksums:  map[string]FakeS3Checksum{"README.md": {SHA256: base64Sha256(readme), Type: s3Types.ChecksumTypeComposite}},
			wantErrMsg: []string{"multipart (composite) SHA256 checksum"},
		},
		{
			name:       "a checksum that is not valid Base64",
			objects:    map[string][]byte{"README.md": readme},
			checksums:  map[string]FakeS3Checksum{"README.md": {SHA256: "not base64!", Type: s3Types.ChecksumTypeFullObject}},
			wantErrMsg: []string{"object key [README.md]", "not base64!"},
		},
		{
			name:       "a metadata request failure names the permission it needs",
			objects:    map[string][]byte{"README.md": readme},
			checksums:  fullObjectChecksums(map[string][]byte{"README.md": readme}),
			headErr:    errInjected,
			wantErrMsg: []string{"injected error", "s3:GetObject"},
		},
		{
			name:       "a listing error propagates",
			objects:    map[string][]byte{"README.md": readme},
			checksums:  fullObjectChecksums(map[string][]byte{"README.md": readme}),
			listErr:    errInjected,
			wantErrMsg: []string{"injected error"},
		},
		{
			name:       "an empty bucket keeps the content-mode message",
			objects:    map[string][]byte{"dir/": nil},
			wantErrMsg: []string{"no matching file or dirs in bucket: [" + fakeS3TestBucketName + "]"},
		},
	} {
		suite.Run(t.name, func() {
			client := &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: t.objects, Checksums: t.checksums,
				ListObjectsV2Err: t.listErr, HeadObjectErr: t.headErr}
			_, err := snapshotMetadata(suite.T(), client, nil)
			require.Error(suite.T(), err)
			for _, want := range t.wantErrMsg {
				require.Contains(suite.T(), err.Error(), want)
			}
		})
	}
}

// Whether an object's checksum is usable is a property of the object, not of
// the connection, so one run reports every such object rather than stopping at
// the first; a bucket-wide migration is then not a guess-and-retry loop.
func (suite *S3MetadataTestSuite) TestReportsEveryUnusableObjectTogether() {
	objects := map[string][]byte{}
	for i := 0; i < 6; i++ {
		objects[fmt.Sprintf("object-%d.txt", i)] = []byte(fmt.Sprintf("content %d", i))
	}
	checksums := fullObjectChecksums(objects)
	delete(checksums, "object-0.txt")
	delete(checksums, "object-2.txt")
	checksums["object-4.txt"] = FakeS3Checksum{SHA256: checksums["object-4.txt"].SHA256 + "-3", Type: s3Types.ChecksumTypeComposite}

	_, err := snapshotMetadata(suite.T(), &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: objects, Checksums: checksums}, nil)
	require.Error(suite.T(), err)
	for _, key := range []string{"object-0.txt", "object-2.txt", "object-4.txt"} {
		require.Contains(suite.T(), err.Error(), "["+key+"]")
	}
	for _, key := range []string{"object-1.txt", "object-3.txt", "object-5.txt"} {
		require.NotContains(suite.T(), err.Error(), "["+key+"]", "objects that are fine must not be named")
	}
	require.Contains(suite.T(), err.Error(), "no SHA256 checksum")
	require.Contains(suite.T(), err.Error(), "multipart (composite)")

	suite.Run("the list is capped", func() {
		many := map[string][]byte{}
		for i := 0; i < 40; i++ {
			many[fmt.Sprintf("object-%02d.txt", i)] = []byte("x")
		}
		_, err := snapshotMetadata(suite.T(), &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: many}, nil)
		require.Error(suite.T(), err)
		require.Contains(suite.T(), err.Error(), "(and 30 more)")
		require.Equal(suite.T(), maxReportedS3KeyProblems, strings.Count(err.Error(), "has no SHA256 checksum"))
	})

	suite.Run("a transport error still stops the run", func() {
		client := &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: objects, Checksums: checksums, HeadObjectErr: errInjected}
		_, err := snapshotMetadata(suite.T(), client, nil)
		require.ErrorIs(suite.T(), err, errInjected)
		require.NotContains(suite.T(), err.Error(), "no SHA256 checksum", "a connection failure is not a per-object report")
	})
}

func TestS3MetadataTestSuite(t *testing.T) {
	suite.Run(t, new(S3MetadataTestSuite))
}
