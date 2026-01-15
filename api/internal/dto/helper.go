package dto

import (
	"github.com/gofiber/fiber/v2"

	"github.com/agenttrace/agenttrace/api/internal/validator"
)

// ParseAndValidate parses the request body into the given struct and validates it.
// Returns a fiber error response if parsing or validation fails.
func ParseAndValidate(c *fiber.Ctx, v any) error {
	if err := c.BodyParser(v); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if err := validator.Validate(v); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Validation Error",
				"message": "Request validation failed",
				"errors":  validationErrors,
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": err.Error(),
		})
	}

	return nil
}
