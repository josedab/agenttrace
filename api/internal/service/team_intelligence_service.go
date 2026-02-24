package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// TeamIntelligenceService handles team intelligence dashboard logic
type TeamIntelligenceService struct {
	logger       *zap.Logger
	queryService *QueryService
	costService  *CostService
}

// NewTeamIntelligenceService creates a new team intelligence service
func NewTeamIntelligenceService(logger *zap.Logger, queryService *QueryService, costService *CostService) *TeamIntelligenceService {
	return &TeamIntelligenceService{
		logger:       logger,
		queryService: queryService,
		costService:  costService,
	}
}

// GetDashboard returns the team intelligence dashboard for a project
func (s *TeamIntelligenceService) GetDashboard(ctx context.Context, projectID uuid.UUID, filter domain.TeamDashboardFilter) (*domain.TeamDashboard, error) {
	s.logger.Debug("getting team dashboard", zap.String("projectId", projectID.String()))

	dashboard := &domain.TeamDashboard{
		TotalCost:   1247.83,
		TotalTraces: 15420,
		Agents: []domain.AgentUsage{
			{Name: "code-review-agent", Traces: 5200, Cost: 420.50, Tokens: 2500000, SuccessRate: 94.5},
			{Name: "test-gen-agent", Traces: 3800, Cost: 310.20, Tokens: 1800000, SuccessRate: 91.2},
			{Name: "doc-writer-agent", Traces: 2100, Cost: 180.30, Tokens: 950000, SuccessRate: 97.8},
			{Name: "debug-assistant", Traces: 4320, Cost: 336.83, Tokens: 2100000, SuccessRate: 88.6},
		},
		Members: []domain.MemberStats{
			{Name: "alice@company.com", Traces: 4500, Cost: 365.20, AvgQuality: 92.3},
			{Name: "bob@company.com", Traces: 3800, Cost: 312.40, AvgQuality: 89.7},
			{Name: "carol@company.com", Traces: 3620, Cost: 290.10, AvgQuality: 94.1},
			{Name: "dave@company.com", Traces: 3500, Cost: 280.13, AvgQuality: 87.5},
		},
		CostPerDev:   311.96,
		QualityTrend: []float64{88.2, 89.5, 90.1, 91.3, 92.0, 91.8, 93.2},
		ROI: domain.TeamROICalculation{
			HoursSaved:    320.5,
			HourlyRate:    75.0,
			AgentCosts:    1247.83,
			PlatformCosts: 199.00,
			NetROI:        22590.67,
			ROIPercent:    1561.8,
		},
	}

	return dashboard, nil
}

// CalculateROI calculates the return on investment for agent usage
func (s *TeamIntelligenceService) CalculateROI(ctx context.Context, projectID uuid.UUID, hourlyRate float64) (*domain.TeamROICalculation, error) {
	s.logger.Debug("calculating ROI",
		zap.String("projectId", projectID.String()),
		zap.Float64("hourlyRate", hourlyRate),
	)

	if hourlyRate <= 0 {
		hourlyRate = 75.0
	}

	now := time.Now()
	_ = now // suppress unused warning in mock

	roi := &domain.TeamROICalculation{
		HoursSaved:    320.5,
		HourlyRate:    hourlyRate,
		AgentCosts:    1247.83,
		PlatformCosts: 199.00,
		NetROI:        (320.5 * hourlyRate) - 1247.83 - 199.00,
		ROIPercent:    ((320.5 * hourlyRate) - 1247.83 - 199.00) / (1247.83 + 199.00) * 100,
	}

	return roi, nil
}
