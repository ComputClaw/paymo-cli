package cache

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/ComputClaw/paymo-cli/internal/api"
)

// mockAPI is a mock implementation of api.PaymoAPI for testing.
type mockAPI struct {
	getMeCalls        int
	getProjectsCalls  int
	getProjectCalls   int
	getTasksCalls     int
	getTaskCalls      int
	getEntriesCalls   int
	getEntryCalls     int
	createProjectErr  error
	archiveProjectErr error
	createTaskErr     error
	completeTaskErr   error
	createEntryErr    error
	networkErr        bool // if true, methods return a network error
}

func (m *mockAPI) GetMe() (*api.User, error) {
	m.getMeCalls++
	if m.networkErr {
		return nil, errors.New("dial tcp: connection refused")
	}
	return &api.User{ID: 1, Name: "Test User", Email: "test@test.com"}, nil
}

func (m *mockAPI) ValidateAuth() error { return nil }

func (m *mockAPI) GetProjects(opts *api.ProjectListOptions) ([]api.Project, error) {
	m.getProjectsCalls++
	if m.networkErr {
		return nil, errors.New("dial tcp: connection refused")
	}
	return []api.Project{
		{ID: 1, Name: "Project One", Active: true},
		{ID: 2, Name: "Project Two", Active: true},
	}, nil
}

func (m *mockAPI) GetProject(id int) (*api.Project, error) {
	m.getProjectCalls++
	if m.networkErr {
		return nil, errors.New("dial tcp: connection refused")
	}
	return &api.Project{ID: id, Name: fmt.Sprintf("Project %d", id), Active: true}, nil
}

func (m *mockAPI) GetProjectByName(name string) (*api.Project, error) {
	return &api.Project{ID: 1, Name: name, Active: true}, nil
}

func (m *mockAPI) CreateProject(req *api.CreateProjectRequest) (*api.Project, error) {
	if m.createProjectErr != nil {
		return nil, m.createProjectErr
	}
	return &api.Project{ID: 99, Name: req.Name, Active: true}, nil
}

func (m *mockAPI) ArchiveProject(id int) error {
	return m.archiveProjectErr
}

func (m *mockAPI) GetTasks(opts *api.TaskListOptions) ([]api.Task, error) {
	m.getTasksCalls++
	if m.networkErr {
		return nil, errors.New("dial tcp: connection refused")
	}
	return []api.Task{
		{ID: 1, Name: "Task One", ProjectID: 10},
		{ID: 2, Name: "Task Two", ProjectID: 10},
	}, nil
}

func (m *mockAPI) GetTask(id int) (*api.Task, error) {
	m.getTaskCalls++
	if m.networkErr {
		return nil, errors.New("dial tcp: connection refused")
	}
	return &api.Task{ID: id, Name: fmt.Sprintf("Task %d", id), ProjectID: 10}, nil
}

func (m *mockAPI) GetTaskByName(projectID int, name string) (*api.Task, error) {
	return &api.Task{ID: 1, Name: name, ProjectID: projectID}, nil
}

func (m *mockAPI) CreateTask(req *api.CreateTaskRequest) (*api.Task, error) {
	if m.createTaskErr != nil {
		return nil, m.createTaskErr
	}
	return &api.Task{ID: 99, Name: req.Name, ProjectID: req.ProjectID}, nil
}

func (m *mockAPI) CompleteTask(id int) error {
	return m.completeTaskErr
}

func (m *mockAPI) GetTaskLists(projectID int) ([]api.TaskList, error) {
	return []api.TaskList{
		{ID: 1, Name: "To Do", ProjectID: projectID},
	}, nil
}

func (m *mockAPI) GetEntries(opts *api.EntryListOptions) ([]api.TimeEntry, error) {
	m.getEntriesCalls++
	if m.networkErr {
		return nil, errors.New("dial tcp: connection refused")
	}
	return []api.TimeEntry{
		{ID: 1, TaskID: 100, Duration: 3600},
	}, nil
}

func (m *mockAPI) GetEntry(id int) (*api.TimeEntry, error) {
	m.getEntryCalls++
	return &api.TimeEntry{ID: id, TaskID: 100, Duration: 3600}, nil
}

func (m *mockAPI) CreateEntry(req *api.CreateTimeEntryRequest) (*api.TimeEntry, error) {
	if m.createEntryErr != nil {
		return nil, m.createEntryErr
	}
	return &api.TimeEntry{ID: 99, TaskID: req.TaskID}, nil
}

func (m *mockAPI) UpdateEntry(id int, req *api.UpdateTimeEntryRequest) (*api.TimeEntry, error) {
	return &api.TimeEntry{ID: id, TaskID: 100, Duration: 7200}, nil
}

func (m *mockAPI) DeleteEntry(id int) error { return nil }

func (m *mockAPI) GetTodayEntries(userID int) ([]api.TimeEntry, error) {
	return []api.TimeEntry{{ID: 1, TaskID: 100}}, nil
}

func (m *mockAPI) GetActiveEntry(userID int) (*api.TimeEntry, error) {
	return &api.TimeEntry{ID: 1, TaskID: 100}, nil
}

func (m *mockAPI) StartEntry(taskID int, description string) (*api.TimeEntry, error) {
	return &api.TimeEntry{ID: 99, TaskID: taskID, Description: description}, nil
}

func (m *mockAPI) StopEntry(id int) (*api.TimeEntry, error) {
	return &api.TimeEntry{ID: id, Duration: 3600}, nil
}

// --- Test helpers ---

func newTestCachedClient(t *testing.T) (*CachedClient, *mockAPI) {
	t.Helper()
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "cache.json"))
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	mock := &mockAPI{}
	return NewCachedClient(mock, store), mock
}

// --- Tests ---

func TestCachedClient_GetMe_CachesResult(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	user1, err := cc.GetMe()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user1.ID != 1 {
		t.Errorf("expected ID 1, got %d", user1.ID)
	}

	// Second call should hit cache
	user2, err := cc.GetMe()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user2.ID != 1 {
		t.Errorf("expected ID 1, got %d", user2.ID)
	}

	if mock.getMeCalls != 1 {
		t.Errorf("expected 1 API call, got %d", mock.getMeCalls)
	}
}

func TestCachedClient_GetProjects_CachesResult(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	projects, err := cc.GetProjects(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}

	// Second call should hit cache
	projects2, err := cc.GetProjects(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects2) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects2))
	}

	if mock.getProjectsCalls != 1 {
		t.Errorf("expected 1 API call, got %d", mock.getProjectsCalls)
	}
}

func TestCachedClient_GetProject_CachesResult(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	p1, _ := cc.GetProject(42)
	if p1.ID != 42 {
		t.Errorf("expected ID 42, got %d", p1.ID)
	}

	p2, _ := cc.GetProject(42)
	if p2.ID != 42 {
		t.Errorf("expected ID 42, got %d", p2.ID)
	}

	if mock.getProjectCalls != 1 {
		t.Errorf("expected 1 API call, got %d", mock.getProjectCalls)
	}
}

func TestCachedClient_CreateProject_InvalidatesCache(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	// Populate cache
	cc.GetProjects(nil)
	if mock.getProjectsCalls != 1 {
		t.Fatalf("expected 1 call, got %d", mock.getProjectsCalls)
	}

	// Create — should invalidate projects list
	_, err := cc.CreateProject(&api.CreateProjectRequest{Name: "New"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Next GetProjects should miss cache
	cc.GetProjects(nil)
	if mock.getProjectsCalls != 2 {
		t.Errorf("expected 2 API calls after invalidation, got %d", mock.getProjectsCalls)
	}
}

func TestCachedClient_ArchiveProject_InvalidatesCache(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	// Populate cache
	cc.GetProject(1)
	cc.GetProjects(nil)

	err := cc.ArchiveProject(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both project and projects should be invalidated
	cc.GetProject(1)
	cc.GetProjects(nil)
	if mock.getProjectCalls != 2 {
		t.Errorf("expected 2 GetProject calls, got %d", mock.getProjectCalls)
	}
	if mock.getProjectsCalls != 2 {
		t.Errorf("expected 2 GetProjects calls, got %d", mock.getProjectsCalls)
	}
}

func TestCachedClient_GetTasks_CachesResult(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	tasks, _ := cc.GetTasks(&api.TaskListOptions{ProjectID: 10})
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}

	// Second call should hit cache
	cc.GetTasks(&api.TaskListOptions{ProjectID: 10})
	if mock.getTasksCalls != 1 {
		t.Errorf("expected 1 API call, got %d", mock.getTasksCalls)
	}
}

func TestCachedClient_GetTask_CachesResult(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	cc.GetTask(42)
	cc.GetTask(42)

	if mock.getTaskCalls != 1 {
		t.Errorf("expected 1 API call, got %d", mock.getTaskCalls)
	}
}

func TestCachedClient_CreateTask_InvalidatesCache(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	cc.GetTasks(nil)

	cc.CreateTask(&api.CreateTaskRequest{Name: "New Task", ProjectID: 10})

	cc.GetTasks(nil)
	if mock.getTasksCalls != 2 {
		t.Errorf("expected 2 API calls after invalidation, got %d", mock.getTasksCalls)
	}
}

func TestCachedClient_CompleteTask_InvalidatesCache(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	cc.GetTask(42)
	cc.GetTasks(nil)

	cc.CompleteTask(42)

	cc.GetTask(42)
	cc.GetTasks(nil)
	if mock.getTaskCalls != 2 {
		t.Errorf("expected 2 GetTask calls, got %d", mock.getTaskCalls)
	}
	if mock.getTasksCalls != 2 {
		t.Errorf("expected 2 GetTasks calls, got %d", mock.getTasksCalls)
	}
}

func TestCachedClient_GetEntries_CachesResult(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	cc.GetEntries(nil)
	cc.GetEntries(nil)

	if mock.getEntriesCalls != 1 {
		t.Errorf("expected 1 API call, got %d", mock.getEntriesCalls)
	}
}

func TestCachedClient_GetEntry_CachesResult(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	cc.GetEntry(1)
	cc.GetEntry(1)

	if mock.getEntryCalls != 1 {
		t.Errorf("expected 1 API call, got %d", mock.getEntryCalls)
	}
}

func TestCachedClient_CreateEntry_InvalidatesCache(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	cc.GetEntries(nil)

	cc.CreateEntry(&api.CreateTimeEntryRequest{TaskID: 100})

	cc.GetEntries(nil)
	if mock.getEntriesCalls != 2 {
		t.Errorf("expected 2 API calls, got %d", mock.getEntriesCalls)
	}
}

func TestCachedClient_UpdateEntry_InvalidatesCache(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	cc.GetEntry(1)
	cc.GetEntries(nil)

	cc.UpdateEntry(1, &api.UpdateTimeEntryRequest{})

	cc.GetEntry(1)
	cc.GetEntries(nil)
	if mock.getEntryCalls != 2 {
		t.Errorf("expected 2 GetEntry calls, got %d", mock.getEntryCalls)
	}
	if mock.getEntriesCalls != 2 {
		t.Errorf("expected 2 GetEntries calls, got %d", mock.getEntriesCalls)
	}
}

func TestCachedClient_DeleteEntry_InvalidatesCache(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	cc.GetEntries(nil)

	cc.DeleteEntry(1)

	cc.GetEntries(nil)
	if mock.getEntriesCalls != 2 {
		t.Errorf("expected 2 API calls, got %d", mock.getEntriesCalls)
	}
}

func TestCachedClient_StartEntry_InvalidatesCache(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	cc.GetEntries(nil)

	cc.StartEntry(100, "Working")

	cc.GetEntries(nil)
	if mock.getEntriesCalls != 2 {
		t.Errorf("expected 2 API calls, got %d", mock.getEntriesCalls)
	}
}

func TestCachedClient_StopEntry_InvalidatesCache(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	cc.GetEntries(nil)

	cc.StopEntry(1)

	cc.GetEntries(nil)
	if mock.getEntriesCalls != 2 {
		t.Errorf("expected 2 API calls, got %d", mock.getEntriesCalls)
	}
}

func TestCachedClient_GetActiveEntry_NeverCached(t *testing.T) {
	cc, _ := newTestCachedClient(t)

	entry, err := cc.GetActiveEntry(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.ID != 1 {
		t.Errorf("expected ID 1, got %d", entry.ID)
	}
}

func TestCachedClient_GetTodayEntries_PassesThrough(t *testing.T) {
	cc, _ := newTestCachedClient(t)

	entries, err := cc.GetTodayEntries(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestCachedClient_ValidateAuth_PassesThrough(t *testing.T) {
	cc, _ := newTestCachedClient(t)

	err := cc.ValidateAuth()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCachedClient_GetProjectByName_CacheLookup(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	// First, cache a project by ID
	cc.GetProject(1)

	// Now lookup by name — if cache has name index, it should resolve to GetProject(1)
	// which is already cached. The mock GetProjectByName always returns the project.
	project, err := cc.GetProjectByName("Project 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.ID != 1 {
		t.Errorf("expected ID 1, got %d", project.ID)
	}

	// GetProject was called once (for the initial fetch), name lookup fell through
	// to inner.GetProjectByName which returned ID=1
	_ = mock
}

func TestCachedClient_GetTaskByName(t *testing.T) {
	cc, _ := newTestCachedClient(t)

	task, err := cc.GetTaskByName(10, "Task One")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Name != "Task One" {
		t.Errorf("expected name 'Task One', got '%s'", task.Name)
	}
}

func TestCachedClient_GetTaskLists_CachesResult(t *testing.T) {
	cc, _ := newTestCachedClient(t)

	lists1, _ := cc.GetTaskLists(10)
	if len(lists1) != 1 {
		t.Errorf("expected 1 task list, got %d", len(lists1))
	}

	// Second call should use cache (mock always returns same data so we can't
	// distinguish easily, but we verify no error)
	lists2, _ := cc.GetTaskLists(10)
	if len(lists2) != 1 {
		t.Errorf("expected 1 task list, got %d", len(lists2))
	}
}

func TestCachedClient_NetworkError_FallsBackToStale(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	// First, populate cache
	projects, _ := cc.GetProjects(nil)
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}

	// Now simulate network failure
	mock.networkErr = true

	// Should fall back to stale cache
	stale, err := cc.GetProjects(nil)
	if err != nil {
		t.Fatalf("expected stale fallback, got error: %v", err)
	}
	if len(stale) != 2 {
		t.Errorf("expected 2 stale projects, got %d", len(stale))
	}
}

func TestCachedClient_NetworkError_GetProject_FallsBack(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	cc.GetProject(1)

	mock.networkErr = true

	p, err := cc.GetProject(1)
	if err != nil {
		t.Fatalf("expected stale fallback, got error: %v", err)
	}
	if p.ID != 1 {
		t.Errorf("expected ID 1, got %d", p.ID)
	}
}

func TestCachedClient_NetworkError_GetTasks_FallsBack(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	cc.GetTasks(nil)

	mock.networkErr = true

	tasks, err := cc.GetTasks(nil)
	if err != nil {
		t.Fatalf("expected stale fallback, got error: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 stale tasks, got %d", len(tasks))
	}
}

func TestCachedClient_NetworkError_GetEntries_FallsBack(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	cc.GetEntries(nil)

	mock.networkErr = true

	entries, err := cc.GetEntries(nil)
	if err != nil {
		t.Fatalf("expected stale fallback, got error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 stale entry, got %d", len(entries))
	}
}

func TestIsNetworkError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"connection refused", errors.New("dial tcp 127.0.0.1:443: connection refused"), true},
		{"no such host", errors.New("dial tcp: lookup example.com: no such host"), true},
		{"network unreachable", errors.New("network is unreachable"), true},
		{"timeout", errors.New("i/o timeout"), true},
		{"dial tcp", errors.New("dial tcp error"), true},
		{"API error", &api.APIError{StatusCode: 500, Message: "server error"}, false},
		{"generic error", errors.New("something went wrong"), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isNetworkError(tc.err)
			if got != tc.expected {
				t.Errorf("isNetworkError(%q) = %v, want %v", tc.err, got, tc.expected)
			}
		})
	}
}

func TestCachedClient_DifferentKeysAreSeparate(t *testing.T) {
	cc, mock := newTestCachedClient(t)

	// Two different project list queries use different cache keys
	cc.GetProjects(nil)
	cc.GetProjects(&api.ProjectListOptions{ActiveOnly: true})

	if mock.getProjectsCalls != 2 {
		t.Errorf("expected 2 API calls for different keys, got %d", mock.getProjectsCalls)
	}
}
