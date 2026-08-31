package github

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/shurcooL/graphql"
	"github.com/stretchr/testify/require"
)

// page builds a pageInfo for a connection that has more results after cursor.
func page(cursor string) pageInfo {
	return pageInfo{HasNextPage: true, EndCursor: graphql.String(cursor)}
}

// lastPage is the pageInfo of a connection with nothing after it.
var lastPage = pageInfo{HasNextPage: false}

func TestPaginate_ReturnsSeedWhenNoNextPage(t *testing.T) {
	calls := 0
	got, err := paginate([]string{"a"}, lastPage, defaultMaxPages,
		func(after graphql.String) ([]string, pageInfo, error) {
			calls++
			return nil, lastPage, nil
		})

	require.NoError(t, err)
	require.Equal(t, []string{"a"}, got)
	require.Zero(t, calls, "should not fetch when the first page is the last")
}

func TestPaginate_ConcatenatesPagesInOrder(t *testing.T) {
	pages := map[graphql.String]struct {
		nodes []string
		next  pageInfo
	}{
		"c1": {[]string{"b", "c"}, page("c2")},
		"c2": {[]string{"d"}, lastPage},
	}
	var asked []graphql.String

	got, err := paginate([]string{"a"}, page("c1"), defaultMaxPages,
		func(after graphql.String) ([]string, pageInfo, error) {
			asked = append(asked, after)
			p := pages[after]
			return p.nodes, p.next, nil
		})

	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c", "d"}, got)
	require.Equal(t, []graphql.String{"c1", "c2"}, asked)
}

func TestPaginate_ErrorsWhenCursorDoesNotAdvance(t *testing.T) {
	calls := 0
	_, err := paginate([]string{"a"}, page("stuck"), defaultMaxPages,
		func(after graphql.String) ([]string, pageInfo, error) {
			calls++
			return []string{"b"}, page("stuck"), nil
		})

	require.ErrorContains(t, err, "did not advance")
	require.Equal(t, 1, calls, "should give up as soon as the cursor repeats")
}

func TestPaginate_ErrorsPastMaxPages(t *testing.T) {
	calls := 0
	_, err := paginate([]string{"a"}, page("c0"), 2,
		func(after graphql.String) ([]string, pageInfo, error) {
			calls++
			return []string{"b"}, page(string(after) + "x"), nil
		})

	require.ErrorContains(t, err, "2 pages")
	require.Equal(t, 1, calls, "seed is page 1, so a cap of 2 allows one fetch")
}

func TestPaginate_PropagatesFetchError(t *testing.T) {
	wantErr := errors.New("boom")
	_, err := paginate([]string{"a"}, page("c1"), defaultMaxPages,
		func(after graphql.String) ([]string, pageInfo, error) {
			return nil, pageInfo{}, wantErr
		})

	require.ErrorIs(t, err, wantErr)
}

// GitHub owner and repo names are case-insensitive, so a config carrying a
// differently-cased name still names the same repo — and must keep its retries.
func TestOwns_IsCaseInsensitive(t *testing.T) {
	config := &GithubConfig{Org: "kosli-dev", Repository: "cli"}

	require.True(t, config.owns(prRef{Owner: "Kosli-Dev", Repo: "CLI", Number: 1}))
	require.True(t, config.owns(prRef{Owner: "kosli-dev", Repo: "cli", Number: 1}))
	require.False(t, config.owns(prRef{Owner: "upstream-org", Repo: "cli", Number: 1}))
	require.False(t, config.owns(prRef{Owner: "kosli-dev", Repo: "other", Number: 1}))
}

// The retry ladder must not sit out a cancelled context. Sleep is left unset so
// the production wait path is exercised.
func TestQueryWithRetry_StopsOnContextCancellation(t *testing.T) {
	ts := newGraphQLTestServer(t, "500", "500", "500", "500")
	config := &GithubConfig{Token: "t", BaseURL: ts.URL, Org: "o", Repository: "r"}
	client := graphql.NewClient(graphqlEndpoint(ts.URL), nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	err := config.queryWithRetry(client)(ctx, &struct{}{}, map[string]any{})
	require.Error(t, err)
	require.Less(t, time.Since(start), 5*time.Second, "must not wait out the ladder")
}

// failThenCancelTransport answers with a 500 and cancels the context before
// returning, so the ladder's first wait always sees an already-cancelled
// context while the server error is already in hand. Cancelling any later —
// from the handler, or on a timer — races the client's body read, and losing
// that race aborts the request itself so the 500 never reaches the caller.
type failThenCancelTransport struct {
	cancel context.CancelFunc
	calls  int
}

func (t *failThenCancelTransport) RoundTrip(*http.Request) (*http.Response, error) {
	t.calls++
	t.cancel()
	return &http.Response{
		StatusCode: http.StatusInternalServerError,
		Status:     "500 Internal Server Error",
		Body:       io.NopCloser(strings.NewReader("boom")),
		Header:     make(http.Header),
	}, nil
}

// Cancellation stops the ladder, but the error that caused every attempt to
// fail is the useful one — it must not be replaced by "context canceled".
func TestQueryWithRetry_CancellationKeepsUnderlyingError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	transport := &failThenCancelTransport{cancel: cancel}

	config := &GithubConfig{Token: "t", Org: "o", Repository: "r"}
	client := graphql.NewClient("http://example.invalid/api/graphql",
		&http.Client{Transport: transport})

	err := config.queryWithRetry(client)(ctx, &struct{}{}, map[string]any{})

	require.Error(t, err)
	require.Equal(t, 1, transport.calls, "the ladder must stop at the first wait")
	require.ErrorIs(t, err, context.Canceled, "cancellation must stay detectable")
	require.ErrorContains(t, err, "500", "the server error must survive too")
}
