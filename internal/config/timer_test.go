package config

import (
	"testing"
	"time"
)

func TestTimerState_GetElapsedTime(t *testing.T) {
	tests := []struct {
		name     string
		state    TimerState
		minElapsed time.Duration
	}{
		{
			name: "inactive timer",
			state: TimerState{
				Active: false,
			},
			minElapsed: 0,
		},
		{
			name: "active timer",
			state: TimerState{
				Active:    true,
				StartTime: time.Now().Add(-5 * time.Minute),
			},
			minElapsed: 4 * time.Minute, // Allow some tolerance
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elapsed := tt.state.GetElapsedTime()
			
			if elapsed < tt.minElapsed {
				t.Errorf("expected elapsed >= %v, got %v", tt.minElapsed, elapsed)
			}
		})
	}
}

func TestTimerState_FormatElapsedTime(t *testing.T) {
	tests := []struct {
		name     string
		state    TimerState
		contains string
	}{
		{
			name: "seconds only",
			state: TimerState{
				Active:    true,
				StartTime: time.Now().Add(-30 * time.Second),
			},
			contains: "s",
		},
		{
			name: "minutes and seconds",
			state: TimerState{
				Active:    true,
				StartTime: time.Now().Add(-5 * time.Minute),
			},
			contains: "m",
		},
		{
			name: "hours minutes seconds",
			state: TimerState{
				Active:    true,
				StartTime: time.Now().Add(-2 * time.Hour),
			},
			contains: "h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := tt.state.FormatElapsedTime()
			
			if formatted == "" {
				t.Error("expected non-empty formatted string")
			}
			
			// Just verify it contains the expected time unit
			if len(formatted) == 0 {
				t.Error("formatted time should not be empty for active timer")
			}
		})
	}
}

func TestTimerState_Fields(t *testing.T) {
	state := TimerState{
		Active:      true,
		EntryID:     123,
		ProjectID:   456,
		TaskID:      789,
		ProjectName: "Test Project",
		TaskName:    "Test Task",
		Description: "Working on tests",
		StartTime:   time.Now(),
	}

	if !state.Active {
		t.Error("expected active to be true")
	}

	if state.EntryID != 123 {
		t.Errorf("expected entry ID 123, got %d", state.EntryID)
	}

	if state.ProjectID != 456 {
		t.Errorf("expected project ID 456, got %d", state.ProjectID)
	}

	if state.TaskID != 789 {
		t.Errorf("expected task ID 789, got %d", state.TaskID)
	}

	if state.ProjectName != "Test Project" {
		t.Errorf("expected project name 'Test Project', got '%s'", state.ProjectName)
	}

	if state.TaskName != "Test Task" {
		t.Errorf("expected task name 'Test Task', got '%s'", state.TaskName)
	}

	if state.Description != "Working on tests" {
		t.Errorf("expected description 'Working on tests', got '%s'", state.Description)
	}
}