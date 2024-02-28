package snyk

import (
	"testing"
)

func TestProcessSnykResultFile(t *testing.T) {
	type want struct {
		tool         string
		version      string
		high_count   int
		medium_count int
		low_count    int
	}
	tests := []struct {
		name    string
		file    string
		want    *want
		wantErr bool
	}{
		{
			name:    "a non-sarif json file causes an error",
			file:    "snyk.json",
			wantErr: true,
		},
		{
			name:    "non-existing file causes an error",
			file:    "non-existing.json",
			wantErr: true,
		},
		{
			name: "a sarif json file is parsed correctly",
			file: "sarif.json",
			want: &want{
				tool:         "SnykCode",
				version:      "1.0.0",
				high_count:   0,
				medium_count: 9,
				low_count:    0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessSnykResultFile(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessSnykResultFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && (tt.want.tool != got.Tool.Name ||
				tt.want.version != got.Tool.Version ||
				tt.want.high_count != got.Results[0].HighCount ||
				tt.want.medium_count != got.Results[0].MediumCount ||
				tt.want.low_count != got.Results[0].LowCount) {
				t.Errorf("ProcessSnykResultFile() = %v, want %v", got, tt.want)
			}

		})
	}
}
