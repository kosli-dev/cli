package main

import (
	"fmt"
	"testing"

	"github.com/kosli-dev/cli/internal/aws"
	"github.com/stretchr/testify/require"
)

func TestResolveDownloadLimitsPicksTheConcurrencyForTheSource(t *testing.T) {
	for _, tc := range []struct {
		name           string
		source         string
		concurrency    int
		concurrencySet bool
		want           int
	}{
		{name: "content mode takes the download default", source: fingerprintSourceContent,
			concurrency: aws.DefaultDownloadLimits.Concurrency, want: aws.DefaultDownloadLimits.Concurrency},
		{name: "metadata mode takes its own default", source: fingerprintSourceMetadata,
			concurrency: aws.DefaultDownloadLimits.Concurrency, want: aws.DefaultMetadataConcurrency},
		{name: "an explicit value wins in metadata mode", source: fingerprintSourceMetadata,
			concurrency: 4, concurrencySet: true, want: 4},
		{name: "an explicit value that equals the download default still wins", source: fingerprintSourceMetadata,
			concurrency: aws.DefaultDownloadLimits.Concurrency, concurrencySet: true, want: aws.DefaultDownloadLimits.Concurrency},
		{name: "an explicit value wins in content mode", source: fingerprintSourceContent,
			concurrency: 4, concurrencySet: true, want: 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := &snapshotS3Options{fingerprintSource: tc.source, downloadConcurrency: tc.concurrency, downloadBudget: defaultDownloadBudget}
			require.NoError(t, o.resolveDownloadLimits(tc.concurrencySet))
			require.Equal(t, tc.want, o.downloadLimits.Concurrency)
		})
	}
}

func TestDefaultMetadataConcurrencyIsWiderThanTheDownloadDefault(t *testing.T) {
	require.Greater(t, aws.DefaultMetadataConcurrency, aws.DefaultDownloadLimits.Concurrency,
		"a source that holds no buffers should not be throttled below the download default")
}

// cobra shows only the flag's own default, so the help text spells out the metadata one.
func TestDownloadConcurrencyHelpStatesTheMetadataDefault(t *testing.T) {
	require.Contains(t, downloadConcurrencyFlag, fmt.Sprintf("defaults to %d", aws.DefaultMetadataConcurrency))
}
