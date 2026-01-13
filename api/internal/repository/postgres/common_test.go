package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// getTestDB returns a database connection for integration tests.
// Returns nil if the database is not available (skips tests).
func getTestDB(t *testing.T) *database.PostgresDB {
	// Check if we're running integration tests
	if os.Getenv("POSTGRES_TEST_HOST") == "" {
		t.Skip("Skipping integration test: POSTGRES_TEST_HOST not set")
		return nil
	}

	cfg := config.PostgresConfig{
		Host:     os.Getenv("POSTGRES_TEST_HOST"),
		Port:     5432,
		User:     os.Getenv("POSTGRES_TEST_USER"),
		Password: os.Getenv("POSTGRES_TEST_PASS"),
		Database: os.Getenv("POSTGRES_TEST_DB"),
		SSLMode:  "disable",
		MaxConns: 5,
		MinConns: 1,
	}

	if cfg.Database == "" {
		cfg.Database = "test_agenttrace"
	}
	if cfg.User == "" {
		cfg.User = "postgres"
	}

	db, err := database.NewPostgres(context.Background(), cfg)
	if err != nil {
		t.Skipf("Skipping integration test: failed to connect to PostgreSQL: %v", err)
		return nil
	}

	return db
}

// cleanupUsers removes test users from the database
func cleanupUsers(t *testing.T, db *database.PostgresDB, emails ...string) {
	ctx := context.Background()
	for _, email := range emails {
		_, _ = db.Pool.Exec(ctx, "DELETE FROM users WHERE email = $1", email)
	}
}

// cleanupOrgs removes test organizations from the database
func cleanupOrgs(t *testing.T, db *database.PostgresDB, names ...string) {
	ctx := context.Background()
	for _, name := range names {
		_, _ = db.Pool.Exec(ctx, "DELETE FROM organizations WHERE name = $1", name)
	}
}

// cleanupProjects removes test projects from the database
func cleanupProjects(t *testing.T, db *database.PostgresDB, names ...string) {
	ctx := context.Background()
	for _, name := range names {
		_, _ = db.Pool.Exec(ctx, "DELETE FROM projects WHERE name = $1", name)
	}
}
