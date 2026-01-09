package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"agenttrace/internal/domain"
	"agenttrace/internal/service"
)

type SSOHandler struct {
	ssoService   *service.SSOService
	auditService *service.AuditService
}

func NewSSOHandler(ssoService *service.SSOService, auditService *service.AuditService) *SSOHandler {
	return &SSOHandler{
		ssoService:   ssoService,
		auditService: auditService,
	}
}

// RegisterRoutes registers SSO routes
func (h *SSOHandler) RegisterRoutes(app fiber.Router) {
	// Public SSO routes (for login flow)
	sso := app.Group("/auth/sso")
	sso.Get("/login/:orgId", h.InitiateLogin)
	sso.Get("/callback", h.HandleCallback)
	sso.Post("/saml/callback", h.HandleSAMLCallback)
	sso.Post("/logout", h.Logout)

	// Admin routes (protected)
	admin := app.Group("/v1/organizations/:orgId/sso")
	admin.Get("/config", h.GetConfiguration)
	admin.Put("/config", h.ConfigureSSO)
	admin.Delete("/config", h.DeleteConfiguration)
	admin.Post("/enable", h.EnableSSO)
	admin.Post("/disable", h.DisableSSO)
	admin.Get("/sessions", h.ListSessions)
}

// InitiateLogin starts the SSO login flow
// @Summary Initiate SSO login
// @Tags SSO
// @Param orgId path string true "Organization ID"
// @Param return_url query string false "URL to redirect to after login"
// @Success 302 "Redirect to IdP"
// @Failure 400 {object} ErrorResponse
// @Router /auth/sso/login/{orgId} [get]
func (h *SSOHandler) InitiateLogin(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid organization ID",
		})
	}

	returnURL := c.Query("return_url", "/")

	// Get SSO configuration
	config, err := h.ssoService.GetConfiguration(c.Context(), orgID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to get SSO configuration",
		})
	}
	if config == nil {
		return c.Status(http.StatusNotFound).JSON(ErrorResponse{
			Error: "SSO not configured for this organization",
		})
	}
	if !config.Enabled {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "SSO is not enabled for this organization",
		})
	}

	var redirectURL string
	switch config.Provider {
	case domain.SSOProviderSAML:
		redirectURL, err = h.ssoService.InitiateSAMLLogin(c.Context(), orgID, returnURL)
	case domain.SSOProviderOIDC, domain.SSOProviderOkta, domain.SSOProviderAzureAD, domain.SSOProviderGoogle:
		redirectURL, err = h.ssoService.InitiateOIDCLogin(c.Context(), orgID, returnURL)
	default:
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Unsupported SSO provider",
		})
	}

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to initiate SSO login: " + err.Error(),
		})
	}

	return c.Redirect(redirectURL, http.StatusFound)
}

// HandleCallback handles the OIDC callback
// @Summary Handle OIDC callback
// @Tags SSO
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 302 "Redirect to application"
// @Failure 400 {object} ErrorResponse
// @Router /auth/sso/callback [get]
func (h *SSOHandler) HandleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")
	errorDesc := c.Query("error_description")

	if errorParam != "" {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: errorDesc,
		})
	}

	if code == "" || state == "" {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Missing code or state parameter",
		})
	}

	session, user, err := h.ssoService.HandleOIDCCallback(c.Context(), code, state)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "SSO authentication failed: " + err.Error(),
		})
	}

	// Log the SSO login
	h.auditService.LogSSOLogin(c.Context(), session.OrganizationID, user.ID, user.Email,
		string(session.Provider), c.IP(), c.Get("User-Agent"))

	// Set session cookie or return JWT
	// In a real implementation, you would create a session cookie or JWT here
	return c.JSON(fiber.Map{
		"success":   true,
		"sessionId": session.ID,
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// HandleSAMLCallback handles the SAML callback
// @Summary Handle SAML callback
// @Tags SSO
// @Param SAMLResponse formData string true "SAML Response"
// @Param RelayState formData string true "Relay State"
// @Success 302 "Redirect to application"
// @Failure 400 {object} ErrorResponse
// @Router /auth/sso/saml/callback [post]
func (h *SSOHandler) HandleSAMLCallback(c *fiber.Ctx) error {
	samlResponse := c.FormValue("SAMLResponse")
	relayState := c.FormValue("RelayState")

	if samlResponse == "" || relayState == "" {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Missing SAML response or relay state",
		})
	}

	session, user, err := h.ssoService.HandleSAMLCallback(c.Context(), samlResponse, relayState)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "SAML authentication failed: " + err.Error(),
		})
	}

	// Log the SSO login
	h.auditService.LogSSOLogin(c.Context(), session.OrganizationID, user.ID, user.Email,
		string(session.Provider), c.IP(), c.Get("User-Agent"))

	// Set session cookie or return JWT
	return c.JSON(fiber.Map{
		"success":   true,
		"sessionId": session.ID,
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// Logout handles SSO logout
// @Summary Logout from SSO
// @Tags SSO
// @Security BearerAuth
// @Success 200 {object} map[string]bool
// @Failure 401 {object} ErrorResponse
// @Router /auth/sso/logout [post]
func (h *SSOHandler) Logout(c *fiber.Ctx) error {
	// Get session ID from request (cookie or header)
	sessionIDStr := c.Get("X-Session-ID")
	if sessionIDStr == "" {
		sessionIDStr = c.Cookies("session_id")
	}

	if sessionIDStr == "" {
		return c.JSON(fiber.Map{"success": true})
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return c.JSON(fiber.Map{"success": true})
	}

	if err := h.ssoService.Logout(c.Context(), sessionID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to logout: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{"success": true})
}

// GetConfiguration returns the SSO configuration for an organization
// @Summary Get SSO configuration
// @Tags SSO Admin
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} domain.SSOConfiguration
// @Failure 404 {object} ErrorResponse
// @Router /v1/organizations/{orgId}/sso/config [get]
func (h *SSOHandler) GetConfiguration(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid organization ID",
		})
	}

	config, err := h.ssoService.GetConfiguration(c.Context(), orgID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to get SSO configuration",
		})
	}

	if config == nil {
		return c.JSON(fiber.Map{
			"configured": false,
		})
	}

	// Don't expose secrets
	config.OIDCClientSecret = ""
	config.SAMLCertificate = "[REDACTED]"

	return c.JSON(config)
}

// ConfigureSSO creates or updates SSO configuration
// @Summary Configure SSO
// @Tags SSO Admin
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param body body domain.SSOConfigurationInput true "SSO Configuration"
// @Success 200 {object} domain.SSOConfiguration
// @Failure 400 {object} ErrorResponse
// @Router /v1/organizations/{orgId}/sso/config [put]
func (h *SSOHandler) ConfigureSSO(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid organization ID",
		})
	}

	var input domain.SSOConfigurationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid request body",
		})
	}

	config, err := h.ssoService.ConfigureSSO(c.Context(), orgID, &input)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to configure SSO: " + err.Error(),
		})
	}

	// Log the configuration
	userID := c.Locals("userID").(uuid.UUID)
	userEmail := c.Locals("userEmail").(string)
	h.auditService.LogSSOConfigured(c.Context(), orgID, userID, userEmail, string(input.Provider))

	// Don't expose secrets
	config.OIDCClientSecret = ""

	return c.JSON(config)
}

// DeleteConfiguration removes SSO configuration
// @Summary Delete SSO configuration
// @Tags SSO Admin
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} ErrorResponse
// @Router /v1/organizations/{orgId}/sso/config [delete]
func (h *SSOHandler) DeleteConfiguration(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid organization ID",
		})
	}

	if err := h.ssoService.DeleteConfiguration(c.Context(), orgID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to delete SSO configuration",
		})
	}

	// Log the deletion
	userID := c.Locals("userID").(uuid.UUID)
	userEmail := c.Locals("userEmail").(string)
	h.auditService.LogSSODisabled(c.Context(), orgID, userID, userEmail)

	return c.JSON(fiber.Map{"success": true})
}

// EnableSSO enables SSO for an organization
// @Summary Enable SSO
// @Tags SSO Admin
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} ErrorResponse
// @Router /v1/organizations/{orgId}/sso/enable [post]
func (h *SSOHandler) EnableSSO(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid organization ID",
		})
	}

	if err := h.ssoService.EnableSSO(c.Context(), orgID, true); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to enable SSO: " + err.Error(),
		})
	}

	// Log the action
	userID := c.Locals("userID").(uuid.UUID)
	userEmail := c.Locals("userEmail").(string)
	h.auditService.LogSSOEnabled(c.Context(), orgID, userID, userEmail)

	return c.JSON(fiber.Map{"success": true})
}

// DisableSSO disables SSO for an organization
// @Summary Disable SSO
// @Tags SSO Admin
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} ErrorResponse
// @Router /v1/organizations/{orgId}/sso/disable [post]
func (h *SSOHandler) DisableSSO(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid organization ID",
		})
	}

	if err := h.ssoService.EnableSSO(c.Context(), orgID, false); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to disable SSO: " + err.Error(),
		})
	}

	// Log the action
	userID := c.Locals("userID").(uuid.UUID)
	userEmail := c.Locals("userEmail").(string)
	h.auditService.LogSSODisabled(c.Context(), orgID, userID, userEmail)

	return c.JSON(fiber.Map{"success": true})
}

// ListSessions lists active SSO sessions for an organization
// @Summary List SSO sessions
// @Tags SSO Admin
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {array} domain.SSOSession
// @Failure 400 {object} ErrorResponse
// @Router /v1/organizations/{orgId}/sso/sessions [get]
func (h *SSOHandler) ListSessions(c *fiber.Ctx) error {
	// This would list active SSO sessions for the organization
	// Implementation would depend on query requirements
	return c.JSON(fiber.Map{
		"sessions": []interface{}{},
		"total":    0,
	})
}
