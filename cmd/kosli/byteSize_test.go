package main

import (
	"testing"

	"github.com/kosli-dev/cli/internal/aws"
	"github.com/stretchr/testify/require"
)

func TestParseByteSize(t *testing.T) {
	const mib = int64(1) << 20
	for _, tc := range []struct {
		input   string
		want    int64
		wantErr string
	}{
		// A bare number is megabytes, matching how Lambda's ephemeral storage is expressed.
		{input: "512", want: 512 * mib},
		{input: "1", want: mib},
		{input: " 64 ", want: 64 * mib},
		// A suffix picks the unit; the trailing B is optional and case does not matter.
		{input: "512M", want: 512 * mib},
		{input: "512MB", want: 512 * mib},
		{input: "512mb", want: 512 * mib},
		{input: "512MiB", want: 512 * mib},
		{input: "8G", want: 8 << 30},
		{input: "8GB", want: 8 << 30},
		{input: "8 GB", want: 8 << 30},
		{input: "2T", want: 2 << 40},
		{input: "1024K", want: 1 << 20},
		{input: "4096KB", want: 4 << 20},
		{input: "1000B", want: 1000},
		{input: "1.5G", want: 3 << 29},
		{input: "0.5M", want: 512 << 10},
		// Rejected: nothing to download into, or not a size at all.
		{input: "", wantErr: "empty"},
		{input: "0", wantErr: "must be at least 1 byte"},
		{input: "0B", wantErr: "must be at least 1 byte"},
		{input: "-1", wantErr: "not a size"},
		{input: "-512M", wantErr: "not a size"},
		{input: "abc", wantErr: "not a size"},
		{input: "M", wantErr: "not a size"},
		{input: "512X", wantErr: `unknown unit "X"`},
		{input: "512 megabytes", wantErr: `unknown unit "megabytes"`},
		{input: "1e3", wantErr: "not a size"},
		{input: "0x10", wantErr: "not a size"},
		{input: "1,024", wantErr: "not a size"},
		{input: "99999999999T", wantErr: "too large"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, err := parseByteSize(tc.input)
			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

// The flag's default is spelled as a string, so pin it to the value the aws
// package uses when no flag is given.
func TestDefaultDownloadBudgetMatchesTheAwsDefault(t *testing.T) {
	got, err := parseByteSize(defaultDownloadBudget)
	require.NoError(t, err)
	require.Equal(t, aws.DefaultDownloadLimits.BytesInFlight, got)
}
