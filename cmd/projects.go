package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// projectsCmd represents the projects command
var projectsCmd = &cobra.Command{
	Use:     "projects",
	Aliases: []string{"project", "proj"},
	Short:   "Project management commands",
	Long:    `Commands for listing, creating, and managing projects in Paymo.`,
}

// listProjectsCmd lists all projects
var listProjectsCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	Long:  `List all projects accessible to your account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("📁 Projects")
		fmt.Println("⚠️  Implementation pending - this will integrate with Paymo API")
		return nil
	},
}

// createProjectCmd creates a new project
var createProjectCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new project",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		fmt.Printf("🆕 Creating project: %s\n", name)
		fmt.Println("⚠️  Implementation pending - this will integrate with Paymo API")
		return nil
	},
}

// showProjectCmd shows project details
var showProjectCmd = &cobra.Command{
	Use:   "show <project>",
	Short: "Show project details",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		project := args[0]
		fmt.Printf("📋 Project Details: %s\n", project)
		fmt.Println("⚠️  Implementation pending - this will integrate with Paymo API")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectsCmd)
	projectsCmd.AddCommand(listProjectsCmd)
	projectsCmd.AddCommand(createProjectCmd)
	projectsCmd.AddCommand(showProjectCmd)

	// Flags for list command
	listProjectsCmd.Flags().BoolP("active", "a", true, "show only active projects")
	listProjectsCmd.Flags().StringP("client", "c", "", "filter by client")

	// Flags for create command
	createProjectCmd.Flags().StringP("client", "c", "", "client name or ID")
	createProjectCmd.Flags().StringP("description", "d", "", "project description")
	createProjectCmd.Flags().BoolP("billable", "b", true, "project is billable")
}