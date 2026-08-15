package filters

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FiltersSuite struct {
	suite.Suite
}

func (suite *FiltersSuite) TestShouldInclude() {
	for _, t := range []struct {
		name    string
		input   string
		filter  *ResourceFilterOptions
		want    bool
		wantErr bool
	}{
		{
			name:  "returns false when input does not match included",
			input: "cli-test",
			filter: &ResourceFilterOptions{
				IncludeNames: []string{"foo"},
			},
			want: false,
		},
		{
			name:  "returns false when input does not match included-regex",
			input: "cli-test",
			filter: &ResourceFilterOptions{
				IncludeNamesRegex: []string{"^foo$"},
			},
			want: false,
		},
		{
			name:  "returns true when input matches included",
			input: "foo",
			filter: &ResourceFilterOptions{
				IncludeNames: []string{"foo", "bar"},
			},
			want: true,
		},
		{
			name:  "returns true when input matches included-regex",
			input: "foo",
			filter: &ResourceFilterOptions{
				IncludeNamesRegex: []string{"^foo$"},
			},
			want: true,
		},
		{
			name:  "error returned when include regex is invalid",
			input: "foo",
			filter: &ResourceFilterOptions{
				IncludeNamesRegex: []string{"^foo["},
			},
			wantErr: true,
		},
		{
			name:  "returns false when input matches excluded",
			input: "foo",
			filter: &ResourceFilterOptions{
				ExcludeNames: []string{"foo"},
			},
			want: false,
		},
		{
			name:  "returns true when input does not match excluded",
			input: "foo",
			filter: &ResourceFilterOptions{
				ExcludeNames: []string{"foo1"},
			},
			want: true,
		},
		{
			name:  "returns false when input matches excluded-regex",
			input: "foo",
			filter: &ResourceFilterOptions{
				ExcludeNamesRegex: []string{"^foo$"},
			},
			want: false,
		},
		{
			name:  "returns true when input does not match excluded-regex",
			input: "foo",
			filter: &ResourceFilterOptions{
				ExcludeNamesRegex: []string{"^foo1.*"},
			},
			want: true,
		},
		{
			name:  "error returned when exclude regex is invalid",
			input: "foo",
			filter: &ResourceFilterOptions{
				ExcludeNamesRegex: []string{"^foo["},
			},
			wantErr: true,
		},
	} {
		suite.Run(t.name, func() {
			answer, err := t.filter.ShouldInclude(t.input)
			require.False(suite.T(), (err != nil) != t.wantErr,
				"ShouldInclude() error = %v, wantErr %v", err, t.wantErr)
			if !t.wantErr {
				require.NoError(suite.T(), err)
				require.Equal(suite.T(), answer, t.want)
			}
		})
	}
}

func (suite *FiltersSuite) TestCompilePatterns() {
	validFilter := &ResourceFilterOptions{
		ExcludeNamesRegex: []string{"^baz.*$", "^qux.*$"},
	}
	err := validFilter.CompilePatterns()
	require.NoError(suite.T(), err)
	require.Len(suite.T(), validFilter.compiledExclude, 2)

	// Subsequent ShouldInclude calls reuse compiled regexes
	inc, err := validFilter.ShouldInclude("baz123")
	require.NoError(suite.T(), err)
	require.False(suite.T(), inc)

	inc, err = validFilter.ShouldInclude("qux456")
	require.NoError(suite.T(), err)
	require.False(suite.T(), inc)

	inc, err = validFilter.ShouldInclude("other")
	require.NoError(suite.T(), err)
	require.True(suite.T(), inc)

	includeFilter := &ResourceFilterOptions{
		IncludeNamesRegex: []string{"^foo.*$", "^bar.*$"},
	}
	err = includeFilter.CompilePatterns()
	require.NoError(suite.T(), err)
	require.Len(suite.T(), includeFilter.compiledInclude, 2)

	inc, err = includeFilter.ShouldInclude("foo123")
	require.NoError(suite.T(), err)
	require.True(suite.T(), inc)

	inc, err = includeFilter.ShouldInclude("bar456")
	require.NoError(suite.T(), err)
	require.True(suite.T(), inc)

	inc, err = includeFilter.ShouldInclude("other")
	require.NoError(suite.T(), err)
	require.False(suite.T(), inc)

	invalidInclude := &ResourceFilterOptions{
		IncludeNamesRegex: []string{"[invalid"},
	}
	require.Error(suite.T(), invalidInclude.CompilePatterns())

	invalidExclude := &ResourceFilterOptions{
		ExcludeNamesRegex: []string{"[invalid"},
	}
	require.Error(suite.T(), invalidExclude.CompilePatterns())
}

func (suite *FiltersSuite) TestConcurrentShouldInclude() {
	filter := &ResourceFilterOptions{
		IncludeNamesRegex: []string{"^include-.*$"},
	}

	type result struct {
		included bool
		err      error
	}

	const workerCount = 20
	results := make([]result, workerCount)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			inc, err := filter.ShouldInclude(fmt.Sprintf("include-%d", idx))
			results[idx] = result{included: inc, err: err}
		}(i)
	}
	wg.Wait()

	for _, res := range results {
		require.NoError(suite.T(), res.err)
		require.True(suite.T(), res.included)
	}

	// Concurrent exclude test
	excludeFilter := &ResourceFilterOptions{
		ExcludeNamesRegex: []string{"^exclude-.*$"},
	}
	excludeResults := make([]result, workerCount)
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			inc, err := excludeFilter.ShouldInclude(fmt.Sprintf("exclude-%d", idx))
			excludeResults[idx] = result{included: inc, err: err}
		}(i)
	}
	wg.Wait()

	for _, res := range excludeResults {
		require.NoError(suite.T(), res.err)
		require.False(suite.T(), res.included)
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFiltersSuite(t *testing.T) {
	suite.Run(t, new(FiltersSuite))
}

func BenchmarkShouldInclude_Regex(b *testing.B) {
	filter := &ResourceFilterOptions{
		IncludeNamesRegex: []string{"^prod-.*$", "^staging-.*$"},
	}
	_ = filter.CompilePatterns()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = filter.ShouldInclude("prod-namespace-app")
	}
}
