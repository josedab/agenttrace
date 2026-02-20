package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// PostgresContainer wraps a testcontainers PostgreSQL instance for integration tests.
type PostgresContainer struct {
	Container testcontainers.Container
	DB        *database.PostgresDB
	DSN       string
}

// SetupPostgres creates a PostgreSQL container, runs migrations, and returns a ready-to-use DB.
// Automatically cleaned up when the test finishes.
func SetupPostgres(t *testing.T) *PostgresContainer {
	t.Helper()

	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test: set INTEGRATION_TEST=1 to run")
	}

	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("test_agenttrace"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get postgres host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get postgres port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/test_agenttrace?sslmode=disable", host, port.Port())

	cfg := config.PostgresConfig{
		Host:     host,
		Port:     port.Int(),
		User:     "test",
		Password: "test",
		Database: "test_agenttrace",
		SSLMode:  "disable",
		MaxConns: 5,
		MinConns: 1,
	}

	db, err := database.NewPostgres(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}

	t.Cleanup(func() {
		db.Pool.Close()
	})

	// Run migrations
	if err := runPostgresMigrations(ctx, db.Pool); err != nil {
		t.Fatalf("failed to run postgres migrations: %v", err)
	}

	return &PostgresContainer{
		Container: container,
		DB:        db,
		DSN:       dsn,
	}
}

// runPostgresMigrations applies all up migrations in order.
func runPostgresMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrationsDir := findMigrationsDir("postgres")
	if migrationsDir == "" {
		return fmt.Errorf("could not find postgres migrations directory")
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	var upFiles []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			if matched, _ := filepath.Match("*.up.sql", e.Name()); matched {
				upFiles = append(upFiles, e.Name())
			}
		}
	}
	sort.Strings(upFiles)

	for _, f := range upFiles {
		sql, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", f, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", f, err)
		}
	}

	return nil
}

// findMigrationsDir walks up from the current directory to find the migrations folder.
func findMigrationsDir(dbType string) string {
	dir, _ := os.Getwd()
	for i := 0; i < 10; i++ {
		candidate := filepath.Join(dir, "migrations", dbType)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
