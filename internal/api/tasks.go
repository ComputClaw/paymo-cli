package api

import (
	"fmt"
	"net/url"
	"strings"
)

// GetTasks returns tasks with optional filtering
func (c *Client) GetTasks(opts *TaskListOptions) ([]Task, error) {
	params := url.Values{}
	
	if opts != nil {
		whereClause := ""
		
		if opts.ProjectID > 0 {
			whereClause = fmt.Sprintf("project_id=%d", opts.ProjectID)
		}
		if opts.TaskListID > 0 {
			if whereClause != "" {
				whereClause += fmt.Sprintf(" and tasklist_id=%d", opts.TaskListID)
			} else {
				whereClause = fmt.Sprintf("tasklist_id=%d", opts.TaskListID)
			}
		}
		if !opts.IncludeCompleted {
			if whereClause != "" {
				whereClause += " and complete=false"
			} else {
				whereClause = "complete=false"
			}
		}
		if opts.UserID > 0 {
			if whereClause != "" {
				whereClause += fmt.Sprintf(" and users in (%d)", opts.UserID)
			} else {
				whereClause = fmt.Sprintf("users in (%d)", opts.UserID)
			}
		}
		
		if whereClause != "" {
			params.Set("where", whereClause)
		}
		
		if opts.IncludeProject {
			params.Set("include", "project")
		}
	}
	
	var resp TasksResponse
	if err := c.GetWithParams("tasks", params, &resp); err != nil {
		return nil, err
	}
	
	return resp.Tasks, nil
}

// TaskListOptions for filtering tasks
type TaskListOptions struct {
	ProjectID        int
	TaskListID       int
	UserID           int
	IncludeCompleted bool
	IncludeProject   bool
}

// GetTask returns a single task by ID
func (c *Client) GetTask(id int) (*Task, error) {
	params := url.Values{}
	params.Set("include", "project")
	
	var resp TasksResponse
	if err := c.GetWithParams(fmt.Sprintf("tasks/%d", id), params, &resp); err != nil {
		return nil, err
	}
	
	if len(resp.Tasks) == 0 {
		return nil, &APIError{StatusCode: 404, Message: "task not found"}
	}
	
	return &resp.Tasks[0], nil
}

// GetTaskByName finds a task by name within a project
func (c *Client) GetTaskByName(projectID int, name string) (*Task, error) {
	params := url.Values{}
	sanitized := strings.NewReplacer("\"", "", "\\", "", "'", "").Replace(name)
	params.Set("where", fmt.Sprintf("project_id=%d and name like \"%%%s%%\"", projectID, sanitized))
	
	var resp TasksResponse
	if err := c.GetWithParams("tasks", params, &resp); err != nil {
		return nil, err
	}
	
	if len(resp.Tasks) == 0 {
		return nil, &APIError{StatusCode: 404, Message: "task not found"}
	}
	
	return &resp.Tasks[0], nil
}

// CreateTask creates a new task
func (c *Client) CreateTask(req *CreateTaskRequest) (*Task, error) {
	var resp TasksResponse
	if err := c.Post("tasks", req, &resp); err != nil {
		return nil, err
	}
	
	if len(resp.Tasks) == 0 {
		return nil, &APIError{StatusCode: 500, Message: "no task returned"}
	}
	
	return &resp.Tasks[0], nil
}

// CompleteTask marks a task as complete
func (c *Client) CompleteTask(id int) error {
	type completeReq struct {
		Complete bool `json:"complete"`
	}
	return c.Put(fmt.Sprintf("tasks/%d", id), &completeReq{Complete: true}, nil)
}

// GetTaskLists returns task lists for a project
func (c *Client) GetTaskLists(projectID int) ([]TaskList, error) {
	params := url.Values{}
	params.Set("where", fmt.Sprintf("project_id=%d", projectID))
	
	var resp struct {
		TaskLists []TaskList `json:"tasklists"`
	}
	if err := c.GetWithParams("tasklists", params, &resp); err != nil {
		return nil, err
	}
	
	return resp.TaskLists, nil
}