package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_GetProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ProjectsResponse{
			Projects: []Project{
				{ID: 1, Name: "Project One", Active: true},
				{ID: 2, Name: "Project Two", Active: false},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	projects, err := client.GetProjects(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
}

func TestClient_GetProjects_ActiveOnly(t *testing.T) {
	var capturedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ProjectsResponse{
			Projects: []Project{},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	_, err := client.GetProjects(&ProjectListOptions{ActiveOnly: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// URL-encoded: active=true becomes active%3Dtrue
	if !strings.Contains(capturedQuery, "active") {
		t.Errorf("expected active filter, got: %s", capturedQuery)
	}
}

func TestClient_GetProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/projects/123") {
			t.Errorf("expected path /projects/123, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ProjectResponse{
			Projects: []Project{
				{ID: 123, Name: "Test Project", Active: true},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	project, err := client.GetProject(123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if project.ID != 123 {
		t.Errorf("expected ID 123, got %d", project.ID)
	}

	if project.Name != "Test Project" {
		t.Errorf("expected name 'Test Project', got '%s'", project.Name)
	}
}

func TestClient_GetProjectByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("where")
		
		if !strings.Contains(query, "name like") {
			t.Errorf("expected name filter, got: %s", query)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ProjectsResponse{
			Projects: []Project{
				{ID: 1, Name: "My Test Project", Active: true},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	project, err := client.GetProjectByName("Test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if project.Name != "My Test Project" {
		t.Errorf("expected name 'My Test Project', got '%s'", project.Name)
	}
}

func TestClient_GetProjectByName_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ProjectsResponse{
			Projects: []Project{},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	_, err := client.GetProjectByName("NonExistent")
	if err == nil {
		t.Error("expected error for non-existent project")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Errorf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestClient_CreateProject(t *testing.T) {
	var receivedBody CreateProjectRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		json.NewDecoder(r.Body).Decode(&receivedBody)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ProjectResponse{
			Projects: []Project{
				{ID: 999, Name: receivedBody.Name, Active: true, Billable: receivedBody.Billable},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	req := &CreateProjectRequest{
		Name:     "New Project",
		Billable: true,
	}

	project, err := client.CreateProject(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if project.ID != 999 {
		t.Errorf("expected ID 999, got %d", project.ID)
	}

	if project.Name != "New Project" {
		t.Errorf("expected name 'New Project', got '%s'", project.Name)
	}

	if !project.Billable {
		t.Error("expected billable to be true")
	}
}

func TestClient_ArchiveProject(t *testing.T) {
	var receivedActive bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var body struct {
			Active bool `json:"active"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		receivedActive = body.Active

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	err := client.ArchiveProject(123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedActive != false {
		t.Error("expected active=false in request body")
	}
}