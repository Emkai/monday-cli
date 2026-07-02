package monday

import "testing"

// Fixtures mirror the real echandia boards inspected during design.

func tasksBoardColumns() []Column {
	return []Column{
		{ID: "name", Title: "Name", Type: "name"},
		{ID: "task_owner", Title: "Owner", Type: "people"},
		{ID: "task_status", Title: "Status", Type: "status", SettingsStr: `{"labels":{"0":"In Progress","1":"Done","2":"Ready for testing","3":"Waiting for review","4":"Removed","5":"","103":"Stuck"}}`},
		{ID: "task_priority", Title: "Priority", Type: "status", SettingsStr: `{"labels":{"0":"Critical","2":"High","7":"Medium","12":"Low"}}`},
		{ID: "task_type", Title: "Type", Type: "status", SettingsStr: `{"labels":{"1":"Feature","2":"Bug","12":"Test"}}`},
		{ID: "task_sprint", Title: "Sprint", Type: "board_relation"},
		// A decoy status column whose title merely contains "status" — must NOT
		// win over task_status.
		{ID: "color_mm2kr4r3", Title: "AI - Status", Type: "status"},
	}
}

func bugsBoardColumns() []Column {
	return []Column{
		{ID: "name", Title: "Name", Type: "name"},
		{ID: "people1", Title: "Reporter", Type: "people"},
		{ID: "bug_status", Title: "Status", Type: "status", SettingsStr: `{"labels":{"1":"Fixed","9":"Awaiting Review","14":"Accepted"}}`},
		{ID: "priority_1", Title: "Priority", Type: "status", SettingsStr: `{"labels":{"0":"Critical","2":"High","7":"Medium"}}`},
		{ID: "color_mkxcj5nd", Title: "Found in release", Type: "status"},
	}
}

func TestDetectRolesTasksBoard(t *testing.T) {
	r := DetectRoles(tasksBoardColumns())
	want := ColumnRoles{
		People:   "task_owner",
		Status:   "task_status",
		Priority: "task_priority",
		Type:     "task_type",
		Sprint:   "task_sprint",
	}
	if r != want {
		t.Fatalf("DetectRoles(tasks) = %+v, want %+v", r, want)
	}
}

func TestDetectRolesBugsBoard(t *testing.T) {
	r := DetectRoles(bugsBoardColumns())
	// Bugs board: people is the Reporter, no Type or Sprint column.
	if r.People != "people1" {
		t.Errorf("People = %q, want people1", r.People)
	}
	if r.Status != "bug_status" {
		t.Errorf("Status = %q, want bug_status", r.Status)
	}
	if r.Priority != "priority_1" {
		t.Errorf("Priority = %q, want priority_1", r.Priority)
	}
	if r.Type != "" {
		t.Errorf("Type = %q, want empty (bugs board has no type column)", r.Type)
	}
	if r.Sprint != "" {
		t.Errorf("Sprint = %q, want empty", r.Sprint)
	}
}

func TestDetectRolesActiveCheckbox(t *testing.T) {
	cols := []Column{{ID: "sprint_activation", Title: "Active sprint", Type: "checkbox"}}
	if r := DetectRoles(cols); r.Active != "sprint_activation" {
		t.Fatalf("Active = %q, want sprint_activation", r.Active)
	}
}

func TestParseLabels(t *testing.T) {
	m := ParseLabels(`{"labels":{"0":"In Progress","1":"Done","4":"Removed","5":""}}`)
	if got := m["in progress"]; got != 0 {
		t.Errorf("in progress -> %d, want 0", got)
	}
	if got := m["done"]; got != 1 {
		t.Errorf("done -> %d, want 1", got)
	}
	if got := m["removed"]; got != 4 {
		t.Errorf("removed -> %d, want 4", got)
	}
	if _, ok := m[""]; ok {
		t.Error("blank label should be dropped")
	}
}

func TestBuildLabels(t *testing.T) {
	roles := DetectRoles(tasksBoardColumns())
	labels := BuildLabels(roles, tasksBoardColumns())
	if labels["status"]["done"] != 1 {
		t.Errorf("status done -> %d, want 1", labels["status"]["done"])
	}
	if labels["priority"]["critical"] != 0 {
		t.Errorf("priority critical -> %d, want 0", labels["priority"]["critical"])
	}
	if labels["type"]["bug"] != 2 {
		t.Errorf("type bug -> %d, want 2", labels["type"]["bug"])
	}
}

func TestItemToTask(t *testing.T) {
	roles := DetectRoles(tasksBoardColumns())
	item := Item{
		ID:   "42",
		Name: "Fix the thing",
		ColumnValues: []ColumnValue{
			{ID: "task_owner", Text: "Mattias"},
			{ID: "task_status", Text: "In Progress"},
			{ID: "task_priority", Text: "High"},
			{ID: "task_type", Text: "Bug"},
		},
	}
	got := ItemToTask(item, roles)
	if got.Owner != "Mattias" || got.Status != "In Progress" || got.Priority != "High" || got.Type != "Bug" {
		t.Fatalf("ItemToTask = %+v", got)
	}
}
