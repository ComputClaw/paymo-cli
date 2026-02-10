package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var clientsCmd = &cobra.Command{
	Use:     "clients",
	Aliases: []string{"client"},
	Short:   "Client management commands",
	Long:    `Commands for listing clients in Paymo.`,
}

var listClientsCmd = &cobra.Command{
	Use:   "list",
	Short: "List clients",
	Long: `List all clients accessible to your account.

Examples:
  paymo clients list
  paymo clients list --format json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		clients, err := client.GetClients()
		if err != nil {
			return fmt.Errorf("fetching clients: %w", err)
		}

		formatter := newFormatter()
		return formatter.FormatClients(clients)
	},
}

func init() {
	rootCmd.AddCommand(clientsCmd)
	clientsCmd.AddCommand(listClientsCmd)
}
