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

// createTestOrg creates an organization with test data
func createTestOrg(name string) *domain.Organization {
	now := time.Now()
	return &domain.Organization{
		ID:        uuid.New(),
		Name:      name,
		Slug:      "test-org-" + uuid.New().String()[:8],
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestOrgRepository_Create(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewOrgRepository(db)
	ctx := context.Background()
	name := "Test Org Create"

	cleanupOrgs(t, db, name)
	defer cleanupOrgs(t, db, name)

	org := createTestOrg(name)

	err := repo.Create(ctx, org)
	require.NoError(t, err)

	// Verify by fetching
	fetched, err := repo.GetByID(ctx, org.ID)
	require.NoError(t, err)
	assert.Equal(t, org.ID, fetched.ID)
	assert.Equal(t, org.Name, fetched.Name)
	assert.Equal(t, org.Slug, fetched.Slug)
}

func TestOrgRepository_GetByID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewOrgRepository(db)
	ctx := context.Background()
	name := "Test Org GetByID"

	cleanupOrgs(t, db, name)
	defer cleanupOrgs(t, db, name)

	org := createTestOrg(name)
	err := repo.Create(ctx, org)
	require.NoError(t, err)

	t.Run("existing org", func(t *testing.T) {
		fetched, err := repo.GetByID(ctx, org.ID)
		require.NoError(t, err)
		assert.Equal(t, org.ID, fetched.ID)
		assert.Equal(t, org.Name, fetched.Name)
	})

	t.Run("non-existent org", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestOrgRepository_GetBySlug(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewOrgRepository(db)
	ctx := context.Background()
	name := "Test Org GetBySlug"

	cleanupOrgs(t, db, name)
	defer cleanupOrgs(t, db, name)

	org := createTestOrg(name)
	err := repo.Create(ctx, org)
	require.NoError(t, err)

	t.Run("existing slug", func(t *testing.T) {
		fetched, err := repo.GetBySlug(ctx, org.Slug)
		require.NoError(t, err)
		assert.Equal(t, org.ID, fetched.ID)
		assert.Equal(t, org.Slug, fetched.Slug)
	})

	t.Run("non-existent slug", func(t *testing.T) {
		_, err := repo.GetBySlug(ctx, "nonexistent-slug")
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestOrgRepository_Update(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewOrgRepository(db)
	ctx := context.Background()
	name := "Test Org Update"
	updatedName := "Test Org Updated"

	cleanupOrgs(t, db, name, updatedName)
	defer cleanupOrgs(t, db, name, updatedName)

	org := createTestOrg(name)
	err := repo.Create(ctx, org)
	require.NoError(t, err)

	// Update org
	org.Name = updatedName
	err = repo.Update(ctx, org)
	require.NoError(t, err)

	// Verify update
	fetched, err := repo.GetByID(ctx, org.ID)
	require.NoError(t, err)
	assert.Equal(t, updatedName, fetched.Name)
}

func TestOrgRepository_Delete(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewOrgRepository(db)
	ctx := context.Background()
	name := "Test Org Delete"

	cleanupOrgs(t, db, name)

	org := createTestOrg(name)
	err := repo.Create(ctx, org)
	require.NoError(t, err)

	// Verify exists
	_, err = repo.GetByID(ctx, org.ID)
	require.NoError(t, err)

	// Delete
	err = repo.Delete(ctx, org.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = repo.GetByID(ctx, org.ID)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestOrgRepository_SlugExists(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewOrgRepository(db)
	ctx := context.Background()
	name := "Test Org SlugExists"

	cleanupOrgs(t, db, name)
	defer cleanupOrgs(t, db, name)

	org := createTestOrg(name)

	t.Run("slug does not exist", func(t *testing.T) {
		exists, err := repo.SlugExists(ctx, org.Slug)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("slug exists", func(t *testing.T) {
		err := repo.Create(ctx, org)
		require.NoError(t, err)

		exists, err := repo.SlugExists(ctx, org.Slug)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestOrgRepository_Members(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()
	orgName := "Test Org Members"
	userEmail := "test-member@example.com"

	cleanupOrgs(t, db, orgName)
	cleanupUsers(t, db, userEmail)
	defer cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)

	// Create org and user
	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Add member
	member := &domain.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           domain.OrgRoleOwner,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	t.Run("add member", func(t *testing.T) {
		err := orgRepo.AddMember(ctx, member)
		require.NoError(t, err)
	})

	t.Run("get member", func(t *testing.T) {
		fetched, err := orgRepo.GetMember(ctx, org.ID, user.ID)
		require.NoError(t, err)
		assert.Equal(t, member.ID, fetched.ID)
		assert.Equal(t, domain.OrgRoleOwner, fetched.Role)
	})

	t.Run("list by user ID", func(t *testing.T) {
		orgs, err := orgRepo.ListByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, orgs, 1)
		assert.Equal(t, org.ID, orgs[0].ID)
	})

	t.Run("list members", func(t *testing.T) {
		members, err := orgRepo.ListMembers(ctx, org.ID)
		require.NoError(t, err)
		assert.Len(t, members, 1)
		assert.Equal(t, user.ID, members[0].UserID)
		assert.NotNil(t, members[0].User)
		assert.Equal(t, user.Email, members[0].User.Email)
	})

	t.Run("update member role", func(t *testing.T) {
		err := orgRepo.UpdateMemberRole(ctx, org.ID, user.ID, domain.OrgRoleMember)
		require.NoError(t, err)

		fetched, err := orgRepo.GetMember(ctx, org.ID, user.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.OrgRoleMember, fetched.Role)
	})

	t.Run("remove member", func(t *testing.T) {
		err := orgRepo.RemoveMember(ctx, org.ID, user.ID)
		require.NoError(t, err)

		_, err = orgRepo.GetMember(ctx, org.ID, user.ID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestOrgRepository_Invitations(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	userRepo := NewUserRepository(db)
	ctx := context.Background()
	orgName := "Test Org Invitations"
	userEmail := "test-inviter@example.com"

	cleanupOrgs(t, db, orgName)
	cleanupUsers(t, db, userEmail)
	defer cleanupOrgs(t, db, orgName)
	defer cleanupUsers(t, db, userEmail)

	// Create org and user
	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	user := createTestUser(userEmail)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create invitation
	invitation := &domain.OrganizationInvitation{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Email:          "invited@example.com",
		Role:           domain.OrgRoleMember,
		InvitedBy:      user.ID,
		Token:          "test-token-" + uuid.New().String(),
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		CreatedAt:      time.Now(),
	}

	t.Run("create invitation", func(t *testing.T) {
		err := orgRepo.CreateInvitation(ctx, invitation)
		require.NoError(t, err)
	})

	t.Run("get invitation by token", func(t *testing.T) {
		fetched, err := orgRepo.GetInvitationByToken(ctx, invitation.Token)
		require.NoError(t, err)
		assert.Equal(t, invitation.ID, fetched.ID)
		assert.Equal(t, invitation.Email, fetched.Email)
	})

	t.Run("list pending invitations", func(t *testing.T) {
		invitations, err := orgRepo.ListPendingInvitations(ctx, org.ID)
		require.NoError(t, err)
		assert.Len(t, invitations, 1)
		assert.Equal(t, invitation.Email, invitations[0].Email)
	})

	t.Run("accept invitation", func(t *testing.T) {
		err := orgRepo.AcceptInvitation(ctx, invitation.Token)
		require.NoError(t, err)

		// Should no longer be found as pending
		_, err = orgRepo.GetInvitationByToken(ctx, invitation.Token)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
