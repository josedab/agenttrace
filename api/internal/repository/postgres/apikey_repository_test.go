package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// createTestAPIKey creates an API key with test data
func createTestAPIKey(projectID, createdBy uuid.UUID, name string) *domain.APIKey {
	now := time.Now()
	return &domain.APIKey{
		ID:               uuid.New(),
		ProjectID:        projectID,
		Name:             name,
		PublicKey:        "pk-" + uuid.New().String()[:16],
		SecretKeyHash:    "$2a$10$testhash",
		SecretKeyPreview: "sk-...abc",
		Scopes:           []string{"traces:read", "traces:write"},
		ExpiresAt:        nil,
		CreatedBy:        &createdBy,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func TestAPIKeyRepository_Create(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	apiKeyRepo := NewAPIKeyRepository(db)
	ctx := context.Background()

	// Setup
	orgName := "Test Org for APIKey Create"
	projectName := "Test Project for APIKey Create"
	userEmail := "test-apikey-create@example.com"

	cleanupUsers(t, db, userEmail)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	apiKey := createTestAPIKey(project.ID, user.ID, "Test API Key")

	err = apiKeyRepo.Create(ctx, apiKey)
	require.NoError(t, err)

	// Verify by fetching
	fetched, err := apiKeyRepo.GetByID(ctx, apiKey.ID)
	require.NoError(t, err)
	assert.Equal(t, apiKey.ID, fetched.ID)
	assert.Equal(t, apiKey.Name, fetched.Name)
	assert.Equal(t, apiKey.PublicKey, fetched.PublicKey)
}

func TestAPIKeyRepository_GetByID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	apiKeyRepo := NewAPIKeyRepository(db)
	ctx := context.Background()

	orgName := "Test Org for APIKey GetByID"
	projectName := "Test Project for APIKey GetByID"
	userEmail := "test-apikey-getbyid@example.com"

	cleanupUsers(t, db, userEmail)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	apiKey := createTestAPIKey(project.ID, user.ID, "Test API Key GetByID")
	err = apiKeyRepo.Create(ctx, apiKey)
	require.NoError(t, err)

	t.Run("existing key", func(t *testing.T) {
		fetched, err := apiKeyRepo.GetByID(ctx, apiKey.ID)
		require.NoError(t, err)
		assert.Equal(t, apiKey.ID, fetched.ID)
		assert.Equal(t, apiKey.Name, fetched.Name)
	})

	t.Run("non-existent key", func(t *testing.T) {
		_, err := apiKeyRepo.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestAPIKeyRepository_GetByPublicKey(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	apiKeyRepo := NewAPIKeyRepository(db)
	ctx := context.Background()

	orgName := "Test Org for APIKey GetByPublicKey"
	projectName := "Test Project for APIKey GetByPublicKey"
	userEmail := "test-apikey-getbypk@example.com"

	cleanupUsers(t, db, userEmail)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	apiKey := createTestAPIKey(project.ID, user.ID, "Test API Key GetByPublicKey")
	err = apiKeyRepo.Create(ctx, apiKey)
	require.NoError(t, err)

	t.Run("existing public key", func(t *testing.T) {
		fetched, err := apiKeyRepo.GetByPublicKey(ctx, apiKey.PublicKey)
		require.NoError(t, err)
		assert.Equal(t, apiKey.ID, fetched.ID)
	})

	t.Run("non-existent public key", func(t *testing.T) {
		_, err := apiKeyRepo.GetByPublicKey(ctx, "pk-nonexistent")
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestAPIKeyRepository_Update(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	apiKeyRepo := NewAPIKeyRepository(db)
	ctx := context.Background()

	orgName := "Test Org for APIKey Update"
	projectName := "Test Project for APIKey Update"
	userEmail := "test-apikey-update@example.com"

	cleanupUsers(t, db, userEmail)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	apiKey := createTestAPIKey(project.ID, user.ID, "Test API Key Update")
	err = apiKeyRepo.Create(ctx, apiKey)
	require.NoError(t, err)

	// Update
	apiKey.Name = "Updated API Key"
	apiKey.Scopes = []string{"traces:read"}
	err = apiKeyRepo.Update(ctx, apiKey)
	require.NoError(t, err)

	// Verify
	fetched, err := apiKeyRepo.GetByID(ctx, apiKey.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated API Key", fetched.Name)
	assert.Equal(t, []string{"traces:read"}, fetched.Scopes)
}

func TestAPIKeyRepository_Delete(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	apiKeyRepo := NewAPIKeyRepository(db)
	ctx := context.Background()

	orgName := "Test Org for APIKey Delete"
	projectName := "Test Project for APIKey Delete"
	userEmail := "test-apikey-delete@example.com"

	cleanupUsers(t, db, userEmail)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	apiKey := createTestAPIKey(project.ID, user.ID, "Test API Key Delete")
	err = apiKeyRepo.Create(ctx, apiKey)
	require.NoError(t, err)

	// Verify exists
	_, err = apiKeyRepo.GetByID(ctx, apiKey.ID)
	require.NoError(t, err)

	// Delete
	err = apiKeyRepo.Delete(ctx, apiKey.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = apiKeyRepo.GetByID(ctx, apiKey.ID)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestAPIKeyRepository_ListByProjectID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	apiKeyRepo := NewAPIKeyRepository(db)
	ctx := context.Background()

	orgName := "Test Org for APIKey List"
	projectName := "Test Project for APIKey List"
	userEmail := "test-apikey-list@example.com"

	cleanupUsers(t, db, userEmail)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create multiple API keys
	for i := 0; i < 3; i++ {
		apiKey := createTestAPIKey(project.ID, user.ID, "Test API Key List")
		err = apiKeyRepo.Create(ctx, apiKey)
		require.NoError(t, err)
	}

	keys, err := apiKeyRepo.ListByProjectID(ctx, project.ID)
	require.NoError(t, err)
	assert.Len(t, keys, 3)
}

func TestAPIKeyRepository_CountByProjectID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	apiKeyRepo := NewAPIKeyRepository(db)
	ctx := context.Background()

	orgName := "Test Org for APIKey Count"
	projectName := "Test Project for APIKey Count"
	userEmail := "test-apikey-count@example.com"

	cleanupUsers(t, db, userEmail)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Initially zero
	count, err := apiKeyRepo.CountByProjectID(ctx, project.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Create keys
	for i := 0; i < 2; i++ {
		apiKey := createTestAPIKey(project.ID, user.ID, "Test API Key Count")
		err = apiKeyRepo.Create(ctx, apiKey)
		require.NoError(t, err)
	}

	count, err = apiKeyRepo.CountByProjectID(ctx, project.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestAPIKeyRepository_UpdateLastUsed(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	apiKeyRepo := NewAPIKeyRepository(db)
	ctx := context.Background()

	orgName := "Test Org for APIKey LastUsed"
	projectName := "Test Project for APIKey LastUsed"
	userEmail := "test-apikey-lastused@example.com"

	cleanupUsers(t, db, userEmail)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	apiKey := createTestAPIKey(project.ID, user.ID, "Test API Key LastUsed")
	err = apiKeyRepo.Create(ctx, apiKey)
	require.NoError(t, err)

	// Initially nil
	fetched, err := apiKeyRepo.GetByID(ctx, apiKey.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched.LastUsedAt)

	// Update last used
	err = apiKeyRepo.UpdateLastUsed(ctx, apiKey.ID)
	require.NoError(t, err)

	// Verify updated
	fetched, err = apiKeyRepo.GetByID(ctx, apiKey.ID)
	require.NoError(t, err)
	assert.NotNil(t, fetched.LastUsedAt)
}
