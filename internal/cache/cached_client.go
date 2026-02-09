package cache

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ComputClaw/paymo-cli/internal/api"
)

// CachedClient wraps a PaymoAPI implementation with SQLite caching.
// Read methods check cache first; mutations pass through and invalidate.
type CachedClient struct {
	inner api.PaymoAPI
	store *Store
}

// NewCachedClient creates a new cached wrapper around the given client.
func NewCachedClient(inner api.PaymoAPI, store *Store) *CachedClient {
	return &CachedClient{inner: inner, store: store}
}

// --- Auth (not cached) ---

func (c *CachedClient) GetMe() (*api.User, error) {
	key := "me"
	var cached api.User
	if err := c.store.Get("me", key, &cached); err == nil {
		return &cached, nil
	}
	user, err := c.inner.GetMe()
	if err != nil {
		return nil, err
	}
	c.store.Set("me", key, user)
	return user, nil
}

func (c *CachedClient) ValidateAuth() error {
	return c.inner.ValidateAuth()
}

// --- Projects ---

func (c *CachedClient) GetProjects(opts *api.ProjectListOptions) ([]api.Project, error) {
	key := projectsKey(opts)
	var cached []api.Project
	if err := c.store.Get("projects", key, &cached); err == nil {
		return cached, nil
	}
	projects, err := c.inner.GetProjects(opts)
	if err != nil {
		if isNetworkError(err) {
			var stale []api.Project
			if c.store.GetStale("projects", key, &stale) == nil {
				return stale, nil
			}
		}
		return nil, err
	}
	c.store.Set("projects", key, projects)
	c.indexProjects(projects)
	return projects, nil
}

func (c *CachedClient) GetProject(id int) (*api.Project, error) {
	key := fmt.Sprintf("%d", id)
	var cached api.Project
	if err := c.store.Get("project", key, &cached); err == nil {
		return &cached, nil
	}
	project, err := c.inner.GetProject(id)
	if err != nil {
		if isNetworkError(err) {
			var stale api.Project
			if c.store.GetStale("project", key, &stale) == nil {
				return &stale, nil
			}
		}
		return nil, err
	}
	c.store.Set("project", key, project)
	c.indexProject(project)
	return project, nil
}

func (c *CachedClient) GetProjectByName(name string) (*api.Project, error) {
	nameLower := strings.ToLower(name)
	// Check name index for fast ID lookup
	if id, err := c.store.LookupName("project", nameLower, 0); err == nil {
		return c.GetProject(id)
	}
	// Cache miss — hit the API
	project, err := c.inner.GetProjectByName(name)
	if err != nil {
		return nil, err
	}
	c.store.Set("project", fmt.Sprintf("%d", project.ID), project)
	c.indexProject(project)
	return project, nil
}

func (c *CachedClient) CreateProject(req *api.CreateProjectRequest) (*api.Project, error) {
	project, err := c.inner.CreateProject(req)
	if err != nil {
		return nil, err
	}
	c.store.InvalidateType("projects")
	c.store.Set("project", fmt.Sprintf("%d", project.ID), project)
	c.indexProject(project)
	return project, nil
}

func (c *CachedClient) ArchiveProject(id int) error {
	if err := c.inner.ArchiveProject(id); err != nil {
		return err
	}
	c.store.InvalidateType("projects", "project", "project_by_name")
	return nil
}

// --- Tasks ---

func (c *CachedClient) GetTasks(opts *api.TaskListOptions) ([]api.Task, error) {
	key := tasksKey(opts)
	var cached []api.Task
	if err := c.store.Get("tasks", key, &cached); err == nil {
		return cached, nil
	}
	tasks, err := c.inner.GetTasks(opts)
	if err != nil {
		if isNetworkError(err) {
			var stale []api.Task
			if c.store.GetStale("tasks", key, &stale) == nil {
				return stale, nil
			}
		}
		return nil, err
	}
	c.store.Set("tasks", key, tasks)
	c.indexTasks(tasks)
	return tasks, nil
}

func (c *CachedClient) GetTask(id int) (*api.Task, error) {
	key := fmt.Sprintf("%d", id)
	var cached api.Task
	if err := c.store.Get("task", key, &cached); err == nil {
		return &cached, nil
	}
	task, err := c.inner.GetTask(id)
	if err != nil {
		if isNetworkError(err) {
			var stale api.Task
			if c.store.GetStale("task", key, &stale) == nil {
				return &stale, nil
			}
		}
		return nil, err
	}
	c.store.Set("task", key, task)
	c.indexTask(task)
	return task, nil
}

func (c *CachedClient) GetTaskByName(projectID int, name string) (*api.Task, error) {
	nameLower := strings.ToLower(name)
	if id, err := c.store.LookupName("task", nameLower, projectID); err == nil {
		return c.GetTask(id)
	}
	task, err := c.inner.GetTaskByName(projectID, name)
	if err != nil {
		return nil, err
	}
	c.store.Set("task", fmt.Sprintf("%d", task.ID), task)
	c.indexTask(task)
	return task, nil
}

func (c *CachedClient) CreateTask(req *api.CreateTaskRequest) (*api.Task, error) {
	task, err := c.inner.CreateTask(req)
	if err != nil {
		return nil, err
	}
	c.store.InvalidateType("tasks")
	c.store.Set("task", fmt.Sprintf("%d", task.ID), task)
	c.indexTask(task)
	return task, nil
}

func (c *CachedClient) CompleteTask(id int) error {
	if err := c.inner.CompleteTask(id); err != nil {
		return err
	}
	c.store.InvalidateType("tasks", "task", "task_by_name")
	return nil
}

func (c *CachedClient) GetTaskLists(projectID int) ([]api.TaskList, error) {
	key := fmt.Sprintf("project=%d", projectID)
	var cached []api.TaskList
	if err := c.store.Get("tasklists", key, &cached); err == nil {
		return cached, nil
	}
	lists, err := c.inner.GetTaskLists(projectID)
	if err != nil {
		return nil, err
	}
	c.store.Set("tasklists", key, lists)
	return lists, nil
}

// --- Time Entries ---

func (c *CachedClient) GetEntries(opts *api.EntryListOptions) ([]api.TimeEntry, error) {
	key := entriesKey(opts)
	var cached []api.TimeEntry
	if err := c.store.Get("entries", key, &cached); err == nil {
		return cached, nil
	}
	entries, err := c.inner.GetEntries(opts)
	if err != nil {
		if isNetworkError(err) {
			var stale []api.TimeEntry
			if c.store.GetStale("entries", key, &stale) == nil {
				return stale, nil
			}
		}
		return nil, err
	}
	c.store.Set("entries", key, entries)
	return entries, nil
}

func (c *CachedClient) GetEntry(id int) (*api.TimeEntry, error) {
	key := fmt.Sprintf("%d", id)
	var cached api.TimeEntry
	if err := c.store.Get("entry", key, &cached); err == nil {
		return &cached, nil
	}
	entry, err := c.inner.GetEntry(id)
	if err != nil {
		return nil, err
	}
	c.store.Set("entry", key, entry)
	return entry, nil
}

func (c *CachedClient) CreateEntry(req *api.CreateTimeEntryRequest) (*api.TimeEntry, error) {
	entry, err := c.inner.CreateEntry(req)
	if err != nil {
		return nil, err
	}
	c.store.InvalidateType("entries", "active_entry")
	return entry, nil
}

func (c *CachedClient) UpdateEntry(id int, req *api.UpdateTimeEntryRequest) (*api.TimeEntry, error) {
	entry, err := c.inner.UpdateEntry(id, req)
	if err != nil {
		return nil, err
	}
	c.store.InvalidateType("entries", "entry")
	return entry, nil
}

func (c *CachedClient) DeleteEntry(id int) error {
	if err := c.inner.DeleteEntry(id); err != nil {
		return err
	}
	c.store.InvalidateType("entries", "entry")
	return nil
}

func (c *CachedClient) GetTodayEntries(userID int) ([]api.TimeEntry, error) {
	// Delegate to inner — this is a convenience wrapper that calls GetEntries
	// with date ranges, and the short TTL on "entries" already covers it.
	return c.inner.GetTodayEntries(userID)
}

func (c *CachedClient) GetActiveEntry(userID int) (*api.TimeEntry, error) {
	// Never cache active entries — stale data here is dangerous
	return c.inner.GetActiveEntry(userID)
}

func (c *CachedClient) StartEntry(taskID int, description string) (*api.TimeEntry, error) {
	entry, err := c.inner.StartEntry(taskID, description)
	if err != nil {
		return nil, err
	}
	c.store.InvalidateType("entries", "active_entry")
	return entry, nil
}

func (c *CachedClient) StopEntry(id int) (*api.TimeEntry, error) {
	entry, err := c.inner.StopEntry(id)
	if err != nil {
		return nil, err
	}
	c.store.InvalidateType("entries", "active_entry")
	return entry, nil
}

// --- Name indexing helpers ---

func (c *CachedClient) indexProject(p *api.Project) {
	c.store.IndexName("project", strings.ToLower(p.Name), p.ID, 0)
}

func (c *CachedClient) indexProjects(projects []api.Project) {
	for i := range projects {
		c.indexProject(&projects[i])
	}
}

func (c *CachedClient) indexTask(t *api.Task) {
	c.store.IndexName("task", strings.ToLower(t.Name), t.ID, t.ProjectID)
}

func (c *CachedClient) indexTasks(tasks []api.Task) {
	for i := range tasks {
		c.indexTask(&tasks[i])
	}
}

// --- Network error detection ---

func isNetworkError(err error) bool {
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		return false // server responded, not a network error
	}
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "network is unreachable") ||
		strings.Contains(errStr, "i/o timeout") ||
		strings.Contains(errStr, "dial tcp")
}
