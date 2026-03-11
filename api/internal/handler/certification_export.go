package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CertificationExportHandler handles certification export HTTP requests
type CertificationExportHandler struct {
	service *service.CertificationExportService
	logger  *zap.Logger
}

// NewCertificationExportHandler creates a new certification export handler
func NewCertificationExportHandler(svc *service.CertificationExportService, logger *zap.Logger) *CertificationExportHandler {
	return &CertificationExportHandler{
		service: svc,
		logger:  logger,
	}
}

// ExportCertification handles POST /api/public/certifications/export
func (h *CertificationExportHandler) ExportCertification(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	var input domain.CertificationExportInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Framework == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Framework is required"})
	}

	userID := uuid.New()
	cert, err := h.service.ExportCertification(c.Context(), projectID, userID, &input)
	if err != nil {
		h.logger.Error("failed to export certification", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(cert)
}

// GetCertification handles GET /api/public/certifications/:certId
func (h *CertificationExportHandler) GetCertification(c *fiber.Ctx) error {
	certID, err := uuid.Parse(c.Params("certId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid certification ID"})
	}

	cert, err := h.service.GetCertification(c.Context(), certID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Certification not found"})
	}

	return c.JSON(cert)
}

// ListCertifications handles GET /api/public/certifications
func (h *CertificationExportHandler) ListCertifications(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	certs, err := h.service.ListCertifications(c.Context(), projectID)
	if err != nil {
		h.logger.Error("operation failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	if certs == nil {
		certs = []domain.Certification{}
	}

	return c.JSON(fiber.Map{"certifications": certs, "count": len(certs)})
}

// Download handles GET /api/public/certifications/:certId/download
func (h *CertificationExportHandler) Download(c *fiber.Ctx) error {
	certID, err := uuid.Parse(c.Params("certId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid certification ID"})
	}

	content, contentType, err := h.service.Download(c.Context(), certID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "attachment; filename=certification-"+certID.String()+".json")
	return c.Send(content)
}

// ListFrameworks handles GET /api/public/certifications/frameworks
func (h *CertificationExportHandler) ListFrameworks(c *fiber.Ctx) error {
	frameworks := h.service.ListFrameworks(c.Context())
	return c.JSON(fiber.Map{"frameworks": frameworks})
}
