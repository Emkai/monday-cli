package monday

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Client represents a Monday.com API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Monday.com API client
func NewClient(apiKey string, timeout int) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.monday.com/v2",
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// GraphQLRequest represents a GraphQL request to Monday.com
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response from Monday.com
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// ExecuteQuery executes a GraphQL query against Monday.com API
func (c *Client) ExecuteQuery(query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var graphqlResp GraphQLResponse
	if err := json.Unmarshal(body, &graphqlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(graphqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", graphqlResp.Errors)
	}

	return &graphqlResp, nil
}

// GetBoard retrieves a specific board by ID
func (c *Client) GetBoard(boardID string) (*Board, error) {
	query := `
		query GetBoard($boardId: ID!) {
			boards(ids: [$boardId]) {
				id
				name
				description
				state
				updated_at
				columns {
					id
					title
					type
					settings_str
				}
			}
		}
	`

	variables := map[string]interface{}{
		"boardId": boardID,
	}

	resp, err := c.ExecuteQuery(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Boards []Board `json:"boards"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal board: %w", err)
	}

	if len(result.Boards) == 0 {
		return nil, fmt.Errorf("board not found")
	}

	return &result.Boards[0], nil
}

// GetBoardItemsByOwner retrieves items from a specific board filtered by owner using pagination
func (c *Client) GetBoardItems(boardID string) ([]Task, []Item, error) {
	board, err := c.GetBoard(boardID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get board: %w", err)
	}

	// Find the owner column ID
	var ownerColumnID string
	for _, column := range board.Columns {
		if strings.Contains(strings.ToLower(column.Title), "owner") {
			ownerColumnID = column.ID
			break
		}
	}
	if ownerColumnID == "" {
		return nil, nil, fmt.Errorf("owner column not found in board")
	}

	var allItems []Item
	cursor := ""
	limit := 25 // Smaller page size for better performance

	for {
		query := `
			query GetBoardItemsByOwner($boardId: ID!, $limit: Int!, $cursor: String) {
				boards(ids: [$boardId]) {
					items_page(limit: $limit, cursor: $cursor) {
						items {
							id
							name
							column_values {
								id
								text
								value
							}
							updated_at
						}
						cursor
					}
				}
			}
		`

		variables := map[string]interface{}{
			"boardId": boardID,
			"limit":   limit,
		}

		if cursor != "" {
			variables["cursor"] = cursor
		}

		resp, err := c.ExecuteQuery(query, variables)
		if err != nil {
			return nil, nil, err
		}

		var result struct {
			Boards []struct {
				ItemsPage struct {
					Items  []Item `json:"items"`
					Cursor string `json:"cursor"`
				} `json:"items_page"`
			} `json:"boards"`
		}

		if err := json.Unmarshal(resp.Data, &result); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal board items: %w", err)
		}

		if len(result.Boards) == 0 {
			return nil, nil, fmt.Errorf("board not found")
		}

		allItems = append(allItems, result.Boards[0].ItemsPage.Items...)

		cursor = result.Boards[0].ItemsPage.Cursor
		if cursor == "" || len(result.Boards[0].ItemsPage.Items) < limit {
			break
		}
		fmt.Printf("Fetching next page... currently %d items\n", len(allItems))
	}

	var allTasks []Task
	localId := 1
	for _, item := range allItems {
		task := Task{
			ID:        item.ID,
			LocalId:   localId,
			Name:      item.Name,
			UpdatedAt: item.UpdatedAt,
		}
		localId++
		for _, cv := range item.ColumnValues {
			if strings.Contains(strings.ToLower(cv.ID), "status") && cv.Text != "" {
				task.Status = Status(cv.Text)
			}
			if strings.Contains(strings.ToLower(cv.ID), "priority") && cv.Text != "" {
				task.Priority = Priority(cv.Text)
			}
			if strings.Contains(strings.ToLower(cv.ID), "type") && cv.Text != "" {
				task.Type = Type(cv.Text)
			}
			if strings.Contains(strings.ToLower(cv.ID), "sprint") && cv.Text != "" {
				task.Sprint = Sprint(cv.Text)
			}
			if strings.Contains(strings.ToLower(cv.ID), "user_name") && cv.Text != "" {
				task.UserName = cv.Text
			}
			if strings.Contains(strings.ToLower(cv.ID), "user_email") && cv.Text != "" {
				task.UserEmail = cv.Text
			}
		}
		allTasks = append(allTasks, task)
	}
	return allTasks, allItems, nil
}

func OrderTasks(tasks []Task) []Task {
	sort.Slice(tasks, func(i, j int) bool {
		statusI := getSortableStatus(tasks[i])
		statusJ := getSortableStatus(tasks[j])

		// First sort by status
		if statusI != statusJ {
			return statusI < statusJ
		}

		// Then by priority
		priorityI := getSortablePriority(tasks[i])
		priorityJ := getSortablePriority(tasks[j])
		if priorityI != priorityJ {
			return priorityI < priorityJ
		}

		// Finally by type
		typeI := getSortableType(tasks[i])
		typeJ := getSortableType(tasks[j])
		return typeI < typeJ

	})
	return tasks
}

func (c *Client) UpdateTaskStatus(boardID, ownerEmail string, task Item, newStatus string) error {
	// First, get the board to find the status column ID
	board, err := c.GetBoard(boardID)
	if err != nil {
		return fmt.Errorf("failed to get board: %w", err)
	}

	// Find the status column ID
	var statusColumnID string
	for _, column := range board.Columns {
		if strings.Contains(strings.ToLower(column.Title), "status") {
			statusColumnID = column.ID
			break
		}
	}
	if statusColumnID == "" {
		return fmt.Errorf("status column not found in board")
	}

	query := `
		mutation UpdateTaskStatus($boardId: ID!, $itemId: ID!, $columnId: String!, $value: JSON!) {
			change_column_value(board_id: $boardId, item_id: $itemId, column_id: $columnId, value: $value) {
				id
			}
		}
	`

	// Use the task's actual ID
	itemID := task.ID

	// Create the JSON value for status column - Monday.com expects a JSON string
	statusValue := fmt.Sprintf(`{"label": "%s"}`, newStatus)

	variables := map[string]interface{}{
		"boardId":  boardID,
		"itemId":   itemID,
		"columnId": statusColumnID,
		"value":    statusValue,
	}

	fmt.Println(variables)
	resp, err := c.ExecuteQuery(query, variables)
	if err != nil {
		return err
	}

	if len(resp.Errors) > 0 {
		return fmt.Errorf("failed to update task status: %v", resp.Errors)
	}

	fmt.Printf("✅ Task %s status updated to %s\n", task.ID, newStatus)

	return nil
}

// UpdateTask updates multiple fields of a task
func (c *Client) UpdateTask(boardID, ownerEmail string, task Task, status, priority, taskType string) (*Task, error) {
	// First, get the board to find the column IDs
	board, err := c.GetBoard(boardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}

	// Find column IDs
	var statusColumnID, priorityColumnID, typeColumnID string
	for _, column := range board.Columns {
		title := strings.ToLower(column.Title)
		if strings.Contains(title, "status") {
			statusColumnID = column.ID
		} else if strings.Contains(title, "priority") {
			priorityColumnID = column.ID
		} else if strings.Contains(title, "type") {
			typeColumnID = column.ID
		}
	}

	// Build column updates
	columnUpdates := make(map[string]string)

	if status != "" && statusColumnID != "" {
		columnUpdates[statusColumnID] = fmt.Sprintf(`{"label": "%s"}`, status)
	}

	if priority != "" && priorityColumnID != "" {
		columnUpdates[priorityColumnID] = fmt.Sprintf(`{"label": "%s"}`, priority)
	}

	if taskType != "" && typeColumnID != "" {
		columnUpdates[typeColumnID] = fmt.Sprintf(`{"label": "%s"}`, taskType)
	}

	// If no fields to update, return the original task
	if len(columnUpdates) == 0 {
		return &task, nil
	}

	// Create the mutation query
	query := `
		mutation UpdateTask($boardId: ID!, $itemId: ID!, $columnValues: JSON!) {
			change_multiple_column_values(board_id: $boardId, item_id: $itemId, column_values: $columnValues) {
				id
			}
		}
	`

	// Create column values JSON
	columnValues := "{"
	first := true
	for columnID, value := range columnUpdates {
		if !first {
			columnValues += ","
		}
		columnValues += fmt.Sprintf(`"%s": %s`, columnID, value)
		first = false
	}
	columnValues += "}"

	variables := map[string]interface{}{
		"boardId":      boardID,
		"itemId":       task.ID,
		"columnValues": columnValues,
	}

	resp, err := c.ExecuteQuery(query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("failed to update task: %v", resp.Errors)
	}

	// Fetch the updated task to return the latest data
	updatedTask, err := c.GetTaskByID(task.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated task: %w", err)
	}

	return updatedTask, nil
}

func (c *Client) CreateTask(boardID, userID, taskName, status, priority, taskType string) (int, *Task, error) {

	// Get board to find column IDs
	board, err := c.GetBoard(boardID)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get board: %w", err)
	}

	// Find column IDs
	var statusColumnID, priorityColumnID, typeColumnID string
	for _, column := range board.Columns {
		title := strings.ToLower(column.Title)
		if strings.Contains(title, "status") {
			statusColumnID = column.ID
		} else if strings.Contains(title, "priority") {
			priorityColumnID = column.ID
		} else if strings.Contains(title, "type") {
			typeColumnID = column.ID
		}
	}

	query := `
		mutation CreateTask($boardId: ID!, $itemName: String!, $columnValues: JSON!) {
			create_item(board_id: $boardId, item_name: $itemName, column_values: $columnValues) {
				id
			}
		}
	`

	// Create column values JSON with all specified values
	columnValues := fmt.Sprintf(`{"task_owner": {"personsAndTeams":[{"id":%s,"kind":"person"}],"changed_at":"%s"}`,
		userID,
		time.Now().Format(time.RFC3339))

	// Add status if provided
	if status != "" && statusColumnID != "" {
		columnValues += fmt.Sprintf(`,"%s": {"label": "%s"}`, statusColumnID, status)
	}

	// Add priority if provided
	if priority != "" && priorityColumnID != "" {
		columnValues += fmt.Sprintf(`,"%s": {"label": "%s"}`, priorityColumnID, priority)
	}

	// Add type if provided
	if taskType != "" && typeColumnID != "" {
		columnValues += fmt.Sprintf(`,"%s": {"label": "%s"}`, typeColumnID, taskType)
	}

	columnValues += "}"

	variables := map[string]interface{}{
		"boardId":      boardID,
		"itemName":     taskName,
		"columnValues": columnValues,
	}

	resp, err := c.ExecuteQuery(query, variables)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create task: %w", err)
	}

	if len(resp.Errors) > 0 {
		return 0, nil, fmt.Errorf("failed to create task: %v", resp.Errors)
	}

	fmt.Printf("✅ Task %s created\n", resp.Data)

	// Parse the response to get the task ID
	var createResult struct {
		CreateItem struct {
			ID string `json:"id"`
		} `json:"create_item"`
	}

	if err := json.Unmarshal(resp.Data, &createResult); err != nil {
		fmt.Printf("Warning: Could not parse created task ID: %v\n", err)
		return 0, nil, fmt.Errorf("failed to parse created task ID: %v", err)
	}

	// Fetch the newly created task and add it to cache
	if createResult.CreateItem.ID != "" {
		localId, task, err := c.fetchAndCacheNewTask(boardID, createResult.CreateItem.ID)
		if err != nil {
			fmt.Printf("Warning: Could not fetch and cache new task: %v\n", err)
		}
		return localId, task, nil
	}

	return 0, nil, fmt.Errorf("failed to create task: %v", resp.Errors)
}

// GetTaskByID retrieves a specific task by ID
func (c *Client) GetTaskByID(taskID string) (*Task, error) {
	query := `
		query GetTask($itemId: ID!) {
			items(ids: [$itemId]) {
				id
				name
				column_values {
					id
					text
					value
				}
				updated_at
			}
		}
	`

	variables := map[string]interface{}{
		"itemId": taskID,
	}

	resp, err := c.ExecuteQuery(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Items []Item `json:"items"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("task not found")
	}

	task := Task{
		ID:        result.Items[0].ID,
		Name:      result.Items[0].Name,
		UpdatedAt: result.Items[0].UpdatedAt,
	}
	for _, cv := range result.Items[0].ColumnValues {
		if strings.Contains(strings.ToLower(cv.ID), "status") && cv.Text != "" {
			task.Status = Status(cv.Text)
		}
		if strings.Contains(strings.ToLower(cv.ID), "priority") && cv.Text != "" {
			task.Priority = Priority(cv.Text)
		}
		if strings.Contains(strings.ToLower(cv.ID), "type") && cv.Text != "" {
			task.Type = Type(cv.Text)
		}
		if strings.Contains(strings.ToLower(cv.ID), "sprint") && cv.Text != "" {
			task.Sprint = Sprint(cv.Text)
		}
		if strings.Contains(strings.ToLower(cv.ID), "user_name") && cv.Text != "" {
			task.UserName = cv.Text
		}
		if strings.Contains(strings.ToLower(cv.ID), "user_email") && cv.Text != "" {
			task.UserEmail = cv.Text
		}
	}

	return &task, nil
}

// fetchAndCacheNewTask fetches a newly created task and adds it to the cache
func (c *Client) fetchAndCacheNewTask(boardID, taskID string) (int, *Task, error) {
	// Get the task details
	task, err := c.GetTaskByID(taskID)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to fetch task: %w", err)
	}

	// Load existing cache
	dataStore := NewDataStore()
	localId, err := dataStore.StoreTaskRequest(boardID, *task)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to store task: %w", err)
	}

	fmt.Printf("📝 Task %s added to local cache with ID %d\n", task.Name, localId)
	return localId, task, nil
}

// GetUserInfo retrieves the current user's information
func (c *Client) GetUserInfo() (*User, error) {
	query := `
		query GetUserInfo {
			me {
				id
				name
				email
				title
				photo_small
				enabled
			}
		}
	`

	resp, err := c.ExecuteQuery(query, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Me User `json:"me"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	return &result.Me, nil
}

// Helper functions for sorting
func getSortableStatus(task Task) int {
	status := strings.ToLower(string(task.Status))
	switch {
	case strings.Contains(status, "done"):
		return 1
	case strings.Contains(status, "in progress"):
		return 2
	case strings.Contains(status, "stuck"):
		return 3
	case strings.Contains(status, "waiting for review"):
		return 4
	case strings.Contains(status, "ready for testing"):
		return 5
	case strings.Contains(status, "removed"):
		return 6
	default:
		return 7
	}
}

func getSortablePriority(task Task) int {
	priority := strings.ToLower(string(task.Priority))
	switch {
	case strings.Contains(priority, "critical"):
		return 1
	case strings.Contains(priority, "high"):
		return 2
	case strings.Contains(priority, "medium"):
		return 3
	case strings.Contains(priority, "low"):
		return 4
	default:
		return 5
	}
}

func getSortableType(task Task) int {
	taskType := strings.ToLower(string(task.Type))
	switch {
	case strings.Contains(taskType, "bug"):
		return 1
	case strings.Contains(taskType, "feature"):
		return 2
	case strings.Contains(taskType, "test"):
		return 3
	case strings.Contains(taskType, "security"):
		return 4
	case strings.Contains(taskType, "quality"):
		return 5
	case strings.Contains(taskType, "other"):
		return 6
	default:
		return 7
	}
}
