package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

type FederatedInstanceRepository struct {
	db *database.PostgresDB
}

func NewFederatedInstanceRepository(db *database.PostgresDB) *FederatedInstanceRepository {
	return &FederatedInstanceRepository{db: db}
}

func (r *FederatedInstanceRepository) SaveInstance(ctx context.Context, instance *domain.FederatedInstance) error {
	query := `
		INSERT INTO federated_instances (id, name, endpoint, api_key_hash,
			privacy_level, last_sync_at, status, metrics_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		instance.ID,
		instance.Name,
		instance.Endpoint,
		instance.APIKey,
		instance.PrivacyLevel,
		instance.LastSyncAt,
		instance.Status,
		instance.MetricsCount,
		instance.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save federated instance: %w", err)
	}

	return nil
}

func (r *FederatedInstanceRepository) GetInstanceByID(ctx context.Context, id uuid.UUID) (*domain.FederatedInstance, error) {
	query := `
		SELECT id, name, endpoint, api_key_hash,
			privacy_level, last_sync_at, status, metrics_count, created_at
		FROM federated_instances
		WHERE id = $1
	`

	var inst domain.FederatedInstance
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&inst.ID, &inst.Name, &inst.Endpoint, &inst.APIKey,
		&inst.PrivacyLevel, &inst.LastSyncAt, &inst.Status,
		&inst.MetricsCount, &inst.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("federated instance")
		}
		return nil, fmt.Errorf("failed to get federated instance: %w", err)
	}

	return &inst, nil
}

func (r *FederatedInstanceRepository) ListInstances(ctx context.Context, projectID uuid.UUID) ([]domain.FederatedInstance, error) {
	query := `
		SELECT id, name, endpoint, api_key_hash,
			privacy_level, last_sync_at, status, metrics_count, created_at
		FROM federated_instances
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list federated instances: %w", err)
	}
	defer rows.Close()

	var instances []domain.FederatedInstance
	for rows.Next() {
		var inst domain.FederatedInstance
		if err := rows.Scan(
			&inst.ID, &inst.Name, &inst.Endpoint, &inst.APIKey,
			&inst.PrivacyLevel, &inst.LastSyncAt, &inst.Status,
			&inst.MetricsCount, &inst.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan federated instance: %w", err)
		}
		instances = append(instances, inst)
	}

	return instances, nil
}

func (r *FederatedInstanceRepository) UpdateInstanceStatus(ctx context.Context, id uuid.UUID, status string, syncAt *time.Time) error {
	query := `
		UPDATE federated_instances
		SET status = $2, last_sync_at = COALESCE($3, last_sync_at)
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query, id, status, syncAt)
	if err != nil {
		return fmt.Errorf("failed to update federated instance status: %w", err)
	}

	return nil
}

func (r *FederatedInstanceRepository) SaveMetric(ctx context.Context, metric *domain.FederatedMetric) error {
	modelDistJSON, err := json.Marshal(metric.ModelDistribution)
	if err != nil {
		return fmt.Errorf("failed to marshal model distribution: %w", err)
	}

	query := `
		INSERT INTO federated_metrics (instance_id, metric_type, period_start, period_end,
			value, p50, p95, p99, std_dev, sample_count, model_distribution)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		metric.InstanceID,
		metric.MetricType,
		metric.Period.Start,
		metric.Period.End,
		metric.Value,
		metric.P50,
		metric.P95,
		metric.P99,
		metric.StdDev,
		metric.SampleCount,
		modelDistJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to save federated metric: %w", err)
	}

	return nil
}

func (r *FederatedInstanceRepository) ListMetrics(ctx context.Context, instanceID uuid.UUID, metricType *domain.FederatedMetricType, limit int) ([]domain.FederatedMetric, error) {
	var query string
	var args []interface{}

	if metricType != nil {
		query = `
			SELECT instance_id, metric_type, period_start, period_end,
				value, p50, p95, p99, std_dev, sample_count, model_distribution
			FROM federated_metrics
			WHERE instance_id = $1 AND metric_type = $2
			ORDER BY period_start DESC
			LIMIT $3
		`
		args = []interface{}{instanceID, *metricType, limit}
	} else {
		query = `
			SELECT instance_id, metric_type, period_start, period_end,
				value, p50, p95, p99, std_dev, sample_count, model_distribution
			FROM federated_metrics
			WHERE instance_id = $1
			ORDER BY period_start DESC
			LIMIT $2
		`
		args = []interface{}{instanceID, limit}
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list federated metrics: %w", err)
	}
	defer rows.Close()

	var metrics []domain.FederatedMetric
	for rows.Next() {
		var m domain.FederatedMetric
		var modelDistJSON []byte

		if err := rows.Scan(
			&m.InstanceID, &m.MetricType, &m.Period.Start, &m.Period.End,
			&m.Value, &m.P50, &m.P95, &m.P99, &m.StdDev,
			&m.SampleCount, &modelDistJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan federated metric: %w", err)
		}

		if len(modelDistJSON) > 0 {
			if err := json.Unmarshal(modelDistJSON, &m.ModelDistribution); err != nil {
				return nil, fmt.Errorf("failed to unmarshal model distribution: %w", err)
			}
		}

		metrics = append(metrics, m)
	}

	return metrics, nil
}
