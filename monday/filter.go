package monday

import (
	"slices"
	"strings"
)

func FilterTasks(tasks []Task, filters Filters) []Task {
	var filteredTasks []Task
	for _, task := range tasks {
		status := strings.ToLower(string(task.Status))
		priority := strings.ToLower(string(task.Priority))
		itemType := strings.ToLower(string(task.Type))
		sprint := strings.ToLower(string(task.Sprint))
		userName := strings.ToLower(string(task.UserName))
		userEmail := strings.ToLower(string(task.UserEmail))
		if len(filters.StatusWhitelist) > 0 && !slices.Contains(filters.StatusWhitelist, status) {
			continue
		}
		if len(filters.StatusBlacklist) > 0 && slices.Contains(filters.StatusBlacklist, status) {
			continue
		}
		if len(filters.PriorityWhitelist) > 0 && !slices.Contains(filters.PriorityWhitelist, priority) {
			continue
		}
		if len(filters.PriorityBlacklist) > 0 && slices.Contains(filters.PriorityBlacklist, priority) {
			continue
		}
		if len(filters.TypeWhitelist) > 0 && !slices.Contains(filters.TypeWhitelist, itemType) {
			continue
		}
		if len(filters.TypeBlacklist) > 0 && slices.Contains(filters.TypeBlacklist, itemType) {
			continue
		}
		if len(filters.SprintWhitelist) > 0 && !slices.Contains(filters.SprintWhitelist, sprint) {
			continue
		}
		if len(filters.SprintBlacklist) > 0 && slices.Contains(filters.SprintBlacklist, sprint) {
			continue
		}
		if len(filters.UserNameWhitelist) > 0 && !slices.Contains(filters.UserNameWhitelist, userName) {
			continue
		}
		if len(filters.UserNameBlacklist) > 0 && slices.Contains(filters.UserNameBlacklist, userName) {
			continue
		}
		if len(filters.UserEmailWhitelist) > 0 && !slices.Contains(filters.UserEmailWhitelist, userEmail) {
			continue
		}
		if len(filters.UserEmailBlacklist) > 0 && slices.Contains(filters.UserEmailBlacklist, userEmail) {
			continue
		}
		filteredTasks = append(filteredTasks, task)
	}
	return filteredTasks
}
