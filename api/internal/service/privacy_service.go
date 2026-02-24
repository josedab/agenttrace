package service

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// PrivacyService manages privacy-preserving analytics
type PrivacyService struct {
	logger           *zap.Logger
	mu               sync.RWMutex
	configs          map[uuid.UUID]*domain.PIIConfig
	deletionRequests map[uuid.UUID]*domain.DataDeletionRequest
	patterns         map[domain.PIIType]*regexp.Regexp
}

// NewPrivacyService creates a new privacy service
func NewPrivacyService(logger *zap.Logger) *PrivacyService {
	return &PrivacyService{
		logger:           logger,
		configs:          make(map[uuid.UUID]*domain.PIIConfig),
		deletionRequests: make(map[uuid.UUID]*domain.DataDeletionRequest),
		patterns: map[domain.PIIType]*regexp.Regexp{
			domain.PIISSN:        regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			domain.PIICreditCard: regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`),
			domain.PIIEmail:      regexp.MustCompile(`\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`),
			domain.PIIPhone:      regexp.MustCompile(`\b(\+?1?[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}\b`),
		},
	}
}

// ScanForPII scans content for personally identifiable information
func (s *PrivacyService) ScanForPII(ctx context.Context, projectID uuid.UUID, content string) (*domain.PIIDetectionResult, error) {
	var findings []domain.PIIFinding

	for piiType, pattern := range s.patterns {
		matches := pattern.FindAllStringIndex(content, -1)
		for _, match := range matches {
			value := content[match[0]:match[1]]
			findings = append(findings, domain.PIIFinding{
				Type:       piiType,
				Value:      s.maskValue(value),
				Location:   fmt.Sprintf("offset:%d-%d", match[0], match[1]),
				Confidence: 0.95,
				Redacted:   false,
			})
		}
	}

	cfg := s.getConfigInternal(projectID)
	redacted := cfg != nil && cfg.AutoRedact

	result := &domain.PIIDetectionResult{
		TotalScanned:    len(content),
		PIIFound:        len(findings),
		Findings:        findings,
		RedactionApplied: redacted,
	}

	s.logger.Info("PII scan completed",
		zap.String("projectId", projectID.String()),
		zap.Int("findings", len(findings)),
	)
	return result, nil
}

// GetConfig returns the PII configuration for a project
func (s *PrivacyService) GetConfig(ctx context.Context, projectID uuid.UUID) (*domain.PIIConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, ok := s.configs[projectID]
	if !ok {
		return &domain.PIIConfig{
			ID:               uuid.New(),
			ProjectID:        projectID,
			Enabled:          true,
			SensitivityLevel: domain.SensitivityMedium,
			AutoRedact:       false,
			DataResidency:    domain.ResidencyGlobal,
			RetentionDays:    90,
			CreatedAt:        time.Now(),
		}, nil
	}
	return cfg, nil
}

// UpdateConfig updates the PII configuration for a project
func (s *PrivacyService) UpdateConfig(ctx context.Context, projectID uuid.UUID, input *domain.PIIConfigInput) (*domain.PIIConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, ok := s.configs[projectID]
	if !ok {
		cfg = &domain.PIIConfig{
			ID:               uuid.New(),
			ProjectID:        projectID,
			Enabled:          true,
			SensitivityLevel: domain.SensitivityMedium,
			AutoRedact:       false,
			DataResidency:    domain.ResidencyGlobal,
			RetentionDays:    90,
			CreatedAt:        time.Now(),
		}
	}

	if input.Enabled != nil {
		cfg.Enabled = *input.Enabled
	}
	if input.SensitivityLevel != nil {
		cfg.SensitivityLevel = *input.SensitivityLevel
	}
	if input.AutoRedact != nil {
		cfg.AutoRedact = *input.AutoRedact
	}
	if input.DataResidency != nil {
		cfg.DataResidency = *input.DataResidency
	}
	if input.RetentionDays != nil {
		cfg.RetentionDays = *input.RetentionDays
	}

	s.configs[projectID] = cfg
	s.logger.Info("updated PII config", zap.String("projectId", projectID.String()))
	return cfg, nil
}

// RequestDeletion creates a data deletion request
func (s *PrivacyService) RequestDeletion(ctx context.Context, projectID uuid.UUID, input *domain.DeletionRequestInput) (*domain.DataDeletionRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	req := &domain.DataDeletionRequest{
		ID:          uuid.New(),
		ProjectID:   projectID,
		RequestType: input.RequestType,
		TargetID:    input.TargetID,
		Status:      domain.DeletionPending,
		RequestedAt: time.Now(),
	}

	s.deletionRequests[req.ID] = req
	s.logger.Info("created deletion request",
		zap.String("id", req.ID.String()),
		zap.String("type", string(input.RequestType)),
	)
	return req, nil
}

// ListDeletionRequests lists all deletion requests for a project
func (s *PrivacyService) ListDeletionRequests(ctx context.Context, projectID uuid.UUID) ([]domain.DataDeletionRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.DataDeletionRequest
	for _, req := range s.deletionRequests {
		if req.ProjectID == projectID {
			result = append(result, *req)
		}
	}
	return result, nil
}

func (s *PrivacyService) getConfigInternal(projectID uuid.UUID) *domain.PIIConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.configs[projectID]
}

func (s *PrivacyService) maskValue(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + "***" + value[len(value)-2:]
}
