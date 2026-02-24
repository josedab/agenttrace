package domain

import (
	"time"

	"github.com/google/uuid"
)

// SecurityRiskType represents the type of security risk detected
type SecurityRiskType string

const (
	SecurityRiskTypePromptInjection     SecurityRiskType = "prompt_injection"
	SecurityRiskTypePIILeakage          SecurityRiskType = "pii_leakage"
	SecurityRiskTypeCodeExecution       SecurityRiskType = "code_execution"
	SecurityRiskTypeExcessivePermissions SecurityRiskType = "excessive_permissions"
	SecurityRiskTypeSupplyChain         SecurityRiskType = "supply_chain"
	SecurityRiskTypeDataExfiltration    SecurityRiskType = "data_exfiltration"
)

// IsValid checks if the security risk type is valid
func (t SecurityRiskType) IsValid() bool {
	switch t {
	case SecurityRiskTypePromptInjection, SecurityRiskTypePIILeakage, SecurityRiskTypeCodeExecution, SecurityRiskTypeExcessivePermissions, SecurityRiskTypeSupplyChain, SecurityRiskTypeDataExfiltration:
		return true
	}
	return false
}

// SecuritySeverity represents the severity level of a security finding
type SecuritySeverity string

const (
	SecuritySeverityLow      SecuritySeverity = "low"
	SecuritySeverityMedium   SecuritySeverity = "medium"
	SecuritySeverityHigh     SecuritySeverity = "high"
	SecuritySeverityCritical SecuritySeverity = "critical"
)

// IsValid checks if the security severity is valid
func (s SecuritySeverity) IsValid() bool {
	switch s {
	case SecuritySeverityLow, SecuritySeverityMedium, SecuritySeverityHigh, SecuritySeverityCritical:
		return true
	}
	return false
}

// ScanAction represents the action to take when a security finding is detected
type ScanAction string

const (
	ScanActionLog        ScanAction = "log"
	ScanActionWarn       ScanAction = "warn"
	ScanActionBlock      ScanAction = "block"
	ScanActionQuarantine ScanAction = "quarantine"
)

// IsValid checks if the scan action is valid
func (a ScanAction) IsValid() bool {
	switch a {
	case ScanActionLog, ScanActionWarn, ScanActionBlock, ScanActionQuarantine:
		return true
	}
	return false
}

// SecurityScanResult represents the result of a security scan on a trace
type SecurityScanResult struct {
	ID             uuid.UUID         `json:"id"`
	ProjectID      uuid.UUID         `json:"projectId"`
	TraceID        uuid.UUID         `json:"traceId"`
	ObservationID  *uuid.UUID        `json:"observationId,omitempty"`
	Findings       []SecurityFinding `json:"findings"`
	OverallRisk    SecuritySeverity  `json:"overallRisk"`
	ScannedAt      time.Time         `json:"scannedAt"`
	ScanDurationMs int64             `json:"scanDurationMs"`
}

// SecurityFinding represents a single security finding within a scan
type SecurityFinding struct {
	ID             uuid.UUID        `json:"id"`
	Type           SecurityRiskType `json:"type"`
	Severity       SecuritySeverity `json:"severity"`
	Title          string           `json:"title"`
	Description    string           `json:"description"`
	Evidence       string           `json:"evidence"`
	Location       string           `json:"location"`
	Recommendation string           `json:"recommendation"`
	FalsePositive  bool             `json:"falsePositive"`
	AcknowledgedAt *time.Time       `json:"acknowledgedAt,omitempty"`
}

// SecurityPolicy represents a security policy for a project
type SecurityPolicy struct {
	ID              uuid.UUID      `json:"id"`
	ProjectID       uuid.UUID      `json:"projectId"`
	Name            string         `json:"name"`
	Enabled         bool           `json:"enabled"`
	Rules           []SecurityRule `json:"rules"`
	Action          ScanAction     `json:"action"`
	ExcludePatterns []string       `json:"excludePatterns,omitempty"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}

// SecurityRule represents a single rule within a security policy
type SecurityRule struct {
	Type          SecurityRiskType `json:"type"`
	Enabled       bool             `json:"enabled"`
	Severity      SecuritySeverity `json:"severity"`
	CustomPattern string           `json:"customPattern,omitempty"`
	Action        ScanAction       `json:"action"`
}

// SecurityDashboard represents the security overview dashboard for a project
type SecurityDashboard struct {
	ProjectID           uuid.UUID                  `json:"projectId"`
	TotalScans          int64                      `json:"totalScans"`
	FindingsCount       int                        `json:"findingsCount"`
	BySeverity          map[SecuritySeverity]int   `json:"bySeverity"`
	ByType              map[SecurityRiskType]int   `json:"byType"`
	TopVulnerableTraces []string                   `json:"topVulnerableTraces"`
	OWASPCoverage       map[string]bool            `json:"owaspCoverage"`
	LastScanAt          *time.Time                 `json:"lastScanAt,omitempty"`
	Policies            []SecurityPolicy           `json:"policies"`
}
