package aws

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
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

func checksummedBucket(objects map[string][]byte) *FakeS3Client {
	return &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: objects, Checksums: fullObjectChecksums(objects)}
}

func snapshotMetadata(t *testing.T, client S3API, excludePaths []string) ([]*S3Data, error) {
	t.Helper()
	return getS3DataFromMetadataClient(client, fakeS3TestBucketName, nil, nil, excludePaths, nil,
		DefaultDownloadLimits, false, logger.NewStandardLogger())
}

// snapshotSkippingUnusable snapshots with unusable checksums skipped and
// returns what was logged as warnings.
func snapshotSkippingUnusable(t *testing.T, client S3API) ([]*S3Data, string, error) {
	t.Helper()
	var warnings bytes.Buffer
	data, err := getS3DataFromMetadataClient(client, fakeS3TestBucketName, nil, nil, nil, nil,
		DefaultDownloadLimits, true, logger.NewLogger(io.Discard, &warnings, false))
	return data, warnings.String(), err
}

// The property the feature rests on: stored checksums fingerprint a bucket
// exactly as downloading does, artifact name and timestamp included.
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
			// '.' sorts before '/', so a flat key sort would diverge from the tree walk.
			name:    "a prefix sharing a name prefix with a sibling object",
			objects: map[string][]byte{"a.txt": []byte("1"), "a/z": []byte("2"), "a/b/c": []byte("3"), "b": []byte("4")},
		},
		{
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

// Existing snapshots on the server carry these fingerprints, so metadata mode
// must reproduce them from checksums alone.
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
			name:             "a single object is fingerprinted as the file, named after the bucket",
			objects:          map[string][]byte{"only/one/file.bin": []byte("solo")},
			wantArtifactName: fakeS3TestBucketName,
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

	mu                  sync.Mutex
	calls               map[string]int
	inFlight            int
	maxInFlight         int
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
		DefaultDownloadLimits, false, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	require.Len(suite.T(), data, 1)
	require.Equal(suite.T(), []string{".kosli_ignore"}, downloads.downloadedKeys(), "only the ignore file may be downloaded")
	require.Equal(suite.T(), map[string]int{"app.js": 1, "lib/util.js": 1}, client.calls, "exactly one HeadObject per contributing object")
	require.Zero(suite.T(), client.checksumModeMissing, "every HeadObject must ask for the stored checksum")
}

// A lone ignore file takes the single-file branch, not the ignore pass, and must
// still be downloaded rather than read through HeadObject.
func (suite *S3MetadataTestSuite) TestALoneIgnoreFileNeedsNoChecksum() {
	objects := map[string][]byte{".kosli_ignore": []byte("logs\n")}
	downloads := &recordingDownloader{S3API: &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: objects}}
	client := &countingHeader{S3API: downloads}

	metadata, err := getS3DataFromMetadataClient(client, fakeS3TestBucketName, nil, nil, nil, nil,
		DefaultDownloadLimits, false, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), []string{".kosli_ignore"}, downloads.downloadedKeys())
	require.Empty(suite.T(), client.calls, "the ignore file must not be read through HeadObject")

	content, err := getS3DataFromClient(&FakeS3Client{Bucket: fakeS3TestBucketName, Objects: objects}, fakeS3TestBucketName,
		nil, nil, nil, nil, DefaultDownloadLimits, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), content, metadata)
}

// A one-byte budget would serialise downloads; HEADs use no disk and must still overlap.
func (suite *S3MetadataTestSuite) TestHeadsAreBoundedByConcurrencyNotBytes() {
	objects := map[string][]byte{}
	for i := 0; i < 40; i++ {
		objects[fmt.Sprintf("dir%d/object-%03d.bin", i%4, i)] = []byte(fmt.Sprintf("%08d", i))
	}
	client := &countingHeader{S3API: checksummedBucket(objects), delay: 5 * time.Millisecond}

	_, err := getS3DataFromMetadataClient(client, fakeS3TestBucketName, nil, nil, nil, nil,
		DownloadLimits{Concurrency: 4, BytesInFlight: 1}, false, logger.NewStandardLogger())
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
			// Pinned in full because the SnapshotS3TestSuite golden repeats it.
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

func (suite *S3MetadataTestSuite) TestCompositeDetectionKeysOffThePartCountSuffix() {
	readme := []byte(fakeReadmeBody)
	for _, t := range []struct {
		name          string
		checksum      FakeS3Checksum
		wantComposite bool
	}{
		{name: "a part-count suffix with no type", checksum: FakeS3Checksum{SHA256: base64Sha256(readme) + "-4"}, wantComposite: true},
		{name: "a COMPOSITE type with no suffix", checksum: FakeS3Checksum{SHA256: base64Sha256(readme), Type: s3Types.ChecksumTypeComposite}, wantComposite: true},
		{name: "a dash that is not a suffix", checksum: FakeS3Checksum{SHA256: "S0tL-S0tLS0_LS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0=", Type: s3Types.ChecksumTypeFullObject}},
		{name: "a dash followed by more than digits", checksum: FakeS3Checksum{SHA256: base64Sha256(readme) + "-4a"}},
		{name: "a trailing dash", checksum: FakeS3Checksum{SHA256: base64Sha256(readme) + "-"}},
	} {
		suite.Run(t.name, func() {
			client := &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: map[string][]byte{"README.md": readme},
				Checksums: map[string]FakeS3Checksum{"README.md": t.checksum}}
			_, err := snapshotMetadata(suite.T(), client, nil)
			require.Error(suite.T(), err)
			if t.wantComposite {
				require.Contains(suite.T(), err.Error(), "multipart (composite)")
				return
			}
			require.Contains(suite.T(), err.Error(), "cannot be decoded")
			require.NotContains(suite.T(), err.Error(), "copy-object", "an undecodable checksum is not fixed by collapsing parts")
		})
	}
}

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

// Skipping an object must leave exactly the snapshot that excluding it leaves,
// so a skipped object behaves as one the bucket never held.
func (suite *S3MetadataTestSuite) TestSkippingUnusableMatchesExcludingThem() {
	composite := func(content []byte) FakeS3Checksum {
		return FakeS3Checksum{SHA256: base64Sha256(content) + "-2", Type: s3Types.ChecksumTypeComposite}
	}
	undecodable := FakeS3Checksum{SHA256: "not base64!", Type: s3Types.ChecksumTypeFullObject}
	later := fakeS3LastModified.Add(time.Hour)

	for _, t := range []struct {
		name         string
		objects      map[string][]byte
		unusable     map[string]*FakeS3Checksum // nil means no checksum at all
		lastModified map[string]time.Time
	}{
		{
			name:     "an object with no checksum",
			objects:  map[string][]byte{"a.txt": []byte("1"), "dir/b.txt": []byte("2"), "dir/c.txt": []byte("3")},
			unusable: map[string]*FakeS3Checksum{"dir/b.txt": nil},
		},
		{
			name:     "a composite and an undecodable checksum",
			objects:  map[string][]byte{"a.txt": []byte("1"), "big.bin": []byte("2"), "odd.bin": []byte("3"), "z/y.txt": []byte("4")},
			unusable: map[string]*FakeS3Checksum{"big.bin": new(composite([]byte("2"))), "odd.bin": &undecodable},
		},
		{
			name:     "one object left is fingerprinted as that file",
			objects:  map[string][]byte{"a.txt": []byte("1"), "dir/b.txt": []byte("2"), "dir/c.txt": []byte("3")},
			unusable: map[string]*FakeS3Checksum{"a.txt": nil, "dir/c.txt": nil},
		},
		{
			name:     "a whole prefix skipped",
			objects:  map[string][]byte{"a.txt": []byte("1"), "b.txt": []byte("2"), "dir/c.txt": []byte("3"), "dir/sub/d.txt": []byte("4")},
			unusable: map[string]*FakeS3Checksum{"dir/c.txt": nil, "dir/sub/d.txt": nil},
		},
		{
			name: "beside a root .kosli_ignore",
			objects: map[string][]byte{
				".kosli_ignore": []byte("logs\n"), "app.js": []byte("app"), "lib/util.js": []byte("util"), "logs/a.log": []byte("a"),
			},
			unusable: map[string]*FakeS3Checksum{"lib/util.js": nil},
		},
		{
			name:     "only the root .kosli_ignore left",
			objects:  map[string][]byte{".kosli_ignore": []byte("logs\n"), "app.js": []byte("app"), "logs/a.log": []byte("a")},
			unusable: map[string]*FakeS3Checksum{"app.js": nil},
		},
		{
			name:         "the newest object skipped",
			objects:      map[string][]byte{"a.txt": []byte("1"), "b.txt": []byte("2"), "c.txt": []byte("3")},
			unusable:     map[string]*FakeS3Checksum{"b.txt": nil},
			lastModified: map[string]time.Time{"b.txt": later},
		},
	} {
		suite.Run(t.name, func() {
			checksums := fullObjectChecksums(t.objects)
			excludeRegex := []string{}
			for key, checksum := range t.unusable {
				delete(checksums, key)
				if checksum != nil {
					checksums[key] = *checksum
				}
				excludeRegex = append(excludeRegex, "^"+regexp.QuoteMeta(key)+"$")
			}
			client := &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: t.objects, Checksums: checksums, LastModified: t.lastModified}

			excluded, err := getS3DataFromClient(client, fakeS3TestBucketName, nil, nil, nil, excludeRegex,
				DefaultDownloadLimits, logger.NewStandardLogger())
			require.NoError(suite.T(), err)

			skipped, _, err := snapshotSkippingUnusable(suite.T(), client)
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), excluded, skipped)
		})
	}
}

func (suite *S3MetadataTestSuite) TestSkippingUnusableWarnsWithTheSkippedKeys() {
	objects := map[string][]byte{}
	for i := 0; i < 6; i++ {
		objects[fmt.Sprintf("object-%d.txt", i)] = []byte(fmt.Sprintf("content %d", i))
	}
	checksums := fullObjectChecksums(objects)
	delete(checksums, "object-0.txt")
	delete(checksums, "object-2.txt")
	checksums["object-4.txt"] = FakeS3Checksum{SHA256: checksums["object-4.txt"].SHA256 + "-3", Type: s3Types.ChecksumTypeComposite}
	client := &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: objects, Checksums: checksums}

	data, warnings, err := snapshotSkippingUnusable(suite.T(), client)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), data, 1)
	require.Equal(suite.T(), "[warning] skipped 3 of the 6 objects whose checksum was read in bucket ["+fakeS3TestBucketName+"]; they have no usable "+
		"SHA256 checksum, so the fingerprint does not cover them:\n[object-0.txt]\n[object-2.txt]\n[object-4.txt]\n", warnings)

	suite.Run("nothing is logged when every checksum is usable", func() {
		_, warnings, err := snapshotSkippingUnusable(suite.T(), checksummedBucket(objects))
		require.NoError(suite.T(), err)
		require.Empty(suite.T(), warnings)
	})

	suite.Run("the list is capped", func() {
		many := map[string][]byte{"kept.txt": []byte("k")}
		for i := 0; i < 40; i++ {
			many[fmt.Sprintf("object-%02d.txt", i)] = []byte("x")
		}
		client := &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: many,
			Checksums: fullObjectChecksums(map[string][]byte{"kept.txt": []byte("k")})}
		_, warnings, err := snapshotSkippingUnusable(suite.T(), client)
		require.NoError(suite.T(), err)
		require.Contains(suite.T(), warnings, "skipped 40 of the 41 objects")
		require.Contains(suite.T(), warnings, "[object-09.txt]\n(and 30 more)\n")
		require.NotContains(suite.T(), warnings, "[object-10.txt]")
	})

	suite.Run("an object .kosli_ignore excludes is not read, so not skipped", func() {
		ignoring := map[string][]byte{".kosli_ignore": []byte("logs\n"), "app.js": []byte("app"), "lib.js": []byte("lib"), "logs/a.log": []byte("a")}
		checksums := fullObjectChecksums(ignoring)
		delete(checksums, "logs/a.log")
		delete(checksums, "lib.js")
		_, warnings, err := snapshotSkippingUnusable(suite.T(), &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: ignoring, Checksums: checksums})
		require.NoError(suite.T(), err)
		require.Contains(suite.T(), warnings, "skipped 1 of the 2 objects")
		require.NotContains(suite.T(), warnings, "logs/a.log")
	})
}

func (suite *S3MetadataTestSuite) TestSkippingEveryObjectReportsAnEmptySnapshot() {
	for _, t := range []struct {
		name    string
		objects map[string][]byte
	}{
		{name: "a single object", objects: map[string][]byte{"README.md": []byte(fakeReadmeBody)}},
		{name: "several objects", objects: map[string][]byte{"README.md": []byte(fakeReadmeBody), "dir/notes.txt": []byte(fakeNotesBody)}},
	} {
		suite.Run(t.name, func() {
			data, warnings, err := snapshotSkippingUnusable(suite.T(), &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: t.objects})
			require.NoError(suite.T(), err)
			require.Empty(suite.T(), data)
			require.NotNil(suite.T(), data, "the payload must carry an empty list, not null")
			require.Contains(suite.T(), warnings, fmt.Sprintf("skipped %d of the %d objects", len(t.objects), len(t.objects)))
			require.Contains(suite.T(), warnings, "[README.md]")
			require.Contains(suite.T(), warnings, "[warning] no object in bucket ["+fakeS3TestBucketName+"] is left to fingerprint; reporting an empty snapshot\n")
		})
	}
}

func (suite *S3MetadataTestSuite) TestSkippingUnusableStillFailsOnOtherErrors() {
	objects := map[string][]byte{"README.md": []byte(fakeReadmeBody), "notes.txt": []byte(fakeNotesBody)}

	suite.Run("a metadata request failure", func() {
		for _, objects := range []map[string][]byte{objects, {"README.md": []byte(fakeReadmeBody)}} {
			_, _, err := snapshotSkippingUnusable(suite.T(), &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: objects, HeadObjectErr: errInjected})
			require.ErrorIs(suite.T(), err, errInjected)
		}
	})

	suite.Run("a bucket with no objects", func() {
		_, _, err := snapshotSkippingUnusable(suite.T(), &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: map[string][]byte{"dir/": nil}})
		require.ErrorContains(suite.T(), err, "no matching file or dirs in bucket")
	})
}

func TestS3MetadataTestSuite(t *testing.T) {
	suite.Run(t, new(S3MetadataTestSuite))
}
