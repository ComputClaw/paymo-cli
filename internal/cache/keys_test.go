package cache

import (
	"testing"
	"time"

	"github.com/ComputClaw/paymo-cli/internal/api"
)

func TestProjectsKey(t *testing.T) {
	tests := []struct {
		name     string
		opts     *api.ProjectListOptions
		expected string
	}{
		{"nil", nil, "all"},
		{"empty", &api.ProjectListOptions{}, "all"},
		{"active only", &api.ProjectListOptions{ActiveOnly: true}, "active=true"},
		{"client filter", &api.ProjectListOptions{ClientID: 5}, "client=5"},
		{"user filter", &api.ProjectListOptions{UserID: 10}, "user=10"},
		{"include tasks", &api.ProjectListOptions{IncludeTasks: true}, "inc_tasks"},
		{"include client", &api.ProjectListOptions{IncludeClient: true}, "inc_client"},
		{"combined", &api.ProjectListOptions{ActiveOnly: true, ClientID: 5, IncludeTasks: true}, "active=true|client=5|inc_tasks"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := projectsKey(tc.opts)
			if got != tc.expected {
				t.Errorf("projectsKey() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestTasksKey(t *testing.T) {
	tests := []struct {
		name     string
		opts     *api.TaskListOptions
		expected string
	}{
		{"nil", nil, "all"},
		{"project filter", &api.TaskListOptions{ProjectID: 10}, "project=10|completed=false"},
		{"include completed", &api.TaskListOptions{IncludeCompleted: true}, "completed=true"},
		{"combined", &api.TaskListOptions{ProjectID: 10, UserID: 5, IncludeCompleted: true, IncludeProject: true}, "project=10|user=5|completed=true|inc_project"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tasksKey(tc.opts)
			if got != tc.expected {
				t.Errorf("tasksKey() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestEntriesKey(t *testing.T) {
	date1 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		opts     *api.EntryListOptions
		expected string
	}{
		{"nil", nil, "all"},
		{"empty", &api.EntryListOptions{}, "all"},
		{"user filter", &api.EntryListOptions{UserID: 1}, "user=1"},
		{"date range", &api.EntryListOptions{StartDate: date1, EndDate: date2}, "start=2026-01-15|end=2026-01-16"},
		{"combined", &api.EntryListOptions{UserID: 1, ProjectID: 5, TaskID: 10, StartDate: date1}, "user=1|project=5|task=10|start=2026-01-15"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := entriesKey(tc.opts)
			if got != tc.expected {
				t.Errorf("entriesKey() = %q, want %q", got, tc.expected)
			}
		})
	}
}
