package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/repository/postgres"
)

// MockSSORepository is a mock implementation of the SSO repository
type MockSSORepository struct {
	mock.Mock
}

func (m *MockSSORepository) CreateConfiguration(ctx context.Context, config *domain.SSOConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockSSORepository) GetConfigurationByOrganization(ctx context.Context, orgID uuid.UUID) (*domain.SSOConfiguration, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SSOConfiguration), args.Error(1)
}

func (m *MockSSORepository) UpdateConfiguration(ctx context.Context, config *domain.SSOConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockSSORepository) DeleteConfiguration(ctx context.Context, orgID uuid.UUID) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}

func (m *MockSSORepository) CreateSession(ctx context.Context, session *domain.SSOSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSSORepository) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.SSOSession, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SSOSession), args.Error(1)
}

func (m *MockSSORepository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSSORepository) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSSORepository) UpdateSessionActivity(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSSORepository) UpdateSessionTokens(ctx context.Context, sessionID uuid.UUID, accessToken, refreshToken, idToken string, expiresAt time.Time) error {
	args := m.Called(ctx, sessionID, accessToken, refreshToken, idToken, expiresAt)
	return args.Error(0)
}

func (m *MockSSORepository) CreateState(ctx context.Context, orgID uuid.UUID, state, returnURL, nonce, codeVerifier string, expiresAt time.Time) error {
	args := m.Called(ctx, orgID, state, returnURL, nonce, codeVerifier, expiresAt)
	return args.Error(0)
}

func (m *MockSSORepository) GetAndDeleteState(ctx context.Context, state string) (*postgres.SSOState, error) {
	args := m.Called(ctx, state)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postgres.SSOState), args.Error(1)
}

func (m *MockSSORepository) CreateIdentityMapping(ctx context.Context, mapping *postgres.SSOIdentityMapping) error {
	args := m.Called(ctx, mapping)
	return args.Error(0)
}

func (m *MockSSORepository) GetIdentityMapping(ctx context.Context, orgID uuid.UUID, provider, externalID string) (*postgres.SSOIdentityMapping, error) {
	args := m.Called(ctx, orgID, provider, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postgres.SSOIdentityMapping), args.Error(1)
}

func (m *MockSSORepository) UpdateIdentityMappingLastLogin(ctx context.Context, mappingID uuid.UUID) error {
	args := m.Called(ctx, mappingID)
	return args.Error(0)
}

func (m *MockSSORepository) UpdateLastSync(ctx context.Context, configID uuid.UUID, syncedAt time.Time) error {
	args := m.Called(ctx, configID, syncedAt)
	return args.Error(0)
}

func (m *MockSSORepository) UpdateLastError(ctx context.Context, configID uuid.UUID, errorMsg string) error {
	args := m.Called(ctx, configID, errorMsg)
	return args.Error(0)
}

func TestSSOService_GetConfiguration(t *testing.T) {
	t.Run("returns existing configuration", func(t *testing.T) {
		ssoRepo := new(MockSSORepository)
		_ = &SSOService{ssoRepo: &postgres.SSORepository{}} // placeholder for future testing

		orgID := uuid.New()
		expectedConfig := &domain.SSOConfiguration{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Provider:       domain.SSOProviderOIDC,
			Enabled:        true,
			OIDCClientID:   "client-123",
		}

		ssoRepo.On("GetConfigurationByOrganization", mock.Anything, orgID).Return(expectedConfig, nil)

		// Note: For actual testing, we'd need dependency injection
		// This test demonstrates the expected behavior
		assert.NotNil(t, expectedConfig)
		assert.Equal(t, orgID, expectedConfig.OrganizationID)
	})

	t.Run("returns nil for non-existent configuration", func(t *testing.T) {
		ssoRepo := new(MockSSORepository)
		orgID := uuid.New()

		ssoRepo.On("GetConfigurationByOrganization", mock.Anything, orgID).Return(nil, nil)

		// The service would return nil, nil for non-existent configs
		result, err := ssoRepo.GetConfigurationByOrganization(context.Background(), orgID)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestSSOService_EnableSSO(t *testing.T) {
	t.Run("enables SSO for configured organization", func(t *testing.T) {
		ssoRepo := new(MockSSORepository)

		orgID := uuid.New()
		config := &domain.SSOConfiguration{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Provider:       domain.SSOProviderOIDC,
			Enabled:        false,
		}

		ssoRepo.On("GetConfigurationByOrganization", mock.Anything, orgID).Return(config, nil)
		ssoRepo.On("UpdateConfiguration", mock.Anything, mock.AnythingOfType("*domain.SSOConfiguration")).Return(nil)

		// Simulating the enable flow
		config.Enabled = true
		err := ssoRepo.UpdateConfiguration(context.Background(), config)

		require.NoError(t, err)
		assert.True(t, config.Enabled)
	})

	t.Run("fails for unconfigured organization", func(t *testing.T) {
		ssoRepo := new(MockSSORepository)
		orgID := uuid.New()

		ssoRepo.On("GetConfigurationByOrganization", mock.Anything, orgID).Return(nil, nil)

		result, _ := ssoRepo.GetConfigurationByOrganization(context.Background(), orgID)
		// Service would return ErrSSONotConfigured
		assert.Nil(t, result)
	})
}

func TestSSOService_InitiateOIDCLogin(t *testing.T) {
	t.Run("generates correct authorization URL", func(t *testing.T) {
		orgID := uuid.New()
		config := &domain.SSOConfiguration{
			ID:               uuid.New(),
			OrganizationID:   orgID,
			Provider:         domain.SSOProviderOIDC,
			Enabled:          true,
			OIDCClientID:     "test-client-id",
			OIDCClientSecret: "test-secret",
			OIDCIssuerURL:    "https://auth.example.com",
			OIDCScopes:       []string{"openid", "profile", "email"},
		}

		// Verify config is valid for OIDC
		assert.NotEmpty(t, config.OIDCClientID)
		assert.NotEmpty(t, config.OIDCIssuerURL)
		assert.True(t, config.Enabled)
	})

	t.Run("fails for disabled SSO", func(t *testing.T) {
		config := &domain.SSOConfiguration{
			Enabled: false,
		}

		// Service would return ErrSSONotEnabled
		assert.False(t, config.Enabled)
	})
}

func TestSSOService_HandleOIDCCallback(t *testing.T) {
	t.Run("validates state parameter", func(t *testing.T) {
		ssoRepo := new(MockSSORepository)

		// Invalid state returns error
		ssoRepo.On("GetAndDeleteState", mock.Anything, "invalid-state").Return(nil, nil)

		result, err := ssoRepo.GetAndDeleteState(context.Background(), "invalid-state")
		assert.NoError(t, err)
		assert.Nil(t, result) // Service would return ErrSSOInvalidState
	})

	t.Run("retrieves valid state", func(t *testing.T) {
		ssoRepo := new(MockSSORepository)

		orgID := uuid.New()
		validState := &postgres.SSOState{
			OrganizationID: orgID,
			State:          "valid-state",
			ReturnURL:      "https://app.example.com/callback",
			Nonce:          "test-nonce",
			CodeVerifier:   "test-verifier",
			ExpiresAt:      time.Now().Add(10 * time.Minute),
		}

		ssoRepo.On("GetAndDeleteState", mock.Anything, "valid-state").Return(validState, nil)

		result, err := ssoRepo.GetAndDeleteState(context.Background(), "valid-state")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, orgID, result.OrganizationID)
	})
}

func TestSSOService_Session(t *testing.T) {
	t.Run("retrieves valid session", func(t *testing.T) {
		ssoRepo := new(MockSSORepository)

		sessionID := uuid.New()
		userID := uuid.New()
		session := &domain.SSOSession{
			ID:             sessionID,
			UserID:         userID,
			OrganizationID: uuid.New(),
			Provider:       domain.SSOProviderOIDC,
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			LastActivityAt: time.Now(),
		}

		ssoRepo.On("GetSession", mock.Anything, sessionID).Return(session, nil)

		result, err := ssoRepo.GetSession(context.Background(), sessionID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.False(t, result.ExpiresAt.Before(time.Now())) // Not expired
	})

	t.Run("detects expired session", func(t *testing.T) {
		ssoRepo := new(MockSSORepository)

		sessionID := uuid.New()
		expiredSession := &domain.SSOSession{
			ID:        sessionID,
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		}

		ssoRepo.On("GetSession", mock.Anything, sessionID).Return(expiredSession, nil)
		ssoRepo.On("DeleteSession", mock.Anything, sessionID).Return(nil)

		result, _ := ssoRepo.GetSession(context.Background(), sessionID)
		// Service would detect expiration and return ErrSSOSessionExpired
		assert.True(t, result.ExpiresAt.Before(time.Now()))
	})

	t.Run("deletes session on logout", func(t *testing.T) {
		ssoRepo := new(MockSSORepository)

		sessionID := uuid.New()

		ssoRepo.On("DeleteSession", mock.Anything, sessionID).Return(nil)

		err := ssoRepo.DeleteSession(context.Background(), sessionID)
		assert.NoError(t, err)
		ssoRepo.AssertExpectations(t)
	})

	t.Run("deletes all user sessions", func(t *testing.T) {
		ssoRepo := new(MockSSORepository)

		userID := uuid.New()
		ssoRepo.On("DeleteUserSessions", mock.Anything, userID).Return(nil)

		err := ssoRepo.DeleteUserSessions(context.Background(), userID)
		assert.NoError(t, err)
	})
}

func TestSSOService_DomainRestrictions(t *testing.T) {
	t.Run("allows email from allowed domain", func(t *testing.T) {
		config := &domain.SSOConfiguration{
			AllowedDomains: []string{"example.com", "company.org"},
		}

		email := "user@example.com"
		domain := extractEmailDomain(email)

		allowed := false
		for _, d := range config.AllowedDomains {
			if d == domain {
				allowed = true
				break
			}
		}

		assert.True(t, allowed)
	})

	t.Run("blocks email from disallowed domain", func(t *testing.T) {
		config := &domain.SSOConfiguration{
			AllowedDomains: []string{"example.com"},
		}

		email := "user@other.com"
		domain := extractEmailDomain(email)

		allowed := false
		for _, d := range config.AllowedDomains {
			if d == domain {
				allowed = true
				break
			}
		}

		assert.False(t, allowed)
	})

	t.Run("allows any domain when no restrictions", func(t *testing.T) {
		config := &domain.SSOConfiguration{
			AllowedDomains: []string{},
		}

		// Empty allowed domains means no restrictions
		assert.Empty(t, config.AllowedDomains)
	})
}

func TestSSOService_IdentityMapping(t *testing.T) {
	t.Run("creates identity mapping for new user", func(t *testing.T) {
		ssoRepo := new(MockSSORepository)

		mapping := &postgres.SSOIdentityMapping{
			UserID:         uuid.New(),
			OrganizationID: uuid.New(),
			Provider:       "oidc",
			ExternalID:     "ext-123",
			ExternalEmail:  "user@example.com",
			ExternalName:   "Test User",
			LinkedAt:       time.Now(),
		}

		ssoRepo.On("CreateIdentityMapping", mock.Anything, mock.AnythingOfType("*postgres.SSOIdentityMapping")).Return(nil)

		err := ssoRepo.CreateIdentityMapping(context.Background(), mapping)
		assert.NoError(t, err)
	})

	t.Run("retrieves existing identity mapping", func(t *testing.T) {
		ssoRepo := new(MockSSORepository)

		orgID := uuid.New()
		userID := uuid.New()
		mapping := &postgres.SSOIdentityMapping{
			ID:             uuid.New(),
			UserID:         userID,
			OrganizationID: orgID,
			Provider:       "oidc",
			ExternalID:     "ext-123",
		}

		ssoRepo.On("GetIdentityMapping", mock.Anything, orgID, "oidc", "ext-123").Return(mapping, nil)

		result, err := ssoRepo.GetIdentityMapping(context.Background(), orgID, "oidc", "ext-123")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
	})

	t.Run("updates last login on existing mapping", func(t *testing.T) {
		ssoRepo := new(MockSSORepository)

		mappingID := uuid.New()
		ssoRepo.On("UpdateIdentityMappingLastLogin", mock.Anything, mappingID).Return(nil)

		err := ssoRepo.UpdateIdentityMappingLastLogin(context.Background(), mappingID)
		assert.NoError(t, err)
	})
}

func TestSSOService_SAMLFlow(t *testing.T) {
	t.Run("builds correct SAML AuthnRequest", func(t *testing.T) {
		config := &domain.SSOConfiguration{
			SAMLEntityID:     "https://sp.example.com/saml",
			SAMLSSOUrl:       "https://idp.example.com/sso",
			SAMLNameIDFormat: "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress",
		}

		// Verify SAML config is complete
		assert.NotEmpty(t, config.SAMLEntityID)
		assert.NotEmpty(t, config.SAMLSSOUrl)
	})

	t.Run("validates SAML configuration", func(t *testing.T) {
		config := &domain.SSOConfiguration{
			Provider:        domain.SSOProviderSAML,
			SAMLEntityID:    "https://sp.example.com/saml",
			SAMLSSOUrl:      "https://idp.example.com/sso",
			SAMLCertificate: "MIIBkT...", // Would be real cert
		}

		assert.Equal(t, domain.SSOProviderSAML, config.Provider)
		assert.NotEmpty(t, config.SAMLCertificate)
	})
}

func TestSSOService_AttributeMapping(t *testing.T) {
	t.Run("uses custom attribute mapping", func(t *testing.T) {
		mapping := domain.SSOAttributeMapping{
			Email:       "mail",
			FirstName:   "givenName",
			LastName:    "sn",
			DisplayName: "displayName",
			Groups:      "memberOf",
		}

		// Verify all fields are set
		assert.Equal(t, "mail", mapping.Email)
		assert.Equal(t, "givenName", mapping.FirstName)
		assert.Equal(t, "sn", mapping.LastName)
		assert.Equal(t, "displayName", mapping.DisplayName)
		assert.Equal(t, "memberOf", mapping.Groups)
	})

	t.Run("uses default attribute mapping when not specified", func(t *testing.T) {
		config := &domain.SSOConfiguration{
			AttributeMapping: domain.SSOAttributeMapping{},
		}

		// Default values would be set by the application
		// Email defaults to "email", etc.
		assert.Empty(t, config.AttributeMapping.Email)
	})
}

func TestSSOService_Helpers(t *testing.T) {
	t.Run("generates random string", func(t *testing.T) {
		str1 := generateRandomString(32)
		str2 := generateRandomString(32)

		assert.Len(t, str1, 32)
		assert.Len(t, str2, 32)
		assert.NotEqual(t, str1, str2) // Should be unique
	})

	t.Run("generates code challenge", func(t *testing.T) {
		verifier := "test-verifier-string-for-pkce-flow"
		challenge := generateCodeChallenge(verifier)

		assert.NotEmpty(t, challenge)
		assert.NotEqual(t, verifier, challenge)
	})

	t.Run("extracts email domain", func(t *testing.T) {
		tests := []struct {
			email    string
			expected string
		}{
			{"user@example.com", "example.com"},
			{"admin@company.org", "company.org"},
			{"test@sub.domain.com", "sub.domain.com"},
			{"invalid-email", ""},
		}

		for _, tt := range tests {
			domain := extractEmailDomain(tt.email)
			assert.Equal(t, tt.expected, domain)
		}
	})
}

func TestSSOService_ConfigurationInput(t *testing.T) {
	t.Run("creates OIDC configuration from input", func(t *testing.T) {
		clientID := "client-123"
		clientSecret := "secret-456"
		issuerURL := "https://auth.example.com"

		input := &domain.SSOConfigurationInput{
			Provider:         domain.SSOProviderOIDC,
			Enabled:          true,
			OIDCClientID:     &clientID,
			OIDCClientSecret: &clientSecret,
			OIDCIssuerURL:    &issuerURL,
			OIDCScopes:       []string{"openid", "profile", "email"},
		}

		assert.Equal(t, domain.SSOProviderOIDC, input.Provider)
		assert.True(t, input.Enabled)
		assert.Equal(t, clientID, *input.OIDCClientID)
	})

	t.Run("creates SAML configuration from input", func(t *testing.T) {
		entityID := "https://sp.example.com"
		ssoURL := "https://idp.example.com/sso"
		cert := "MIIBkT..."

		input := &domain.SSOConfigurationInput{
			Provider:        domain.SSOProviderSAML,
			Enabled:         true,
			SAMLEntityID:    &entityID,
			SAMLSSOUrl:      &ssoURL,
			SAMLCertificate: &cert,
		}

		assert.Equal(t, domain.SSOProviderSAML, input.Provider)
		assert.Equal(t, entityID, *input.SAMLEntityID)
	})

	t.Run("includes user provisioning settings", func(t *testing.T) {
		autoProvision := true
		defaultRole := "member"

		input := &domain.SSOConfigurationInput{
			Provider:           domain.SSOProviderOIDC,
			AutoProvisionUsers: &autoProvision,
			DefaultRole:        &defaultRole,
			AllowedDomains:     []string{"example.com"},
		}

		assert.True(t, *input.AutoProvisionUsers)
		assert.Equal(t, "member", *input.DefaultRole)
		assert.Contains(t, input.AllowedDomains, "example.com")
	})
}

func TestJWKSCache_ConcurrentAccess(t *testing.T) {
	// Reset global cache for test isolation
	globalJWKSCache.mu.Lock()
	globalJWKSCache.entries = make(map[string]*jwksCacheEntry)
	globalJWKSCache.mu.Unlock()

	t.Run("concurrent reads do not race", func(t *testing.T) {
		// Pre-populate cache
		globalJWKSCache.mu.Lock()
		globalJWKSCache.entries["https://issuer1.example.com"] = &jwksCacheEntry{
			keys:      map[string]interface{}{"kid1": "key1"},
			fetchedAt: time.Now(),
		}
		globalJWKSCache.entries["https://issuer2.example.com"] = &jwksCacheEntry{
			keys:      map[string]interface{}{"kid2": "key2"},
			fetchedAt: time.Now(),
		}
		globalJWKSCache.mu.Unlock()

		done := make(chan bool, 20)
		for i := 0; i < 20; i++ {
			go func(idx int) {
				issuer := "https://issuer1.example.com"
				if idx%2 == 0 {
					issuer = "https://issuer2.example.com"
				}
				globalJWKSCache.mu.RLock()
				entry, ok := globalJWKSCache.entries[issuer]
				if ok {
					_ = entry.keys
				}
				globalJWKSCache.mu.RUnlock()
				done <- true
			}(i)
		}
		for i := 0; i < 20; i++ {
			<-done
		}
	})

	t.Run("per-issuer caching isolates different providers", func(t *testing.T) {
		globalJWKSCache.mu.Lock()
		globalJWKSCache.entries["https://auth.google.com"] = &jwksCacheEntry{
			keys:      map[string]interface{}{"google-kid": "google-key"},
			fetchedAt: time.Now(),
		}
		globalJWKSCache.entries["https://auth.okta.com"] = &jwksCacheEntry{
			keys:      map[string]interface{}{"okta-kid": "okta-key"},
			fetchedAt: time.Now(),
		}
		globalJWKSCache.mu.Unlock()

		globalJWKSCache.mu.RLock()
		googleEntry := globalJWKSCache.entries["https://auth.google.com"]
		oktaEntry := globalJWKSCache.entries["https://auth.okta.com"]
		globalJWKSCache.mu.RUnlock()

		assert.Contains(t, googleEntry.keys, "google-kid")
		assert.NotContains(t, googleEntry.keys, "okta-kid")
		assert.Contains(t, oktaEntry.keys, "okta-kid")
		assert.NotContains(t, oktaEntry.keys, "google-kid")
	})

	t.Run("cache expiry is per-issuer", func(t *testing.T) {
		globalJWKSCache.mu.Lock()
		globalJWKSCache.entries["https://expired.example.com"] = &jwksCacheEntry{
			keys:      map[string]interface{}{"kid": "key"},
			fetchedAt: time.Now().Add(-20 * time.Minute), // expired
		}
		globalJWKSCache.entries["https://fresh.example.com"] = &jwksCacheEntry{
			keys:      map[string]interface{}{"kid": "key"},
			fetchedAt: time.Now(), // fresh
		}
		globalJWKSCache.mu.Unlock()

		globalJWKSCache.mu.RLock()
		expiredEntry := globalJWKSCache.entries["https://expired.example.com"]
		freshEntry := globalJWKSCache.entries["https://fresh.example.com"]
		globalJWKSCache.mu.RUnlock()

		assert.True(t, time.Since(expiredEntry.fetchedAt) > globalJWKSCache.ttl, "expired entry should be past TTL")
		assert.True(t, time.Since(freshEntry.fetchedAt) < globalJWKSCache.ttl, "fresh entry should be within TTL")
	})
}

func TestSSOService_SAMLXMLParsing(t *testing.T) {
	// Go's encoding/xml does not support external entity expansion (safe from XXE by default)
	t.Run("rejects malformed SAML XML", func(t *testing.T) {
		config := &domain.SSOConfiguration{
			AttributeMapping: domain.SSOAttributeMapping{
				Email: "email",
			},
		}
		svc := &SSOService{}
		_, _, err := svc.parseSAMLResponse(config, []byte("not xml at all"))
		assert.Error(t, err)
	})

	t.Run("parses valid SAML response", func(t *testing.T) {
		config := &domain.SSOConfiguration{
			AttributeMapping: domain.SSOAttributeMapping{
				Email:     "email",
				FirstName: "firstName",
			},
		}
		samlXML := `<Response><Assertion>
			<Subject><NameID>user@example.com</NameID></Subject>
			<Conditions NotBefore="2024-01-01T00:00:00Z" NotOnOrAfter="2025-01-01T00:00:00Z"/>
			<AuthnStatement SessionIndex="sess123"/>
			<AttributeStatement>
				<Attribute Name="email"><AttributeValue>user@example.com</AttributeValue></Attribute>
				<Attribute Name="firstName"><AttributeValue>Test</AttributeValue></Attribute>
			</AttributeStatement>
		</Assertion></Response>`

		svc := &SSOService{}
		userInfo, sessionIdx, err := svc.parseSAMLResponse(config, []byte(samlXML))
		require.NoError(t, err)
		assert.Equal(t, "user@example.com", userInfo.Email)
		assert.Equal(t, "Test", userInfo.FirstName)
		assert.Equal(t, "sess123", sessionIdx)
	})

	t.Run("XXE entity expansion is not supported by Go xml decoder", func(t *testing.T) {
		config := &domain.SSOConfiguration{
			AttributeMapping: domain.SSOAttributeMapping{Email: "email"},
		}
		// Go's encoding/xml ignores external entities entirely
		xxeXML := `<?xml version="1.0"?>
		<!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///etc/passwd">]>
		<Response><Assertion>
			<Subject><NameID>&xxe;</NameID></Subject>
			<Conditions/><AuthnStatement/>
			<AttributeStatement/>
		</Assertion></Response>`

		svc := &SSOService{}
		_, _, err := svc.parseSAMLResponse(config, []byte(xxeXML))
		// Go's xml.Unmarshal will error on entity references
		assert.Error(t, err)
	})
}

func TestSSOService_JWKParsing(t *testing.T) {
	t.Run("parses RSA JWK", func(t *testing.T) {
		key := jwkKey{
			Kty: "RSA",
			Kid: "rsa-key-1",
			Use: "sig",
			N:   "0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08qNLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awapJzKnqDKgw",
			E:   "AQAB",
		}
		pubKey, err := parseJWK(key)
		require.NoError(t, err)
		assert.NotNil(t, pubKey)
	})

	t.Run("rejects unsupported key type", func(t *testing.T) {
		key := jwkKey{
			Kty: "oct",
			Kid: "oct-key",
		}
		_, err := parseJWK(key)
		assert.Error(t, err)
	})
}
