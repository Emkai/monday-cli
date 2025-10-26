package monday

import (
	"encoding/json"
	"time"
)

type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"
	PriorityNone     Priority = ""
)

type Status string

const (
	StatusDone             Status = "done"
	StatusInProgress       Status = "in progress"
	StatusStuck            Status = "stuck"
	StatusWaitingForReview Status = "waiting for review"
	StatusReadyForTesting  Status = "ready for testing"
	StatusRemoved          Status = "removed"
	StatusNone             Status = ""
)

type Type string

const (
	TypeBug      Type = "Bug"
	TypeFeature  Type = "Feature"
	TypeTest     Type = "Test"
	TypeSecurity Type = "Security"
	TypeQuality  Type = "Quality"
	TypeOther    Type = "Other"
	TypeNone     Type = ""
)

type Sprint string

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
type Task struct {
	localId   int       `json:"local_id"`
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    Status    `json:"status"`
	Priority  Priority  `json:"priority"`
	Type      Type      `json:"type"`
	Sprint    Sprint    `json:"sprint"`
	UserName  string    `json:"user_name"`
	UserEmail string    `json:"user_email"`
	UpdatedAt time.Time `json:"updated_at"`
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

// User represents a Monday.com user
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Title    string `json:"title"`
	PhotoURL string `json:"photo_small"`
	Enabled  bool   `json:"enabled"`
}
