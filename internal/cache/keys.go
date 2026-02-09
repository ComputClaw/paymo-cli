package cache

import (
	"fmt"
	"strings"

	"github.com/ComputClaw/paymo-cli/internal/api"
)

// projectsKey derives a cache key for GetProjects with the given options.
func projectsKey(opts *api.ProjectListOptions) string {
	if opts == nil {
		return "all"
	}
	var parts []string
	if opts.ActiveOnly {
		parts = append(parts, "active=true")
	}
	if opts.ClientID > 0 {
		parts = append(parts, fmt.Sprintf("client=%d", opts.ClientID))
	}
	if opts.UserID > 0 {
		parts = append(parts, fmt.Sprintf("user=%d", opts.UserID))
	}
	if opts.IncludeTasks {
		parts = append(parts, "inc_tasks")
	}
	if opts.IncludeClient {
		parts = append(parts, "inc_client")
	}
	if len(parts) == 0 {
		return "all"
	}
	return strings.Join(parts, "|")
}

// tasksKey derives a cache key for GetTasks with the given options.
func tasksKey(opts *api.TaskListOptions) string {
	if opts == nil {
		return "all"
	}
	var parts []string
	if opts.ProjectID > 0 {
		parts = append(parts, fmt.Sprintf("project=%d", opts.ProjectID))
	}
	if opts.TaskListID > 0 {
		parts = append(parts, fmt.Sprintf("tasklist=%d", opts.TaskListID))
	}
	if opts.UserID > 0 {
		parts = append(parts, fmt.Sprintf("user=%d", opts.UserID))
	}
	if opts.IncludeCompleted {
		parts = append(parts, "completed=true")
	} else {
		parts = append(parts, "completed=false")
	}
	if opts.IncludeProject {
		parts = append(parts, "inc_project")
	}
	if len(parts) == 0 {
		return "all"
	}
	return strings.Join(parts, "|")
}

// entriesKey derives a cache key for GetEntries with the given options.
func entriesKey(opts *api.EntryListOptions) string {
	if opts == nil {
		return "all"
	}
	var parts []string
	if opts.UserID > 0 {
		parts = append(parts, fmt.Sprintf("user=%d", opts.UserID))
	}
	if opts.ProjectID > 0 {
		parts = append(parts, fmt.Sprintf("project=%d", opts.ProjectID))
	}
	if opts.TaskID > 0 {
		parts = append(parts, fmt.Sprintf("task=%d", opts.TaskID))
	}
	if !opts.StartDate.IsZero() {
		parts = append(parts, fmt.Sprintf("start=%s", opts.StartDate.Format("2006-01-02")))
	}
	if !opts.EndDate.IsZero() {
		parts = append(parts, fmt.Sprintf("end=%s", opts.EndDate.Format("2006-01-02")))
	}
	if len(parts) == 0 {
		return "all"
	}
	return strings.Join(parts, "|")
}
