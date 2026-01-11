package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

type SSORepository struct {
	db *sqlx.DB
}

func NewSSORepository(db *sqlx.DB) *SSORepository {
	return &SSORepository{db: db}
}

// Configuration methods

func (r *SSORepository) CreateConfiguration(ctx context.Context, config *domain.SSOConfiguration) error {
	attributeMappingJSON, err := json.Marshal(config.AttributeMapping)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO sso_configurations (
			id, organization_id, provider, enabled, enforce_sso, allowed_domains,
			saml_entity_id, saml_sso_url, saml_slo_url, saml_certificate,
			saml_sign_requests, saml_name_id_format,
			oidc_client_id, oidc_client_secret, oidc_issuer_url, oidc_scopes,
			attribute_mapping, auto_provision_users, default_role, auto_assign_projects,
			metadata_url, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23
		)`

	_, err = r.db.ExecContext(ctx, query,
		config.ID, config.OrganizationID, config.Provider, config.Enabled, config.EnforceSSO,
		pq.Array(config.AllowedDomains),
		config.SAMLEntityID, config.SAMLSSOUrl, config.SAMLSLOUrl, config.SAMLCertificate,
		config.SAMLSignRequests, config.SAMLNameIDFormat,
		config.OIDCClientID, config.OIDCClientSecret, config.OIDCIssuerURL,
		pq.Array(config.OIDCScopes),
		attributeMappingJSON, config.AutoProvisionUsers, config.DefaultRole,
		pq.Array(config.AutoAssignProjects),
		config.MetadataURL, config.CreatedAt, config.UpdatedAt,
	)
	return err
}

func (r *SSORepository) GetConfigurationByOrganization(ctx context.Context, orgID uuid.UUID) (*domain.SSOConfiguration, error) {
	query := `
		SELECT id, organization_id, provider, enabled, enforce_sso, allowed_domains,
			saml_entity_id, saml_sso_url, saml_slo_url, saml_certificate,
			saml_sign_requests, saml_name_id_format,
			oidc_client_id, oidc_client_secret, oidc_issuer_url, oidc_scopes,
			attribute_mapping, auto_provision_users, default_role, auto_assign_projects,
			metadata_url, last_sync_at, last_error_at, last_error, created_at, updated_at
		FROM sso_configurations
		WHERE organization_id = $1`

	var config domain.SSOConfiguration
	var allowedDomains, oidcScopes pq.StringArray
	var autoAssignProjects pq.StringArray
	var attributeMappingJSON []byte

	err := r.db.QueryRowContext(ctx, query, orgID).Scan(
		&config.ID, &config.OrganizationID, &config.Provider, &config.Enabled,
		&config.EnforceSSO, &allowedDomains,
		&config.SAMLEntityID, &config.SAMLSSOUrl, &config.SAMLSLOUrl, &config.SAMLCertificate,
		&config.SAMLSignRequests, &config.SAMLNameIDFormat,
		&config.OIDCClientID, &config.OIDCClientSecret, &config.OIDCIssuerURL, &oidcScopes,
		&attributeMappingJSON, &config.AutoProvisionUsers, &config.DefaultRole, &autoAssignProjects,
		&config.MetadataURL, &config.LastSyncAt, &config.LastErrorAt, &config.LastError,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	config.AllowedDomains = []string(allowedDomains)
	config.OIDCScopes = []string(oidcScopes)

	// Parse auto-assign projects
	config.AutoAssignProjects = make([]uuid.UUID, 0, len(autoAssignProjects))
	for _, s := range autoAssignProjects {
		if id, err := uuid.Parse(s); err == nil {
			config.AutoAssignProjects = append(config.AutoAssignProjects, id)
		}
	}

	if err := json.Unmarshal(attributeMappingJSON, &config.AttributeMapping); err != nil {
		return nil, err
	}

	return &config, nil
}

func (r *SSORepository) UpdateConfiguration(ctx context.Context, config *domain.SSOConfiguration) error {
	attributeMappingJSON, err := json.Marshal(config.AttributeMapping)
	if err != nil {
		return err
	}

	query := `
		UPDATE sso_configurations SET
			provider = $2, enabled = $3, enforce_sso = $4, allowed_domains = $5,
			saml_entity_id = $6, saml_sso_url = $7, saml_slo_url = $8, saml_certificate = $9,
			saml_sign_requests = $10, saml_name_id_format = $11,
			oidc_client_id = $12, oidc_client_secret = $13, oidc_issuer_url = $14, oidc_scopes = $15,
			attribute_mapping = $16, auto_provision_users = $17, default_role = $18,
			auto_assign_projects = $19, metadata_url = $20, updated_at = $21
		WHERE id = $1`

	_, err = r.db.ExecContext(ctx, query,
		config.ID, config.Provider, config.Enabled, config.EnforceSSO,
		pq.Array(config.AllowedDomains),
		config.SAMLEntityID, config.SAMLSSOUrl, config.SAMLSLOUrl, config.SAMLCertificate,
		config.SAMLSignRequests, config.SAMLNameIDFormat,
		config.OIDCClientID, config.OIDCClientSecret, config.OIDCIssuerURL,
		pq.Array(config.OIDCScopes),
		attributeMappingJSON, config.AutoProvisionUsers, config.DefaultRole,
		pq.Array(config.AutoAssignProjects),
		config.MetadataURL, time.Now(),
	)
	return err
}

func (r *SSORepository) DeleteConfiguration(ctx context.Context, orgID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sso_configurations WHERE organization_id = $1", orgID)
	return err
}

func (r *SSORepository) UpdateLastSync(ctx context.Context, configID uuid.UUID, syncTime time.Time) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE sso_configurations SET last_sync_at = $2, updated_at = NOW() WHERE id = $1",
		configID, syncTime)
	return err
}

func (r *SSORepository) UpdateLastError(ctx context.Context, configID uuid.UUID, errorMsg string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE sso_configurations SET last_error_at = NOW(), last_error = $2, updated_at = NOW() WHERE id = $1",
		configID, errorMsg)
	return err
}

// Session methods

func (r *SSORepository) CreateSession(ctx context.Context, session *domain.SSOSession) error {
	query := `
		INSERT INTO sso_sessions (
			id, user_id, organization_id, provider, external_id, session_index,
			access_token, refresh_token, id_token, expires_at, last_activity_at,
			ip_address, user_agent, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.OrganizationID, session.Provider,
		session.ExternalID, session.SessionIndex,
		session.AccessToken, session.RefreshToken, session.IDToken,
		session.ExpiresAt, session.LastActivityAt,
		session.IPAddress, session.UserAgent, session.CreatedAt,
	)
	return err
}

func (r *SSORepository) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.SSOSession, error) {
	query := `
		SELECT id, user_id, organization_id, provider, external_id, session_index,
			access_token, refresh_token, id_token, expires_at, last_activity_at,
			ip_address, user_agent, created_at
		FROM sso_sessions
		WHERE id = $1`

	var session domain.SSOSession
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID, &session.UserID, &session.OrganizationID, &session.Provider,
		&session.ExternalID, &session.SessionIndex,
		&session.AccessToken, &session.RefreshToken, &session.IDToken,
		&session.ExpiresAt, &session.LastActivityAt,
		&session.IPAddress, &session.UserAgent, &session.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &session, err
}

func (r *SSORepository) GetSessionsByUser(ctx context.Context, userID uuid.UUID) ([]domain.SSOSession, error) {
	query := `
		SELECT id, user_id, organization_id, provider, external_id, session_index,
			access_token, refresh_token, id_token, expires_at, last_activity_at,
			ip_address, user_agent, created_at
		FROM sso_sessions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC`

	var sessions []domain.SSOSession
	err := r.db.SelectContext(ctx, &sessions, query, userID)
	return sessions, err
}

func (r *SSORepository) UpdateSessionActivity(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE sso_sessions SET last_activity_at = NOW() WHERE id = $1",
		sessionID)
	return err
}

func (r *SSORepository) UpdateSessionTokens(ctx context.Context, sessionID uuid.UUID, accessToken, refreshToken, idToken string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sso_sessions SET
			access_token = $2, refresh_token = $3, id_token = $4,
			expires_at = $5, last_activity_at = NOW()
		WHERE id = $1`,
		sessionID, accessToken, refreshToken, idToken, expiresAt)
	return err
}

func (r *SSORepository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sso_sessions WHERE id = $1", sessionID)
	return err
}

func (r *SSORepository) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sso_sessions WHERE user_id = $1", userID)
	return err
}

func (r *SSORepository) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM sso_sessions WHERE expires_at < NOW()")
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// State methods (for OAuth flow)

func (r *SSORepository) CreateState(ctx context.Context, orgID uuid.UUID, state, returnURL, nonce, codeVerifier string, expiresAt time.Time) error {
	query := `
		INSERT INTO sso_states (organization_id, state, return_url, nonce, code_verifier, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, orgID, state, returnURL, nonce, codeVerifier, expiresAt)
	return err
}

type SSOState struct {
	ID             uuid.UUID `db:"id"`
	OrganizationID uuid.UUID `db:"organization_id"`
	State          string    `db:"state"`
	ReturnURL      string    `db:"return_url"`
	Nonce          string    `db:"nonce"`
	CodeVerifier   string    `db:"code_verifier"`
	ExpiresAt      time.Time `db:"expires_at"`
	CreatedAt      time.Time `db:"created_at"`
}

func (r *SSORepository) GetAndDeleteState(ctx context.Context, state string) (*SSOState, error) {
	query := `
		DELETE FROM sso_states
		WHERE state = $1 AND expires_at > NOW()
		RETURNING id, organization_id, state, return_url, nonce, code_verifier, expires_at, created_at`

	var s SSOState
	err := r.db.QueryRowContext(ctx, query, state).Scan(
		&s.ID, &s.OrganizationID, &s.State, &s.ReturnURL, &s.Nonce, &s.CodeVerifier, &s.ExpiresAt, &s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *SSORepository) CleanupExpiredStates(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM sso_states WHERE expires_at < NOW()")
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Identity Mapping methods

type SSOIdentityMapping struct {
	ID             uuid.UUID      `db:"id"`
	UserID         uuid.UUID      `db:"user_id"`
	OrganizationID uuid.UUID      `db:"organization_id"`
	Provider       string         `db:"provider"`
	ExternalID     string         `db:"external_id"`
	ExternalEmail  string         `db:"external_email"`
	ExternalName   string         `db:"external_name"`
	Attributes     map[string]any `db:"attributes"`
	LinkedAt       time.Time      `db:"linked_at"`
	LastLoginAt    *time.Time     `db:"last_login_at"`
}

func (r *SSORepository) CreateIdentityMapping(ctx context.Context, mapping *SSOIdentityMapping) error {
	attributesJSON, err := json.Marshal(mapping.Attributes)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO sso_identity_mappings (
			user_id, organization_id, provider, external_id, external_email,
			external_name, attributes, linked_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		mapping.UserID, mapping.OrganizationID, mapping.Provider, mapping.ExternalID,
		mapping.ExternalEmail, mapping.ExternalName, attributesJSON, mapping.LinkedAt,
	).Scan(&mapping.ID)
}

func (r *SSORepository) GetIdentityMapping(ctx context.Context, orgID uuid.UUID, provider, externalID string) (*SSOIdentityMapping, error) {
	query := `
		SELECT id, user_id, organization_id, provider, external_id, external_email,
			external_name, attributes, linked_at, last_login_at
		FROM sso_identity_mappings
		WHERE organization_id = $1 AND provider = $2 AND external_id = $3`

	var mapping SSOIdentityMapping
	var attributesJSON []byte

	err := r.db.QueryRowContext(ctx, query, orgID, provider, externalID).Scan(
		&mapping.ID, &mapping.UserID, &mapping.OrganizationID, &mapping.Provider,
		&mapping.ExternalID, &mapping.ExternalEmail, &mapping.ExternalName,
		&attributesJSON, &mapping.LinkedAt, &mapping.LastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(attributesJSON, &mapping.Attributes); err != nil {
		return nil, err
	}

	return &mapping, nil
}

func (r *SSORepository) GetIdentityMappingsByUser(ctx context.Context, userID uuid.UUID) ([]SSOIdentityMapping, error) {
	query := `
		SELECT id, user_id, organization_id, provider, external_id, external_email,
			external_name, attributes, linked_at, last_login_at
		FROM sso_identity_mappings
		WHERE user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mappings []SSOIdentityMapping
	for rows.Next() {
		var mapping SSOIdentityMapping
		var attributesJSON []byte

		if err := rows.Scan(
			&mapping.ID, &mapping.UserID, &mapping.OrganizationID, &mapping.Provider,
			&mapping.ExternalID, &mapping.ExternalEmail, &mapping.ExternalName,
			&attributesJSON, &mapping.LinkedAt, &mapping.LastLoginAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(attributesJSON, &mapping.Attributes); err != nil {
			return nil, err
		}

		mappings = append(mappings, mapping)
	}

	return mappings, rows.Err()
}

func (r *SSORepository) UpdateIdentityMappingLastLogin(ctx context.Context, mappingID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE sso_identity_mappings SET last_login_at = NOW() WHERE id = $1",
		mappingID)
	return err
}

func (r *SSORepository) DeleteIdentityMapping(ctx context.Context, mappingID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sso_identity_mappings WHERE id = $1", mappingID)
	return err
}
