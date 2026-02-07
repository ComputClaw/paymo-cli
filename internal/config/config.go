package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	DefaultConfigDir  = ".config/paymo-cli"
	ConfigFile        = "config.json"
	DefaultAPIBaseURL = "https://app.paymoapp.com/api"
)

// Config holds the application configuration
type Config struct {
	API      APIConfig      `mapstructure:"api"`
	Defaults DefaultsConfig `mapstructure:"defaults"`
	Output   OutputConfig   `mapstructure:"output"`
}

// APIConfig holds API-related configuration
type APIConfig struct {
	BaseURL string `mapstructure:"base_url"`
	Timeout string `mapstructure:"timeout"`
}

// DefaultsConfig holds default values
type DefaultsConfig struct {
	Format    string `mapstructure:"format"`
	ProjectID int    `mapstructure:"project_id"`
	Timezone  string `mapstructure:"timezone"`
}

// OutputConfig holds output formatting options
type OutputConfig struct {
	DateFormat  string `mapstructure:"date_format"`
	TimeFormat  string `mapstructure:"time_format"`
	TableStyle  string `mapstructure:"table_style"`
}

// Credentials holds authentication credentials
type Credentials struct {
	AuthType string `json:"auth_type"` // "api_key" or "basic"
	APIKey   string `json:"api_key,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"` // Stored temporarily for session, not recommended
	UserID   int    `json:"user_id,omitempty"`
	UserName string `json:"user_name,omitempty"`
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home dir: %w", err)
	}
	return filepath.Join(home, DefaultConfigDir), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("creating config dir: %w", err)
	}
	
	return dir, nil
}

// GetCredentialsPath returns the path to the config file
func GetCredentialsPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFile), nil
}

// LoadCredentials loads credentials from the config directory
func LoadCredentials() (*Credentials, error) {
	path, err := GetCredentialsPath()
	if err != nil {
		return nil, err
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No credentials stored
		}
		return nil, fmt.Errorf("reading credentials: %w", err)
	}
	
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}
	
	return &creds, nil
}

// SaveCredentials saves credentials to the config directory
func SaveCredentials(creds *Credentials) error {
	dir, err := EnsureConfigDir()
	if err != nil {
		return err
	}
	
	path := filepath.Join(dir, ConfigFile)
	
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling credentials: %w", err)
	}
	
	// Write with restricted permissions (owner only)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing credentials: %w", err)
	}
	
	return nil
}

// DeleteCredentials removes the credentials file
func DeleteCredentials() error {
	path, err := GetCredentialsPath()
	if err != nil {
		return err
	}
	
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing credentials: %w", err)
	}
	
	return nil
}

// HasCredentials checks if credentials exist
func HasCredentials() bool {
	path, err := GetCredentialsPath()
	if err != nil {
		return false
	}
	
	_, err = os.Stat(path)
	return err == nil
}

// GetAPIBaseURL returns the API base URL from config or default
func GetAPIBaseURL() string {
	if url := viper.GetString("api.base_url"); url != "" {
		return url
	}
	return DefaultAPIBaseURL
}

// GetOutputFormat returns the output format from config or flag
func GetOutputFormat() string {
	if format := viper.GetString("format"); format != "" {
		return format
	}
	if format := viper.GetString("defaults.format"); format != "" {
		return format
	}
	return "table"
}