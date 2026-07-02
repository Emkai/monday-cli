package monday

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// boardCache holds the last fetched tasks for one board, keyed for both
// id-based and short LocalID-based lookup.
type boardCache struct {
	Tasks      map[string]Task `json:"tasks"`
	LocalIDMap map[int]string  `json:"local_id_map"`
	Timestamp  time.Time       `json:"timestamp"`
}

// DataStore is a per-board on-disk cache of fetched tasks. The LocalID short
// index lets the CLI reference items by a small integer (e.g. `items show 3`).
type DataStore struct {
	cache map[string]boardCache
}

// NewDataStore loads the cache from disk once. A missing/corrupt cache starts empty.
func NewDataStore() *DataStore {
	ds := &DataStore{cache: map[string]boardCache{}}
	_ = ds.load()
	return ds
}

// StoreTasks replaces the cached task set for a board, assigning LocalIDs from
// each task's LocalID field (see TasksFromItems).
func (ds *DataStore) StoreTasks(boardID string, tasks []Task) error {
	bc := boardCache{
		Tasks:      make(map[string]Task, len(tasks)),
		LocalIDMap: make(map[int]string, len(tasks)),
		Timestamp:  time.Now(),
	}
	for _, t := range tasks {
		bc.Tasks[t.ID] = t
		bc.LocalIDMap[t.LocalID] = t.ID
	}
	ds.cache[boardID] = bc
	return ds.save()
}

// CachedTasks returns the cached tasks for a board (sorted by LocalID) and the
// time they were fetched.
func (ds *DataStore) CachedTasks(boardID string) ([]Task, time.Time, bool) {
	bc, ok := ds.cache[boardID]
	if !ok {
		return nil, time.Time{}, false
	}
	tasks := make([]Task, 0, len(bc.Tasks))
	for _, t := range bc.Tasks {
		tasks = append(tasks, t)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].LocalID < tasks[j].LocalID })
	return tasks, bc.Timestamp, true
}

// TaskByLocalID looks up a single cached task by its short index.
func (ds *DataStore) TaskByLocalID(boardID string, localID int) (Task, bool) {
	bc, ok := ds.cache[boardID]
	if !ok {
		return Task{}, false
	}
	id, ok := bc.LocalIDMap[localID]
	if !ok {
		return Task{}, false
	}
	t, ok := bc.Tasks[id]
	return t, ok
}

// UpdateTask replaces a cached task by its ID, preserving its LocalID mapping.
func (ds *DataStore) UpdateTask(boardID string, task Task) error {
	bc, ok := ds.cache[boardID]
	if !ok {
		return fmt.Errorf("board %s not in cache", boardID)
	}
	if existing, ok := bc.Tasks[task.ID]; ok && task.LocalID == 0 {
		task.LocalID = existing.LocalID
	}
	bc.Tasks[task.ID] = task
	ds.cache[boardID] = bc
	return ds.save()
}

// ClearCache drops the cache for a board.
func (ds *DataStore) ClearCache(boardID string) error {
	delete(ds.cache, boardID)
	return ds.save()
}

func cachePath() (string, error) {
	if p := os.Getenv("MONDAY_CACHE"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".cache", "monday-cli", "tasks.json"), nil
}

func (ds *DataStore) save() error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	data, err := json.Marshal(ds.cache)
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}
	return nil
}

func (ds *DataStore) load() error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read cache file: %w", err)
	}
	return json.Unmarshal(data, &ds.cache)
}
