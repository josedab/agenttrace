package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
)

// Pagination represents pagination parameters for list operations.
type Pagination struct {
	Limit  int
	Offset int
}

// DefaultPagination provides default pagination values.
var DefaultPagination = Pagination{Limit: 50, Offset: 0}

// RequireProjectID extracts the project ID from the request context.
// If the project ID is not found, it sends an unauthorized response and returns an error.
// Returns the project ID and nil on success.
func RequireProjectID(c *fiber.Ctx) (uuid.UUID, error) {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return uuid.Nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}
	return projectID, nil
}

// RequireUserID extracts the user ID from the request context.
// If the user ID is not found, it sends an unauthorized response and returns an error.
// Returns the user ID and nil on success.
func RequireUserID(c *fiber.Ctx) (uuid.UUID, error) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return uuid.Nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "User ID not found",
		})
	}
	return userID, nil
}

// ParsePagination extracts limit and offset query parameters with validation.
// maxLimit specifies the maximum allowed limit (0 for no maximum).
func ParsePagination(c *fiber.Ctx, maxLimit int) Pagination {
	p := Pagination{
		Limit:  parseQueryInt(c, "limit", DefaultPagination.Limit),
		Offset: parseQueryInt(c, "offset", DefaultPagination.Offset),
	}

	if p.Limit < 0 {
		p.Limit = DefaultPagination.Limit
	}
	if maxLimit > 0 && p.Limit > maxLimit {
		p.Limit = maxLimit
	}
	if p.Offset < 0 {
		p.Offset = 0
	}

	return p
}

// parseQueryInt parses an integer query parameter with a default value.
func parseQueryInt(c *fiber.Ctx, key string, defaultValue int) int {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return intVal
}

// parseQueryUUID parses a UUID query parameter.
// Returns nil if the parameter is empty or invalid.
func parseQueryUUID(c *fiber.Ctx, key string) *uuid.UUID {
	val := c.Query(key)
	if val == "" {
		return nil
	}
	id, err := uuid.Parse(val)
	if err != nil {
		return nil
	}
	return &id
}

// ErrorResponse represents a standardized error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// errorResponse creates a standardized JSON error response.
func errorResponse(c *fiber.Ctx, statusCode int, message string) error {
	errorName := "Error"
	switch statusCode {
	case fiber.StatusBadRequest:
		errorName = "Bad Request"
	case fiber.StatusUnauthorized:
		errorName = "Unauthorized"
	case fiber.StatusForbidden:
		errorName = "Forbidden"
	case fiber.StatusNotFound:
		errorName = "Not Found"
	case fiber.StatusConflict:
		errorName = "Conflict"
	case fiber.StatusInternalServerError:
		errorName = "Internal Server Error"
	}

	return c.Status(statusCode).JSON(ErrorResponse{
		Error:   errorName,
		Message: message,
	})
}
