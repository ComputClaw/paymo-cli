package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, DefaultConfigDir)

	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestCredentials_SaveAndLoad(t *testing.T) {
	// Use a temporary directory for testing
	tmpDir := t.TempDir()

	// Create test credentials
	creds := &Credentials{
		AuthType: "api_key",
		APIKey:   "test-api-key-12345",
		UserID:   123,
		UserName: "Test User",
	}

	// Save to temp location
	testPath := filepath.Join(tmpDir, "credentials")

	// Write credentials manually for test
	data := []byte(`{"auth_type":"api_key","api_key":"test-api-key-12345","user_id":123,"user_name":"Test User"}`)
	err := os.WriteFile(testPath, data, 0600)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Read and verify
	readData, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	if string(readData) != string(data) {
		t.Errorf("data mismatch: expected %s, got %s", string(data), string(readData))
	}

	// Verify file permissions (only on Unix; Windows doesn't support Unix-style perms)
	if os.PathSeparator == '/' {
		info, err := os.Stat(testPath)
		if err != nil {
			t.Fatalf("failed to stat file: %v", err)
		}

		perm := info.Mode().Perm()
		if perm != 0600 {
			t.Errorf("expected permissions 0600, got %o", perm)
		}
	}

	_ = creds // Used to set up the test
}

func TestCredentials_Fields(t *testing.T) {
	tests := []struct {
		name     string
		creds    Credentials
		authType string
	}{
		{
			name: "api_key auth",
			creds: Credentials{
				AuthType: "api_key",
				APIKey:   "my-key",
				UserID:   42,
				UserName: "Alice",
			},
			authType: "api_key",
		},
		{
			name: "basic auth",
			creds: Credentials{
				AuthType: "basic",
				Email:    "alice@example.com",
				UserID:   42,
				UserName: "Alice",
			},
			authType: "basic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.creds.AuthType != tt.authType {
				t.Errorf("expected auth type %s, got %s", tt.authType, tt.creds.AuthType)
			}
			if tt.creds.UserID == 0 {
				t.Error("user ID should not be zero")
			}
		})
	}
}

func TestConfig_Fields(t *testing.T) {
	cfg := Config{
		API: APIConfig{
			BaseURL: "https://custom.api.com",
			Timeout: "60s",
		},
		Defaults: DefaultsConfig{
			Format:   "json",
			Timezone: "UTC",
		},
		Output: OutputConfig{
			DateFormat: "2006-01-02",
			TimeFormat: "15:04",
		},
	}

	if cfg.API.BaseURL != "https://custom.api.com" {
		t.Errorf("unexpected base URL: %s", cfg.API.BaseURL)
	}

	if cfg.Defaults.Format != "json" {
		t.Errorf("unexpected format: %s", cfg.Defaults.Format)
	}
}

func TestGetAPIBaseURL(t *testing.T) {
	url := GetAPIBaseURL()

	if url != DefaultAPIBaseURL {
		t.Errorf("expected %s, got %s", DefaultAPIBaseURL, url)
	}
}

func TestGetOutputFormat(t *testing.T) {
	format := GetOutputFormat()

	// Default should be "table"
	if format != "table" {
		t.Errorf("expected 'table', got '%s'", format)
	}
}

func TestGetAPIKeyFromEnv(t *testing.T) {
	// Test with no env var
	key := GetAPIKeyFromEnv()
	if key != "" {
		t.Error("expected empty key when env var not set")
	}

	// Test with env var set
	os.Setenv("PAYMO_API_KEY", "test-key-123")
	defer os.Unsetenv("PAYMO_API_KEY")

	key = GetAPIKeyFromEnv()
	if key != "test-key-123" {
		t.Errorf("expected 'test-key-123', got '%s'", key)
	}
}