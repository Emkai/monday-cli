package cli

import (
	"emkai/go-cli-gui/monday"
	"fmt"
	"strconv"
	"strings"
)

// ANSI color codes
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorGray    = "\033[90m"
)

// Color helper functions
func colorize(text, color string) string {
	return color + text + ColorReset
}

// Maps for assigning colors by value
var statusColorMap = map[string]string{
	"done":        ColorGreen,
	"completed":   ColorGreen,
	"in progress": ColorBlue,
	"progress":    ColorBlue,
	"review":      ColorMagenta,
	"stuck":       ColorRed,
	"blocked":     ColorRed,
	"testing":     ColorCyan,
	"not started": ColorGray,
	"removed":     ColorGray,
}

var priorityColorMap = map[string]string{
	"critical": ColorRed,
	"high":     ColorYellow,
	"medium":   ColorBlue,
	"low":      ColorGreen,
}

var typeColorMap = map[string]string{
	"bug":      ColorRed,
	"feature":  ColorGreen,
	"test":     ColorCyan,
	"security": ColorMagenta,
	"quality":  ColorBlue,
	"other":    ColorWhite,
}

func (c *CLI) PrintItems(tasks map[string]monday.Task) {
	tasksList := make([]monday.Task, 0, len(tasks))
	for _, task := range tasks {
		tasksList = append(tasksList, task)
	}

	filteredTasks := monday.FilterTasks(tasksList, c.config.GetFilters())

	fmt.Printf("👤 Found %d tasks to matching filters:\n\n", len(filteredTasks))

	sortedTasks := monday.OrderTasks(filteredTasks)

	currentStatus := ""
	activeCount := 0
	for i := 0; i < len(sortedTasks); i++ {
		task := sortedTasks[i]
		if string(task.Status) != currentStatus {
			currentStatus = string(task.Status)
			statusIcon := getStatusIcon(currentStatus)
			statusColor := getStatusColor(currentStatus)
			if currentStatus == "" {
				fmt.Printf("\n%s %s\n", statusIcon, colorize("None", ColorWhite))
			} else {
				fmt.Printf("\n%s %s\n", statusIcon, colorize(currentStatus, statusColor))
			}
		}
		if isActiveStatus(string(task.Status)) {
			activeCount++
		}
		PrintTask(task)
	}

	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("📊 Active tasks: %d\n", activeCount)
}

// Define which statuses are considered 'active'
func isActiveStatus(status string) bool {
	status = strings.ToLower(status)
	// You can adjust these according to your workflow
	return !(strings.Contains(status, "done") || strings.Contains(status, "completed") || strings.Contains(status, "removed"))
}

func PrintTask(task monday.Task) {
	// Extract status, priority, and type
	priorityColor := getPriorityColor(string(task.Priority))
	taskTypeIcon := getTypeIcon(string(task.Type))

	fmt.Printf("%s. %s [%s] %s, (%s, %s)\n",
		padLocalId(task.LocalId),
		taskTypeIcon,
		colorize(padPriority(string(task.Priority)), priorityColor),
		task.Name,
		task.UserName,
		task.UserEmail,
	)
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

func PrintUserInfo(user *monday.User) {
	fmt.Printf("👤 User Information\n")
	fmt.Println("-" + strings.Repeat("-", 50))
	fmt.Printf("🆔 ID: %s\n", user.ID)
	fmt.Printf("👤 Name: %s\n", user.Name)
	fmt.Printf("📧 Email: %s\n", user.Email)
	if user.Title != "" {
		fmt.Printf("💼 Title: %s\n", user.Title)
	}
	if user.PhotoURL != "" {
		fmt.Printf("🖼️  Photo: %s\n", user.PhotoURL)
	}
	status := "❌ Disabled"
	if user.Enabled {
		status = "✅ Enabled"
	}
	fmt.Printf("🔐 Status: %s\n", status)
	fmt.Println("=" + strings.Repeat("=", 50))
}

func PrintCommand(cmd Command) {
	fmt.Println("Command: " + cmd.Command)
	fmt.Println("Args:")
	for _, arg := range cmd.Args {
		fmt.Println("    Arg: " + arg)
	}
	fmt.Println("Flags:")
	for _, flag := range cmd.Flags {
		fmt.Println("    Flag: " + flag.Flag + " Value: " + flag.Value)
	}
}

// Color assignment logic
func getStatusColor(status string) string {
	status = strings.ToLower(status)
	for k, c := range statusColorMap {
		if strings.Contains(status, k) {
			return c
		}
	}
	return ColorWhite
}

func getPriorityColor(priority string) string {
	priority = strings.ToLower(priority)
	for k, c := range priorityColorMap {
		if strings.Contains(priority, k) {
			return c
		}
	}
	return ColorWhite
}

func getTypeColor(taskType string) string {
	taskType = strings.ToLower(taskType)
	for k, c := range typeColorMap {
		if strings.Contains(taskType, k) {
			return c
		}
	}
	return ColorWhite
}

func padPriority(priority string) string {
	maxLen := 8 // "critical" is the longest priority string (8 letters)
	padding := maxLen - len(priority)
	leftPad := padding / 2
	rightPad := padding - leftPad
	return strings.Repeat(" ", leftPad+1) + priority + strings.Repeat(" ", rightPad+1)
}

func padLocalId(localId int) string {
	s := strconv.Itoa(localId)
	for len(s) < 4 {
		s = " " + s
	}
	return s
}
