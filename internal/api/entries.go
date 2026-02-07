package api

import (
	"fmt"
	"net/url"
	"time"
)

// GetEntries returns time entries with optional filtering
func (c *Client) GetEntries(opts *EntryListOptions) ([]TimeEntry, error) {
	params := url.Values{}
	
	if opts != nil {
		if opts.UserID > 0 {
			params.Set("where", fmt.Sprintf("user_id=%d", opts.UserID))
		}
		if opts.ProjectID > 0 {
			if params.Get("where") != "" {
				params.Set("where", params.Get("where")+fmt.Sprintf(" and project_id=%d", opts.ProjectID))
			} else {
				params.Set("where", fmt.Sprintf("project_id=%d", opts.ProjectID))
			}
		}
		if opts.TaskID > 0 {
			if params.Get("where") != "" {
				params.Set("where", params.Get("where")+fmt.Sprintf(" and task_id=%d", opts.TaskID))
			} else {
				params.Set("where", fmt.Sprintf("task_id=%d", opts.TaskID))
			}
		}
		if !opts.StartDate.IsZero() {
			dateStr := opts.StartDate.Format("2006-01-02T15:04:05Z")
			if params.Get("where") != "" {
				params.Set("where", params.Get("where")+fmt.Sprintf(" and start_time>=\"%s\"", dateStr))
			} else {
				params.Set("where", fmt.Sprintf("start_time>=\"%s\"", dateStr))
			}
		}
		if !opts.EndDate.IsZero() {
			dateStr := opts.EndDate.Format("2006-01-02T15:04:05Z")
			if params.Get("where") != "" {
				params.Set("where", params.Get("where")+fmt.Sprintf(" and start_time<=\"%s\"", dateStr))
			} else {
				params.Set("where", fmt.Sprintf("start_time<=\"%s\"", dateStr))
			}
		}
		if opts.IncludeTask {
			params.Set("include", "task")
		}
		if opts.IncludeProject {
			if params.Get("include") != "" {
				params.Set("include", params.Get("include")+",task.project")
			} else {
				params.Set("include", "task.project")
			}
		}
	}
	
	var resp TimeEntriesResponse
	if err := c.GetWithParams("entries", params, &resp); err != nil {
		return nil, err
	}
	
	return resp.Entries, nil
}

// EntryListOptions for filtering time entries
type EntryListOptions struct {
	UserID         int
	ProjectID      int
	TaskID         int
	StartDate      time.Time
	EndDate        time.Time
	IncludeTask    bool
	IncludeProject bool
}

// GetEntry returns a single time entry by ID
func (c *Client) GetEntry(id int) (*TimeEntry, error) {
	var resp TimeEntryResponse
	if err := c.Get(fmt.Sprintf("entries/%d", id), &resp); err != nil {
		return nil, err
	}
	
	if len(resp.Entries) == 0 {
		return nil, &APIError{StatusCode: 404, Message: "entry not found"}
	}
	
	return &resp.Entries[0], nil
}

// CreateEntry creates a new time entry
func (c *Client) CreateEntry(req *CreateTimeEntryRequest) (*TimeEntry, error) {
	var resp TimeEntryResponse
	if err := c.Post("entries", req, &resp); err != nil {
		return nil, err
	}
	
	if len(resp.Entries) == 0 {
		return nil, &APIError{StatusCode: 500, Message: "no entry returned"}
	}
	
	return &resp.Entries[0], nil
}

// UpdateEntry updates an existing time entry
func (c *Client) UpdateEntry(id int, req *UpdateTimeEntryRequest) (*TimeEntry, error) {
	var resp TimeEntryResponse
	if err := c.Put(fmt.Sprintf("entries/%d", id), req, &resp); err != nil {
		return nil, err
	}
	
	if len(resp.Entries) == 0 {
		return nil, &APIError{StatusCode: 500, Message: "no entry returned"}
	}
	
	return &resp.Entries[0], nil
}

// DeleteEntry deletes a time entry
func (c *Client) DeleteEntry(id int) error {
	return c.Delete(fmt.Sprintf("entries/%d", id))
}

// GetTodayEntries returns entries for today
func (c *Client) GetTodayEntries(userID int) ([]TimeEntry, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	return c.GetEntries(&EntryListOptions{
		UserID:         userID,
		StartDate:      startOfDay,
		EndDate:        endOfDay,
		IncludeTask:    true,
		IncludeProject: true,
	})
}

// GetActiveEntry returns the currently running entry (no end_time) for a user
func (c *Client) GetActiveEntry(userID int) (*TimeEntry, error) {
	params := url.Values{}
	params.Set("where", fmt.Sprintf("user_id=%d and end_time=\"\"", userID))
	params.Set("include", "task.project")
	
	var resp TimeEntriesResponse
	if err := c.GetWithParams("entries", params, &resp); err != nil {
		return nil, err
	}
	
	if len(resp.Entries) == 0 {
		return nil, nil // No active entry
	}
	
	return &resp.Entries[0], nil
}

// StartEntry creates a new time entry with only start time (running timer)
func (c *Client) StartEntry(taskID int, description string) (*TimeEntry, error) {
	req := &CreateTimeEntryRequest{
		TaskID:      taskID,
		StartTime:   time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		Description: description,
	}
	
	return c.CreateEntry(req)
}

// StopEntry stops a running time entry by setting the end time
func (c *Client) StopEntry(id int) (*TimeEntry, error) {
	endTime := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	req := &UpdateTimeEntryRequest{
		EndTime: &endTime,
	}
	
	return c.UpdateEntry(id, req)
}