package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	auth := &APIKeyAuth{APIKey: "test-key"}
	client := NewClient(auth)

	if client.BaseURL != DefaultBaseURL {
		t.Errorf("expected base URL %s, got %s", DefaultBaseURL, client.BaseURL)
	}

	if client.Auth == nil {
		t.Error("expected auth to be set")
	}
}

func TestNewClientWithBaseURL(t *testing.T) {
	auth := &APIKeyAuth{APIKey: "test-key"}
	customURL := "https://custom.paymo.com/api"
	client := NewClientWithBaseURL(customURL, auth)

	if client.BaseURL != customURL {
		t.Errorf("expected base URL %s, got %s", customURL, client.BaseURL)
	}
}

func TestAPIKeyAuth_SetAuth(t *testing.T) {
	auth := &APIKeyAuth{APIKey: "my-api-key"}
	req, _ := http.NewRequest("GET", "https://example.com", nil)

	err := auth.SetAuth(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	username, password, ok := req.BasicAuth()
	if !ok {
		t.Fatal("expected basic auth to be set")
	}

	if username != "my-api-key" {
		t.Errorf("expected username 'my-api-key', got '%s'", username)
	}

	if password != "x" {
		t.Errorf("expected password 'x', got '%s'", password)
	}
}

func TestBasicAuth_SetAuth(t *testing.T) {
	auth := &BasicAuth{Email: "user@example.com", Password: "secret"}
	req, _ := http.NewRequest("GET", "https://example.com", nil)

	err := auth.SetAuth(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	username, password, ok := req.BasicAuth()
	if !ok {
		t.Fatal("expected basic auth to be set")
	}

	if username != "user@example.com" {
		t.Errorf("expected username 'user@example.com', got '%s'", username)
	}

	if password != "secret" {
		t.Errorf("expected password 'secret', got '%s'", password)
	}
}

func TestClient_Get(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		// Verify auth header
		username, _, ok := r.BasicAuth()
		if !ok || username != "test-key" {
			t.Error("expected basic auth with test-key")
		}

		// Verify accept header
		if r.Header.Get("Accept") != "application/json" {
			t.Error("expected Accept: application/json header")
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	auth := &APIKeyAuth{APIKey: "test-key"}
	client := NewClientWithBaseURL(server.URL, auth)

	var result map[string]string
	err := client.Get("test", &result)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", result["status"])
	}
}

func TestClient_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid credentials"})
	}))
	defer server.Close()

	auth := &APIKeyAuth{APIKey: "bad-key"}
	client := NewClientWithBaseURL(server.URL, auth)

	var result map[string]string
	err := client.Get("test", &result)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != 401 {
		t.Errorf("expected status code 401, got %d", apiErr.StatusCode)
	}

	if apiErr.Message != "Invalid credentials" {
		t.Errorf("expected message 'Invalid credentials', got '%s'", apiErr.Message)
	}
}

func TestClient_GetMe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me" {
			t.Errorf("expected path /me, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(MeResponse{
			Users: []User{
				{ID: 123, Name: "Test User", Email: "test@example.com"},
			},
		})
	}))
	defer server.Close()

	auth := &APIKeyAuth{APIKey: "test-key"}
	client := NewClientWithBaseURL(server.URL, auth)

	user, err := client.GetMe()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID != 123 {
		t.Errorf("expected user ID 123, got %d", user.ID)
	}

	if user.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", user.Name)
	}

	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", user.Email)
	}
}

func TestClient_RateLimitHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Ratelimit-Limit", "100")
		w.Header().Set("X-Ratelimit-Remaining", "99")
		w.Header().Set("X-Ratelimit-Decay-Period", "60")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	auth := &APIKeyAuth{APIKey: "test-key"}
	client := NewClientWithBaseURL(server.URL, auth)

	var result map[string]string
	err := client.Get("test", &result)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.rateLimit != 100 {
		t.Errorf("expected rate limit 100, got %d", client.rateLimit)
	}

	if client.rateRemaining != 99 {
		t.Errorf("expected rate remaining 99, got %d", client.rateRemaining)
	}
}