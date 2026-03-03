package domain

import (
	"time"

	"github.com/google/uuid"
)

// CertificationFramework represents a compliance certification framework
type CertificationFramework string

const (
	CertFrameworkSOC2   CertificationFramework = "soc2"
	CertFrameworkHIPAA  CertificationFramework = "hipaa"
	CertFrameworkEUAI   CertificationFramework = "eu_ai_act"
	CertFrameworkISO27001 CertificationFramework = "iso_27001"
	CertFrameworkGDPR   CertificationFramework = "gdpr"
)

// CertificationOutputFormat represents the output format
type CertificationOutputFormat string

const (
	CertOutputPDF  CertificationOutputFormat = "pdf"
	CertOutputJSON CertificationOutputFormat = "json"
	CertOutputCSV  CertificationOutputFormat = "csv"
)

// CertificationStatus represents the status of a certification export
type CertificationStatus string

const (
	CertStatusPending    CertificationStatus = "pending"
	CertStatusGenerating CertificationStatus = "generating"
	CertStatusCompleted  CertificationStatus = "completed"
	CertStatusFailed     CertificationStatus = "failed"
)

// Certification represents a compliance certification export
type Certification struct {
	ID           uuid.UUID                 `json:"id"`
	ProjectID    uuid.UUID                 `json:"projectId"`
	Framework    CertificationFramework    `json:"framework"`
	Format       CertificationOutputFormat `json:"format"`
	Status       CertificationStatus       `json:"status"`
	DateFrom     time.Time                 `json:"dateFrom"`
	DateTo       time.Time                 `json:"dateTo"`
	Sections     []CertSection             `json:"sections"`
	Summary      CertSummary               `json:"summary"`
	Attestation  *CertAttestation          `json:"attestation,omitempty"`
	DownloadURL  string                    `json:"downloadUrl,omitempty"`
	CreatedBy    uuid.UUID                 `json:"createdBy"`
	CreatedAt    time.Time                 `json:"createdAt"`
	CompletedAt  *time.Time                `json:"completedAt,omitempty"`
}

// CertSection represents a section of the certification report
type CertSection struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Status      string          `json:"status"` // "pass", "fail", "warning", "not_applicable"
	Evidence    []CertEvidence  `json:"evidence"`
	Score       float64         `json:"score"`
}

// CertEvidence represents a piece of compliance evidence
type CertEvidence struct {
	Type        string `json:"type"` // "trace", "pii_scan", "model_risk", "fairness", "data_lineage"
	Description string `json:"description"`
	Reference   string `json:"reference"`
	Timestamp   time.Time `json:"timestamp"`
}

// CertSummary represents the overall certification summary
type CertSummary struct {
	OverallScore    float64 `json:"overallScore"`
	TotalSections   int     `json:"totalSections"`
	PassedSections  int     `json:"passedSections"`
	FailedSections  int     `json:"failedSections"`
	WarningSections int     `json:"warningSections"`
	TracesAnalyzed  int     `json:"tracesAnalyzed"`
	PIIFindings     int     `json:"piiFindings"`
	RiskLevel       string  `json:"riskLevel"` // "low", "medium", "high"
}

// CertAttestation represents a digital attestation signature
type CertAttestation struct {
	SignedBy    string    `json:"signedBy"`
	SignedAt    time.Time `json:"signedAt"`
	Fingerprint string   `json:"fingerprint"`
	Algorithm   string   `json:"algorithm"`
}

// CertificationExportInput represents input for creating a certification
type CertificationExportInput struct {
	Framework    CertificationFramework    `json:"framework" validate:"required"`
	Format       CertificationOutputFormat `json:"format,omitempty"`
	DateFrom     string                    `json:"dateFrom" validate:"required"`
	DateTo       string                    `json:"dateTo" validate:"required"`
	IncludeTraces bool                     `json:"includeTraces"`
	IncludePII    bool                     `json:"includePii"`
	IncludeRisk   bool                     `json:"includeRisk"`
	IncludeFairness bool                   `json:"includeFairness"`
	SignedBy      string                   `json:"signedBy,omitempty"`
}

// CertFrameworkInfo provides information about a compliance framework
type CertFrameworkInfo struct {
	Framework   CertificationFramework `json:"framework"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Sections    []string               `json:"sections"`
}
