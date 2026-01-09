package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"agenttrace/internal/domain"
)

type AuditRepository struct {
	db *sqlx.DB
}

func NewAuditRepository(db *sqlx.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// CreateAuditLog creates a new audit log entry
func (r *AuditRepository) CreateAuditLog(ctx context.Context, input *domain.AuditLogInput) (*domain.AuditLog, error) {
	id := uuid.New()
	now := time.Now()

	metadataJSON, err := json.Marshal(input.Metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	var changesJSON []byte
	if input.Changes != nil {
		changesJSON, err = json.Marshal(input.Changes)
		if err != nil {
			changesJSON = nil
		}
	}

	query := `
		INSERT INTO audit_logs (
			id, organization_id, actor_id, actor_email, actor_type,
			action, resource_type, resource_id, resource_name, description,
			metadata, changes, ip_address, user_agent, request_id, session_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`

	_, err = r.db.ExecContext(ctx, query,
		id, input.OrganizationID, input.ActorID, input.ActorEmail, input.ActorType,
		input.Action, input.ResourceType, input.ResourceID, input.ResourceName, input.Description,
		metadataJSON, changesJSON, input.IPAddress, input.UserAgent, input.RequestID, input.SessionID, now,
	)
	if err != nil {
		return nil, err
	}

	return &domain.AuditLog{
		ID:             id,
		OrganizationID: input.OrganizationID,
		ActorID:        input.ActorID,
		ActorEmail:     input.ActorEmail,
		ActorType:      input.ActorType,
		Action:         input.Action,
		ResourceType:   input.ResourceType,
		ResourceID:     input.ResourceID,
		ResourceName:   input.ResourceName,
		Description:    input.Description,
		Metadata:       input.Metadata,
		Changes:        input.Changes,
		IPAddress:      input.IPAddress,
		UserAgent:      input.UserAgent,
		RequestID:      input.RequestID,
		SessionID:      input.SessionID,
		CreatedAt:      now,
	}, nil
}

// GetAuditLog retrieves a single audit log entry
func (r *AuditRepository) GetAuditLog(ctx context.Context, orgID, logID uuid.UUID) (*domain.AuditLog, error) {
	query := `
		SELECT id, organization_id, actor_id, actor_email, actor_type,
			action, resource_type, resource_id, resource_name, description,
			metadata, changes, ip_address, user_agent, request_id, session_id, created_at
		FROM audit_logs
		WHERE id = $1 AND organization_id = $2`

	var log domain.AuditLog
	var metadataJSON, changesJSON []byte

	err := r.db.QueryRowContext(ctx, query, logID, orgID).Scan(
		&log.ID, &log.OrganizationID, &log.ActorID, &log.ActorEmail, &log.ActorType,
		&log.Action, &log.ResourceType, &log.ResourceID, &log.ResourceName, &log.Description,
		&metadataJSON, &changesJSON, &log.IPAddress, &log.UserAgent, &log.RequestID, &log.SessionID, &log.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &log.Metadata)
	}
	if changesJSON != nil {
		json.Unmarshal(changesJSON, &log.Changes)
	}

	return &log, nil
}

// ListAuditLogs retrieves audit logs with filtering and pagination
func (r *AuditRepository) ListAuditLogs(ctx context.Context, filter *domain.AuditLogFilter) (*domain.AuditLogList, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.OrganizationID != nil {
		conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argNum))
		args = append(args, *filter.OrganizationID)
		argNum++
	}

	if filter.ActorID != nil {
		conditions = append(conditions, fmt.Sprintf("actor_id = $%d", argNum))
		args = append(args, *filter.ActorID)
		argNum++
	}

	if filter.Action != nil {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argNum))
		args = append(args, *filter.Action)
		argNum++
	}

	if len(filter.Actions) > 0 {
		placeholders := make([]string, len(filter.Actions))
		for i, action := range filter.Actions {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, action)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("action IN (%s)", strings.Join(placeholders, ", ")))
	}

	if filter.ResourceType != nil {
		conditions = append(conditions, fmt.Sprintf("resource_type = $%d", argNum))
		args = append(args, *filter.ResourceType)
		argNum++
	}

	if filter.ResourceID != nil {
		conditions = append(conditions, fmt.Sprintf("resource_id = $%d", argNum))
		args = append(args, *filter.ResourceID)
		argNum++
	}

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argNum))
		args = append(args, *filter.StartTime)
		argNum++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argNum))
		args = append(args, *filter.EndTime)
		argNum++
	}

	if filter.IPAddress != nil {
		conditions = append(conditions, fmt.Sprintf("ip_address = $%d", argNum))
		args = append(args, *filter.IPAddress)
		argNum++
	}

	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		conditions = append(conditions, fmt.Sprintf(
			"to_tsvector('english', coalesce(description, '') || ' ' || coalesce(resource_name, '')) @@ plainto_tsquery('english', $%d)",
			argNum,
		))
		args = append(args, *filter.SearchQuery)
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", whereClause)
	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, err
	}

	// Get data
	limit := 50
	if filter.Limit > 0 && filter.Limit <= 1000 {
		limit = filter.Limit
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, organization_id, actor_id, actor_email, actor_type,
			action, resource_type, resource_id, resource_name, description,
			metadata, changes, ip_address, user_agent, request_id, session_id, created_at
		FROM audit_logs
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argNum, argNum+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		var metadataJSON, changesJSON []byte

		if err := rows.Scan(
			&log.ID, &log.OrganizationID, &log.ActorID, &log.ActorEmail, &log.ActorType,
			&log.Action, &log.ResourceType, &log.ResourceID, &log.ResourceName, &log.Description,
			&metadataJSON, &changesJSON, &log.IPAddress, &log.UserAgent, &log.RequestID, &log.SessionID, &log.CreatedAt,
		); err != nil {
			return nil, err
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &log.Metadata)
		}
		if changesJSON != nil {
			json.Unmarshal(changesJSON, &log.Changes)
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.AuditLogList{
		Data:       logs,
		TotalCount: totalCount,
		HasMore:    offset+len(logs) < totalCount,
	}, nil
}

// GetAuditSummary returns aggregated audit statistics
func (r *AuditRepository) GetAuditSummary(ctx context.Context, orgID uuid.UUID, period string) (*domain.AuditSummary, error) {
	var interval string
	switch period {
	case "day":
		interval = "1 day"
	case "week":
		interval = "7 days"
	case "month":
		interval = "30 days"
	default:
		interval = "7 days"
		period = "week"
	}

	summary := &domain.AuditSummary{
		OrganizationID:   orgID,
		Period:           period,
		EventsByAction:   make(map[domain.AuditAction]int),
		EventsByResource: make(map[domain.AuditResourceType]int),
	}

	// Total events
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM audit_logs WHERE organization_id = $1 AND created_at > NOW() - $2::INTERVAL",
		orgID, interval,
	).Scan(&summary.TotalEvents)
	if err != nil {
		return nil, err
	}

	// Events by action
	actionRows, err := r.db.QueryContext(ctx, `
		SELECT action, COUNT(*) as count
		FROM audit_logs
		WHERE organization_id = $1 AND created_at > NOW() - $2::INTERVAL
		GROUP BY action`,
		orgID, interval)
	if err != nil {
		return nil, err
	}
	defer actionRows.Close()

	for actionRows.Next() {
		var action string
		var count int
		if err := actionRows.Scan(&action, &count); err != nil {
			return nil, err
		}
		summary.EventsByAction[domain.AuditAction(action)] = count
	}

	// Events by resource type
	resourceRows, err := r.db.QueryContext(ctx, `
		SELECT resource_type, COUNT(*) as count
		FROM audit_logs
		WHERE organization_id = $1 AND created_at > NOW() - $2::INTERVAL
		GROUP BY resource_type`,
		orgID, interval)
	if err != nil {
		return nil, err
	}
	defer resourceRows.Close()

	for resourceRows.Next() {
		var resourceType string
		var count int
		if err := resourceRows.Scan(&resourceType, &count); err != nil {
			return nil, err
		}
		summary.EventsByResource[domain.AuditResourceType(resourceType)] = count
	}

	// Unique actors
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT actor_email)
		FROM audit_logs
		WHERE organization_id = $1 AND created_at > NOW() - $2::INTERVAL`,
		orgID, interval,
	).Scan(&summary.UniqueActors)
	if err != nil {
		return nil, err
	}

	// Top actors
	topActorRows, err := r.db.QueryContext(ctx, `
		SELECT actor_id, actor_email, actor_type, COUNT(*) as event_count
		FROM audit_logs
		WHERE organization_id = $1 AND created_at > NOW() - $2::INTERVAL
		GROUP BY actor_id, actor_email, actor_type
		ORDER BY event_count DESC
		LIMIT 10`,
		orgID, interval)
	if err != nil {
		return nil, err
	}
	defer topActorRows.Close()

	for topActorRows.Next() {
		var actor domain.AuditActorSummary
		if err := topActorRows.Scan(&actor.ActorID, &actor.ActorEmail, &actor.ActorType, &actor.EventCount); err != nil {
			return nil, err
		}
		summary.TopActors = append(summary.TopActors, actor)
	}

	return summary, nil
}

// Retention Policy methods

func (r *AuditRepository) GetRetentionPolicy(ctx context.Context, orgID uuid.UUID) (*domain.AuditRetentionPolicy, error) {
	query := `
		SELECT id, organization_id, retention_days, enabled, created_at, updated_at
		FROM audit_retention_policies
		WHERE organization_id = $1`

	var policy domain.AuditRetentionPolicy
	err := r.db.QueryRowContext(ctx, query, orgID).Scan(
		&policy.ID, &policy.OrganizationID, &policy.RetentionDays, &policy.Enabled,
		&policy.CreatedAt, &policy.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &policy, err
}

func (r *AuditRepository) UpsertRetentionPolicy(ctx context.Context, policy *domain.AuditRetentionPolicy) error {
	query := `
		INSERT INTO audit_retention_policies (id, organization_id, retention_days, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (organization_id) DO UPDATE SET
			retention_days = EXCLUDED.retention_days,
			enabled = EXCLUDED.enabled,
			updated_at = NOW()`

	if policy.ID == uuid.Nil {
		policy.ID = uuid.New()
	}
	now := time.Now()
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = now
	}
	policy.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		policy.ID, policy.OrganizationID, policy.RetentionDays, policy.Enabled,
		policy.CreatedAt, policy.UpdatedAt,
	)
	return err
}

func (r *AuditRepository) ApplyRetentionPolicy(ctx context.Context, orgID uuid.UUID) (int64, error) {
	var deletedCount int64
	err := r.db.QueryRowContext(ctx, "SELECT apply_audit_retention_policy($1)", orgID).Scan(&deletedCount)
	return deletedCount, err
}

// DeleteAuditLogsBefore deletes audit logs older than the specified time
func (r *AuditRepository) DeleteAuditLogsBefore(ctx context.Context, orgID uuid.UUID, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM audit_logs WHERE organization_id = $1 AND created_at < $2",
		orgID, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Export Job methods

type AuditExportJob struct {
	ID             uuid.UUID  `db:"id"`
	OrganizationID uuid.UUID  `db:"organization_id"`
	RequestedBy    *uuid.UUID `db:"requested_by"`
	Status         string     `db:"status"`
	Filter         string     `db:"filter"`
	Format         string     `db:"format"`
	Compress       bool       `db:"compress"`
	FilePath       *string    `db:"file_path"`
	FileSize       *int64     `db:"file_size"`
	RecordCount    *int       `db:"record_count"`
	Error          *string    `db:"error"`
	StartedAt      *time.Time `db:"started_at"`
	CompletedAt    *time.Time `db:"completed_at"`
	ExpiresAt      *time.Time `db:"expires_at"`
	CreatedAt      time.Time  `db:"created_at"`
}

func (r *AuditRepository) CreateExportJob(ctx context.Context, orgID uuid.UUID, requestedBy *uuid.UUID, filter *domain.AuditLogFilter, format string, compress bool) (*AuditExportJob, error) {
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}

	job := &AuditExportJob{
		ID:             uuid.New(),
		OrganizationID: orgID,
		RequestedBy:    requestedBy,
		Status:         "pending",
		Filter:         string(filterJSON),
		Format:         format,
		Compress:       compress,
		CreatedAt:      time.Now(),
	}

	query := `
		INSERT INTO audit_export_jobs (id, organization_id, requested_by, status, filter, format, compress, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = r.db.ExecContext(ctx, query,
		job.ID, job.OrganizationID, job.RequestedBy, job.Status, job.Filter, job.Format, job.Compress, job.CreatedAt)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (r *AuditRepository) GetExportJob(ctx context.Context, jobID uuid.UUID) (*AuditExportJob, error) {
	var job AuditExportJob
	err := r.db.GetContext(ctx, &job, "SELECT * FROM audit_export_jobs WHERE id = $1", jobID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &job, err
}

func (r *AuditRepository) UpdateExportJobStatus(ctx context.Context, jobID uuid.UUID, status string, filePath *string, fileSize *int64, recordCount *int, errMsg *string) error {
	now := time.Now()
	var startedAt, completedAt, expiresAt *time.Time

	if status == "processing" {
		startedAt = &now
	}
	if status == "completed" || status == "failed" {
		completedAt = &now
		if status == "completed" {
			expires := now.Add(24 * time.Hour)
			expiresAt = &expires
		}
	}

	query := `
		UPDATE audit_export_jobs SET
			status = $2,
			file_path = COALESCE($3, file_path),
			file_size = COALESCE($4, file_size),
			record_count = COALESCE($5, record_count),
			error = $6,
			started_at = COALESCE($7, started_at),
			completed_at = COALESCE($8, completed_at),
			expires_at = COALESCE($9, expires_at)
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, jobID, status, filePath, fileSize, recordCount, errMsg, startedAt, completedAt, expiresAt)
	return err
}

func (r *AuditRepository) ListExportJobs(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]AuditExportJob, error) {
	var jobs []AuditExportJob
	err := r.db.SelectContext(ctx, &jobs, `
		SELECT * FROM audit_export_jobs
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		orgID, limit, offset)
	return jobs, err
}
