package filters

import (
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

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFiltersSuite(t *testing.T) {
	suite.Run(t, new(FiltersSuite))
}
