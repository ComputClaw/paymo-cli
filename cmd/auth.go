package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"

	"github.com/ComputClaw/paymo-cli/internal/api"
	"github.com/ComputClaw/paymo-cli/internal/cache"
	"github.com/ComputClaw/paymo-cli/internal/config"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  `Commands for managing authentication with Paymo.`,
}

// loginCmd handles authentication setup
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Paymo",
	Long: `Set up authentication credentials for Paymo API access.
You can use either email/password or API key authentication.

API Key (recommended):
  paymo auth login --api-key YOUR_API_KEY

Interactive login:
  paymo auth login`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, _ := cmd.Flags().GetString("api-key")

		var auth api.Authenticator
		var creds *config.Credentials

		if apiKey != "" {
			// API key authentication
			auth = &api.APIKeyAuth{APIKey: apiKey}
			creds = &config.Credentials{
				AuthType: "api_key",
				APIKey:   apiKey,
			}
		} else {
			// Interactive email/password
			email, _ := cmd.Flags().GetString("email")
			if email == "" {
				fmt.Print("Email: ")
				fmt.Scanln(&email)
			}

			fmt.Print("Password: ")
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("error reading password: %v", err)
			}
			password := string(bytePassword)
			fmt.Println() // New line after password input

			auth = &api.BasicAuth{Email: email, Password: password}
			creds = &config.Credentials{
				AuthType: "basic",
				Email:    email,
				// Note: We don't store the password for security
			}
		}

		// Validate credentials by making an API call
		formatter := newFormatter()
		if formatter.Format != "json" && !formatter.Quiet {
			fmt.Print("Validating credentials... ")
		}

		client := api.NewClient(auth)
		user, err := client.GetMe()
		if err != nil {
			if formatter.Format != "json" && !formatter.Quiet {
				fmt.Println("failed")
			}
			return fmt.Errorf("authentication failed: %v", err)
		}

		if formatter.Format != "json" && !formatter.Quiet {
			fmt.Println("ok")
		}

		// Store user info in credentials
		creds.UserID = user.ID
		creds.UserName = user.Name

		// Save credentials
		if err := config.SaveCredentials(creds); err != nil {
			return fmt.Errorf("saving credentials: %v", err)
		}

		// Sync core data into the cache (non-fatal on error)
		syncAfterLogin(formatter, user)

		return formatter.FormatSuccess(
			fmt.Sprintf("Successfully authenticated as %s (%s)", user.Name, user.Email),
			user.ID,
		)
	},
}

// logoutCmd clears authentication
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear authentication credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !config.HasCredentials() {
			formatter := newFormatter()
			return formatter.FormatSuccess("Not currently logged in.", 0)
		}

		if err := config.DeleteCredentials(); err != nil {
			return fmt.Errorf("removing credentials: %v", err)
		}

		// Also clear any active timer state
		config.ClearTimerState()

		formatter := newFormatter()
		return formatter.FormatSuccess("Successfully logged out.", 0)
	},
}

// statusAuthCmd shows authentication status
var statusAuthCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := config.LoadCredentials()
		if err != nil {
			return fmt.Errorf("loading credentials: %v", err)
		}

		formatter := newFormatter()

		if creds == nil {
			if formatter.Format == "json" {
				return formatter.FormatTimerStatus(map[string]interface{}{
					"authenticated": false,
				})
			}
			if !formatter.Quiet {
				fmt.Fprintln(formatter.Writer, "Not authenticated")
				fmt.Fprintln(formatter.Writer, "\nRun 'paymo auth login' to authenticate.")
			}
			return nil
		}

		// Try to validate the credentials are still valid
		client, clientErr := getAPIClient()
		valid := false
		if clientErr == nil {
			if err := client.ValidateAuth(); err == nil {
				valid = true
			}
		}

		if formatter.Format == "json" {
			status := map[string]interface{}{
				"authenticated": true,
				"method":        creds.AuthType,
				"valid":         valid,
			}
			if creds.UserName != "" {
				status["user_name"] = creds.UserName
				status["user_id"] = creds.UserID
			}
			return formatter.FormatTimerStatus(status)
		}

		if !formatter.Quiet {
			fmt.Fprintln(formatter.Writer, "Authenticated")
			fmt.Fprintf(formatter.Writer, "  Method: %s\n", creds.AuthType)
			if creds.UserName != "" {
				fmt.Fprintf(formatter.Writer, "  User:   %s (ID: %d)\n", creds.UserName, creds.UserID)
			}
			if valid {
				fmt.Fprintf(formatter.Writer, "  Status: Valid\n")
			} else {
				fmt.Fprintf(formatter.Writer, "  Status: Invalid or expired\n")
			}
		}

		return nil
	},
}

// getAPIClient creates an API client from stored credentials or environment.
// When caching is enabled, the returned client transparently caches reads.
// Defined as a var to allow test injection.
var getAPIClient = func() (api.PaymoAPI, error) {
	// Check environment variable first
	if envKey := config.GetAPIKeyFromEnv(); envKey != "" {
		auth := &api.APIKeyAuth{APIKey: envKey}
		client := api.NewClientWithBaseURL(config.GetAPIBaseURL(), auth)
		return wrapWithCache(client), nil
	}

	// Check credentials file
	creds, err := config.LoadCredentials()
	if err != nil {
		return nil, fmt.Errorf("loading credentials: %w", err)
	}

	if creds == nil {
		return nil, fmt.Errorf("not authenticated - run 'paymo auth login' first\n   or set PAYMO_API_KEY environment variable")
	}

	// Warn about insecure permissions
	if err := config.CheckCredentialsPermissions(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	var auth api.Authenticator
	switch creds.AuthType {
	case "api_key":
		auth = &api.APIKeyAuth{APIKey: creds.APIKey}
	case "basic":
		// Basic auth requires password which we don't store
		return nil, fmt.Errorf("session expired - please login again with 'paymo auth login'")
	default:
		return nil, fmt.Errorf("unknown auth type: %s", creds.AuthType)
	}

	client := api.NewClientWithBaseURL(config.GetAPIBaseURL(), auth)
	return wrapWithCache(client), nil
}


// wrapWithCache wraps a client with the JSON file cache layer if enabled.
func wrapWithCache(client api.PaymoAPI) api.PaymoAPI {
	if viper.GetBool("no_cache") {
		return client
	}
	cacheDir, err := config.GetConfigDir()
	if err != nil {
		return client
	}
	cachePath := filepath.Join(cacheDir, "cache.json")
	store, err := cache.Open(cachePath)
	if err != nil {
		if viper.GetBool("verbose") {
			fmt.Fprintf(os.Stderr, "Warning: cache unavailable: %v\n", err)
		}
		return client
	}
	return cache.NewCachedClient(client, store)
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(statusAuthCmd)

	// Flags for login command
	loginCmd.Flags().StringP("api-key", "k", "", "authenticate using API key")
	loginCmd.Flags().StringP("email", "e", "", "email address")
}
