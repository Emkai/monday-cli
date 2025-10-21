package monday

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// BoardService handles board-related operations
type BoardService struct {
	client *Client
}

// NewBoardService creates a new board service
func NewBoardService(client *Client) *BoardService {
	return &BoardService{client: client}
}

// GetAllBoards retrieves all boards with optional filtering
func (bs *BoardService) GetAllBoards() ([]Board, error) {
	boards, err := bs.client.GetBoards()
	if err != nil {
		return nil, fmt.Errorf("failed to get boards: %w", err)
	}

	// Sort boards by name
	sort.Slice(boards, func(i, j int) bool {
		return strings.ToLower(boards[i].Name) < strings.ToLower(boards[j].Name)
	})

	return boards, nil
}

// GetBoardByID retrieves a specific board with all its data
func (bs *BoardService) GetBoardByID(boardID string) (*Board, error) {
	board, err := bs.client.GetBoard(boardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get board %s: %w", boardID, err)
	}

	return board, nil
}

// SearchBoards searches boards by name
func (bs *BoardService) SearchBoards(query string) ([]Board, error) {
	boards, err := bs.GetAllBoards()
	if err != nil {
		return nil, err
	}

	var results []Board
	query = strings.ToLower(query)

	for _, board := range boards {
		if strings.Contains(strings.ToLower(board.Name), query) ||
			strings.Contains(strings.ToLower(board.Description), query) {
			results = append(results, board)
		}
	}

	return results, nil
}

// ItemService handles item-related operations
type ItemService struct {
	client *Client
}

// NewItemService creates a new item service
func NewItemService(client *Client) *ItemService {
	return &ItemService{client: client}
}

// CreateItem creates a new item in a board
func (is *ItemService) CreateItem(boardID, itemName string, columnValues map[string]string) (*Item, error) {
	item, err := is.client.CreateItem(boardID, itemName, columnValues)
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	return item, nil
}

// UpdateItem updates an item's column values
func (is *ItemService) UpdateItem(itemID string, columnValues map[string]string) (*Item, error) {
	item, err := is.client.UpdateItem(itemID, columnValues)
	if err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	return item, nil
}

// GetItemColumnValue retrieves a specific column value for an item
func (is *ItemService) GetItemColumnValue(item Item, columnID string) *ColumnValue {
	for _, cv := range item.ColumnValues {
		if cv.ID == columnID {
			return &cv
		}
	}
	return nil
}

// FilterItemsByColumn filters items by a specific column value
func (is *ItemService) FilterItemsByColumn(items []Item, columnID, value string) []Item {
	var results []Item
	value = strings.ToLower(value)

	for _, item := range items {
		columnValue := is.GetItemColumnValue(item, columnID)
		if columnValue != nil && strings.Contains(strings.ToLower(columnValue.Text), value) {
			results = append(results, item)
		}
	}

	return results
}

// SortItemsByColumn sorts items by a specific column value
func (is *ItemService) SortItemsByColumn(items []Item, columnID string, ascending bool) []Item {
	sort.Slice(items, func(i, j int) bool {
		colI := is.GetItemColumnValue(items[i], columnID)
		colJ := is.GetItemColumnValue(items[j], columnID)

		textI := ""
		textJ := ""

		if colI != nil {
			textI = colI.Text
		}
		if colJ != nil {
			textJ = colJ.Text
		}

		if ascending {
			return strings.ToLower(textI) < strings.ToLower(textJ)
		}
		return strings.ToLower(textI) > strings.ToLower(textJ)
	})

	return items
}

// WorkspaceService handles workspace-related operations
type WorkspaceService struct {
	client *Client
}

// NewWorkspaceService creates a new workspace service
func NewWorkspaceService(client *Client) *WorkspaceService {
	return &WorkspaceService{client: client}
}

// GetWorkspaces retrieves all workspaces
func (ws *WorkspaceService) GetWorkspaces() ([]Workspace, error) {
	query := `
		query {
			workspaces {
				id
				name
				description
				kind
			}
		}
	`

	resp, err := ws.client.ExecuteQuery(query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces: %w", err)
	}

	var result struct {
		Workspaces []Workspace `json:"workspaces"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workspaces: %w", err)
	}

	return result.Workspaces, nil
}

// AnalyticsService handles analytics and reporting
type AnalyticsService struct {
	client *Client
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(client *Client) *AnalyticsService {
	return &AnalyticsService{client: client}
}

// GetBoardActivity retrieves activity for a specific board
func (as *AnalyticsService) GetBoardActivity(boardID string, limit int) ([]ActivityLog, error) {
	query := `
		query GetBoardActivity($boardId: ID!, $limit: Int!) {
			boards(ids: [$boardId]) {
				activity_logs(limit: $limit) {
					id
					event
					created_at
					data
				}
			}
		}
	`

	variables := map[string]interface{}{
		"boardId": boardID,
		"limit":   limit,
	}

	resp, err := as.client.ExecuteQuery(query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get board activity: %w", err)
	}

	var result struct {
		Boards []struct {
			ActivityLogs []ActivityLog `json:"activity_logs"`
		} `json:"boards"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal activity logs: %w", err)
	}

	if len(result.Boards) == 0 {
		return []ActivityLog{}, nil
	}

	return result.Boards[0].ActivityLogs, nil
}

// GetItemCountByStatus returns a count of items by status
func (as *AnalyticsService) GetItemCountByStatus(board *Board, statusColumnID string) map[string]int {
	statusCount := make(map[string]int)

	for _, item := range board.Items {
		statusValue := as.getItemStatus(item, statusColumnID)
		statusCount[statusValue]++
	}

	return statusCount
}

// getItemStatus gets the status value for an item
func (as *AnalyticsService) getItemStatus(item Item, statusColumnID string) string {
	for _, cv := range item.ColumnValues {
		if cv.ID == statusColumnID {
			return cv.Text
		}
	}
	return "Unknown"
}

// GetRecentItems returns items updated in the last N days
func (as *AnalyticsService) GetRecentItems(board *Board, days int) []Item {
	var recentItems []Item
	cutoff := time.Now().AddDate(0, 0, -days)

	for _, item := range board.Items {
		if item.UpdatedAt.After(cutoff) {
			recentItems = append(recentItems, item)
		}
	}

	// Sort by update time, most recent first
	sort.Slice(recentItems, func(i, j int) bool {
		return recentItems[i].UpdatedAt.After(recentItems[j].UpdatedAt)
	})

	return recentItems
}
