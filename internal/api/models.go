package api

import (
	"time"
)

// User represents a Paymo user
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Type      string    `json:"type"`
	Active    bool      `json:"active"`
	Timezone  string    `json:"timezone"`
	CreatedOn time.Time `json:"created_on"`
	UpdatedOn time.Time `json:"updated_on"`
}

// MeResponse is the response from /api/me
type MeResponse struct {
	Users []User `json:"users"`
}

// Project represents a Paymo project
type Project struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Code         string    `json:"code,omitempty"`
	Description  string    `json:"description,omitempty"`
	ClientID     int       `json:"client_id,omitempty"`
	Active       bool      `json:"active"`
	Billable     bool      `json:"billable"`
	BudgetHours  float64   `json:"budget_hours,omitempty"`
	PricePerHour float64   `json:"price_per_hour,omitempty"`
	Color        string    `json:"color,omitempty"`
	Users        []int     `json:"users,omitempty"`
	Managers     []int     `json:"managers,omitempty"`
	CreatedOn    time.Time `json:"created_on"`
	UpdatedOn    time.Time `json:"updated_on"`
}

// ProjectsResponse is the response from /api/projects
type ProjectsResponse struct {
	Projects []Project `json:"projects"`
}

// ProjectResponse is the response for a single project
type ProjectResponse struct {
	Projects []Project `json:"projects"`
}

// PaymoClient represents a Paymo client (customer)
type PaymoClient struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Address   string    `json:"address,omitempty"`
	City      string    `json:"city,omitempty"`
	Country   string    `json:"country,omitempty"`
	Active    bool      `json:"active"`
	CreatedOn time.Time `json:"created_on"`
	UpdatedOn time.Time `json:"updated_on"`
}

// ClientsResponse is the response from /api/clients
type ClientsResponse struct {
	Clients []PaymoClient `json:"clients"`
}

// TaskList represents a Paymo task list
type TaskList struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	ProjectID int       `json:"project_id"`
	Seq       int       `json:"seq"`
	CreatedOn time.Time `json:"created_on"`
	UpdatedOn time.Time `json:"updated_on"`
}

// Task represents a Paymo task
type Task struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code,omitempty"`
	ProjectID   int       `json:"project_id"`
	TaskListID  int       `json:"tasklist_id"`
	Description string    `json:"description,omitempty"`
	Complete    bool      `json:"complete"`
	Billable    bool      `json:"billable"`
	DueDate     string    `json:"due_date,omitempty"`
	Users       []int     `json:"users,omitempty"`
	Priority    int       `json:"priority,omitempty"`
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"updated_on"`
}

// TasksResponse is the response from /api/tasks
type TasksResponse struct {
	Tasks []Task `json:"tasks"`
}

// TimeEntry represents a Paymo time entry
type TimeEntry struct {
	ID          int       `json:"id"`
	TaskID      int       `json:"task_id"`
	UserID      int       `json:"user_id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time,omitempty"`
	Duration    int       `json:"duration"` // in seconds
	Description string    `json:"description,omitempty"`
	Billable    bool      `json:"billable"`
	Billed      bool      `json:"billed"`
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"updated_on"`
	
	// Included relations (when requested)
	Task    *Task    `json:"task,omitempty"`
	Project *Project `json:"project,omitempty"`
}

// TimeEntriesResponse is the response from /api/entries
type TimeEntriesResponse struct {
	Entries []TimeEntry `json:"entries"`
}

// TimeEntryResponse is the response for a single time entry
type TimeEntryResponse struct {
	Entries []TimeEntry `json:"entries"`
}

// CreateTimeEntryRequest is the request body for creating a time entry
type CreateTimeEntryRequest struct {
	TaskID      int    `json:"task_id"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time,omitempty"`
	Duration    int    `json:"duration,omitempty"`
	Description string `json:"description,omitempty"`
}

// UpdateTimeEntryRequest is the request body for updating a time entry
type UpdateTimeEntryRequest struct {
	TaskID      *int    `json:"task_id,omitempty"`
	StartTime   *string `json:"start_time,omitempty"`
	EndTime     *string `json:"end_time,omitempty"`
	Duration    *int    `json:"duration,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Timer represents an active timer (not a Paymo native concept, we track locally)
type Timer struct {
	Active      bool      `json:"active"`
	EntryID     int       `json:"entry_id,omitempty"`
	ProjectID   int       `json:"project_id"`
	TaskID      int       `json:"task_id"`
	ProjectName string    `json:"project_name"`
	TaskName    string    `json:"task_name"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
}

// CreateProjectRequest is the request body for creating a project
type CreateProjectRequest struct {
	Name         string  `json:"name"`
	ClientID     *int    `json:"client_id,omitempty"`
	Description  string  `json:"description,omitempty"`
	Billable     bool    `json:"billable"`
	BudgetHours  float64 `json:"budget_hours,omitempty"`
	PricePerHour float64 `json:"price_per_hour,omitempty"`
}

// CreateTaskRequest is the request body for creating a task
type CreateTaskRequest struct {
	Name        string `json:"name"`
	ProjectID   int    `json:"project_id"`
	TaskListID  int    `json:"tasklist_id,omitempty"`
	Description string `json:"description,omitempty"`
	Billable    bool   `json:"billable"`
	DueDate     string `json:"due_date,omitempty"`
}