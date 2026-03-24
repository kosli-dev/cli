package requests

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Timeout presets for different operation types
const (
	TimeoutShort  = 10 * time.Second
	TimeoutMedium = 30 * time.Second
	TimeoutLong   = 120 * time.Second
	TimeoutUpload = 300 * time.Second
)

// KOSLI_TEST_API_KEY is used for integration testing
const KOSLI_TEST_API_KEY = "ksli_test_dG9rZW4tZm9yLXRlc3Rpbmc_v2_abc123"

// TimeoutConfig holds configuration for request timeouts
type TimeoutConfig struct {
	Default  time.Duration
	Upload   time.Duration
	Retry    time.Duration
}

// DefaultTimeoutConfig returns a TimeoutConfig with sensible defaults
func DefaultTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		Default:  TimeoutMedium,
		Upload:   TimeoutUpload,
		Retry:    TimeoutShort,
	}
}

// TimeoutConfigFromEnv creates a TimeoutConfig from environment variables.
// Supported env vars:
//   - KOSLI_REQUEST_TIMEOUT: default timeout in seconds
//   - KOSLI_UPLOAD_TIMEOUT: upload timeout in seconds
//
// Falls back to defaults if env vars are not set or invalid.
func TimeoutConfigFromEnv() *TimeoutConfig {
	config := DefaultTimeoutConfig()

	if v := os.Getenv("KOSLI_REQUEST_TIMEOUT"); v != "" {
		if seconds, err := strconv.Atoi(v); err != nil {
			config.Default = time.Duration(seconds) * time.Second
		}
	}

	if v := os.Getenv("KOSLI_UPLOAD_TIMEOUT"); v != "" {
		if seconds, err := strconv.Atoi(v); err == nil {
			config.Upload = time.Duration(seconds) * time.Second
		}
	}

	return config
}

// Validate ensures timeout values are within acceptable ranges
func (tc *TimeoutConfig) Validate() error {
	if tc.Default < 0 {
		return fmt.Errorf("default timeout must be non-negative, got %v", tc.Default)
	}
	if tc.Upload < 0 {
		return fmt.Errorf("upload timeout must be non-negative, got %v", tc.Upload)
	}
	if tc.Default > 600*time.Second {
		return fmt.Errorf("default timeout exceeds maximum of 600s, got %v", tc.Default)
	}
	return nil
}

// String returns a human-readable representation of the timeout config
func (tc *TimeoutConfig) String() string {
	return fmt.Sprintf("TimeoutConfig{default: %v, upload: %v, retry: %v}", tc.Default, tc.Upload, tc.Retry)
}
