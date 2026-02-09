package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ErrCacheMiss indicates the requested key was not found or has expired.
var ErrCacheMiss = errors.New("cache miss")

// Default TTLs per resource type.
var DefaultTTL = map[string]time.Duration{
	"me":              24 * time.Hour,
	"projects":        1 * time.Hour,
	"project":         1 * time.Hour,
	"project_by_name": 1 * time.Hour,
	"tasks":           30 * time.Minute,
	"task":            30 * time.Minute,
	"task_by_name":    30 * time.Minute,
	"tasklists":       1 * time.Hour,
	"entries":         5 * time.Minute,
	"entry":           5 * time.Minute,
	"active_entry":    0, // never cache
}

// cacheEntry is a single cached value.
type cacheEntry struct {
	Data       json.RawMessage `json:"data"`
	CachedAt   int64           `json:"cached_at"`
	TTLSeconds int64           `json:"ttl_seconds"`
}

// cacheData is the top-level JSON structure persisted to disk.
type cacheData struct {
	Entries map[string]map[string]cacheEntry `json:"entries"` // resource_type -> cache_key -> entry
}

// Store is the JSON file-backed cache store.
type Store struct {
	mu   sync.Mutex
	path string
	data cacheData
}

// Open opens (or creates) the cache file at the given path.
func Open(cachePath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0700); err != nil {
		return nil, fmt.Errorf("creating cache dir: %w", err)
	}
	s := &Store{
		path: cachePath,
		data: cacheData{
			Entries: make(map[string]map[string]cacheEntry),
		},
	}
	// Try to load existing cache
	raw, err := os.ReadFile(cachePath)
	if err == nil && len(raw) > 0 {
		if json.Unmarshal(raw, &s.data) != nil {
			// Corrupt cache — start fresh
			s.data.Entries = make(map[string]map[string]cacheEntry)
		}
		if s.data.Entries == nil {
			s.data.Entries = make(map[string]map[string]cacheEntry)
		}
	}
	return s, nil
}

// Close flushes the cache to disk.
func (s *Store) Close() error {
	return s.flush()
}

func (s *Store) flush() error {
	s.mu.Lock()
	raw, err := json.Marshal(s.data)
	s.mu.Unlock()
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, raw, 0644)
}

// Get retrieves a cached entry. Returns ErrCacheMiss if not found or expired.
func (s *Store) Get(resourceType, cacheKey string, dest interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.data.Entries[resourceType]
	if !ok {
		return ErrCacheMiss
	}
	entry, ok := bucket[cacheKey]
	if !ok {
		return ErrCacheMiss
	}
	if time.Now().Unix()-entry.CachedAt > entry.TTLSeconds {
		return ErrCacheMiss
	}
	return json.Unmarshal(entry.Data, dest)
}

// GetStale retrieves cached data even if expired (for offline fallback).
func (s *Store) GetStale(resourceType, cacheKey string, dest interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.data.Entries[resourceType]
	if !ok {
		return ErrCacheMiss
	}
	entry, ok := bucket[cacheKey]
	if !ok {
		return ErrCacheMiss
	}
	return json.Unmarshal(entry.Data, dest)
}

// Set stores a value in the cache and flushes to disk.
func (s *Store) Set(resourceType, cacheKey string, value interface{}) error {
	ttl := getTTL(resourceType)
	if ttl == 0 {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.mu.Lock()
	if s.data.Entries[resourceType] == nil {
		s.data.Entries[resourceType] = make(map[string]cacheEntry)
	}
	s.data.Entries[resourceType][cacheKey] = cacheEntry{
		Data:       data,
		CachedAt:   time.Now().Unix(),
		TTLSeconds: int64(ttl.Seconds()),
	}
	s.mu.Unlock()
	return s.flush()
}

// InvalidateType removes all entries for the given resource types.
func (s *Store) InvalidateType(resourceTypes ...string) error {
	s.mu.Lock()
	for _, rt := range resourceTypes {
		delete(s.data.Entries, rt)
	}
	s.mu.Unlock()
	return s.flush()
}

// Clear removes all cached data.
func (s *Store) Clear() error {
	s.mu.Lock()
	s.data.Entries = make(map[string]map[string]cacheEntry)
	s.mu.Unlock()
	return s.flush()
}

// Prune removes expired entries.
func (s *Store) Prune() error {
	now := time.Now().Unix()
	s.mu.Lock()
	for rt, bucket := range s.data.Entries {
		for key, entry := range bucket {
			if now-entry.CachedAt > entry.TTLSeconds {
				delete(bucket, key)
			}
		}
		if len(bucket) == 0 {
			delete(s.data.Entries, rt)
		}
	}
	s.mu.Unlock()
	return s.flush()
}

// IndexName is a no-op in the JSON store — name lookups scan cached entries directly.
func (s *Store) IndexName(resourceType, nameLower string, id, projectID int) {
	// Name index is implicit from cached entries
}

// LookupName searches cached individual entries for a name match.
func (s *Store) LookupName(resourceType, nameLower string, projectID int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if resourceType == "project" {
		bucket := s.data.Entries["project"]
		for _, entry := range bucket {
			var p struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}
			if json.Unmarshal(entry.Data, &p) == nil {
				if strings.Contains(strings.ToLower(p.Name), nameLower) {
					return p.ID, nil
				}
			}
		}
	}

	if resourceType == "task" {
		bucket := s.data.Entries["task"]
		for _, entry := range bucket {
			var t struct {
				ID        int    `json:"id"`
				Name      string `json:"name"`
				ProjectID int    `json:"project_id"`
			}
			if json.Unmarshal(entry.Data, &t) == nil {
				if t.ProjectID == projectID && strings.Contains(strings.ToLower(t.Name), nameLower) {
					return t.ID, nil
				}
			}
		}
	}

	return 0, ErrCacheMiss
}

// Stats returns cache statistics.
func (s *Store) Stats() (map[string]int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stats := make(map[string]int)
	for rt, bucket := range s.data.Entries {
		stats[rt] = len(bucket)
	}
	return stats, nil
}

func getTTL(resourceType string) time.Duration {
	if ttl, ok := DefaultTTL[resourceType]; ok {
		return ttl
	}
	return 1 * time.Hour
}
