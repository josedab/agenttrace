package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// DiffAnalysisRepository handles persistence for diff analyses
type DiffAnalysisRepository struct {
	db *database.PostgresDB
}

// NewDiffAnalysisRepository creates a new repository
func NewDiffAnalysisRepository(db *database.PostgresDB) *DiffAnalysisRepository {
	return &DiffAnalysisRepository{db: db}
}

// Save saves a diff analysis
func (r *DiffAnalysisRepository) Save(ctx context.Context, analysis *domain.DiffAnalysis) error {
	findingsJSON, err := json.Marshal(analysis.Findings)
	if err != nil {
		return fmt.Errorf("failed to marshal findings: %w", err)
	}

	fileAnalysesJSON, err := json.Marshal(analysis.FileAnalyses)
	if err != nil {
		return fmt.Errorf("failed to marshal file analyses: %w", err)
	}

	dimensionScoresJSON, err := json.Marshal(analysis.DimensionScores)
	if err != nil {
		return fmt.Errorf("failed to marshal dimension scores: %w", err)
	}

	query := `
		INSERT INTO diff_analyses (
			id, project_id, trace_id, status,
			files_added, files_modified, files_deleted, lines_added, lines_removed,
			overall_score, dimension_scores, findings, file_analyses,
			agent_name, git_commit_sha, git_branch, created_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			overall_score = EXCLUDED.overall_score,
			dimension_scores = EXCLUDED.dimension_scores,
			findings = EXCLUDED.findings,
			file_analyses = EXCLUDED.file_analyses,
			completed_at = EXCLUDED.completed_at
	`

	_, err = r.db.Pool.Exec(ctx, query,
		analysis.ID, analysis.ProjectID, analysis.TraceID, analysis.Status,
		analysis.FilesAdded, analysis.FilesModified, analysis.FilesDeleted,
		analysis.LinesAdded, analysis.LinesRemoved,
		analysis.OverallScore, dimensionScoresJSON, findingsJSON, fileAnalysesJSON,
		analysis.AgentName, analysis.GitCommitSha, analysis.GitBranch,
		analysis.CreatedAt, analysis.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save diff analysis: %w", err)
	}

	return nil
}

// GetByID retrieves a diff analysis by ID
func (r *DiffAnalysisRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.DiffAnalysis, error) {
	query := `
		SELECT id, project_id, trace_id, status,
		       files_added, files_modified, files_deleted, lines_added, lines_removed,
		       overall_score, dimension_scores, findings, file_analyses,
		       agent_name, git_commit_sha, git_branch, created_at, completed_at
		FROM diff_analyses WHERE id = $1
	`

	var analysis domain.DiffAnalysis
	var findingsJSON, fileAnalysesJSON, dimensionScoresJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&analysis.ID, &analysis.ProjectID, &analysis.TraceID, &analysis.Status,
		&analysis.FilesAdded, &analysis.FilesModified, &analysis.FilesDeleted,
		&analysis.LinesAdded, &analysis.LinesRemoved,
		&analysis.OverallScore, &dimensionScoresJSON, &findingsJSON, &fileAnalysesJSON,
		&analysis.AgentName, &analysis.GitCommitSha, &analysis.GitBranch,
		&analysis.CreatedAt, &analysis.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff analysis: %w", err)
	}

	json.Unmarshal(findingsJSON, &analysis.Findings)
	json.Unmarshal(fileAnalysesJSON, &analysis.FileAnalyses)
	json.Unmarshal(dimensionScoresJSON, &analysis.DimensionScores)

	return &analysis, nil
}

// GetByTraceID retrieves analyses for a trace
func (r *DiffAnalysisRepository) GetByTraceID(ctx context.Context, traceID uuid.UUID) ([]domain.DiffAnalysisSummary, error) {
	query := `
		SELECT id, trace_id, status, overall_score,
		       COALESCE(jsonb_array_length(findings::jsonb), 0) as finding_count,
		       files_added + files_modified + files_deleted as files_changed,
		       created_at
		FROM diff_analyses WHERE trace_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list diff analyses: %w", err)
	}
	defer rows.Close()

	var summaries []domain.DiffAnalysisSummary
	for rows.Next() {
		var s domain.DiffAnalysisSummary
		if err := rows.Scan(&s.ID, &s.TraceID, &s.Status, &s.OverallScore, &s.FindingCount, &s.FilesChanged, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan diff analysis summary: %w", err)
		}
		summaries = append(summaries, s)
	}

	return summaries, nil
}

// List retrieves analyses with filtering
func (r *DiffAnalysisRepository) List(ctx context.Context, filter *domain.DiffAnalysisFilter, limit, offset int) ([]domain.DiffAnalysisSummary, int64, error) {
	countQuery := `SELECT COUNT(*) FROM diff_analyses WHERE project_id = $1`
	var totalCount int64
	r.db.Pool.QueryRow(ctx, countQuery, filter.ProjectID).Scan(&totalCount)

	query := `
		SELECT id, trace_id, status, overall_score,
		       COALESCE(jsonb_array_length(findings::jsonb), 0) as finding_count,
		       files_added + files_modified + files_deleted as files_changed,
		       created_at
		FROM diff_analyses WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, filter.ProjectID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list diff analyses: %w", err)
	}
	defer rows.Close()

	var summaries []domain.DiffAnalysisSummary
	for rows.Next() {
		var s domain.DiffAnalysisSummary
		if err := rows.Scan(&s.ID, &s.TraceID, &s.Status, &s.OverallScore, &s.FindingCount, &s.FilesChanged, &s.CreatedAt); err != nil {
			return nil, 0, err
		}
		summaries = append(summaries, s)
	}

	return summaries, totalCount, nil
}

// GetQualityTrend retrieves quality scores over time
func (r *DiffAnalysisRepository) GetQualityTrend(ctx context.Context, projectID uuid.UUID, since time.Time) (*domain.QualityTrend, error) {
	query := `
		SELECT overall_score, trace_id, created_at
		FROM diff_analyses
		WHERE project_id = $1 AND created_at >= $2 AND status = 'completed'
		ORDER BY created_at ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get quality trend: %w", err)
	}
	defer rows.Close()

	trend := &domain.QualityTrend{}
	var sum float64
	for rows.Next() {
		var p domain.QualityTrendPoint
		if err := rows.Scan(&p.OverallScore, &p.TraceID, &p.Timestamp); err != nil {
			return nil, err
		}
		trend.Points = append(trend.Points, p)
		sum += p.OverallScore
	}

	if len(trend.Points) > 0 {
		trend.Average = sum / float64(len(trend.Points))
		if len(trend.Points) >= 2 {
			first := trend.Points[0].OverallScore
			last := trend.Points[len(trend.Points)-1].OverallScore
			if last > first+5 {
				trend.Trend = "improving"
			} else if last < first-5 {
				trend.Trend = "declining"
			} else {
				trend.Trend = "stable"
			}
		} else {
			trend.Trend = "stable"
		}
	}

	return trend, nil
}
