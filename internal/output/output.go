package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ComputClaw/paymo-cli/internal/api"
)

// Formatter handles output formatting
type Formatter struct {
	Format string
	Writer io.Writer
}

// NewFormatter creates a new formatter with the specified format
func NewFormatter(format string) *Formatter {
	return &Formatter{
		Format: strings.ToLower(format),
		Writer: os.Stdout,
	}
}

// FormatTimeEntries outputs time entries in the specified format
func (f *Formatter) FormatTimeEntries(entries []api.TimeEntry) error {
	switch f.Format {
	case "json":
		return f.formatJSON(entries)
	case "csv":
		return f.formatEntriesCSV(entries)
	default:
		return f.formatEntriesTable(entries)
	}
}

// FormatProjects outputs projects in the specified format
func (f *Formatter) FormatProjects(projects []api.Project) error {
	switch f.Format {
	case "json":
		return f.formatJSON(projects)
	case "csv":
		return f.formatProjectsCSV(projects)
	default:
		return f.formatProjectsTable(projects)
	}
}

// FormatTasks outputs tasks in the specified format
func (f *Formatter) FormatTasks(tasks []api.Task) error {
	switch f.Format {
	case "json":
		return f.formatJSON(tasks)
	case "csv":
		return f.formatTasksCSV(tasks)
	default:
		return f.formatTasksTable(tasks)
	}
}

// formatJSON outputs data as JSON
func (f *Formatter) formatJSON(data interface{}) error {
	encoder := json.NewEncoder(f.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// formatEntriesTable outputs entries as a table
func (f *Formatter) formatEntriesTable(entries []api.TimeEntry) error {
	if len(entries) == 0 {
		fmt.Fprintln(f.Writer, "No time entries found.")
		return nil
	}

	// Calculate column widths
	idWidth := 6
	projectWidth := 20
	taskWidth := 20
	durationWidth := 10
	dateWidth := 10
	descWidth := 30

	// Print header
	fmt.Fprintf(f.Writer, "┌%s┬%s┬%s┬%s┬%s┬%s┐\n",
		strings.Repeat("─", idWidth+2),
		strings.Repeat("─", projectWidth+2),
		strings.Repeat("─", taskWidth+2),
		strings.Repeat("─", durationWidth+2),
		strings.Repeat("─", dateWidth+2),
		strings.Repeat("─", descWidth+2))

	fmt.Fprintf(f.Writer, "│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
		idWidth, "ID",
		projectWidth, "Project",
		taskWidth, "Task",
		durationWidth, "Duration",
		dateWidth, "Date",
		descWidth, "Description")

	fmt.Fprintf(f.Writer, "├%s┼%s┼%s┼%s┼%s┼%s┤\n",
		strings.Repeat("─", idWidth+2),
		strings.Repeat("─", projectWidth+2),
		strings.Repeat("─", taskWidth+2),
		strings.Repeat("─", durationWidth+2),
		strings.Repeat("─", dateWidth+2),
		strings.Repeat("─", descWidth+2))

	// Print entries
	var totalDuration int
	for _, e := range entries {
		projectName := "Unknown"
		taskName := "Unknown"
		if e.Project != nil {
			projectName = truncate(e.Project.Name, projectWidth)
		}
		if e.Task != nil {
			taskName = truncate(e.Task.Name, taskWidth)
		}

		duration := formatDuration(e.Duration)
		date := e.StartTime.Format("2006-01-02")
		desc := truncate(e.Description, descWidth)
		totalDuration += e.Duration

		fmt.Fprintf(f.Writer, "│ %-*d │ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
			idWidth, e.ID,
			projectWidth, projectName,
			taskWidth, taskName,
			durationWidth, duration,
			dateWidth, date,
			descWidth, desc)
	}

	// Print footer with total
	fmt.Fprintf(f.Writer, "└%s┴%s┴%s┴%s┴%s┴%s┘\n",
		strings.Repeat("─", idWidth+2),
		strings.Repeat("─", projectWidth+2),
		strings.Repeat("─", taskWidth+2),
		strings.Repeat("─", durationWidth+2),
		strings.Repeat("─", dateWidth+2),
		strings.Repeat("─", descWidth+2))

	fmt.Fprintf(f.Writer, "\nTotal: %s (%d entries)\n", formatDuration(totalDuration), len(entries))

	return nil
}

// formatProjectsTable outputs projects as a table
func (f *Formatter) formatProjectsTable(projects []api.Project) error {
	if len(projects) == 0 {
		fmt.Fprintln(f.Writer, "No projects found.")
		return nil
	}

	idWidth := 8
	nameWidth := 30
	codeWidth := 10
	statusWidth := 8
	billableWidth := 8

	fmt.Fprintf(f.Writer, "┌%s┬%s┬%s┬%s┬%s┐\n",
		strings.Repeat("─", idWidth+2),
		strings.Repeat("─", nameWidth+2),
		strings.Repeat("─", codeWidth+2),
		strings.Repeat("─", statusWidth+2),
		strings.Repeat("─", billableWidth+2))

	fmt.Fprintf(f.Writer, "│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
		idWidth, "ID",
		nameWidth, "Name",
		codeWidth, "Code",
		statusWidth, "Status",
		billableWidth, "Billable")

	fmt.Fprintf(f.Writer, "├%s┼%s┼%s┼%s┼%s┤\n",
		strings.Repeat("─", idWidth+2),
		strings.Repeat("─", nameWidth+2),
		strings.Repeat("─", codeWidth+2),
		strings.Repeat("─", statusWidth+2),
		strings.Repeat("─", billableWidth+2))

	for _, p := range projects {
		status := "Inactive"
		if p.Active {
			status = "Active"
		}
		billable := "No"
		if p.Billable {
			billable = "Yes"
		}

		fmt.Fprintf(f.Writer, "│ %-*d │ %-*s │ %-*s │ %-*s │ %-*s │\n",
			idWidth, p.ID,
			nameWidth, truncate(p.Name, nameWidth),
			codeWidth, truncate(p.Code, codeWidth),
			statusWidth, status,
			billableWidth, billable)
	}

	fmt.Fprintf(f.Writer, "└%s┴%s┴%s┴%s┴%s┘\n",
		strings.Repeat("─", idWidth+2),
		strings.Repeat("─", nameWidth+2),
		strings.Repeat("─", codeWidth+2),
		strings.Repeat("─", statusWidth+2),
		strings.Repeat("─", billableWidth+2))

	fmt.Fprintf(f.Writer, "\nTotal: %d projects\n", len(projects))

	return nil
}

// formatTasksTable outputs tasks as a table
func (f *Formatter) formatTasksTable(tasks []api.Task) error {
	if len(tasks) == 0 {
		fmt.Fprintln(f.Writer, "No tasks found.")
		return nil
	}

	idWidth := 8
	nameWidth := 35
	projectWidth := 20
	statusWidth := 10
	dueDateWidth := 12

	fmt.Fprintf(f.Writer, "┌%s┬%s┬%s┬%s┬%s┐\n",
		strings.Repeat("─", idWidth+2),
		strings.Repeat("─", nameWidth+2),
		strings.Repeat("─", projectWidth+2),
		strings.Repeat("─", statusWidth+2),
		strings.Repeat("─", dueDateWidth+2))

	fmt.Fprintf(f.Writer, "│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
		idWidth, "ID",
		nameWidth, "Name",
		projectWidth, "Project",
		statusWidth, "Status",
		dueDateWidth, "Due Date")

	fmt.Fprintf(f.Writer, "├%s┼%s┼%s┼%s┼%s┤\n",
		strings.Repeat("─", idWidth+2),
		strings.Repeat("─", nameWidth+2),
		strings.Repeat("─", projectWidth+2),
		strings.Repeat("─", statusWidth+2),
		strings.Repeat("─", dueDateWidth+2))

	for _, t := range tasks {
		status := "Open"
		if t.Complete {
			status = "Complete"
		}
		dueDate := "-"
		if t.DueDate != "" {
			dueDate = t.DueDate
		}

		fmt.Fprintf(f.Writer, "│ %-*d │ %-*s │ %-*d │ %-*s │ %-*s │\n",
			idWidth, t.ID,
			nameWidth, truncate(t.Name, nameWidth),
			projectWidth, t.ProjectID,
			statusWidth, status,
			dueDateWidth, dueDate)
	}

	fmt.Fprintf(f.Writer, "└%s┴%s┴%s┴%s┴%s┘\n",
		strings.Repeat("─", idWidth+2),
		strings.Repeat("─", nameWidth+2),
		strings.Repeat("─", projectWidth+2),
		strings.Repeat("─", statusWidth+2),
		strings.Repeat("─", dueDateWidth+2))

	fmt.Fprintf(f.Writer, "\nTotal: %d tasks\n", len(tasks))

	return nil
}

// CSV formatters
func (f *Formatter) formatEntriesCSV(entries []api.TimeEntry) error {
	w := csv.NewWriter(f.Writer)
	defer w.Flush()

	w.Write([]string{"id", "project_id", "project_name", "task_id", "task_name", "duration", "date", "description"})

	for _, e := range entries {
		projectName := ""
		taskName := ""
		if e.Project != nil {
			projectName = e.Project.Name
		}
		if e.Task != nil {
			taskName = e.Task.Name
		}

		w.Write([]string{
			fmt.Sprintf("%d", e.ID),
			fmt.Sprintf("%d", e.Task.ProjectID),
			projectName,
			fmt.Sprintf("%d", e.TaskID),
			taskName,
			fmt.Sprintf("%d", e.Duration),
			e.StartTime.Format("2006-01-02"),
			e.Description,
		})
	}

	return nil
}

func (f *Formatter) formatProjectsCSV(projects []api.Project) error {
	w := csv.NewWriter(f.Writer)
	defer w.Flush()

	w.Write([]string{"id", "name", "code", "active", "billable", "client_id"})

	for _, p := range projects {
		w.Write([]string{
			fmt.Sprintf("%d", p.ID),
			p.Name,
			p.Code,
			fmt.Sprintf("%t", p.Active),
			fmt.Sprintf("%t", p.Billable),
			fmt.Sprintf("%d", p.ClientID),
		})
	}

	return nil
}

func (f *Formatter) formatTasksCSV(tasks []api.Task) error {
	w := csv.NewWriter(f.Writer)
	defer w.Flush()

	w.Write([]string{"id", "name", "project_id", "complete", "billable", "due_date"})

	for _, t := range tasks {
		w.Write([]string{
			fmt.Sprintf("%d", t.ID),
			t.Name,
			fmt.Sprintf("%d", t.ProjectID),
			fmt.Sprintf("%t", t.Complete),
			fmt.Sprintf("%t", t.Billable),
			t.DueDate,
		})
	}

	return nil
}

// Helper functions
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatDuration(seconds int) string {
	d := time.Duration(seconds) * time.Second
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}