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
		wantErr string
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
			wantErr: "invalid include name regex pattern ^foo[: error parsing regexp: missing closing ]: `[`",
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
			wantErr: "invalid exclude name regex pattern ^foo[: error parsing regexp: missing closing ]: `[`",
		},
	} {
		suite.Run(t.name, func() {
			included, err := t.filter.Compile().ShouldInclude(t.input)
			if t.wantErr != "" {
				require.EqualError(suite.T(), err, t.wantErr)
				return
			}
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), t.want, included)
		})
	}
}

// TestShouldIncludeReportsInvalidPatternsLazily pins the reporting semantics an invalid
// pattern has always had: it is reported only when matching a name actually reaches it.
// A name settled by a literal or by an earlier pattern never reaches a later invalid one,
// and patterns in the branch that is not taken are never reached at all.
func (suite *FiltersSuite) TestShouldIncludeReportsInvalidPatternsLazily() {
	for _, t := range []struct {
		name    string
		input   string
		filter  *ResourceFilterOptions
		want    bool
		wantErr string
	}{
		{
			name:  "an excluded literal name is settled before the invalid pattern is reached",
			input: "foo",
			filter: &ResourceFilterOptions{
				ExcludeNames:      []string{"foo"},
				ExcludeNamesRegex: []string{"^foo["},
			},
			want: false,
		},
		{
			name:  "a name that matches no literal reaches the invalid exclude pattern",
			input: "bar",
			filter: &ResourceFilterOptions{
				ExcludeNames:      []string{"foo"},
				ExcludeNamesRegex: []string{"^foo["},
			},
			wantErr: "invalid exclude name regex pattern ^foo[: error parsing regexp: missing closing ]: `[`",
		},
		{
			name:  "an included literal name is settled before the invalid pattern is reached",
			input: "foo",
			filter: &ResourceFilterOptions{
				IncludeNames:      []string{"foo"},
				IncludeNamesRegex: []string{"^foo["},
			},
			want: true,
		},
		{
			name:  "an earlier matching exclude pattern is reached before the invalid one",
			input: "bar1",
			filter: &ResourceFilterOptions{
				ExcludeNamesRegex: []string{"^bar.*$", "^foo["},
			},
			want: false,
		},
		{
			name:  "an earlier matching include pattern is reached before the invalid one",
			input: "bar1",
			filter: &ResourceFilterOptions{
				IncludeNamesRegex: []string{"^bar.*$", "^foo["},
			},
			want: true,
		},
		{
			name:  "an invalid include pattern is never reached while excluding",
			input: "bar",
			filter: &ResourceFilterOptions{
				ExcludeNames:      []string{"foo"},
				IncludeNamesRegex: []string{"^foo["},
			},
			want: true,
		},
		{
			name:  "the first invalid pattern reached is the one reported",
			input: "baz",
			filter: &ResourceFilterOptions{
				ExcludeNamesRegex: []string{"^foo[", "^bar("},
			},
			wantErr: "invalid exclude name regex pattern ^foo[: error parsing regexp: missing closing ]: `[`",
		},
	} {
		suite.Run(t.name, func() {
			included, err := t.filter.Compile().ShouldInclude(t.input)
			if t.wantErr != "" {
				require.EqualError(suite.T(), err, t.wantErr)
				return
			}
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), t.want, included)
		})
	}
}

func (suite *FiltersSuite) TestCompile() {
	for _, t := range []struct {
		name        string
		filter      *ResourceFilterOptions
		wantIsSet   bool
		included    []string
		notIncluded []string
	}{
		{
			name:      "an empty filter includes everything",
			filter:    &ResourceFilterOptions{},
			wantIsSet: false,
			included:  []string{"foo", "bar"},
		},
		{
			name: "include names and regex patterns are compiled",
			filter: &ResourceFilterOptions{
				IncludeNames:      []string{"foo"},
				IncludeNamesRegex: []string{"^bar.*$", "^baz.*$"},
			},
			wantIsSet:   true,
			included:    []string{"foo", "bar1", "baz2"},
			notIncluded: []string{"cli-test", "foo1"},
		},
		{
			name: "exclude names and regex patterns are compiled",
			filter: &ResourceFilterOptions{
				ExcludeNames:      []string{"foo"},
				ExcludeNamesRegex: []string{"^bar.*$", "^baz.*$"},
			},
			wantIsSet:   true,
			included:    []string{"cli-test", "foo1"},
			notIncluded: []string{"foo", "bar1", "baz2"},
		},
		{
			name: "an invalid pattern still yields a usable filter",
			filter: &ResourceFilterOptions{
				ExcludeNames:      []string{"foo"},
				ExcludeNamesRegex: []string{"^foo["},
			},
			wantIsSet:   true,
			notIncluded: []string{"foo"},
		},
	} {
		suite.Run(t.name, func() {
			compiled := t.filter.Compile()
			require.Equal(suite.T(), t.wantIsSet, compiled.IsSet())
			require.Equal(suite.T(), t.filter.IsSet(), compiled.IsSet())

			for _, name := range t.included {
				included, err := compiled.ShouldInclude(name)
				require.NoError(suite.T(), err)
				require.True(suite.T(), included, "expected %s to be included", name)
			}
			for _, name := range t.notIncluded {
				included, err := compiled.ShouldInclude(name)
				require.NoError(suite.T(), err)
				require.False(suite.T(), included, "expected %s to NOT be included", name)
			}
		})
	}
}

// TestCompiledFilterConcurrentShouldInclude verifies that one compiled filter can be
// shared by many goroutines, as the k8s and ECS snapshot paths do. The filter carries an
// invalid pattern so that the error path is exercised concurrently too. Run with -race.
func (suite *FiltersSuite) TestCompiledFilterConcurrentShouldInclude() {
	filter := &ResourceFilterOptions{
		IncludeNamesRegex: []string{"^include-.*$", "^never-reached["},
	}
	compiled := filter.Compile()

	const workers = 20
	type outcome struct {
		included bool
		err      error
	}
	results := make([]outcome, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// even workers match the first pattern and never reach the invalid one,
			// odd workers match nothing and so are reported the invalid pattern
			name := fmt.Sprintf("include-%d", idx)
			if idx%2 == 1 {
				name = fmt.Sprintf("other-%d", idx)
			}
			included, err := compiled.ShouldInclude(name)
			results[idx] = outcome{included: included, err: err}
		}(i)
	}
	wg.Wait()

	for i, result := range results {
		if i%2 == 1 {
			require.EqualError(suite.T(), result.err,
				"invalid include name regex pattern ^never-reached[: error parsing regexp: missing closing ]: `[`")
			continue
		}
		require.NoError(suite.T(), result.err)
		require.True(suite.T(), result.included, "expected include-%d to be included", i)
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFiltersSuite(t *testing.T) {
	suite.Run(t, new(FiltersSuite))
}

const benchmarkNamesCount = 100

func benchmarkNames() []string {
	names := make([]string, benchmarkNamesCount)
	for i := range names {
		names[i] = fmt.Sprintf("prod-namespace-%d", i)
	}
	return names
}

// BenchmarkCompiledResourceFilterShouldInclude filters many names with the
// patterns compiled once.
func BenchmarkCompiledResourceFilterShouldInclude(b *testing.B) {
	filter := &ResourceFilterOptions{
		IncludeNamesRegex: []string{"^prod-.*$", "^staging-.*$"},
	}
	names := benchmarkNames()
	for b.Loop() {
		compiled := filter.Compile()
		for _, name := range names {
			if _, err := compiled.ShouldInclude(name); err != nil {
				b.Fatal(err)
			}
		}
	}
}
