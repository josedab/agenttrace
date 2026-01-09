package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	EmailVerified bool       `json:"emailVerified"`
	Name          string     `json:"name,omitempty"`
	Image         string     `json:"image,omitempty"`
	PasswordHash  string     `json:"-"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`

	// Related data (populated by resolver)
	Organizations []OrganizationMember `json:"organizations,omitempty"`
}

// UserInput represents input for creating/updating a user
type UserInput struct {
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password,omitempty" validate:"omitempty,min=8"`
	Name     *string `json:"name,omitempty"`
	Image    *string `json:"image,omitempty"`
}

// UserUpdateInput represents input for updating a user
type UserUpdateInput struct {
	Name  *string `json:"name,omitempty"`
	Image *string `json:"image,omitempty"`
}

// Account represents an OAuth account linked to a user
type Account struct {
	ID                uuid.UUID  `json:"id"`
	UserID            uuid.UUID  `json:"userId"`
	Provider          string     `json:"provider"`
	ProviderAccountID string     `json:"providerAccountId"`
	AccessToken       *string    `json:"-"`
	RefreshToken      *string    `json:"-"`
	ExpiresAt         *time.Time `json:"expiresAt,omitempty"`
	TokenType         *string    `json:"tokenType,omitempty"`
	Scope             *string    `json:"scope,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

// UserSession represents a user authentication session
type UserSession struct {
	ID           uuid.UUID `json:"id"`
	SessionToken string    `json:"sessionToken"`
	UserID       uuid.UUID `json:"userId"`
	ExpiresAt    time.Time `json:"expiresAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

// AuthResult represents the result of an authentication attempt
type AuthResult struct {
	User         *User     `json:"user"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken,omitempty"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// EmailVerificationToken represents an email verification token
type EmailVerificationToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// LoginInput represents input for login
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RegisterInput represents input for registration
type RegisterInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name,omitempty"`
}

// OAuthCallbackInput represents input from OAuth callback
type OAuthCallbackInput struct {
	Provider          string     `json:"provider"`
	ProviderAccountID string     `json:"providerAccountId"`
	Email             string     `json:"email"`
	Name              string     `json:"name,omitempty"`
	Image             string     `json:"image,omitempty"`
	AccessToken       *string    `json:"accessToken"`
	RefreshToken      *string    `json:"refreshToken,omitempty"`
	ExpiresAt         *time.Time `json:"expiresAt,omitempty"`
	TokenType         *string    `json:"tokenType,omitempty"`
	Scope             *string    `json:"scope,omitempty"`
}
