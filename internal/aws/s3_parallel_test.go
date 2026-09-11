package aws

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/utils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type S3ParallelTestSuite struct {
	suite.Suite
}

// trackingDownloader records how many downloads, and how many listed bytes,
// are in flight at once, and how often each key is fetched. A delay keeps
// downloads overlapping so the bounds are actually exercised.
type trackingDownloader struct {
	S3API
	sizes map[string]int64
	delay time.Duration
	// hook, when set, runs in place of the delegate for that key.
	hook func(ctx context.Context, key string) error

	mu               sync.Mutex
	inFlight         int
	maxInFlight      int
	bytesInFlight    int64
	maxBytesInFlight int64
	calls            map[string]int
}

func (d *trackingDownloader) DownloadObject(ctx context.Context, params *transfermanager.DownloadObjectInput, optFns ...func(*transfermanager.Options)) (*transfermanager.DownloadObjectOutput, error) {
	key := *params.Key
	d.mu.Lock()
	if d.calls == nil {
		d.calls = map[string]int{}
	}
	d.calls[key]++
	d.inFlight++
	d.bytesInFlight += d.sizes[key]
	d.maxInFlight = max(d.maxInFlight, d.inFlight)
	d.maxBytesInFlight = max(d.maxBytesInFlight, d.bytesInFlight)
	d.mu.Unlock()
	defer func() {
		d.mu.Lock()
		d.inFlight--
		d.bytesInFlight -= d.sizes[key]
		d.mu.Unlock()
	}()

	time.Sleep(d.delay)
	if d.hook != nil {
		if err := d.hook(ctx, key); err != nil {
			return nil, err
		}
	}
	return d.S3API.DownloadObject(ctx, params, optFns...)
}

func (d *trackingDownloader) currentInFlight() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.inFlight
}

// bucketOf builds a fake bucket of n objects of the given size and returns the
// tracking downloader plus the listing fingerprintS3Objects takes.
func bucketOf(n int, size int, delay time.Duration) (*trackingDownloader, []s3Object) {
	objects := map[string][]byte{}
	listing := make([]s3Object, 0, n)
	sizes := map[string]int64{}
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("dir%d/object-%03d.bin", i%4, i)
		body := []byte(fmt.Sprintf("%0*d", size, i))
		objects[key] = body
		sizes[key] = int64(len(body))
		listing = append(listing, s3Object{key: key, size: int64(len(body))})
	}
	fake := &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: objects}
	return &trackingDownloader{S3API: fake, sizes: sizes, delay: delay}, listing
}

func (suite *S3ParallelTestSuite) TestMakesExactlyOneCallPerObject() {
	client, listing := bucketOf(60, 8, 0)
	_, parallel, err := fingerprintS3Objects(client, fakeS3TestBucketName, listing, downloadLimits{concurrency: 8, bytesInFlight: 1 << 30}, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	require.Len(suite.T(), client.calls, 60)
	for key, n := range client.calls {
		require.Equal(suite.T(), 1, n, "key %s", key)
	}

	sequential, _ := bucketOf(60, 8, 0)
	_, want, err := fingerprintS3Objects(sequential, fakeS3TestBucketName, listing, downloadLimits{concurrency: 1, bytesInFlight: 1 << 30}, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), want, parallel)
}

func (suite *S3ParallelTestSuite) TestRespectsTheConcurrencyBound() {
	client, listing := bucketOf(40, 8, 5*time.Millisecond)
	_, _, err := fingerprintS3Objects(client, fakeS3TestBucketName, listing, downloadLimits{concurrency: 4, bytesInFlight: 1 << 30}, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	require.LessOrEqual(suite.T(), client.maxInFlight, 4)
	require.GreaterOrEqual(suite.T(), client.maxInFlight, 2, "downloads must actually overlap for the bound to be tested")
}

// The byte budget binds before the count bound here: eight slots would allow
// eight 100-byte objects, the budget allows two.
func (suite *S3ParallelTestSuite) TestRespectsTheByteBudget() {
	client, listing := bucketOf(20, 100, 5*time.Millisecond)
	_, _, err := fingerprintS3Objects(client, fakeS3TestBucketName, listing, downloadLimits{concurrency: 8, bytesInFlight: 250}, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	require.LessOrEqual(suite.T(), client.maxBytesInFlight, int64(250))
	require.LessOrEqual(suite.T(), client.maxInFlight, 2)
	require.GreaterOrEqual(suite.T(), client.maxInFlight, 2, "downloads must actually overlap for the budget to be tested")
}

// An object larger than the whole budget must still download, and runs alone.
func (suite *S3ParallelTestSuite) TestAnObjectLargerThanTheBudgetRunsAlone() {
	client, listing := bucketOf(12, 100, 5*time.Millisecond)
	big := "big/huge.bin"
	body := make([]byte, 1000)
	client.S3API.(*FakeS3Client).Objects[big] = body
	client.sizes[big] = 1000
	listing = append([]s3Object{{key: big, size: 1000}}, listing...)

	var aloneChecks int
	client.hook = func(_ context.Context, key string) error {
		if key == big {
			suite.mu().Lock()
			aloneChecks++
			suite.mu().Unlock()
			require.Equal(suite.T(), 1, client.currentInFlight(), "the oversized object must be the only download in flight")
		}
		return nil
	}

	_, _, err := fingerprintS3Objects(client, fakeS3TestBucketName, listing, downloadLimits{concurrency: 8, bytesInFlight: 250}, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, aloneChecks)
	require.Equal(suite.T(), 1, client.calls[big])
}

var suiteMu sync.Mutex

func (suite *S3ParallelTestSuite) mu() *sync.Mutex { return &suiteMu }

// Random per-object delays reorder completion; the fingerprint must equal what
// DirSha256 gives the same tree on disk, every time.
func (suite *S3ParallelTestSuite) TestFingerprintIsIndependentOfCompletionOrder() {
	tree := map[string]string{}
	for i := 0; i < 30; i++ {
		tree[fmt.Sprintf("d%d/sub%d/f%02d.txt", i%3, i%5, i)] = fmt.Sprintf("content %d", i)
	}
	root := suite.T().TempDir()
	objects := map[string][]byte{}
	sizes := map[string]int64{}
	listing := []s3Object{}
	for p, content := range tree {
		require.NoError(suite.T(), utils.CreateFileWithContent(filepath.Join(root, filepath.FromSlash(p)), content))
		objects[p] = []byte(content)
		sizes[p] = int64(len(content))
		listing = append(listing, s3Object{key: p, size: int64(len(content))})
	}
	want, err := digest.DirSha256(root, nil, logger.NewStandardLogger())
	require.NoError(suite.T(), err)

	random := rand.New(rand.NewSource(5))
	for round := 0; round < 3; round++ {
		client := &trackingDownloader{S3API: &FakeS3Client{Bucket: fakeS3TestBucketName, Objects: objects}, sizes: sizes}
		client.hook = func(_ context.Context, _ string) error {
			suite.mu().Lock()
			d := time.Duration(random.Intn(4)) * time.Millisecond
			suite.mu().Unlock()
			time.Sleep(d)
			return nil
		}
		name, got, err := fingerprintS3Objects(client, fakeS3TestBucketName, listing, downloadLimits{concurrency: 8, bytesInFlight: 1 << 30}, logger.NewStandardLogger())
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), fakeS3TestBucketName, name)
		require.Equal(suite.T(), want, got, "round %d", round)
	}
}

// The first transport error cancels the shared context: an in-flight download
// sees it and returns, and no further download starts. Goroutines race for the
// slots, so the hook decides by arrival rather than by key: the first download
// to arrive fails, every other one blocks until it is cancelled.
func (suite *S3ParallelTestSuite) TestATransportErrorStopsRemainingWork() {
	client, listing := bucketOf(10, 8, 0)
	errBoom := errors.New("boom")
	var arrivals int
	var failingKey string
	client.hook = func(ctx context.Context, key string) error {
		suite.mu().Lock()
		arrivals++
		first := arrivals == 1
		if first {
			failingKey = key
		}
		suite.mu().Unlock()
		if first {
			time.Sleep(5 * time.Millisecond) // let the second slot fill first
			return errBoom
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return errors.New("the context was never cancelled")
		}
	}

	_, _, err := fingerprintS3Objects(client, fakeS3TestBucketName, listing, downloadLimits{concurrency: 2, bytesInFlight: 1 << 30}, logger.NewStandardLogger())
	require.Error(suite.T(), err)
	require.ErrorIs(suite.T(), err, errBoom)
	require.Contains(suite.T(), err.Error(), fmt.Sprintf("object key [%s]", failingKey))
	require.Len(suite.T(), client.calls, 2, "only the two downloads in flight at the failure may have started: %v", client.calls)
}

func (suite *S3ParallelTestSuite) TestDefaultLimitsAreSane() {
	require.GreaterOrEqual(suite.T(), defaultDownloadLimits.concurrency, 2)
	require.GreaterOrEqual(suite.T(), defaultDownloadLimits.bytesInFlight, int64(64<<20))
}

func TestS3ParallelTestSuite(t *testing.T) {
	suite.Run(t, new(S3ParallelTestSuite))
}
