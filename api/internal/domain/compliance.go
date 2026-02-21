package domain

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ComplianceRiskLevel represents the EU AI Act risk classification level
type ComplianceRiskLevel string

const (
	ComplianceRiskLevelHigh    ComplianceRiskLevel = "HIGH_RISK"
	ComplianceRiskLevelLimited ComplianceRiskLevel = "LIMITED_RISK"
	ComplianceRiskLevelMinimal ComplianceRiskLevel = "MINIMAL_RISK"
)

// IsValid checks if the compliance risk level is valid
func (r ComplianceRiskLevel) IsValid() bool {
	switch r {
	case ComplianceRiskLevelHigh, ComplianceRiskLevelLimited, ComplianceRiskLevelMinimal:
		return true
	}
	return false
}

// ComplianceStatus represents the compliance assessment status
type ComplianceStatus string

const (
	ComplianceStatusCompliant    ComplianceStatus = "COMPLIANT"
	ComplianceStatusNonCompliant ComplianceStatus = "NON_COMPLIANT"
	ComplianceStatusUnderReview  ComplianceStatus = "UNDER_REVIEW"
	ComplianceStatusNotAssessed  ComplianceStatus = "NOT_ASSESSED"
)

// IsValid checks if the compliance status is valid
func (s ComplianceStatus) IsValid() bool {
	switch s {
	case ComplianceStatusCompliant, ComplianceStatusNonCompliant, ComplianceStatusUnderReview, ComplianceStatusNotAssessed:
		return true
	}
	return false
}

// ComplianceRecord represents a compliance assessment record for a project
type ComplianceRecord struct {
	ID             uuid.UUID           `json:"id"`
	ProjectID      uuid.UUID           `json:"projectId"`
	RiskLevel      ComplianceRiskLevel `json:"riskLevel"`
	Status         ComplianceStatus    `json:"status"`
	AssessmentDate time.Time           `json:"assessmentDate"`
	NextReviewDate time.Time           `json:"nextReviewDate"`
	Findings       []ComplianceFinding `json:"findings"`
	AuditorNotes   string              `json:"auditorNotes"`
	CreatedAt      time.Time           `json:"createdAt"`
	UpdatedAt      time.Time           `json:"updatedAt"`
}

// ComplianceFinding represents an individual finding within a compliance assessment
type ComplianceFinding struct {
	ID          string `json:"id"`
	Category    string `json:"category"`    // transparency, fairness, robustness, accountability
	Severity    string `json:"severity"`    // critical, major, minor
	Description string `json:"description"`
	Remediation string `json:"remediation"`
	Status      string `json:"status"` // open, resolved
}

// ImmutableAuditEntry represents a tamper-evident audit log entry with hash chaining
type ImmutableAuditEntry struct {
	ID           uuid.UUID `json:"id"`
	ProjectID    uuid.UUID `json:"projectId"`
	TraceID      string    `json:"traceId"`
	EntryType    string    `json:"entryType"`
	Actor        string    `json:"actor"`
	Action       string    `json:"action"`
	Details      string    `json:"details"` // JSON payload
	PreviousHash string    `json:"previousHash"`
	Hash         string    `json:"hash"`
	Timestamp    time.Time `json:"timestamp"`
}

// ComputeHash computes a SHA-256 hash of the entry fields for tamper detection
func (e *ImmutableAuditEntry) ComputeHash() string {
	data := e.EntryType + e.Actor + e.Action + e.Details + e.PreviousHash + e.Timestamp.UTC().Format(time.RFC3339Nano)
	h := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", h)
}

// ConformityAssessment represents an EU AI Act conformity assessment for an AI system
type ConformityAssessment struct {
	ID                  uuid.UUID           `json:"id"`
	ProjectID           uuid.UUID           `json:"projectId"`
	SystemName          string              `json:"systemName"`
	SystemDescription   string              `json:"systemDescription"`
	RiskLevel           ComplianceRiskLevel `json:"riskLevel"`
	Provider            string              `json:"provider"`
	DeploymentDate      *time.Time          `json:"deploymentDate,omitempty"`
	TransparencyScore   float64             `json:"transparencyScore"`
	FairnessScore       float64             `json:"fairnessScore"`
	RobustnessScore     float64             `json:"robustnessScore"`
	HumanOversightLevel string              `json:"humanOversightLevel"`
	DataGovernanceNotes string              `json:"dataGovernanceNotes"`
	Status              ComplianceStatus    `json:"status"`
	CreatedAt           time.Time           `json:"createdAt"`
	UpdatedAt           time.Time           `json:"updatedAt"`
}

// ComplianceReportInput represents input parameters for generating a compliance report
type ComplianceReportInput struct {
	ProjectID  uuid.UUID `json:"projectId"`
	ReportType string    `json:"reportType"` // conformity_assessment, audit_summary, risk_report
	DateRange  DateRange `json:"dateRange"`
}
