package monday

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	defaultBaseURL    = "https://api.monday.com/v2"
	defaultAPIVersion = "2025-01"
	defaultTimeout    = 30
)

// ColumnRoles maps semantic roles to a board's actual column IDs. Roles that a
// board does not have are left empty.
type ColumnRoles struct {
	People   string `json:"people,omitempty"`
	Status   string `json:"status,omitempty"`
	Priority string `json:"priority,omitempty"`
	Type     string `json:"type,omitempty"`
	Sprint   string `json:"sprint,omitempty"`
	Active   string `json:"active,omitempty"`
}

// Get returns the column ID for a role name ("people", "status", ...).
func (r ColumnRoles) Get(role string) string {
	switch strings.ToLower(role) {
	case "people":
		return r.People
	case "status":
		return r.Status
	case "priority":
		return r.Priority
	case "type":
		return r.Type
	case "sprint":
		return r.Sprint
	case "active":
		return r.Active
	}
	return ""
}

// Set assigns a column ID to a role name. Returns false for an unknown role.
func (r *ColumnRoles) Set(role, columnID string) bool {
	switch strings.ToLower(role) {
	case "people":
		r.People = columnID
	case "status":
		r.Status = columnID
	case "priority":
		r.Priority = columnID
	case "type":
		r.Type = columnID
	case "sprint":
		r.Sprint = columnID
	case "active":
		r.Active = columnID
	default:
		return false
	}
	return true
}

// RoleNames is the set of valid role names.
var RoleNames = []string{"people", "status", "priority", "type", "sprint", "active"}

// Favorite is a board the user has registered under a name, carrying its own
// column-role map plus label indexes for its status-like columns so filtering
// and mutations can translate label text -> index without re-fetching.
type Favorite struct {
	ID          string      `json:"id"`
	Terminology string      `json:"terminology,omitempty"`
	Columns     ColumnRoles `json:"columns"`
	// Labels maps a role name ("status"/"priority"/"type") to a
	// lower-cased-label -> index map.
	Labels map[string]map[string]int `json:"labels,omitempty"`
}

// Config is the on-disk CLI configuration.
type Config struct {
	APIToken     string              `json:"api_token"`
	BaseURL      string              `json:"base_url"`
	APIVersion   string              `json:"api_version"`
	Timeout      int                 `json:"timeout_seconds"`
	User         User                `json:"user"`
	DefaultBoard string              `json:"default_board,omitempty"`
	Boards       map[string]Favorite `json:"boards"`
}

// DefaultConfig returns a fresh configuration with sane defaults.
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    defaultBaseURL,
		APIVersion: defaultAPIVersion,
		Timeout:    defaultTimeout,
		Boards:     map[string]Favorite{},
	}
}

// ConfigPath returns the config file location, honoring the MONDAY_CONFIG env var.
func ConfigPath() string {
	if p := os.Getenv("MONDAY_CONFIG"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "./monday-config.json"
	}
	return filepath.Join(home, ".config", "monday-cli", "config.json")
}

// LoadConfig reads the config file, creating a default one if it is absent, and
// migrates any legacy single-board config it finds.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			return cfg, cfg.Save(path)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	cfg.migrateLegacy(data)
	if cfg.Boards == nil {
		cfg.Boards = map[string]Favorite{}
	}
	return cfg, nil
}

// migrateLegacy carries over the token and user identity from the old flat
// schema (api_key / user_* fields). Board setup is not migrated because the old
// schema stored no column-role map — the user re-registers boards with `board add`.
func (c *Config) migrateLegacy(data []byte) {
	if c.APIToken != "" {
		return
	}
	var legacy struct {
		APIKey    string `json:"api_key"`
		UserID    string `json:"user_id"`
		UserName  string `json:"user_name"`
		UserEmail string `json:"user_email"`
		UserTitle string `json:"user_title"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return
	}
	c.APIToken = legacy.APIKey
	if c.User.ID == "" {
		c.User = User{ID: legacy.UserID, Name: legacy.UserName, Email: legacy.UserEmail, Title: legacy.UserTitle}
	}
}

// Save writes the config to disk (0600 — it holds a secret token).
func (c *Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// Token returns the API token, preferring the MONDAY_API_TOKEN env var.
func (c *Config) Token() string {
	if t := os.Getenv("MONDAY_API_TOKEN"); t != "" {
		return t
	}
	return c.APIToken
}

// ResolveBoard resolves a favorite by name, falling back to the default board
// when name is empty. It returns the resolved name alongside the favorite.
func (c *Config) ResolveBoard(name string) (string, Favorite, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = c.DefaultBoard
	}
	if name == "" {
		return "", Favorite{}, fmt.Errorf("no board specified and no default set — use -b <name> or `mon board default <name>`")
	}
	fav, ok := c.Boards[name]
	if !ok {
		return "", Favorite{}, fmt.Errorf("unknown board %q — see `mon board ls` (registered: %s)", name, strings.Join(c.BoardNames(), ", "))
	}
	return name, fav, nil
}

// BoardNames returns the registered favorite names, sorted.
func (c *Config) BoardNames() []string {
	names := make([]string, 0, len(c.Boards))
	for n := range c.Boards {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
