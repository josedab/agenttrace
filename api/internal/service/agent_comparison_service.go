package service

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// AgentComparisonService handles agent comparison and profiling logic
type AgentComparisonService struct {
	logger *zap.Logger
}

// NewAgentComparisonService creates a new agent comparison service
func NewAgentComparisonService(logger *zap.Logger) *AgentComparisonService {
	return &AgentComparisonService{
		logger: logger,
	}
}

type agentCharacteristics struct {
	name            string
	agentType       domain.AgentType
	quality         float64
	costPerTrace    float64
	speedMs         float64
	tokenEfficiency float64
	errorRate       float64
}

var defaultAgentChars = []agentCharacteristics{
	{name: "Claude Code", agentType: domain.AgentTypeClaudeCode, quality: 0.91, costPerTrace: 0.10, speedMs: 3500, tokenEfficiency: 0.85, errorRate: 0.05},
	{name: "Copilot", agentType: domain.AgentTypeCopilot, quality: 0.78, costPerTrace: 0.05, speedMs: 2000, tokenEfficiency: 0.75, errorRate: 0.10},
	{name: "Cursor", agentType: domain.AgentTypeCursor, quality: 0.86, costPerTrace: 0.08, speedMs: 2750, tokenEfficiency: 0.82, errorRate: 0.06},
	{name: "Aider", agentType: domain.AgentTypeAider, quality: 0.72, costPerTrace: 0.03, speedMs: 4000, tokenEfficiency: 0.70, errorRate: 0.12},
}

// CreateProfile creates a new agent profile
func (s *AgentComparisonService) CreateProfile(ctx context.Context, projectID uuid.UUID, input domain.AgentProfileInput) (*domain.AgentProfile, error) {
	now := time.Now()
	profile := &domain.AgentProfile{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        input.Name,
		AgentType:   input.AgentType,
		ModelName:   input.ModelName,
		Description: input.Description,
		Config:      input.Config,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.logger.Info("created agent profile",
		zap.String("profileId", profile.ID.String()),
		zap.String("name", profile.Name),
		zap.String("agentType", string(profile.AgentType)),
	)

	return profile, nil
}

// ListProfiles returns agent profiles for a project with demo data
func (s *AgentComparisonService) ListProfiles(ctx context.Context, projectID uuid.UUID, limit, offset int) (*domain.AgentProfileList, error) {
	s.logger.Info("listing agent profiles",
		zap.String("projectId", projectID.String()),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	now := time.Now()
	profiles := make([]domain.AgentProfile, 0, len(defaultAgentChars))

	for _, ac := range defaultAgentChars {
		profile := domain.AgentProfile{
			ID:        uuid.New(),
			ProjectID: projectID,
			Name:      ac.name,
			AgentType: ac.agentType,
			ModelName: modelNameForAgentType(ac.agentType),
			AverageMetrics: &domain.AgentMetricsSummary{
				TotalTraces:       150,
				AvgCostPerTrace:   ac.costPerTrace,
				AvgLatencyMs:      ac.speedMs,
				AvgTokensPerTrace: math.Round(ac.tokenEfficiency * 1500),
				AvgQualityScore:   ac.quality,
				ErrorRate:         ac.errorRate,
				P50LatencyMs:      math.Round(ac.speedMs * 0.85),
				P95LatencyMs:      math.Round(ac.speedMs * 1.40),
				P99LatencyMs:      math.Round(ac.speedMs * 1.80),
			},
			CreatedAt: now.Add(-30 * 24 * time.Hour),
			UpdatedAt: now,
		}
		profiles = append(profiles, profile)
	}

	total := len(profiles)
	if offset >= total {
		profiles = []domain.AgentProfile{}
	} else {
		end := offset + limit
		if end > total {
			end = total
		}
		profiles = profiles[offset:end]
	}

	return &domain.AgentProfileList{
		Profiles:   profiles,
		TotalCount: int64(total),
		HasMore:    offset+limit < total,
	}, nil
}

// GetProfile returns a specific agent profile by ID
func (s *AgentComparisonService) GetProfile(ctx context.Context, projectID, profileID uuid.UUID) (*domain.AgentProfile, error) {
	s.logger.Info("fetching agent profile",
		zap.String("projectId", projectID.String()),
		zap.String("profileId", profileID.String()),
	)

	ac := defaultAgentChars[0]
	now := time.Now()

	profile := &domain.AgentProfile{
		ID:        profileID,
		ProjectID: projectID,
		Name:      ac.name,
		AgentType: ac.agentType,
		ModelName: modelNameForAgentType(ac.agentType),
		AverageMetrics: &domain.AgentMetricsSummary{
			TotalTraces:       150,
			AvgCostPerTrace:   ac.costPerTrace,
			AvgLatencyMs:      ac.speedMs,
			AvgTokensPerTrace: math.Round(ac.tokenEfficiency * 1500),
			AvgQualityScore:   ac.quality,
			ErrorRate:         ac.errorRate,
			P50LatencyMs:      math.Round(ac.speedMs * 0.85),
			P95LatencyMs:      math.Round(ac.speedMs * 1.40),
			P99LatencyMs:      math.Round(ac.speedMs * 1.80),
		},
		CreatedAt: now.Add(-30 * 24 * time.Hour),
		UpdatedAt: now,
	}

	return profile, nil
}

// RunComparison executes a comparison across agents, generating realistic mock metrics
func (s *AgentComparisonService) RunComparison(ctx context.Context, projectID uuid.UUID, input domain.AgentComparisonInput) (*domain.AgentComparisonRun, error) {
	s.logger.Info("running agent comparison",
		zap.String("projectId", projectID.String()),
		zap.String("name", input.Name),
		zap.Int("agentCount", len(input.AgentIDs)),
	)

	metricTypes := input.MetricTypes
	if len(metricTypes) == 0 {
		metricTypes = []domain.ComparisonMetricType{
			domain.ComparisonMetricCost,
			domain.ComparisonMetricQuality,
			domain.ComparisonMetricSpeed,
			domain.ComparisonMetricTokenEfficiency,
			domain.ComparisonMetricErrorRate,
		}
	}

	type agentEntry struct {
		id    uuid.UUID
		chars agentCharacteristics
	}
	agents := make([]agentEntry, len(input.AgentIDs))
	for i, id := range input.AgentIDs {
		idx := i % len(defaultAgentChars)
		agents[i] = agentEntry{id: id, chars: defaultAgentChars[idx]}
	}

	var metrics []domain.ComparisonMetric
	normalizedScores := make(map[string]map[string]float64)
	winnerByMetric := make(map[string]uuid.UUID)

	for _, mt := range metricTypes {
		type rawEntry struct {
			agentID uuid.UUID
			name    string
			aType   domain.AgentType
			value   float64
		}
		entries := make([]rawEntry, len(agents))
		for i, a := range agents {
			entries[i] = rawEntry{
				agentID: a.id,
				name:    a.chars.name,
				aType:   a.chars.agentType,
				value:   rawValueForMetric(a.chars, mt),
			}
		}

		lowerIsBetter := mt == domain.ComparisonMetricCost ||
			mt == domain.ComparisonMetricSpeed ||
			mt == domain.ComparisonMetricErrorRate

		bestVal := entries[0].value
		for _, e := range entries[1:] {
			if lowerIsBetter && e.value < bestVal {
				bestVal = e.value
			} else if !lowerIsBetter && e.value > bestVal {
				bestVal = e.value
			}
		}

		sorted := make([]rawEntry, len(entries))
		copy(sorted, entries)
		sort.Slice(sorted, func(i, j int) bool {
			if lowerIsBetter {
				return sorted[i].value < sorted[j].value
			}
			return sorted[i].value > sorted[j].value
		})

		rankMap := make(map[uuid.UUID]int)
		for i, e := range sorted {
			rankMap[e.agentID] = i + 1
		}
		winnerByMetric[string(mt)] = sorted[0].agentID

		for _, e := range entries {
			var normalized float64
			if lowerIsBetter {
				if e.value > 0 {
					normalized = math.Round((bestVal / e.value) * 100)
				}
			} else {
				if bestVal > 0 {
					normalized = math.Round((e.value / bestVal) * 100)
				}
			}

			metrics = append(metrics, domain.ComparisonMetric{
				AgentID:         e.agentID,
				AgentName:       e.name,
				AgentType:       e.aType,
				MetricType:      mt,
				RawValue:        e.value,
				NormalizedValue: normalized,
				Rank:            rankMap[e.agentID],
			})

			agentKey := e.agentID.String()
			if normalizedScores[agentKey] == nil {
				normalizedScores[agentKey] = make(map[string]float64)
			}
			normalizedScores[agentKey][string(mt)] = normalized
		}
	}

	// Determine overall winner by highest average normalized score
	var bestAvg float64
	var overallWinner uuid.UUID
	for i, a := range agents {
		agentScores := normalizedScores[a.id.String()]
		var sum float64
		for _, v := range agentScores {
			sum += v
		}
		avg := sum / float64(len(agentScores))
		if i == 0 || avg > bestAvg {
			bestAvg = avg
			overallWinner = a.id
		}
	}

	var dateRange *domain.ComparisonDateRange
	if input.FromDate != nil && input.ToDate != nil {
		dateRange = &domain.ComparisonDateRange{
			From: *input.FromDate,
			To:   *input.ToDate,
		}
	}

	run := &domain.AgentComparisonRun{
		ID:               uuid.New(),
		ProjectID:        projectID,
		Name:             input.Name,
		AgentIDs:         input.AgentIDs,
		DateRange:        dateRange,
		Metrics:          metrics,
		NormalizedScores: normalizedScores,
		WinnerByMetric:   winnerByMetric,
		OverallWinner:    &overallWinner,
		CreatedAt:        time.Now(),
	}

	s.logger.Info("completed agent comparison",
		zap.String("comparisonId", run.ID.String()),
		zap.String("overallWinner", overallWinner.String()),
	)

	return run, nil
}

// ListComparisons returns comparisons for a project
func (s *AgentComparisonService) ListComparisons(ctx context.Context, projectID uuid.UUID, limit, offset int) (*domain.AgentComparisonList, error) {
	s.logger.Info("listing agent comparisons",
		zap.String("projectId", projectID.String()),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	return &domain.AgentComparisonList{
		Comparisons: []domain.AgentComparisonRun{},
		TotalCount:  0,
		HasMore:     false,
	}, nil
}

// GetComparison returns a specific comparison run by ID
func (s *AgentComparisonService) GetComparison(ctx context.Context, projectID, comparisonID uuid.UUID) (*domain.AgentComparisonRun, error) {
	s.logger.Info("fetching agent comparison",
		zap.String("projectId", projectID.String()),
		zap.String("comparisonId", comparisonID.String()),
	)

	run := &domain.AgentComparisonRun{
		ID:        comparisonID,
		ProjectID: projectID,
		Name:      "Demo Comparison",
		AgentIDs:  []uuid.UUID{},
		Metrics:   []domain.ComparisonMetric{},
		CreatedAt: time.Now(),
	}

	return run, nil
}

// GetTrends generates time-series trend data for agent metrics
func (s *AgentComparisonService) GetTrends(ctx context.Context, projectID uuid.UUID, agentIDs []uuid.UUID, metricType domain.ComparisonMetricType, days int) ([]domain.AgentTrendPoint, error) {
	s.logger.Info("fetching agent trends",
		zap.String("projectId", projectID.String()),
		zap.String("metricType", string(metricType)),
		zap.Int("days", days),
		zap.Int("agents", len(agentIDs)),
	)

	now := time.Now()
	var points []domain.AgentTrendPoint

	for i, agentID := range agentIDs {
		idx := i % len(defaultAgentChars)
		chars := defaultAgentChars[idx]
		baseValue := rawValueForMetric(chars, metricType)

		for d := days - 1; d >= 0; d-- {
			ts := now.Add(-time.Duration(d) * 24 * time.Hour).Truncate(24 * time.Hour)
			variation := 1.0 + float64(d%7-3)*0.02
			value := math.Round(baseValue*variation*1000) / 1000

			points = append(points, domain.AgentTrendPoint{
				Timestamp:  ts,
				AgentID:    agentID,
				MetricType: metricType,
				Value:      value,
			})
		}
	}

	return points, nil
}

// GetDashboardSummary returns a dashboard overview for agent comparisons
func (s *AgentComparisonService) GetDashboardSummary(ctx context.Context, projectID uuid.UUID) (*domain.AgentComparisonSummary, error) {
	s.logger.Info("fetching dashboard summary",
		zap.String("projectId", projectID.String()),
	)

	now := time.Now()
	qualityLeader := buildProfileFromChars(projectID, defaultAgentChars[0], now)
	costLeader := buildProfileFromChars(projectID, defaultAgentChars[3], now)
	speedLeader := buildProfileFromChars(projectID, defaultAgentChars[1], now)

	summary := &domain.AgentComparisonSummary{
		TotalProfiles:     4,
		TotalComparisons:  12,
		TopAgent:          qualityLeader,
		CostLeader:        costLeader,
		SpeedLeader:       speedLeader,
		QualityLeader:     qualityLeader,
		RecentComparisons: []domain.AgentComparisonRun{},
	}

	return summary, nil
}

func modelNameForAgentType(agentType domain.AgentType) string {
	switch agentType {
	case domain.AgentTypeClaudeCode:
		return "claude-sonnet-4-20250514"
	case domain.AgentTypeCopilot:
		return "gpt-4o"
	case domain.AgentTypeCursor:
		return "claude-sonnet-4-20250514"
	case domain.AgentTypeAider:
		return "deepseek-coder-v2"
	default:
		return "unknown"
	}
}

func rawValueForMetric(chars agentCharacteristics, metricType domain.ComparisonMetricType) float64 {
	switch metricType {
	case domain.ComparisonMetricQuality:
		return chars.quality
	case domain.ComparisonMetricCost:
		return chars.costPerTrace
	case domain.ComparisonMetricSpeed:
		return chars.speedMs
	case domain.ComparisonMetricTokenEfficiency:
		return chars.tokenEfficiency
	case domain.ComparisonMetricErrorRate:
		return chars.errorRate
	default:
		return 0
	}
}

func buildProfileFromChars(projectID uuid.UUID, chars agentCharacteristics, now time.Time) *domain.AgentProfile {
	return &domain.AgentProfile{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      chars.name,
		AgentType: chars.agentType,
		ModelName: modelNameForAgentType(chars.agentType),
		AverageMetrics: &domain.AgentMetricsSummary{
			TotalTraces:       150,
			AvgCostPerTrace:   chars.costPerTrace,
			AvgLatencyMs:      chars.speedMs,
			AvgTokensPerTrace: math.Round(chars.tokenEfficiency * 1500),
			AvgQualityScore:   chars.quality,
			ErrorRate:         chars.errorRate,
			P50LatencyMs:      math.Round(chars.speedMs * 0.85),
			P95LatencyMs:      math.Round(chars.speedMs * 1.40),
			P99LatencyMs:      math.Round(chars.speedMs * 1.80),
		},
		CreatedAt: now.Add(-30 * 24 * time.Hour),
		UpdatedAt: now,
	}
}
