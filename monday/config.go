package monday

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents Monday.com configuration
type Config struct {
	APIKey     string `json:"api_key"`
	BaseURL    string `json:"base_url"`
	Timeout    int    `json:"timeout_seconds"`
	OwnerEmail string `json:"owner_email"`
	BoardID    string `json:"board_id"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    "https://api.monday.com/v2",
		Timeout:    30,
		OwnerEmail: "",
		BoardID:    "",
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
	return c.APIKey != "" && c.OwnerEmail != "" && c.BoardID != ""
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./monday-config.json"
	}
	return filepath.Join(homeDir, ".config", "monday-cli", "config.json")
}
