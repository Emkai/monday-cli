package monday

import "strings"

func isStatusType(t string) bool { return t == "status" || t == "color" }

// DetectRoles infers a board's column-role map from its columns, using column
// type + title heuristics. It is deliberately conservative: a role left empty
// simply means the CLI won't offer that filter for the board until the user
// overrides it with `board set-col`.
func DetectRoles(cols []Column) ColumnRoles {
	var r ColumnRoles
	for _, c := range cols {
		title := strings.ToLower(strings.TrimSpace(c.Title))
		switch {
		case c.Type == "people" && r.People == "":
			r.People = c.ID
		case isStatusType(c.Type) && strings.Contains(title, "priority") && r.Priority == "":
			r.Priority = c.ID
		case isStatusType(c.Type) && title == "type" && r.Type == "":
			r.Type = c.ID
		case isStatusType(c.Type) && (c.ID == "status" || title == "status") && r.Status == "":
			r.Status = c.ID
		case c.Type == "board_relation" && strings.Contains(title, "sprint") && r.Sprint == "":
			r.Sprint = c.ID
		case c.Type == "checkbox" && strings.Contains(title, "active") && r.Active == "":
			r.Active = c.ID
		}
	}
	// Fallback: a status column whose title merely contains "status" (e.g. a
	// custom column ID), avoiding the ones already claimed as priority/type.
	if r.Status == "" {
		for _, c := range cols {
			if isStatusType(c.Type) && strings.Contains(strings.ToLower(c.Title), "status") &&
				c.ID != r.Priority && c.ID != r.Type {
				r.Status = c.ID
				break
			}
		}
	}
	return r
}

// BuildLabels extracts label -> index maps for a favorite's status-like roles so
// filters and mutations can translate label text to the index Monday expects.
func BuildLabels(roles ColumnRoles, cols []Column) map[string]map[string]int {
	byID := make(map[string]Column, len(cols))
	for _, c := range cols {
		byID[c.ID] = c
	}
	labels := map[string]map[string]int{}
	for _, role := range []string{"status", "priority", "type"} {
		id := roles.Get(role)
		if id == "" {
			continue
		}
		if c, ok := byID[id]; ok {
			if m := ParseLabels(c.SettingsStr); len(m) > 0 {
				labels[role] = m
			}
		}
	}
	return labels
}
