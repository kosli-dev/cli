package github

import (
	"errors"
	"testing"

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
