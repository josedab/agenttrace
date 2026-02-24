package domain

import "time"

// TeamDashboard represents the team intelligence dashboard
type TeamDashboard struct {
	TotalCost    float64        `json:"totalCost"`
	TotalTraces  int            `json:"totalTraces"`
	Agents       []AgentUsage   `json:"agents"`
	Members      []MemberStats  `json:"members"`
	CostPerDev   float64        `json:"costPerDev"`
	QualityTrend []float64      `json:"qualityTrend"`
	ROI          TeamROICalculation `json:"roi"`
}

// MemberStats represents individual team member statistics
type MemberStats struct {
	Name       string  `json:"name"`
	Traces     int     `json:"traces"`
	Cost       float64 `json:"cost"`
	AvgQuality float64 `json:"avgQuality"`
}

// AgentUsage represents usage statistics for a specific agent
type AgentUsage struct {
	Name        string  `json:"name"`
	Traces      int     `json:"traces"`
	Cost        float64 `json:"cost"`
	Tokens      int64   `json:"tokens"`
	SuccessRate float64 `json:"successRate"`
}

// TeamROICalculation represents the return on investment calculation for team intelligence
type TeamROICalculation struct {
	HoursSaved    float64 `json:"hoursSaved"`
	HourlyRate    float64 `json:"hourlyRate"`
	AgentCosts    float64 `json:"agentCosts"`
	PlatformCosts float64 `json:"platformCosts"`
	NetROI        float64 `json:"netROI"`
	ROIPercent    float64 `json:"roiPercent"`
}

// TeamDashboardFilter represents filters for the team dashboard
type TeamDashboardFilter struct {
	ProjectID string     `json:"projectId"`
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
}
