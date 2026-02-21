package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// ComplianceExportRepository defines repository operations for compliance exports
type ComplianceExportRepository interface {
	SaveJob(ctx context.Context, job *domain.ComplianceExportJob) error
	GetJob(ctx context.Context, jobID uuid.UUID) (*domain.ComplianceExportJob, error)
	ListJobs(ctx context.Context, projectID uuid.UUID) ([]domain.ComplianceExportJob, error)
}

// ComplianceExportService generates compliance-ready export documents in
// various regulatory formats such as SOC2, ISO 42001, and EU AI Act.
type ComplianceExportService struct {
	logger            *zap.Logger
	exportRepo        ComplianceExportRepository
	complianceService *ComplianceService
	auditService      *AuditService
}

// NewComplianceExportService creates a new compliance export service
func NewComplianceExportService(
	logger *zap.Logger,
	exportRepo ComplianceExportRepository,
	complianceService *ComplianceService,
	auditService *AuditService,
) *ComplianceExportService {
	return &ComplianceExportService{
		logger:            logger,
		exportRepo:        exportRepo,
		complianceService: complianceService,
		auditService:      auditService,
	}
}

// StartExport creates a new compliance export job in pending status.
func (s *ComplianceExportService) StartExport(ctx context.Context, projectID uuid.UUID, input domain.ComplianceExportInput) (*domain.ComplianceExportJob, error) {
	if !input.Format.IsValid() {
		return nil, fmt.Errorf("invalid compliance export format: %s", input.Format)
	}

	now := time.Now()
	job := &domain.ComplianceExportJob{
		ID:        uuid.New(),
		ProjectID: projectID,
		Format:    input.Format,
		Status:    "pending",
		DateRange: input.DateRange,
		CreatedAt: now,
	}

	if err := s.exportRepo.SaveJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to save compliance export job: %w", err)
	}

	s.logger.Info("started compliance export",
		zap.String("projectId", projectID.String()),
		zap.String("jobId", job.ID.String()),
		zap.String("format", string(input.Format)),
	)

	return job, nil
}

// GetExportJob retrieves a compliance export job by ID.
func (s *ComplianceExportService) GetExportJob(ctx context.Context, jobID uuid.UUID) (*domain.ComplianceExportJob, error) {
	job, err := s.exportRepo.GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance export job: %w", err)
	}
	return job, nil
}

// ListExports retrieves all compliance export jobs for a project.
func (s *ComplianceExportService) ListExports(ctx context.Context, projectID uuid.UUID) ([]domain.ComplianceExportJob, error) {
	jobs, err := s.exportRepo.ListJobs(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list compliance export jobs: %w", err)
	}
	return jobs, nil
}

// GetTemplates returns the available compliance export templates.
func (s *ComplianceExportService) GetTemplates() []domain.ComplianceTemplate {
	return []domain.ComplianceTemplate{
		{
			Name:   "SOC 2 Type II Report",
			Format: domain.ComplianceExportFormatSOC2,
			Sections: []domain.TemplateSection{
				{Title: "System Description", Description: "Overview of the AI system and its boundaries", DataQuery: "project_metadata", Required: true},
				{Title: "Control Activities", Description: "Security, availability, and processing integrity controls", DataQuery: "audit_entries", Required: true},
				{Title: "Monitoring", Description: "Continuous monitoring and anomaly detection summary", DataQuery: "anomaly_summary", Required: true},
			},
		},
		{
			Name:   "ISO 42001 AI Management System",
			Format: domain.ComplianceExportFormatISO42001,
			Sections: []domain.TemplateSection{
				{Title: "AI Policy", Description: "Organizational AI governance policies", DataQuery: "compliance_records", Required: true},
				{Title: "Risk Assessment", Description: "AI risk identification and treatment", DataQuery: "risk_assessments", Required: true},
				{Title: "Performance Evaluation", Description: "AI system performance metrics and evaluations", DataQuery: "scorecard_data", Required: false},
			},
		},
		{
			Name:   "EU AI Act Conformity Assessment",
			Format: domain.ComplianceExportFormatEUAIAct,
			Sections: []domain.TemplateSection{
				{Title: "Risk Classification", Description: "Risk level determination and justification", DataQuery: "conformity_assessment", Required: true},
				{Title: "Transparency Requirements", Description: "Disclosure and transparency measures", DataQuery: "transparency_checks", Required: true},
				{Title: "Human Oversight", Description: "Human-in-the-loop mechanisms", DataQuery: "oversight_records", Required: true},
				{Title: "Audit Trail", Description: "Immutable audit log for the reporting period", DataQuery: "audit_trail", Required: true},
			},
		},
	}
}

// GenerateReport performs the actual report generation for a pending job,
// updating its status to completed or failed.
func (s *ComplianceExportService) GenerateReport(ctx context.Context, jobID uuid.UUID) error {
	job, err := s.exportRepo.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job for report generation: %w", err)
	}

	if job.Status != "pending" {
		return fmt.Errorf("job %s is not in pending status (current: %s)", jobID, job.Status)
	}

	// Mark as generating
	job.Status = "generating"
	if err := s.exportRepo.SaveJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update job status to generating: %w", err)
	}

	// Gather compliance data
	_, compErr := s.complianceService.GetComplianceStatus(ctx, job.ProjectID)
	if compErr != nil {
		job.Status = "failed"
		_ = s.exportRepo.SaveJob(ctx, job)
		return fmt.Errorf("failed to gather compliance data: %w", compErr)
	}

	// Mark as completed
	now := time.Now()
	job.Status = "completed"
	job.CompletedAt = &now
	job.FileURL = fmt.Sprintf("/exports/%s/%s.json", job.ProjectID, job.ID)

	if err := s.exportRepo.SaveJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update job status to completed: %w", err)
	}

	s.logger.Info("generated compliance report",
		zap.String("jobId", jobID.String()),
		zap.String("format", string(job.Format)),
	)

	return nil
}
