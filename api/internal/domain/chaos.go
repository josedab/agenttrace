package domain

import (
	"time"

	"github.com/google/uuid"
)

// ChaosExperimentStatus represents the status of a chaos experiment
type ChaosExperimentStatus string

const (
	ChaosStatusDraft     ChaosExperimentStatus = "draft"
	ChaosStatusRunning   ChaosExperimentStatus = "running"
	ChaosStatusCompleted ChaosExperimentStatus = "completed"
)

// FaultInjectionType represents the type of fault to inject
type FaultInjectionType string

const (
	FaultTypeLatency           FaultInjectionType = "latency"
	FaultTypeError             FaultInjectionType = "error"
	FaultTypeRateLimit         FaultInjectionType = "rate_limit"
	FaultTypeModelDowngrade    FaultInjectionType = "model_downgrade"
	FaultTypeContextTruncation FaultInjectionType = "context_truncation"
)

// FaultConfig holds the configuration for a specific fault injection
type FaultConfig struct {
	LatencyMs       *int     `json:"latencyMs,omitempty"`
	ErrorCode       *int     `json:"errorCode,omitempty"`
	ErrorMessage    string   `json:"errorMessage,omitempty"`
	SubstituteModel string   `json:"substituteModel,omitempty"`
	TruncatePercent *float64 `json:"truncatePercent,omitempty"`
}

// FaultInjection represents a single fault to inject during a chaos experiment
type FaultInjection struct {
	ID       uuid.UUID          `json:"id"`
	Type     FaultInjectionType `json:"type"`
	Target   string             `json:"target"`
	Config   FaultConfig        `json:"config"`
	Duration int                `json:"duration"`
}

// ChaosResults holds the results of a completed chaos experiment
type ChaosResults struct {
	Resilience           float64            `json:"resilience"`
	GracefulDegradation  bool               `json:"gracefulDegradation"`
	ErrorRecovery        bool               `json:"errorRecovery"`
	FallbackBehavior     string             `json:"fallbackBehavior"`
	Metrics              map[string]float64 `json:"metrics"`
}

// ChaosExperiment represents a chaos engineering experiment
type ChaosExperiment struct {
	ID          uuid.UUID             `json:"id"`
	ProjectID   uuid.UUID             `json:"projectId"`
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	Status      ChaosExperimentStatus `json:"status"`
	Faults      []FaultInjection      `json:"faults"`
	Results     *ChaosResults         `json:"results,omitempty"`
	CreatedAt   time.Time             `json:"createdAt"`
	CompletedAt *time.Time            `json:"completedAt,omitempty"`
}

// ResilienceScorecard summarizes an agent's resilience to faults
type ResilienceScorecard struct {
	AgentName       string             `json:"agentName"`
	OverallScore    float64            `json:"overallScore"`
	Scores          map[string]float64 `json:"scores"`
	TestedScenarios int                `json:"testedScenarios"`
	PassedScenarios int                `json:"passedScenarios"`
	Recommendations []string           `json:"recommendations"`
}

// ChaosExperimentInput is the input for creating a chaos experiment
type ChaosExperimentInput struct {
	Name   string           `json:"name"`
	Faults []FaultInjection `json:"faults"`
}
