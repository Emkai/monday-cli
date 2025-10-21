package monday

import (
	"fmt"
)

// BoardService handles board-related operations
type BoardService struct {
	client *Client
}

// NewBoardService creates a new board service
func NewBoardService(client *Client) *BoardService {
	return &BoardService{client: client}
}

// GetBoardByID retrieves a specific board with all its data
func (bs *BoardService) GetBoardByID(boardID string) (*Board, error) {
	board, err := bs.client.GetBoard(boardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get board %s: %w", boardID, err)
	}

	return board, nil
}
