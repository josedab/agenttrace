package domain

import (
	"time"

	"github.com/google/uuid"
)

// CanaryStatus represents the status of a canary deployment
type CanaryStatus string

const (
	CanaryStatusPending    CanaryStatus = "pending"
	CanaryStatusRunning    CanaryStatus = "running"
	CanaryStatusPromoting  CanaryStatus = "promoting"
	CanaryStatusCompleted  CanaryStatus = "completed"
	CanaryStatusRolledBack CanaryStatus = "rolled_back"
	CanaryStatusFailed     CanaryStatus = "failed"
)

// CanaryStage represents a traffic percentage stage
type CanaryStage struct {
	Percentage    int           `json:"percentage"`    // 5, 25, 50, 100
	MinDuration   string        `json:"minDuration"`   // minimum time at this stage
	AutoPromote   bool          `json:"autoPromote"`
}

// PromotionCriteria defines when a canary stage can be promoted
type PromotionCriteria struct {
	MinEvalScore    *float64 `json:"minEvalScore,omitempty"`
	MaxCostIncrease *float64 `json:"maxCostIncrease,omitempty"` // percentage
	MaxLatencyMs    *int64   `json:"maxLatencyMs,omitempty"`
	MaxErrorRate    *float64 `json:"maxErrorRate,omitempty"`
	MinSampleSize   int      `json:"minSampleSize"`
}

// CanaryDeployment represents a canary deployment of an agent configuration
type CanaryDeployment struct {
	ID              uuid.UUID         `json:"id"`
	ProjectID       uuid.UUID         `json:"projectId"`
	Name            string            `json:"name"`
	Description     string            `json:"description,omitempty"`
	Status          CanaryStatus      `json:"status"`
	BaselineVersion string            `json:"baselineVersion"`
	CanaryVersion   string            `json:"canaryVersion"`
	Stages          []CanaryStage     `json:"stages"`
	CurrentStage    int               `json:"currentStage"`
	Criteria        PromotionCriteria `json:"criteria"`
	Metrics         CanaryMetrics     `json:"metrics"`
	CreatedBy       uuid.UUID         `json:"createdBy"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
	CompletedAt     *time.Time        `json:"completedAt,omitempty"`
}

// CanaryMetrics represents metrics for baseline and canary versions
type CanaryMetrics struct {
	Baseline CanaryVersionMetrics `json:"baseline"`
	Canary   CanaryVersionMetrics `json:"canary"`
}

// CanaryVersionMetrics represents metrics for a single version in canary deployment
type CanaryVersionMetrics struct {
	RequestCount  int64   `json:"requestCount"`
	AvgLatencyMs  float64 `json:"avgLatencyMs"`
	ErrorRate     float64 `json:"errorRate"`
	AvgEvalScore  float64 `json:"avgEvalScore"`
	TotalCost     float64 `json:"totalCost"`
	P95LatencyMs  float64 `json:"p95LatencyMs"`
}

// CanaryDeploymentInput represents input for creating a canary deployment
type CanaryDeploymentInput struct {
	Name            string            `json:"name" validate:"required"`
	Description     string            `json:"description,omitempty"`
	BaselineVersion string            `json:"baselineVersion" validate:"required"`
	CanaryVersion   string            `json:"canaryVersion" validate:"required"`
	Stages          []CanaryStage     `json:"stages,omitempty"`
	Criteria        *PromotionCriteria `json:"criteria,omitempty"`
}

// CanaryDeploymentList represents a paginated list of deployments
type CanaryDeploymentList struct {
	Deployments []CanaryDeployment `json:"deployments"`
	TotalCount  int                `json:"totalCount"`
}

// ActiveVersion represents the currently active agent version
type ActiveVersion struct {
	Version    string  `json:"version"`
	IsCanary   bool    `json:"isCanary"`
	Percentage int     `json:"percentage"`
	DeploymentID *uuid.UUID `json:"deploymentId,omitempty"`
}
