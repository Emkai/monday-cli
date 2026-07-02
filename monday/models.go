package monday

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// Workspace is the workspace a board belongs to.
type Workspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Board represents a Monday.com board.
type Board struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	State       string     `json:"state"`
	BoardKind   string     `json:"board_kind"`
	Terminology string     `json:"item_terminology"`
	ItemsCount  int        `json:"items_count"`
	Workspace   *Workspace `json:"workspace,omitempty"`
	Columns     []Column   `json:"columns,omitempty"`
}

// Column represents a board column. For status-type columns, SettingsStr holds the
// label definitions (parse with ParseLabels).
type Column struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	SettingsStr string `json:"settings_str"`
}

// Task is a flattened, display-oriented view of a board item, resolved through a
// board's column-role map. Status/Priority/Type/Sprint hold the raw label text as
// shown in Monday (they differ per board, so they are plain strings, not enums).
type Task struct {
	LocalID   int       `json:"local_id"`
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Priority  string    `json:"priority"`
	Type      string    `json:"type"`
	Sprint    string    `json:"sprint"`
	Owner     string    `json:"owner"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Item is a raw Monday.com board item.
type Item struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	ColumnValues []ColumnValue `json:"column_values"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// ColumnValue is a single column's value on an item.
type ColumnValue struct {
	ID    string          `json:"id"`
	Text  string          `json:"text"`
	Value json.RawMessage `json:"value,omitempty"`
}

// User represents a Monday.com user.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Title string `json:"title"`
}

// ItemToTask projects a raw Item onto the Task model using a board's column roles.
// Empty role IDs simply never match (a ColumnValue ID is never empty), so boards
// missing a role (e.g. the Bugs board has no Type column) degrade gracefully.
func ItemToTask(item Item, roles ColumnRoles) Task {
	t := Task{ID: item.ID, Name: item.Name, UpdatedAt: item.UpdatedAt}
	for _, cv := range item.ColumnValues {
		switch cv.ID {
		case roles.People:
			t.Owner = cv.Text
		case roles.Status:
			t.Status = cv.Text
		case roles.Priority:
			t.Priority = cv.Text
		case roles.Type:
			t.Type = cv.Text
		case roles.Sprint:
			t.Sprint = cv.Text
		}
	}
	return t
}

// TasksFromItems converts a page of items to Tasks, assigning sequential 1-based
// LocalIDs for short reference from the CLI.
func TasksFromItems(items []Item, roles ColumnRoles) []Task {
	tasks := make([]Task, 0, len(items))
	for i, item := range items {
		t := ItemToTask(item, roles)
		t.LocalID = i + 1
		tasks = append(tasks, t)
	}
	return tasks
}

// ParseLabels parses a status-type column's settings_str into a map of
// lower-cased label text -> label index. Monday encodes settings_str as a JSON
// string of the form {"labels":{"0":"In Progress","1":"Done",...}}.
func ParseLabels(settingsStr string) map[string]int {
	var s struct {
		Labels map[string]string `json:"labels"`
	}
	if err := json.Unmarshal([]byte(settingsStr), &s); err != nil {
		return nil
	}
	out := make(map[string]int, len(s.Labels))
	for idxStr, label := range s.Labels {
		// The map is keyed by index (as string) -> label text.
		// We want label -> index, dropping blank labels.
		trimmed := strings.TrimSpace(label)
		if trimmed == "" {
			continue
		}
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			continue
		}
		out[strings.ToLower(trimmed)] = idx
	}
	return out
}
