package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// EvalHubRepository persists packages, immutable versions, and executions.
type EvalHubRepository struct {
	db *database.PostgresDB
}

// pgUniqueViolation is the PostgreSQL SQLSTATE for a unique constraint failure.
const pgUniqueViolation = "23505"

const evalHubAccessClause = `(
	owner_project_id = $1
	OR visibility = 'public'
	OR (visibility = 'organization' AND organization_id = $2)
)`

// NewEvalHubRepository creates an Eval Hub repository.
func NewEvalHubRepository(db *database.PostgresDB) *EvalHubRepository {
	return &EvalHubRepository{db: db}
}

// SavePackageVersion atomically creates or updates a package and stores an immutable version.
func (r *EvalHubRepository) SavePackageVersion(
	ctx context.Context,
	pkg *domain.EvalHubPackage,
	version *domain.EvalHubVersion,
	createPackage bool,
) error {
	return database.Transaction(ctx, r.db, func(tx pgx.Tx) error {
		if createPackage {
			_, err := tx.Exec(ctx, `
				INSERT INTO eval_hub_packages (
					id, owner_project_id, organization_id, kind, name, description,
					visibility, latest_version, forked_from_package_id,
					forked_from_version, published_by, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			`,
				pkg.ID,
				pkg.OwnerProjectID,
				pkg.OrganizationID,
				pkg.Kind,
				pkg.Name,
				pkg.Description,
				pkg.Visibility,
				pkg.LatestVersion,
				pkg.ForkedFromPackageID,
				pkg.ForkedFromVersion,
				pkg.PublishedBy,
				pkg.CreatedAt,
				pkg.UpdatedAt,
			)
			if err != nil {
				return fmt.Errorf("create eval hub package: %w", err)
			}
		} else {
			result, err := tx.Exec(ctx, `
				UPDATE eval_hub_packages
				SET name = $3, description = $4, visibility = $5,
					latest_version = $6, updated_at = $7
				WHERE owner_project_id = $1 AND id = $2
			`,
				pkg.OwnerProjectID,
				pkg.ID,
				pkg.Name,
				pkg.Description,
				pkg.Visibility,
				pkg.LatestVersion,
				pkg.UpdatedAt,
			)
			if err != nil {
				return fmt.Errorf("update eval hub package: %w", err)
			}
			if result.RowsAffected() == 0 {
				return apperrors.NotFound("eval hub package")
			}
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO eval_hub_versions (
				id, package_id, version, source_resource_id, manifest, checksum,
				version_note, created_by, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
			version.ID,
			version.PackageID,
			version.Version,
			version.SourceResourceID,
			version.Manifest,
			version.Checksum,
			version.VersionNote,
			version.CreatedBy,
			version.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("create eval hub package version: %w", err)
		}
		return nil
	})
}

// GetOwnedPackage retrieves a package only from its owner project.
func (r *EvalHubRepository) GetOwnedPackage(
	ctx context.Context,
	projectID, packageID uuid.UUID,
) (*domain.EvalHubPackage, error) {
	return r.getPackage(ctx, `
		WHERE package_id = $1 AND owner_project_id = $2
	`, packageID, projectID)
}

// GetAccessiblePackage enforces public, organization, and owner visibility boundaries.
func (r *EvalHubRepository) GetAccessiblePackage(
	ctx context.Context,
	packageID, requesterProjectID, organizationID uuid.UUID,
) (*domain.EvalHubPackage, error) {
	return r.getPackage(ctx, `
		WHERE package_id = $1
			AND (
				owner_project_id = $2
				OR visibility = 'public'
				OR (visibility = 'organization' AND organization_id = $3)
			)
	`, packageID, requesterProjectID, organizationID)
}

func (r *EvalHubRepository) getPackage(
	ctx context.Context,
	where string,
	args ...interface{},
) (*domain.EvalHubPackage, error) {
	query := `
		SELECT
			package_id, owner_project_id, organization_id, kind, name, description,
			visibility, latest_version, forked_from_package_id, forked_from_version,
			published_by, package_created_at, package_updated_at,
			version_id, source_resource_id, manifest, checksum, version_note,
			version_created_by, version_created_at
		FROM eval_hub_package_latest
	` + where

	var pkg domain.EvalHubPackage
	var version domain.EvalHubVersion
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(
		&pkg.ID,
		&pkg.OwnerProjectID,
		&pkg.OrganizationID,
		&pkg.Kind,
		&pkg.Name,
		&pkg.Description,
		&pkg.Visibility,
		&pkg.LatestVersion,
		&pkg.ForkedFromPackageID,
		&pkg.ForkedFromVersion,
		&pkg.PublishedBy,
		&pkg.CreatedAt,
		&pkg.UpdatedAt,
		&version.ID,
		&version.SourceResourceID,
		&version.Manifest,
		&version.Checksum,
		&version.VersionNote,
		&version.CreatedBy,
		&version.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("eval hub package")
		}
		return nil, fmt.Errorf("get eval hub package: %w", err)
	}
	version.PackageID = pkg.ID
	version.Version = pkg.LatestVersion
	pkg.Version = &version
	return &pkg, nil
}

// GetVersion retrieves an immutable package version.
func (r *EvalHubRepository) GetVersion(
	ctx context.Context,
	packageID uuid.UUID,
	versionNumber int,
) (*domain.EvalHubVersion, error) {
	var version domain.EvalHubVersion
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, package_id, version, source_resource_id, manifest, checksum,
			version_note, created_by, created_at
		FROM eval_hub_versions
		WHERE package_id = $1 AND version = $2
	`, packageID, versionNumber).Scan(
		&version.ID,
		&version.PackageID,
		&version.Version,
		&version.SourceResourceID,
		&version.Manifest,
		&version.Checksum,
		&version.VersionNote,
		&version.CreatedBy,
		&version.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("eval hub package version")
		}
		return nil, fmt.Errorf("get eval hub package version: %w", err)
	}
	return &version, nil
}

// ListAccessiblePackages lists only packages visible to the requester.
func (r *EvalHubRepository) ListAccessiblePackages(
	ctx context.Context,
	filter domain.EvalHubPackageFilter,
) (*domain.EvalHubPackageList, error) {
	conditions := []string{evalHubAccessClause}
	args := []interface{}{filter.RequesterProjectID, filter.OrganizationID}
	nextArg := 3

	if filter.Kind != nil {
		conditions = append(conditions, fmt.Sprintf("kind = $%d", nextArg))
		args = append(args, *filter.Kind)
		nextArg++
	}
	if filter.Visibility != nil {
		conditions = append(conditions, fmt.Sprintf("visibility = $%d", nextArg))
		args = append(args, *filter.Visibility)
		nextArg++
	}
	if filter.Query != "" {
		conditions = append(
			conditions,
			fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", nextArg, nextArg),
		)
		args = append(args, "%"+filter.Query+"%")
		nextArg++
	}
	where := strings.Join(conditions, " AND ")

	var totalCount int64
	if err := r.db.Pool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM eval_hub_packages WHERE "+where,
		args...,
	).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("count eval hub packages: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT
			package_id, owner_project_id, organization_id, kind, name, description,
			visibility, latest_version, forked_from_package_id, forked_from_version,
			published_by, package_created_at, package_updated_at,
			version_id, source_resource_id, manifest, checksum, version_note,
			version_created_by, version_created_at
		FROM eval_hub_package_latest
		WHERE %s
		ORDER BY package_updated_at DESC
		LIMIT $%d OFFSET $%d
	`, where, nextArg, nextArg+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list eval hub packages: %w", err)
	}
	defer rows.Close()

	packages := make([]domain.EvalHubPackage, 0)
	for rows.Next() {
		var pkg domain.EvalHubPackage
		var version domain.EvalHubVersion
		if err := rows.Scan(
			&pkg.ID,
			&pkg.OwnerProjectID,
			&pkg.OrganizationID,
			&pkg.Kind,
			&pkg.Name,
			&pkg.Description,
			&pkg.Visibility,
			&pkg.LatestVersion,
			&pkg.ForkedFromPackageID,
			&pkg.ForkedFromVersion,
			&pkg.PublishedBy,
			&pkg.CreatedAt,
			&pkg.UpdatedAt,
			&version.ID,
			&version.SourceResourceID,
			&version.Manifest,
			&version.Checksum,
			&version.VersionNote,
			&version.CreatedBy,
			&version.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan eval hub package: %w", err)
		}
		version.PackageID = pkg.ID
		version.Version = pkg.LatestVersion
		pkg.Version = &version
		packages = append(packages, pkg)
	}

	return &domain.EvalHubPackageList{
		Packages:   packages,
		TotalCount: totalCount,
		HasMore:    int64(filter.Offset+len(packages)) < totalCount,
	}, nil
}

// CreateRun persists an Eval Hub execution.
func (r *EvalHubRepository) CreateRun(ctx context.Context, run *domain.EvalHubRun) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO eval_hub_runs (
			id, project_id, package_id, package_version, status, dataset_run_id,
			experiment_id, result, capability_message, idempotency_key,
			created_by, started_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''), $11, $12, $13)
	`,
		run.ID,
		run.ProjectID,
		run.PackageID,
		run.PackageVersion,
		run.Status,
		run.DatasetRunID,
		run.ExperimentID,
		run.Result,
		run.CapabilityMessage,
		run.IdempotencyKey,
		run.CreatedBy,
		run.StartedAt,
		run.CompletedAt,
	)
	if err != nil {
		// A duplicate idempotency key means a concurrent request already created
		// the run; callers re-read it instead of starting a second execution.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return apperrors.Conflict("eval hub run already exists for this idempotency key")
		}
		return fmt.Errorf("create eval hub run: %w", err)
	}
	return nil
}

// UpdateRun persists the outcome of a durable run within its project.
func (r *EvalHubRepository) UpdateRun(ctx context.Context, run *domain.EvalHubRun) error {
	commandTag, err := r.db.Pool.Exec(ctx, `
		UPDATE eval_hub_runs
		SET status = $3, dataset_run_id = $4, experiment_id = $5, result = $6,
			capability_message = $7, completed_at = $8
		WHERE project_id = $1 AND id = $2
	`,
		run.ProjectID,
		run.ID,
		run.Status,
		run.DatasetRunID,
		run.ExperimentID,
		run.Result,
		run.CapabilityMessage,
		run.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("update eval hub run: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return apperrors.NotFound("eval hub run")
	}
	return nil
}

// ListRuns lists executions within one project.
func (r *EvalHubRepository) ListRuns(
	ctx context.Context,
	projectID uuid.UUID,
	limit, offset int,
) (*domain.EvalHubRunList, error) {
	var totalCount int64
	if err := r.db.Pool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM eval_hub_runs WHERE project_id = $1",
		projectID,
	).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("count eval hub runs: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, project_id, package_id, package_version, status, dataset_run_id,
			experiment_id, result, capability_message, COALESCE(idempotency_key, ''),
			created_by, started_at, completed_at
		FROM eval_hub_runs
		WHERE project_id = $1
		ORDER BY started_at DESC
		LIMIT $2 OFFSET $3
	`, projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list eval hub runs: %w", err)
	}
	defer rows.Close()

	runs := make([]domain.EvalHubRun, 0)
	for rows.Next() {
		var run domain.EvalHubRun
		if err := rows.Scan(
			&run.ID,
			&run.ProjectID,
			&run.PackageID,
			&run.PackageVersion,
			&run.Status,
			&run.DatasetRunID,
			&run.ExperimentID,
			&run.Result,
			&run.CapabilityMessage,
			&run.IdempotencyKey,
			&run.CreatedBy,
			&run.StartedAt,
			&run.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan eval hub run: %w", err)
		}
		runs = append(runs, run)
	}
	return &domain.EvalHubRunList{
		Runs:       runs,
		TotalCount: totalCount,
		HasMore:    int64(offset+len(runs)) < totalCount,
	}, nil
}

// GetRunByID retrieves a run within its project.
func (r *EvalHubRepository) GetRunByID(
	ctx context.Context,
	projectID, runID uuid.UUID,
) (*domain.EvalHubRun, error) {
	return r.getRun(ctx, "project_id = $1 AND id = $2", projectID, runID)
}

// GetRunByIdempotencyKey returns an existing idempotent run.
func (r *EvalHubRepository) GetRunByIdempotencyKey(
	ctx context.Context,
	projectID uuid.UUID,
	key string,
) (*domain.EvalHubRun, error) {
	return r.getRun(ctx, "project_id = $1 AND idempotency_key = $2", projectID, key)
}

func (r *EvalHubRepository) getRun(
	ctx context.Context,
	where string,
	args ...interface{},
) (*domain.EvalHubRun, error) {
	var run domain.EvalHubRun
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, project_id, package_id, package_version, status, dataset_run_id,
			experiment_id, result, capability_message, COALESCE(idempotency_key, ''),
			created_by, started_at, completed_at
		FROM eval_hub_runs
		WHERE `+where,
		args...,
	).Scan(
		&run.ID,
		&run.ProjectID,
		&run.PackageID,
		&run.PackageVersion,
		&run.Status,
		&run.DatasetRunID,
		&run.ExperimentID,
		&run.Result,
		&run.CapabilityMessage,
		&run.IdempotencyKey,
		&run.CreatedBy,
		&run.StartedAt,
		&run.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("eval hub run")
		}
		return nil, fmt.Errorf("get eval hub run: %w", err)
	}
	return &run, nil
}
