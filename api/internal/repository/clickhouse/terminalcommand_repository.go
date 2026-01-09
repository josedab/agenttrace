package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// TerminalCommandRepository handles terminal command data in ClickHouse
type TerminalCommandRepository struct {
	db *database.ClickHouseDB
}

// NewTerminalCommandRepository creates a new terminal command repository
func NewTerminalCommandRepository(db *database.ClickHouseDB) *TerminalCommandRepository {
	return &TerminalCommandRepository{db: db}
}

// Create inserts a new terminal command
func (r *TerminalCommandRepository) Create(ctx context.Context, cmd *domain.TerminalCommand) error {
	query := `
		INSERT INTO terminal_commands (
			id, project_id, trace_id, observation_id, command, args,
			working_directory, shell, env_vars, started_at, completed_at,
			duration_ms, exit_code, stdout, stderr, stdout_truncated,
			stderr_truncated, success, timed_out, killed, max_memory_bytes,
			cpu_time_ms, tool_name, reason
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		cmd.ID,
		cmd.ProjectID,
		cmd.TraceID,
		cmd.ObservationID,
		cmd.Command,
		cmd.Args,
		cmd.WorkingDirectory,
		cmd.Shell,
		cmd.EnvVars,
		cmd.StartedAt,
		cmd.CompletedAt,
		cmd.DurationMs,
		cmd.ExitCode,
		cmd.Stdout,
		cmd.Stderr,
		cmd.StdoutTruncated,
		cmd.StderrTruncated,
		cmd.Success,
		cmd.TimedOut,
		cmd.Killed,
		cmd.MaxMemoryBytes,
		cmd.CPUTimeMs,
		cmd.ToolName,
		cmd.Reason,
	)
}

// CreateBatch inserts multiple terminal commands
func (r *TerminalCommandRepository) CreateBatch(ctx context.Context, cmds []*domain.TerminalCommand) error {
	if len(cmds) == 0 {
		return nil
	}

	batch, err := r.db.PrepareBatch(ctx, `
		INSERT INTO terminal_commands (
			id, project_id, trace_id, observation_id, command, args,
			working_directory, shell, env_vars, started_at, completed_at,
			duration_ms, exit_code, stdout, stderr, stdout_truncated,
			stderr_truncated, success, timed_out, killed, max_memory_bytes,
			cpu_time_ms, tool_name, reason
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, cmd := range cmds {
		if err := batch.Append(
			cmd.ID,
			cmd.ProjectID,
			cmd.TraceID,
			cmd.ObservationID,
			cmd.Command,
			cmd.Args,
			cmd.WorkingDirectory,
			cmd.Shell,
			cmd.EnvVars,
			cmd.StartedAt,
			cmd.CompletedAt,
			cmd.DurationMs,
			cmd.ExitCode,
			cmd.Stdout,
			cmd.Stderr,
			cmd.StdoutTruncated,
			cmd.StderrTruncated,
			cmd.Success,
			cmd.TimedOut,
			cmd.Killed,
			cmd.MaxMemoryBytes,
			cmd.CPUTimeMs,
			cmd.ToolName,
			cmd.Reason,
		); err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	return batch.Send()
}

// GetByTraceID retrieves all terminal commands for a trace
func (r *TerminalCommandRepository) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.TerminalCommand, error) {
	query := `
		SELECT
			id, project_id, trace_id, observation_id, command, args,
			working_directory, shell, env_vars, started_at, completed_at,
			duration_ms, exit_code, stdout, stderr, stdout_truncated,
			stderr_truncated, success, timed_out, killed, max_memory_bytes,
			cpu_time_ms, tool_name, reason
		FROM terminal_commands FINAL
		WHERE project_id = ? AND trace_id = ?
		ORDER BY started_at ASC
	`

	var cmds []domain.TerminalCommand
	if err := r.db.Select(ctx, &cmds, query, projectID, traceID); err != nil {
		return nil, err
	}

	return cmds, nil
}

// List retrieves terminal commands with filtering
func (r *TerminalCommandRepository) List(ctx context.Context, filter *domain.TerminalCommandFilter, limit, offset int) (*domain.TerminalCommandList, error) {
	conditions := []string{"project_id = ?"}
	args := []interface{}{filter.ProjectID}

	if filter.TraceID != nil {
		conditions = append(conditions, "trace_id = ?")
		args = append(args, *filter.TraceID)
	}

	if filter.ObservationID != nil {
		conditions = append(conditions, "observation_id = ?")
		args = append(args, *filter.ObservationID)
	}

	if filter.Command != nil {
		conditions = append(conditions, "command LIKE ?")
		args = append(args, "%"+*filter.Command+"%")
	}

	if filter.ExitCode != nil {
		conditions = append(conditions, "exit_code = ?")
		args = append(args, *filter.ExitCode)
	}

	if filter.Success != nil {
		conditions = append(conditions, "success = ?")
		args = append(args, *filter.Success)
	}

	if filter.FromTime != nil {
		conditions = append(conditions, "started_at >= ?")
		args = append(args, *filter.FromTime)
	}

	if filter.ToTime != nil {
		conditions = append(conditions, "started_at <= ?")
		args = append(args, *filter.ToTime)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Get count
	countQuery := fmt.Sprintf("SELECT count() FROM terminal_commands FINAL WHERE %s", whereClause)
	var totalCount int64
	row := r.db.QueryRow(ctx, countQuery, args...)
	if err := row.Scan(&totalCount); err != nil {
		return nil, err
	}

	// Get terminal commands
	query := fmt.Sprintf(`
		SELECT
			id, project_id, trace_id, observation_id, command, args,
			working_directory, shell, env_vars, started_at, completed_at,
			duration_ms, exit_code, stdout, stderr, stdout_truncated,
			stderr_truncated, success, timed_out, killed, max_memory_bytes,
			cpu_time_ms, tool_name, reason
		FROM terminal_commands FINAL
		WHERE %s
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)

	var cmds []domain.TerminalCommand
	if err := r.db.Select(ctx, &cmds, query, args...); err != nil {
		return nil, err
	}

	return &domain.TerminalCommandList{
		TerminalCommands: cmds,
		TotalCount:       totalCount,
		HasMore:          int64(offset+len(cmds)) < totalCount,
	}, nil
}

// GetStats retrieves terminal command statistics
func (r *TerminalCommandRepository) GetStats(ctx context.Context, projectID uuid.UUID, traceID *string) (*domain.TerminalCommandStats, error) {
	conditions := []string{"project_id = ?"}
	args := []interface{}{projectID}

	if traceID != nil {
		conditions = append(conditions, "trace_id = ?")
		args = append(args, *traceID)
	}

	whereClause := strings.Join(conditions, " AND ")

	query := fmt.Sprintf(`
		SELECT
			count() AS total_commands,
			countIf(success = true) AS success_count,
			countIf(success = false) AS failure_count,
			countIf(timed_out = true) AS timeout_count,
			avg(duration_ms) AS avg_duration_ms,
			sum(duration_ms) AS total_duration_ms
		FROM terminal_commands FINAL
		WHERE %s
	`, whereClause)

	var stats domain.TerminalCommandStats
	row := r.db.QueryRow(ctx, query, args...)
	err := row.Scan(
		&stats.TotalCommands,
		&stats.SuccessCount,
		&stats.FailureCount,
		&stats.TimeoutCount,
		&stats.AvgDurationMs,
		&stats.TotalDurationMs,
	)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
