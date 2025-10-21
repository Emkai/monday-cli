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
	var allItems []Item
	cursor := ""
	limit := 50 // Increased limit to get more items per request

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

		// Add items to our collection
		allItems = append(allItems, result.Boards[0].ItemsPage.Items...)

		// Check if we have more pages
		cursor = result.Boards[0].ItemsPage.Cursor
		if cursor == "" || len(result.Boards[0].ItemsPage.Items) < limit {
			break
		}
	}

	// Filter items by owner
	var filteredItems []Item
	for _, item := range allItems {
		for _, cv := range item.ColumnValues {
			if strings.Contains(strings.ToLower(cv.ID), "owner") &&
				strings.Contains(strings.ToLower(cv.Text), strings.ToLower(ownerEmail)) {
				filteredItems = append(filteredItems, item)
				break
			}
		}
	}

	// Sort items by status, priority, then type
	sort.Slice(filteredItems, func(i, j int) bool {
		// Get status, priority, and type for both items
		statusI := getSortableStatus(filteredItems[i])
		statusJ := getSortableStatus(filteredItems[j])

		// First sort by status
		if statusI != statusJ {
			return statusI < statusJ
		}

		// Then by priority
		priorityI := getSortablePriority(filteredItems[i])
		priorityJ := getSortablePriority(filteredItems[j])
		if priorityI != priorityJ {
			return priorityI < priorityJ
		}

		// Finally by type
		typeI := getSortableType(filteredItems[i])
		typeJ := getSortableType(filteredItems[j])
		return typeI < typeJ
	})

	return filteredItems, nil
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
			case strings.Contains(status, "todo") || strings.Contains(status, "not started"):
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
