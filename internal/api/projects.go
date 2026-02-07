package api

import (
	"fmt"
	"net/url"
)

// GetProjects returns all projects with optional filtering
func (c *Client) GetProjects(opts *ProjectListOptions) ([]Project, error) {
	params := url.Values{}
	
	if opts != nil {
		whereClause := ""
		
		if opts.ActiveOnly {
			whereClause = "active=true"
		}
		if opts.ClientID > 0 {
			if whereClause != "" {
				whereClause += fmt.Sprintf(" and client_id=%d", opts.ClientID)
			} else {
				whereClause = fmt.Sprintf("client_id=%d", opts.ClientID)
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
		
		if opts.IncludeTasks {
			params.Set("include", "tasklists.tasks")
		}
		if opts.IncludeClient {
			if params.Get("include") != "" {
				params.Set("include", params.Get("include")+",client")
			} else {
				params.Set("include", "client")
			}
		}
	}
	
	var resp ProjectsResponse
	if err := c.GetWithParams("projects", params, &resp); err != nil {
		return nil, err
	}
	
	return resp.Projects, nil
}

// ProjectListOptions for filtering projects
type ProjectListOptions struct {
	ActiveOnly    bool
	ClientID      int
	UserID        int
	IncludeTasks  bool
	IncludeClient bool
}

// GetProject returns a single project by ID
func (c *Client) GetProject(id int) (*Project, error) {
	params := url.Values{}
	params.Set("include", "tasklists.tasks,client")
	
	var resp ProjectResponse
	if err := c.GetWithParams(fmt.Sprintf("projects/%d", id), params, &resp); err != nil {
		return nil, err
	}
	
	if len(resp.Projects) == 0 {
		return nil, &APIError{StatusCode: 404, Message: "project not found"}
	}
	
	return &resp.Projects[0], nil
}

// GetProjectByName finds a project by name (case-insensitive partial match)
func (c *Client) GetProjectByName(name string) (*Project, error) {
	params := url.Values{}
	params.Set("where", fmt.Sprintf("name like \"%%%s%%\"", name))
	
	var resp ProjectsResponse
	if err := c.GetWithParams("projects", params, &resp); err != nil {
		return nil, err
	}
	
	if len(resp.Projects) == 0 {
		return nil, &APIError{StatusCode: 404, Message: "project not found"}
	}
	
	// Return first match
	return &resp.Projects[0], nil
}

// CreateProject creates a new project
func (c *Client) CreateProject(req *CreateProjectRequest) (*Project, error) {
	var resp ProjectResponse
	if err := c.Post("projects", req, &resp); err != nil {
		return nil, err
	}
	
	if len(resp.Projects) == 0 {
		return nil, &APIError{StatusCode: 500, Message: "no project returned"}
	}
	
	return &resp.Projects[0], nil
}

// ArchiveProject archives a project
func (c *Client) ArchiveProject(id int) error {
	type archiveReq struct {
		Active bool `json:"active"`
	}
	return c.Put(fmt.Sprintf("projects/%d", id), &archiveReq{Active: false}, nil)
}