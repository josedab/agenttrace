package clickhouse

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// SkillProfileRepository queries ClickHouse for agent performance aggregations
type SkillProfileRepository struct {
	db *database.ClickHouseDB
}

// NewSkillProfileRepository creates a new repository
func NewSkillProfileRepository(db *database.ClickHouseDB) *SkillProfileRepository {
	return &SkillProfileRepository{db: db}
}

// GetAgentNames returns distinct agent names (from trace metadata) for a project
func (r *SkillProfileRepository) GetAgentNames(ctx context.Context, projectID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT JSONExtractString(metadata, 'agent_name') as agent_name
		FROM agenttrace.traces
		WHERE project_id = ?
		AND agent_name != ''
		ORDER BY agent_name
		LIMIT 100
	`
	rows, err := r.db.Conn.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent names: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		if name != "" {
			names = append(names, name)
		}
	}
	return names, nil
}

// GetAgentTraceStats returns aggregate trace statistics for an agent
func (r *SkillProfileRepository) GetAgentTraceStats(ctx context.Context, projectID uuid.UUID, agentName string) (*domain.AgentSkillProfile, error) {
	query := `
		SELECT
			count() as trace_count,
			avg(total_cost) as avg_cost,
			avg(duration_ms) as avg_latency,
			countIf(level != 'ERROR') / count() as success_rate,
			max(start_time) as last_active
		FROM agenttrace.traces
		WHERE project_id = ?
		AND JSONExtractString(metadata, 'agent_name') = ?
		AND start_time >= now() - INTERVAL 30 DAY
	`

	profile := &domain.AgentSkillProfile{
		AgentName:     agentName,
		ProjectID:     projectID,
		Skills:        make(map[domain.SkillDimension]domain.SkillScore),
		LanguageStats: make(map[string]domain.LanguageStat),
		ModelStats:    make(map[string]domain.ModelStat),
	}

	row := r.db.Conn.QueryRow(ctx, query, projectID, agentName)
	err := row.Scan(
		&profile.TotalTraces,
		&profile.AvgCostPerTask,
		&profile.AvgLatencyMs,
		&profile.SuccessRate,
		&profile.LastActive,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent stats: %w", err)
	}

	return profile, nil
}

// GetModelBreakdown returns per-model statistics for an agent
func (r *SkillProfileRepository) GetModelBreakdown(ctx context.Context, projectID uuid.UUID, agentName string) (map[string]domain.ModelStat, error) {
	query := `
		SELECT
			o.model,
			count() as trace_count,
			avg(o.total_cost) as avg_cost,
			avg(o.duration_ms) as avg_latency,
			countIf(o.level != 'ERROR') / count() as success_rate
		FROM agenttrace.observations o
		INNER JOIN agenttrace.traces t ON o.trace_id = t.id AND o.project_id = t.project_id
		WHERE t.project_id = ?
		AND JSONExtractString(t.metadata, 'agent_name') = ?
		AND o.type = 'GENERATION'
		AND o.model != ''
		GROUP BY o.model
		ORDER BY trace_count DESC
		LIMIT 20
	`

	rows, err := r.db.Conn.Query(ctx, query, projectID, agentName)
	if err != nil {
		return nil, fmt.Errorf("failed to get model breakdown: %w", err)
	}
	defer rows.Close()

	result := make(map[string]domain.ModelStat)
	for rows.Next() {
		var stat domain.ModelStat
		if err := rows.Scan(&stat.Model, &stat.TraceCount, &stat.AvgCost, &stat.AvgLatency, &stat.SuccessRate); err != nil {
			return nil, err
		}
		result[stat.Model] = stat
	}
	return result, nil
}
