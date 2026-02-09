package api

// PaymoAPI defines the contract for all Paymo API operations.
// Both the raw Client and the cached wrapper implement this interface.
type PaymoAPI interface {
	// Auth
	GetMe() (*User, error)
	ValidateAuth() error

	// Projects
	GetProjects(opts *ProjectListOptions) ([]Project, error)
	GetProject(id int) (*Project, error)
	GetProjectByName(name string) (*Project, error)
	CreateProject(req *CreateProjectRequest) (*Project, error)
	ArchiveProject(id int) error

	// Tasks
	GetTasks(opts *TaskListOptions) ([]Task, error)
	GetTask(id int) (*Task, error)
	GetTaskByName(projectID int, name string) (*Task, error)
	CreateTask(req *CreateTaskRequest) (*Task, error)
	CompleteTask(id int) error
	GetTaskLists(projectID int) ([]TaskList, error)

	// Time Entries
	GetEntries(opts *EntryListOptions) ([]TimeEntry, error)
	GetEntry(id int) (*TimeEntry, error)
	CreateEntry(req *CreateTimeEntryRequest) (*TimeEntry, error)
	UpdateEntry(id int, req *UpdateTimeEntryRequest) (*TimeEntry, error)
	DeleteEntry(id int) error
	GetTodayEntries(userID int) ([]TimeEntry, error)
	GetActiveEntry(userID int) (*TimeEntry, error)
	StartEntry(taskID int, description string) (*TimeEntry, error)
	StopEntry(id int) (*TimeEntry, error)
}

// Compile-time check: *Client implements PaymoAPI
var _ PaymoAPI = (*Client)(nil)
