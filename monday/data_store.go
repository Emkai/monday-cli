package monday

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TaskCache represents a cached task request
type TaskCache struct {
	Tasks     map[int]Item
	Timestamp time.Time
}

// DataStore manages caching of task requests
type DataStore struct {
	cache map[string]TaskCache
}

// NewDataStore creates a new DataStore instance
func NewDataStore() *DataStore {
	ds := &DataStore{
		cache: make(map[string]TaskCache),
	}
	if err := ds.Load(); err != nil {
		// Initialize empty cache if load fails
		ds.cache = make(map[string]TaskCache)
	}
	return ds
}

// StoreTaskRequest caches a task request result
func (ds *DataStore) StoreTaskRequest(boardID, ownerEmail string, tasks map[int]Item) {
	key := boardID + ownerEmail
	ds.cache[key] = TaskCache{
		Tasks:     tasks,
		Timestamp: time.Now(),
	}

	// Save cache to disk after update
	if err := ds.Save(); err != nil {
		fmt.Printf("Failed to save cache: %v\n", err)
	}
}

// GetCachedTasks retrieves cached tasks if available
func (ds *DataStore) GetCachedTasks(boardID, ownerEmail string) (map[int]Item, time.Time, bool) {
	if err := ds.Load(); err != nil {
		return make(map[int]Item), time.Time{}, false
	}

	key := boardID + ownerEmail
	if cached, exists := ds.cache[key]; exists {
		return cached.Tasks, cached.Timestamp, true
	}
	return nil, time.Time{}, false
}

func (ds *DataStore) GetCachedTask(boardID, ownerEmail string, taskID int) (Item, time.Time, bool) {
	if err := ds.Load(); err != nil {
		return Item{}, time.Time{}, false
	}
	key := boardID + ownerEmail
	if cached, exists := ds.cache[key]; exists {
		return cached.Tasks[taskID], cached.Timestamp, true
	}
	return Item{}, time.Time{}, false
}

func (ds *DataStore) UpdateCachedTask(boardID, ownerEmail string, taskID int, task Item) {
	key := boardID + ownerEmail
	ds.cache[key].Tasks[taskID] = task
	if err := ds.Save(); err != nil {
		fmt.Printf("Failed to update cached task: %v\n", err)
	}
}

// ClearCache removes all cached entries
func (ds *DataStore) ClearCache(boardID, ownerEmail string) {
	key := boardID + ownerEmail
	delete(ds.cache, key)

	// Save cache to disk after update
	if err := ds.Save(); err != nil {
		fmt.Printf("Failed to save cache: %v\n", err)
	}
}

// getCachePath returns the path to the cache file
func getCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".cache", "monday-cli", "tasks.json"), nil
}

// Save persists the cache to disk
func (ds *DataStore) Save() error {
	cachePath, err := getCachePath()
	if err != nil {
		return err
	}

	// Ensure cache directory exists
	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.Marshal(ds.cache)
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Load reads the cache from disk
func (ds *DataStore) Load() error {
	cachePath, err := getCachePath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Not an error if cache doesn't exist yet
		}
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	if err := json.Unmarshal(data, &ds.cache); err != nil {
		return fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	return nil
}
