package monday

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Client is a Monday.com GraphQL API client.
type Client struct {
	token      string
	baseURL    string
	apiVersion string
	httpClient *http.Client
}

// NewClient builds a client. baseURL/apiVersion fall back to sane defaults when empty.
func NewClient(token, baseURL, apiVersion string, timeout int) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if apiVersion == "" {
		apiVersion = defaultAPIVersion
	}
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return &Client{
		token:      token,
		baseURL:    baseURL,
		apiVersion: apiVersion,
		httpClient: &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}
}

// Rule is a single server-side filter rule for items_page(query_params:).
type Rule struct {
	ColumnID     string `json:"column_id"`
	CompareValue any    `json:"compare_value"`
	Operator     string `json:"operator"`
}

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

func (r *graphQLResponse) errorString() string {
	msgs := make([]string, 0, len(r.Errors))
	for _, e := range r.Errors {
		msgs = append(msgs, e.Message)
	}
	return strings.Join(msgs, "; ")
}

// execute runs a GraphQL operation, checking HTTP status and GraphQL errors,
// retrying once on a 429 rate limit.
func (c *Client) execute(query string, variables map[string]any) (json.RawMessage, error) {
	payload, err := json.Marshal(graphQLRequest{Query: query, Variables: variables})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	const maxAttempts = 2
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.baseURL, bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", c.token)
		req.Header.Set("API-Version", c.apiVersion)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request to Monday.com failed: %w", err)
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("failed to read response: %w", readErr)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			if attempt < maxAttempts {
				time.Sleep(retryAfter(resp.Header.Get("Retry-After")))
				continue
			}
			return nil, fmt.Errorf("rate limited by Monday.com (429) — try again shortly")
		}
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return nil, fmt.Errorf("authentication failed (HTTP %d) — check your token with `mon login`", resp.StatusCode)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Monday.com returned HTTP %d: %s", resp.StatusCode, snippet(body))
		}

		var gr graphQLResponse
		if err := json.Unmarshal(body, &gr); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, snippet(body))
		}
		if len(gr.Errors) > 0 {
			return nil, fmt.Errorf("Monday.com API error: %s", gr.errorString())
		}
		return gr.Data, nil
	}
	return nil, fmt.Errorf("rate limited by Monday.com (429) — try again shortly")
}

func retryAfter(header string) time.Duration {
	secs, err := strconv.Atoi(strings.TrimSpace(header))
	if err != nil || secs <= 0 {
		secs = 2
	}
	if secs > 30 {
		secs = 30
	}
	return time.Duration(secs) * time.Second
}

func snippet(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 200 {
		return s[:200] + "…"
	}
	return s
}

// Me returns the authenticated user.
func (c *Client) Me() (*User, error) {
	data, err := c.execute(`query { me { id name email title } }`, nil)
	if err != nil {
		return nil, err
	}
	var r struct {
		Me User `json:"me"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("failed to parse user: %w", err)
	}
	return &r.Me, nil
}

// ListBoards lists active boards, optionally scoped to workspaces, paginating fully.
func (c *Client) ListBoards(workspaceIDs []string) ([]Board, error) {
	const query = `
		query($limit: Int!, $page: Int!, $wsIds: [ID!]) {
			boards(limit: $limit, page: $page, state: active, order_by: used_at, workspace_ids: $wsIds) {
				id
				name
				board_kind
				item_terminology
				items_count
				workspace { id name }
			}
		}`
	var all []Board
	const pageSize = 100
	for page := 1; ; page++ {
		vars := map[string]any{"limit": pageSize, "page": page}
		if len(workspaceIDs) > 0 {
			vars["wsIds"] = workspaceIDs
		}
		data, err := c.execute(query, vars)
		if err != nil {
			return nil, err
		}
		var r struct {
			Boards []Board `json:"boards"`
		}
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, fmt.Errorf("failed to parse boards: %w", err)
		}
		all = append(all, r.Boards...)
		if len(r.Boards) < pageSize {
			break
		}
	}
	return all, nil
}

// GetBoard fetches a board's metadata and columns (used for role auto-detection).
func (c *Client) GetBoard(boardID string) (*Board, error) {
	const query = `
		query($id: ID!) {
			boards(ids: [$id]) {
				id
				name
				board_kind
				item_terminology
				columns { id title type settings_str }
			}
		}`
	data, err := c.execute(query, map[string]any{"id": boardID})
	if err != nil {
		return nil, err
	}
	var r struct {
		Boards []Board `json:"boards"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("failed to parse board: %w", err)
	}
	if len(r.Boards) == 0 {
		return nil, fmt.Errorf("board %s not found (or not accessible with this token)", boardID)
	}
	return &r.Boards[0], nil
}

type itemsPage struct {
	Cursor string `json:"cursor"`
	Items  []Item `json:"items"`
}

// FetchItems returns board items filtered by server-side rules. limit <= 0 fetches
// all matching items; otherwise it caps the total. ruleOperator is "and"/"or".
func (c *Client) FetchItems(boardID string, rules []Rule, ruleOperator string, limit int) ([]Item, error) {
	const firstQuery = `
		query($id: ID!, $limit: Int!, $qp: ItemsQuery) {
			boards(ids: [$id]) {
				items_page(limit: $limit, query_params: $qp) {
					cursor
					items { id name updated_at column_values { id text } }
				}
			}
		}`
	const nextQuery = `
		query($cursor: String!, $limit: Int!) {
			next_items_page(cursor: $cursor, limit: $limit) {
				cursor
				items { id name updated_at column_values { id text } }
			}
		}`

	pageSize := 100
	if limit > 0 && limit < pageSize {
		pageSize = limit
	}

	vars := map[string]any{"id": boardID, "limit": pageSize}
	if len(rules) > 0 {
		if ruleOperator == "" {
			ruleOperator = "and"
		}
		vars["qp"] = map[string]any{"rules": rules, "operator": ruleOperator}
	}
	data, err := c.execute(firstQuery, vars)
	if err != nil {
		return nil, err
	}
	var first struct {
		Boards []struct {
			ItemsPage itemsPage `json:"items_page"`
		} `json:"boards"`
	}
	if err := json.Unmarshal(data, &first); err != nil {
		return nil, fmt.Errorf("failed to parse items: %w", err)
	}
	if len(first.Boards) == 0 {
		return nil, fmt.Errorf("board %s not found", boardID)
	}

	items := first.Boards[0].ItemsPage.Items
	cursor := first.Boards[0].ItemsPage.Cursor
	for cursor != "" && (limit <= 0 || len(items) < limit) {
		data, err := c.execute(nextQuery, map[string]any{"cursor": cursor, "limit": pageSize})
		if err != nil {
			return nil, err
		}
		var next struct {
			NextItemsPage itemsPage `json:"next_items_page"`
		}
		if err := json.Unmarshal(data, &next); err != nil {
			return nil, fmt.Errorf("failed to parse items page: %w", err)
		}
		items = append(items, next.NextItemsPage.Items...)
		cursor = next.NextItemsPage.Cursor
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

// GetItem fetches a single item by ID.
func (c *Client) GetItem(itemID string) (*Item, error) {
	const query = `
		query($id: ID!) {
			items(ids: [$id]) { id name updated_at column_values { id text } }
		}`
	data, err := c.execute(query, map[string]any{"id": itemID})
	if err != nil {
		return nil, err
	}
	var r struct {
		Items []Item `json:"items"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("failed to parse item: %w", err)
	}
	if len(r.Items) == 0 {
		return nil, fmt.Errorf("item %s not found", itemID)
	}
	return &r.Items[0], nil
}

// CreateItem creates an item. columnValues maps column IDs to Monday value objects
// (e.g. {"index": 1} for status). It is JSON-encoded safely before sending.
func (c *Client) CreateItem(boardID, name string, columnValues map[string]any) (string, error) {
	const mutation = `
		mutation($boardId: ID!, $name: String!, $cols: JSON!) {
			create_item(board_id: $boardId, item_name: $name, column_values: $cols) { id }
		}`
	colsJSON, err := json.Marshal(columnValues)
	if err != nil {
		return "", fmt.Errorf("failed to encode column values: %w", err)
	}
	data, err := c.execute(mutation, map[string]any{
		"boardId": boardID,
		"name":    name,
		"cols":    string(colsJSON),
	})
	if err != nil {
		return "", err
	}
	var r struct {
		CreateItem struct {
			ID string `json:"id"`
		} `json:"create_item"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return "", fmt.Errorf("failed to parse created item: %w", err)
	}
	return r.CreateItem.ID, nil
}

// ChangeColumnValues updates one or more columns on an item.
func (c *Client) ChangeColumnValues(boardID, itemID string, columnValues map[string]any) error {
	const mutation = `
		mutation($boardId: ID!, $itemId: ID!, $cols: JSON!) {
			change_multiple_column_values(board_id: $boardId, item_id: $itemId, column_values: $cols) { id }
		}`
	colsJSON, err := json.Marshal(columnValues)
	if err != nil {
		return fmt.Errorf("failed to encode column values: %w", err)
	}
	_, err = c.execute(mutation, map[string]any{
		"boardId": boardID,
		"itemId":  itemID,
		"cols":    string(colsJSON),
	})
	return err
}
