package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) CreateAccount(ctx context.Context, account *domain.Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockUserRepository) GetAccountByProvider(ctx context.Context, provider, providerAccountID string) (*domain.Account, error) {
	args := m.Called(ctx, provider, providerAccountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockUserRepository) CreateSession(ctx context.Context, session *domain.UserSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockUserRepository) GetSessionByToken(ctx context.Context, token string) (*domain.UserSession, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserSession), args.Error(1)
}

func (m *MockUserRepository) DeleteSession(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// MockAPIKeyRepository is a mock implementation of APIKeyRepository
type MockAPIKeyRepository struct {
	mock.Mock
}

func (m *MockAPIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockAPIKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) GetByPublicKey(ctx context.Context, publicKey string) (*domain.APIKey, error) {
	args := m.Called(ctx, publicKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) Update(ctx context.Context, key *domain.APIKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockAPIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAPIKeyRepository) ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]domain.APIKey, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAPIKeyRepository) GetProjectIDByPublicKey(ctx context.Context, publicKey string) (*uuid.UUID, error) {
	args := m.Called(ctx, publicKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*uuid.UUID), args.Error(1)
}

// MockOrgRepository is a mock implementation of OrgRepository
type MockOrgRepository struct {
	mock.Mock
}

func (m *MockOrgRepository) Create(ctx context.Context, org *domain.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockOrgRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *MockOrgRepository) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *MockOrgRepository) Update(ctx context.Context, org *domain.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockOrgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrgRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Organization, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Organization), args.Error(1)
}

func (m *MockOrgRepository) AddMember(ctx context.Context, member *domain.OrganizationMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockOrgRepository) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*domain.OrganizationMember, error) {
	args := m.Called(ctx, orgID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrganizationMember), args.Error(1)
}

func (m *MockOrgRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	args := m.Called(ctx, slug)
	return args.Bool(0), args.Error(1)
}

// MockProjectRepository is a mock implementation of ProjectRepository
type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *MockProjectRepository) GetUserRoleForProject(ctx context.Context, projectID, userID uuid.UUID) (*domain.OrgRole, error) {
	args := m.Called(ctx, projectID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrgRole), args.Error(1)
}

// Helper function to create test config
func testConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:        "test-secret-key-for-testing-purposes-only",
			Issuer:        "agenttrace-test",
			AccessExpiry:  15,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
	}
}

func TestAuthService_Register(t *testing.T) {
	t.Run("successfully registers new user", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		userRepo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
		userRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
		orgRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Organization")).Return(nil)
		orgRepo.On("AddMember", mock.Anything, mock.AnythingOfType("*domain.OrganizationMember")).Return(nil)
		projectRepo.On("Create", mock.Anything, mock.MatchedBy(func(project *domain.Project) bool {
			return project.Name == "Default Project" &&
				project.Slug == "default-project" &&
				project.Settings == `{"systemProvisioned":true}` &&
				project.OrganizationID != uuid.Nil
		})).Return(nil)
		requestStartedAt := time.Now()
		userRepo.On(
			"CreateSession",
			mock.Anything,
			mock.MatchedBy(func(session *domain.UserSession) bool {
				expectedExpiry := requestStartedAt.Add(7 * 24 * time.Hour)
				return session.ExpiresAt.Sub(expectedExpiry).Abs() < time.Second
			}),
		).Return(nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		result, err := svc.Register(context.Background(), &domain.RegisterInput{
			Email:    "test@example.com",
			Password: "securepassword123",
			Name:     "Test User",
		})

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)
		assert.Equal(t, "test@example.com", result.User.Email)
		assert.Equal(t, "Test User", result.User.Name)

		userRepo.AssertExpectations(t)
		orgRepo.AssertExpectations(t)
		projectRepo.AssertExpectations(t)
	})

	t.Run("fails if email already exists", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		userRepo.On("ExistsByEmail", mock.Anything, "existing@example.com").Return(true, nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		result, err := svc.Register(context.Background(), &domain.RegisterInput{
			Email:    "existing@example.com",
			Password: "password123",
			Name:     "Test User",
		})

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperrors.IsValidation(err))
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Run("successfully logs in with valid credentials", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		// Generate valid password hash for "correctpassword"
		passwordHash, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &domain.User{
			ID:           uuid.New(),
			Email:        "user@example.com",
			Name:         "Test User",
			PasswordHash: string(passwordHash),
		}

		userRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
		userRepo.On("CreateSession", mock.Anything, mock.AnythingOfType("*domain.UserSession")).Return(nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		result, err := svc.Login(context.Background(), &domain.LoginInput{
			Email:    "user@example.com",
			Password: "correctpassword",
		})

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)
	})

	t.Run("fails with wrong password", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		// Generate valid password hash for "correctpassword"
		passwordHash, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &domain.User{
			ID:           uuid.New(),
			Email:        "user@example.com",
			PasswordHash: string(passwordHash),
		}

		userRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		result, err := svc.Login(context.Background(), &domain.LoginInput{
			Email:    "user@example.com",
			Password: "wrongpassword",
		})

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperrors.IsUnauthorized(err))
	})

	t.Run("fails for non-existent user", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		userRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, apperrors.NotFound("user not found"))

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		result, err := svc.Login(context.Background(), &domain.LoginInput{
			Email:    "notfound@example.com",
			Password: "password123",
		})

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperrors.IsUnauthorized(err))
	})

	t.Run("fails for OAuth-only user", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		user := &domain.User{
			ID:           uuid.New(),
			Email:        "oauth@example.com",
			PasswordHash: "", // No password hash for OAuth users
		}

		userRepo.On("GetByEmail", mock.Anything, "oauth@example.com").Return(user, nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		result, err := svc.Login(context.Background(), &domain.LoginInput{
			Email:    "oauth@example.com",
			Password: "anypassword",
		})

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperrors.IsUnauthorized(err))
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	t.Run("successfully refreshes token", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		userID := uuid.New()
		session := &domain.UserSession{
			ID:           uuid.New(),
			SessionToken: "valid-refresh-token",
			UserID:       userID,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}
		user := &domain.User{
			ID:    userID,
			Email: "user@example.com",
		}

		userRepo.On("GetSessionByToken", mock.Anything, "valid-refresh-token").Return(session, nil)
		userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
		userRepo.On("DeleteSession", mock.Anything, "valid-refresh-token").Return(nil)
		userRepo.On("CreateSession", mock.Anything, mock.AnythingOfType("*domain.UserSession")).Return(nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		result, err := svc.RefreshToken(context.Background(), "valid-refresh-token")

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.AccessToken)
		// Refresh token should be rotated (different from the original)
		assert.NotEqual(t, "valid-refresh-token", result.RefreshToken)
		assert.NotEmpty(t, result.RefreshToken)
		userRepo.AssertExpectations(t)
	})

	t.Run("fails with invalid refresh token", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		userRepo.On("GetSessionByToken", mock.Anything, "invalid-token").Return(nil, apperrors.NotFound("session not found"))

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		result, err := svc.RefreshToken(context.Background(), "invalid-token")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperrors.IsUnauthorized(err))
	})
}

func TestAuthService_Logout(t *testing.T) {
	t.Run("successfully logs out", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		userRepo.On("DeleteSession", mock.Anything, "session-token").Return(nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		err := svc.Logout(context.Background(), "session-token")

		require.NoError(t, err)
		userRepo.AssertExpectations(t)
	})
}

func TestAuthService_ValidateJWT(t *testing.T) {
	t.Run("validates valid JWT token", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		// First generate a valid token
		user := &domain.User{
			ID:    uuid.New(),
			Email: "user@example.com",
		}
		token, err := svc.generateAccessToken(user)
		require.NoError(t, err)

		// Then validate it
		claims, err := svc.ValidateJWT(context.Background(), token)

		require.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, user.ID.String(), claims.UserID)
		assert.Equal(t, "user@example.com", claims.Email)
	})

	t.Run("fails with invalid token", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		claims, err := svc.ValidateJWT(context.Background(), "invalid.jwt.token")

		require.Error(t, err)
		assert.Nil(t, claims)
		assert.True(t, apperrors.IsUnauthorized(err))
	})

	t.Run("fails with wrong secret", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		cfg1 := testConfig()
		cfg1.JWT.Secret = "secret-one"
		svc1 := NewAuthService(cfg1, userRepo, apiKeyRepo, orgRepo, projectRepo)

		cfg2 := testConfig()
		cfg2.JWT.Secret = "secret-two"
		svc2 := NewAuthService(cfg2, userRepo, apiKeyRepo, orgRepo, projectRepo)

		// Generate token with first secret
		user := &domain.User{ID: uuid.New(), Email: "user@example.com"}
		token, _ := svc1.generateAccessToken(user)

		// Try to validate with different secret
		claims, err := svc2.ValidateJWT(context.Background(), token)

		require.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestAuthService_ValidateAPIKeyPair(t *testing.T) {
	t.Run("validates valid API key pair", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		// Generate a key pair for testing
		secretKey := "sk-testsecretkey1234567890abcdef1234567890abcdef"
		secretKeyHash, err := svc.hashSecretKey(secretKey)
		require.NoError(t, err)

		projectID := uuid.New()
		apiKey := &domain.APIKey{
			ID:            uuid.New(),
			ProjectID:     projectID,
			PublicKey:     "pk-testpublickey12345678",
			SecretKeyHash: secretKeyHash,
			ExpiresAt:     nil, // Never expires
		}

		apiKeyRepo.On("GetByPublicKey", mock.Anything, "pk-testpublickey12345678").Return(apiKey, nil)
		apiKeyRepo.On("UpdateLastUsed", mock.Anything, apiKey.ID).Return(nil)

		result, err := svc.validateAPIKeyPair(context.Background(), "pk-testpublickey12345678", secretKey)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, projectID, result.ProjectID)
	})

	t.Run("fails with invalid public key", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		apiKeyRepo.On("GetByPublicKey", mock.Anything, "pk-invalid").Return(nil, apperrors.NotFound("not found"))

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		result, err := svc.validateAPIKeyPair(context.Background(), "pk-invalid", "sk-anything")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperrors.IsUnauthorized(err))
	})

	t.Run("fails with wrong secret key", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		correctSecretHash, err := svc.hashSecretKey("sk-correctkey")
		require.NoError(t, err)
		apiKey := &domain.APIKey{
			ID:            uuid.New(),
			ProjectID:     uuid.New(),
			PublicKey:     "pk-test",
			SecretKeyHash: correctSecretHash,
		}

		apiKeyRepo.On("GetByPublicKey", mock.Anything, "pk-test").Return(apiKey, nil)

		result, err := svc.validateAPIKeyPair(context.Background(), "pk-test", "sk-wrongkey")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperrors.IsUnauthorized(err))
	})

	t.Run("fails with expired key", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		secretKey := "sk-expiredsecret"
		secretKeyHash, err := svc.hashSecretKey(secretKey)
		require.NoError(t, err)
		expiredTime := time.Now().Add(-24 * time.Hour) // Expired yesterday

		apiKey := &domain.APIKey{
			ID:            uuid.New(),
			ProjectID:     uuid.New(),
			PublicKey:     "pk-expired",
			SecretKeyHash: secretKeyHash,
			ExpiresAt:     &expiredTime,
		}

		apiKeyRepo.On("GetByPublicKey", mock.Anything, "pk-expired").Return(apiKey, nil)

		result, err := svc.validateAPIKeyPair(context.Background(), "pk-expired", secretKey)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperrors.IsUnauthorized(err))
	})
}

func TestAuthService_AuthenticateAPIKey(t *testing.T) {
	t.Run("validates canonical secret credential", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)
		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		keyID := "0123456789abcdef0123456789abcdef"
		credential := "sk-at-" + keyID + ".secret"
		secretKeyHash, err := svc.hashSecretKey(credential)
		require.NoError(t, err)
		apiKey := &domain.APIKey{
			ID:            uuid.New(),
			ProjectID:     uuid.New(),
			PublicKey:     "pk-at-" + keyID,
			SecretKeyHash: secretKeyHash,
			Scopes:        []string{"traces:read"},
		}
		apiKeyRepo.On("GetByPublicKey", mock.Anything, apiKey.PublicKey).Return(apiKey, nil)
		apiKeyRepo.On("UpdateLastUsed", mock.Anything, apiKey.ID).Return(nil)

		result, err := svc.AuthenticateAPIKey(context.Background(), credential)

		require.NoError(t, err)
		assert.Equal(t, apiKey.ID, result.APIKeyID)
		assert.Equal(t, apiKey.ProjectID, result.ProjectID)
		assert.Equal(t, apiKey.Scopes, result.Scopes)
	})

	t.Run("rejects public identifier without secret", func(t *testing.T) {
		svc := NewAuthService(
			testConfig(),
			new(MockUserRepository),
			new(MockAPIKeyRepository),
			new(MockOrgRepository),
			new(MockProjectRepository),
		)

		result, err := svc.AuthenticateAPIKey(
			context.Background(),
			"pk-at-0123456789abcdef0123456789abcdef",
		)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperrors.IsUnauthorized(err))
	})

	t.Run("populates UserID from owning creator", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)
		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		keyID := "0123456789abcdef0123456789abcdef"
		credential := "sk-at-" + keyID + ".secret"
		secretKeyHash, err := svc.hashSecretKey(credential)
		require.NoError(t, err)
		owner := uuid.New()
		apiKey := &domain.APIKey{
			ID:            uuid.New(),
			ProjectID:     uuid.New(),
			PublicKey:     "pk-at-" + keyID,
			SecretKeyHash: secretKeyHash,
			Scopes:        []string{"traces:read"},
			CreatedBy:     &owner,
		}
		apiKeyRepo.On("GetByPublicKey", mock.Anything, apiKey.PublicKey).Return(apiKey, nil)
		apiKeyRepo.On("UpdateLastUsed", mock.Anything, apiKey.ID).Return(nil)

		result, err := svc.AuthenticateAPIKey(context.Background(), credential)

		require.NoError(t, err)
		require.NotNil(t, result.UserID)
		assert.Equal(t, owner, *result.UserID)
		// The returned pointer must be a copy so callers cannot mutate the key.
		assert.NotSame(t, apiKey.CreatedBy, result.UserID)
	})

	t.Run("leaves UserID nil for unowned legacy key", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)
		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		keyID := "abcdef0123456789abcdef0123456789"
		credential := "sk-at-" + keyID + ".secret"
		secretKeyHash, err := svc.hashSecretKey(credential)
		require.NoError(t, err)
		apiKey := &domain.APIKey{
			ID:            uuid.New(),
			ProjectID:     uuid.New(),
			PublicKey:     "pk-at-" + keyID,
			SecretKeyHash: secretKeyHash,
			Scopes:        []string{"traces:read"},
			CreatedBy:     nil,
		}
		apiKeyRepo.On("GetByPublicKey", mock.Anything, apiKey.PublicKey).Return(apiKey, nil)
		apiKeyRepo.On("UpdateLastUsed", mock.Anything, apiKey.ID).Return(nil)

		result, err := svc.AuthenticateAPIKey(context.Background(), credential)

		require.NoError(t, err)
		assert.Nil(t, result.UserID)
	})
}

func TestAuthService_CreateAPIKey(t *testing.T) {
	t.Run("creates API key with default scopes", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		apiKeyRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.APIKey")).Return(nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		projectID := uuid.New()
		userID := uuid.New()
		result, err := svc.CreateAPIKey(context.Background(), projectID, &domain.APIKeyInput{
			Name: "Test Key",
		}, userID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.APIKey)
		assert.NotEmpty(t, result.SecretKey)
		assert.True(t, strings.HasPrefix(result.APIKey.PublicKey, "pk-at-"))
		assert.True(t, strings.HasPrefix(result.SecretKey, "sk-at-"))
		parsedPublicKey, parsedSecretKey, parseErr := parseAPIKeyCredential(result.SecretKey)
		require.NoError(t, parseErr)
		assert.Equal(t, result.APIKey.PublicKey, parsedPublicKey)
		assert.Equal(t, result.SecretKey, parsedSecretKey)
		assert.Equal(t, "Test Key", result.APIKey.Name)
		assert.Equal(t, projectID, result.APIKey.ProjectID)
	})

	t.Run("creates API key with custom scopes", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		apiKeyRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.APIKey")).Return(nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		projectID := uuid.New()
		userID := uuid.New()
		customScopes := []string{"traces:read", "traces:write"}
		result, err := svc.CreateAPIKey(context.Background(), projectID, &domain.APIKeyInput{
			Name:   "Limited Key",
			Scopes: customScopes,
		}, userID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, customScopes, result.APIKey.Scopes)
	})

	t.Run("rejects unknown scopes", func(t *testing.T) {
		svc := NewAuthService(
			testConfig(),
			new(MockUserRepository),
			new(MockAPIKeyRepository),
			new(MockOrgRepository),
			new(MockProjectRepository),
		)

		result, err := svc.CreateAPIKey(
			context.Background(),
			uuid.New(),
			&domain.APIKeyInput{Name: "Invalid", Scopes: []string{"everything:*"}},
			uuid.New(),
		)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperrors.IsValidation(err))
	})

	t.Run("creates API key with expiration", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		apiKeyRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.APIKey")).Return(nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		projectID := uuid.New()
		userID := uuid.New()
		expiresAt := time.Now().Add(30 * 24 * time.Hour)
		result, err := svc.CreateAPIKey(context.Background(), projectID, &domain.APIKeyInput{
			Name:      "Expiring Key",
			ExpiresAt: &expiresAt,
		}, userID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.APIKey.ExpiresAt)
		assert.Equal(t, expiresAt.Unix(), result.APIKey.ExpiresAt.Unix())
	})
}

func TestAuthService_DeleteAPIKey(t *testing.T) {
	t.Run("deletes API key", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		keyID := uuid.New()
		apiKeyRepo.On("Delete", mock.Anything, keyID).Return(nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		err := svc.DeleteAPIKey(context.Background(), keyID)

		require.NoError(t, err)
		apiKeyRepo.AssertExpectations(t)
	})
}

func TestAuthService_ListAPIKeys(t *testing.T) {
	t.Run("lists project API keys", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		projectID := uuid.New()
		keys := []domain.APIKey{
			{ID: uuid.New(), Name: "Key 1", ProjectID: projectID},
			{ID: uuid.New(), Name: "Key 2", ProjectID: projectID},
		}
		apiKeyRepo.On("ListByProjectID", mock.Anything, projectID).Return(keys, nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		result, err := svc.ListAPIKeys(context.Background(), projectID)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Key 1", result[0].Name)
		assert.Equal(t, "Key 2", result[1].Name)
	})
}

func TestAuthService_GetUserByID(t *testing.T) {
	t.Run("gets user by ID", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		userID := uuid.New()
		user := &domain.User{
			ID:    userID,
			Email: "user@example.com",
			Name:  "Test User",
		}
		userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		result, err := svc.GetUserByID(context.Background(), userID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.ID)
		assert.Equal(t, "user@example.com", result.Email)
	})
}

func TestAuthService_CheckProjectAccess(t *testing.T) {
	t.Run("allows access when user has sufficient role", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		projectID := uuid.New()
		userID := uuid.New()
		adminRole := domain.OrgRoleAdmin
		projectRepo.On("GetUserRoleForProject", mock.Anything, projectID, userID).Return(&adminRole, nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		err := svc.CheckProjectAccess(context.Background(), projectID, userID, domain.OrgRoleMember)

		require.NoError(t, err)
	})

	t.Run("denies access when user has insufficient role", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		projectID := uuid.New()
		userID := uuid.New()
		viewerRole := domain.OrgRoleViewer
		projectRepo.On("GetUserRoleForProject", mock.Anything, projectID, userID).Return(&viewerRole, nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		err := svc.CheckProjectAccess(context.Background(), projectID, userID, domain.OrgRoleAdmin)

		require.Error(t, err)
		assert.True(t, apperrors.IsForbidden(err))
	})

	t.Run("denies access when user has no role", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		projectID := uuid.New()
		userID := uuid.New()
		projectRepo.On("GetUserRoleForProject", mock.Anything, projectID, userID).Return(nil, nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		err := svc.CheckProjectAccess(context.Background(), projectID, userID, domain.OrgRoleViewer)

		require.Error(t, err)
		assert.True(t, apperrors.IsForbidden(err))
	})
}

func TestAuthService_HasRequiredRole(t *testing.T) {
	svc := &AuthService{}

	tests := []struct {
		name         string
		userRole     domain.OrgRole
		requiredRole domain.OrgRole
		expected     bool
	}{
		{"owner >= owner", domain.OrgRoleOwner, domain.OrgRoleOwner, true},
		{"owner >= admin", domain.OrgRoleOwner, domain.OrgRoleAdmin, true},
		{"owner >= member", domain.OrgRoleOwner, domain.OrgRoleMember, true},
		{"owner >= viewer", domain.OrgRoleOwner, domain.OrgRoleViewer, true},
		{"admin >= admin", domain.OrgRoleAdmin, domain.OrgRoleAdmin, true},
		{"admin >= member", domain.OrgRoleAdmin, domain.OrgRoleMember, true},
		{"admin >= viewer", domain.OrgRoleAdmin, domain.OrgRoleViewer, true},
		{"admin < owner", domain.OrgRoleAdmin, domain.OrgRoleOwner, false},
		{"member >= member", domain.OrgRoleMember, domain.OrgRoleMember, true},
		{"member >= viewer", domain.OrgRoleMember, domain.OrgRoleViewer, true},
		{"member < admin", domain.OrgRoleMember, domain.OrgRoleAdmin, false},
		{"viewer >= viewer", domain.OrgRoleViewer, domain.OrgRoleViewer, true},
		{"viewer < member", domain.OrgRoleViewer, domain.OrgRoleMember, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.hasRequiredRole(tt.userRole, tt.requiredRole)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthService_APIKeyGeneration(t *testing.T) {
	svc := &AuthService{}

	t.Run("generates valid key pair", func(t *testing.T) {
		publicKey, secretKey, err := svc.generateAPIKeyPair()

		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(publicKey, "pk-at-"))
		assert.True(t, strings.HasPrefix(secretKey, "sk-at-"))
		assert.Len(t, publicKey, 38)
		assert.Len(t, secretKey, 71)
	})

	t.Run("hash and verify secret key", func(t *testing.T) {
		secretKey := "sk-testsecret1234567890"
		hash, err := svc.hashSecretKey(secretKey)
		require.NoError(t, err)

		assert.NotEqual(t, secretKey, hash)
		assert.True(t, svc.verifySecretKey(secretKey, hash))
		assert.False(t, svc.verifySecretKey("wrong-key", hash))
	})

	t.Run("rejects secrets longer than bcrypt supports", func(t *testing.T) {
		_, err := svc.hashSecretKey(strings.Repeat("x", 73))
		require.Error(t, err)
	})
}

func TestAuthService_ValidateOAuthCallbackSecret(t *testing.T) {
	cfg := testConfig()
	cfg.OAuth.CallbackSecret = "trusted-callback-secret"
	svc := NewAuthService(
		cfg,
		new(MockUserRepository),
		new(MockAPIKeyRepository),
		new(MockOrgRepository),
		new(MockProjectRepository),
	)

	assert.True(t, svc.ValidateOAuthCallbackSecret("trusted-callback-secret"))
	assert.False(t, svc.ValidateOAuthCallbackSecret("wrong-secret"))
	assert.False(t, svc.ValidateOAuthCallbackSecret(""))
}

func TestAuthService_RefreshToken_ExpiredSession(t *testing.T) {
	t.Run("expired session returns Unauthorized", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		userID := uuid.New()
		session := &domain.UserSession{
			ID:           uuid.New(),
			SessionToken: "expired-refresh-token",
			UserID:       userID,
			ExpiresAt:    time.Now().Add(-1 * time.Hour), // expired 1 hour ago
		}

		userRepo.On("GetSessionByToken", mock.Anything, "expired-refresh-token").Return(session, nil)
		userRepo.On("DeleteSession", mock.Anything, "expired-refresh-token").Return(nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		result, err := svc.RefreshToken(context.Background(), "expired-refresh-token")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperrors.IsUnauthorized(err))
	})

	t.Run("caps legacy overflowed expiration", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		session := &domain.UserSession{
			ID:           uuid.New(),
			SessionToken: "overflowed-refresh-token",
			UserID:       uuid.New(),
			CreatedAt:    time.Now().Add(-8 * 24 * time.Hour),
			ExpiresAt:    time.Now().Add(100 * 365 * 24 * time.Hour),
		}
		userRepo.On(
			"GetSessionByToken",
			mock.Anything,
			"overflowed-refresh-token",
		).Return(session, nil)
		userRepo.On(
			"DeleteSession",
			mock.Anything,
			"overflowed-refresh-token",
		).Return(nil)

		svc := NewAuthService(
			testConfig(),
			userRepo,
			new(MockAPIKeyRepository),
			new(MockOrgRepository),
			new(MockProjectRepository),
		)

		result, err := svc.RefreshToken(
			context.Background(),
			"overflowed-refresh-token",
		)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperrors.IsUnauthorized(err))
	})

	t.Run("session expiring within seconds - race window", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		userID := uuid.New()
		// Session that expires in 1 millisecond (race window)
		session := &domain.UserSession{
			ID:           uuid.New(),
			SessionToken: "near-expired-token",
			UserID:       userID,
			ExpiresAt:    time.Now().Add(-1 * time.Millisecond),
		}

		userRepo.On("GetSessionByToken", mock.Anything, "near-expired-token").Return(session, nil)
		userRepo.On("DeleteSession", mock.Anything, "near-expired-token").Return(nil)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		result, err := svc.RefreshToken(context.Background(), "near-expired-token")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperrors.IsUnauthorized(err))
	})

	t.Run("refresh after session deleted - concurrent logout", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		userRepo.On("GetSessionByToken", mock.Anything, "deleted-token").Return(nil, apperrors.NotFound("session not found"))

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		result, err := svc.RefreshToken(context.Background(), "deleted-token")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperrors.IsUnauthorized(err))
	})

	t.Run("ignored DeleteSession error does not break flow", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		apiKeyRepo := new(MockAPIKeyRepository)
		orgRepo := new(MockOrgRepository)
		projectRepo := new(MockProjectRepository)

		userID := uuid.New()
		session := &domain.UserSession{
			ID:           uuid.New(),
			SessionToken: "expired-token-with-delete-error",
			UserID:       userID,
			ExpiresAt:    time.Now().Add(-1 * time.Hour),
		}

		userRepo.On("GetSessionByToken", mock.Anything, "expired-token-with-delete-error").Return(session, nil)
		// DeleteSession fails, but the error is ignored (line 304: _ = s.userRepo.DeleteSession())
		userRepo.On("DeleteSession", mock.Anything, "expired-token-with-delete-error").Return(assert.AnError)

		svc := NewAuthService(testConfig(), userRepo, apiKeyRepo, orgRepo, projectRepo)

		result, err := svc.RefreshToken(context.Background(), "expired-token-with-delete-error")

		// Should still return Unauthorized regardless of DeleteSession failure
		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperrors.IsUnauthorized(err))
	})
}
