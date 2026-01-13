package service

import (
	"context"
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
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

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

type SSOService struct {
	ssoRepo   *postgres.SSORepository
	userRepo  *postgres.UserRepository
	orgRepo   *postgres.OrgRepository
	auditSvc  *AuditService
	baseURL   string
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
		authURL, _ = url.Parse(config.OIDCIssuerURL)
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
	// Parse without verification for now (in production, fetch JWKS and verify)
	token, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse ID token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

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
		s.ssoRepo.DeleteSession(ctx, sessionID)
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

