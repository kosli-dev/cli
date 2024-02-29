package snyk

import (
	"testing"
)

func TestProcessSnykResultFile(t *testing.T) {
	type result struct {
		high_count   int
		medium_count int
		low_count    int
	}
	type want struct {
		tool    string
		version string
		results []result
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
			name: "a sarif code test json file is parsed correctly",
			file: "sarif-code.json",
			want: &want{
				tool:    "SnykCode",
				version: "1.0.0",
				results: []result{
					{
						high_count:   0,
						medium_count: 9,
						low_count:    0,
					},
				},
			},
		},
		{
			name: "a sarif container test json file is parsed correctly",
			file: "sarif-container.json",
			want: &want{
				tool:    "Snyk Container",
				version: "",
				results: []result{
					{
						high_count:   0,
						medium_count: 0,
						low_count:    0,
					},
					{
						high_count:   2,
						medium_count: 1,
						low_count:    0,
					},
				},
			},
		},
		{
			name: "a sarif empty container test json file is parsed correctly",
			file: "sarif-empty-container.json",
			want: &want{
				tool:    "Snyk Container",
				version: "",
				results: []result{
					{
						high_count:   0,
						medium_count: 0,
						low_count:    0,
					},
					{
						high_count:   0,
						medium_count: 0,
						low_count:    0,
					},
				},
			},
		},
		{
			name: "a sarif IaC (terraform) test json file is parsed correctly",
			file: "sarif-iac.json",
			want: &want{
				tool:    "Snyk IaC",
				version: "1.1177.0",
				results: []result{
					{
						high_count:   4,
						medium_count: 9,
						low_count:    113,
					},
				},
			},
		},
		{
			name: "a sarif IaC (helm) test json file is parsed correctly",
			file: "sarif-helm.json",
			want: &want{
				tool:    "Snyk IaC",
				version: "1.1177.0",
				results: []result{
					{
						high_count:   0,
						medium_count: 3,
						low_count:    2,
					},
				},
			},
		},
		{
			name: "a sarif IaC (Serverless) test json file is parsed correctly",
			file: "sarif-serverless.json",
			want: &want{
				tool:    "Snyk IaC",
				version: "1.1177.0",
				results: []result{
					{
						high_count:   0,
						medium_count: 0,
						low_count:    7,
					},
				},
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
				tt.want.version != got.Tool.Version) {
				t.Errorf("ProcessSnykResultFile() failed, want: Tool: %s (got %s) -- Version: %s (got %s)", tt.want.tool, got.Tool.Name, tt.want.version, got.Tool.Version)
			}

			if tt.want != nil && len(tt.want.results) > 0 {
				for i, wantResult := range tt.want.results {
					if wantResult.high_count != got.Results[i].HighCount ||
						wantResult.medium_count != got.Results[i].MediumCount ||
						wantResult.low_count != got.Results[i].LowCount {
						t.Errorf("ProcessSnykResultFile() failed for Result [%d], want %v -- got %v", i, wantResult, got.Results[i])
					}
				}
			}
		})
	}
}
