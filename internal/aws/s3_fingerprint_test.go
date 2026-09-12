package aws

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/utils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type S3FingerprintTestSuite struct {
	suite.Suite
}

// recordingDownloader wraps an S3API and records every DownloadObject call: the
// key, and the local file the bytes were written to.
type recordingDownloader struct {
	S3API
	mu    sync.Mutex
	keys  []string
	files []string
	// onDownload runs inside each DownloadObject call, before delegating.
	onDownload func(key string, file *os.File)
}

func (r *recordingDownloader) DownloadObject(ctx context.Context, params *transfermanager.DownloadObjectInput, optFns ...func(*transfermanager.Options)) (*transfermanager.DownloadObjectOutput, error) {
	file, _ := params.WriterAt.(*os.File)
	r.mu.Lock()
	r.keys = append(r.keys, *params.Key)
	if file != nil {
		r.files = append(r.files, file.Name())
	}
	r.mu.Unlock()
	if r.onDownload != nil {
		r.onDownload(*params.Key, file)
	}
	return r.S3API.DownloadObject(ctx, params, optFns...)
}

func (r *recordingDownloader) downloadedKeys() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	keys := append([]string{}, r.keys...)
	sort.Strings(keys)
	return keys
}

func snapshotFake(t *testing.T, client S3API) (artifactName, fingerprint string) {
	t.Helper()
	data, err := getS3DataFromClient(client, fakeS3TestBucketName, nil, nil, nil, nil, DefaultDownloadLimits, logger.NewStandardLogger())
	require.NoError(t, err)
	require.Len(t, data, 1)
	require.Len(t, data[0].Digests, 1)
	for name, sha := range data[0].Digests {
		return name, sha
	}
	return "", ""
}

// TestPinnedFingerprints holds the fingerprints recorded against main before
// content mode stopped writing objects under their keys. They must never move:
// every existing environment snapshot on the server was computed this way.
func (suite *S3FingerprintTestSuite) TestPinnedFingerprints() {
	for _, t := range []struct {
		name             string
		objects          map[string][]byte
		wantArtifactName string
		wantFingerprint  string
	}{
		{
			name: "unusual key shapes fold as filepath.Join folded them",
			objects: map[string][]byte{
				"/lead.txt": []byte("u\n"), "a//b": []byte("o\n"), "./c.txt": []byte("t\n"), `d\e.txt`: []byte("n\n"),
			},
			wantArtifactName: fakeS3TestBucketName,
			wantFingerprint:  "27e2d8aa07677b7818b8cf101b45c928aa8fbf7d17a8c8efc84469e24a106ec3",
		},
		{
			name: "a dot sorts before a slash",
			objects: map[string][]byte{
				"a.txt": []byte("1"), "a/z": []byte("2"), "a/b/c": []byte("3"), "b": []byte("4"),
			},
			wantArtifactName: fakeS3TestBucketName,
			wantFingerprint:  "aaddb8f3e299e12316d42fa9ac36d4ed7d38c34e7ec002239269c86a033bb9fd",
		},
		{
			name: "nested prefixes with folder markers",
			objects: map[string][]byte{
				"dir/": nil, "dir/sub/": nil, "dir/sub/x.yml": []byte("x"), "dir/y.txt": []byte("y"), "README.md": []byte("r"),
			},
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
			name, sha := snapshotFake(suite.T(), &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: t.objects})
			require.Equal(suite.T(), t.wantArtifactName, name)
			require.Equal(suite.T(), t.wantFingerprint, sha)
		})
	}
}

// TestMatchesAttestedDirectory is the property the whole command exists for:
// a bucket holding the files of a directory fingerprints as that directory does
// at attestation time, and a single object as that file does.
func (suite *S3FingerprintTestSuite) TestMatchesAttestedDirectory() {
	tree := map[string]string{
		"README.md":                  "# readme\n",
		"dummy/dummy_2/template.yml": "key: value\n",
		"dummy/other.txt":            "other\n",
		"a.txt":                      "a\n",
		"a/z":                        "z\n",
		".kosli_ignore":              "logs\n*.tmp\n",
		"logs/noise.log":             "noise\n",
		"scratch.tmp":                "tmp\n",
	}
	root := suite.T().TempDir()
	objects := map[string][]byte{}
	for p, content := range tree {
		require.NoError(suite.T(), utils.CreateFileWithContent(filepath.Join(root, filepath.FromSlash(p)), content))
		objects[p] = []byte(content)
	}
	attested, err := digest.DirSha256(root, nil, logger.NewStandardLogger())
	require.NoError(suite.T(), err)

	name, sha := snapshotFake(suite.T(), &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: objects})
	require.Equal(suite.T(), fakeS3TestBucketName, name)
	require.Equal(suite.T(), attested, sha)

	suite.Run("a single object", func() {
		path := filepath.Join(suite.T().TempDir(), "release.bin")
		require.NoError(suite.T(), utils.CreateFileWithContent(path, "the release"))
		attestedFile, err := digest.FileSha256(path, logger.NewStandardLogger())
		require.NoError(suite.T(), err)

		name, sha := snapshotFake(suite.T(), &FakeS3Client{Bucket: fakeS3TestBucketName,
			Objects: map[string][]byte{"builds/v1/release.bin": []byte("the release")}})
		require.Equal(suite.T(), "release.bin", name)
		require.Equal(suite.T(), attestedFile, sha)
	})
}

// A root .kosli_ignore in the bucket applies its rules, as DirSha256 applies
// them on disk: a bucket with an excluded directory fingerprints as one that
// never held it.
func (suite *S3FingerprintTestSuite) TestHonoursRootKosliIgnore() {
	ignore := []byte("logs\n")
	_, withLogs := snapshotFake(suite.T(), &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: map[string][]byte{
		".kosli_ignore": ignore, "app.js": []byte("app"), "logs/a.log": []byte("a"), "logs/deep/b.log": []byte("b"),
	}})
	_, withoutLogs := snapshotFake(suite.T(), &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: map[string][]byte{
		".kosli_ignore": ignore, "app.js": []byte("app"),
	}})
	require.Equal(suite.T(), withoutLogs, withLogs)

	_, noIgnore := snapshotFake(suite.T(), &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: map[string][]byte{
		"app.js": []byte("app"), "logs/a.log": []byte("a"), "logs/deep/b.log": []byte("b"),
	}})
	require.NotEqual(suite.T(), noIgnore, withLogs, "the ignore file must have had an effect")
}

// Exactly the objects that contribute content are downloaded: not folder
// markers, not filter-excluded keys, not keys the root .kosli_ignore excludes.
// Nothing else is skipped, so no object is dropped silently.
func (suite *S3FingerprintTestSuite) TestDownloadsExactlyTheContributingObjects() {
	client := &recordingDownloader{S3API: &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: map[string][]byte{
		".kosli_ignore":    []byte("logs\n*.tmp\n"),
		"app.js":           []byte("app"),
		"lib/util.js":      []byte("util"),
		"lib/":             nil,
		"logs/a.log":       []byte("a"),
		"logs/deep/b.log":  []byte("b"),
		"scratch.tmp":      []byte("tmp"),
		"filtered/out.txt": []byte("out"),
	}}}
	data, err := getS3DataFromClient(client, fakeS3TestBucketName, nil, nil, []string{"filtered/"}, nil, DefaultDownloadLimits, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	require.Len(suite.T(), data, 1)
	require.Equal(suite.T(), []string{".kosli_ignore", "app.js", "lib/util.js"}, client.downloadedKeys())
}

// Objects are downloaded to files whose names owe nothing to the key, every
// file is removed once hashed, and the download directory is gone at the end.
// Downloads overlap, so how many files exist at once is the byte budget's
// concern (see S3ParallelTestSuite), not this test's.
func (suite *S3FingerprintTestSuite) TestObjectsNeverLandUnderTheirKeyAndDoNotLinger() {
	keys := []string{"alpha.bin", "beta/gamma.bin", "delta/epsilon/zeta.bin"}
	objects := map[string][]byte{}
	for _, key := range keys {
		objects[key] = []byte(key)
	}
	client := &recordingDownloader{S3API: &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: objects}}
	client.onDownload = func(key string, file *os.File) {
		require.NotNil(suite.T(), file, "the transfer manager must be handed a real file")
		require.NotContains(suite.T(), filepath.Base(file.Name()), filepath.Base(key),
			"the local file name must owe nothing to the key")
	}

	_, err := getS3DataFromClient(client, fakeS3TestBucketName, nil, nil, nil, nil, DefaultDownloadLimits, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	require.Len(suite.T(), client.files, len(keys))
	for _, file := range client.files {
		_, err := os.Stat(file)
		require.ErrorIs(suite.T(), err, os.ErrNotExist)
		_, err = os.Stat(filepath.Dir(file))
		require.ErrorIs(suite.T(), err, os.ErrNotExist, "the download directory must be removed")
	}
}

func (suite *S3FingerprintTestSuite) TestADownloadErrorNamesTheKey() {
	client := &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: map[string][]byte{
		"README.md": []byte(fakeReadmeBody), "notes.txt": []byte(fakeNotesBody),
	}, DownloadObjectErr: os.ErrDeadlineExceeded}
	_, err := getS3DataFromClient(client, fakeS3TestBucketName, nil, nil, nil, nil, DefaultDownloadLimits, logger.NewStandardLogger())
	require.Error(suite.T(), err)
	require.ErrorIs(suite.T(), err, os.ErrDeadlineExceeded)
	// Downloads overlap, so either object may be the first to fail.
	require.Regexp(suite.T(), `object key \[(README\.md|notes\.txt)\]`, err.Error())
	require.NotContains(suite.T(), err.Error(), "--exclude-regex", "a transport failure must not advise dropping the object")
}

func TestS3FingerprintTestSuite(t *testing.T) {
	suite.Run(t, new(S3FingerprintTestSuite))
}
