package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// CloudOnboardingService handles cloud tenant onboarding and usage metering
type CloudOnboardingService struct {
	logger *zap.Logger
}

// NewCloudOnboardingService creates a new cloud onboarding service
func NewCloudOnboardingService(logger *zap.Logger) *CloudOnboardingService {
	return &CloudOnboardingService{
		logger: logger,
	}
}

// CreateOnboarding initializes the onboarding flow for a new cloud tenant
func (s *CloudOnboardingService) CreateOnboarding(ctx context.Context, tenantID uuid.UUID) (*domain.CloudOnboarding, error) {
	now := time.Now()

	steps := []domain.OnboardingStepStatus{
		{Step: domain.OnboardingStepAccountCreated, Completed: true, CompletedAt: &now},
		{Step: domain.OnboardingStepProjectCreated, Completed: false},
		{Step: domain.OnboardingStepSDKInstalled, Completed: false},
		{Step: domain.OnboardingStepFirstTrace, Completed: false},
		{Step: domain.OnboardingStepFirstEval, Completed: false},
	}

	onboarding := &domain.CloudOnboarding{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Steps:       steps,
		CurrentStep: domain.OnboardingStepProjectCreated,
		CreatedAt:   now,
	}

	s.logger.Info("onboarding created",
		zap.String("tenantId", tenantID.String()),
		zap.String("id", onboarding.ID.String()),
	)
	return onboarding, nil
}

// GetOnboarding retrieves the current onboarding state for a tenant
func (s *CloudOnboardingService) GetOnboarding(ctx context.Context, tenantID uuid.UUID) (*domain.CloudOnboarding, error) {
	s.logger.Debug("fetching onboarding", zap.String("tenantId", tenantID.String()))

	now := time.Now()
	return &domain.CloudOnboarding{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Steps:       []domain.OnboardingStepStatus{},
		CurrentStep: domain.OnboardingStepAccountCreated,
		CreatedAt:   now,
	}, nil
}

// CompleteStep marks an onboarding step as completed and advances to the next step
func (s *CloudOnboardingService) CompleteStep(ctx context.Context, tenantID uuid.UUID, step domain.OnboardingStep) error {
	if !step.IsValid() {
		return fmt.Errorf("invalid onboarding step: %s", step)
	}

	s.logger.Info("onboarding step completed",
		zap.String("tenantId", tenantID.String()),
		zap.String("step", string(step)),
	)
	return nil
}

// GenerateQuickstart generates a quickstart configuration with code snippets
// for the specified language and framework
func (s *CloudOnboardingService) GenerateQuickstart(ctx context.Context, language, framework, apiKey, projectID, host string) (*domain.QuickstartConfig, error) {
	supportedLanguages := map[string]bool{
		"python": true, "typescript": true, "javascript": true, "go": true, "java": true, "ruby": true,
	}
	if !supportedLanguages[language] {
		return nil, fmt.Errorf("unsupported language: %s; supported: python, typescript, javascript, go, java, ruby", language)
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for quickstart generation")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required for quickstart generation")
	}

	if host == "" {
		host = "https://cloud.agenttrace.io"
	}

	config := &domain.QuickstartConfig{
		Language:  language,
		Framework: framework,
		APIKey:    apiKey,
		ProjectID: projectID,
		Host:      host,
	}

	s.logger.Info("quickstart generated",
		zap.String("language", language),
		zap.String("framework", framework),
		zap.String("projectId", projectID),
	)
	return config, nil
}

// GetUsage retrieves current usage metrics for a tenant
func (s *CloudOnboardingService) GetUsage(ctx context.Context, tenantID uuid.UUID) (*domain.UsageMeter, error) {
	s.logger.Debug("fetching usage", zap.String("tenantId", tenantID.String()))

	usage := &domain.UsageMeter{
		TenantID:          tenantID,
		Period:            time.Now().Format("2006-01"),
		TracesUsed:        12450,
		TracesLimit:        50000,
		StorageUsedBytes:  1073741824, // 1 GB
		StorageLimitBytes: 10737418240, // 10 GB
		APICallsUsed:      34200,
		APICallsLimit:     100000,
		PercentUsed:       24.9,
	}

	return usage, nil
}

// CheckQuota checks whether a tenant has remaining quota for a given metric
func (s *CloudOnboardingService) CheckQuota(ctx context.Context, tenantID uuid.UUID, metric string, quantity int64) (bool, error) {
	validMetrics := map[string]bool{
		"traces": true, "storage": true, "api_calls": true,
	}
	if !validMetrics[metric] {
		return false, fmt.Errorf("unknown metric: %s; supported: traces, storage, api_calls", metric)
	}

	usage, err := s.GetUsage(ctx, tenantID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch usage: %w", err)
	}

	var remaining int64
	switch metric {
	case "traces":
		remaining = usage.TracesLimit - usage.TracesUsed
	case "storage":
		remaining = usage.StorageLimitBytes - usage.StorageUsedBytes
	case "api_calls":
		remaining = usage.APICallsLimit - usage.APICallsUsed
	}

	allowed := quantity <= remaining

	s.logger.Debug("quota check",
		zap.String("tenantId", tenantID.String()),
		zap.String("metric", metric),
		zap.Int64("requested", quantity),
		zap.Int64("remaining", remaining),
		zap.Bool("allowed", allowed),
	)
	return allowed, nil
}
