package main

import (
	"fmt"
	"os"
	"strings"

	"emkai/go-cli-gui/monday"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	// Check if we have subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "config":
			handleConfigCommand()
			return
		case "tasks":
			handleTasksCommand()
			return
		case "help":
			showHelp()
			return
		}
	}

	// Default TUI mode
	showHelp()
}

func handleConfigCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: monday-cli config <subcommand>")
		fmt.Println("Subcommands:")
		fmt.Println("  set-api-key <api-key>")
		fmt.Println("  set-owner-email <email>")
		fmt.Println("  set-board-id <board-id>")
		fmt.Println("  show")
		os.Exit(1)
	}

	subcommand := os.Args[2]

	switch subcommand {
	case "set-api-key":
		if len(os.Args) < 4 {
			fmt.Println("Usage: monday-cli config set-api-key <api-key>")
			os.Exit(1)
		}
		apiKey := os.Args[3]
		if err := setAPIKey(apiKey); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("API key set successfully!")
	case "set-owner-email":
		if len(os.Args) < 4 {
			fmt.Println("Usage: monday-cli config set-owner-email <email>")
			os.Exit(1)
		}
		ownerEmail := os.Args[3]
		if err := setOwnerEmail(ownerEmail); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Owner email set successfully!")
	case "set-board-id":
		if len(os.Args) < 4 {
			fmt.Println("Usage: monday-cli config set-board-id <board-id>")
			os.Exit(1)
		}
		boardID := os.Args[3]
		if err := setBoardID(boardID); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Board ID set successfully!")

	case "show":
		if err := showConfig(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func handleTasksCommand() {
	client, err := getClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get configured board ID and owner email
	config, err := loadConfig("")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	boardID := config.GetBoardID()
	ownerEmail := config.GetOwnerEmail()

	if boardID == "" {
		fmt.Println("❌ Board ID not configured. Run: monday-cli config set-board-id <board-id>")
		os.Exit(1)
	}

	if ownerEmail == "" {
		fmt.Println("❌ Owner email not configured. Run: monday-cli config set-owner-email <email>")
		os.Exit(1)
	}

	fmt.Printf("🔍 Fetching tasks assigned to %s...\n", ownerEmail)
	fmt.Println("=" + strings.Repeat("=", 50))

	// Get board info first
	boardService := monday.NewBoardService(client)
	board, err := boardService.GetBoardByID(boardID)
	if err != nil {
		fmt.Printf("❌ Error getting board: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📋 Board: %s (ID: %s)\n", board.Name, board.ID)
	fmt.Println("-" + strings.Repeat("-", len(board.Name)+20))

	// Get items from the board using configured owner email
	items, err := client.GetBoardItemsByOwner(boardID, ownerEmail)
	if err != nil {
		fmt.Printf("❌ Error getting tasks: %v\n", err)
		os.Exit(1)
	}

	if len(items) == 0 {
		fmt.Printf("👤 No tasks assigned to %s in %s\n", ownerEmail, board.Name)
		return
	}

	// Filter out completed tasks
	var activeItems []monday.Item
	for _, item := range items {
		skipTask := false
		for _, cv := range item.ColumnValues {
			if strings.Contains(strings.ToLower(cv.ID), "status") && cv.Text != "" {
				status := strings.ToLower(cv.Text)
				if strings.Contains(status, "done") || strings.Contains(status, "completed") {
					skipTask = true
					break
				}
			}
		}
		if !skipTask {
			activeItems = append(activeItems, item)
		}
	}

	if len(activeItems) == 0 {
		fmt.Printf("🎉 No active tasks assigned to %s in %s\n", ownerEmail, board.Name)
		return
	}

	fmt.Printf("👤 Found %d tasks assigned to %s:\n\n", len(items), ownerEmail)

	fmt.Println("Type [Status Priority] Task Name")
	for _, item := range activeItems {
		// Extract status, priority, and type
		status := "📋"
		priority := "⚪"
		taskType := "📋"

		for _, cv := range item.ColumnValues {
			if strings.Contains(strings.ToLower(cv.ID), "status") && cv.Text != "" {
				status = getStatusIcon(cv.Text)
			} else if strings.Contains(strings.ToLower(cv.ID), "priority") && cv.Text != "" {
				priority = getPriorityIcon(cv.Text)
			} else if strings.Contains(strings.ToLower(cv.ID), "type") && cv.Text != "" {
				taskType = getTypeIcon(cv.Text)
			}
		}

		// Format: [icon] Task Name | Priority Type
		fmt.Printf("%s [%s %s] %s\n", taskType, status, priority, item.Name)
	}

	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("📊 Active tasks: %d\n", len(activeItems))
}

func showHelp() {
	fmt.Println("Monday.com CLI - Task Management Tool")
	fmt.Println("")
	fmt.Println("Usage: monday-cli [command]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  tasks          Show your assigned tasks")
	fmt.Println("  config         Manage configuration")
	fmt.Println("  help           Show this help")
	fmt.Println("")
	fmt.Println("Configuration Commands:")
	fmt.Println("  monday-cli config set-api-key <api-key>")
	fmt.Println("  monday-cli config set-owner-email <email>")
	fmt.Println("  monday-cli config set-board-id <board-id>")
	fmt.Println("  monday-cli config show")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  monday-cli tasks")
	fmt.Println("  monday-cli config set-owner-email your.email@company.com")
	fmt.Println("  monday-cli config set-board-id 0123456789")
}

// Helper functions for configuration
func loadConfig(configPath string) (*monday.Config, error) {
	if configPath == "" {
		configPath = monday.GetConfigPath()
	}
	return monday.LoadConfig(configPath)
}

func getClient() (*monday.Client, error) {
	config, err := loadConfig("")
	if err != nil {
		return nil, err
	}

	if !config.IsConfigured() {
		return nil, fmt.Errorf("configuration error: API key, owner email, and board ID are required")
	}

	return monday.NewClient(config.GetAPIKey(), config.Timeout), nil
}

func setAPIKey(apiKey string) error {
	config, err := loadConfig("")
	if err != nil {
		return err
	}

	config.SetAPIKey(apiKey)

	configPath := monday.GetConfigPath()
	if err := config.Save(configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

func setOwnerEmail(ownerEmail string) error {
	config, err := loadConfig("")
	if err != nil {
		return err
	}

	config.SetOwnerEmail(ownerEmail)

	configPath := monday.GetConfigPath()
	if err := config.Save(configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

func setBoardID(boardID string) error {
	config, err := loadConfig("")
	if err != nil {
		return err
	}

	config.SetBoardID(boardID)

	configPath := monday.GetConfigPath()
	if err := config.Save(configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

func getAPIKey() error {
	config, err := loadConfig("")
	if err != nil {
		return err
	}

	apiKey := config.GetAPIKey()
	if apiKey == "" {
		fmt.Println("No API key configured")
		return nil
	}

	masked := maskAPIKey(apiKey)
	fmt.Printf("API key: %s\n", masked)
	return nil
}

func showConfig() error {
	config, err := loadConfig("")
	if err != nil {
		return err
	}

	fmt.Println("Current configuration:")
	fmt.Printf("  API Key: %s\n", maskAPIKey(config.GetAPIKey()))
	fmt.Printf("  Owner Email: %s\n", config.GetOwnerEmail())
	fmt.Printf("  Board ID: %s\n", config.GetBoardID())
	fmt.Printf("  Base URL: %s\n", config.BaseURL)
	fmt.Printf("  Timeout: %d seconds\n", config.Timeout)
	return nil
}

func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 4 {
		return strings.Repeat("*", len(apiKey))
	}
	return strings.Repeat("*", len(apiKey)-4) + apiKey[len(apiKey)-4:]
}

// Icon helper functions
func getStatusIcon(status string) string {
	status = strings.ToLower(status)
	switch {
	case strings.Contains(status, "done") || strings.Contains(status, "completed"):
		return "✅"
	case strings.Contains(status, "progress") || strings.Contains(status, "in progress"):
		return "🔄"
	case strings.Contains(status, "stuck") || strings.Contains(status, "blocked"):
		return "🚫"
	case strings.Contains(status, "review"):
		return "👀"
	default:
		return "📋"
	}
}

func getPriorityIcon(priority string) string {
	priority = strings.ToLower(priority)
	switch {
	case strings.Contains(priority, "critical"):
		return "🔴"
	case strings.Contains(priority, "high"):
		return "🟠"
	case strings.Contains(priority, "medium"):
		return "🟡"
	case strings.Contains(priority, "low"):
		return "🟢"
	default:
		return "⚪"
	}
}

func getTypeIcon(taskType string) string {
	taskType = strings.ToLower(taskType)
	switch {
	case strings.Contains(taskType, "bug"):
		return "🐛"
	case strings.Contains(taskType, "feature"):
		return "✨"
	case strings.Contains(taskType, "test"):
		return "🧪"
	case strings.Contains(taskType, "security"):
		return "🔒"
	case strings.Contains(taskType, "improvement"):
		return "📈"
	case strings.Contains(taskType, "documentation"):
		return "📝"
	default:
		return "📝"
	}
}

type menuModel struct {
	choices []string
	cursor  int
}

func (m *menuModel) Init() tea.Cmd {
	return nil
}

func (m *menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			choice := m.choices[m.cursor]
			switch choice {
			case "Connect":
				// TODO: Implement connection logic
				return m, tea.Quit
			case "Settings":
				// TODO: Implement settings logic
				return m, tea.Quit
			case "About":
				// TODO: Implement about logic
				return m, tea.Quit
			case "Exit":
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m *menuModel) View() string {
	s := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Render("Monday.com CLI")
	s += "\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\nPress q to quit.\n"

	return s
}
