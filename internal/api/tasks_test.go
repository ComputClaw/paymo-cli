package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_GetTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TasksResponse{
			Tasks: []Task{
				{ID: 1, Name: "Task One", ProjectID: 10},
				{ID: 2, Name: "Task Two", ProjectID: 10},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	tasks, err := client.GetTasks(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestClient_GetTasks_WithProjectFilter(t *testing.T) {
	var capturedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TasksResponse{
			Tasks: []Task{},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	_, err := client.GetTasks(&TaskListOptions{ProjectID: 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedQuery, "project_id") {
		t.Errorf("expected project_id filter, got: %s", capturedQuery)
	}
}

func TestClient_GetTasks_WithCompletedFilter(t *testing.T) {
	var capturedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TasksResponse{Tasks: []Task{}})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	// Default: IncludeCompleted=false should add complete=false
	_, err := client.GetTasks(&TaskListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedQuery, "complete") {
		t.Errorf("expected complete filter, got: %s", capturedQuery)
	}
}

func TestClient_GetTasks_WithIncludeProject(t *testing.T) {
	var capturedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TasksResponse{Tasks: []Task{}})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	_, err := client.GetTasks(&TaskListOptions{IncludeProject: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedQuery, "include=project") {
		t.Errorf("expected include=project, got: %s", capturedQuery)
	}
}

func TestClient_GetTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/tasks/42") {
			t.Errorf("expected path /tasks/42, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TasksResponse{
			Tasks: []Task{
				{ID: 42, Name: "Design Phase", ProjectID: 10},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	task, err := client.GetTask(42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task.ID != 42 {
		t.Errorf("expected ID 42, got %d", task.ID)
	}
	if task.Name != "Design Phase" {
		t.Errorf("expected name 'Design Phase', got '%s'", task.Name)
	}
}

func TestClient_GetTask_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TasksResponse{Tasks: []Task{}})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	_, err := client.GetTask(999)
	if err == nil {
		t.Fatal("expected error for non-existent task")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestClient_GetTaskByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("where")

		if !strings.Contains(query, "name like") {
			t.Errorf("expected name filter, got: %s", query)
		}
		if !strings.Contains(query, "project_id=10") {
			t.Errorf("expected project_id filter, got: %s", query)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TasksResponse{
			Tasks: []Task{
				{ID: 5, Name: "Design Homepage", ProjectID: 10},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	task, err := client.GetTaskByName(10, "Design")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task.Name != "Design Homepage" {
		t.Errorf("expected name 'Design Homepage', got '%s'", task.Name)
	}
}

func TestClient_GetTaskByName_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TasksResponse{Tasks: []Task{}})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	_, err := client.GetTaskByName(10, "NonExistent")
	if err == nil {
		t.Fatal("expected error for non-existent task")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestClient_GetTaskByName_Sanitization(t *testing.T) {
	var capturedWhere string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedWhere = r.URL.Query().Get("where")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TasksResponse{
			Tasks: []Task{{ID: 1, Name: "safe", ProjectID: 1}},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	// Input with quotes/backslashes that should be stripped
	client.GetTaskByName(1, `test"inject\`)

	if strings.Contains(capturedWhere, `"inject`) {
		t.Errorf("double quotes should be sanitized from name, got: %s", capturedWhere)
	}
	if strings.Contains(capturedWhere, `\`) {
		t.Errorf("backslashes should be sanitized from name, got: %s", capturedWhere)
	}
}

func TestClient_CreateTask(t *testing.T) {
	var receivedBody CreateTaskRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		json.NewDecoder(r.Body).Decode(&receivedBody)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TasksResponse{
			Tasks: []Task{
				{ID: 100, Name: receivedBody.Name, ProjectID: receivedBody.ProjectID},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	req := &CreateTaskRequest{
		Name:      "New Task",
		ProjectID: 10,
		Billable:  true,
	}

	task, err := client.CreateTask(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task.ID != 100 {
		t.Errorf("expected ID 100, got %d", task.ID)
	}
	if task.Name != "New Task" {
		t.Errorf("expected name 'New Task', got '%s'", task.Name)
	}
	if receivedBody.ProjectID != 10 {
		t.Errorf("expected project ID 10 in request, got %d", receivedBody.ProjectID)
	}
}

func TestClient_CompleteTask(t *testing.T) {
	var receivedComplete bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/tasks/42") {
			t.Errorf("expected path /tasks/42, got %s", r.URL.Path)
		}

		var body struct {
			Complete bool `json:"complete"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		receivedComplete = body.Complete

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	err := client.CompleteTask(42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !receivedComplete {
		t.Error("expected complete=true in request body")
	}
}

func TestClient_GetTaskLists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("where")
		if !strings.Contains(query, "project_id=10") {
			t.Errorf("expected project_id filter, got: %s", query)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			TaskLists []TaskList `json:"tasklists"`
		}{
			TaskLists: []TaskList{
				{ID: 1, Name: "To Do", ProjectID: 10},
				{ID: 2, Name: "In Progress", ProjectID: 10},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	lists, err := client.GetTaskLists(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(lists) != 2 {
		t.Errorf("expected 2 task lists, got %d", len(lists))
	}
	if lists[0].Name != "To Do" {
		t.Errorf("expected name 'To Do', got '%s'", lists[0].Name)
	}
}
