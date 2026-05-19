package migration

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	// Register migration database and source drivers.
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"

	"github.com/agenttrace/agenttrace/api/internal/config"
)

const migrationLockID int64 = 6847281401

// Target selects which database migrations to run.
type Target string

const (
	// TargetAll runs PostgreSQL and ClickHouse migrations.
	TargetAll Target = "all"
	// TargetPostgres runs only PostgreSQL migrations.
	TargetPostgres Target = "postgres"
	// TargetClickHouse runs only ClickHouse migrations.
	TargetClickHouse Target = "clickhouse"
)

// Runner executes application database migrations.
type Runner struct {
	Config      *config.Config
	Path        string
	WaitTimeout time.Duration
}

// Up applies outstanding migrations.
func (r Runner) Up(ctx context.Context, target Target) error {
	switch target {
	case TargetAll:
		return r.withPostgresLock(ctx, func() error {
			if err := r.migratePostgres(true); err != nil {
				return err
			}
			return r.migrateClickHouse(ctx, true)
		})
	case TargetPostgres:
		return r.withPostgresLock(ctx, func() error {
			return r.migratePostgres(true)
		})
	case TargetClickHouse:
		return r.migrateClickHouse(ctx, true)
	default:
		return fmt.Errorf("unsupported migration target %q", target)
	}
}

// Down rolls each selected database back by one migration.
func (r Runner) Down(ctx context.Context, target Target) error {
	switch target {
	case TargetAll:
		return errors.New(
			"combined rollback is not supported; roll back PostgreSQL and ClickHouse explicitly",
		)
	case TargetPostgres:
		return r.withPostgresLock(ctx, func() error {
			return r.migratePostgres(false)
		})
	case TargetClickHouse:
		return r.migrateClickHouse(ctx, false)
	default:
		return fmt.Errorf("unsupported migration target %q", target)
	}
}

func (r Runner) withPostgresLock(ctx context.Context, fn func() error) error {
	timeout := r.WaitTimeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	deadline := time.Now().Add(timeout)

	var conn *pgx.Conn
	var err error
	for {
		conn, err = pgx.Connect(ctx, r.Config.Postgres.DSN())
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("connect to PostgreSQL for migration lock: %w", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationLockID); err != nil {
		closeErr := conn.Close(context.Background())
		return errors.Join(fmt.Errorf("acquire migration lock: %w", err), closeErr)
	}

	runErr := fn()
	_, unlockErr := conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", migrationLockID)
	closeErr := conn.Close(context.Background())
	return errors.Join(runErr, unlockErr, closeErr)
}

func (r Runner) migratePostgres(up bool) error {
	migrations, err := migrate.New(
		fileSource(filepath.Join(r.Path, "postgres")),
		r.Config.Postgres.DSN(),
	)
	if err != nil {
		return fmt.Errorf("initialize PostgreSQL migrations: %w", err)
	}
	if up {
		err = migrations.Up()
	} else {
		err = migrations.Steps(-1)
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		err = fmt.Errorf("run PostgreSQL migrations: %w", err)
	} else {
		err = nil
	}
	return errors.Join(err, closeMigrations(migrations))
}

func (r Runner) migrateClickHouse(ctx context.Context, up bool) error {
	databaseURL, err := NormalizeClickHouseURL(r.Config.ClickHouse.MigrationURL())
	if err != nil {
		return fmt.Errorf("normalize ClickHouse migration URL: %w", err)
	}

	if up {
		state, err := r.waitForClickHouse(ctx, databaseURL)
		if err != nil {
			return err
		}

		migrations, err := migrate.New(
			fileSource(filepath.Join(r.Path, "clickhouse")),
			databaseURL,
		)
		if err != nil {
			return fmt.Errorf("initialize ClickHouse migrations: %w", err)
		}
		if state == ClickHouseSchemaLegacy {
			if err := migrations.Force(9); err != nil {
				return errors.Join(
					fmt.Errorf("adopt legacy ClickHouse schema: %w", err),
					closeMigrations(migrations),
				)
			}
		}
		runErr := migrations.Up()
		if runErr != nil && !errors.Is(runErr, migrate.ErrNoChange) {
			runErr = fmt.Errorf("run ClickHouse migrations: %w", runErr)
		} else {
			runErr = nil
		}
		return errors.Join(runErr, closeMigrations(migrations))
	}

	migrations, err := migrate.New(
		fileSource(filepath.Join(r.Path, "clickhouse")),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("initialize ClickHouse migrations: %w", err)
	}
	runErr := migrations.Steps(-1)
	if runErr != nil && !errors.Is(runErr, migrate.ErrNoChange) {
		runErr = fmt.Errorf("roll back ClickHouse migration: %w", runErr)
	} else {
		runErr = nil
	}
	return errors.Join(runErr, closeMigrations(migrations))
}

func (r Runner) waitForClickHouse(ctx context.Context, databaseURL string) (ClickHouseSchemaState, error) {
	timeout := r.WaitTimeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	deadline := time.Now().Add(timeout)

	for {
		state, err := InspectClickHouseSchema(ctx, databaseURL)
		if err == nil {
			return state, nil
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("connect to ClickHouse for migrations: %w", err)
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

func fileSource(path string) string {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		absolutePath = path
	}
	return (&url.URL{Scheme: "file", Path: absolutePath}).String()
}

func closeMigrations(migrations *migrate.Migrate) error {
	sourceErr, databaseErr := migrations.Close()
	return errors.Join(sourceErr, databaseErr)
}
