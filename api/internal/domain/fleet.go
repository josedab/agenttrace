package domain

import (
	"time"

	"github.com/google/uuid"
)

// FleetAgentStatus represents the health status of a fleet agent
type FleetAgentStatus string

const (
	FleetAgentHealthy  FleetAgentStatus = "healthy"
	FleetAgentDegraded FleetAgentStatus = "degraded"
	FleetAgentDown     FleetAgentStatus = "down"
)

// FleetPolicyType represents the type of fleet policy
type FleetPolicyType string

const (
	PolicyCostLimit          FleetPolicyType = "cost_limit"
	PolicyQualityMinimum     FleetPolicyType = "quality_minimum"
	PolicyUptimeRequirement  FleetPolicyType = "uptime_requirement"
)

// FleetPolicyScope represents the scope of a fleet policy
type FleetPolicyScope string

const (
	ScopeFleet   FleetPolicyScope = "fleet"
	ScopeProject FleetPolicyScope = "project"
	ScopeAgent   FleetPolicyScope = "agent"
)

// FleetDashboard represents the fleet management dashboard
type FleetDashboard struct {
	TotalAgents   int            `json:"totalAgents"`
	HealthyCount  int            `json:"healthyCount"`
	DegradedCount int            `json:"degradedCount"`
	DownCount     int            `json:"downCount"`
	TotalCost     float64        `json:"totalCost"`
	TotalTraces   int            `json:"totalTraces"`
	Agents        []FleetAgent   `json:"agents"`
	Policies      []FleetPolicy  `json:"policies"`
}

// FleetAgent represents an agent in the fleet
type FleetAgent struct {
	Name        string           `json:"name"`
	ProjectID   uuid.UUID        `json:"projectId"`
	Status      FleetAgentStatus `json:"status"`
	Traces      int              `json:"traces"`
	Cost        float64          `json:"cost"`
	LastActive  time.Time        `json:"lastActive"`
	Version     string           `json:"version"`
	HealthScore float64          `json:"healthScore"`
}

// FleetPolicy represents a fleet management policy
type FleetPolicy struct {
	ID        uuid.UUID        `json:"id"`
	Name      string           `json:"name"`
	Type      FleetPolicyType  `json:"type"`
	Config    map[string]any   `json:"config"`
	Scope     FleetPolicyScope `json:"scope"`
	Enabled   bool             `json:"enabled"`
	CreatedAt time.Time        `json:"createdAt"`
}

// FleetPolicyInput represents input for creating a fleet policy
type FleetPolicyInput struct {
	Name    string           `json:"name" validate:"required"`
	Type    FleetPolicyType  `json:"type" validate:"required"`
	Config  map[string]any   `json:"config"`
	Scope   FleetPolicyScope `json:"scope"`
	Enabled bool             `json:"enabled"`
}

// BulkConfigUpdate represents a bulk configuration update for agents
type BulkConfigUpdate struct {
	AgentNames []string       `json:"agentNames" validate:"required"`
	Config     map[string]any `json:"config" validate:"required"`
	Note       string         `json:"note,omitempty"`
}

// ScalingAction represents the recommended scaling action
type ScalingAction string

const (
	ScaleUp   ScalingAction = "scale_up"
	ScaleDown ScalingAction = "scale_down"
	Maintain  ScalingAction = "maintain"
)

// FleetScalingRecommendation represents a scaling recommendation for an agent
type FleetScalingRecommendation struct {
	AgentName         string        `json:"agentName"`
	CurrentLoad       float64       `json:"currentLoad"`
	RecommendedAction ScalingAction `json:"recommendedAction"`
	Reason            string        `json:"reason"`
	EstimatedSavings  float64       `json:"estimatedSavings"`
}
