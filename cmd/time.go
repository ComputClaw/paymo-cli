package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ComputClaw/paymo-cli/internal/api"
	"github.com/ComputClaw/paymo-cli/internal/config"
)

// timeCmd represents the time command
var timeCmd = &cobra.Command{
	Use:   "time",
	Short: "Time tracking commands",
	Long:  `Commands for tracking time entries, starting/stopping timers, and managing time data.`,
}

// startCmd starts a time tracking session
var startCmd = &cobra.Command{
	Use:   "start [project] [task] [description]",
	Short: "Start a new time tracking session",
	Long: `Start tracking time for a project and task.

Examples:
  paymo time start "My Project" "Development"   # By name
  paymo time start -p 123 -t 456 "Bug fixing"   # By ID with description
  paymo time start -p "My Project" -t "Dev"      # By name with flags`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		// Check if timer is already running
		state, err := config.LoadTimerState()
		if err != nil {
			return fmt.Errorf("loading timer state: %w", err)
		}

		if state.Active {
			return fmt.Errorf("timer already running for '%s' / '%s'\nRun 'paymo time stop' first", state.ProjectName, state.TaskName)
		}

		// Get project and task
		projectFlag, _ := cmd.Flags().GetString("project")
		taskFlag, _ := cmd.Flags().GetString("task")
		descFlag, _ := cmd.Flags().GetString("description")

		var projectID, taskID int
		var projectName, taskName string

		// Determine project
		projectArg := projectFlag
		if projectArg == "" && len(args) > 0 {
			projectArg = args[0]
		}
		if projectArg == "" {
			return fmt.Errorf("project is required - use 'paymo time start <project>' or '-p <id>'")
		}

		project, err := resolveProject(client, projectArg)
		if err != nil {
			return err
		}
		projectID = project.ID
		projectName = project.Name

		// Determine task
		taskArg := taskFlag
		if taskArg == "" && len(args) > 1 {
			taskArg = args[1]
		}
		if taskArg == "" {
			return fmt.Errorf("task is required - use 'paymo time start <project> <task>' or '-t <id>'")
		}

		// Resolve task with project context for name-based lookup
		task, err := resolveTask(client, taskArg, fmt.Sprintf("%d", projectID))
		if err != nil {
			return err
		}
		taskID = task.ID
		taskName = task.Name

		// Get description
		description := descFlag
		if description == "" && len(args) > 2 {
			description = args[2]
		}

		// Start the entry via API
		entry, err := client.StartEntry(taskID, description)
		if err != nil {
			return fmt.Errorf("starting timer: %w", err)
		}

		// Save timer state locally
		timerState := &config.TimerState{
			Active:      true,
			EntryID:     entry.ID,
			ProjectID:   projectID,
			TaskID:      taskID,
			ProjectName: projectName,
			TaskName:    taskName,
			Description: description,
			StartTime:   time.Now(),
		}

		if err := config.SaveTimerState(timerState); err != nil {
			return fmt.Errorf("saving timer state: %w", err)
		}

		formatter := newFormatter()
		if formatter.Format == "json" {
			return formatter.FormatTimeEntry(entry)
		}
		if !formatter.Quiet {
			fmt.Fprintf(formatter.Writer, "Timer started\n")
			fmt.Fprintf(formatter.Writer, "  Project:     %s\n", projectName)
			fmt.Fprintf(formatter.Writer, "  Task:        %s\n", taskName)
			if description != "" {
				fmt.Fprintf(formatter.Writer, "  Description: %s\n", description)
			}
			fmt.Fprintf(formatter.Writer, "  Started:     %s\n", time.Now().Format("15:04:05"))
		} else {
			fmt.Fprintf(formatter.Writer, "%d\n", entry.ID)
		}

		return nil
	},
}

// stopCmd stops the current time tracking session
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the current time tracking session",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		// Check timer state
		state, err := config.LoadTimerState()
		if err != nil {
			return fmt.Errorf("loading timer state: %w", err)
		}

		if !state.Active {
			formatter := newFormatter()
			return formatter.FormatSuccess("No timer is currently running.", 0)
		}

		// Stop the entry via API
		entry, err := client.StopEntry(state.EntryID)
		if err != nil {
			return fmt.Errorf("stopping timer: %w", err)
		}

		// Clear timer state
		if err := config.ClearTimerState(); err != nil {
			return fmt.Errorf("clearing timer state: %w", err)
		}

		formatter := newFormatter()
		if formatter.Format == "json" {
			return formatter.FormatTimeEntry(entry)
		}
		if !formatter.Quiet {
			elapsed := state.FormatElapsedTime()
			fmt.Fprintf(formatter.Writer, "Timer stopped\n")
			fmt.Fprintf(formatter.Writer, "  Project:  %s\n", state.ProjectName)
			fmt.Fprintf(formatter.Writer, "  Task:     %s\n", state.TaskName)
			fmt.Fprintf(formatter.Writer, "  Duration: %s\n", elapsed)
			fmt.Fprintf(formatter.Writer, "  Entry ID: %d\n", entry.ID)
		} else {
			fmt.Fprintf(formatter.Writer, "%d\n", entry.ID)
		}

		return nil
	},
}

// statusCmd shows current time tracking status
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current time tracking status",
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := config.LoadTimerState()
		if err != nil {
			return fmt.Errorf("loading timer state: %w", err)
		}

		formatter := newFormatter()

		if !state.Active {
			if formatter.Format == "json" {
				return formatter.FormatTimerStatus(map[string]interface{}{
					"active": false,
				})
			}
			if !formatter.Quiet {
				fmt.Fprintln(formatter.Writer, "No timer is currently running.")
				fmt.Fprintln(formatter.Writer, "\nRun 'paymo time start <project> <task>' to start tracking.")
			}
			return nil
		}

		if formatter.Format == "json" {
			return formatter.FormatTimerStatus(map[string]interface{}{
				"active":       true,
				"entry_id":     state.EntryID,
				"project_id":   state.ProjectID,
				"project_name": state.ProjectName,
				"task_id":      state.TaskID,
				"task_name":    state.TaskName,
				"description":  state.Description,
				"start_time":   state.StartTime.Format(time.RFC3339),
				"elapsed":      state.FormatElapsedTime(),
			})
		}
		if !formatter.Quiet {
			fmt.Fprintf(formatter.Writer, "Timer Running\n")
			fmt.Fprintf(formatter.Writer, "  Project:     %s\n", state.ProjectName)
			fmt.Fprintf(formatter.Writer, "  Task:        %s\n", state.TaskName)
			if state.Description != "" {
				fmt.Fprintf(formatter.Writer, "  Description: %s\n", state.Description)
			}
			fmt.Fprintf(formatter.Writer, "  Started:     %s\n", state.StartTime.Format("15:04:05"))
			fmt.Fprintf(formatter.Writer, "  Elapsed:     %s\n", state.FormatElapsedTime())
		}

		return nil
	},
}

// logCmd shows time entries
var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show time entries",
	Long: `Display time entries with filtering options.

Examples:
  paymo time log                    # Today's entries
  paymo time log --date yesterday   # Yesterday's entries
  paymo time log --date 2026-02-01  # Specific date
  paymo time log --project 123      # Filter by project
  paymo time log --project "Proj"   # Filter by project name`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		// Get user ID
		creds, _ := config.LoadCredentials()
		userID := 0
		if creds != nil {
			userID = creds.UserID
		}

		// Parse date filter
		dateFlag, _ := cmd.Flags().GetString("date")
		projectFlag, _ := cmd.Flags().GetString("project")

		opts := &api.EntryListOptions{
			UserID:         userID,
			IncludeTask:    true,
			IncludeProject: true,
		}

		// Handle date filter
		now := time.Now()
		switch dateFlag {
		case "today", "":
			opts.StartDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			opts.EndDate = opts.StartDate.Add(24 * time.Hour)
		case "yesterday":
			opts.StartDate = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
			opts.EndDate = opts.StartDate.Add(24 * time.Hour)
		case "this-week":
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			opts.StartDate = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
			opts.EndDate = now
		case "last-week":
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			endOfLastWeek := time.Date(now.Year(), now.Month(), now.Day()-weekday, 23, 59, 59, 0, now.Location())
			opts.StartDate = endOfLastWeek.AddDate(0, 0, -6)
			opts.EndDate = endOfLastWeek
		default:
			// Try to parse as date
			date, err := time.Parse("2006-01-02", dateFlag)
			if err != nil {
				return fmt.Errorf("invalid date format: %s (use YYYY-MM-DD)", dateFlag)
			}
			opts.StartDate = date
			opts.EndDate = date.Add(24 * time.Hour)
		}

		// Handle project filter
		if projectFlag != "" {
			projectID, err := resolveProjectID(client, projectFlag)
			if err != nil {
				return err
			}
			opts.ProjectID = projectID
		}

		// Fetch entries
		entries, err := client.GetEntries(opts)
		if err != nil {
			return fmt.Errorf("fetching entries: %w", err)
		}

		// Format output
		formatter := newFormatter()
		return formatter.FormatTimeEntries(entries)
	},
}

func init() {
	rootCmd.AddCommand(timeCmd)
	timeCmd.AddCommand(startCmd)
	timeCmd.AddCommand(stopCmd)
	timeCmd.AddCommand(statusCmd)
	timeCmd.AddCommand(logCmd)

	// Flags for start command
	startCmd.Flags().StringP("project", "p", "", "project name or ID")
	startCmd.Flags().StringP("task", "t", "", "task name or ID")
	startCmd.Flags().StringP("description", "d", "", "time entry description")

	// Flags for log command
	logCmd.Flags().StringP("date", "", "today", "date filter (today, yesterday, this-week, last-week, YYYY-MM-DD)")
	logCmd.Flags().StringP("project", "p", "", "filter by project")
	logCmd.Flags().IntP("limit", "l", 50, "number of entries to show")
}
