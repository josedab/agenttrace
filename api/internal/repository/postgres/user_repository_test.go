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

// createTestUser creates a user with test data
func createTestUser(email string) *domain.User {
	now := time.Now()
	return &domain.User{
		ID:            uuid.New(),
		Email:         email,
		EmailVerified: false,
		Name:          "Test User",
		Image:         "https://example.com/avatar.png",
		PasswordHash:  "$2a$10$testpasswordhash",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func TestUserRepository_Create(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()
	email := "test-create@example.com"

	// Cleanup before and after
	cleanupUsers(t, db, email)
	defer cleanupUsers(t, db, email)

	user := createTestUser(email)

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Verify by fetching
	fetched, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, fetched.ID)
	assert.Equal(t, user.Email, fetched.Email)
	assert.Equal(t, user.Name, fetched.Name)
}

func TestUserRepository_GetByID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()
	email := "test-getbyid@example.com"

	cleanupUsers(t, db, email)
	defer cleanupUsers(t, db, email)

	user := createTestUser(email)
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("existing user", func(t *testing.T) {
		fetched, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, fetched.ID)
		assert.Equal(t, user.Email, fetched.Email)
		assert.Equal(t, user.Name, fetched.Name)
	})

	t.Run("non-existent user", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()
	email := "test-getbyemail@example.com"

	cleanupUsers(t, db, email)
	defer cleanupUsers(t, db, email)

	user := createTestUser(email)
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("existing email", func(t *testing.T) {
		fetched, err := repo.GetByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, user.ID, fetched.ID)
		assert.Equal(t, user.Email, fetched.Email)
	})

	t.Run("non-existent email", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestUserRepository_Update(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()
	email := "test-update@example.com"

	cleanupUsers(t, db, email)
	defer cleanupUsers(t, db, email)

	user := createTestUser(email)
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Update user
	user.Name = "Updated Name"
	user.EmailVerified = true
	err = repo.Update(ctx, user)
	require.NoError(t, err)

	// Verify update
	fetched, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", fetched.Name)
	assert.True(t, fetched.EmailVerified)
}

func TestUserRepository_Delete(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()
	email := "test-delete@example.com"

	cleanupUsers(t, db, email)

	user := createTestUser(email)
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Verify exists
	_, err = repo.GetByID(ctx, user.ID)
	require.NoError(t, err)

	// Delete
	err = repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = repo.GetByID(ctx, user.ID)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestUserRepository_ExistsByEmail(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()
	email := "test-exists@example.com"

	cleanupUsers(t, db, email)
	defer cleanupUsers(t, db, email)

	t.Run("user does not exist", func(t *testing.T) {
		exists, err := repo.ExistsByEmail(ctx, email)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("user exists", func(t *testing.T) {
		user := createTestUser(email)
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		exists, err := repo.ExistsByEmail(ctx, email)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestUserRepository_Session(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()
	email := "test-session@example.com"

	cleanupUsers(t, db, email)
	defer cleanupUsers(t, db, email)

	// Create user first
	user := createTestUser(email)
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create session
	session := &domain.UserSession{
		ID:           uuid.New(),
		SessionToken: "test-session-token-" + uuid.New().String(),
		UserID:       user.ID,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	err = repo.CreateSession(ctx, session)
	require.NoError(t, err)

	t.Run("get session by token", func(t *testing.T) {
		fetched, err := repo.GetSessionByToken(ctx, session.SessionToken)
		require.NoError(t, err)
		assert.Equal(t, session.ID, fetched.ID)
		assert.Equal(t, session.UserID, fetched.UserID)
	})

	t.Run("get non-existent session", func(t *testing.T) {
		_, err := repo.GetSessionByToken(ctx, "nonexistent-token")
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("delete session", func(t *testing.T) {
		err := repo.DeleteSession(ctx, session.SessionToken)
		require.NoError(t, err)

		_, err = repo.GetSessionByToken(ctx, session.SessionToken)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestUserRepository_Account(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()
	email := "test-account@example.com"

	cleanupUsers(t, db, email)
	defer cleanupUsers(t, db, email)

	// Create user first
	user := createTestUser(email)
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create account
	accessToken := "test-access-token"
	refreshToken := "test-refresh-token"
	expiresAt := time.Now().Add(24 * time.Hour)
	tokenType := "Bearer"
	scope := "user:email"
	account := &domain.Account{
		ID:                uuid.New(),
		UserID:            user.ID,
		Provider:          "github",
		ProviderAccountID: "12345",
		AccessToken:       &accessToken,
		RefreshToken:      &refreshToken,
		ExpiresAt:         &expiresAt,
		TokenType:         &tokenType,
		Scope:             &scope,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err = repo.CreateAccount(ctx, account)
	require.NoError(t, err)

	t.Run("get account by provider", func(t *testing.T) {
		fetched, err := repo.GetAccountByProvider(ctx, "github", "12345")
		require.NoError(t, err)
		assert.Equal(t, account.ID, fetched.ID)
		assert.Equal(t, account.UserID, fetched.UserID)
		assert.Equal(t, account.Provider, fetched.Provider)
	})

	t.Run("get non-existent account", func(t *testing.T) {
		_, err := repo.GetAccountByProvider(ctx, "github", "nonexistent")
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("upsert account", func(t *testing.T) {
		// Update access token via upsert
		updatedAccessToken := "updated-access-token"
		account.AccessToken = &updatedAccessToken
		err := repo.CreateAccount(ctx, account)
		require.NoError(t, err)

		fetched, err := repo.GetAccountByProvider(ctx, "github", "12345")
		require.NoError(t, err)
		assert.Equal(t, "updated-access-token", *fetched.AccessToken)
	})
}
