package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/ComputClaw/paymo-cli/internal/api"
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
		fmt.Print("🔐 Validating credentials... ")
		
		client := api.NewClient(auth)
		user, err := client.GetMe()
		if err != nil {
			fmt.Println("❌")
			return fmt.Errorf("authentication failed: %v", err)
		}
		
		fmt.Println("✅")
		
		// Store user info in credentials
		creds.UserID = user.ID
		creds.UserName = user.Name
		
		// Save credentials
		if err := config.SaveCredentials(creds); err != nil {
			return fmt.Errorf("saving credentials: %v", err)
		}
		
		fmt.Printf("\n🎉 Successfully authenticated as %s (%s)\n", user.Name, user.Email)
		fmt.Printf("   Credentials saved to ~/.config/paymo-cli/credentials\n")
		
		return nil
	},
}

// logoutCmd clears authentication
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear authentication credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !config.HasCredentials() {
			fmt.Println("Not currently logged in.")
			return nil
		}
		
		if err := config.DeleteCredentials(); err != nil {
			return fmt.Errorf("removing credentials: %v", err)
		}
		
		// Also clear any active timer state
		config.ClearTimerState()
		
		fmt.Println("🚪 Successfully logged out.")
		return nil
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
		
		if creds == nil {
			fmt.Println("🔒 Not authenticated")
			fmt.Println("\nRun 'paymo auth login' to authenticate.")
			return nil
		}
		
		fmt.Println("🔓 Authenticated")
		fmt.Printf("   Method: %s\n", creds.AuthType)
		if creds.UserName != "" {
			fmt.Printf("   User: %s (ID: %d)\n", creds.UserName, creds.UserID)
		}
		
		// Try to validate the credentials are still valid
		client, err := getAPIClient()
		if err != nil {
			fmt.Println("   Status: ⚠️  Unable to verify (no valid auth)")
			return nil
		}
		
		if err := client.ValidateAuth(); err != nil {
			fmt.Println("   Status: ❌ Invalid or expired")
		} else {
			fmt.Println("   Status: ✅ Valid")
		}
		
		return nil
	},
}

// getAPIClient creates an API client from stored credentials or environment
func getAPIClient() (*api.Client, error) {
	// Check environment variable first
	if envKey := config.GetAPIKeyFromEnv(); envKey != "" {
		auth := &api.APIKeyAuth{APIKey: envKey}
		return api.NewClientWithBaseURL(config.GetAPIBaseURL(), auth), nil
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
		fmt.Fprintf(os.Stderr, "⚠️  Warning: %v\n", err)
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

	return api.NewClientWithBaseURL(config.GetAPIBaseURL(), auth), nil
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