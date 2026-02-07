package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const TimerStateFile = "timer.json"

// TimerState holds the current timer state
type TimerState struct {
	Active      bool      `json:"active"`
	EntryID     int       `json:"entry_id,omitempty"`
	ProjectID   int       `json:"project_id,omitempty"`
	TaskID      int       `json:"task_id,omitempty"`
	ProjectName string    `json:"project_name,omitempty"`
	TaskName    string    `json:"task_name,omitempty"`
	Description string    `json:"description,omitempty"`
	StartTime   time.Time `json:"start_time,omitempty"`
}

// GetTimerStatePath returns the path to the timer state file
func GetTimerStatePath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, TimerStateFile), nil
}

// LoadTimerState loads the current timer state
func LoadTimerState() (*TimerState, error) {
	path, err := GetTimerStatePath()
	if err != nil {
		return nil, err
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &TimerState{Active: false}, nil
		}
		return nil, fmt.Errorf("reading timer state: %w", err)
	}
	
	var state TimerState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing timer state: %w", err)
	}
	
	return &state, nil
}

// SaveTimerState saves the current timer state
func SaveTimerState(state *TimerState) error {
	dir, err := EnsureConfigDir()
	if err != nil {
		return err
	}
	
	path := filepath.Join(dir, TimerStateFile)
	
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling timer state: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing timer state: %w", err)
	}
	
	return nil
}

// ClearTimerState clears the timer state (timer stopped)
func ClearTimerState() error {
	return SaveTimerState(&TimerState{Active: false})
}

// GetElapsedTime returns the elapsed time for the current timer
func (s *TimerState) GetElapsedTime() time.Duration {
	if !s.Active || s.StartTime.IsZero() {
		return 0
	}
	return time.Since(s.StartTime)
}

// FormatElapsedTime returns a human-readable elapsed time
func (s *TimerState) FormatElapsedTime() string {
	elapsed := s.GetElapsedTime()
	hours := int(elapsed.Hours())
	minutes := int(elapsed.Minutes()) % 60
	seconds := int(elapsed.Seconds()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}