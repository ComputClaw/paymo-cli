package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_GetEntries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		if !strings.HasPrefix(r.URL.Path, "/entries") {
			t.Errorf("expected path /entries, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TimeEntriesResponse{
			Entries: []TimeEntry{
				{
					ID:          1,
					TaskID:      100,
					UserID:      10,
					StartTime:   time.Now(),
					Duration:    3600,
					Description: "Test entry",
				},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	entries, err := client.GetEntries(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].ID != 1 {
		t.Errorf("expected ID 1, got %d", entries[0].ID)
	}
}

func TestClient_GetEntries_WithFilters(t *testing.T) {
	var capturedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TimeEntriesResponse{
			Entries: []TimeEntry{},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	opts := &EntryListOptions{
		UserID:      123,
		IncludeTask: true,
	}

	_, err := client.GetEntries(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// URL-encoded: user_id=123 becomes user_id%3D123
	if !strings.Contains(capturedQuery, "user_id") {
		t.Errorf("expected user_id filter in query, got: %s", capturedQuery)
	}

	if !strings.Contains(capturedQuery, "include=task") {
		t.Errorf("expected include=task in query, got: %s", capturedQuery)
	}
}

func TestClient_CreateEntry(t *testing.T) {
	var receivedBody CreateTimeEntryRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		json.NewDecoder(r.Body).Decode(&receivedBody)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TimeEntryResponse{
			Entries: []TimeEntry{
				{
					ID:          999,
					TaskID:      receivedBody.TaskID,
					Description: receivedBody.Description,
					StartTime:   time.Now(),
				},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	req := &CreateTimeEntryRequest{
		TaskID:      100,
		StartTime:   time.Now().Format("2006-01-02T15:04:05Z"),
		Description: "New entry",
	}

	entry, err := client.CreateEntry(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.ID != 999 {
		t.Errorf("expected ID 999, got %d", entry.ID)
	}

	if receivedBody.TaskID != 100 {
		t.Errorf("expected task ID 100 in request, got %d", receivedBody.TaskID)
	}

	if receivedBody.Description != "New entry" {
		t.Errorf("expected description 'New entry', got '%s'", receivedBody.Description)
	}
}

func TestClient_StartEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateTimeEntryRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Verify start time is set
		if req.StartTime == "" {
			t.Error("expected start time to be set")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TimeEntryResponse{
			Entries: []TimeEntry{
				{
					ID:          1,
					TaskID:      req.TaskID,
					Description: req.Description,
					StartTime:   time.Now(),
				},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	entry, err := client.StartEntry(100, "Starting work")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.TaskID != 100 {
		t.Errorf("expected task ID 100, got %d", entry.TaskID)
	}

	if entry.Description != "Starting work" {
		t.Errorf("expected description 'Starting work', got '%s'", entry.Description)
	}
}

func TestClient_StopEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var req UpdateTimeEntryRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Verify end time is set
		if req.EndTime == nil {
			t.Error("expected end time to be set")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TimeEntryResponse{
			Entries: []TimeEntry{
				{
					ID:        1,
					StartTime: time.Now().Add(-1 * time.Hour),
					EndTime:   time.Now(),
					Duration:  3600,
				},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	entry, err := client.StopEntry(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.Duration != 3600 {
		t.Errorf("expected duration 3600, got %d", entry.Duration)
	}
}

func TestClient_GetActiveEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("where")
		
		if !strings.Contains(query, "end_time=\"\"") {
			t.Errorf("expected end_time filter, got: %s", query)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TimeEntriesResponse{
			Entries: []TimeEntry{
				{
					ID:        1,
					TaskID:    100,
					StartTime: time.Now(),
				},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	entry, err := client.GetActiveEntry(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry == nil {
		t.Fatal("expected entry, got nil")
	}

	if entry.ID != 1 {
		t.Errorf("expected ID 1, got %d", entry.ID)
	}
}

func TestClient_GetActiveEntry_NoActive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TimeEntriesResponse{
			Entries: []TimeEntry{},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL, &APIKeyAuth{APIKey: "test"})

	entry, err := client.GetActiveEntry(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry != nil {
		t.Error("expected nil entry when no active timer")
	}
}