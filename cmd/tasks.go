package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ComputClaw/paymo-cli/internal/api"
	"github.com/ComputClaw/paymo-cli/internal/output"
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
			if id, err := strconv.Atoi(projectFlag); err == nil {
				opts.ProjectID = id
			} else {
				project, err := client.GetProjectByName(projectFlag)
				if err != nil {
					return fmt.Errorf("project not found: %w", err)
				}
				opts.ProjectID = project.ID
			}
		}
		
		tasks, err := client.GetTasks(opts)
		if err != nil {
			return fmt.Errorf("fetching tasks: %w", err)
		}
		
		format := viper.GetString("format")
		formatter := output.NewFormatter(format)
		return formatter.FormatTasks(tasks)
	},
}

// showTaskCmd shows task details
var showTaskCmd = &cobra.Command{
	Use:   "show <task-id>",
	Short: "Show task details",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}
		
		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task ID: %s", args[0])
		}
		
		task, err := client.GetTask(taskID)
		if err != nil {
			return fmt.Errorf("task not found: %w", err)
		}
		
		format := viper.GetString("format")
		if format == "json" {
			formatter := output.NewFormatter(format)
			return formatter.FormatTasks([]api.Task{*task})
		}
		
		// Pretty print
		fmt.Printf("📋 Task: %s\n", task.Name)
		fmt.Printf("   ID: %d\n", task.ID)
		if task.Code != "" {
			fmt.Printf("   Code: %s\n", task.Code)
		}
		fmt.Printf("   Project ID: %d\n", task.ProjectID)
		fmt.Printf("   Status: %s\n", taskStatusString(task.Complete))
		fmt.Printf("   Billable: %s\n", boolString(task.Billable))
		if task.DueDate != "" {
			fmt.Printf("   Due Date: %s\n", task.DueDate)
		}
		if task.Description != "" {
			fmt.Printf("   Description: %s\n", task.Description)
		}
		fmt.Printf("   Created: %s\n", task.CreatedOn.Format("2006-01-02"))
		
		return nil
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
		
		var projectID int
		if id, err := strconv.Atoi(projectFlag); err == nil {
			projectID = id
		} else {
			project, err := client.GetProjectByName(projectFlag)
			if err != nil {
				return fmt.Errorf("project not found: %w", err)
			}
			projectID = project.ID
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
		
		fmt.Printf("✅ Task created successfully\n")
		fmt.Printf("   ID: %d\n", task.ID)
		fmt.Printf("   Name: %s\n", task.Name)
		
		return nil
	},
}

// completeTaskCmd marks a task as complete
var completeTaskCmd = &cobra.Command{
	Use:   "complete <task-id>",
	Short: "Mark a task as complete",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}
		
		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task ID: %s", args[0])
		}
		
		// Get task name first for confirmation
		task, err := client.GetTask(taskID)
		if err != nil {
			return fmt.Errorf("task not found: %w", err)
		}
		
		if err := client.CompleteTask(taskID); err != nil {
			return fmt.Errorf("completing task: %w", err)
		}
		
		fmt.Printf("✅ Task '%s' marked as complete.\n", task.Name)
		
		return nil
	},
}

func taskStatusString(complete bool) string {
	if complete {
		return "Complete"
	}
	return "Open"
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

	// Flags for create command
	createTaskCmd.Flags().StringP("project", "p", "", "project ID or name (required)")
	createTaskCmd.Flags().StringP("description", "d", "", "task description")
	createTaskCmd.Flags().BoolP("billable", "b", true, "task is billable")
	createTaskCmd.Flags().String("due", "", "due date (YYYY-MM-DD)")
}