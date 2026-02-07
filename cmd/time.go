package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ComputClaw/paymo-cli/internal/api"
	"github.com/ComputClaw/paymo-cli/internal/config"
	"github.com/ComputClaw/paymo-cli/internal/output"
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
  paymo time start                              # Interactive selection
  paymo time start "My Project" "Development"   # By name
  paymo time start -p 123 -t 456 "Bug fixing"   # By ID with description`,
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
		if projectFlag != "" {
			// Try as ID first
			if id, err := strconv.Atoi(projectFlag); err == nil {
				project, err := client.GetProject(id)
				if err != nil {
					return fmt.Errorf("project not found: %w", err)
				}
				projectID = project.ID
				projectName = project.Name
			} else {
				// Try as name
				project, err := client.GetProjectByName(projectFlag)
				if err != nil {
					return fmt.Errorf("project not found: %w", err)
				}
				projectID = project.ID
				projectName = project.Name
			}
		} else if len(args) > 0 {
			project, err := client.GetProjectByName(args[0])
			if err != nil {
				return fmt.Errorf("project not found: %w", err)
			}
			projectID = project.ID
			projectName = project.Name
		} else {
			return fmt.Errorf("project is required - use 'paymo time start <project>' or '-p <id>'")
		}
		
		// Determine task
		if taskFlag != "" {
			if id, err := strconv.Atoi(taskFlag); err == nil {
				task, err := client.GetTask(id)
				if err != nil {
					return fmt.Errorf("task not found: %w", err)
				}
				taskID = task.ID
				taskName = task.Name
			} else {
				task, err := client.GetTaskByName(projectID, taskFlag)
				if err != nil {
					return fmt.Errorf("task not found: %w", err)
				}
				taskID = task.ID
				taskName = task.Name
			}
		} else if len(args) > 1 {
			task, err := client.GetTaskByName(projectID, args[1])
			if err != nil {
				return fmt.Errorf("task not found: %w", err)
			}
			taskID = task.ID
			taskName = task.Name
		} else {
			return fmt.Errorf("task is required - use 'paymo time start <project> <task>' or '-t <id>'")
		}
		
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
		
		fmt.Printf("🚀 Timer started\n")
		fmt.Printf("   Project: %s\n", projectName)
		fmt.Printf("   Task: %s\n", taskName)
		if description != "" {
			fmt.Printf("   Description: %s\n", description)
		}
		fmt.Printf("   Started: %s\n", time.Now().Format("15:04:05"))
		
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
			fmt.Println("⚠️  No timer is currently running.")
			return nil
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
		
		elapsed := state.FormatElapsedTime()
		
		fmt.Printf("⏹️  Timer stopped\n")
		fmt.Printf("   Project: %s\n", state.ProjectName)
		fmt.Printf("   Task: %s\n", state.TaskName)
		fmt.Printf("   Duration: %s\n", elapsed)
		fmt.Printf("   Entry ID: %d\n", entry.ID)
		
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
		
		if !state.Active {
			fmt.Println("⏸️  No timer is currently running.")
			fmt.Println("\nRun 'paymo time start <project> <task>' to start tracking.")
			return nil
		}
		
		fmt.Printf("⏱️  Timer Running\n")
		fmt.Printf("   Project: %s\n", state.ProjectName)
		fmt.Printf("   Task: %s\n", state.TaskName)
		if state.Description != "" {
			fmt.Printf("   Description: %s\n", state.Description)
		}
		fmt.Printf("   Started: %s\n", state.StartTime.Format("15:04:05"))
		fmt.Printf("   Elapsed: %s\n", state.FormatElapsedTime())
		
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
  paymo time log --project 123      # Filter by project`,
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
		
		// Fetch entries
		entries, err := client.GetEntries(opts)
		if err != nil {
			return fmt.Errorf("fetching entries: %w", err)
		}
		
		// Format output
		format := viper.GetString("format")
		formatter := output.NewFormatter(format)
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