package azure

import (
	"bytes"
	"io"
	"testing"

	armappservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/kosli-dev/cli/internal/filters"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
)

func site(name string) *armappservice.Site {
	return &armappservice.Site{Name: &name}
}

func siteNames(apps []*armappservice.Site) []string {
	names := make([]string, 0, len(apps))
	for _, app := range apps {
		names = append(names, *app.Name)
	}
	return names
}

func TestFilterApps(t *testing.T) {
	apps := []*armappservice.Site{site("web-api"), site("web-ui"), site("zip-func"), site("zip-worker")}

	for _, tc := range []struct {
		name   string
		filter *filters.ResourceFilterOptions
		want   []string
	}{
		{
			name:   "no filter keeps every app",
			filter: &filters.ResourceFilterOptions{},
			want:   []string{"web-api", "web-ui", "zip-func", "zip-worker"},
		},
		{
			name:   "exclude by literal name",
			filter: &filters.ResourceFilterOptions{ExcludeNames: []string{"zip-func"}},
			want:   []string{"web-api", "web-ui", "zip-worker"},
		},
		{
			name:   "exclude by regex",
			filter: &filters.ResourceFilterOptions{ExcludeNamesRegex: []string{"^zip-"}},
			want:   []string{"web-api", "web-ui"},
		},
		{
			name: "literal and regex excludes together",
			filter: &filters.ResourceFilterOptions{
				ExcludeNames:      []string{"web-ui"},
				ExcludeNamesRegex: []string{"-worker$"},
			},
			want: []string{"web-api", "zip-func"},
		},
		{
			name:   "include by literal name",
			filter: &filters.ResourceFilterOptions{IncludeNames: []string{"web-ui"}},
			want:   []string{"web-ui"},
		},
		{
			name:   "include by regex",
			filter: &filters.ResourceFilterOptions{IncludeNamesRegex: []string{"^web-"}},
			want:   []string{"web-api", "web-ui"},
		},
		{
			name:   "exclude regex is case-sensitive",
			filter: &filters.ResourceFilterOptions{ExcludeNamesRegex: []string{"^ZIP-"}},
			want:   []string{"web-api", "web-ui", "zip-func", "zip-worker"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := filterApps(apps, tc.filter.Compile(), logger.NewStandardLogger())
			require.NoError(t, err)
			require.Equal(t, tc.want, siteNames(got))
		})
	}

	t.Run("nil filter keeps every app", func(t *testing.T) {
		got, err := filterApps(apps, nil, logger.NewStandardLogger())
		require.NoError(t, err)
		require.Equal(t, siteNames(apps), siteNames(got))
	})

	t.Run("invalid regex is reported", func(t *testing.T) {
		filter := &filters.ResourceFilterOptions{ExcludeNamesRegex: []string{"[invalid"}}
		_, err := filterApps(apps, filter.Compile(), logger.NewStandardLogger())
		require.ErrorContains(t, err, "invalid exclude name regex pattern [invalid")
	})

	t.Run("an excluded app is dropped before its properties are read", func(t *testing.T) {
		// site() leaves Properties nil, so reading State or SiteConfig would panic.
		filter := &filters.ResourceFilterOptions{ExcludeNames: []string{"zip-func"}}
		got, err := filterApps([]*armappservice.Site{site("zip-func")}, filter.Compile(), logger.NewStandardLogger())
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("warns when the filters leave out every app", func(t *testing.T) {
		var warnings bytes.Buffer
		filter := &filters.ResourceFilterOptions{IncludeNames: []string{"web-apo"}}
		got, err := filterApps(apps, filter.Compile(), logger.NewLogger(io.Discard, &warnings, false))
		require.NoError(t, err)
		require.Empty(t, got)
		require.Contains(t, warnings.String(), "the include/exclude filters left out all 4 apps in the resource group")
	})

	t.Run("no warning when the filters keep an app", func(t *testing.T) {
		var warnings bytes.Buffer
		filter := &filters.ResourceFilterOptions{IncludeNames: []string{"web-api"}}
		_, err := filterApps(apps, filter.Compile(), logger.NewLogger(io.Discard, &warnings, false))
		require.NoError(t, err)
		require.Empty(t, warnings.String())
	})

	t.Run("no warning for an empty resource group", func(t *testing.T) {
		var warnings bytes.Buffer
		filter := &filters.ResourceFilterOptions{IncludeNames: []string{"web-api"}}
		_, err := filterApps(nil, filter.Compile(), logger.NewLogger(io.Discard, &warnings, false))
		require.NoError(t, err)
		require.Empty(t, warnings.String())
	})
}
