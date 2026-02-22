package domain

import (
	"time"

	"github.com/google/uuid"
)

// AgentScorecard represents a periodic performance scorecard for an agent
type AgentScorecard struct {
	ID          uuid.UUID        `json:"id"`
	ProjectID   uuid.UUID        `json:"projectId"`
	AgentName   string           `json:"agentName"`
	Period      string           `json:"period"` // weekly, monthly
	PeriodStart time.Time        `json:"periodStart"`
	PeriodEnd   time.Time        `json:"periodEnd"`
	Metrics     ScorecardMetrics `json:"metrics"`
	Trends      ScorecardTrends  `json:"trends"`
	Grade       string           `json:"grade"` // A, B, C, D, F
	Summary     string           `json:"summary"`
	CreatedAt   time.Time        `json:"createdAt"`
}

// ScorecardMetrics holds the computed metrics for a scorecard period
type ScorecardMetrics struct {
	TotalTraces       int      `json:"totalTraces"`
	SuccessRate       float64  `json:"successRate"`
	AvgLatencyMs      float64  `json:"avgLatencyMs"`
	P95LatencyMs      float64  `json:"p95LatencyMs"`
	TotalCostCents    int64    `json:"totalCostCents"`
	CostPerTrace      float64  `json:"costPerTrace"`
	ErrorRate         float64  `json:"errorRate"`
	AvgTokensPerTrace int      `json:"avgTokensPerTrace"`
	UserSatisfaction  *float64 `json:"userSatisfaction,omitempty"`
}

// ScorecardTrends holds period-over-period trend deltas
type ScorecardTrends struct {
	SuccessRateDelta float64 `json:"successRateDelta"`
	LatencyDelta     float64 `json:"latencyDelta"`
	CostDelta        float64 `json:"costDelta"`
	ErrorRateDelta   float64 `json:"errorRateDelta"`
	VolumeChange     float64 `json:"volumeChange"`
}

// ScorecardConfig configures automatic scorecard generation and delivery
type ScorecardConfig struct {
	ProjectID    uuid.UUID `json:"projectId"`
	AgentName    string    `json:"agentName"`
	Period       string    `json:"period"` // weekly, monthly
	Recipients   []string  `json:"recipients"`
	SlackWebhook string    `json:"slackWebhook,omitempty"`
	Enabled      bool      `json:"enabled"`
}

// ScorecardInput represents input for generating a scorecard on demand
type ScorecardInput struct {
	AgentName string `json:"agentName"`
	Period    string `json:"period"` // weekly, monthly
}
