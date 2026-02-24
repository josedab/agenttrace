package domain

import (
	"time"

	"github.com/google/uuid"
)

// EmbedConfig represents the configuration for a white-label embed
type EmbedConfig struct {
	ID             uuid.UUID     `json:"id"`
	ProjectID      uuid.UUID     `json:"projectId"`
	Theme          EmbedTheme    `json:"theme"`
	AllowedOrigins []string      `json:"allowedOrigins"`
	Features       EmbedFeatures `json:"features"`
	APIKeyID       uuid.UUID     `json:"apiKeyId"`
	Enabled        bool          `json:"enabled"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

// EmbedTheme represents theming options for the embed
type EmbedTheme struct {
	PrimaryColor    string `json:"primaryColor,omitempty"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	TextColor       string `json:"textColor,omitempty"`
	FontFamily      string `json:"fontFamily,omitempty"`
	LogoURL         string `json:"logoUrl,omitempty"`
	HideBranding    bool   `json:"hideBranding"`
	BorderRadius    string `json:"borderRadius,omitempty"`
}

// EmbedFeatures represents which features are enabled in the embed
type EmbedFeatures struct {
	TraceViewer   bool `json:"traceViewer"`
	CostDashboard bool `json:"costDashboard"`
	QualityScore  bool `json:"qualityScore"`
	ActivityFeed  bool `json:"activityFeed"`
}

// EmbedToken represents an embed access token
type EmbedToken struct {
	Token     string    `json:"token"`
	ConfigID  uuid.UUID `json:"configId"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// EmbedConfigInput represents the input for creating or updating an embed config
type EmbedConfigInput struct {
	Theme          EmbedTheme    `json:"theme"`
	AllowedOrigins []string      `json:"allowedOrigins"`
	Features       EmbedFeatures `json:"features"`
}
