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
	Tasks      map[string]Task
	LocalIdMap map[int]string // Maps local index to task ID
	RawItems   map[string]Item
	Users      map[string]User // Maps user ID to User
	Sprints    []Sprint        // List of sprints found on the board
	Timestamp  time.Time
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

func (ds *DataStore) StoreRawItems(boardID string, items []Item) {
	if _, exists := ds.cache[boardID]; !exists {
		ds.cache[boardID] = TaskCache{
			Tasks:      make(map[string]Task),
			LocalIdMap: make(map[int]string),
			RawItems:   make(map[string]Item),
			Users:      make(map[string]User),
			Timestamp:  time.Now(),
		}
	}
	for _, item := range items {
		ds.cache[boardID].RawItems[item.ID] = item
	}
	if err := ds.Save(); err != nil {
		fmt.Printf("Failed to save cache: %v\n", err)
	}
}

// StoreBoardUsers stores board users in the cache
func (ds *DataStore) StoreBoardUsers(boardID string, users []User) {
	if _, exists := ds.cache[boardID]; !exists {
		ds.cache[boardID] = TaskCache{
			Tasks:      make(map[string]Task),
			LocalIdMap: make(map[int]string),
			RawItems:   make(map[string]Item),
			Users:      make(map[string]User),
			Timestamp:  time.Now(),
		}
	}
	for _, user := range users {
		ds.cache[boardID].Users[user.ID] = user
	}
	if err := ds.Save(); err != nil {
		fmt.Printf("Failed to save cache: %v\n", err)
	}
}

// GetCachedBoardUsers retrieves cached board users
func (ds *DataStore) GetCachedBoardUsers(boardID string) ([]User, time.Time, bool) {
	if err := ds.Load(); err != nil {
		return []User{}, time.Time{}, false
	}

	if cached, exists := ds.cache[boardID]; exists {
		var users []User
		for _, user := range cached.Users {
			users = append(users, user)
		}
		return users, cached.Timestamp, true
	}
	return []User{}, time.Time{}, false
}

// StoreBoardSprints stores a slice of Sprint objects in the cache
func (ds *DataStore) StoreBoardSprints(boardID string, sprints []Sprint) {
	if err := ds.Load(); err != nil {
		fmt.Printf("Failed to load cache: %v\n", err)
		return
	}

	if _, exists := ds.cache[boardID]; !exists {
		ds.cache[boardID] = TaskCache{
			Tasks:      make(map[string]Task),
			LocalIdMap: make(map[int]string),
			RawItems:   make(map[string]Item),
			Users:      make(map[string]User),
			Sprints:    []Sprint{},
			Timestamp:  time.Now(),
		}
	}

	cache := ds.cache[boardID]
	cache.Sprints = sprints
	cache.Timestamp = time.Now()
	ds.cache[boardID] = cache

	if err := ds.Save(); err != nil {
		fmt.Printf("Failed to save cache: %v\n", err)
	}
}

// GetCachedBoardSprints retrieves cached Sprint objects
func (ds *DataStore) GetCachedBoardSprints(boardID string) ([]Sprint, time.Time, bool) {
	if err := ds.Load(); err != nil {
		return []Sprint{}, time.Time{}, false
	}

	if cached, exists := ds.cache[boardID]; exists {
		return cached.Sprints, cached.Timestamp, true
	}
	return []Sprint{}, time.Time{}, false
}

// StoreTasksRequest caches a task request result
func (ds *DataStore) StoreTasksRequest(boardID string, tasks []Task, rawItems []Item) {
	localIdMap := make(map[int]string)
	tasksMap := make(map[string]Task)
	for _, task := range tasks {
		tasksMap[task.ID] = task
		if _, exists := localIdMap[task.LocalId]; exists {
			fmt.Printf("Local ID %d already exists for task %s\n", task.LocalId, task.ID)
		}
		localIdMap[task.LocalId] = task.ID
	}
	rawItemsMap := make(map[string]Item)
	for _, item := range rawItems {
		rawItemsMap[item.ID] = item
	}

	ds.cache[boardID] = TaskCache{
		Tasks:      tasksMap,
		LocalIdMap: localIdMap,
		RawItems:   rawItemsMap,
		Users:      make(map[string]User),
		Sprints:    []Sprint{},
		Timestamp:  time.Now(),
	}

	if err := ds.Save(); err != nil {
		fmt.Printf("Failed to save cache: %v\n", err)
	}
}

// StoreTaskRequest caches a task request result
func (ds *DataStore) StoreTaskRequest(boardID string, task Task) (int, error) {
	if _, exists := ds.cache[boardID]; !exists {
		ds.cache[boardID] = TaskCache{
			Tasks:      make(map[string]Task),
			LocalIdMap: make(map[int]string),
			Timestamp:  time.Now(),
		}
	}
	ds.cache[boardID].Tasks[task.ID] = task
	localId, err := ds.GetTaskLocalIdByID(boardID, task.ID)
	if err != nil {
		fmt.Printf("Failed to get task local ID: %v\n", err)
		localId = len(ds.cache[boardID].LocalIdMap) + 1
	}
	task.LocalId = localId
	ds.cache[boardID].LocalIdMap[localId] = task.ID

	// Save cache to disk after update
	if err := ds.Save(); err != nil {
		fmt.Printf("Failed to save cache: %v\n", err)
		return 0, fmt.Errorf("failed to save cache: %v", err)
	}
	return localId, nil
}

// GetCachedTasks retrieves cached tasks if available
func (ds *DataStore) GetCachedTasks(boardID string) (map[string]Task, time.Time, bool) {
	if err := ds.Load(); err != nil {
		return make(map[string]Task), time.Time{}, false
	}

	if cached, exists := ds.cache[boardID]; exists {
		return cached.Tasks, cached.Timestamp, true
	}
	return nil, time.Time{}, false
}

func (ds *DataStore) GetCachedTask(boardID string, taskID string) (Task, time.Time, bool) {
	if err := ds.Load(); err != nil {
		return Task{}, time.Time{}, false
	}
	if cached, exists := ds.cache[boardID]; exists {
		return cached.Tasks[taskID], cached.Timestamp, true
	}
	return Task{}, time.Time{}, false
}

// GetCachedTaskByIndex retrieves a task by local index
func (ds *DataStore) GetCachedTaskByLocalId(boardID string, localId int) (Task, time.Time, bool) {
	if err := ds.Load(); err != nil {
		return Task{}, time.Time{}, false
	}
	if cached, exists := ds.cache[boardID]; exists {
		if taskID, exists := cached.LocalIdMap[localId]; exists {
			return cached.Tasks[taskID], cached.Timestamp, true
		}
	}
	return Task{}, time.Time{}, false
}

// GetIndexMap retrieves the index mapping for a board/owner combination
func (ds *DataStore) GetLocalIdMap(boardID string) (map[int]string, error) {
	if err := ds.Load(); err != nil {
		return nil, fmt.Errorf("failed to load cache: %w", err)
	}
	if cached, exists := ds.cache[boardID]; exists {
		return cached.LocalIdMap, nil
	}
	return nil, fmt.Errorf("board %s not found", boardID)
}

func (ds *DataStore) UpdateCachedTask(boardID string, taskID string, task Task) {
	ds.cache[boardID].Tasks[taskID] = task
	if err := ds.Save(); err != nil {
		fmt.Printf("Failed to update cached task: %v\n", err)
	}
}

// UpdateCachedTaskByIndex updates a task by local index
func (ds *DataStore) UpdateCachedTaskByLocalId(boardID string, localId int, task Task) {
	if cached, exists := ds.cache[boardID]; exists {
		if taskID, exists := cached.LocalIdMap[localId]; exists {
			cached.Tasks[taskID] = task
			if err := ds.Save(); err != nil {
				fmt.Printf("Failed to update cached task: %v\n", err)
			}
		}
	}
}

// ClearCache removes all cached entries
func (ds *DataStore) ClearCache(boardID string) {
	delete(ds.cache, boardID)

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

func (ds *DataStore) GetTaskLocalIdByID(boardID string, taskID string) (int, error) {
	if cached, exists := ds.cache[boardID]; exists {
		for localId, id := range cached.LocalIdMap {
			if id == taskID {
				return localId, nil
			}
		}
		ds.cache[boardID].LocalIdMap[len(cached.LocalIdMap)+1] = taskID
		return len(cached.LocalIdMap) + 1, nil
	}
	return -1, fmt.Errorf("board %s not found", boardID)
}
