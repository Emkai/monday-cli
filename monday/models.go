package monday

import (
	"encoding/json"
	"time"
)

// Board represents a Monday.com board
type Board struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	UpdatedAt   time.Time `json:"updated_at"`
	Columns     []Column  `json:"columns,omitempty"`
	Items       []Item    `json:"items,omitempty"`
}

// Column represents a Monday.com board column
type Column struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Type        string          `json:"type"`
	SettingsStr string          `json:"settings_str"`
	Settings    json.RawMessage `json:"settings,omitempty"`
}

// Item represents a Monday.com board item
type Item struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	ColumnValues []ColumnValue `json:"column_values"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// ColumnValue represents a column value for an item
type ColumnValue struct {
	ID    string          `json:"id"`
	Text  string          `json:"text"`
	Value json.RawMessage `json:"value"`
}

// Update represents a Monday.com update (comment)
type Update struct {
	ID        string    `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	Creator   User      `json:"creator"`
}

// User represents a Monday.com user
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Workspace represents a Monday.com workspace
type Workspace struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
}

// Group represents a Monday.com board group
type Group struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Color string `json:"color"`
}

// Tag represents a Monday.com tag
type Tag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Asset represents a Monday.com asset (file)
type Asset struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	FileSize int64  `json:"file_size"`
}

// Notification represents a Monday.com notification
type Notification struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	Read      bool      `json:"read"`
}

// Team represents a Monday.com team
type Team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// BoardView represents a Monday.com board view
type BoardView struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Settings json.RawMessage `json:"settings"`
}

// ActivityLog represents a Monday.com activity log entry
type ActivityLog struct {
	ID        string          `json:"id"`
	Event     string          `json:"event"`
	CreatedAt time.Time       `json:"created_at"`
	Data      json.RawMessage `json:"data"`
}

// Webhook represents a Monday.com webhook
type Webhook struct {
	ID       string          `json:"id"`
	BoardID  string          `json:"board_id"`
	URL      string          `json:"url"`
	IsActive bool            `json:"is_active"`
	Config   json.RawMessage `json:"config"`
}

// Integration represents a Monday.com integration
type Integration struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	IsActive bool            `json:"is_active"`
	Settings json.RawMessage `json:"settings"`
}
