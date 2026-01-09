package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// UserRepository handles user data operations in PostgreSQL
type UserRepository struct {
	db *database.PostgresDB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.PostgresDB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, email_verified, name, image, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.EmailVerified,
		user.Name,
		user.Image,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, email_verified, name, image, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.EmailVerified,
		&user.Name,
		&user.Image,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("user")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, email_verified, name, image, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.EmailVerified,
		&user.Name,
		&user.Image,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("user")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $2, email_verified = $3, name = $4, image = $5, password_hash = $6, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.EmailVerified,
		user.Name,
		user.Image,
		user.PasswordHash,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// CreateAccount creates a new OAuth account
func (r *UserRepository) CreateAccount(ctx context.Context, account *domain.Account) error {
	query := `
		INSERT INTO accounts (id, user_id, provider, provider_account_id, access_token, refresh_token, expires_at, token_type, scope, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (provider, provider_account_id)
		DO UPDATE SET access_token = $5, refresh_token = $6, expires_at = $7, updated_at = NOW()
	`

	_, err := r.db.Pool.Exec(ctx, query,
		account.ID,
		account.UserID,
		account.Provider,
		account.ProviderAccountID,
		account.AccessToken,
		account.RefreshToken,
		account.ExpiresAt,
		account.TokenType,
		account.Scope,
		account.CreatedAt,
		account.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

// GetAccountByProvider retrieves an OAuth account by provider
func (r *UserRepository) GetAccountByProvider(ctx context.Context, provider, providerAccountID string) (*domain.Account, error) {
	query := `
		SELECT id, user_id, provider, provider_account_id, access_token, refresh_token, expires_at, token_type, scope, created_at, updated_at
		FROM accounts
		WHERE provider = $1 AND provider_account_id = $2
	`

	var account domain.Account
	err := r.db.Pool.QueryRow(ctx, query, provider, providerAccountID).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderAccountID,
		&account.AccessToken,
		&account.RefreshToken,
		&account.ExpiresAt,
		&account.TokenType,
		&account.Scope,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("account")
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

// CreateSession creates a new session
func (r *UserRepository) CreateSession(ctx context.Context, session *domain.UserSession) error {
	query := `
		INSERT INTO sessions (id, session_token, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		session.ID,
		session.SessionToken,
		session.UserID,
		session.ExpiresAt,
		session.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSessionByToken retrieves a session by token
func (r *UserRepository) GetSessionByToken(ctx context.Context, token string) (*domain.UserSession, error) {
	query := `
		SELECT id, session_token, user_id, expires_at, created_at
		FROM sessions
		WHERE session_token = $1 AND expires_at > NOW()
	`

	var session domain.UserSession
	err := r.db.Pool.QueryRow(ctx, query, token).Scan(
		&session.ID,
		&session.SessionToken,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("session")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// DeleteSession deletes a session
func (r *UserRepository) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE session_token = $1`

	_, err := r.db.Pool.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteExpiredSessions deletes expired sessions
func (r *UserRepository) DeleteExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at <= NOW()`

	_, err := r.db.Pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}

// ExistsByEmail checks if a user exists by email
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}
