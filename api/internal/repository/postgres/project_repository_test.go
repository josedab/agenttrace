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

// createTestProject creates a project with test data
func createTestProject(name string, orgID uuid.UUID) *domain.Project {
	now := time.Now()
	rateLimit := 1000
	return &domain.Project{
		ID:              uuid.New(),
		OrganizationID:  orgID,
		Name:            name,
		Slug:            "test-project-" + uuid.New().String()[:8],
		Description:     "Test project description",
		Settings:        "",
		RetentionDays:   90,
		RateLimitPerMin: &rateLimit,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func TestProjectRepository_Create(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	ctx := context.Background()
	orgName := "Test Org for Project Create"
	projectName := "Test Project Create"

	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	// Create org first
	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)

	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	// Verify by fetching
	fetched, err := projectRepo.GetByID(ctx, project.ID)
	require.NoError(t, err)
	assert.Equal(t, project.ID, fetched.ID)
	assert.Equal(t, project.Name, fetched.Name)
	assert.Equal(t, project.OrganizationID, fetched.OrganizationID)
}

func TestProjectRepository_GetByID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	ctx := context.Background()
	orgName := "Test Org for Project GetByID"
	projectName := "Test Project GetByID"

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

	t.Run("existing project", func(t *testing.T) {
		fetched, err := projectRepo.GetByID(ctx, project.ID)
		require.NoError(t, err)
		assert.Equal(t, project.ID, fetched.ID)
		assert.Equal(t, project.Name, fetched.Name)
	})

	t.Run("non-existent project", func(t *testing.T) {
		_, err := projectRepo.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestProjectRepository_GetBySlug(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	ctx := context.Background()
	orgName := "Test Org for Project GetBySlug"
	projectName := "Test Project GetBySlug"

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

	t.Run("existing slug", func(t *testing.T) {
		fetched, err := projectRepo.GetBySlug(ctx, org.ID, project.Slug)
		require.NoError(t, err)
		assert.Equal(t, project.ID, fetched.ID)
		assert.Equal(t, project.Slug, fetched.Slug)
	})

	t.Run("non-existent slug", func(t *testing.T) {
		_, err := projectRepo.GetBySlug(ctx, org.ID, "nonexistent-slug")
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestProjectRepository_Update(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	ctx := context.Background()
	orgName := "Test Org for Project Update"
	projectName := "Test Project Update"

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

	// Update project
	project.Description = "Updated description"
	project.RetentionDays = 180
	err = projectRepo.Update(ctx, project)
	require.NoError(t, err)

	// Verify update
	fetched, err := projectRepo.GetByID(ctx, project.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", fetched.Description)
	assert.Equal(t, 180, fetched.RetentionDays)
}

func TestProjectRepository_Delete(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	ctx := context.Background()
	orgName := "Test Org for Project Delete"
	projectName := "Test Project Delete"

	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	// Verify exists
	_, err = projectRepo.GetByID(ctx, project.ID)
	require.NoError(t, err)

	// Delete
	err = projectRepo.Delete(ctx, project.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = projectRepo.GetByID(ctx, project.ID)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestProjectRepository_ListByOrganizationID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	ctx := context.Background()
	orgName := "Test Org for Project List"
	projectName1 := "Test Project List 1"
	projectName2 := "Test Project List 2"

	cleanupProjects(t, db, projectName1, projectName2)
	cleanupOrgs(t, db, orgName)
	defer cleanupProjects(t, db, projectName1, projectName2)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project1 := createTestProject(projectName1, org.ID)
	err = projectRepo.Create(ctx, project1)
	require.NoError(t, err)

	project2 := createTestProject(projectName2, org.ID)
	err = projectRepo.Create(ctx, project2)
	require.NoError(t, err)

	projects, err := projectRepo.ListByOrganizationID(ctx, org.ID)
	require.NoError(t, err)
	assert.Len(t, projects, 2)
}

func TestProjectRepository_SlugExists(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	ctx := context.Background()
	orgName := "Test Org for Project SlugExists"
	projectName := "Test Project SlugExists"

	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)

	t.Run("slug does not exist", func(t *testing.T) {
		exists, err := projectRepo.SlugExists(ctx, org.ID, project.Slug)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("slug exists", func(t *testing.T) {
		err := projectRepo.Create(ctx, project)
		require.NoError(t, err)

		exists, err := projectRepo.SlugExists(ctx, org.ID, project.Slug)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestProjectRepository_Members(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()
	orgName := "Test Org for Project Members"
	projectName := "Test Project Members"
	userEmail := "test-project-member@example.com"

	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	cleanupUsers(t, db, userEmail)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)

	// Create org, project, and user
	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Add member
	member := &domain.ProjectMember{
		ID:        uuid.New(),
		ProjectID: project.ID,
		UserID:    user.ID,
		Role:      domain.OrgRoleAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("add member", func(t *testing.T) {
		err := projectRepo.AddMember(ctx, member)
		require.NoError(t, err)
	})

	t.Run("get member", func(t *testing.T) {
		fetched, err := projectRepo.GetMember(ctx, project.ID, user.ID)
		require.NoError(t, err)
		assert.Equal(t, member.ID, fetched.ID)
		assert.Equal(t, domain.OrgRoleAdmin, fetched.Role)
	})

	t.Run("remove member", func(t *testing.T) {
		err := projectRepo.RemoveMember(ctx, project.ID, user.ID)
		require.NoError(t, err)

		_, err = projectRepo.GetMember(ctx, project.ID, user.ID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
