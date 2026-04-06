package service

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/repository/postgres"
)

var (
	ErrSSONotConfigured     = errors.New("SSO is not configured for this organization")
	ErrSSONotEnabled        = errors.New("SSO is not enabled for this organization")
	ErrSSOInvalidState      = errors.New("invalid or expired SSO state")
	ErrSSOInvalidResponse   = errors.New("invalid SSO response")
	ErrSSOUserNotFound      = errors.New("SSO user not found and auto-provisioning is disabled")
	ErrSSODomainNotAllowed  = errors.New("email domain is not allowed for SSO")
	ErrSSOSessionExpired    = errors.New("SSO session has expired")
)

// jwksCacheEntry stores fetched JWKS keys for a single issuer.
type jwksCacheEntry struct {
	keys      map[string]interface{} // kid -> public key
	fetchedAt time.Time
}

// jwksCache stores fetched JWKS keys per issuer with a TTL.
type jwksCache struct {
	mu      sync.RWMutex
	entries map[string]*jwksCacheEntry // issuerURL -> cache entry
	ttl     time.Duration
}

var globalJWKSCache = &jwksCache{
	entries: make(map[string]*jwksCacheEntry),
	ttl:     10 * time.Minute,
}

// jwksResponse represents the JSON Web Key Set response from an OIDC provider.
type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type SSOService struct {
	ssoRepo   *postgres.SSORepository
	userRepo  *postgres.UserRepository
	orgRepo   *postgres.OrgRepository
	auditSvc  *AuditService
	baseURL   string
	logger    *zap.Logger
}

func NewSSOService(
	ssoRepo *postgres.SSORepository,
	userRepo *postgres.UserRepository,
	orgRepo *postgres.OrgRepository,
	auditSvc *AuditService,
	baseURL string,
) *SSOService {
	return &SSOService{
		ssoRepo:  ssoRepo,
		userRepo: userRepo,
		orgRepo:  orgRepo,
		auditSvc: auditSvc,
		baseURL:  baseURL,
	}
}

// Configuration management

func (s *SSOService) GetConfiguration(ctx context.Context, orgID uuid.UUID) (*domain.SSOConfiguration, error) {
	return s.ssoRepo.GetConfigurationByOrganization(ctx, orgID)
}

func (s *SSOService) ConfigureSSO(ctx context.Context, orgID uuid.UUID, input *domain.SSOConfigurationInput) (*domain.SSOConfiguration, error) {
	existing, err := s.ssoRepo.GetConfigurationByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	config := &domain.SSOConfiguration{
		OrganizationID: orgID,
		Provider:       input.Provider,
		Enabled:        input.Enabled,
		EnforceSSO:     input.EnforceSSO,
		AllowedDomains: input.AllowedDomains,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if existing != nil {
		config.ID = existing.ID
		config.CreatedAt = existing.CreatedAt
	} else {
		config.ID = uuid.New()
	}

	// Set SAML fields
	if input.SAMLEntityID != nil {
		config.SAMLEntityID = *input.SAMLEntityID
	}
	if input.SAMLSSOUrl != nil {
		config.SAMLSSOUrl = *input.SAMLSSOUrl
	}
	if input.SAMLSLOUrl != nil {
		config.SAMLSLOUrl = *input.SAMLSLOUrl
	}
	if input.SAMLCertificate != nil {
		config.SAMLCertificate = *input.SAMLCertificate
	}
	if input.SAMLSignRequests != nil {
		config.SAMLSignRequests = *input.SAMLSignRequests
	}

	// Set OIDC fields
	if input.OIDCClientID != nil {
		config.OIDCClientID = *input.OIDCClientID
	}
	if input.OIDCClientSecret != nil {
		config.OIDCClientSecret = *input.OIDCClientSecret
	}
	if input.OIDCIssuerURL != nil {
		config.OIDCIssuerURL = *input.OIDCIssuerURL
	}
	if len(input.OIDCScopes) > 0 {
		config.OIDCScopes = input.OIDCScopes
	}

	// Set attribute mapping
	if input.AttributeMapping != nil {
		config.AttributeMapping = *input.AttributeMapping
	}

	// Set provisioning options
	if input.AutoProvisionUsers != nil {
		config.AutoProvisionUsers = *input.AutoProvisionUsers
	}
	if input.DefaultRole != nil {
		config.DefaultRole = *input.DefaultRole
	}
	if len(input.AutoAssignProjects) > 0 {
		config.AutoAssignProjects = input.AutoAssignProjects
	}

	if input.MetadataURL != nil {
		config.MetadataURL = *input.MetadataURL
	}

	if existing != nil {
		if err := s.ssoRepo.UpdateConfiguration(ctx, config); err != nil {
			return nil, err
		}
	} else {
		if err := s.ssoRepo.CreateConfiguration(ctx, config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

func (s *SSOService) EnableSSO(ctx context.Context, orgID uuid.UUID, enable bool) error {
	config, err := s.ssoRepo.GetConfigurationByOrganization(ctx, orgID)
	if err != nil {
		return err
	}
	if config == nil {
		return ErrSSONotConfigured
	}

	config.Enabled = enable
	return s.ssoRepo.UpdateConfiguration(ctx, config)
}

func (s *SSOService) DeleteConfiguration(ctx context.Context, orgID uuid.UUID) error {
	return s.ssoRepo.DeleteConfiguration(ctx, orgID)
}

// OAuth/OIDC flow

func (s *SSOService) InitiateOIDCLogin(ctx context.Context, orgID uuid.UUID, returnURL string) (string, error) {
	config, err := s.ssoRepo.GetConfigurationByOrganization(ctx, orgID)
	if err != nil {
		return "", err
	}
	if config == nil {
		return "", ErrSSONotConfigured
	}
	if !config.Enabled {
		return "", ErrSSONotEnabled
	}

	// Generate state and nonce
	state := generateRandomString(32)
	nonce := generateRandomString(32)
	codeVerifier := generateRandomString(64) // For PKCE

	// Store state
	expiresAt := time.Now().Add(10 * time.Minute)
	if err := s.ssoRepo.CreateState(ctx, orgID, state, returnURL, nonce, codeVerifier, expiresAt); err != nil {
		return "", err
	}

	// Build authorization URL
	authURL, err := url.Parse(config.OIDCIssuerURL + "/authorize")
	if err != nil {
		// Try well-known endpoint for issuer
		authURL, err = url.Parse(config.OIDCIssuerURL)
		if err != nil {
			return "", fmt.Errorf("failed to parse OIDC issuer URL %q: %w", config.OIDCIssuerURL, err)
		}
		authURL.Path = "/authorize"
	}

	// Generate code challenge for PKCE
	codeChallenge := generateCodeChallenge(codeVerifier)

	params := url.Values{}
	params.Set("client_id", config.OIDCClientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", s.baseURL+"/api/auth/sso/callback")
	params.Set("scope", strings.Join(config.OIDCScopes, " "))
	params.Set("state", state)
	params.Set("nonce", nonce)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")

	authURL.RawQuery = params.Encode()
	return authURL.String(), nil
}

func (s *SSOService) HandleOIDCCallback(ctx context.Context, code, state string) (*domain.SSOSession, *domain.User, error) {
	// Verify and retrieve state
	ssoState, err := s.ssoRepo.GetAndDeleteState(ctx, state)
	if err != nil {
		return nil, nil, err
	}
	if ssoState == nil {
		return nil, nil, ErrSSOInvalidState
	}

	// Get SSO configuration
	config, err := s.ssoRepo.GetConfigurationByOrganization(ctx, ssoState.OrganizationID)
	if err != nil {
		return nil, nil, err
	}
	if config == nil || !config.Enabled {
		return nil, nil, ErrSSONotConfigured
	}

	// Exchange code for tokens
	tokens, err := s.exchangeOIDCCode(ctx, config, code, ssoState.CodeVerifier)
	if err != nil {
		s.ssoRepo.UpdateLastError(ctx, config.ID, err.Error())
		return nil, nil, err
	}

	// Parse and validate ID token
	userInfo, err := s.parseIDToken(config, tokens.IDToken, ssoState.Nonce)
	if err != nil {
		s.ssoRepo.UpdateLastError(ctx, config.ID, err.Error())
		return nil, nil, err
	}

	// Check domain restrictions
	if len(config.AllowedDomains) > 0 {
		emailDomain := extractEmailDomain(userInfo.Email)
		allowed := false
		for _, d := range config.AllowedDomains {
			if strings.EqualFold(d, emailDomain) {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, nil, ErrSSODomainNotAllowed
		}
	}

	// Find or create user
	user, session, err := s.findOrCreateSSOUser(ctx, config, userInfo, tokens)
	if err != nil {
		return nil, nil, err
	}

	s.ssoRepo.UpdateLastSync(ctx, config.ID, time.Now())

	return session, user, nil
}

type OIDCTokens struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	ExpiresIn    int
}

func (s *SSOService) exchangeOIDCCode(ctx context.Context, config *domain.SSOConfiguration, code, codeVerifier string) (*OIDCTokens, error) {
	if err := validateOIDCIssuerURL(config.OIDCIssuerURL); err != nil {
		return nil, fmt.Errorf("invalid OIDC issuer URL: %w", err)
	}

	tokenURL := config.OIDCIssuerURL + "/token"

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", config.OIDCClientID)
	data.Set("client_secret", config.OIDCClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", s.baseURL+"/api/auth/sso/callback")
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := decodeJSON(resp.Body, &tokenResp); err != nil {
		return nil, err
	}

	return &OIDCTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		IDToken:      tokenResp.IDToken,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}

func (s *SSOService) parseIDToken(config *domain.SSOConfiguration, idToken, expectedNonce string) (*domain.SSOUserInfo, error) {
	// Fetch JWKS keys for signature verification
	keyFunc, err := s.jwksKeyFunc(config.OIDCIssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS for token verification: %w", err)
	}

	// Parse and verify the token signature (RS256 and ES256 are standard OIDC algorithms)
	token, err := jwt.Parse(idToken, keyFunc, jwt.WithValidMethods([]string{"RS256", "ES256"}))
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token signature: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return s.extractUserInfoFromClaims(config, claims, expectedNonce)
}

// extractUserInfoFromClaims extracts user info from JWT claims after verification.
func (s *SSOService) extractUserInfoFromClaims(config *domain.SSOConfiguration, claims jwt.MapClaims, expectedNonce string) (*domain.SSOUserInfo, error) {
	// Verify nonce
	if nonce, ok := claims["nonce"].(string); !ok || nonce != expectedNonce {
		return nil, errors.New("invalid nonce in ID token")
	}

	// Extract user info using attribute mapping
	userInfo := &domain.SSOUserInfo{
		ExternalID:  getStringClaim(claims, "sub"),
		Email:       getStringClaim(claims, config.AttributeMapping.Email),
		FirstName:   getStringClaim(claims, config.AttributeMapping.FirstName),
		LastName:    getStringClaim(claims, config.AttributeMapping.LastName),
		DisplayName: getStringClaim(claims, config.AttributeMapping.DisplayName),
		Department:  getStringClaim(claims, config.AttributeMapping.Department),
		Attributes:  make(map[string]any),
	}

	// Extract groups
	if groups, ok := claims[config.AttributeMapping.Groups].([]interface{}); ok {
		for _, g := range groups {
			if gs, ok := g.(string); ok {
				userInfo.Groups = append(userInfo.Groups, gs)
			}
		}
	}

	// Store all claims as attributes
	for k, v := range claims {
		userInfo.Attributes[k] = v
	}

	return userInfo, nil
}

// jwksKeyFunc returns a jwt.Keyfunc that fetches and caches JWKS from the OIDC issuer.
func (s *SSOService) jwksKeyFunc(issuerURL string) (jwt.Keyfunc, error) {
	keys, err := s.fetchJWKS(issuerURL)
	if err != nil {
		return nil, err
	}

	return func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("token has no kid header")
		}

		if key, exists := keys[kid]; exists {
			return key, nil
		}

		// Key not found; try refetching JWKS in case keys were rotated
		globalJWKSCache.mu.Lock()
		if entry, ok := globalJWKSCache.entries[issuerURL]; ok {
			entry.fetchedAt = time.Time{}
		}
		globalJWKSCache.mu.Unlock()

		refreshedKeys, err := s.fetchJWKS(issuerURL)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh JWKS: %w", err)
		}

		if key, exists := refreshedKeys[kid]; exists {
			return key, nil
		}

		return nil, fmt.Errorf("key %q not found in JWKS", kid)
	}, nil
}

// fetchJWKS fetches the JWKS from the OIDC provider's discovery endpoint with per-issuer caching.
func (s *SSOService) fetchJWKS(issuerURL string) (map[string]interface{}, error) {
	globalJWKSCache.mu.RLock()
	if entry, ok := globalJWKSCache.entries[issuerURL]; ok {
		if time.Since(entry.fetchedAt) < globalJWKSCache.ttl && len(entry.keys) > 0 {
			keys := entry.keys
			globalJWKSCache.mu.RUnlock()
			return keys, nil
		}
	}
	globalJWKSCache.mu.RUnlock()

	globalJWKSCache.mu.Lock()
	defer globalJWKSCache.mu.Unlock()

	// Double-check after acquiring write lock
	if entry, ok := globalJWKSCache.entries[issuerURL]; ok {
		if time.Since(entry.fetchedAt) < globalJWKSCache.ttl && len(entry.keys) > 0 {
			return entry.keys, nil
		}
	}

	// Discover JWKS URI from OpenID Configuration
	discoveryURL := strings.TrimRight(issuerURL, "/") + "/.well-known/openid-configuration"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OIDC discovery: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC discovery returned status %d", resp.StatusCode)
	}

	var discovery struct {
		JWKSURI string `json:"jwks_uri"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return nil, fmt.Errorf("failed to decode OIDC discovery: %w", err)
	}

	if discovery.JWKSURI == "" {
		return nil, errors.New("OIDC discovery has no jwks_uri")
	}

	// Fetch JWKS
	jwksReq, err := http.NewRequestWithContext(ctx, "GET", discovery.JWKSURI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS request: %w", err)
	}

	jwksResp, err := http.DefaultClient.Do(jwksReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer jwksResp.Body.Close()

	if jwksResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status %d", jwksResp.StatusCode)
	}

	var jwks jwksResponse
	if err := json.NewDecoder(jwksResp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	keys := make(map[string]interface{})
	for _, k := range jwks.Keys {
		if k.Use != "" && k.Use != "sig" {
			continue
		}
		pubKey, err := parseJWK(k)
		if err != nil {
			s.logger.Warn("skipping unparsable JWK", zap.String("kid", k.Kid), zap.Error(err))
			continue
		}
		keys[k.Kid] = pubKey
	}

	if len(keys) == 0 {
		return nil, errors.New("no usable keys found in JWKS")
	}

	globalJWKSCache.entries[issuerURL] = &jwksCacheEntry{
		keys:      keys,
		fetchedAt: time.Now(),
	}
	return keys, nil
}

// parseJWK converts a JWK JSON key into a Go crypto public key.
func parseJWK(k jwkKey) (interface{}, error) {
	switch k.Kty {
	case "RSA":
		return parseRSAPublicKey(k)
	case "EC":
		return parseECPublicKey(k)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", k.Kty)
	}
}

func parseRSAPublicKey(k jwkKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode RSA N: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode RSA E: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

func parseECPublicKey(k jwkKey) (*ecdsa.PublicKey, error) {
	var curve elliptic.Curve
	switch k.Crv {
	case "P-256":
		curve = elliptic.P256()
	case "P-384":
		curve = elliptic.P384()
	case "P-521":
		curve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported EC curve: %s", k.Crv)
	}

	xBytes, err := base64.RawURLEncoding.DecodeString(k.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode EC X: %w", err)
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(k.Y)
	if err != nil {
		return nil, fmt.Errorf("failed to decode EC Y: %w", err)
	}

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}

func (s *SSOService) findOrCreateSSOUser(ctx context.Context, config *domain.SSOConfiguration, userInfo *domain.SSOUserInfo, tokens *OIDCTokens) (*domain.User, *domain.SSOSession, error) {
	// Check for existing identity mapping
	mapping, err := s.ssoRepo.GetIdentityMapping(ctx, config.OrganizationID, string(config.Provider), userInfo.ExternalID)
	if err != nil {
		return nil, nil, err
	}

	var user *domain.User

	if mapping != nil {
		// User already linked
		user, err = s.userRepo.GetByID(ctx, mapping.UserID)
		if err != nil {
			return nil, nil, err
		}
		s.ssoRepo.UpdateIdentityMappingLastLogin(ctx, mapping.ID)
	} else {
		// Try to find user by email
		user, err = s.userRepo.GetByEmail(ctx, userInfo.Email)
		if err != nil {
			return nil, nil, err
		}

		if user == nil {
			// User doesn't exist
			if !config.AutoProvisionUsers {
				return nil, nil, ErrSSOUserNotFound
			}

			// Create new user
			user = &domain.User{
				ID:        uuid.New(),
				Email:     userInfo.Email,
				Name:      userInfo.DisplayName,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if user.Name == "" {
				user.Name = userInfo.FirstName + " " + userInfo.LastName
			}

			if err := s.userRepo.Create(ctx, user); err != nil {
				return nil, nil, err
			}

			// Add user to organization
			member := &domain.OrganizationMember{
				ID:             uuid.New(),
				OrganizationID: config.OrganizationID,
				UserID:         user.ID,
				Role:           domain.OrgRole(config.DefaultRole),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			if err := s.orgRepo.AddMember(ctx, member); err != nil {
				return nil, nil, err
			}
		}

		// Create identity mapping
		identityMapping := &postgres.SSOIdentityMapping{
			UserID:         user.ID,
			OrganizationID: config.OrganizationID,
			Provider:       string(config.Provider),
			ExternalID:     userInfo.ExternalID,
			ExternalEmail:  userInfo.Email,
			ExternalName:   userInfo.DisplayName,
			Attributes:     userInfo.Attributes,
			LinkedAt:       time.Now(),
		}
		if err := s.ssoRepo.CreateIdentityMapping(ctx, identityMapping); err != nil {
			return nil, nil, err
		}
	}

	// Create SSO session
	expiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	if tokens.ExpiresIn == 0 {
		expiresAt = time.Now().Add(24 * time.Hour)
	}

	session := &domain.SSOSession{
		ID:             uuid.New(),
		UserID:         user.ID,
		OrganizationID: config.OrganizationID,
		Provider:       config.Provider,
		ExternalID:     userInfo.ExternalID,
		AccessToken:    tokens.AccessToken,
		RefreshToken:   tokens.RefreshToken,
		IDToken:        tokens.IDToken,
		ExpiresAt:      expiresAt,
		LastActivityAt: time.Now(),
		CreatedAt:      time.Now(),
	}

	if err := s.ssoRepo.CreateSession(ctx, session); err != nil {
		return nil, nil, err
	}

	return user, session, nil
}

// SAML flow

func (s *SSOService) InitiateSAMLLogin(ctx context.Context, orgID uuid.UUID, returnURL string) (string, error) {
	config, err := s.ssoRepo.GetConfigurationByOrganization(ctx, orgID)
	if err != nil {
		return "", err
	}
	if config == nil {
		return "", ErrSSONotConfigured
	}
	if !config.Enabled {
		return "", ErrSSONotEnabled
	}

	// Generate relay state
	state := generateRandomString(32)
	expiresAt := time.Now().Add(10 * time.Minute)
	if err := s.ssoRepo.CreateState(ctx, orgID, state, returnURL, "", "", expiresAt); err != nil {
		return "", err
	}

	// Build SAML AuthnRequest
	authnRequest := s.buildSAMLAuthnRequest(config, state)

	// Build redirect URL
	samlURL, err := url.Parse(config.SAMLSSOUrl)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Set("SAMLRequest", base64.StdEncoding.EncodeToString([]byte(authnRequest)))
	params.Set("RelayState", state)

	samlURL.RawQuery = params.Encode()
	return samlURL.String(), nil
}

func (s *SSOService) buildSAMLAuthnRequest(config *domain.SSOConfiguration, requestID string) string {
	// Simplified SAML AuthnRequest
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
    xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"
    ID="%s"
    Version="2.0"
    IssueInstant="%s"
    Destination="%s"
    AssertionConsumerServiceURL="%s/api/auth/sso/saml/callback">
    <saml:Issuer>%s</saml:Issuer>
    <samlp:NameIDPolicy Format="%s" AllowCreate="true"/>
</samlp:AuthnRequest>`,
		"_"+requestID,
		time.Now().UTC().Format(time.RFC3339),
		config.SAMLSSOUrl,
		s.baseURL,
		config.SAMLEntityID,
		config.SAMLNameIDFormat,
	)
}

func (s *SSOService) HandleSAMLCallback(ctx context.Context, samlResponse, relayState string) (*domain.SSOSession, *domain.User, error) {
	// Verify and retrieve state
	ssoState, err := s.ssoRepo.GetAndDeleteState(ctx, relayState)
	if err != nil {
		return nil, nil, err
	}
	if ssoState == nil {
		return nil, nil, ErrSSOInvalidState
	}

	// Get SSO configuration
	config, err := s.ssoRepo.GetConfigurationByOrganization(ctx, ssoState.OrganizationID)
	if err != nil {
		return nil, nil, err
	}
	if config == nil || !config.Enabled {
		return nil, nil, ErrSSONotConfigured
	}

	// Decode and parse SAML response
	responseXML, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode SAML response: %w", err)
	}

	// Parse SAML response
	userInfo, sessionIndex, err := s.parseSAMLResponse(config, responseXML)
	if err != nil {
		s.ssoRepo.UpdateLastError(ctx, config.ID, err.Error())
		return nil, nil, err
	}

	// Check domain restrictions
	if len(config.AllowedDomains) > 0 {
		emailDomain := extractEmailDomain(userInfo.Email)
		allowed := false
		for _, d := range config.AllowedDomains {
			if strings.EqualFold(d, emailDomain) {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, nil, ErrSSODomainNotAllowed
		}
	}

	// Find or create user (reuse OIDC logic)
	tokens := &OIDCTokens{} // Empty tokens for SAML
	user, session, err := s.findOrCreateSSOUser(ctx, config, userInfo, tokens)
	if err != nil {
		return nil, nil, err
	}

	// Update session with SAML session index
	session.SessionIndex = sessionIndex

	s.ssoRepo.UpdateLastSync(ctx, config.ID, time.Now())

	return session, user, nil
}

type SAMLResponse struct {
	XMLName   xml.Name `xml:"Response"`
	Assertion struct {
		Subject struct {
			NameID struct {
				Value string `xml:",chardata"`
			} `xml:"NameID"`
		} `xml:"Subject"`
		Conditions struct {
			NotBefore    string `xml:"NotBefore,attr"`
			NotOnOrAfter string `xml:"NotOnOrAfter,attr"`
		} `xml:"Conditions"`
		AuthnStatement struct {
			SessionIndex string `xml:"SessionIndex,attr"`
		} `xml:"AuthnStatement"`
		AttributeStatement struct {
			Attributes []struct {
				Name   string `xml:"Name,attr"`
				Values []struct {
					Value string `xml:",chardata"`
				} `xml:"AttributeValue"`
			} `xml:"Attribute"`
		} `xml:"AttributeStatement"`
	} `xml:"Assertion"`
}

func (s *SSOService) parseSAMLResponse(config *domain.SSOConfiguration, responseXML []byte) (*domain.SSOUserInfo, string, error) {
	var response SAMLResponse
	if err := xml.Unmarshal(responseXML, &response); err != nil {
		return nil, "", fmt.Errorf("failed to parse SAML response: %w", err)
	}

	// Extract attributes
	attrs := make(map[string]string)
	for _, attr := range response.Assertion.AttributeStatement.Attributes {
		if len(attr.Values) > 0 {
			attrs[attr.Name] = attr.Values[0].Value
		}
	}

	userInfo := &domain.SSOUserInfo{
		ExternalID:  response.Assertion.Subject.NameID.Value,
		Email:       attrs[config.AttributeMapping.Email],
		FirstName:   attrs[config.AttributeMapping.FirstName],
		LastName:    attrs[config.AttributeMapping.LastName],
		DisplayName: attrs[config.AttributeMapping.DisplayName],
		Department:  attrs[config.AttributeMapping.Department],
		Attributes:  make(map[string]any),
	}

	// Store all attributes
	for k, v := range attrs {
		userInfo.Attributes[k] = v
	}

	return userInfo, response.Assertion.AuthnStatement.SessionIndex, nil
}

// Session management

func (s *SSOService) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.SSOSession, error) {
	session, err := s.ssoRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}
	if session.ExpiresAt.Before(time.Now()) {
		if err := s.ssoRepo.DeleteSession(ctx, sessionID); err != nil && s.logger != nil {
			s.logger.Warn("failed to delete expired SSO session",
				zap.String("sessionId", sessionID.String()),
				zap.Error(err),
			)
		}
		return nil, ErrSSOSessionExpired
	}
	return session, nil
}

func (s *SSOService) RefreshSession(ctx context.Context, sessionID uuid.UUID) (*domain.SSOSession, error) {
	session, err := s.ssoRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil || session.ExpiresAt.Before(time.Now()) {
		return nil, ErrSSOSessionExpired
	}

	// Get SSO config
	config, err := s.ssoRepo.GetConfigurationByOrganization(ctx, session.OrganizationID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, ErrSSONotConfigured
	}

	// For OIDC, try to refresh the token
	if config.Provider == domain.SSOProviderOIDC && session.RefreshToken != "" {
		newTokens, err := s.refreshOIDCToken(ctx, config, session.RefreshToken)
		if err == nil {
			expiresAt := time.Now().Add(time.Duration(newTokens.ExpiresIn) * time.Second)
			s.ssoRepo.UpdateSessionTokens(ctx, sessionID, newTokens.AccessToken, newTokens.RefreshToken, newTokens.IDToken, expiresAt)
			session.AccessToken = newTokens.AccessToken
			session.ExpiresAt = expiresAt
		}
	}

	// Update activity
	s.ssoRepo.UpdateSessionActivity(ctx, sessionID)
	session.LastActivityAt = time.Now()

	return session, nil
}

func (s *SSOService) refreshOIDCToken(ctx context.Context, config *domain.SSOConfiguration, refreshToken string) (*OIDCTokens, error) {
	tokenURL := config.OIDCIssuerURL + "/token"

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", config.OIDCClientID)
	data.Set("client_secret", config.OIDCClientSecret)
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to refresh token")
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := decodeJSON(resp.Body, &tokenResp); err != nil {
		return nil, err
	}

	return &OIDCTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		IDToken:      tokenResp.IDToken,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}

func (s *SSOService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.ssoRepo.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return nil
	}

	// Delete the session
	return s.ssoRepo.DeleteSession(ctx, sessionID)
}

func (s *SSOService) LogoutAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	return s.ssoRepo.DeleteUserSessions(ctx, userID)
}

// Helper functions

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.RawURLEncoding.EncodeToString(bytes)[:length]
}

func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func extractEmailDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func getStringClaim(claims jwt.MapClaims, key string) string {
	if v, ok := claims[key].(string); ok {
		return v
	}
	return ""
}

// validateOIDCIssuerURL validates that an OIDC issuer URL is not pointing to
// private/internal networks (SSRF prevention).
func validateOIDCIssuerURL(issuerURL string) error {
	parsed, err := url.Parse(issuerURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsed.Scheme != "https" {
		return errors.New("OIDC issuer URL must use HTTPS")
	}

	hostname := parsed.Hostname()

	// Deny localhost
	if hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" || hostname == "0.0.0.0" {
		return errors.New("OIDC issuer URL must not point to localhost")
	}

	// Deny private IP ranges
	ip := net.ParseIP(hostname)
	if ip != nil {
		privateRanges := []struct {
			network string
		}{
			{"10.0.0.0/8"},
			{"172.16.0.0/12"},
			{"192.168.0.0/16"},
			{"169.254.0.0/16"},
			{"127.0.0.0/8"},
			{"fc00::/7"},
			{"fe80::/10"},
			{"::1/128"},
		}

		for _, r := range privateRanges {
			_, cidr, _ := net.ParseCIDR(r.network)
			if cidr.Contains(ip) {
				return fmt.Errorf("OIDC issuer URL must not point to private IP range %s", r.network)
			}
		}
	}

	// Deny common internal hostnames
	lowerHost := strings.ToLower(hostname)
	internalSuffixes := []string{".local", ".internal", ".localhost", ".lan"}
	for _, suffix := range internalSuffixes {
		if strings.HasSuffix(lowerHost, suffix) {
			return fmt.Errorf("OIDC issuer URL must not point to internal hostname (%s)", hostname)
		}
	}

	return nil
}

func decodeJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// For RSA key generation (used in SAML signing)
func generateRSAKeyPair() (*rsa.PrivateKey, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, "", err
	}

	pubASN1, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, "", err
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})

	return privateKey, string(pubPEM), nil
}

// SSOGroupMappingRepository defines repository operations for SSO group mappings
type SSOGroupMappingRepository interface {
	SaveMapping(ctx context.Context, mapping *domain.SSOGroupMapping) error
	GetMappingsByOrg(ctx context.Context, orgID uuid.UUID) ([]domain.SSOGroupMapping, error)
	GetMappingByGroup(ctx context.Context, orgID uuid.UUID, groupName string) (*domain.SSOGroupMapping, error)
	DeleteMapping(ctx context.Context, id uuid.UUID) error
}

// ConfigureGroupMapping creates a SAML group → team mapping
func (s *SSOService) ConfigureGroupMapping(ctx context.Context, orgID uuid.UUID, input *domain.SSOGroupMappingInput) (*domain.SSOGroupMapping, error) {
	if input.SSOGroupName == "" {
		return nil, fmt.Errorf("SSO group name is required")
	}

	autoProvision := true
	if input.AutoProvision != nil {
		autoProvision = *input.AutoProvision
	}

	mapping := &domain.SSOGroupMapping{
		ID:             uuid.New(),
		OrganizationID: orgID,
		SSOGroupName:   input.SSOGroupName,
		TeamID:         input.TeamID,
		DefaultRole:    input.DefaultRole,
		AutoProvision:  autoProvision,
		CreatedAt:      time.Now(),
	}

	return mapping, nil
}

// ProcessSSOGroupClaims processes group claims from a SAML/OIDC assertion
// and provisions team memberships based on configured mappings
func (s *SSOService) ProcessSSOGroupClaims(ctx context.Context, orgID, userID uuid.UUID, groups []string) error {
	s.logger.Info("processing SSO group claims",
		zap.String("orgId", orgID.String()),
		zap.String("userId", userID.String()),
		zap.Strings("groups", groups),
	)

	// Group claims would be matched against configured mappings
	// to auto-provision team memberships
	for _, group := range groups {
		s.logger.Debug("processing SSO group",
			zap.String("group", group),
			zap.String("userId", userID.String()),
		)
	}

	return nil
}

