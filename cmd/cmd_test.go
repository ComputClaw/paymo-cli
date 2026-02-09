package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ComputClaw/paymo-cli/internal/api"
)

// mockPaymoAPI implements api.PaymoAPI for cmd/ testing.
type mockPaymoAPI struct {
	projects     []api.Project
	tasks        []api.Task
	entries      []api.TimeEntry
	tasklists    []api.TaskList
	user         *api.User
	activeEntry  *api.TimeEntry
	createErr    error
	archiveErr   error
	completeErr  error
	deleteErr    error
}

func newMockAPI() *mockPaymoAPI {
	return &mockPaymoAPI{
		user: &api.User{ID: 1, Name: "Test User", Email: "test@example.com"},
		projects: []api.Project{
			{ID: 1, Name: "Project Alpha", Active: true, Billable: true},
			{ID: 2, Name: "Project Beta", Active: true, Billable: false},
		},
		tasks: []api.Task{
			{ID: 10, Name: "Design", ProjectID: 1, Complete: false},
			{ID: 11, Name: "Development", ProjectID: 1, Complete: false},
		},
		entries: []api.TimeEntry{
			{ID: 100, TaskID: 10, UserID: 1, Duration: 3600, Description: "Working on design"},
		},
		tasklists: []api.TaskList{
			{ID: 1, Name: "To Do", ProjectID: 1},
		},
	}
}

func (m *mockPaymoAPI) GetMe() (*api.User, error) { return m.user, nil }
func (m *mockPaymoAPI) ValidateAuth() error        { return nil }

func (m *mockPaymoAPI) GetProjects(opts *api.ProjectListOptions) ([]api.Project, error) {
	return m.projects, nil
}

func (m *mockPaymoAPI) GetProject(id int) (*api.Project, error) {
	for _, p := range m.projects {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, &api.APIError{StatusCode: 404, Code: "NOT_FOUND", Message: "project not found"}
}

func (m *mockPaymoAPI) GetProjectByName(name string) (*api.Project, error) {
	nameLower := strings.ToLower(name)
	for _, p := range m.projects {
		if strings.Contains(strings.ToLower(p.Name), nameLower) {
			return &p, nil
		}
	}
	return nil, &api.APIError{StatusCode: 404, Code: "NOT_FOUND", Message: "project not found"}
}

func (m *mockPaymoAPI) CreateProject(req *api.CreateProjectRequest) (*api.Project, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return &api.Project{ID: 99, Name: req.Name, Active: true, Billable: req.Billable}, nil
}

func (m *mockPaymoAPI) ArchiveProject(id int) error {
	return m.archiveErr
}

func (m *mockPaymoAPI) GetTasks(opts *api.TaskListOptions) ([]api.Task, error) {
	return m.tasks, nil
}

func (m *mockPaymoAPI) GetTask(id int) (*api.Task, error) {
	for _, t := range m.tasks {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, &api.APIError{StatusCode: 404, Code: "NOT_FOUND", Message: "task not found"}
}

func (m *mockPaymoAPI) GetTaskByName(projectID int, name string) (*api.Task, error) {
	nameLower := strings.ToLower(name)
	for _, t := range m.tasks {
		if t.ProjectID == projectID && strings.Contains(strings.ToLower(t.Name), nameLower) {
			return &t, nil
		}
	}
	return nil, &api.APIError{StatusCode: 404, Code: "NOT_FOUND", Message: "task not found"}
}

func (m *mockPaymoAPI) CreateTask(req *api.CreateTaskRequest) (*api.Task, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return &api.Task{ID: 99, Name: req.Name, ProjectID: req.ProjectID}, nil
}

func (m *mockPaymoAPI) CompleteTask(id int) error {
	return m.completeErr
}

func (m *mockPaymoAPI) GetTaskLists(projectID int) ([]api.TaskList, error) {
	return m.tasklists, nil
}

func (m *mockPaymoAPI) GetEntries(opts *api.EntryListOptions) ([]api.TimeEntry, error) {
	return m.entries, nil
}

func (m *mockPaymoAPI) GetEntry(id int) (*api.TimeEntry, error) {
	for _, e := range m.entries {
		if e.ID == id {
			return &e, nil
		}
	}
	return nil, &api.APIError{StatusCode: 404, Code: "NOT_FOUND", Message: "entry not found"}
}

func (m *mockPaymoAPI) CreateEntry(req *api.CreateTimeEntryRequest) (*api.TimeEntry, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return &api.TimeEntry{ID: 99, TaskID: req.TaskID, Description: req.Description, StartTime: time.Now()}, nil
}

func (m *mockPaymoAPI) UpdateEntry(id int, req *api.UpdateTimeEntryRequest) (*api.TimeEntry, error) {
	return &api.TimeEntry{ID: id, TaskID: 10, Duration: 7200}, nil
}

func (m *mockPaymoAPI) DeleteEntry(id int) error {
	return m.deleteErr
}

func (m *mockPaymoAPI) GetTodayEntries(userID int) ([]api.TimeEntry, error) {
	return m.entries, nil
}

func (m *mockPaymoAPI) GetActiveEntry(userID int) (*api.TimeEntry, error) {
	return m.activeEntry, nil
}

func (m *mockPaymoAPI) StartEntry(taskID int, description string) (*api.TimeEntry, error) {
	return &api.TimeEntry{ID: 99, TaskID: taskID, Description: description, StartTime: time.Now()}, nil
}

func (m *mockPaymoAPI) StopEntry(id int) (*api.TimeEntry, error) {
	return &api.TimeEntry{ID: id, Duration: 3600, EndTime: time.Now()}, nil
}

// --- Test helpers ---

// runCommand runs a command with mock API and returns error only.
// Output goes to os.Stdout (formatter uses os.Stdout); we test behavior not output.
// Output formatting is already tested in internal/output/output_test.go.
func runCommand(mock api.PaymoAPI, args ...string) error {
	origClient := getAPIClient
	defer func() { getAPIClient = origClient }()

	getAPIClient = func() (api.PaymoAPI, error) {
		return mock, nil
	}

	// Reset persistent flag state to avoid bleeding between tests.
	// Cobra doesn't reset flag values between Execute() calls.
	resetCommandFlags(showTaskCmd, "project")
	resetCommandFlags(completeTaskCmd, "project")
	resetCommandFlags(createTaskCmd, "project")
	resetCommandFlags(listTasksCmd, "project")
	resetCommandFlags(logCmd, "date", "project")

	rootCmd.SetArgs(args)
	viper.Set("format", "json")
	viper.Set("quiet", false)
	viper.Set("no_cache", true)

	return rootCmd.Execute()
}

// resetCommandFlags resets named flags on a command to their default values.
func resetCommandFlags(cmd *cobra.Command, flagNames ...string) {
	for _, name := range flagNames {
		f := cmd.Flags().Lookup(name)
		if f != nil {
			f.Value.Set(f.DefValue)
			f.Changed = false
		}
	}
}

// --- Project command tests ---

func TestProjectsList(t *testing.T) {
	err := runCommand(newMockAPI(), "projects", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectsShow_ByID(t *testing.T) {
	err := runCommand(newMockAPI(), "projects", "show", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectsShow_ByName(t *testing.T) {
	err := runCommand(newMockAPI(), "projects", "show", "Alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectsShow_NotFound(t *testing.T) {
	err := runCommand(newMockAPI(), "projects", "show", "999")
	if err == nil {
		t.Fatal("expected error for non-existent project")
	}
}

func TestProjectsCreate(t *testing.T) {
	err := runCommand(newMockAPI(), "projects", "create", "New Project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectsCreate_Error(t *testing.T) {
	mock := newMockAPI()
	mock.createErr = errors.New("API error")
	err := runCommand(mock, "projects", "create", "Fail")
	if err == nil {
		t.Fatal("expected error from create")
	}
}

func TestProjectsCreate_MissingArgs(t *testing.T) {
	err := runCommand(newMockAPI(), "projects", "create")
	if err == nil {
		t.Fatal("expected error when no name provided")
	}
}

func TestProjectsArchive(t *testing.T) {
	err := runCommand(newMockAPI(), "projects", "archive", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectsArchive_Error(t *testing.T) {
	mock := newMockAPI()
	mock.archiveErr = errors.New("API error")
	err := runCommand(mock, "projects", "archive", "1")
	if err == nil {
		t.Fatal("expected error from archive")
	}
}

func TestProjectsTasks(t *testing.T) {
	err := runCommand(newMockAPI(), "projects", "tasks", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectsTasks_ByName(t *testing.T) {
	err := runCommand(newMockAPI(), "projects", "tasks", "Alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Task command tests ---

func TestTasksList(t *testing.T) {
	err := runCommand(newMockAPI(), "tasks", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTasksList_WithProject(t *testing.T) {
	err := runCommand(newMockAPI(), "tasks", "list", "--project", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTasksList_WithProjectName(t *testing.T) {
	err := runCommand(newMockAPI(), "tasks", "list", "--project", "Alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTasksShow_ByID(t *testing.T) {
	err := runCommand(newMockAPI(), "tasks", "show", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTasksShow_ByName(t *testing.T) {
	err := runCommand(newMockAPI(), "tasks", "show", "Design", "--project", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTasksShow_ByName_NoProject(t *testing.T) {
	err := runCommand(newMockAPI(), "tasks", "show", "Design")
	if err == nil {
		t.Fatal("expected error when name-based lookup without --project")
	}
}

func TestTasksCreate(t *testing.T) {
	err := runCommand(newMockAPI(), "tasks", "create", "New Task", "--project", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTasksCreate_NoProject(t *testing.T) {
	err := runCommand(newMockAPI(), "tasks", "create", "New Task")
	if err == nil {
		t.Fatal("expected error when --project is missing")
	}
}

func TestTasksCreate_MissingArgs(t *testing.T) {
	err := runCommand(newMockAPI(), "tasks", "create")
	if err == nil {
		t.Fatal("expected error when no name provided")
	}
}

func TestTasksComplete(t *testing.T) {
	err := runCommand(newMockAPI(), "tasks", "complete", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTasksComplete_Error(t *testing.T) {
	mock := newMockAPI()
	mock.completeErr = errors.New("API error")
	err := runCommand(mock, "tasks", "complete", "10")
	if err == nil {
		t.Fatal("expected error from complete")
	}
}

func TestTasksComplete_NotFound(t *testing.T) {
	err := runCommand(newMockAPI(), "tasks", "complete", "999")
	if err == nil {
		t.Fatal("expected error for non-existent task")
	}
}

// --- Time command tests ---

func TestTimeLog(t *testing.T) {
	err := runCommand(newMockAPI(), "time", "log")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTimeLog_WithProject(t *testing.T) {
	err := runCommand(newMockAPI(), "time", "log", "--project", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTimeLog_WithDate(t *testing.T) {
	err := runCommand(newMockAPI(), "time", "log", "--date", "yesterday")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTimeLog_WithDateThisWeek(t *testing.T) {
	err := runCommand(newMockAPI(), "time", "log", "--date", "this-week")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTimeLog_WithDateLastWeek(t *testing.T) {
	err := runCommand(newMockAPI(), "time", "log", "--date", "last-week")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTimeLog_WithDateSpecific(t *testing.T) {
	err := runCommand(newMockAPI(), "time", "log", "--date", "2026-01-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTimeLog_InvalidDate(t *testing.T) {
	err := runCommand(newMockAPI(), "time", "log", "--date", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid date format")
	}
}

func TestTimeStatus(t *testing.T) {
	// No timer state file means no active timer
	err := runCommand(newMockAPI(), "time", "status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Resolver tests (unit tests for helpers.go) ---

func TestResolveProjectID_Numeric(t *testing.T) {
	mock := newMockAPI()
	id, err := resolveProjectID(mock, "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("expected 42, got %d", id)
	}
}

func TestResolveProjectID_Name(t *testing.T) {
	mock := newMockAPI()
	id, err := resolveProjectID(mock, "Alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 1 {
		t.Errorf("expected 1, got %d", id)
	}
}

func TestResolveProjectID_NotFound(t *testing.T) {
	mock := newMockAPI()
	_, err := resolveProjectID(mock, "NonExistent")
	if err == nil {
		t.Fatal("expected error for non-existent project")
	}
}

func TestResolveProject_ByID(t *testing.T) {
	mock := newMockAPI()
	p, err := resolveProject(mock, "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "Project Alpha" {
		t.Errorf("expected 'Project Alpha', got '%s'", p.Name)
	}
}

func TestResolveProject_ByName(t *testing.T) {
	mock := newMockAPI()
	p, err := resolveProject(mock, "Beta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "Project Beta" {
		t.Errorf("expected 'Project Beta', got '%s'", p.Name)
	}
}

func TestResolveProject_NotFound_ByID(t *testing.T) {
	mock := newMockAPI()
	_, err := resolveProject(mock, "999")
	if err == nil {
		t.Fatal("expected error for non-existent project ID")
	}
}

func TestResolveTask_ByID(t *testing.T) {
	mock := newMockAPI()
	task, err := resolveTask(mock, "10", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Name != "Design" {
		t.Errorf("expected 'Design', got '%s'", task.Name)
	}
}

func TestResolveTask_ByName(t *testing.T) {
	mock := newMockAPI()
	task, err := resolveTask(mock, "Design", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Name != "Design" {
		t.Errorf("expected 'Design', got '%s'", task.Name)
	}
}

func TestResolveTask_ByName_WithProjectName(t *testing.T) {
	mock := newMockAPI()
	task, err := resolveTask(mock, "Design", "Alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Name != "Design" {
		t.Errorf("expected 'Design', got '%s'", task.Name)
	}
}

func TestResolveTask_ByName_NoProject(t *testing.T) {
	mock := newMockAPI()
	_, err := resolveTask(mock, "Design", "")
	if err == nil {
		t.Fatal("expected error for name-based lookup without project")
	}
	if !strings.Contains(err.Error(), "--project") {
		t.Errorf("expected error about --project flag, got: %v", err)
	}
}

func TestResolveTask_NotFound_ByID(t *testing.T) {
	mock := newMockAPI()
	_, err := resolveTask(mock, "999", "")
	if err == nil {
		t.Fatal("expected error for non-existent task ID")
	}
}

// --- newFormatter test ---

func TestNewFormatter(t *testing.T) {
	viper.Set("format", "json")
	viper.Set("quiet", true)

	f := newFormatter()
	if f.Format != "json" {
		t.Errorf("expected format 'json', got '%s'", f.Format)
	}
	if !f.Quiet {
		t.Error("expected quiet to be true")
	}

	// Reset
	viper.Set("format", "table")
	viper.Set("quiet", false)
}

// --- GetOutputFormat test ---

func TestGetOutputFormat(t *testing.T) {
	viper.Set("format", "csv")
	got := GetOutputFormat()
	if got != "csv" {
		t.Errorf("expected 'csv', got '%s'", got)
	}

	viper.Set("format", "")
	got = GetOutputFormat()
	if got != "table" {
		t.Errorf("expected 'table' as default, got '%s'", got)
	}

	// Reset
	viper.Set("format", "table")
}

// --- Schema command test ---

func TestSchemaCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"schema"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Error cases ---

func TestGetAPIClient_NoAuth(t *testing.T) {
	origClient := getAPIClient
	defer func() { getAPIClient = origClient }()

	getAPIClient = func() (api.PaymoAPI, error) {
		return nil, fmt.Errorf("not authenticated")
	}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"projects", "list"})
	viper.Set("format", "json")
	viper.Set("no_cache", true)

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when not authenticated")
	}
}

func TestProjectsArchive_MissingArgs(t *testing.T) {
	err := runCommand(newMockAPI(), "projects", "archive")
	if err == nil {
		t.Fatal("expected error when no project specified")
	}
}

func TestTasksComplete_MissingArgs(t *testing.T) {
	err := runCommand(newMockAPI(), "tasks", "complete")
	if err == nil {
		t.Fatal("expected error when no task specified")
	}
}

// Verify the mock implements the full interface
var _ api.PaymoAPI = (*mockPaymoAPI)(nil)
