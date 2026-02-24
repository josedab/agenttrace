package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// RCARepository handles persistence for root cause analysis reports
type RCARepository struct {
	db *database.PostgresDB
}

// NewRCARepository creates a new RCA repository
func NewRCARepository(db *database.PostgresDB) *RCARepository {
	return &RCARepository{db: db}
}

// Save persists an RCA report
func (r *RCARepository) Save(ctx context.Context, report *domain.RCAReport) error {
	factorsJSON, _ := json.Marshal(report.ContributingFactors)
	remediationsJSON, _ := json.Marshal(report.Remediations)
	query := `INSERT INTO rca_reports (id, project_id, trace_id, primary_category, confidence, summary, detailed_analysis, contributing_factors, remediations, analyzed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.Pool.Exec(ctx, query, report.ID, report.ProjectID, report.TraceID.String(),
		report.PrimaryCategory, report.Confidence, report.Summary, report.DetailedAnalysis,
		factorsJSON, remediationsJSON, report.AnalyzedAt)
	if err != nil {
		return fmt.Errorf("failed to save RCA report: %w", err)
	}
	return nil
}

// GetByID retrieves an RCA report by ID
func (r *RCARepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.RCAReport, error) {
	query := `SELECT id, project_id, trace_id, primary_category, confidence, summary, detailed_analysis, contributing_factors, remediations, analyzed_at FROM rca_reports WHERE id = $1`
	var report domain.RCAReport
	var factorsJSON, remediationsJSON []byte
	var traceIDStr string
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&report.ID, &report.ProjectID, &traceIDStr,
		&report.PrimaryCategory, &report.Confidence, &report.Summary, &report.DetailedAnalysis,
		&factorsJSON, &remediationsJSON, &report.AnalyzedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get RCA report: %w", err)
	}
	report.TraceID, _ = uuid.Parse(traceIDStr)
	json.Unmarshal(factorsJSON, &report.ContributingFactors)
	json.Unmarshal(remediationsJSON, &report.Remediations)
	return &report, nil
}

// ListByProject returns RCA reports for a project
func (r *RCARepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.RCAReport, error) {
	query := `SELECT id, project_id, trace_id, primary_category, confidence, summary, analyzed_at FROM rca_reports WHERE project_id = $1 ORDER BY analyzed_at DESC LIMIT 100`
	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list RCA reports: %w", err)
	}
	defer rows.Close()
	var reports []domain.RCAReport
	for rows.Next() {
		var rpt domain.RCAReport
		var traceIDStr string
		if err := rows.Scan(&rpt.ID, &rpt.ProjectID, &traceIDStr, &rpt.PrimaryCategory, &rpt.Confidence, &rpt.Summary, &rpt.AnalyzedAt); err != nil {
			return nil, fmt.Errorf("failed to scan RCA report: %w", err)
		}
		rpt.TraceID, _ = uuid.Parse(traceIDStr)
		reports = append(reports, rpt)
	}
	return reports, nil
}
