package monday

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Filters struct {
	UserNameWhitelist  []string `json:"user_name_whitelist"`
	UserNameBlacklist  []string `json:"user_name_blacklist"`
	UserEmailWhitelist []string `json:"user_email_whitelist"`
	UserEmailBlacklist []string `json:"user_email_blacklist"`
	StatusWhitelist    []string `json:"status_whitelist"`
	StatusBlacklist    []string `json:"status_blacklist"`
	PriorityWhitelist  []string `json:"priority_whitelist"`
	PriorityBlacklist  []string `json:"priority_blacklist"`
	TypeWhitelist      []string `json:"type_whitelist"`
	TypeBlacklist      []string `json:"type_blacklist"`
	SprintWhitelist    []string `json:"sprint_whitelist"`
	SprintBlacklist    []string `json:"sprint_blacklist"`
}

// Config represents Monday.com configuration
type Config struct {
	APIKey     string  `json:"api_key"`
	BaseURL    string  `json:"base_url"`
	Timeout    int     `json:"timeout_seconds"`
	OwnerEmail string  `json:"owner_email"`
	BoardID    string  `json:"board_id"`
	SprintID   string  `json:"sprint_id"`
	UserID     string  `json:"user_id"`
	UserName   string  `json:"user_name"`
	UserEmail  string  `json:"user_email"`
	UserTitle  string  `json:"user_title"`
	Filters    Filters `json:"filters"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    "https://api.monday.com/v2",
		Timeout:    30,
		OwnerEmail: "",
		BoardID:    "",
		SprintID:   "",
		Filters: Filters{
			UserNameWhitelist:  []string{},
			UserNameBlacklist:  []string{},
			UserEmailWhitelist: []string{},
			UserEmailBlacklist: []string{},
			StatusWhitelist:    []string{},
			StatusBlacklist:    []string{},
			PriorityWhitelist:  []string{},
			PriorityBlacklist:  []string{},
			TypeWhitelist:      []string{},
			TypeBlacklist:      []string{},
			SprintWhitelist:    []string{},
			SprintBlacklist:    []string{},
		},
	}
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config if it doesn't exist
		config := DefaultConfig()
		if err := config.Save(configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save saves configuration to file
func (c *Config) Save(configPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the default config file path
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./monday-config.json"
	}
	return filepath.Join(homeDir, ".config", "monday-cli", "config.json")
}

// SetAPIKey sets the API key in the configuration
func (c *Config) SetAPIKey(apiKey string) {
	c.APIKey = apiKey
}

// GetAPIKey returns the API key
func (c *Config) GetAPIKey() string {
	return c.APIKey
}

// SetOwnerEmail sets the owner email in the configuration
func (c *Config) SetOwnerEmail(ownerEmail string) {
	c.OwnerEmail = ownerEmail
}

// GetOwnerEmail returns the owner email
func (c *Config) GetOwnerEmail() string {
	return c.OwnerEmail
}

// SetBoardID sets the board ID in the configuration
func (c *Config) SetBoardID(boardID string) {
	c.BoardID = boardID
}

// GetBoardID returns the board ID
func (c *Config) GetBoardID() string {
	return c.BoardID
}

// IsConfigured checks if the configuration is complete
func (c *Config) IsConfigured() bool {
	return c.APIKey != "" && c.HasUserInfo() && c.BoardID != ""
}

// SetSprintID sets the sprint ID in the configuration
func (c *Config) SetSprintID(sprintID string) {
	c.SprintID = sprintID
}

// GetSprintID returns the sprint ID
func (c *Config) GetSprintID() string {
	return c.SprintID
}

func (c *Config) AddStatusWhitelist(status string) {
	c.Filters.StatusWhitelist = append(c.Filters.StatusWhitelist, status)
}

func (c *Config) RemoveStatusWhitelist(status string) {
	c.Filters.StatusWhitelist = removeFromSlice(c.Filters.StatusWhitelist, status)
}

func (c *Config) AddStatusBlacklist(status string) {
	c.Filters.StatusBlacklist = append(c.Filters.StatusBlacklist, status)
}

func (c *Config) RemoveStatusBlacklist(status string) {
	c.Filters.StatusBlacklist = removeFromSlice(c.Filters.StatusBlacklist, status)
}

func (c *Config) AddPriorityWhitelist(priority string) {
	c.Filters.PriorityWhitelist = append(c.Filters.PriorityWhitelist, priority)
}

func (c *Config) RemovePriorityWhitelist(priority string) {
	c.Filters.PriorityWhitelist = removeFromSlice(c.Filters.PriorityWhitelist, priority)
}

func (c *Config) AddPriorityBlacklist(priority string) {
	c.Filters.PriorityBlacklist = append(c.Filters.PriorityBlacklist, priority)
}

func (c *Config) RemovePriorityBlacklist(priority string) {
	c.Filters.PriorityBlacklist = removeFromSlice(c.Filters.PriorityBlacklist, priority)
}

func (c *Config) AddTypeWhitelist(taskType string) {
	c.Filters.TypeWhitelist = append(c.Filters.TypeWhitelist, taskType)
}

func (c *Config) RemoveTypeWhitelist(taskType string) {
	c.Filters.TypeWhitelist = removeFromSlice(c.Filters.TypeWhitelist, taskType)
}

func (c *Config) AddTypeBlacklist(taskType string) {
	c.Filters.TypeBlacklist = append(c.Filters.TypeBlacklist, taskType)
}

func (c *Config) RemoveTypeBlacklist(taskType string) {
	c.Filters.TypeBlacklist = removeFromSlice(c.Filters.TypeBlacklist, taskType)
}

func (c *Config) AddSprintWhitelist(sprint string) {
	c.Filters.SprintWhitelist = append(c.Filters.SprintWhitelist, sprint)
}

func (c *Config) RemoveSprintWhitelist(sprint string) {
	c.Filters.SprintWhitelist = removeFromSlice(c.Filters.SprintWhitelist, sprint)
}

func (c *Config) AddSprintBlacklist(sprint string) {
	c.Filters.SprintBlacklist = append(c.Filters.SprintBlacklist, sprint)
}

func (c *Config) RemoveSprintBlacklist(sprint string) {
	c.Filters.SprintBlacklist = removeFromSlice(c.Filters.SprintBlacklist, sprint)
}

func (c *Config) AddUserNameWhitelist(userName string) {
	c.Filters.UserNameWhitelist = append(c.Filters.UserNameWhitelist, userName)
}
func (c *Config) RemoveUserNameWhitelist(userName string) {
	c.Filters.UserNameWhitelist = removeFromSlice(c.Filters.UserNameWhitelist, userName)
}

func (c *Config) AddUserNameBlacklist(userName string) {
	c.Filters.UserNameBlacklist = append(c.Filters.UserNameBlacklist, userName)
}

func (c *Config) RemoveUserNameBlacklist(userName string) {
	c.Filters.UserNameBlacklist = removeFromSlice(c.Filters.UserNameBlacklist, userName)
}

func (c *Config) AddUserEmailWhitelist(userEmail string) {
	c.Filters.UserEmailWhitelist = append(c.Filters.UserEmailWhitelist, userEmail)
}

func (c *Config) RemoveUserEmailWhitelist(userEmail string) {
	c.Filters.UserEmailWhitelist = removeFromSlice(c.Filters.UserEmailWhitelist, userEmail)
}

func (c *Config) AddUserEmailBlacklist(userEmail string) {
	c.Filters.UserEmailBlacklist = append(c.Filters.UserEmailBlacklist, userEmail)
}

func (c *Config) RemoveUserEmailBlacklist(userEmail string) {
	c.Filters.UserEmailBlacklist = removeFromSlice(c.Filters.UserEmailBlacklist, userEmail)
}

// GetFilters returns the filters
func (c *Config) GetFilters() Filters {
	return c.Filters
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./monday-config.json"
	}
	return filepath.Join(homeDir, ".config", "monday-cli", "config.json")
}

// SetUserInfo sets the user information in the configuration
func (c *Config) SetUserInfo(user *User) {
	c.UserID = user.ID
	c.UserName = user.Name
	c.UserEmail = user.Email
	c.UserTitle = user.Title
	// Also set the owner email to the user's email for backward compatibility
	c.OwnerEmail = user.Email
}

// GetUserInfo returns the user information from the configuration
func (c *Config) GetUserInfo() *User {
	return &User{
		ID:    c.UserID,
		Name:  c.UserName,
		Email: c.UserEmail,
		Title: c.UserTitle,
	}
}

// HasUserInfo checks if user information is available
func (c *Config) HasUserInfo() bool {
	return c.UserID != "" && c.UserEmail != ""
}

// GetUserEmail returns the user email
func (c *Config) GetUserEmail() string {
	return c.UserEmail
}

func removeFromSlice(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
