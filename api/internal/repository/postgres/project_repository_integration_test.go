//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/repository/postgres"
	"github.com/agenttrace/agenttrace/api/internal/testutil"
)

func TestProjectRepository_Integration(t *testing.T) {
	pg := testutil.SetupPostgres(t)
	repo := postgres.NewProjectRepository(pg.DB)
	ctx := context.Background()

	// Create an org first (projects require a foreign key)
	orgID := uuid.New()
	_, err := pg.DB.Pool.Exec(ctx, `
		INSERT INTO organizations (id, name, slug, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, orgID, "Test Org", "test-org", time.Now(), time.Now())
	require.NoError(t, err)

	t.Run("Create and Get", func(t *testing.T) {
		project := &domain.Project{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Name:           "Test Project",
			Slug:           "test-project-" + uuid.New().String()[:8],
			Description:    "A test project",
			RetentionDays:  90,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := repo.Create(ctx, project)
		require.NoError(t, err)

		got, err := repo.GetByID(ctx, project.ID)
		require.NoError(t, err)
		assert.Equal(t, project.Name, got.Name)
		assert.Equal(t, project.Slug, got.Slug)
		assert.Equal(t, project.Description, got.Description)
	})

	t.Run("ListByOrganizationID", func(t *testing.T) {
		projects, err := repo.ListByOrganizationID(ctx, orgID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(projects), 1)
	})

	t.Run("Update", func(t *testing.T) {
		project := &domain.Project{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Name:           "Update Test",
			Slug:           "update-test-" + uuid.New().String()[:8],
			RetentionDays:  30,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := repo.Create(ctx, project)
		require.NoError(t, err)

		project.Name = "Updated Name"
		project.UpdatedAt = time.Now()
		err = repo.Update(ctx, project)
		require.NoError(t, err)

		got, err := repo.GetByID(ctx, project.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", got.Name)
	})

	t.Run("Delete", func(t *testing.T) {
		project := &domain.Project{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Name:           "Delete Test",
			Slug:           "delete-test-" + uuid.New().String()[:8],
			RetentionDays:  30,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := repo.Create(ctx, project)
		require.NoError(t, err)

		err = repo.Delete(ctx, project.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, project.ID)
		assert.Error(t, err)
	})
}
