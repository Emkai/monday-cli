package monday

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
	APIKey        string  `json:"api_key"`
	BaseURL       string  `json:"base_url"`
	Timeout       int     `json:"timeout_seconds"`
	BoardID       string  `json:"board_id"`
	SprintID      string  `json:"sprint_id"`
	SprintBoardId string  `json:"sprint_board_id"`
	UserID        string  `json:"user_id"`
	UserName      string  `json:"user_name"`
	UserEmail     string  `json:"user_email"`
	UserTitle     string  `json:"user_title"`
	Filters       Filters `json:"filters"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:       "https://api.monday.com/v2",
		Timeout:       30,
		BoardID:       "",
		SprintID:      "",
		SprintBoardId: "",
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

// SetSprintBoardID sets the sprint board ID in the configuration
func (c *Config) SetSprintBoardID(sprintBoardID string) {
	c.SprintBoardId = sprintBoardID
}

// GetSprintBoardID returns the sprint board ID
func (c *Config) GetSprintBoardID() string {
	return c.SprintBoardId
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

// FilterType represents the type of filter
type FilterType string

const (
	FilterStatus    FilterType = "status"
	FilterPriority  FilterType = "priority"
	FilterTaskType  FilterType = "type"
	FilterSprint    FilterType = "sprint"
	FilterUserName  FilterType = "user_name"
	FilterUserEmail FilterType = "user_email"
)

// FilterListType represents whether it's a whitelist or blacklist
type FilterListType string

const (
	Whitelist FilterListType = "whitelist"
	Blacklist FilterListType = "blacklist"
)

// AddFilter adds a value to the specified filter list
func (c *Config) AddFilter(filterType FilterType, listType FilterListType, value string) error {
	value = strings.ToLower(value)
	switch filterType {
	case FilterStatus:
		if listType == Whitelist {
			c.AddStatusWhitelist(value)
		} else {
			c.AddStatusBlacklist(value)
		}
	case FilterPriority:
		if listType == Whitelist {
			c.AddPriorityWhitelist(value)
		} else {
			c.AddPriorityBlacklist(value)
		}
	case FilterTaskType:
		if listType == Whitelist {
			c.AddTypeWhitelist(value)
		} else {
			c.AddTypeBlacklist(value)
		}
	case FilterSprint:
		if listType == Whitelist {
			c.AddSprintWhitelist(value)
		} else {
			c.AddSprintBlacklist(value)
		}
	case FilterUserName:
		if listType == Whitelist {
			c.AddUserNameWhitelist(value)
		} else {
			c.AddUserNameBlacklist(value)
		}
	case FilterUserEmail:
		if listType == Whitelist {
			c.AddUserEmailWhitelist(value)
		} else {
			c.AddUserEmailBlacklist(value)
		}
	default:
		return fmt.Errorf("unknown filter type: %s", filterType)
	}
	return nil
}

// RemoveFilter removes a value from the specified filter list
func (c *Config) RemoveFilter(filterType FilterType, listType FilterListType, value string) error {
	value = strings.ToLower(value)
	switch filterType {
	case FilterStatus:
		if listType == Whitelist {
			c.RemoveStatusWhitelist(value)
		} else {
			c.RemoveStatusBlacklist(value)
		}
	case FilterPriority:
		if listType == Whitelist {
			c.RemovePriorityWhitelist(value)
		} else {
			c.RemovePriorityBlacklist(value)
		}
	case FilterTaskType:
		if listType == Whitelist {
			c.RemoveTypeWhitelist(value)
		} else {
			c.RemoveTypeBlacklist(value)
		}
	case FilterSprint:
		if listType == Whitelist {
			c.RemoveSprintWhitelist(value)
		} else {
			c.RemoveSprintBlacklist(value)
		}
	case FilterUserName:
		if listType == Whitelist {
			c.RemoveUserNameWhitelist(value)
		} else {
			c.RemoveUserNameBlacklist(value)
		}
	case FilterUserEmail:
		if listType == Whitelist {
			c.RemoveUserEmailWhitelist(value)
		} else {
			c.RemoveUserEmailBlacklist(value)
		}
	default:
		return fmt.Errorf("unknown filter type: %s", filterType)
	}
	return nil
}

// ClearFilter clears all values from the specified filter list
func (c *Config) ClearFilter(filterType FilterType, listType FilterListType) error {
	switch filterType {
	case FilterStatus:
		if listType == Whitelist {
			c.Filters.StatusWhitelist = []string{}
		} else {
			c.Filters.StatusBlacklist = []string{}
		}
	case FilterPriority:
		if listType == Whitelist {
			c.Filters.PriorityWhitelist = []string{}
		} else {
			c.Filters.PriorityBlacklist = []string{}
		}
	case FilterTaskType:
		if listType == Whitelist {
			c.Filters.TypeWhitelist = []string{}
		} else {
			c.Filters.TypeBlacklist = []string{}
		}
	case FilterSprint:
		if listType == Whitelist {
			c.Filters.SprintWhitelist = []string{}
		} else {
			c.Filters.SprintBlacklist = []string{}
		}
	case FilterUserName:
		if listType == Whitelist {
			c.Filters.UserNameWhitelist = []string{}
		} else {
			c.Filters.UserNameBlacklist = []string{}
		}
	case FilterUserEmail:
		if listType == Whitelist {
			c.Filters.UserEmailWhitelist = []string{}
		} else {
			c.Filters.UserEmailBlacklist = []string{}
		}
	default:
		return fmt.Errorf("unknown filter type: %s", filterType)
	}
	return nil
}

// GetFilterValues returns the values for the specified filter list
func (c *Config) GetFilterValues(filterType FilterType, listType FilterListType) []string {
	switch filterType {
	case FilterStatus:
		if listType == Whitelist {
			return c.Filters.StatusWhitelist
		} else {
			return c.Filters.StatusBlacklist
		}
	case FilterPriority:
		if listType == Whitelist {
			return c.Filters.PriorityWhitelist
		} else {
			return c.Filters.PriorityBlacklist
		}
	case FilterTaskType:
		if listType == Whitelist {
			return c.Filters.TypeWhitelist
		} else {
			return c.Filters.TypeBlacklist
		}
	case FilterSprint:
		if listType == Whitelist {
			return c.Filters.SprintWhitelist
		} else {
			return c.Filters.SprintBlacklist
		}
	case FilterUserName:
		if listType == Whitelist {
			return c.Filters.UserNameWhitelist
		} else {
			return c.Filters.UserNameBlacklist
		}
	case FilterUserEmail:
		if listType == Whitelist {
			return c.Filters.UserEmailWhitelist
		} else {
			return c.Filters.UserEmailBlacklist
		}
	default:
		return []string{}
	}
}

// ClearAllFilters clears all filter lists
func (c *Config) ClearAllFilters() {
	c.Filters = Filters{
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
	}
}

// Convenience methods for current user filtering
// FilterToCurrentUser sets filters to show only tasks assigned to the current user
func (c *Config) FilterToCurrentUser() error {
	if !c.HasUserInfo() {
		return fmt.Errorf("user information not available - run 'user info' first")
	}

	// Clear existing user filters
	c.Filters.UserNameWhitelist = []string{}
	c.Filters.UserNameBlacklist = []string{}
	c.Filters.UserEmailWhitelist = []string{}
	c.Filters.UserEmailBlacklist = []string{}

	// Add current user to whitelist
	c.Filters.UserNameWhitelist = append(c.Filters.UserNameWhitelist, strings.ToLower(c.UserName))
	c.Filters.UserEmailWhitelist = append(c.Filters.UserEmailWhitelist, strings.ToLower(c.UserEmail))

	return nil
}

// AddCurrentUserToWhitelist adds the current user to the user whitelist
func (c *Config) AddCurrentUserToWhitelist() error {
	if !c.HasUserInfo() {
		return fmt.Errorf("user information not available - run 'user info' first")
	}

	// Add current user to whitelist if not already present
	userName := strings.ToLower(c.UserName)
	userEmail := strings.ToLower(c.UserEmail)

	if !slices.Contains(c.Filters.UserNameWhitelist, userName) {
		c.Filters.UserNameWhitelist = append(c.Filters.UserNameWhitelist, userName)
	}
	if !slices.Contains(c.Filters.UserEmailWhitelist, userEmail) {
		c.Filters.UserEmailWhitelist = append(c.Filters.UserEmailWhitelist, userEmail)
	}

	return nil
}

// RemoveCurrentUserFromWhitelist removes the current user from the user whitelist
func (c *Config) RemoveCurrentUserFromWhitelist() error {
	if !c.HasUserInfo() {
		return fmt.Errorf("user information not available - run 'user info' first")
	}

	userName := strings.ToLower(c.UserName)
	userEmail := strings.ToLower(c.UserEmail)

	c.Filters.UserNameWhitelist = removeFromSlice(c.Filters.UserNameWhitelist, userName)
	c.Filters.UserEmailWhitelist = removeFromSlice(c.Filters.UserEmailWhitelist, userEmail)

	return nil
}

// AddCurrentUserToBlacklist adds the current user to the user blacklist
func (c *Config) AddCurrentUserToBlacklist() error {
	if !c.HasUserInfo() {
		return fmt.Errorf("user information not available - run 'user info' first")
	}

	userName := strings.ToLower(c.UserName)
	userEmail := strings.ToLower(c.UserEmail)

	if !slices.Contains(c.Filters.UserNameBlacklist, userName) {
		c.Filters.UserNameBlacklist = append(c.Filters.UserNameBlacklist, userName)
	}
	if !slices.Contains(c.Filters.UserEmailBlacklist, userEmail) {
		c.Filters.UserEmailBlacklist = append(c.Filters.UserEmailBlacklist, userEmail)
	}

	return nil
}

// RemoveCurrentUserFromBlacklist removes the current user from the user blacklist
func (c *Config) RemoveCurrentUserFromBlacklist() error {
	if !c.HasUserInfo() {
		return fmt.Errorf("user information not available - run 'user info' first")
	}

	userName := strings.ToLower(c.UserName)
	userEmail := strings.ToLower(c.UserEmail)

	c.Filters.UserNameBlacklist = removeFromSlice(c.Filters.UserNameBlacklist, userName)
	c.Filters.UserEmailBlacklist = removeFromSlice(c.Filters.UserEmailBlacklist, userEmail)

	return nil
}

// Convenience methods for current sprint filtering
// FilterToCurrentSprint sets filters to show only tasks from the current sprint
func (c *Config) FilterToCurrentSprint() error {
	if c.SprintID == "" {
		return fmt.Errorf("current sprint not set - run 'config set-sprint-id <sprint-id>' first")
	}

	// Clear existing sprint filters
	c.Filters.SprintWhitelist = []string{}
	c.Filters.SprintBlacklist = []string{}

	// Add current sprint to whitelist
	c.Filters.SprintWhitelist = append(c.Filters.SprintWhitelist, strings.ToLower(c.SprintID))

	return nil
}

// AddCurrentSprintToWhitelist adds the current sprint to the sprint whitelist
func (c *Config) AddCurrentSprintToWhitelist() error {
	if c.SprintID == "" {
		return fmt.Errorf("current sprint not set - run 'config set-sprint-id <sprint-id>' first")
	}

	sprintID := strings.ToLower(c.SprintID)

	if !slices.Contains(c.Filters.SprintWhitelist, sprintID) {
		c.Filters.SprintWhitelist = append(c.Filters.SprintWhitelist, sprintID)
	}

	return nil
}

// RemoveCurrentSprintFromWhitelist removes the current sprint from the sprint whitelist
func (c *Config) RemoveCurrentSprintFromWhitelist() error {
	if c.SprintID == "" {
		return fmt.Errorf("current sprint not set - run 'config set-sprint-id <sprint-id>' first")
	}

	sprintID := strings.ToLower(c.SprintID)
	c.Filters.SprintWhitelist = removeFromSlice(c.Filters.SprintWhitelist, sprintID)

	return nil
}

// AddCurrentSprintToBlacklist adds the current sprint to the sprint blacklist
func (c *Config) AddCurrentSprintToBlacklist() error {
	if c.SprintID == "" {
		return fmt.Errorf("current sprint not set - run 'config set-sprint-id <sprint-id>' first")
	}

	sprintID := strings.ToLower(c.SprintID)

	if !slices.Contains(c.Filters.SprintBlacklist, sprintID) {
		c.Filters.SprintBlacklist = append(c.Filters.SprintBlacklist, sprintID)
	}

	return nil
}

// RemoveCurrentSprintFromBlacklist removes the current sprint from the sprint blacklist
func (c *Config) RemoveCurrentSprintFromBlacklist() error {
	if c.SprintID == "" {
		return fmt.Errorf("current sprint not set - run 'config set-sprint-id <sprint-id>' first")
	}

	sprintID := strings.ToLower(c.SprintID)
	c.Filters.SprintBlacklist = removeFromSlice(c.Filters.SprintBlacklist, sprintID)

	return nil
}

func removeFromSlice(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
