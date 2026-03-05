package domain

import (
	"time"

	"github.com/google/uuid"
)

// DashboardConfig represents a real-time streaming dashboard configuration
type DashboardConfig struct {
	ID              uuid.UUID         `json:"id"`
	ProjectID       uuid.UUID         `json:"projectId"`
	Name            string            `json:"name"`
	Layout          []StreamingDashboardWidget `json:"layout"`
	RefreshInterval int               `json:"refreshInterval"` // seconds
	IsDefault       bool              `json:"isDefault"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}

// StreamingDashboardWidget represents a widget in the streaming dashboard layout
type StreamingDashboardWidget struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"` // active_sessions, cost_ticker, token_stream, error_feed, progress_bar
	Position WidgetPosition `json:"position"`
	Config   map[string]any `json:"config,omitempty"`
}

// WidgetPosition represents the position and size of a widget
type WidgetPosition struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// DashboardSnapshot represents a point-in-time snapshot of dashboard metrics
type DashboardSnapshot struct {
	ProjectID      uuid.UUID    `json:"projectId"`
	ActiveSessions int          `json:"activeSessions"`
	TotalCost      float64      `json:"totalCost"`
	TotalTokens    int64        `json:"totalTokens"`
	ErrorCount     int          `json:"errorCount"`
	ActiveStreams  []LiveMetrics `json:"activeStreams"`
	TopModels      []ModelUsage `json:"topModels"`
	Timestamp      time.Time    `json:"timestamp"`
}

// ModelUsage represents usage statistics for a specific model
type ModelUsage struct {
	Model        string  `json:"model"`
	RequestCount int     `json:"requestCount"`
	TotalTokens  int64   `json:"totalTokens"`
	TotalCost    float64 `json:"totalCost"`
}
