package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOpen_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer store.Close()

	stats, _ := store.Stats()
	if len(stats) != 0 {
		t.Errorf("expected empty cache, got %d types", len(stats))
	}
}

func TestOpen_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	// Write a valid cache file
	data := `{"entries":{"project":{"1":{"data":"{\"id\":1,\"name\":\"Test\"}","cached_at":` +
		itoa(time.Now().Unix()) + `,"ttl_seconds":3600}}}}`
	os.WriteFile(path, []byte(data), 0644)

	store, err := Open(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer store.Close()

	stats, _ := store.Stats()
	if stats["project"] != 1 {
		t.Errorf("expected 1 project entry, got %d", stats["project"])
	}
}

func TestOpen_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	os.WriteFile(path, []byte("not json"), 0644)

	store, err := Open(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer store.Close()

	// Should start fresh
	stats, _ := store.Stats()
	if len(stats) != 0 {
		t.Errorf("expected empty cache after corrupt file, got %d types", len(stats))
	}
}

func TestSetAndGet(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	type item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	err := store.Set("project", "1", &item{ID: 1, Name: "Test"})
	if err != nil {
		t.Fatalf("Set error: %v", err)
	}

	var got item
	err = store.Get("project", "1", &got)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}

	if got.ID != 1 || got.Name != "Test" {
		t.Errorf("expected {1, Test}, got {%d, %s}", got.ID, got.Name)
	}
}

func TestGet_Miss(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	var got struct{}
	err := store.Get("project", "nonexistent", &got)
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss, got %v", err)
	}
}

func TestGet_MissBucket(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	var got struct{}
	err := store.Get("nonexistent_type", "1", &got)
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss, got %v", err)
	}
}

func TestGet_Expired(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// Manually inject an expired entry
	store.mu.Lock()
	store.data.Entries["project"] = map[string]cacheEntry{
		"1": {
			Data:       []byte(`{"id":1}`),
			CachedAt:   time.Now().Add(-2 * time.Hour).Unix(),
			TTLSeconds: 3600, // 1 hour
		},
	}
	store.mu.Unlock()

	var got struct{ ID int }
	err := store.Get("project", "1", &got)
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss for expired entry, got %v", err)
	}
}

func TestGetStale_ReturnsExpired(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// Manually inject an expired entry
	store.mu.Lock()
	store.data.Entries["project"] = map[string]cacheEntry{
		"1": {
			Data:       []byte(`{"id":1,"name":"Stale"}`),
			CachedAt:   time.Now().Add(-2 * time.Hour).Unix(),
			TTLSeconds: 3600,
		},
	}
	store.mu.Unlock()

	var got struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	err := store.GetStale("project", "1", &got)
	if err != nil {
		t.Fatalf("GetStale error: %v", err)
	}

	if got.Name != "Stale" {
		t.Errorf("expected name 'Stale', got '%s'", got.Name)
	}
}

func TestGetStale_Miss(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	var got struct{}
	err := store.GetStale("project", "nonexistent", &got)
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss, got %v", err)
	}
}

func TestSet_ZeroTTL_Skips(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// "active_entry" has TTL=0, should not be cached
	err := store.Set("active_entry", "1", map[string]int{"id": 1})
	if err != nil {
		t.Fatalf("Set error: %v", err)
	}

	var got struct{}
	err = store.Get("active_entry", "1", &got)
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss for zero-TTL type, got %v", err)
	}
}

func TestInvalidateType(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	store.Set("projects", "all", []int{1, 2})
	store.Set("project", "1", map[string]int{"id": 1})
	store.Set("tasks", "all", []int{3})

	err := store.InvalidateType("projects", "project")
	if err != nil {
		t.Fatalf("InvalidateType error: %v", err)
	}

	var got struct{}
	if store.Get("projects", "all", &got) != ErrCacheMiss {
		t.Error("expected projects to be invalidated")
	}
	if store.Get("project", "1", &got) != ErrCacheMiss {
		t.Error("expected project to be invalidated")
	}
	// tasks should still be present
	var tasks []int
	if err := store.Get("tasks", "all", &tasks); err != nil {
		t.Errorf("tasks should still be cached, got error: %v", err)
	}
}

func TestClear(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	store.Set("project", "1", map[string]int{"id": 1})
	store.Set("tasks", "all", []int{1})

	err := store.Clear()
	if err != nil {
		t.Fatalf("Clear error: %v", err)
	}

	stats, _ := store.Stats()
	if len(stats) != 0 {
		t.Errorf("expected empty cache after Clear, got %d types", len(stats))
	}
}

func TestPrune(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// Add a fresh entry
	store.Set("project", "1", map[string]int{"id": 1})

	// Manually inject an expired entry
	store.mu.Lock()
	if store.data.Entries["project"] == nil {
		store.data.Entries["project"] = make(map[string]cacheEntry)
	}
	store.data.Entries["project"]["2"] = cacheEntry{
		Data:       []byte(`{"id":2}`),
		CachedAt:   time.Now().Add(-2 * time.Hour).Unix(),
		TTLSeconds: 3600,
	}
	store.mu.Unlock()

	err := store.Prune()
	if err != nil {
		t.Fatalf("Prune error: %v", err)
	}

	// Fresh entry should remain
	var got struct{ ID int }
	if err := store.Get("project", "1", &got); err != nil {
		t.Errorf("fresh entry should remain after prune, got error: %v", err)
	}

	// Expired entry should be gone
	if err := store.GetStale("project", "2", &got); err != ErrCacheMiss {
		t.Errorf("expired entry should be pruned, got: %v", err)
	}
}

func TestLookupName_Project(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// Cache a project entry
	store.Set("project", "42", map[string]interface{}{
		"id":   42,
		"name": "My Test Project",
	})

	id, err := store.LookupName("project", "test", 0)
	if err != nil {
		t.Fatalf("LookupName error: %v", err)
	}
	if id != 42 {
		t.Errorf("expected ID 42, got %d", id)
	}
}

func TestLookupName_Task(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	store.Set("task", "10", map[string]interface{}{
		"id":         10,
		"name":       "Design Homepage",
		"project_id": 5,
	})

	id, err := store.LookupName("task", "design", 5)
	if err != nil {
		t.Fatalf("LookupName error: %v", err)
	}
	if id != 10 {
		t.Errorf("expected ID 10, got %d", id)
	}
}

func TestLookupName_TaskWrongProject(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	store.Set("task", "10", map[string]interface{}{
		"id":         10,
		"name":       "Design Homepage",
		"project_id": 5,
	})

	_, err := store.LookupName("task", "design", 999)
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss for wrong project, got %v", err)
	}
}

func TestLookupName_Miss(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	_, err := store.LookupName("project", "nonexistent", 0)
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss, got %v", err)
	}
}

func TestStats(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	store.Set("project", "1", map[string]int{"id": 1})
	store.Set("project", "2", map[string]int{"id": 2})
	store.Set("tasks", "all", []int{1, 2, 3})

	stats, err := store.Stats()
	if err != nil {
		t.Fatalf("Stats error: %v", err)
	}

	if stats["project"] != 2 {
		t.Errorf("expected 2 project entries, got %d", stats["project"])
	}
	if stats["tasks"] != 1 {
		t.Errorf("expected 1 tasks entry, got %d", stats["tasks"])
	}
}

func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	// Write and close
	store1, _ := Open(path)
	store1.Set("project", "1", map[string]int{"id": 1})
	store1.Close()

	// Reopen and read
	store2, _ := Open(path)
	defer store2.Close()

	var got struct{ ID int }
	err := store2.Get("project", "1", &got)
	if err != nil {
		t.Fatalf("expected cached data to persist, got error: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("expected ID 1, got %d", got.ID)
	}
}

func TestGetTTL(t *testing.T) {
	tests := []struct {
		resourceType string
		expected     time.Duration
	}{
		{"me", 24 * time.Hour},
		{"projects", 1 * time.Hour},
		{"project", 1 * time.Hour},
		{"tasks", 30 * time.Minute},
		{"task", 30 * time.Minute},
		{"entries", 5 * time.Minute},
		{"active_entry", 0},
		{"unknown_type", 1 * time.Hour}, // default
	}

	for _, tc := range tests {
		t.Run(tc.resourceType, func(t *testing.T) {
			got := getTTL(tc.resourceType)
			if got != tc.expected {
				t.Errorf("getTTL(%q) = %v, want %v", tc.resourceType, got, tc.expected)
			}
		})
	}
}

// --- helpers ---

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open test store: %v", err)
	}
	return store
}

func itoa(n int64) string {
	return fmt.Sprintf("%d", n)
}
