package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// UserRepository defines user repository operations
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	CreateAccount(ctx context.Context, account *domain.Account) error
	GetAccountByProvider(ctx context.Context, provider, providerAccountID string) (*domain.Account, error)
	CreateSession(ctx context.Context, session *domain.UserSession) error
	GetSessionByToken(ctx context.Context, token string) (*domain.UserSession, error)
	DeleteSession(ctx context.Context, token string) error
}

// APIKeyRepository defines API key repository operations
type APIKeyRepository interface {
	Create(ctx context.Context, key *domain.APIKey) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error)
	GetByPublicKey(ctx context.Context, publicKey string) (*domain.APIKey, error)
	Update(ctx context.Context, key *domain.APIKey) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]domain.APIKey, error)
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
	GetProjectIDByPublicKey(ctx context.Context, publicKey string) (*uuid.UUID, error)
}

// OrgRepository defines organization repository operations
type OrgRepository interface {
	Create(ctx context.Context, org *domain.Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	Update(ctx context.Context, org *domain.Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Organization, error)
	AddMember(ctx context.Context, member *domain.OrganizationMember) error
	GetMember(ctx context.Context, orgID, userID uuid.UUID) (*domain.OrganizationMember, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
}

// ProjectRepository defines project repository operations
type ProjectRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	GetUserRoleForProject(ctx context.Context, projectID, userID uuid.UUID) (*domain.OrgRole, error)
}

// AuditLogger defines the interface for audit logging
type AuditLogger interface {
	LogLogin(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, email, ipAddress, userAgent string) error
	LogLoginFailed(ctx context.Context, orgID uuid.UUID, email, ipAddress, userAgent, reason string) error
	LogLogout(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, email string) error
	LogSSOLogin(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, email, provider, ipAddress, userAgent string) error
	LogAPIKeyCreated(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail string, keyID uuid.UUID, keyName string) error
	LogAPIKeyRevoked(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail string, keyID uuid.UUID, keyName string) error
	LogUserCreated(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail string, newUserID uuid.UUID, newUserEmail string) error
}

// AuthService handles authentication and authorization
type AuthService struct {
	cfg         *config.Config
	userRepo    UserRepository
	apiKeyRepo  APIKeyRepository
	orgRepo     OrgRepository
	projectRepo ProjectRepository
	auditLogger AuditLogger
}

// NewAuthService creates a new auth service
func NewAuthService(
	cfg *config.Config,
	userRepo UserRepository,
	apiKeyRepo APIKeyRepository,
	orgRepo OrgRepository,
	projectRepo ProjectRepository,
) *AuthService {
	return &AuthService{
		cfg:         cfg,
		userRepo:    userRepo,
		apiKeyRepo:  apiKeyRepo,
		orgRepo:     orgRepo,
		projectRepo: projectRepo,
	}
}

// SetAuditLogger sets the audit logger for the auth service
// This allows optional audit logging without making it a required dependency
func (s *AuthService) SetAuditLogger(logger AuditLogger) {
	s.auditLogger = logger
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, input *domain.RegisterInput) (*domain.AuthResult, error) {
	// Check if email exists
	exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, apperrors.Validation("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	now := time.Now()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        input.Email,
		Name:         input.Name,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default organization
	org := &domain.Organization{
		ID:        uuid.New(),
		Name:      input.Name + "'s Organization",
		Slug:      domain.GenerateSlug(input.Name + "'s Organization"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Add user as owner
	member := &domain.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           domain.OrgRoleOwner,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.orgRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token as session
	session := &domain.UserSession{
		ID:           uuid.New(),
		SessionToken: refreshToken,
		UserID:       user.ID,
		ExpiresAt:    time.Now().Add(time.Duration(s.cfg.JWT.RefreshExpiry) * time.Hour),
		CreatedAt:    now,
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Audit log: user registration
	if s.auditLogger != nil {
		go func() {
			_ = s.auditLogger.LogUserCreated(context.Background(), org.ID, user.ID, user.Email, user.ID, user.Email)
		}()
	}

	return &domain.AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(s.cfg.JWT.AccessExpiry) * time.Minute),
	}, nil
}

// Login authenticates a user with email and password
func (s *AuthService) Login(ctx context.Context, input *domain.LoginInput) (*domain.AuthResult, error) {
	return s.LoginWithContext(ctx, input, "", "")
}

// LoginWithContext authenticates a user with email/password and request context for audit logging
func (s *AuthService) LoginWithContext(ctx context.Context, input *domain.LoginInput, ipAddress, userAgent string) (*domain.AuthResult, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, apperrors.Unauthorized("invalid credentials")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get user's primary organization for audit logging
	var primaryOrgID uuid.UUID
	if s.auditLogger != nil {
		orgs, err := s.orgRepo.ListByUserID(ctx, user.ID)
		if err == nil && len(orgs) > 0 {
			primaryOrgID = orgs[0].ID
		}
	}

	if user.PasswordHash == "" {
		// Audit log: failed login (OAuth user tried password login)
		if s.auditLogger != nil && primaryOrgID != uuid.Nil {
			go func() {
				_ = s.auditLogger.LogLoginFailed(context.Background(), primaryOrgID, input.Email, ipAddress, userAgent, "OAuth user attempted password login")
			}()
		}
		return nil, apperrors.Unauthorized("please login with your OAuth provider")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		// Audit log: failed login (invalid password)
		if s.auditLogger != nil && primaryOrgID != uuid.Nil {
			go func() {
				_ = s.auditLogger.LogLoginFailed(context.Background(), primaryOrgID, input.Email, ipAddress, userAgent, "invalid password")
			}()
		}
		return nil, apperrors.Unauthorized("invalid credentials")
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token as session
	now := time.Now()
	session := &domain.UserSession{
		ID:           uuid.New(),
		SessionToken: refreshToken,
		UserID:       user.ID,
		ExpiresAt:    time.Now().Add(time.Duration(s.cfg.JWT.RefreshExpiry) * time.Hour),
		CreatedAt:    now,
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Audit log: successful login
	if s.auditLogger != nil && primaryOrgID != uuid.Nil {
		go func() {
			_ = s.auditLogger.LogLogin(context.Background(), primaryOrgID, user.ID, user.Email, ipAddress, userAgent)
		}()
	}

	return &domain.AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(s.cfg.JWT.AccessExpiry) * time.Minute),
	}, nil
}

// RefreshToken generates new tokens from a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResult, error) {
	session, err := s.userRepo.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, apperrors.Unauthorized("invalid refresh token")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &domain.AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(s.cfg.JWT.AccessExpiry) * time.Minute),
	}, nil
}

// Logout invalidates a session
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.LogoutWithContext(ctx, refreshToken, uuid.Nil, "")
}

// LogoutWithContext invalidates a session with audit logging context
func (s *AuthService) LogoutWithContext(ctx context.Context, refreshToken string, userID uuid.UUID, userEmail string) error {
	// If we have audit logger and user context, log the logout
	if s.auditLogger != nil && userID != uuid.Nil {
		// Get user's primary organization for audit logging
		orgs, err := s.orgRepo.ListByUserID(ctx, userID)
		if err == nil && len(orgs) > 0 {
			go func() {
				_ = s.auditLogger.LogLogout(context.Background(), orgs[0].ID, userID, userEmail)
			}()
		}
	}

	return s.userRepo.DeleteSession(ctx, refreshToken)
}

// ValidateJWT validates a JWT access token
func (s *AuthService) ValidateJWT(ctx context.Context, tokenString string) (*domain.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, apperrors.Unauthorized("invalid token")
	}

	claims, ok := token.Claims.(*domain.JWTClaims)
	if !ok || !token.Valid {
		return nil, apperrors.Unauthorized("invalid token")
	}

	return claims, nil
}

// ValidateAPIKey validates an API key and returns project info
func (s *AuthService) ValidateAPIKey(ctx context.Context, publicKey, secretKey string) (*uuid.UUID, error) {
	apiKey, err := s.apiKeyRepo.GetByPublicKey(ctx, publicKey)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, apperrors.Unauthorized("invalid API key")
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	// Check expiration
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, apperrors.Unauthorized("API key expired")
	}

	// Verify secret key
	if !s.verifySecretKey(secretKey, apiKey.SecretKeyHash) {
		return nil, apperrors.Unauthorized("invalid API key")
	}

	// Update last used (async, don't fail on error)
	go func() {
		_ = s.apiKeyRepo.UpdateLastUsed(context.Background(), apiKey.ID)
	}()

	return &apiKey.ProjectID, nil
}

// ValidateAPIKeyPublicOnly validates an API key by public key only (for read operations)
func (s *AuthService) ValidateAPIKeyPublicOnly(ctx context.Context, publicKey string) (*uuid.UUID, error) {
	projectID, err := s.apiKeyRepo.GetProjectIDByPublicKey(ctx, publicKey)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, apperrors.Unauthorized("invalid API key")
		}
		return nil, fmt.Errorf("failed to get project ID: %w", err)
	}

	return projectID, nil
}

// CreateAPIKey creates a new API key
func (s *AuthService) CreateAPIKey(ctx context.Context, projectID uuid.UUID, input *domain.APIKeyInput, userID uuid.UUID) (*domain.APIKeyCreateResult, error) {
	return s.CreateAPIKeyWithContext(ctx, projectID, input, userID, "")
}

// CreateAPIKeyWithContext creates a new API key with audit logging context
func (s *AuthService) CreateAPIKeyWithContext(ctx context.Context, projectID uuid.UUID, input *domain.APIKeyInput, userID uuid.UUID, userEmail string) (*domain.APIKeyCreateResult, error) {
	// Generate keys
	publicKey, secretKey, err := s.generateAPIKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Hash secret key
	secretKeyHash := s.hashSecretKey(secretKey)

	// Get preview (last 4 characters)
	secretKeyPreview := secretKey[len(secretKey)-4:]

	now := time.Now()

	// Set default expiration to 1 year if not specified
	// This prevents permanent API keys which are a security risk if leaked
	expiresAt := input.ExpiresAt
	if expiresAt == nil {
		defaultExpiry := now.AddDate(1, 0, 0) // 1 year from now
		expiresAt = &defaultExpiry
	}

	apiKey := &domain.APIKey{
		ID:               uuid.New(),
		ProjectID:        projectID,
		Name:             input.Name,
		PublicKey:        publicKey,
		SecretKeyHash:    secretKeyHash,
		SecretKeyPreview: secretKeyPreview,
		Scopes:           input.Scopes,
		ExpiresAt:        expiresAt,
		CreatedBy:        &userID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if len(apiKey.Scopes) == 0 {
		apiKey.Scopes = domain.DefaultScopes()
	}

	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	// Audit log: API key created
	if s.auditLogger != nil && userID != uuid.Nil {
		// Get organization ID from project
		project, err := s.projectRepo.GetByID(ctx, projectID)
		if err == nil && project != nil {
			go func() {
				_ = s.auditLogger.LogAPIKeyCreated(context.Background(), project.OrganizationID, userID, userEmail, apiKey.ID, apiKey.Name)
			}()
		}
	}

	return &domain.APIKeyCreateResult{
		APIKey:    apiKey,
		SecretKey: secretKey,
	}, nil
}

// DeleteAPIKey deletes an API key
func (s *AuthService) DeleteAPIKey(ctx context.Context, id uuid.UUID) error {
	return s.DeleteAPIKeyWithContext(ctx, id, uuid.Nil, "")
}

// DeleteAPIKeyWithContext deletes an API key with audit logging context
func (s *AuthService) DeleteAPIKeyWithContext(ctx context.Context, id uuid.UUID, actorID uuid.UUID, actorEmail string) error {
	// Get API key details before deletion for audit logging
	var apiKey *domain.APIKey
	var orgID uuid.UUID
	if s.auditLogger != nil && actorID != uuid.Nil {
		var err error
		apiKey, err = s.apiKeyRepo.GetByID(ctx, id)
		if err == nil && apiKey != nil {
			// Get organization ID from project
			project, err := s.projectRepo.GetByID(ctx, apiKey.ProjectID)
			if err == nil && project != nil {
				orgID = project.OrganizationID
			}
		}
	}

	// Delete the API key
	if err := s.apiKeyRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Audit log: API key revoked
	if s.auditLogger != nil && actorID != uuid.Nil && apiKey != nil && orgID != uuid.Nil {
		go func() {
			_ = s.auditLogger.LogAPIKeyRevoked(context.Background(), orgID, actorID, actorEmail, apiKey.ID, apiKey.Name)
		}()
	}

	return nil
}

// ListAPIKeys lists API keys for a project
func (s *AuthService) ListAPIKeys(ctx context.Context, projectID uuid.UUID) ([]domain.APIKey, error) {
	return s.apiKeyRepo.ListByProjectID(ctx, projectID)
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// CheckProjectAccess checks if a user has access to a project
func (s *AuthService) CheckProjectAccess(ctx context.Context, projectID, userID uuid.UUID, requiredRole domain.OrgRole) error {
	role, err := s.projectRepo.GetUserRoleForProject(ctx, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	if role == nil {
		return apperrors.Forbidden("no access to project")
	}

	// Check role hierarchy
	if !s.hasRequiredRole(*role, requiredRole) {
		return apperrors.Forbidden("insufficient permissions")
	}

	return nil
}

// hasRequiredRole checks if the user's role meets the required level
func (s *AuthService) hasRequiredRole(userRole, requiredRole domain.OrgRole) bool {
	roleLevel := map[domain.OrgRole]int{
		domain.OrgRoleViewer: 1,
		domain.OrgRoleMember: 2,
		domain.OrgRoleAdmin:  3,
		domain.OrgRoleOwner:  4,
	}

	return roleLevel[userRole] >= roleLevel[requiredRole]
}

// generateAccessToken generates a JWT access token
func (s *AuthService) generateAccessToken(user *domain.User) (string, error) {
	claims := &domain.JWTClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.JWT.AccessExpiry) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.cfg.JWT.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

// generateRefreshToken generates a random refresh token
func (s *AuthService) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// generateAPIKeyPair generates a public/secret key pair
func (s *AuthService) generateAPIKeyPair() (publicKey, secretKey string, err error) {
	// Public key: pk-xxxx
	pubBytes := make([]byte, 16)
	if _, err := rand.Read(pubBytes); err != nil {
		return "", "", err
	}
	publicKey = "pk-" + hex.EncodeToString(pubBytes)

	// Secret key: sk-xxxx
	secBytes := make([]byte, 32)
	if _, err := rand.Read(secBytes); err != nil {
		return "", "", err
	}
	secretKey = "sk-" + hex.EncodeToString(secBytes)

	return publicKey, secretKey, nil
}

// hashSecretKey creates a bcrypt hash of the secret key
func (s *AuthService) hashSecretKey(secretKey string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(secretKey), bcrypt.DefaultCost)
	if err != nil {
		// This should never happen with valid input, but fall back to empty string
		// which will fail verification
		return ""
	}
	return string(hash)
}

// verifySecretKey verifies a secret key against its bcrypt hash
func (s *AuthService) verifySecretKey(secretKey, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(secretKey))
	return err == nil
}

// HandleOAuthCallback handles OAuth authentication callback
func (s *AuthService) HandleOAuthCallback(ctx context.Context, input *domain.OAuthCallbackInput) (*domain.AuthResult, error) {
	return s.HandleOAuthCallbackWithContext(ctx, input, "", "")
}

// HandleOAuthCallbackWithContext handles OAuth authentication callback with audit logging context
func (s *AuthService) HandleOAuthCallbackWithContext(ctx context.Context, input *domain.OAuthCallbackInput, ipAddress, userAgent string) (*domain.AuthResult, error) {
	// Check if account exists
	account, err := s.userRepo.GetAccountByProvider(ctx, input.Provider, input.ProviderAccountID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	var user *domain.User
	now := time.Now()

	if account != nil {
		// Existing account - get user
		user, err = s.userRepo.GetByID(ctx, account.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		// Update account tokens
		account.AccessToken = input.AccessToken
		account.RefreshToken = input.RefreshToken
		account.ExpiresAt = input.ExpiresAt
		account.UpdatedAt = now

		if err := s.userRepo.CreateAccount(ctx, account); err != nil {
			return nil, fmt.Errorf("failed to update account: %w", err)
		}
	} else {
		// New account - check if user exists by email
		user, err = s.userRepo.GetByEmail(ctx, input.Email)
		if err != nil {
			if !apperrors.IsNotFound(err) {
				return nil, fmt.Errorf("failed to get user: %w", err)
			}

			// Create new user
			user = &domain.User{
				ID:            uuid.New(),
				Email:         input.Email,
				EmailVerified: true,
				Name:          input.Name,
				Image:         input.Image,
				CreatedAt:     now,
				UpdatedAt:     now,
			}

			if err := s.userRepo.Create(ctx, user); err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}

			// Create default organization
			org := &domain.Organization{
				ID:        uuid.New(),
				Name:      input.Name + "'s Organization",
				Slug:      domain.GenerateSlug(input.Name + "'s Organization"),
				CreatedAt: now,
				UpdatedAt: now,
			}

			if err := s.orgRepo.Create(ctx, org); err != nil {
				return nil, fmt.Errorf("failed to create organization: %w", err)
			}

			// Add user as owner
			member := &domain.OrganizationMember{
				ID:             uuid.New(),
				OrganizationID: org.ID,
				UserID:         user.ID,
				Role:           domain.OrgRoleOwner,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := s.orgRepo.AddMember(ctx, member); err != nil {
				return nil, fmt.Errorf("failed to add member: %w", err)
			}
		}

		// Create account link
		account = &domain.Account{
			ID:                uuid.New(),
			UserID:            user.ID,
			Provider:          input.Provider,
			ProviderAccountID: input.ProviderAccountID,
			AccessToken:       input.AccessToken,
			RefreshToken:      input.RefreshToken,
			ExpiresAt:         input.ExpiresAt,
			TokenType:         input.TokenType,
			Scope:             input.Scope,
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		if err := s.userRepo.CreateAccount(ctx, account); err != nil {
			return nil, fmt.Errorf("failed to create account: %w", err)
		}
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token as session
	session := &domain.UserSession{
		ID:           uuid.New(),
		SessionToken: refreshToken,
		UserID:       user.ID,
		ExpiresAt:    time.Now().Add(time.Duration(s.cfg.JWT.RefreshExpiry) * time.Hour),
		CreatedAt:    now,
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Audit log: SSO login
	if s.auditLogger != nil {
		// Get user's primary organization for audit logging
		orgs, err := s.orgRepo.ListByUserID(ctx, user.ID)
		if err == nil && len(orgs) > 0 {
			go func() {
				_ = s.auditLogger.LogSSOLogin(context.Background(), orgs[0].ID, user.ID, user.Email, input.Provider, ipAddress, userAgent)
			}()
		}
	}

	return &domain.AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(s.cfg.JWT.AccessExpiry) * time.Minute),
	}, nil
}
