package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReportTemplate represents the type of compliance report template
type ReportTemplate string

const (
	ReportEUAIAct ReportTemplate = "eu_ai_act"
	ReportSOC2    ReportTemplate = "soc2"
	ReportCustom  ReportTemplate = "custom"
)

// ComplianceReport represents a generated compliance report
type ComplianceReport struct {
	ID          uuid.UUID       `json:"id"`
	ProjectID   uuid.UUID       `json:"projectId"`
	Template    ReportTemplate  `json:"template"`
	Title       string          `json:"title"`
	Status      string          `json:"status"` // generating, ready, error
	Sections    []ReportSection `json:"sections"`
	Summary     string          `json:"summary"`
	Score       float64         `json:"score"` // 0-100 compliance score
	GeneratedAt time.Time       `json:"generatedAt"`
	Period      DateRange       `json:"period"`
}

// ReportSection represents a section within a compliance report
type ReportSection struct {
	Title           string           `json:"title"`
	Article         string           `json:"article,omitempty"` // EU AI Act article reference
	Status          string           `json:"status"`            // compliant, partial, non_compliant
	Description     string           `json:"description"`
	Evidence        []ReportEvidence `json:"evidence"`
	Recommendations []string         `json:"recommendations,omitempty"`
}

// ReportEvidence represents evidence within a report section
type ReportEvidence struct {
	Type        string `json:"type"` // metric, config, log, policy
	Description string `json:"description"`
	Value       string `json:"value"`
	Source      string `json:"source"`
}

// GenerateReportInput represents input for generating a compliance report
type GenerateReportInput struct {
	Template ReportTemplate `json:"template" validate:"required"`
	Title    string         `json:"title,omitempty"`
	Period   *DateRange     `json:"period,omitempty"`
}
