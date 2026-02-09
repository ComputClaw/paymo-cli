package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ComputClaw/paymo-cli/internal/api"
)

// tasksCmd represents the tasks command
var tasksCmd = &cobra.Command{
	Use:     "tasks",
	Aliases: []string{"task"},
	Short:   "Task management commands",
	Long:    `Commands for listing, creating, and managing tasks in Paymo.`,
}

// listTasksCmd lists tasks
var listTasksCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long: `List tasks with optional filtering.

Examples:
  paymo tasks list                      # List all incomplete tasks
  paymo tasks list --project 123        # Filter by project
  paymo tasks list --project "My Proj"  # Filter by project name
  paymo tasks list --all                # Include completed tasks`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		projectFlag, _ := cmd.Flags().GetString("project")
		includeCompleted, _ := cmd.Flags().GetBool("all")

		opts := &api.TaskListOptions{
			IncludeCompleted: includeCompleted,
			IncludeProject:   true,
		}

		if projectFlag != "" {
			projectID, err := resolveProjectID(client, projectFlag)
			if err != nil {
				return err
			}
			opts.ProjectID = projectID
		}

		tasks, err := client.GetTasks(opts)
		if err != nil {
			return fmt.Errorf("fetching tasks: %w", err)
		}

		formatter := newFormatter()
		return formatter.FormatTasks(tasks)
	},
}

// showTaskCmd shows task details
var showTaskCmd = &cobra.Command{
	Use:   "show <task>",
	Short: "Show task details",
	Long: `Show details for a specific task.

Examples:
  paymo tasks show 456                            # By ID
  paymo tasks show "Bug Fix" --project "My Proj"  # By name (requires --project)`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		projectFlag, _ := cmd.Flags().GetString("project")

		task, err := resolveTask(client, args[0], projectFlag)
		if err != nil {
			return err
		}

		formatter := newFormatter()
		return formatter.FormatTask(task)
	},
}

// createTaskCmd creates a new task
var createTaskCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new task",
	Long: `Create a new task in a project.

Examples:
  paymo tasks create "New Feature" --project 123
  paymo tasks create "Bug Fix" -p "My Project" --due 2026-02-15`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		name := args[0]
		projectFlag, _ := cmd.Flags().GetString("project")
		description, _ := cmd.Flags().GetString("description")
		billable, _ := cmd.Flags().GetBool("billable")
		dueDate, _ := cmd.Flags().GetString("due")

		if projectFlag == "" {
			return fmt.Errorf("project is required - use --project flag")
		}

		projectID, err := resolveProjectID(client, projectFlag)
		if err != nil {
			return err
		}

		req := &api.CreateTaskRequest{
			Name:        name,
			ProjectID:   projectID,
			Description: description,
			Billable:    billable,
			DueDate:     dueDate,
		}

		task, err := client.CreateTask(req)
		if err != nil {
			return fmt.Errorf("creating task: %w", err)
		}

		formatter := newFormatter()
		return formatter.FormatTask(task)
	},
}

// completeTaskCmd marks a task as complete
var completeTaskCmd = &cobra.Command{
	Use:   "complete <task>",
	Short: "Mark a task as complete",
	Long: `Mark a task as complete.

Examples:
  paymo tasks complete 456                            # By ID
  paymo tasks complete "Bug Fix" --project "My Proj"  # By name (requires --project)`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		projectFlag, _ := cmd.Flags().GetString("project")

		task, err := resolveTask(client, args[0], projectFlag)
		if err != nil {
			return err
		}

		if err := client.CompleteTask(task.ID); err != nil {
			return fmt.Errorf("completing task: %w", err)
		}

		formatter := newFormatter()
		return formatter.FormatSuccess(fmt.Sprintf("Task '%s' marked as complete.", task.Name), task.ID)
	},
}

func init() {
	rootCmd.AddCommand(tasksCmd)
	tasksCmd.AddCommand(listTasksCmd)
	tasksCmd.AddCommand(showTaskCmd)
	tasksCmd.AddCommand(createTaskCmd)
	tasksCmd.AddCommand(completeTaskCmd)

	// Flags for list command
	listTasksCmd.Flags().StringP("project", "p", "", "filter by project ID or name")
	listTasksCmd.Flags().Bool("all", false, "include completed tasks")

	// Flags for show command
	showTaskCmd.Flags().StringP("project", "p", "", "project ID or name (required for name-based task lookup)")

	// Flags for create command
	createTaskCmd.Flags().StringP("project", "p", "", "project ID or name (required)")
	createTaskCmd.Flags().StringP("description", "d", "", "task description")
	createTaskCmd.Flags().BoolP("billable", "b", true, "task is billable")
	createTaskCmd.Flags().String("due", "", "due date (YYYY-MM-DD)")

	// Flags for complete command
	completeTaskCmd.Flags().StringP("project", "p", "", "project ID or name (required for name-based task lookup)")
}
