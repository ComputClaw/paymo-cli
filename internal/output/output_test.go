package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ComputClaw/paymo-cli/internal/api"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		{"table", "table"},
		{"TABLE", "table"},
		{"json", "json"},
		{"JSON", "json"},
		{"csv", "csv"},
		{"CSV", "csv"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			f := NewFormatter(tt.format)
			if f.Format != tt.expected {
				t.Errorf("expected format '%s', got '%s'", tt.expected, f.Format)
			}
		})
	}
}

func TestFormatTimeEntries_JSON(t *testing.T) {
	entries := []api.TimeEntry{
		{
			ID:          1,
			TaskID:      100,
			UserID:      10,
			StartTime:   time.Date(2026, 2, 7, 9, 0, 0, 0, time.UTC),
			Duration:    3600,
			Description: "Test entry",
		},
	}

	var buf bytes.Buffer
	f := NewFormatter("json")
	f.Writer = &buf

	err := f.FormatTimeEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's valid JSON
	var result []api.TimeEntry
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 entry, got %d", len(result))
	}

	if result[0].ID != 1 {
		t.Errorf("expected ID 1, got %d", result[0].ID)
	}
}

func TestFormatTimeEntries_Table(t *testing.T) {
	project := &api.Project{ID: 1, Name: "Test Project"}
	task := &api.Task{ID: 100, Name: "Test Task", ProjectID: 1}

	entries := []api.TimeEntry{
		{
			ID:          1,
			TaskID:      100,
			UserID:      10,
			StartTime:   time.Date(2026, 2, 7, 9, 0, 0, 0, time.UTC),
			Duration:    3600,
			Description: "Test entry",
			Project:     project,
			Task:        task,
		},
	}

	var buf bytes.Buffer
	f := NewFormatter("table")
	f.Writer = &buf

	err := f.FormatTimeEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check for table structure
	if !strings.Contains(output, "─") {
		t.Error("expected table borders in output")
	}

	// Check for headers
	if !strings.Contains(output, "ID") {
		t.Error("expected 'ID' header in output")
	}

	if !strings.Contains(output, "Project") {
		t.Error("expected 'Project' header in output")
	}

	// Check for data
	if !strings.Contains(output, "Test Project") {
		t.Error("expected project name in output")
	}

	if !strings.Contains(output, "Test Task") {
		t.Error("expected task name in output")
	}

	// Check for total
	if !strings.Contains(output, "Total:") {
		t.Error("expected 'Total:' in output")
	}
}

func TestFormatTimeEntries_Empty(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("table")
	f.Writer = &buf

	err := f.FormatTimeEntries([]api.TimeEntry{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No time entries found") {
		t.Error("expected 'No time entries found' message")
	}
}

func TestFormatProjects_JSON(t *testing.T) {
	projects := []api.Project{
		{
			ID:       1,
			Name:     "Project One",
			Code:     "P1",
			Active:   true,
			Billable: true,
		},
		{
			ID:       2,
			Name:     "Project Two",
			Code:     "P2",
			Active:   false,
			Billable: false,
		},
	}

	var buf bytes.Buffer
	f := NewFormatter("json")
	f.Writer = &buf

	err := f.FormatProjects(projects)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []api.Project
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 projects, got %d", len(result))
	}
}

func TestFormatProjects_Table(t *testing.T) {
	projects := []api.Project{
		{
			ID:       1,
			Name:     "Project One",
			Code:     "P1",
			Active:   true,
			Billable: true,
		},
	}

	var buf bytes.Buffer
	f := NewFormatter("table")
	f.Writer = &buf

	err := f.FormatProjects(projects)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Project One") {
		t.Error("expected project name in output")
	}

	if !strings.Contains(output, "Active") {
		t.Error("expected 'Active' status in output")
	}

	if !strings.Contains(output, "Total:") {
		t.Error("expected 'Total:' in output")
	}
}

func TestFormatTasks_JSON(t *testing.T) {
	tasks := []api.Task{
		{
			ID:        1,
			Name:      "Task One",
			ProjectID: 100,
			Complete:  false,
			Billable:  true,
		},
	}

	var buf bytes.Buffer
	f := NewFormatter("json")
	f.Writer = &buf

	err := f.FormatTasks(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []api.Task
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 task, got %d", len(result))
	}

	if result[0].Name != "Task One" {
		t.Errorf("expected name 'Task One', got '%s'", result[0].Name)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, expected %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds  int
		expected string
	}{
		{0, "0m"},
		{60, "1m"},
		{90, "1m"},
		{3600, "1h 0m"},
		{3660, "1h 1m"},
		{7200, "2h 0m"},
		{7320, "2h 2m"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.seconds)
			if result != tt.expected {
				t.Errorf("formatDuration(%d) = %q, expected %q", tt.seconds, result, tt.expected)
			}
		})
	}
}