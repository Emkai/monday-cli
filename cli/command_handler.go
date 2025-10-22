package cli

import (
	"emkai/go-cli-gui/monday"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func (c *CLI) HandleCommand() {
	switch c.command.Command {
	case "help", "h":
		c.ShowHelp()
	case "config", "cfg":
		c.HandleConfigCommand()
	case "tasks", "ts":
		c.HandleTasksCommand()
	case "task", "t":
		c.HandleTaskCommand()
	default:
		c.ShowHelp()
	}
}

func (c *CLI) ShowHelp() {
	fmt.Println("Monday CLI - Task Management Tool")
	fmt.Println("")
	fmt.Println("Usage: <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  tasks (ts)     Show your assigned tasks")
	fmt.Println("  task (t)       Specific task operations")
	fmt.Println("  config (cfg)   Manage configuration")
	fmt.Println("  help (h)       Show this help")
	fmt.Println("")
}

func (c *CLI) HandleConfigCommand() {
	if len(c.command.Args) == 0 {
		c.HelpConfigCommand()
		return
	}
	subcommand := c.command.Args[0]
	switch subcommand {
	case "set-api-key", "key":
		if len(c.command.Args) < 2 {
			fmt.Println("Usage: monday-cli config set-api-key <api-key>")
			return
		}
		c.config.SetAPIKey(c.command.Args[1])
		c.config.Save(monday.GetConfigPath())
		return
	case "set-owner-email", "email":
		if len(c.command.Args) < 2 {
			fmt.Println("Usage: monday-cli config set-owner-email <email>")
			return
		}
		c.config.SetOwnerEmail(c.command.Args[1])
		c.config.Save(monday.GetConfigPath())
		return
	case "set-board-id", "board":
		if len(c.command.Args) < 2 {
			fmt.Println("Usage: monday-cli config set-board-id <board-id>")
			return
		}
		c.config.SetBoardID(c.command.Args[1])
		c.config.Save(monday.GetConfigPath())
		return
	case "set-sprint-id", "sprint":
		if len(c.command.Args) < 2 {
			fmt.Println("Usage: monday-cli config set-sprint-id <sprint-id>")
			return
		}
		c.config.SetSprintID(c.command.Args[1])
		c.config.Save(monday.GetConfigPath())
		return
	case "show", "s":
		fmt.Println("API Key:", maskAPIKey(c.config.GetAPIKey()))
		fmt.Println("Owner Email:", c.config.GetOwnerEmail())
		fmt.Println("Board ID:", c.config.GetBoardID())
		fmt.Println("Sprint ID:", c.config.GetSprintID())
		return
	default:
		c.HelpConfigCommand()
		return
	}
}

func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 4 {
		return strings.Repeat("*", len(apiKey))
	}
	return strings.Repeat("*", len(apiKey)-4) + apiKey[len(apiKey)-4:]
}

func (c *CLI) HelpConfigCommand() {
	fmt.Println("Config Commands:")
	fmt.Println("  config set-api-key (key) <api-key>")
	fmt.Println("  config set-owner-email (email) <email>")
	fmt.Println("  config set-board-id (board) <board-id>")
	fmt.Println("  config set-sprint-id (sprint) <sprint-id>")
	fmt.Println("  config show (s)")
}

func (c *CLI) HandleTasksCommand() {
	if len(c.command.Args) == 0 {
		c.HelpTasksCommand()
		return
	}
	subcommand := c.command.Args[0]
	switch subcommand {
	case "list", "ls":
		dataStore := monday.NewDataStore()
		tasks, timestamp, _ := dataStore.GetCachedTasks(c.config.GetBoardID(), c.config.GetOwnerEmail())
		fmt.Println("Tasks cached at: " + timestamp.Format(time.RFC3339))
		// Convert map to slice for consistent ordering
		var sortedTasks []monday.Item
		for _, task := range tasks {
			sortedTasks = append(sortedTasks, task)
		}
		c.PrintItems(tasks, sortedTasks)
		return
	case "fetch", "f":
		client := monday.NewClient(c.config.GetAPIKey(), c.config.Timeout)

		boardID := c.config.GetBoardID()
		ownerEmail := c.config.GetOwnerEmail()

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

		itemsMap := make(map[int]monday.Item)
		for i, item := range items {
			itemsMap[i] = item
		}

		dataStore := monday.NewDataStore()
		dataStore.ClearCache(boardID, ownerEmail)
		dataStore.StoreTaskRequest(boardID, ownerEmail, itemsMap)
		c.PrintItems(itemsMap, items)
		return
	default:
		c.HelpTasksCommand()
		return
	}
}

func (c *CLI) HelpTasksCommand() {
	fmt.Println("Tasks Commands:")
	fmt.Println("  tasks list (ls)      Show your assigned tasks")
	fmt.Println("  tasks fetch (f)      Fetch your assigned tasks")
}

func (c *CLI) HandleTaskCommand() {
	var err error
	if len(c.command.Args) == 0 {
		c.HelpTaskCommand()
		return
	}
	subcommand := c.command.Args[0]
	switch subcommand {
	case "show", "s":
		if len(c.command.Args) < 2 {
			fmt.Println("Usage: monday-cli task show <task-id>")
			return
		}
		taskID, err := strconv.Atoi(c.command.Args[1])
		if err != nil {
			fmt.Printf("❌ Invalid task ID: %v\n", err)
			os.Exit(1)
		}
		dataStore := monday.NewDataStore()
		task, timestamp, ok := dataStore.GetCachedTask(c.config.GetBoardID(), c.config.GetOwnerEmail(), taskID)
		if !ok {
			fmt.Printf("❌ Task %d not found\n", taskID)
			os.Exit(1)
		}
		fmt.Println("Task cached at: " + timestamp.Format(time.RFC3339))
		c.PrintTask(taskID, task)
		return
	case "create", "c":
		if len(c.command.Args) < 2 {
			fmt.Println("Usage: monday-cli task create <task-name>")
			return
		}

		taskName := strings.Join(c.command.Args[1:], " ")

		fmt.Printf("Creating task: %s\n", taskName)
		client := monday.NewClient(c.config.GetAPIKey(), c.config.Timeout)
		err = client.CreateTask(c.config.GetBoardID(), c.config.GetOwnerEmail(), taskName)
		if err != nil {
			fmt.Printf("❌ Error creating task: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Task %s created\n", taskName)
		return
	case "edit", "e":
		if len(c.command.Args) < 2 {
			fmt.Println("Usage: monday-cli task edit <task-id> <new-status>")
			fmt.Println("New status: done(d), in progress(p), stuck(s), waiting review(r), ready for testing(t), removed(rm)")
			return
		}
		taskID, err := strconv.Atoi(c.command.Args[1])
		if err != nil {
			fmt.Printf("❌ Invalid task ID: %v\n", err)
			os.Exit(1)
		}
		newStatus := getStatusValue(c.command.Args[2])
		if newStatus == "" {
			fmt.Printf("❌ Invalid status: %s\n", c.command.Args[2])
			os.Exit(1)
		}
		dataStore := monday.NewDataStore()
		task, _, ok := dataStore.GetCachedTask(c.config.GetBoardID(), c.config.GetOwnerEmail(), taskID)
		if !ok {
			fmt.Printf("❌ Task %d not found\n", taskID)
			os.Exit(1)
		}
		client := monday.NewClient(c.config.GetAPIKey(), c.config.Timeout)
		err = client.UpdateTaskStatus(c.config.GetBoardID(), c.config.GetOwnerEmail(), task, newStatus)
		if err != nil {
			fmt.Printf("❌ Error updating task status: %v\n", err)
			os.Exit(1)
		}
		dataStore.UpdateCachedTask(c.config.GetBoardID(), c.config.GetOwnerEmail(), taskID, task)
		fmt.Printf("✅ Task %s status updated to %s\n", taskID, newStatus)
		return
	default:
		c.HelpTaskCommand()
		return
	}
}

func getStatusValue(status string) string {
	switch status {
	case "done", "d":
		return "Done"
	case "in progress", "p":
		return "In Progress"
	case "stuck", "s":
		return "Stuck"
	case "waiting review", "r":
		return "Waiting for review"
	case "ready for testing", "t":
		return "Ready for testing"
	case "removed", "rm":
		return "Removed"
	default:
		return ""
	}
}

func (c *CLI) HelpTaskCommand() {
	fmt.Println("Task Commands:")
	fmt.Println("  task show (s) <task-id> Show a specific task")
	fmt.Println("  task edit (e) <task-id> <new-status> Edit a specific task")
}
