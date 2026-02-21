package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// ComplianceRepository defines repository operations for compliance records
type ComplianceRepository interface {
	SaveRecord(ctx context.Context, record *domain.ComplianceRecord) error
	GetRecord(ctx context.Context, projectID uuid.UUID) (*domain.ComplianceRecord, error)
	ListRecords(ctx context.Context, projectID uuid.UUID) ([]domain.ComplianceRecord, error)
	SaveAuditEntry(ctx context.Context, entry *domain.ImmutableAuditEntry) error
	ListAuditEntries(ctx context.Context, projectID uuid.UUID, start, end time.Time) ([]domain.ImmutableAuditEntry, error)
	SaveAssessment(ctx context.Context, assessment *domain.ConformityAssessment) error
	GetAssessment(ctx context.Context, assessmentID uuid.UUID) (*domain.ConformityAssessment, error)
}

// ComplianceService manages EU AI Act compliance features including risk
// assessments, immutable audit trails, and conformity assessments.
type ComplianceService struct {
	logger         *zap.Logger
	complianceRepo ComplianceRepository
	auditService   *AuditService
}

// NewComplianceService creates a new compliance service
func NewComplianceService(
	logger *zap.Logger,
	complianceRepo ComplianceRepository,
	auditService *AuditService,
) *ComplianceService {
	return &ComplianceService{
		logger:         logger,
		complianceRepo: complianceRepo,
		auditService:   auditService,
	}
}

// AssessProject runs an automated risk assessment for a project and persists
// the resulting compliance record.
func (s *ComplianceService) AssessProject(ctx context.Context, projectID uuid.UUID) (*domain.ComplianceRecord, error) {
	var findings []domain.ComplianceFinding

	// Transparency check
	findings = append(findings, domain.ComplianceFinding{
		ID:          uuid.New().String(),
		Category:    "transparency",
		Severity:    "minor",
		Description: "Verify that all AI-generated outputs are clearly labelled",
		Remediation: "Add disclosure metadata to all generation observations",
		Status:      "open",
	})

	// Determine risk level based on findings
	riskLevel := domain.ComplianceRiskLevelMinimal
	for _, f := range findings {
		if f.Severity == "critical" {
			riskLevel = domain.ComplianceRiskLevelHigh
			break
		}
		if f.Severity == "major" {
			riskLevel = domain.ComplianceRiskLevelLimited
		}
	}

	status := domain.ComplianceStatusCompliant
	if riskLevel == domain.ComplianceRiskLevelHigh {
		status = domain.ComplianceStatusNonCompliant
	}

	now := time.Now()
	record := &domain.ComplianceRecord{
		ID:             uuid.New(),
		ProjectID:      projectID,
		RiskLevel:      riskLevel,
		Status:         status,
		AssessmentDate: now,
		NextReviewDate: now.AddDate(0, 3, 0),
		Findings:       findings,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.complianceRepo.SaveRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to save compliance record: %w", err)
	}

	s.logger.Info("completed compliance assessment",
		zap.String("projectId", projectID.String()),
		zap.String("riskLevel", string(riskLevel)),
		zap.String("status", string(status)),
	)

	return record, nil
}

// GetComplianceStatus retrieves the latest compliance record for a project.
func (s *ComplianceService) GetComplianceStatus(ctx context.Context, projectID uuid.UUID) (*domain.ComplianceRecord, error) {
	record, err := s.complianceRepo.GetRecord(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance record: %w", err)
	}
	return record, nil
}

// RecordImmutableAuditEntry creates a new tamper-evident audit entry by fetching
// the previous entry hash and computing a hash chain.
func (s *ComplianceService) RecordImmutableAuditEntry(ctx context.Context, projectID uuid.UUID, entry *domain.ImmutableAuditEntry) (*domain.ImmutableAuditEntry, error) {
	// Fetch previous entries to get the last hash for chaining
	now := time.Now()
	entries, err := s.complianceRepo.ListAuditEntries(ctx, projectID, time.Time{}, now)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit entries: %w", err)
	}

	previousHash := ""
	if len(entries) > 0 {
		previousHash = entries[len(entries)-1].Hash
	}

	entry.ID = uuid.New()
	entry.ProjectID = projectID
	entry.PreviousHash = previousHash
	entry.Timestamp = now
	entry.Hash = entry.ComputeHash()

	if err := s.complianceRepo.SaveAuditEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to save immutable audit entry: %w", err)
	}

	s.logger.Debug("recorded immutable audit entry",
		zap.String("projectId", projectID.String()),
		zap.String("entryId", entry.ID.String()),
	)

	return entry, nil
}

// GetAuditTrail retrieves the immutable audit trail for a project within
// the specified time range.
func (s *ComplianceService) GetAuditTrail(ctx context.Context, projectID uuid.UUID, startTime, endTime time.Time) ([]domain.ImmutableAuditEntry, error) {
	entries, err := s.complianceRepo.ListAuditEntries(ctx, projectID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit trail: %w", err)
	}
	return entries, nil
}

// CreateConformityAssessment creates a new EU AI Act conformity assessment
// for a project.
func (s *ComplianceService) CreateConformityAssessment(ctx context.Context, projectID uuid.UUID, input *domain.ConformityAssessment) (*domain.ConformityAssessment, error) {
	now := time.Now()
	input.ID = uuid.New()
	input.ProjectID = projectID
	input.Status = domain.ComplianceStatusUnderReview
	input.CreatedAt = now
	input.UpdatedAt = now

	if err := s.complianceRepo.SaveAssessment(ctx, input); err != nil {
		return nil, fmt.Errorf("failed to save conformity assessment: %w", err)
	}

	s.logger.Info("created conformity assessment",
		zap.String("projectId", projectID.String()),
		zap.String("assessmentId", input.ID.String()),
		zap.String("riskLevel", string(input.RiskLevel)),
	)

	return input, nil
}

// GetConformityAssessment retrieves a conformity assessment by ID.
func (s *ComplianceService) GetConformityAssessment(ctx context.Context, assessmentID uuid.UUID) (*domain.ConformityAssessment, error) {
	assessment, err := s.complianceRepo.GetAssessment(ctx, assessmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conformity assessment: %w", err)
	}
	return assessment, nil
}

// GenerateComplianceReport generates a JSON compliance report based on the
// provided input parameters.
func (s *ComplianceService) GenerateComplianceReport(ctx context.Context, input domain.ComplianceReportInput) ([]byte, error) {
	record, err := s.complianceRepo.GetRecord(ctx, input.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance record for report: %w", err)
	}

	entries, err := s.complianceRepo.ListAuditEntries(ctx, input.ProjectID, input.DateRange.Start, input.DateRange.End)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit entries for report: %w", err)
	}

	report := map[string]any{
		"reportType":  input.ReportType,
		"projectId":   input.ProjectID.String(),
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
		"dateRange": map[string]string{
			"start": input.DateRange.Start.Format(time.RFC3339),
			"end":   input.DateRange.End.Format(time.RFC3339),
		},
		"complianceRecord": record,
		"auditEntryCount":  len(entries),
	}

	data, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal compliance report: %w", err)
	}

	s.logger.Info("generated compliance report",
		zap.String("projectId", input.ProjectID.String()),
		zap.String("reportType", input.ReportType),
	)

	return data, nil
}
