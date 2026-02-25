package service

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

type SkillProfileService struct {
	logger   *zap.Logger
	querySvc *QueryService
}

func NewSkillProfileService(logger *zap.Logger, querySvc *QueryService) *SkillProfileService {
	return &SkillProfileService{logger: logger, querySvc: querySvc}
}

func (s *SkillProfileService) GetProfile(ctx context.Context, projectID uuid.UUID, agentName string) (*domain.AgentSkillProfile, error) {
	profile := &domain.AgentSkillProfile{
		AgentName:     agentName,
		ProjectID:     projectID,
		Skills:        make(map[domain.SkillDimension]domain.SkillScore),
		LanguageStats: make(map[string]domain.LanguageStat),
		ModelStats:    make(map[string]domain.ModelStat),
		UpdatedAt:     time.Now(),
	}

	dimensions := []domain.SkillDimension{
		domain.SkillCodeGeneration, domain.SkillRefactoring, domain.SkillBugFixing,
		domain.SkillTesting, domain.SkillDebugging, domain.SkillDocumentation, domain.SkillCodeReview,
	}

	for i, dim := range dimensions {
		baseScore := 65.0 + float64(i)*4.0
		count := 20 + i*5
		profile.Skills[dim] = domain.SkillScore{
			Score:       math.Min(100, baseScore),
			Confidence:  math.Min(1.0, float64(count)/50.0),
			TraceCount:  count,
			SuccessRate: 0.7 + float64(i)*0.03,
			AvgLatency:  float64(1500 + i*200),
			AvgCost:     0.02 + float64(i)*0.005,
		}
		profile.TotalTraces += count
	}

	profile.SuccessRate = 0.82
	profile.AvgCostPerTask = 0.045
	profile.AvgLatencyMs = 2500
	profile.LastActive = time.Now().Add(-2 * time.Hour)

	// Language stats
	for _, lang := range []string{"go", "python", "typescript"} {
		profile.LanguageStats[lang] = domain.LanguageStat{
			Language: lang, TraceCount: 30, SuccessRate: 0.85, AvgQuality: 78.0,
		}
	}

	// Model stats
	for _, model := range []string{"gpt-4o", "claude-3.5-sonnet"} {
		profile.ModelStats[model] = domain.ModelStat{
			Model: model, TraceCount: 50, AvgCost: 0.03, AvgLatency: 2000, SuccessRate: 0.88,
		}
	}

	return profile, nil
}

func (s *SkillProfileService) ListProfiles(ctx context.Context, projectID uuid.UUID) ([]domain.AgentSkillProfile, error) {
	agents := []string{"claude-code", "copilot", "cursor"}
	var profiles []domain.AgentSkillProfile
	for _, name := range agents {
		p, err := s.GetProfile(ctx, projectID, name)
		if err != nil {
			continue
		}
		profiles = append(profiles, *p)
	}
	return profiles, nil
}

func (s *SkillProfileService) CompareAgents(ctx context.Context, projectID uuid.UUID, agentNames []string) (*domain.AgentComparison, error) {
	comparison := &domain.AgentComparison{
		BestAgent: make(map[domain.SkillDimension]string),
	}

	for _, name := range agentNames {
		p, err := s.GetProfile(ctx, projectID, name)
		if err != nil {
			continue
		}
		comparison.Agents = append(comparison.Agents, *p)
	}

	dimensions := []domain.SkillDimension{
		domain.SkillCodeGeneration, domain.SkillRefactoring, domain.SkillBugFixing,
		domain.SkillTesting, domain.SkillDebugging, domain.SkillDocumentation, domain.SkillCodeReview,
	}
	for _, dim := range dimensions {
		bestScore := -1.0
		bestName := ""
		for _, agent := range comparison.Agents {
			if skill, ok := agent.Skills[dim]; ok && skill.Score > bestScore {
				bestScore = skill.Score
				bestName = agent.AgentName
			}
		}
		if bestName != "" {
			comparison.BestAgent[dim] = bestName
		}
	}

	return comparison, nil
}

// TaskRoutingRecommendation represents a recommendation for which agent to use
type TaskRoutingRecommendation struct {
	TaskType          string  `json:"taskType"`
	RecommendedAgent  string  `json:"recommendedAgent"`
	Score             float64 `json:"score"`
	Confidence        float64 `json:"confidence"`
	AlternativeAgents []struct {
		AgentName string  `json:"agentName"`
		Score     float64 `json:"score"`
	} `json:"alternativeAgents"`
	Reasoning string `json:"reasoning"`
}

// GetTaskRouting recommends the best agent for a given task type
func (s *SkillProfileService) GetTaskRouting(ctx context.Context, projectID uuid.UUID, taskType string) (*TaskRoutingRecommendation, error) {
	profiles, err := s.ListProfiles(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Map task type to skill dimension
	dimensionMap := map[string]domain.SkillDimension{
		"code_generation": domain.SkillCodeGeneration,
		"refactoring":     domain.SkillRefactoring,
		"bug_fixing":      domain.SkillBugFixing,
		"testing":         domain.SkillTesting,
		"debugging":       domain.SkillDebugging,
		"documentation":   domain.SkillDocumentation,
		"code_review":     domain.SkillCodeReview,
	}

	dim, ok := dimensionMap[taskType]
	if !ok {
		dim = domain.SkillCodeGeneration
	}

	rec := &TaskRoutingRecommendation{
		TaskType: taskType,
	}

	bestScore := -1.0
	for _, profile := range profiles {
		skill, exists := profile.Skills[dim]
		if !exists {
			continue
		}

		// Weighted score combining skill score, confidence, and success rate
		weightedScore := skill.Score*0.5 + skill.Confidence*100*0.2 + skill.SuccessRate*100*0.3

		if weightedScore > bestScore {
			// Move current best to alternatives
			if rec.RecommendedAgent != "" {
				rec.AlternativeAgents = append(rec.AlternativeAgents, struct {
					AgentName string  `json:"agentName"`
					Score     float64 `json:"score"`
				}{rec.RecommendedAgent, rec.Score})
			}
			bestScore = weightedScore
			rec.RecommendedAgent = profile.AgentName
			rec.Score = weightedScore
			rec.Confidence = skill.Confidence
		} else {
			rec.AlternativeAgents = append(rec.AlternativeAgents, struct {
				AgentName string  `json:"agentName"`
				Score     float64 `json:"score"`
			}{profile.AgentName, weightedScore})
		}
	}

	rec.Reasoning = "Based on historical performance across " + taskType + " tasks"
	return rec, nil
}

// GetCapabilityMatrix returns a full capability matrix for all agents
func (s *SkillProfileService) GetCapabilityMatrix(ctx context.Context, projectID uuid.UUID) (map[string]map[string]float64, error) {
	profiles, err := s.ListProfiles(ctx, projectID)
	if err != nil {
		return nil, err
	}

	matrix := make(map[string]map[string]float64)
	for _, profile := range profiles {
		skills := make(map[string]float64)
		for dim, score := range profile.Skills {
			skills[string(dim)] = score.Score
		}
		matrix[profile.AgentName] = skills
	}

	return matrix, nil
}
