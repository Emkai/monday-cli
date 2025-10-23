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
func (c *Client) GetBoardItemsByOwner(boardID, ownerEmail string) ([]Item, error) {
	// First, get the board to find the owner column ID
	board, err := c.GetBoard(boardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
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
		return nil, fmt.Errorf("owner column not found in board")
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
			return nil, err
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
			return nil, fmt.Errorf("failed to unmarshal board items: %w", err)
		}

		if len(result.Boards) == 0 {
			return nil, fmt.Errorf("board not found")
		}

		// Filter items by owner and add to collection
		for _, item := range result.Boards[0].ItemsPage.Items {
			for _, cv := range item.ColumnValues {
				if cv.ID == ownerColumnID && strings.Contains(strings.ToLower(cv.Text), strings.ToLower(ownerEmail)) {
					allItems = append(allItems, item)
					break
				}
			}
		}

		// Check if we have enough items or no more pages
		cursor = result.Boards[0].ItemsPage.Cursor
		if cursor == "" || len(result.Boards[0].ItemsPage.Items) < limit {
			break
		}
		fmt.Printf("Fetching next page... currently %d items for %s\n", len(allItems), ownerEmail)
	}

	allItems = OrderItems(allItems)
	return allItems, nil
}

func OrderItems(items []Item) []Item {
	// Sort items by status, priority, then type
	sort.Slice(items, func(i, j int) bool {
		statusI := getSortableStatus(items[i])
		statusJ := getSortableStatus(items[j])

		// First sort by status
		if statusI != statusJ {
			return statusI < statusJ
		}

		// Then by priority
		priorityI := getSortablePriority(items[i])
		priorityJ := getSortablePriority(items[j])
		if priorityI != priorityJ {
			return priorityI < priorityJ
		}

		// Finally by type
		typeI := getSortableType(items[i])
		typeJ := getSortableType(items[j])
		return typeI < typeJ

	})
	return items
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

func (c *Client) CreateTask(boardID, ownerEmail, taskName, status, priority, taskType string) error {
	// First get the user ID for the owner email
	userQuery := `
		query GetUser($emails: [String!]!) {
			users(emails: $emails) {
				id
			}
		}
	`

	userVars := map[string]interface{}{
		"emails": []string{ownerEmail},
	}

	userResp, err := c.ExecuteQuery(userQuery, userVars)
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	if len(userResp.Errors) > 0 {
		return fmt.Errorf("failed to get user ID: %v", userResp.Errors)
	}

	var userData struct {
		Users []struct {
			ID string `json:"id"`
		} `json:"users"`
	}
	if err := json.Unmarshal(userResp.Data, &userData); err != nil {
		return fmt.Errorf("failed to parse user data: %w", err)
	}

	if len(userData.Users) == 0 {
		return fmt.Errorf("user not found for email: %s", ownerEmail)
	}

	userID := userData.Users[0].ID

	// Get board to find column IDs
	board, err := c.GetBoard(boardID)
	if err != nil {
		return fmt.Errorf("failed to get board: %w", err)
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
		return fmt.Errorf("failed to create task: %w", err)
	}

	if len(resp.Errors) > 0 {
		return fmt.Errorf("failed to create task: %v", resp.Errors)
	}

	return nil
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
func getSortableStatus(item Item) int {
	for _, cv := range item.ColumnValues {
		if strings.Contains(strings.ToLower(cv.ID), "status") && cv.Text != "" {
			status := strings.ToLower(cv.Text)
			switch {
			case strings.Contains(status, "done") || strings.Contains(status, "completed"):
				return 1 // Done first
			case strings.Contains(status, "progress") || strings.Contains(status, "in progress"):
				return 2 // In progress second
			case strings.Contains(status, "stuck") || strings.Contains(status, "blocked"):
				return 3 // Stuck third
			case strings.Contains(status, "review"):
				return 4 // Review fourth
			case strings.Contains(status, "testing") || strings.Contains(status, "not started"):
				return 5 // Todo last
			default:
				return 6 // Unknown status last
			}
		}
	}
	return 6 // Default to last if no status found
}

func getSortablePriority(item Item) int {
	for _, cv := range item.ColumnValues {
		if strings.Contains(strings.ToLower(cv.ID), "priority") && cv.Text != "" {
			priority := strings.ToLower(cv.Text)
			switch {
			case strings.Contains(priority, "critical"):
				return 1 // Critical first
			case strings.Contains(priority, "high"):
				return 2 // High second
			case strings.Contains(priority, "medium"):
				return 3 // Medium third
			case strings.Contains(priority, "low"):
				return 4 // Low fourth
			default:
				return 5 // Unknown priority last
			}
		}
	}
	return 5 // Default to last if no priority found
}

func getSortableType(item Item) int {
	for _, cv := range item.ColumnValues {
		if strings.Contains(strings.ToLower(cv.ID), "type") && cv.Text != "" {
			taskType := strings.ToLower(cv.Text)
			switch {
			case strings.Contains(taskType, "bug"):
				return 1 // Bug first
			case strings.Contains(taskType, "feature"):
				return 2 // Feature second
			case strings.Contains(taskType, "test"):
				return 3 // Test third
			case strings.Contains(taskType, "security"):
				return 4 // Security fourth
			case strings.Contains(taskType, "improvement"):
				return 5 // Improvement fifth
			default:
				return 6 // Other types last
			}
		}
	}
	return 6 // Default to last if no type found
}
