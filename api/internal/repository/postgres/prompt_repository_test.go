package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// createTestPrompt creates a prompt with test data
func createTestPrompt(name string, projectID uuid.UUID) *domain.Prompt {
	now := time.Now()
	return &domain.Prompt{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        name,
		Type:        "text",
		Description: "Test prompt description",
		Tags:        []string{"test", "prompt"},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// createTestPromptVersion creates a prompt version with test data
func createTestPromptVersion(promptID uuid.UUID, createdBy uuid.UUID) *domain.PromptVersion {
	now := time.Now()
	return &domain.PromptVersion{
		ID:            uuid.New(),
		PromptID:      promptID,
		Content:       "Hello {{name}}, welcome to {{place}}!",
		Config:        `{"model": "gpt-4"}`,
		Labels:        []string{"production"},
		CreatedBy:     &createdBy,
		CommitMessage: "Initial version",
		CreatedAt:     now,
	}
}

func cleanupPrompts(t *testing.T, db *database.PostgresDB, names ...string) {
	ctx := context.Background()
	for _, name := range names {
		_, _ = db.Pool.Exec(ctx, "DELETE FROM prompt_versions WHERE prompt_id IN (SELECT id FROM prompts WHERE name = $1)", name)
		_, _ = db.Pool.Exec(ctx, "DELETE FROM prompts WHERE name = $1", name)
	}
}

func TestPromptRepository_Create(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	promptRepo := NewPromptRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Prompt Create"
	projectName := "Test Project for Prompt Create"
	promptName := "Test Prompt Create"

	cleanupPrompts(t, db, promptName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupPrompts(t, db, promptName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	prompt := createTestPrompt(promptName, project.ID)
	err = promptRepo.Create(ctx, prompt)
	require.NoError(t, err)

	// Verify by fetching
	fetched, err := promptRepo.GetByID(ctx, prompt.ID)
	require.NoError(t, err)
	assert.Equal(t, prompt.ID, fetched.ID)
	assert.Equal(t, prompt.Name, fetched.Name)
	assert.Equal(t, prompt.ProjectID, fetched.ProjectID)
}

func TestPromptRepository_GetByID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	promptRepo := NewPromptRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Prompt GetByID"
	projectName := "Test Project for Prompt GetByID"
	promptName := "Test Prompt GetByID"

	cleanupPrompts(t, db, promptName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupPrompts(t, db, promptName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	prompt := createTestPrompt(promptName, project.ID)
	err = promptRepo.Create(ctx, prompt)
	require.NoError(t, err)

	t.Run("existing prompt", func(t *testing.T) {
		fetched, err := promptRepo.GetByID(ctx, prompt.ID)
		require.NoError(t, err)
		assert.Equal(t, prompt.ID, fetched.ID)
		assert.Equal(t, prompt.Name, fetched.Name)
	})

	t.Run("non-existent prompt", func(t *testing.T) {
		_, err := promptRepo.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestPromptRepository_GetByName(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	promptRepo := NewPromptRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Prompt GetByName"
	projectName := "Test Project for Prompt GetByName"
	promptName := "Test Prompt GetByName"

	cleanupPrompts(t, db, promptName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupPrompts(t, db, promptName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	prompt := createTestPrompt(promptName, project.ID)
	err = promptRepo.Create(ctx, prompt)
	require.NoError(t, err)

	t.Run("existing name", func(t *testing.T) {
		fetched, err := promptRepo.GetByName(ctx, project.ID, promptName)
		require.NoError(t, err)
		assert.Equal(t, prompt.ID, fetched.ID)
	})

	t.Run("non-existent name", func(t *testing.T) {
		_, err := promptRepo.GetByName(ctx, project.ID, "nonexistent")
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestPromptRepository_Update(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	promptRepo := NewPromptRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Prompt Update"
	projectName := "Test Project for Prompt Update"
	promptName := "Test Prompt Update"

	cleanupPrompts(t, db, promptName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupPrompts(t, db, promptName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	prompt := createTestPrompt(promptName, project.ID)
	err = promptRepo.Create(ctx, prompt)
	require.NoError(t, err)

	// Update
	prompt.Description = "Updated description"
	prompt.Tags = []string{"updated", "tags"}
	err = promptRepo.Update(ctx, prompt)
	require.NoError(t, err)

	// Verify
	fetched, err := promptRepo.GetByID(ctx, prompt.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", fetched.Description)
	assert.Equal(t, []string{"updated", "tags"}, fetched.Tags)
}

func TestPromptRepository_Delete(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	promptRepo := NewPromptRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Prompt Delete"
	projectName := "Test Project for Prompt Delete"
	promptName := "Test Prompt Delete"

	cleanupPrompts(t, db, promptName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	prompt := createTestPrompt(promptName, project.ID)
	err = promptRepo.Create(ctx, prompt)
	require.NoError(t, err)

	// Verify exists
	_, err = promptRepo.GetByID(ctx, prompt.ID)
	require.NoError(t, err)

	// Delete
	err = promptRepo.Delete(ctx, prompt.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = promptRepo.GetByID(ctx, prompt.ID)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestPromptRepository_List(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	promptRepo := NewPromptRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Prompt List"
	projectName := "Test Project for Prompt List"
	promptName1 := "Test Prompt List A"
	promptName2 := "Test Prompt List B"
	promptName3 := "Test Prompt List C"

	cleanupPrompts(t, db, promptName1, promptName2, promptName3)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupPrompts(t, db, promptName1, promptName2, promptName3)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	// Create prompts
	for _, name := range []string{promptName1, promptName2, promptName3} {
		prompt := createTestPrompt(name, project.ID)
		err = promptRepo.Create(ctx, prompt)
		require.NoError(t, err)
	}

	t.Run("basic list", func(t *testing.T) {
		filter := &domain.PromptFilter{ProjectID: project.ID}
		list, err := promptRepo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(3), list.TotalCount)
		assert.Len(t, list.Prompts, 3)
	})

	t.Run("with limit", func(t *testing.T) {
		filter := &domain.PromptFilter{ProjectID: project.ID}
		list, err := promptRepo.List(ctx, filter, 2, 0)
		require.NoError(t, err)
		assert.Len(t, list.Prompts, 2)
		assert.True(t, list.HasMore)
	})

	t.Run("filter by name", func(t *testing.T) {
		nameFilter := "List A"
		filter := &domain.PromptFilter{
			ProjectID: project.ID,
			Name:      &nameFilter,
		}
		list, err := promptRepo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(1), list.TotalCount)
	})
}

func TestPromptRepository_NameExists(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	promptRepo := NewPromptRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Prompt NameExists"
	projectName := "Test Project for Prompt NameExists"
	promptName := "Test Prompt NameExists"

	cleanupPrompts(t, db, promptName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupPrompts(t, db, promptName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	t.Run("name does not exist", func(t *testing.T) {
		exists, err := promptRepo.NameExists(ctx, project.ID, promptName)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("name exists", func(t *testing.T) {
		prompt := createTestPrompt(promptName, project.ID)
		err := promptRepo.Create(ctx, prompt)
		require.NoError(t, err)

		exists, err := promptRepo.NameExists(ctx, project.ID, promptName)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestPromptRepository_Versions(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	promptRepo := NewPromptRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Prompt Versions"
	projectName := "Test Project for Prompt Versions"
	promptName := "Test Prompt Versions"
	userEmail := "test-prompt-versions@example.com"

	cleanupPrompts(t, db, promptName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	cleanupUsers(t, db, userEmail)
	defer cleanupPrompts(t, db, promptName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	prompt := createTestPrompt(promptName, project.ID)
	err = promptRepo.Create(ctx, prompt)
	require.NoError(t, err)

	t.Run("create version", func(t *testing.T) {
		version := createTestPromptVersion(prompt.ID, user.ID)
		err := promptRepo.CreateVersion(ctx, version)
		require.NoError(t, err)
		assert.Equal(t, 1, version.Version)
	})

	t.Run("get version", func(t *testing.T) {
		fetched, err := promptRepo.GetVersion(ctx, prompt.ID, 1)
		require.NoError(t, err)
		assert.Equal(t, 1, fetched.Version)
		assert.Contains(t, fetched.Content, "{{name}}")
		assert.Contains(t, fetched.Variables, "name")
		assert.Contains(t, fetched.Variables, "place")
	})

	t.Run("get latest version", func(t *testing.T) {
		// Create another version
		version2 := createTestPromptVersion(prompt.ID, user.ID)
		version2.Content = "Version 2 content"
		version2.Labels = []string{"staging"}
		err := promptRepo.CreateVersion(ctx, version2)
		require.NoError(t, err)
		assert.Equal(t, 2, version2.Version)

		latest, err := promptRepo.GetLatestVersion(ctx, prompt.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, latest.Version)
	})

	t.Run("get version by label", func(t *testing.T) {
		fetched, err := promptRepo.GetVersionByLabel(ctx, prompt.ID, "production")
		require.NoError(t, err)
		assert.Equal(t, 1, fetched.Version)
	})

	t.Run("list versions", func(t *testing.T) {
		versions, err := promptRepo.ListVersions(ctx, prompt.ID)
		require.NoError(t, err)
		assert.Len(t, versions, 2)
		// Should be ordered by version DESC
		assert.Equal(t, 2, versions[0].Version)
		assert.Equal(t, 1, versions[1].Version)
	})

	t.Run("update version labels", func(t *testing.T) {
		versions, err := promptRepo.ListVersions(ctx, prompt.ID)
		require.NoError(t, err)

		newLabels := []string{"production", "latest"}
		err = promptRepo.UpdateVersionLabels(ctx, versions[0].ID, newLabels)
		require.NoError(t, err)

		// Verify by getting by label
		fetched, err := promptRepo.GetVersionByLabel(ctx, prompt.ID, "latest")
		require.NoError(t, err)
		assert.Equal(t, versions[0].Version, fetched.Version)
	})
}
