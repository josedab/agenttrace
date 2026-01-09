package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// Cursor represents a pagination cursor
type Cursor struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"ts"`
	Offset    int       `json:"off,omitempty"`
}

// Encode encodes the cursor to a string
func (c *Cursor) Encode() string {
	data, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeCursor decodes a cursor string
func DecodeCursor(s string) (*Cursor, error) {
	if s == "" {
		return nil, nil
	}

	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor encoding: %w", err)
	}

	var cursor Cursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, fmt.Errorf("invalid cursor format: %w", err)
	}

	return &cursor, nil
}

// NewCursor creates a new cursor from an ID and timestamp
func NewCursor(id string, timestamp time.Time) *Cursor {
	return &Cursor{
		ID:        id,
		Timestamp: timestamp,
	}
}

// NewOffsetCursor creates a new offset-based cursor
func NewOffsetCursor(offset int) *Cursor {
	return &Cursor{
		Offset: offset,
	}
}

// PageInfo contains pagination information
type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *string `json:"startCursor,omitempty"`
	EndCursor       *string `json:"endCursor,omitempty"`
	TotalCount      *int64  `json:"totalCount,omitempty"`
}

// PageParams contains pagination parameters
type PageParams struct {
	First  *int
	Last   *int
	After  *string
	Before *string
	Limit  int
	Offset int
}

// NewPageParams creates new pagination parameters with defaults
func NewPageParams() *PageParams {
	limit := 50
	return &PageParams{
		Limit: limit,
	}
}

// WithFirst sets the first parameter
func (p *PageParams) WithFirst(first int) *PageParams {
	p.First = &first
	p.Limit = first
	return p
}

// WithAfter sets the after cursor
func (p *PageParams) WithAfter(after string) *PageParams {
	p.After = &after
	return p
}

// WithLast sets the last parameter
func (p *PageParams) WithLast(last int) *PageParams {
	p.Last = &last
	p.Limit = last
	return p
}

// WithBefore sets the before cursor
func (p *PageParams) WithBefore(before string) *PageParams {
	p.Before = &before
	return p
}

// GetLimit returns the limit, defaulting to 50
func (p *PageParams) GetLimit() int {
	if p.First != nil {
		return *p.First
	}
	if p.Last != nil {
		return *p.Last
	}
	if p.Limit > 0 {
		return p.Limit
	}
	return 50
}

// GetAfterCursor decodes and returns the after cursor
func (p *PageParams) GetAfterCursor() (*Cursor, error) {
	if p.After == nil {
		return nil, nil
	}
	return DecodeCursor(*p.After)
}

// GetBeforeCursor decodes and returns the before cursor
func (p *PageParams) GetBeforeCursor() (*Cursor, error) {
	if p.Before == nil {
		return nil, nil
	}
	return DecodeCursor(*p.Before)
}

// Connection represents a paginated connection
type Connection[T any] struct {
	Edges      []Edge[T] `json:"edges"`
	PageInfo   PageInfo  `json:"pageInfo"`
	TotalCount int64     `json:"totalCount"`
}

// Edge represents an edge in a connection
type Edge[T any] struct {
	Node   T      `json:"node"`
	Cursor string `json:"cursor"`
}

// NewConnection creates a new connection from items
func NewConnection[T any](items []T, getCursor func(T) string, hasMore bool, totalCount int64) *Connection[T] {
	edges := make([]Edge[T], len(items))
	for i, item := range items {
		edges[i] = Edge[T]{
			Node:   item,
			Cursor: getCursor(item),
		}
	}

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &Connection[T]{
		Edges: edges,
		PageInfo: PageInfo{
			HasNextPage:     hasMore,
			HasPreviousPage: false, // Would need offset tracking
			StartCursor:     startCursor,
			EndCursor:       endCursor,
			TotalCount:      &totalCount,
		},
		TotalCount: totalCount,
	}
}
