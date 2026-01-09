package domain

import (
	"time"

	"github.com/google/uuid"
)

// SSOProvider represents supported SSO providers
type SSOProvider string

const (
	SSOProviderSAML   SSOProvider = "saml"
	SSOProviderOIDC   SSOProvider = "oidc"
	SSOProviderOkta   SSOProvider = "okta"
	SSOProviderAzureAD SSOProvider = "azure_ad"
	SSOProviderGoogle SSOProvider = "google"
)

// SSOConfiguration represents an organization's SSO configuration
type SSOConfiguration struct {
	ID             uuid.UUID   `json:"id" db:"id"`
	OrganizationID uuid.UUID   `json:"organizationId" db:"organization_id"`
	Provider       SSOProvider `json:"provider" db:"provider"`
	Enabled        bool        `json:"enabled" db:"enabled"`
	EnforceSSO     bool        `json:"enforceSSO" db:"enforce_sso"` // Require SSO for all users
	AllowedDomains []string    `json:"allowedDomains" db:"allowed_domains"`

	// SAML Configuration
	SAMLEntityID         string `json:"samlEntityId,omitempty" db:"saml_entity_id"`
	SAMLSSOUrl           string `json:"samlSsoUrl,omitempty" db:"saml_sso_url"`
	SAMLSLOUrl           string `json:"samlSloUrl,omitempty" db:"saml_slo_url"`
	SAMLCertificate      string `json:"samlCertificate,omitempty" db:"saml_certificate"`
	SAMLSignRequests     bool   `json:"samlSignRequests" db:"saml_sign_requests"`
	SAMLNameIDFormat     string `json:"samlNameIdFormat,omitempty" db:"saml_name_id_format"`

	// OIDC Configuration
	OIDCClientID     string   `json:"oidcClientId,omitempty" db:"oidc_client_id"`
	OIDCClientSecret string   `json:"-" db:"oidc_client_secret"` // Never exposed in JSON
	OIDCIssuerURL    string   `json:"oidcIssuerUrl,omitempty" db:"oidc_issuer_url"`
	OIDCScopes       []string `json:"oidcScopes,omitempty" db:"oidc_scopes"`

	// Attribute Mapping
	AttributeMapping SSOAttributeMapping `json:"attributeMapping" db:"attribute_mapping"`

	// Auto-provisioning
	AutoProvisionUsers  bool `json:"autoProvisionUsers" db:"auto_provision_users"`
	DefaultRole         string `json:"defaultRole" db:"default_role"`
	AutoAssignProjects  []uuid.UUID `json:"autoAssignProjects,omitempty" db:"auto_assign_projects"`

	// Metadata
	MetadataURL  string     `json:"metadataUrl,omitempty" db:"metadata_url"`
	LastSyncAt   *time.Time `json:"lastSyncAt,omitempty" db:"last_sync_at"`
	LastErrorAt  *time.Time `json:"lastErrorAt,omitempty" db:"last_error_at"`
	LastError    *string    `json:"lastError,omitempty" db:"last_error"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// SSOAttributeMapping maps IdP attributes to AgentTrace user fields
type SSOAttributeMapping struct {
	Email      string `json:"email"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	DisplayName string `json:"displayName"`
	Groups     string `json:"groups"`
	Department string `json:"department"`
}

// SSOSession represents an active SSO session
type SSOSession struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	UserID           uuid.UUID  `json:"userId" db:"user_id"`
	OrganizationID   uuid.UUID  `json:"organizationId" db:"organization_id"`
	Provider         SSOProvider `json:"provider" db:"provider"`
	ExternalID       string     `json:"externalId" db:"external_id"` // IdP user ID
	SessionIndex     string     `json:"sessionIndex,omitempty" db:"session_index"` // SAML session index
	AccessToken      string     `json:"-" db:"access_token"` // For OIDC
	RefreshToken     string     `json:"-" db:"refresh_token"`
	IDToken          string     `json:"-" db:"id_token"`
	ExpiresAt        time.Time  `json:"expiresAt" db:"expires_at"`
	LastActivityAt   time.Time  `json:"lastActivityAt" db:"last_activity_at"`
	IPAddress        string     `json:"ipAddress" db:"ip_address"`
	UserAgent        string     `json:"userAgent" db:"user_agent"`
	CreatedAt        time.Time  `json:"createdAt" db:"created_at"`
}

// SSOConfigurationInput represents input for creating/updating SSO configuration
type SSOConfigurationInput struct {
	Provider        SSOProvider         `json:"provider" validate:"required"`
	Enabled         bool                `json:"enabled"`
	EnforceSSO      bool                `json:"enforceSSO"`
	AllowedDomains  []string            `json:"allowedDomains"`

	// SAML
	SAMLEntityID     *string `json:"samlEntityId"`
	SAMLSSOUrl       *string `json:"samlSsoUrl"`
	SAMLSLOUrl       *string `json:"samlSloUrl"`
	SAMLCertificate  *string `json:"samlCertificate"`
	SAMLSignRequests *bool   `json:"samlSignRequests"`

	// OIDC
	OIDCClientID     *string  `json:"oidcClientId"`
	OIDCClientSecret *string  `json:"oidcClientSecret"`
	OIDCIssuerURL    *string  `json:"oidcIssuerUrl"`
	OIDCScopes       []string `json:"oidcScopes"`

	// Attribute mapping
	AttributeMapping *SSOAttributeMapping `json:"attributeMapping"`

	// Provisioning
	AutoProvisionUsers *bool       `json:"autoProvisionUsers"`
	DefaultRole        *string     `json:"defaultRole"`
	AutoAssignProjects []uuid.UUID `json:"autoAssignProjects"`

	MetadataURL *string `json:"metadataUrl"`
}

// SSOLoginRequest represents a request to initiate SSO login
type SSOLoginRequest struct {
	OrganizationID uuid.UUID `json:"organizationId"`
	ReturnURL      string    `json:"returnUrl"`
	State          string    `json:"state"`
}

// SSOCallbackRequest represents the callback from IdP
type SSOCallbackRequest struct {
	SAMLResponse string `json:"SAMLResponse,omitempty"`
	RelayState   string `json:"RelayState,omitempty"`
	Code         string `json:"code,omitempty"` // OIDC
	State        string `json:"state,omitempty"`
	Error        string `json:"error,omitempty"`
	ErrorDesc    string `json:"error_description,omitempty"`
}

// SSOUserInfo represents user info from IdP
type SSOUserInfo struct {
	ExternalID  string   `json:"externalId"`
	Email       string   `json:"email"`
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	DisplayName string   `json:"displayName"`
	Groups      []string `json:"groups"`
	Department  string   `json:"department"`
	Attributes  map[string]any `json:"attributes"`
}
