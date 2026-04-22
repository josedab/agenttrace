package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// ReplayPlanRepository persists safe replay plans.
type ReplayPlanRepository struct {
	db *database.PostgresDB
}

// NewReplayPlanRepository creates a replay plan repository.
func NewReplayPlanRepository(db *database.PostgresDB) *ReplayPlanRepository {
	return &ReplayPlanRepository{db: db}
}

// Create persists a replay plan.
func (r *ReplayPlanRepository) Create(ctx context.Context, plan *domain.ReplayPlan) error {
	request, capabilities, result, err := marshalReplayPlan(plan)
	if err != nil {
		return err
	}

	_, err = r.db.Pool.Exec(ctx, `
		INSERT INTO replay_plans (
			id, project_id, trace_id, checkpoint_id, status, request,
			capabilities, result, failure_reason, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`,
		plan.ID,
		plan.ProjectID,
		plan.TraceID,
		plan.CheckpointID,
		plan.Status,
		request,
		capabilities,
		result,
		plan.FailureReason,
		plan.CreatedBy,
		plan.CreatedAt,
		plan.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create replay plan: %w", err)
	}
	return nil
}

// GetByID retrieves a replay plan only within its project.
func (r *ReplayPlanRepository) GetByID(
	ctx context.Context,
	projectID, planID uuid.UUID,
) (*domain.ReplayPlan, error) {
	var plan domain.ReplayPlan
	var requestJSON, capabilitiesJSON, resultJSON []byte

	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, project_id, trace_id, checkpoint_id, status, request,
			capabilities, result, failure_reason, created_by, created_at, updated_at
		FROM replay_plans
		WHERE project_id = $1 AND id = $2
	`, projectID, planID).Scan(
		&plan.ID,
		&plan.ProjectID,
		&plan.TraceID,
		&plan.CheckpointID,
		&plan.Status,
		&requestJSON,
		&capabilitiesJSON,
		&resultJSON,
		&plan.FailureReason,
		&plan.CreatedBy,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("replay plan")
		}
		return nil, fmt.Errorf("get replay plan: %w", err)
	}

	if err := unmarshalReplayPlan(&plan, requestJSON, capabilitiesJSON, resultJSON); err != nil {
		return nil, err
	}
	return &plan, nil
}

// TransitionStatus atomically moves a plan into a new status only when it is in
// the expected status, so concurrent executions cannot both start.
// It returns the updated plan, a conflict when the plan is in another status,
// or a not-found error when the plan does not belong to the project.
func (r *ReplayPlanRepository) TransitionStatus(
	ctx context.Context,
	projectID, planID uuid.UUID,
	transition domain.ReplayPlanTransition,
) (*domain.ReplayPlan, error) {
	query := `
		UPDATE replay_plans
		SET status = $4, updated_at = $5, failure_reason = ''
		WHERE project_id = $1 AND id = $2 AND status = $3
		RETURNING id, project_id, trace_id, checkpoint_id, status, request,
			capabilities, result, failure_reason, created_by, created_at, updated_at
	`
	args := []any{projectID, planID, transition.From, transition.To, transition.At}

	var plan domain.ReplayPlan
	var requestJSON, capabilitiesJSON, resultJSON []byte
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(
		&plan.ID,
		&plan.ProjectID,
		&plan.TraceID,
		&plan.CheckpointID,
		&plan.Status,
		&requestJSON,
		&capabilitiesJSON,
		&resultJSON,
		&plan.FailureReason,
		&plan.CreatedBy,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, r.transitionRejection(ctx, projectID, planID, transition)
		}
		return nil, fmt.Errorf("transition replay plan: %w", err)
	}

	if err := unmarshalReplayPlan(&plan, requestJSON, capabilitiesJSON, resultJSON); err != nil {
		return nil, err
	}
	return &plan, nil
}

// transitionRejection explains why a conditional transition matched no row.
func (r *ReplayPlanRepository) transitionRejection(
	ctx context.Context,
	projectID, planID uuid.UUID,
	transition domain.ReplayPlanTransition,
) error {
	current, err := r.GetByID(ctx, projectID, planID)
	if err != nil {
		return err
	}
	return apperrors.Conflict(
		fmt.Sprintf(
			"replay plan must be %s, current status is %s",
			transition.From,
			current.Status,
		),
	)
}

// Update persists replay plan state transitions and results.
func (r *ReplayPlanRepository) Update(ctx context.Context, plan *domain.ReplayPlan) error {
	request, capabilities, result, err := marshalReplayPlan(plan)
	if err != nil {
		return err
	}

	commandTag, err := r.db.Pool.Exec(ctx, `
		UPDATE replay_plans
		SET status = $3, request = $4, capabilities = $5, result = $6,
			failure_reason = $7, updated_at = $8
		WHERE project_id = $1 AND id = $2
	`,
		plan.ProjectID,
		plan.ID,
		plan.Status,
		request,
		capabilities,
		result,
		plan.FailureReason,
		plan.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update replay plan: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return apperrors.NotFound("replay plan")
	}
	return nil
}

func marshalReplayPlan(
	plan *domain.ReplayPlan,
) (requestJSON, capabilitiesJSON, resultJSON []byte, resultErr error) {
	requestJSON, err := json.Marshal(plan.Request)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("marshal replay request: %w", err)
	}
	capabilitiesJSON, err = json.Marshal(plan.Capabilities)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("marshal replay capabilities: %w", err)
	}
	if plan.Result != nil {
		resultJSON, err = json.Marshal(plan.Result)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("marshal replay result: %w", err)
		}
	}
	return requestJSON, capabilitiesJSON, resultJSON, nil
}

func unmarshalReplayPlan(
	plan *domain.ReplayPlan,
	requestJSON, capabilitiesJSON, resultJSON []byte,
) error {
	if err := json.Unmarshal(requestJSON, &plan.Request); err != nil {
		return fmt.Errorf("unmarshal replay request: %w", err)
	}
	if err := json.Unmarshal(capabilitiesJSON, &plan.Capabilities); err != nil {
		return fmt.Errorf("unmarshal replay capabilities: %w", err)
	}
	if len(resultJSON) > 0 && string(resultJSON) != "null" {
		plan.Result = &domain.ReplayPlanResult{}
		if err := json.Unmarshal(resultJSON, plan.Result); err != nil {
			return fmt.Errorf("unmarshal replay result: %w", err)
		}
	}
	return nil
}
