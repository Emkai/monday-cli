package cli

import (
	"emkai/go-cli-gui/monday"
	"fmt"
	"strings"
)

func (c *CLI) PrintItems(items map[int]monday.Item, sortedItems []monday.Item) {
	var activeItems map[int]monday.Item = make(map[int]monday.Item)
	sprintID := c.config.GetSprintID()
	for id, item := range items {
		skipTask := false
		for _, cv := range item.ColumnValues {
			if strings.Contains(strings.ToLower(cv.ID), "status") && cv.Text != "" {
				status := strings.ToLower(cv.Text)
				if strings.Contains(status, "done") || strings.Contains(status, "completed") {
					skipTask = true
					break
				}
			}
			if sprintID != "" && strings.Contains(strings.ToLower(cv.ID), "sprint") && cv.Text != "" {
				sprint := strings.ToLower(cv.Text)
				if strings.Contains(sprint, sprintID) {
					skipTask = true
					break
				}
			}
		}
		if !skipTask {
			activeItems[id] = item
		}
	}

	if len(activeItems) == 0 {
		fmt.Printf("🎉 No active tasks assigned to %s in %s\n", c.config.GetOwnerEmail(), c.config.GetBoardID())
		return
	}

	fmt.Printf("👤 Found %d tasks assigned to %s:\n\n", len(items), c.config.GetOwnerEmail())

	fmt.Println("Type [Status Priority] Task Name")
	// Use sortedItems to maintain order, but only print active items
	for _, item := range sortedItems {
		// Find the ID for this item in the map
		for id, mapItem := range items {
			if mapItem.ID == item.ID {
				if _, isActive := activeItems[id]; isActive {
					c.PrintTask(id, item)
				}
				break
			}
		}
	}

	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("📊 Active tasks: %d\n", len(activeItems))

}

func (c *CLI) PrintTask(id int, item monday.Item) {
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
	fmt.Printf("%d. %s [%s %s] %s\n", id, taskType, status, priority, item.Name)
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
	case strings.Contains(status, "testing") || strings.Contains(status, "not started"):
		return "🧪"
	case strings.Contains(status, "removed"):
		return "🗑️"
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
		return "🟡"
	case strings.Contains(priority, "medium"):
		return "🔵"
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
	case strings.Contains(taskType, "quality"):
		return "📈"
	case strings.Contains(taskType, "other"):
		return "📝"
	default:
		return "📝"
	}
}
