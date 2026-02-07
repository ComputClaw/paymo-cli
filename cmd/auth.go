package cmd

import (
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
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
You can use either email/password or API key authentication.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, _ := cmd.Flags().GetString("api-key")
		
		if apiKey != "" {
			fmt.Printf("🔑 Setting up API key authentication...\n")
			fmt.Println("⚠️  Implementation pending - will store API key securely")
			return nil
		}

		// Interactive email/password setup
		fmt.Print("Email: ")
		var email string
		fmt.Scanln(&email)

		fmt.Print("Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("error reading password: %v", err)
		}
		password := string(bytePassword)
		fmt.Println() // New line after password input

		fmt.Printf("🔐 Authenticating %s...\n", email)
		fmt.Println("⚠️  Implementation pending - this will authenticate with Paymo API")
		
		// Hide the actual password in output
		_ = password

		return nil
	},
}

// logoutCmd clears authentication
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear authentication credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🚪 Logging out...")
		fmt.Println("⚠️  Implementation pending - will clear stored credentials")
		return nil
	},
}

// statusAuthCmd shows authentication status
var statusAuthCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🔍 Authentication Status")
		fmt.Println("⚠️  Implementation pending - will show current auth state")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(statusAuthCmd)

	// Flags for login command
	loginCmd.Flags().StringP("api-key", "k", "", "authenticate using API key")
	loginCmd.Flags().StringP("email", "e", "", "email address")
	loginCmd.Flags().StringP("server", "s", "https://app.paymoapp.com", "Paymo server URL")
}