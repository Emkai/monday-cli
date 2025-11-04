package cli

import (
	"fmt"
	"monday-cli/monday"
	"os"
	"strconv"
	"strings"
	"time"
)

type CommandString string

const (
	CSHelp CommandString = "help"
	CSConfig CommandString = "config"
	CSTasks CommandString = "tasks"
	CSTask CommandString = "task"
	CSUser CommandString = "user"
)

func (cs *CommandString) ToString() string {
	return string(*cs)
} 

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
	case "user", "u":
		c.HandleUserCommand()
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
	fmt.Println("  user (u)       User information")
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
		fmt.Println("API Key set successfully")

		// Automatically fetch user info after setting API key
		fmt.Println("🔍 Fetching user information...")
		client := monday.NewClient(c.config.GetAPIKey(), c.config.Timeout)
		user, err := client.GetUserInfo()
		if err != nil {
			fmt.Printf("❌ Error getting user info: %v\n", err)
			fmt.Println("You can run 'user info' later to fetch user information")
			return
		}

		// Save user info to config
		c.config.SetUserInfo(user)
		c.config.Save(monday.GetConfigPath())
		fmt.Println("💾 User information saved to configuration")
		fmt.Println("")

		// Show user info
		PrintUserInfo(user)
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
	case "set-sprint-board-id", "sprint-board":
		if len(c.command.Args) < 2 {
			fmt.Println("Usage: monday-cli config set-sprint-board-id <sprint-board-id>")
			return
		}
		c.config.SetSprintBoardID(c.command.Args[1])
		c.config.Save(monday.GetConfigPath())
		return
	case "show", "s":
		fmt.Println("API Key:", maskAPIKey(c.config.GetAPIKey()))
		if c.config.HasUserInfo() {
			user := c.config.GetUserInfo()
			fmt.Println("User ID:", user.ID)
			fmt.Println("User Name:", user.Name)
			fmt.Println("User Email:", user.Email)
			if user.Title != "" {
				fmt.Println("User Title:", user.Title)
			}
		} else {
			fmt.Println("User Info: Not configured (run 'user info' to fetch)")
		}
		fmt.Println("Board ID:", c.config.GetBoardID())
		fmt.Println("Sprint ID:", c.config.GetSprintID())
		fmt.Println("Sprint Board ID:", c.config.GetSprintBoardID())
		return
	case "add-filter", "addf":
		c.HandleAddFilterCommand()
		return
	case "remove-filter", "remf":
		c.HandleRemoveFilterCommand()
		return
	case "clear-filter", "clrf":
		c.HandleClearFilterCommand()
		return
	case "list-filters", "listf":
		c.HandleListFiltersCommand()
		return
	case "clear-all-filters", "clearallf":
		c.HandleClearAllFiltersCommand()
		return
	case "filter-to-me", "me":
		c.HandleFilterToMeCommand()
		return
	case "add-me", "addme":
		c.HandleAddMeCommand()
		return
	case "remove-me", "removeme":
		c.HandleRemoveMeCommand()
		return
	case "filter-to-sprint", "sprint-filter":
		c.HandleFilterToSprintCommand()
		return
	case "add-sprint", "add-s":
		c.HandleAddSprintCommand()
		return
	case "remove-sprint", "rm-s":
		c.HandleRemoveSprintCommand()
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
	fmt.Println("  config set-board-id (board) <board-id>")
	fmt.Println("  config set-sprint-id (sprint) <sprint-id>")
	fmt.Println("  config set-sprint-board-id (sprint-board) <sprint-board-id>")
	fmt.Println("  config show (s)")
	fmt.Println("")
	fmt.Println("Filter Commands:")
	fmt.Println("  config add-filter (addf) <type> <whitelist|blacklist> <value>")
	fmt.Println("  config remove-filter (remf) <type> <whitelist|blacklist> <value>")
	fmt.Println("  config clear-filter (clrf) <type> <whitelist|blacklist>")
	fmt.Println("  config list-filters (listf)")
	fmt.Println("  config clear-all-filters (clearallf)")
	fmt.Println("")
	fmt.Println("User Filter Commands:")
	fmt.Println("  config filter-to-me (me)           Show only tasks assigned to you")
	fmt.Println("  config add-me (addme)              Add yourself to user whitelist")
	fmt.Println("  config remove-me (removeme)        Remove yourself from user whitelist")
	fmt.Println("  config filter-to-sprint (sprint-filter)  Filter to show only current sprint tasks")
	fmt.Println("  config add-sprint (add-s)          Add current sprint to whitelist")
	fmt.Println("  config remove-sprint (rm-s)        Remove current sprint from whitelist")
	fmt.Println("")
	fmt.Println("Filter Types: status, priority, type, sprint, user_name, user_email")
	fmt.Println("Examples:")
	fmt.Println("  config add-filter status whitelist 'in progress'")
	fmt.Println("  config add-filter priority blacklist 'low'")
	fmt.Println("  config remove-filter type whitelist 'bug'")
	fmt.Println("  config filter-to-me")
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
		tasks, timestamp, _ := dataStore.GetCachedTasks(c.config.GetBoardID())
		fmt.Println("Tasks cached at: " + timestamp.Format(time.RFC3339))
		c.PrintItems(tasks)
		return
	case "fetch", "f":
		client := monday.NewClient(c.config.GetAPIKey(), c.config.Timeout)

		boardID := c.config.GetBoardID()

		fmt.Printf("🔍 Fetching tasks in board %s...\n", boardID)
		fmt.Println("=" + strings.Repeat("=", 50))

		boardService := monday.NewBoardService(client)
		board, err := boardService.GetBoardByID(boardID)
		if err != nil {
			fmt.Printf("❌ Error getting board: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("📋 Board: %s (ID: %s)\n", board.Name, board.ID)
		fmt.Println("-" + strings.Repeat("-", len(board.Name)+20))

		items, rawItems, err := client.GetBoardItems(boardID)
		if err != nil {
			fmt.Printf("❌ Error getting tasks: %v\n", err)
			os.Exit(1)
		}

		if len(items) == 0 {
			fmt.Printf("👤 No tasks in %s\n", board.Name)
			return
		}

		// Fetch board users
		fmt.Printf("👥 Fetching board users...\n")
		users, err := client.GetBoardUsers(boardID)
		if err != nil {
			fmt.Printf("⚠️  Warning: Could not fetch board users: %v\n", err)
			users = []monday.User{} // Continue without users
		} else {
			fmt.Printf("👥 Found %d users on board\n", len(users))
		}

		// Fetch board sprints from sprint board
		sprintBoardID := c.config.GetSprintBoardID()
		var sprints []monday.Sprint
		if sprintBoardID != "" {
			fmt.Printf("🏃 Fetching sprints from sprint board...\n")
			sprints, err = client.GetBoardSprints(sprintBoardID)
			if err != nil {
				fmt.Printf("⚠️  Warning: Could not fetch board sprints: %v\n", err)
				sprints = []monday.Sprint{} // Continue without sprints
			} else {
				fmt.Printf("🏃 Found %d sprints on sprint board\n", len(sprints))
			}
		} else {
			fmt.Printf("⚠️  Warning: No sprint board ID configured, skipping sprint fetch\n")
			sprints = []monday.Sprint{}
		}

		dataStore := monday.NewDataStore()
		dataStore.ClearCache(boardID)
		dataStore.StoreTasksRequest(boardID, items, rawItems)
		dataStore.StoreBoardUsers(boardID, users)
		if sprintBoardID != "" {
			dataStore.StoreBoardSprints(sprintBoardID, sprints)
		}
		cacheItems, _, _ := dataStore.GetCachedTasks(boardID)
		c.PrintItems(cacheItems)
		return
	case "users", "u":
		c.HandleListBoardUsersCommand()
		return
	case "sprints", "s":
		c.HandleListBoardSprintsCommand()
		return
	case "sprint", "sp":
		c.HandleSprintCommand()
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
	fmt.Println("  tasks users (u)      Show board users")
	fmt.Println("  tasks sprints (s)    Show board sprints")
	fmt.Println("  tasks sprint (sp)    Sprint-specific commands")
}

func (c *CLI) HandleTaskCommand() {
	if len(c.command.Args) == 0 {
		c.HelpTaskCommand()
		return
	}
	subcommand := c.command.Args[0]
	switch subcommand {
	case "show", "s":
		if len(c.command.Args) < 2 {
			fmt.Println("Usage: monday-cli task show <task-index>")
			return
		}
		localId, err := strconv.Atoi(c.command.Args[1])
		if err != nil {
			fmt.Printf("❌ Invalid task local ID: %v\n", err)
			os.Exit(1)
		}
		dataStore := monday.NewDataStore()
		task, timestamp, ok := dataStore.GetCachedTaskByLocalId(c.config.GetBoardID(), localId)
		if !ok {
			fmt.Printf("❌ Task %d not found\n", localId)
			os.Exit(1)
		}
		fmt.Println("Task cached at: " + timestamp.Format(time.RFC3339))
		PrintTask(task)
		return
	case "create", "c":
		if len(c.command.Args) < 2 {
			fmt.Println("Usage: monday-cli task create <task-name> [flags]")
			fmt.Println("Flags:")
			fmt.Println("  -status, -s <status>     Set task status (done/d, in progress/p, stuck/s, etc.)")
			fmt.Println("  -priority, -p <priority> Set task priority (critical/c, high/h, medium/m, low/l)")
			fmt.Println("  -type, -t <type>         Set task type (bug/b, feature/f, test/t, security/s, improvement/i)")
			return
		}

		taskName := c.command.Args[1]

		// Parse flags
		var status, priority, taskType string
		for _, flag := range c.command.Flags {
			switch flag.Flag {
			case "-status", "-s":
				status = getStatusValue(flag.Value)
				if status == "" {
					fmt.Printf("❌ Invalid status: %s\n", flag.Value)
					fmt.Println("Valid status values: done(d), in progress(p), stuck(s), waiting review(r), ready for testing(t), removed(rm)")
					os.Exit(1)
				}
			case "-priority", "-p":
				priority = getPriorityValue(flag.Value)
				if priority == "" {
					fmt.Printf("❌ Invalid priority: %s\n", flag.Value)
					fmt.Println("Valid priority values: critical(c), high(h), medium(m), low(l)")
					os.Exit(1)
				}
			case "-type", "-t":
				taskType = getTypeValue(flag.Value)
				if taskType == "" {
					fmt.Printf("❌ Invalid type: %s\n", flag.Value)
					fmt.Println("Valid type values: bug(b), feature(f), test(t), security(s), quality(q)")
					os.Exit(1)
				}
			}
		}

		fmt.Printf("Creating task: %s\n", taskName)
		if status != "" {
			fmt.Printf("  Status: %s\n", status)
		}
		if priority != "" {
			fmt.Printf("  Priority: %s\n", priority)
		}
		if taskType != "" {
			fmt.Printf("  Type: %s\n", taskType)
		}

		client := monday.NewClient(c.config.GetAPIKey(), c.config.Timeout)
		localId, task, err := client.CreateTask(c.config.GetBoardID(), c.config.GetUserInfo().ID, taskName, status, priority, taskType)
		if err != nil {
			fmt.Printf("❌ Error creating task: %v\n", err)
			return
		}
		fmt.Printf("✅ Task %s created with ID %d\n", task.Name, localId)
		PrintTask(*task)
		return
	case "edit", "e":
		if len(c.command.Args) < 2 {
			fmt.Println("Usage: monday-cli task edit <task-index> [flags]")
			fmt.Println("Flags:")
			fmt.Println("  -status, -s <status>     Set task status (done/d, in progress/p, stuck/s, etc.)")
			fmt.Println("  -priority, -p <priority> Set task priority (critical/c, high/h, medium/m, low/l)")
			fmt.Println("  -type, -t <type>         Set task type (bug/b, feature/f, test/t, security/s, improvement/i)")
			return
		}
		taskIndex, err := strconv.Atoi(c.command.Args[1])
		if err != nil {
			fmt.Printf("❌ Invalid task index: %v\n", err)
			os.Exit(1)
		}

		// Parse flags
		var status, priority, taskType string
		for _, flag := range c.command.Flags {
			switch flag.Flag {
			case "-status", "-s":
				status = getStatusValue(flag.Value)
				if status == "" {
					fmt.Printf("❌ Invalid status: %s\n", flag.Value)
					fmt.Println("Valid status values: done(d), in progress(p), stuck(s), waiting review(r), ready for testing(t), removed(rm)")
					os.Exit(1)
				}
			case "-priority", "-p":
				priority = getPriorityValue(flag.Value)
				if priority == "" {
					fmt.Printf("❌ Invalid priority: %s\n", flag.Value)
					fmt.Println("Valid priority values: critical(c), high(h), medium(m), low(l)")
					os.Exit(1)
				}
			case "-type", "-t":
				taskType = getTypeValue(flag.Value)
				if taskType == "" {
					fmt.Printf("❌ Invalid type: %s\n", flag.Value)
					fmt.Println("Valid type values: bug(b), feature(f), test(t), security(s), quality(q)")
					os.Exit(1)
				}
			}
		}

		// Check if at least one field is being updated
		if status == "" && priority == "" && taskType == "" {
			fmt.Println("❌ No fields to update. Please specify at least one flag (-status, -priority, or -type)")
			return
		}

		dataStore := monday.NewDataStore()
		task, _, ok := dataStore.GetCachedTaskByLocalId(c.config.GetBoardID(), taskIndex)
		if !ok {
			fmt.Printf("❌ Task %d not found\n", taskIndex)
			os.Exit(1)
		}

		fmt.Printf("Updating task %d: %s\n", taskIndex, task.Name)
		if status != "" {
			fmt.Printf("  Status: %s\n", status)
		}
		if priority != "" {
			fmt.Printf("  Priority: %s\n", priority)
		}
		if taskType != "" {
			fmt.Printf("  Type: %s\n", taskType)
		}

		client := monday.NewClient(c.config.GetAPIKey(), c.config.Timeout)
		updatedTask, err := client.UpdateTask(c.config.GetBoardID(), c.config.GetUserEmail(), task, status, priority, taskType)
		if err != nil {
			fmt.Printf("❌ Error updating task: %v\n", err)
			os.Exit(1)
		}
		dataStore.UpdateCachedTaskByLocalId(c.config.GetBoardID(), taskIndex, *updatedTask)
		fmt.Printf("✅ Task %d updated successfully\n", taskIndex)
		PrintTask(*updatedTask)
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
	case "waiting for review", "r":
		return "Waiting for review"
	case "ready for testing", "t":
		return "Ready for testing"
	case "removed", "rm":
		return "Removed"
	default:
		return ""
	}
}

func getPriorityValue(priority string) string {
	switch priority {
	case "critical", "c":
		return "Critical"
	case "high", "h":
		return "High"
	case "medium", "m":
		return "Medium"
	case "low", "l":
		return "Low"
	default:
		return ""
	}
}

func getTypeValue(taskType string) string {
	switch taskType {
	case "bug", "b":
		return "Bug"
	case "feature", "f":
		return "Feature"
	case "test", "t":
		return "Test"
	case "security", "s":
		return "Security"
	case "quality", "q":
		return "Quality"
	default:
		return ""
	}
}

func (c *CLI) HelpTaskCommand() {
	fmt.Println("Task Commands:")
	fmt.Println("  task show (s) <task-index> Show a specific task")
	fmt.Println("  task create (c) <task-name> [flags] Create a new task")
	fmt.Println("    Flags:")
	fmt.Println("      -status, -s <status>     Set task status (done/d, in progress/p, stuck/s, etc.)")
	fmt.Println("      -priority, -p <priority> Set task priority (critical/c, high/h, medium/m, low/l)")
	fmt.Println("      -type, -t <type>         Set task type (bug/b, feature/f, test/t, security/s, improvement/i)")
	fmt.Println("  task edit (e) <task-index> [flags] Edit a specific task")
	fmt.Println("    Flags:")
	fmt.Println("      -status, -s <status>     Set task status (done/d, in progress/p, stuck/s, etc.)")
	fmt.Println("      -priority, -p <priority> Set task priority (critical/c, high/h, medium/m, low/l)")
	fmt.Println("      -type, -t <type>         Set task type (bug/b, feature/f, test/t, security/s, improvement/i)")
}

func (c *CLI) HandleUserCommand() {
	if len(c.command.Args) == 0 {
		c.HelpUserCommand()
		return
	}
	subcommand := c.command.Args[0]
	switch subcommand {
	case "info", "i":
		client := monday.NewClient(c.config.GetAPIKey(), c.config.Timeout)

		fmt.Println("🔍 Fetching user information...")
		fmt.Println("=" + strings.Repeat("=", 50))

		user, err := client.GetUserInfo()
		if err != nil {
			fmt.Printf("❌ Error getting user info: %v\n", err)
			os.Exit(1)
		}

		// Save user info to config
		c.config.SetUserInfo(user)
		c.config.Save(monday.GetConfigPath())
		fmt.Println("💾 User information saved to configuration")
		fmt.Println("")

		PrintUserInfo(user)
		return
	default:
		c.HelpUserCommand()
		return
	}
}

func (c *CLI) HelpUserCommand() {
	fmt.Println("User Commands:")
	fmt.Println("  user info (i)   Show current user information")
}

// Filter command handlers
func (c *CLI) HandleAddFilterCommand() {
	if len(c.command.Args) < 4 {
		fmt.Println("Usage: monday-cli config add-filter <type> <whitelist|blacklist> <value>")
		fmt.Println("Types: status, priority, type, sprint, user_name, user_email")
		fmt.Println("Example: monday-cli config add-filter status whitelist 'in progress'")
		return
	}

	filterType := monday.FilterType(c.command.Args[1])
	listType := monday.FilterListType(c.command.Args[2])
	value := c.command.Args[3]

	// Validate filter type
	validTypes := []monday.FilterType{
		monday.FilterStatus, monday.FilterPriority, monday.FilterTaskType,
		monday.FilterSprint, monday.FilterUserName, monday.FilterUserEmail,
	}
	validType := false
	for _, vt := range validTypes {
		if filterType == vt {
			validType = true
			break
		}
	}
	if !validType {
		fmt.Printf("❌ Invalid filter type: %s\n", filterType)
		fmt.Println("Valid types: status, priority, type, sprint, user_name, user_email")
		return
	}

	// Validate list type
	if listType != monday.Whitelist && listType != monday.Blacklist {
		fmt.Printf("❌ Invalid list type: %s\n", listType)
		fmt.Println("Valid types: whitelist, blacklist")
		return
	}

	err := c.config.AddFilter(filterType, listType, value)
	if err != nil {
		fmt.Printf("❌ Error adding filter: %v\n", err)
		return
	}

	c.config.Save(monday.GetConfigPath())
	fmt.Printf("✅ Added '%s' to %s %s\n", value, listType, filterType)
}

func (c *CLI) HandleRemoveFilterCommand() {
	if len(c.command.Args) < 4 {
		fmt.Println("Usage: monday-cli config remove-filter <type> <whitelist|blacklist> <value>")
		fmt.Println("Types: status, priority, type, sprint, user_name, user_email")
		fmt.Println("Example: monday-cli config remove-filter status whitelist 'in progress'")
		return
	}

	filterType := monday.FilterType(c.command.Args[1])
	listType := monday.FilterListType(c.command.Args[2])
	value := c.command.Args[3]

	// Validate filter type
	validTypes := []monday.FilterType{
		monday.FilterStatus, monday.FilterPriority, monday.FilterTaskType,
		monday.FilterSprint, monday.FilterUserName, monday.FilterUserEmail,
	}
	validType := false
	for _, vt := range validTypes {
		if filterType == vt {
			validType = true
			break
		}
	}
	if !validType {
		fmt.Printf("❌ Invalid filter type: %s\n", filterType)
		fmt.Println("Valid types: status, priority, type, sprint, user_name, user_email")
		return
	}

	// Validate list type
	if listType != monday.Whitelist && listType != monday.Blacklist {
		fmt.Printf("❌ Invalid list type: %s\n", listType)
		fmt.Println("Valid types: whitelist, blacklist")
		return
	}

	err := c.config.RemoveFilter(filterType, listType, value)
	if err != nil {
		fmt.Printf("❌ Error removing filter: %v\n", err)
		return
	}

	c.config.Save(monday.GetConfigPath())
	fmt.Printf("✅ Removed '%s' from %s %s\n", value, listType, filterType)
}

func (c *CLI) HandleClearFilterCommand() {
	if len(c.command.Args) < 3 {
		fmt.Println("Usage: monday-cli config clear-filter <type> <whitelist|blacklist>")
		fmt.Println("Types: status, priority, type, sprint, user_name, user_email")
		fmt.Println("Example: monday-cli config clear-filter status whitelist")
		return
	}

	filterType := monday.FilterType(c.command.Args[1])
	listType := monday.FilterListType(c.command.Args[2])

	// Validate filter type
	validTypes := []monday.FilterType{
		monday.FilterStatus, monday.FilterPriority, monday.FilterTaskType,
		monday.FilterSprint, monday.FilterUserName, monday.FilterUserEmail,
	}
	validType := false
	for _, vt := range validTypes {
		if filterType == vt {
			validType = true
			break
		}
	}
	if !validType {
		fmt.Printf("❌ Invalid filter type: %s\n", filterType)
		fmt.Println("Valid types: status, priority, type, sprint, user_name, user_email")
		return
	}

	// Validate list type
	if listType != monday.Whitelist && listType != monday.Blacklist {
		fmt.Printf("❌ Invalid list type: %s\n", listType)
		fmt.Println("Valid types: whitelist, blacklist")
		return
	}

	err := c.config.ClearFilter(filterType, listType)
	if err != nil {
		fmt.Printf("❌ Error clearing filter: %v\n", err)
		return
	}

	c.config.Save(monday.GetConfigPath())
	fmt.Printf("✅ Cleared %s %s\n", listType, filterType)
}

func (c *CLI) HandleListFiltersCommand() {
	fmt.Println("🔍 Current Filters:")
	fmt.Println("=" + strings.Repeat("=", 50))

	filterTypes := []monday.FilterType{
		monday.FilterStatus, monday.FilterPriority, monday.FilterTaskType,
		monday.FilterSprint, monday.FilterUserName, monday.FilterUserEmail,
	}

	for _, filterType := range filterTypes {
		fmt.Printf("\n📋 %s:\n", strings.ToUpper(string(filterType)))

		// Show whitelist
		whitelist := c.config.GetFilterValues(filterType, monday.Whitelist)
		if len(whitelist) > 0 {
			fmt.Printf("  ✅ Whitelist: %v\n", whitelist)
		} else {
			fmt.Printf("  ✅ Whitelist: (empty)\n")
		}

		// Show blacklist
		blacklist := c.config.GetFilterValues(filterType, monday.Blacklist)
		if len(blacklist) > 0 {
			fmt.Printf("  ❌ Blacklist: %v\n", blacklist)
		} else {
			fmt.Printf("  ❌ Blacklist: (empty)\n")
		}
	}
}

func (c *CLI) HandleClearAllFiltersCommand() {
	c.config.ClearAllFilters()
	c.config.Save(monday.GetConfigPath())
	fmt.Println("✅ Cleared all filters")
}

func (c *CLI) HandleFilterToMeCommand() {
	err := c.config.FilterToCurrentUser()
	if err != nil {
		fmt.Printf("❌ Error filtering to current user: %v\n", err)
		return
	}

	c.config.Save(monday.GetConfigPath())
	fmt.Println("✅ Filtered to show only tasks assigned to you")
}

func (c *CLI) HandleAddMeCommand() {
	err := c.config.AddCurrentUserToWhitelist()
	if err != nil {
		fmt.Printf("❌ Error adding current user to whitelist: %v\n", err)
		return
	}

	c.config.Save(monday.GetConfigPath())
	fmt.Println("✅ Added current user to whitelist")
}

func (c *CLI) HandleRemoveMeCommand() {
	err := c.config.RemoveCurrentUserFromWhitelist()
	if err != nil {
		fmt.Printf("❌ Error removing current user from whitelist: %v\n", err)
		return
	}

	c.config.Save(monday.GetConfigPath())
	fmt.Println("✅ Removed current user from whitelist")
}

func (c *CLI) HandleListBoardUsersCommand() {
	dataStore := monday.NewDataStore()
	users, timestamp, ok := dataStore.GetCachedBoardUsers(c.config.GetBoardID())

	if !ok || len(users) == 0 {
		fmt.Println("❌ No board users found in cache")
		fmt.Println("💡 Run 'tasks fetch' first to fetch board users")
		return
	}

	fmt.Printf("👥 Board Users (cached at: %s)\n", timestamp.Format(time.RFC3339))
	fmt.Println("=" + strings.Repeat("=", 50))

	for i, user := range users {
		status := "❌ Disabled"
		if user.Enabled {
			status = "✅ Enabled"
		}

		fmt.Printf("%d. %s (%s)\n", i+1, user.Name, user.Email)
		fmt.Printf("   🆔 ID: %s\n", user.ID)
		if user.Title != "" {
			fmt.Printf("   💼 Title: %s\n", user.Title)
		}
		fmt.Printf("   🔐 Status: %s\n", status)
		if user.PhotoURL != "" {
			fmt.Printf("   🖼️  Photo: %s\n", user.PhotoURL)
		}
		fmt.Println()
	}

	fmt.Printf("📊 Total users: %d\n", len(users))
}

// HandleListBoardSprintsCommand lists all sprints found on the sprint board
func (c *CLI) HandleListBoardSprintsCommand() {
	sprintBoardID := c.config.GetSprintBoardID()
	if sprintBoardID == "" {
		fmt.Println("❌ No sprint board ID configured")
		fmt.Println("💡 Run 'config set-sprint-board-id <sprint-board-id>' first")
		return
	}

	dataStore := monday.NewDataStore()
	sprints, timestamp, ok := dataStore.GetCachedBoardSprints(sprintBoardID)

	if !ok || len(sprints) == 0 {
		fmt.Println("❌ No board sprints found in cache")
		fmt.Println("💡 Run 'tasks fetch' first to fetch board sprints")
		return
	}

	fmt.Printf("🏃 Sprint Board Sprints (cached at: %s)\n", timestamp.Format(time.RFC3339))
	fmt.Println("=" + strings.Repeat("=", 50))

	for i, sprint := range sprints {
		fmt.Printf("%d. %s\n", i+1, sprint)
	}

	fmt.Printf("📊 Total sprints: %d\n", len(sprints))
}

// HandleSprintCommand handles sprint-specific commands
func (c *CLI) HandleSprintCommand() {
	if len(c.command.Args) < 1 {
		c.HelpSprintCommand()
		return
	}

	subcommand := c.command.Args[1]
	switch subcommand {
	case "fetch", "f":
		c.HandleSprintFetchCommand()
		return
	case "list", "ls":
		c.HandleSprintListCommand()
		return
	default:
		c.HelpSprintCommand()
		return
	}
}

// HandleSprintFetchCommand fetches items from the current sprint
func (c *CLI) HandleSprintFetchCommand() {
	sprintID := c.config.GetSprintID()
	if sprintID == "" {
		fmt.Println("❌ No sprint ID configured")
		fmt.Println("💡 Run 'config set-sprint-id <sprint-id>' first")
		return
	}

	boardID := c.config.GetBoardID()
	if boardID == "" {
		fmt.Println("❌ No board ID configured")
		fmt.Println("💡 Run 'config set-board-id <board-id>' first")
		return
	}

	client := monday.NewClient(c.config.GetAPIKey(), c.config.Timeout)

	fmt.Printf("🔍 Fetching items from sprint %s...\n", sprintID)

	tasks, items, err := client.GetSprintItems(sprintID)
	if err != nil {
		fmt.Printf("❌ Error fetching sprint items: %v\n", err)
		return
	}

	if len(tasks) == 0 {
		fmt.Printf("👤 No tasks found in sprint %s\n", sprintID)
		return
	}

	// Merge sprint tasks into board cache (same array as regular tasks)
	dataStore := monday.NewDataStore()
	dataStore.MergeSprintTasksIntoBoard(boardID, tasks, items)

	// Also store in sprint cache for backward compatibility
	dataStore.StoreSprintItems(sprintID, tasks, items)

	// Get merged tasks from board cache to display
	cachedTasks, _, _ := dataStore.GetCachedTasks(boardID)

	// Display the tasks
	c.PrintItems(cachedTasks)
}

// HandleSprintListCommand lists items from the current sprint
func (c *CLI) HandleSprintListCommand() {
	sprintID := c.config.GetSprintID()
	if sprintID == "" {
		fmt.Println("❌ No sprint ID configured")
		fmt.Println("💡 Run 'config set-sprint-id <sprint-id>' first")
		return
	}

	boardID := c.config.GetBoardID()
	if boardID == "" {
		fmt.Println("❌ No board ID configured")
		fmt.Println("💡 Run 'config set-board-id <board-id>' first")
		return
	}

	// Sprint tasks are now stored in the board cache with regular tasks
	dataStore := monday.NewDataStore()
	tasksMap, _, ok := dataStore.GetCachedTasks(boardID)

	if !ok || len(tasksMap) == 0 {
		fmt.Println("❌ No tasks found in cache")
		fmt.Println("💡 Run 'tasks fetch' or 'tasks sprint fetch' first to fetch tasks")
		return
	}

	// Filter tasks by sprint if needed (tasks fetched from sprint will have Sprint field set)
	c.PrintItems(tasksMap)
}

// HelpSprintCommand shows help for sprint commands
func (c *CLI) HelpSprintCommand() {
	fmt.Println("Sprint Commands:")
	fmt.Println("  tasks sprint fetch (f)    Fetch items from current sprint")
	fmt.Println("  tasks sprint list (ls)   List cached sprint items")
	fmt.Println("")
	fmt.Println("Configuration:")
	fmt.Println("  config set-sprint-id <id>  Set the current sprint ID")
	fmt.Println("  config show                Show current sprint ID")
}

// HandleFilterToSprintCommand filters to show only tasks from the current sprint
func (c *CLI) HandleFilterToSprintCommand() {
	err := c.config.FilterToCurrentSprint()
	if err != nil {
		fmt.Printf("❌ Error filtering to current sprint: %v\n", err)
		return
	}

	c.config.Save(monday.GetConfigPath())
	fmt.Println("✅ Filtered to show only tasks from current sprint")
}

// HandleAddSprintCommand adds the current sprint to the whitelist
func (c *CLI) HandleAddSprintCommand() {
	err := c.config.AddCurrentSprintToWhitelist()
	if err != nil {
		fmt.Printf("❌ Error adding current sprint to whitelist: %v\n", err)
		return
	}

	c.config.Save(monday.GetConfigPath())
	fmt.Println("✅ Added current sprint to whitelist")
}

// HandleRemoveSprintCommand removes the current sprint from the whitelist
func (c *CLI) HandleRemoveSprintCommand() {
	err := c.config.RemoveCurrentSprintFromWhitelist()
	if err != nil {
		fmt.Printf("❌ Error removing current sprint from whitelist: %v\n", err)
		return
	}

	c.config.Save(monday.GetConfigPath())
	fmt.Println("✅ Removed current sprint from whitelist")
}
