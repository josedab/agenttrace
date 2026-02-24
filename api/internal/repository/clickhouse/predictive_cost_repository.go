package clickhouse

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// PredictiveCostStats contains historical statistics for cost prediction
type PredictiveCostStats struct {
	AvgCost      float64 `json:"avgCost"`
	AvgLatencyMs float64 `json:"avgLatencyMs"`
	AvgTokens    float64 `json:"avgTokens"`
	AvgQuality   float64 `json:"avgQuality"`
	TraceCount   int     `json:"traceCount"`
	P50Cost      float64 `json:"p50Cost"`
	P90Cost      float64 `json:"p90Cost"`
	P99Cost      float64 `json:"p99Cost"`
}

// PredictiveCostRepository queries ClickHouse for historical cost data
type PredictiveCostRepository struct {
	db *database.ClickHouseDB
}

// NewPredictiveCostRepository creates a new repository
func NewPredictiveCostRepository(db *database.ClickHouseDB) *PredictiveCostRepository {
	return &PredictiveCostRepository{db: db}
}

// GetHistoricalStats returns aggregate cost statistics for prediction
func (r *PredictiveCostRepository) GetHistoricalStats(ctx context.Context, projectID uuid.UUID, model string) (*PredictiveCostStats, error) {
	query := `
		SELECT
			avg(total_cost) as avg_cost,
			avg(duration_ms) as avg_latency_ms,
			avg(total_tokens) as avg_tokens,
			count() as trace_count,
			quantile(0.5)(total_cost) as p50_cost,
			quantile(0.9)(total_cost) as p90_cost,
			quantile(0.99)(total_cost) as p99_cost
		FROM agenttrace.traces
		WHERE project_id = ?
		AND start_time >= now() - INTERVAL 30 DAY
	`

	args := []any{projectID}
	if model != "" {
		query = `
			SELECT
				avg(o.total_cost) as avg_cost,
				avg(o.duration_ms) as avg_latency_ms,
				avg(o.usage_input_tokens + o.usage_output_tokens) as avg_tokens,
				count() as trace_count,
				quantile(0.5)(o.total_cost) as p50_cost,
				quantile(0.9)(o.total_cost) as p90_cost,
				quantile(0.99)(o.total_cost) as p99_cost
			FROM agenttrace.observations o
			WHERE o.project_id = ? AND o.model = ?
			AND o.start_time >= now() - INTERVAL 30 DAY
		`
		args = append(args, model)
	}

	stats := &PredictiveCostStats{}
	row := r.db.Conn.QueryRow(ctx, query, args...)
	err := row.Scan(&stats.AvgCost, &stats.AvgLatencyMs, &stats.AvgTokens,
		&stats.TraceCount, &stats.P50Cost, &stats.P90Cost, &stats.P99Cost)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical stats: %w", err)
	}
	return stats, nil
}

// GetModelCostDistribution returns cost distribution per model
func (r *PredictiveCostRepository) GetModelCostDistribution(ctx context.Context, projectID uuid.UUID) (map[string]*PredictiveCostStats, error) {
	query := `
		SELECT
			o.model,
			avg(o.total_cost) as avg_cost,
			avg(o.duration_ms) as avg_latency_ms,
			avg(o.usage_input_tokens + o.usage_output_tokens) as avg_tokens,
			count() as trace_count,
			quantile(0.5)(o.total_cost) as p50_cost,
			quantile(0.9)(o.total_cost) as p90_cost,
			quantile(0.99)(o.total_cost) as p99_cost
		FROM agenttrace.observations o
		WHERE o.project_id = ? AND o.model != ''
		AND o.start_time >= now() - INTERVAL 30 DAY
		GROUP BY o.model
		ORDER BY trace_count DESC
		LIMIT 20
	`

	rows, err := r.db.Conn.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model distribution: %w", err)
	}
	defer rows.Close()

	result := make(map[string]*PredictiveCostStats)
	for rows.Next() {
		var model string
		stats := &PredictiveCostStats{}
		if err := rows.Scan(&model, &stats.AvgCost, &stats.AvgLatencyMs, &stats.AvgTokens,
			&stats.TraceCount, &stats.P50Cost, &stats.P90Cost, &stats.P99Cost); err != nil {
			return nil, err
		}
		result[model] = stats
	}
	return result, nil
}
