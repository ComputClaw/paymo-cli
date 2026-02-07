package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ComputClaw/paymo-cli/internal/api"
	"github.com/ComputClaw/paymo-cli/internal/output"
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
	Long: `List all projects accessible to your account.

Examples:
  paymo projects list             # List active projects
  paymo projects list --all       # Include inactive projects
  paymo projects list --format json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}
		
		activeOnly, _ := cmd.Flags().GetBool("active")
		allProjects, _ := cmd.Flags().GetBool("all")
		clientFilter, _ := cmd.Flags().GetString("client")
		
		opts := &api.ProjectListOptions{
			ActiveOnly: activeOnly && !allProjects,
		}
		
		if clientFilter != "" {
			if id, err := strconv.Atoi(clientFilter); err == nil {
				opts.ClientID = id
			}
		}
		
		projects, err := client.GetProjects(opts)
		if err != nil {
			return fmt.Errorf("fetching projects: %w", err)
		}
		
		format := viper.GetString("format")
		formatter := output.NewFormatter(format)
		return formatter.FormatProjects(projects)
	},
}

// createProjectCmd creates a new project
var createProjectCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new project",
	Long: `Create a new project in Paymo.

Examples:
  paymo projects create "New Project"
  paymo projects create "Client Work" --client 123 --billable`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}
		
		name := args[0]
		description, _ := cmd.Flags().GetString("description")
		billable, _ := cmd.Flags().GetBool("billable")
		clientID, _ := cmd.Flags().GetInt("client")
		
		req := &api.CreateProjectRequest{
			Name:        name,
			Description: description,
			Billable:    billable,
		}
		
		if clientID > 0 {
			req.ClientID = &clientID
		}
		
		project, err := client.CreateProject(req)
		if err != nil {
			return fmt.Errorf("creating project: %w", err)
		}
		
		fmt.Printf("✅ Project created successfully\n")
		fmt.Printf("   ID: %d\n", project.ID)
		fmt.Printf("   Name: %s\n", project.Name)
		if project.Code != "" {
			fmt.Printf("   Code: %s\n", project.Code)
		}
		
		return nil
	},
}

// showProjectCmd shows project details
var showProjectCmd = &cobra.Command{
	Use:   "show <project>",
	Short: "Show project details",
	Long: `Show details for a specific project.

Examples:
  paymo projects show 123           # By ID
  paymo projects show "My Project"  # By name`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}
		
		projectArg := args[0]
		
		var project *api.Project
		
		// Try as ID first
		if id, err := strconv.Atoi(projectArg); err == nil {
			project, err = client.GetProject(id)
			if err != nil {
				return fmt.Errorf("project not found: %w", err)
			}
		} else {
			// Try as name
			project, err = client.GetProjectByName(projectArg)
			if err != nil {
				return fmt.Errorf("project not found: %w", err)
			}
		}
		
		format := viper.GetString("format")
		if format == "json" {
			formatter := output.NewFormatter(format)
			return formatter.FormatProjects([]api.Project{*project})
		}
		
		// Pretty print for table format
		fmt.Printf("📁 Project: %s\n", project.Name)
		fmt.Printf("   ID: %d\n", project.ID)
		if project.Code != "" {
			fmt.Printf("   Code: %s\n", project.Code)
		}
		if project.Description != "" {
			fmt.Printf("   Description: %s\n", project.Description)
		}
		fmt.Printf("   Status: %s\n", statusString(project.Active))
		fmt.Printf("   Billable: %s\n", boolString(project.Billable))
		if project.BudgetHours > 0 {
			fmt.Printf("   Budget: %.1f hours\n", project.BudgetHours)
		}
		if project.PricePerHour > 0 {
			fmt.Printf("   Rate: $%.2f/hour\n", project.PricePerHour)
		}
		fmt.Printf("   Created: %s\n", project.CreatedOn.Format("2006-01-02"))
		
		return nil
	},
}

// tasksProjectCmd lists tasks for a project
var tasksProjectCmd = &cobra.Command{
	Use:   "tasks <project>",
	Short: "List tasks for a project",
	Long: `List all tasks for a specific project.

Examples:
  paymo projects tasks 123
  paymo projects tasks "My Project" --all`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}
		
		projectArg := args[0]
		includeCompleted, _ := cmd.Flags().GetBool("all")
		
		var projectID int
		
		// Try as ID first
		if id, err := strconv.Atoi(projectArg); err == nil {
			projectID = id
		} else {
			// Try as name
			project, err := client.GetProjectByName(projectArg)
			if err != nil {
				return fmt.Errorf("project not found: %w", err)
			}
			projectID = project.ID
		}
		
		tasks, err := client.GetTasks(&api.TaskListOptions{
			ProjectID:        projectID,
			IncludeCompleted: includeCompleted,
		})
		if err != nil {
			return fmt.Errorf("fetching tasks: %w", err)
		}
		
		format := viper.GetString("format")
		formatter := output.NewFormatter(format)
		return formatter.FormatTasks(tasks)
	},
}

// archiveProjectCmd archives a project
var archiveProjectCmd = &cobra.Command{
	Use:   "archive <project>",
	Short: "Archive a project",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}
		
		projectArg := args[0]
		
		var projectID int
		var projectName string
		
		// Try as ID first
		if id, err := strconv.Atoi(projectArg); err == nil {
			project, err := client.GetProject(id)
			if err != nil {
				return fmt.Errorf("project not found: %w", err)
			}
			projectID = project.ID
			projectName = project.Name
		} else {
			// Try as name
			project, err := client.GetProjectByName(projectArg)
			if err != nil {
				return fmt.Errorf("project not found: %w", err)
			}
			projectID = project.ID
			projectName = project.Name
		}
		
		if err := client.ArchiveProject(projectID); err != nil {
			return fmt.Errorf("archiving project: %w", err)
		}
		
		fmt.Printf("📦 Project '%s' has been archived.\n", projectName)
		
		return nil
	},
}

func statusString(active bool) string {
	if active {
		return "Active"
	}
	return "Inactive"
}

func boolString(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func init() {
	rootCmd.AddCommand(projectsCmd)
	projectsCmd.AddCommand(listProjectsCmd)
	projectsCmd.AddCommand(createProjectCmd)
	projectsCmd.AddCommand(showProjectCmd)
	projectsCmd.AddCommand(tasksProjectCmd)
	projectsCmd.AddCommand(archiveProjectCmd)

	// Flags for list command
	listProjectsCmd.Flags().BoolP("active", "a", true, "show only active projects")
	listProjectsCmd.Flags().Bool("all", false, "show all projects including inactive")
	listProjectsCmd.Flags().StringP("client", "c", "", "filter by client ID")

	// Flags for create command
	createProjectCmd.Flags().StringP("description", "d", "", "project description")
	createProjectCmd.Flags().IntP("client", "c", 0, "client ID")
	createProjectCmd.Flags().BoolP("billable", "b", true, "project is billable")

	// Flags for tasks command
	tasksProjectCmd.Flags().Bool("all", false, "include completed tasks")
}