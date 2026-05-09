package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// FederationHandler handles federation and export endpoints
type FederationHandler struct {
	federationService *service.FederationService
	logger            *zap.Logger
}

// NewFederationHandler creates a new federation handler
func NewFederationHandler(federationService *service.FederationService, logger *zap.Logger) *FederationHandler {
	return &FederationHandler{
		federationService: federationService,
		logger:            logger,
	}
}

// AddPeer handles POST /api/public/federation/peers
func (h *FederationHandler) AddPeer(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.OTelFederationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	peer, err := h.federationService.AddPeer(c.Context(), projectID, &input)
	if err != nil {
		return respondServiceError(
			c,
			h.logger,
			err,
			fiber.StatusInternalServerError,
			"Failed to add peer",
		)
	}

	return c.Status(fiber.StatusCreated).JSON(redactFederationPeer(peer))
}

// ListPeers handles GET /api/public/federation/peers
func (h *FederationHandler) ListPeers(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	peers := h.federationService.ListPeers(c.Context(), projectID)
	for i := range peers {
		peers[i].URL = redactEndpoint(peers[i].URL)
		peers[i].APIKey = ""
	}
	return c.JSON(fiber.Map{"peers": peers, "count": len(peers)})
}

// RemovePeer handles DELETE /api/public/federation/peers/:peerId
func (h *FederationHandler) RemovePeer(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	peerID, err := uuid.Parse(c.Params("peerId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid peer ID"})
	}

	if err := h.federationService.RemovePeer(c.Context(), projectID, peerID); err != nil {
		return respondServiceError(
			c,
			h.logger,
			err,
			fiber.StatusInternalServerError,
			"Failed to remove peer",
		)
	}

	return c.JSON(fiber.Map{"status": "removed"})
}

// FederatedQuery handles POST /api/public/federation/query
func (h *FederationHandler) FederatedQuery(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var query domain.FederationQuery
	if err := c.BodyParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.federationService.FederatedQuery(c.Context(), projectID, &query)
	if err != nil {
		return respondServiceError(
			c,
			h.logger,
			err,
			fiber.StatusInternalServerError,
			"Federation query failed",
		)
	}

	return c.JSON(result)
}

// CreateExportDestination handles POST /api/public/federation/destinations
func (h *FederationHandler) CreateExportDestination(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.ExportDestinationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	dest, err := h.federationService.CreateExportDestination(c.Context(), projectID, &input)
	if err != nil {
		return respondServiceError(
			c,
			h.logger,
			err,
			fiber.StatusInternalServerError,
			"Failed to create destination",
		)
	}

	return c.Status(fiber.StatusCreated).JSON(redactExportDestination(dest))
}

// ListExportDestinations handles GET /api/public/federation/destinations
func (h *FederationHandler) ListExportDestinations(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	destinations := h.federationService.ListExportDestinations(c.Context(), projectID)
	for i := range destinations {
		destinations[i] = *redactExportDestination(&destinations[i])
	}
	return c.JSON(fiber.Map{"destinations": destinations, "count": len(destinations)})
}
