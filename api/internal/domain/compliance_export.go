package domain

import (
	"time"

	"github.com/google/uuid"
)

// ComplianceExportFormat represents a supported compliance export format
type ComplianceExportFormat string

const (
	ComplianceExportFormatSOC2      ComplianceExportFormat = "SOC2"
	ComplianceExportFormatISO42001  ComplianceExportFormat = "ISO_42001"
	ComplianceExportFormatEUAIAct   ComplianceExportFormat = "EU_AI_ACT"
	ComplianceExportFormatPDFSummary ComplianceExportFormat = "PDF_SUMMARY"
	ComplianceExportFormatJSONFull  ComplianceExportFormat = "JSON_FULL"
)

// IsValid checks if the compliance export format is valid
func (f ComplianceExportFormat) IsValid() bool {
	switch f {
	case ComplianceExportFormatSOC2,
		ComplianceExportFormatISO42001,
		ComplianceExportFormatEUAIAct,
		ComplianceExportFormatPDFSummary,
		ComplianceExportFormatJSONFull:
		return true
	}
	return false
}

// ComplianceExportJob represents an asynchronous compliance export job
type ComplianceExportJob struct {
	ID            uuid.UUID              `json:"id"`
	ProjectID     uuid.UUID              `json:"projectId"`
	Format        ComplianceExportFormat `json:"format"`
	Status        string                 `json:"status"` // pending, generating, completed, failed
	DateRange     DateRange              `json:"dateRange"`
	FileURL       string                 `json:"fileUrl,omitempty"`
	FileSizeBytes int64                  `json:"fileSizeBytes,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
	CompletedAt   *time.Time             `json:"completedAt,omitempty"`
}

// ComplianceExportInput represents input for starting a compliance export
type ComplianceExportInput struct {
	Format          ComplianceExportFormat `json:"format"`
	DateRange       DateRange              `json:"dateRange"`
	IncludeTraces   bool                   `json:"includeTraces"`
	IncludeScores   bool                   `json:"includeScores"`
	IncludeAuditLog bool                   `json:"includeAuditLog"`
}

// ComplianceTemplate defines a template for a compliance export format
type ComplianceTemplate struct {
	Name     string                 `json:"name"`
	Format   ComplianceExportFormat `json:"format"`
	Sections []TemplateSection      `json:"sections"`
}

// TemplateSection describes a section within a compliance template
type TemplateSection struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DataQuery   string `json:"dataQuery"`
	Required    bool   `json:"required"`
}
