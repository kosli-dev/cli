package requests

import (
	"os"
	"testing"
	"time"
)

func TestDefaultTimeoutConfig(t *testing.T) {
	config := DefaultTimeoutConfig()
	if config.Default != TimeoutMedium {
		t.Errorf("expected default timeout %v, got %v", TimeoutMedium, config.Default)
	}
	if config.Upload != TimeoutUpload {
		t.Errorf("expected upload timeout %v, got %v", TimeoutUpload, config.Upload)
	}
}

func TestTimeoutConfigFromEnv(t *testing.T) {
	// Test with env vars set
	os.Setenv("KOSLI_REQUEST_TIMEOUT", "60")
	os.Setenv("KOSLI_UPLOAD_TIMEOUT", "600")
	defer os.Unsetenv("KOSLI_REQUEST_TIMEOUT")
	defer os.Unsetenv("KOSLI_UPLOAD_TIMEOUT")

	config := TimeoutConfigFromEnv()

	if config.Default != 60*time.Second {
		t.Errorf("expected default timeout 60s, got %v", config.Default)
	}
	if config.Upload != 600*time.Second {
		t.Errorf("expected upload timeout 600s, got %v", config.Upload)
	}
}

func TestTimeoutConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *TimeoutConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultTimeoutConfig(),
			wantErr: false,
		},
		{
			name: "negative default timeout",
			config: &TimeoutConfig{
				Default: -1 * time.Second,
				Upload:  TimeoutUpload,
			},
			wantErr: true,
		},
		{
			name: "excessive default timeout",
			config: &TimeoutConfig{
				Default: 1000 * time.Second,
				Upload:  TimeoutUpload,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
