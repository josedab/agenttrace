package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// MarketplaceHandler handles marketplace HTTP requests
type MarketplaceHandler struct {
	service *service.MarketplaceService
	logger  *zap.Logger
}

// NewMarketplaceHandler creates a new marketplace handler
func NewMarketplaceHandler(svc *service.MarketplaceService, logger *zap.Logger) *MarketplaceHandler {
	return &MarketplaceHandler{
		service: svc,
		logger:  logger,
	}
}

// Publish handles POST /api/public/marketplace
func (h *MarketplaceHandler) Publish(c *fiber.Ctx) error {
	var input domain.PackagePublishInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}
	if input.Type == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Type is required"})
	}
	if input.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Content is required"})
	}

	pkg, err := h.service.PublishPackage(c.Context(), &input)
	if err != nil {
		h.logger.Error("failed to publish package", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to publish package"})
	}

	return c.Status(fiber.StatusCreated).JSON(pkg)
}

// Search handles GET /api/public/marketplace
func (h *MarketplaceHandler) Search(c *fiber.Ctx) error {
	search := &domain.MarketplaceSearch{
		Query:  c.Query("query"),
		SortBy: c.Query("sortBy", "downloads"),
		Limit:  ParsePagination(c, 100).Limit,
		Offset: ParsePagination(c, 100).Offset,
	}

	if typeStr := c.Query("type"); typeStr != "" {
		t := domain.PackageType(typeStr)
		search.Type = &t
	}

	if tagsStr := c.Query("tags"); tagsStr != "" {
		search.Tags = splitTags(tagsStr)
	}

	packages, total := h.service.SearchPackages(c.Context(), search)
	if packages == nil {
		packages = []domain.MarketplacePackage{}
	}

	return c.JSON(fiber.Map{
		"packages": packages,
		"total":    total,
		"count":    len(packages),
	})
}

// Get handles GET /api/public/marketplace/:packageId
func (h *MarketplaceHandler) Get(c *fiber.Ctx) error {
	packageID, err := uuid.Parse(c.Params("packageId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid package ID"})
	}

	pkg, err := h.service.GetPackage(c.Context(), packageID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Package not found"})
	}

	return c.JSON(pkg)
}

// Install handles POST /api/public/marketplace/:packageId/install
func (h *MarketplaceHandler) Install(c *fiber.Ctx) error {
	packageID, err := uuid.Parse(c.Params("packageId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid package ID"})
	}

	pkg, err := h.service.InstallPackage(c.Context(), packageID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Package not found"})
	}

	return c.JSON(fiber.Map{"status": "installed", "package": pkg})
}

// Rate handles POST /api/public/marketplace/:packageId/rate
func (h *MarketplaceHandler) Rate(c *fiber.Ctx) error {
	packageID, err := uuid.Parse(c.Params("packageId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid package ID"})
	}

	var input struct {
		Score  int    `json:"score"`
		Review string `json:"review,omitempty"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Score < 1 || input.Score > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Score must be between 1 and 5"})
	}

	pkg, err := h.service.RatePackage(c.Context(), packageID, input.Score, input.Review)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
	}

	return c.JSON(pkg)
}

// Featured handles GET /api/public/marketplace/featured
func (h *MarketplaceHandler) Featured(c *fiber.Ctx) error {
	packages := h.service.GetFeatured(c.Context())
	if packages == nil {
		packages = []domain.MarketplacePackage{}
	}

	return c.JSON(fiber.Map{"packages": packages, "count": len(packages)})
}

// Categories handles GET /api/public/marketplace/categories
func (h *MarketplaceHandler) Categories(c *fiber.Ctx) error {
	categories := h.service.GetCategories(c.Context())
	return c.JSON(fiber.Map{"categories": categories})
}

// StarterKits handles GET /api/public/marketplace/starter-kits
func (h *MarketplaceHandler) StarterKits(c *fiber.Ctx) error {
	kits := h.service.GetStarterKits(c.Context())
	return c.JSON(fiber.Map{"starterKits": kits, "count": len(kits)})
}

// Reviews handles GET /api/public/marketplace/:packageId/reviews
func (h *MarketplaceHandler) Reviews(c *fiber.Ctx) error {
	packageID, err := uuid.Parse(c.Params("packageId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid package ID"})
	}

	reviews := h.service.GetReviews(c.Context(), packageID)
	if reviews == nil {
		reviews = []domain.PackageRating{}
	}

	return c.JSON(fiber.Map{"reviews": reviews, "count": len(reviews)})
}

// splitTags splits a comma-separated tags string
func splitTags(tags string) []string {
	var result []string
	for _, tag := range strings.Split(tags, ",") {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
