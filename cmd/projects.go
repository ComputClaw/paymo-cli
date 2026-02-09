package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ComputClaw/paymo-cli/internal/api"
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

		formatter := newFormatter()
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

		formatter := newFormatter()
		return formatter.FormatProject(project)
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

		project, err := resolveProject(client, args[0])
		if err != nil {
			return err
		}

		formatter := newFormatter()
		return formatter.FormatProject(project)
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

		includeCompleted, _ := cmd.Flags().GetBool("all")

		projectID, err := resolveProjectID(client, args[0])
		if err != nil {
			return err
		}

		tasks, err := client.GetTasks(&api.TaskListOptions{
			ProjectID:        projectID,
			IncludeCompleted: includeCompleted,
		})
		if err != nil {
			return fmt.Errorf("fetching tasks: %w", err)
		}

		formatter := newFormatter()
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

		project, err := resolveProject(client, args[0])
		if err != nil {
			return err
		}

		if err := client.ArchiveProject(project.ID); err != nil {
			return fmt.Errorf("archiving project: %w", err)
		}

		formatter := newFormatter()
		return formatter.FormatSuccess(fmt.Sprintf("Project '%s' has been archived.", project.Name), project.ID)
	},
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
