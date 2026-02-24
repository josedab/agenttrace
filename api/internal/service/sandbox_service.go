package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

type SandboxService struct {
	logger   *zap.Logger
	reviews  map[uuid.UUID]*domain.SandboxReview
	policies map[uuid.UUID]*domain.SandboxPolicy
}

func NewSandboxService(logger *zap.Logger) *SandboxService {
	return &SandboxService{
		logger:   logger,
		reviews:  make(map[uuid.UUID]*domain.SandboxReview),
		policies: make(map[uuid.UUID]*domain.SandboxPolicy),
	}
}

func (s *SandboxService) SubmitForReview(ctx context.Context, projectID uuid.UUID, input *domain.SandboxReviewInput) (*domain.SandboxReview, error) {
	review := &domain.SandboxReview{
		ID:              uuid.New(),
		ProjectID:       projectID,
		TraceID:         input.TraceID,
		Status:          domain.SandboxStatusPending,
		ProposedActions: input.ProposedActions,
		CreatedAt:       time.Now(),
		ExpiresAt:       time.Now().Add(24 * time.Hour),
	}

	// Assess risk
	review.RiskScore, review.RiskLevel = s.assessRisk(input.ProposedActions)

	s.reviews[review.ID] = review
	s.logger.Info("sandbox review submitted",
		zap.String("reviewId", review.ID.String()),
		zap.String("riskLevel", review.RiskLevel),
		zap.Float64("riskScore", review.RiskScore),
	)
	return review, nil
}

func (s *SandboxService) ReviewDecision(ctx context.Context, reviewID uuid.UUID, decision *domain.SandboxDecision) (*domain.SandboxReview, error) {
	review, ok := s.reviews[reviewID]
	if !ok {
		return nil, nil
	}

	now := time.Now()
	review.ReviewedAt = &now
	review.ReviewNote = decision.Note

	switch decision.Action {
	case "approve":
		review.Status = domain.SandboxStatusApproved
		for i := range review.ProposedActions {
			approved := true
			review.ProposedActions[i].Approved = &approved
		}
	case "reject":
		review.Status = domain.SandboxStatusRejected
		for i := range review.ProposedActions {
			rejected := false
			review.ProposedActions[i].Approved = &rejected
		}
	case "approve_partial":
		review.Status = domain.SandboxStatusApproved
		approvedSet := make(map[string]bool)
		for _, id := range decision.ActionIDs {
			approvedSet[id] = true
		}
		for i := range review.ProposedActions {
			val := approvedSet[review.ProposedActions[i].ID]
			review.ProposedActions[i].Approved = &val
		}
	}

	return review, nil
}

func (s *SandboxService) GetReview(ctx context.Context, reviewID uuid.UUID) (*domain.SandboxReview, error) {
	review, ok := s.reviews[reviewID]
	if !ok {
		return nil, nil
	}
	return review, nil
}

func (s *SandboxService) ListPendingReviews(ctx context.Context, projectID uuid.UUID) ([]domain.SandboxReview, error) {
	var result []domain.SandboxReview
	for _, r := range s.reviews {
		if r.ProjectID == projectID && r.Status == domain.SandboxStatusPending {
			result = append(result, *r)
		}
	}
	return result, nil
}

func (s *SandboxService) CreatePolicy(ctx context.Context, projectID uuid.UUID, input *domain.SandboxPolicyInput) (*domain.SandboxPolicy, error) {
	policy := &domain.SandboxPolicy{
		ID:              uuid.New(),
		ProjectID:       projectID,
		Name:            input.Name,
		AllowedPaths:    input.AllowedPaths,
		BlockedPaths:    input.BlockedPaths,
		AllowedCommands: input.AllowedCommands,
		BlockedCommands: input.BlockedCommands,
		RequireReview:   input.RequireReview,
		CreatedAt:       time.Now(),
	}
	if input.AllowNetwork != nil {
		policy.AllowNetwork = *input.AllowNetwork
	}
	if input.AllowEnvAccess != nil {
		policy.AllowEnvAccess = *input.AllowEnvAccess
	}
	if input.MaxFileSize != nil {
		policy.MaxFileSize = *input.MaxFileSize
	}
	if policy.RequireReview == "" {
		policy.RequireReview = "high_risk"
	}
	s.policies[policy.ID] = policy
	return policy, nil
}

func (s *SandboxService) ListPolicies(ctx context.Context, projectID uuid.UUID) ([]domain.SandboxPolicy, error) {
	var result []domain.SandboxPolicy
	for _, p := range s.policies {
		if p.ProjectID == projectID {
			result = append(result, *p)
		}
	}
	return result, nil
}

func (s *SandboxService) GetStats(ctx context.Context, projectID uuid.UUID) (*domain.SandboxStats, error) {
	stats := &domain.SandboxStats{
		ByRiskLevel: make(map[string]int),
		ByStatus:    make(map[string]int),
	}
	for _, r := range s.reviews {
		if r.ProjectID == projectID {
			stats.TotalReviews++
			stats.ByRiskLevel[r.RiskLevel]++
			stats.ByStatus[string(r.Status)]++
			if r.Status == domain.SandboxStatusPending {
				stats.PendingReviews++
			}
		}
	}
	if stats.TotalReviews > 0 {
		approved := stats.ByStatus[string(domain.SandboxStatusApproved)]
		stats.ApprovalRate = float64(approved) / float64(stats.TotalReviews)
	}
	return stats, nil
}

func (s *SandboxService) assessRisk(actions []domain.SandboxAction) (float64, string) {
	maxRisk := 0.0
	for _, a := range actions {
		risk := 20.0
		switch a.Type {
		case "file_delete":
			risk = 80.0
		case "command_exec":
			risk = 70.0
		case "network_request":
			risk = 60.0
		case "env_access":
			risk = 75.0
		case "file_write":
			risk = 40.0
			if strings.Contains(a.Target, ".env") || strings.Contains(a.Target, "secret") || strings.Contains(a.Target, ".key") {
				risk = 90.0
			}
		}
		if risk > maxRisk {
			maxRisk = risk
		}
	}
	level := "low"
	if maxRisk >= 80 {
		level = "critical"
	} else if maxRisk >= 60 {
		level = "high"
	} else if maxRisk >= 40 {
		level = "medium"
	}
	return maxRisk, level
}
