package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

type SkillProfileHandler struct {
	skillProfileService *service.SkillProfileService
	logger              *zap.Logger
}

func NewSkillProfileHandler(svc *service.SkillProfileService, logger *zap.Logger) *SkillProfileHandler {
	return &SkillProfileHandler{skillProfileService: svc, logger: logger}
}

func (h *SkillProfileHandler) GetProfile(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	agentName := c.Params("agentName")
	if agentName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Agent name required"})
	}
	profile, err := h.skillProfileService.GetProfile(c.Context(), projectID, agentName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get profile"})
	}
	return c.JSON(profile)
}

func (h *SkillProfileHandler) ListProfiles(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	profiles, err := h.skillProfileService.ListProfiles(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list profiles"})
	}
	return c.JSON(fiber.Map{"profiles": profiles, "count": len(profiles)})
}

func (h *SkillProfileHandler) CompareAgents(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	agentsParam := c.Query("agents", "")
	if agentsParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "agents query param required"})
	}
	agents := strings.Split(agentsParam, ",")
	comparison, err := h.skillProfileService.CompareAgents(c.Context(), projectID, agents)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to compare agents"})
	}
	return c.JSON(comparison)
}
