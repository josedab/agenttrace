package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/testutil"
)

// MockOrgService mocks the org service for testing
type MockOrgService struct {
	mock.Mock
}

func (m *MockOrgService) Get(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *MockOrgService) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *MockOrgService) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Organization, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Organization), args.Error(1)
}

func (m *MockOrgService) Create(ctx context.Context, name string, ownerID uuid.UUID) (*domain.Organization, error) {
	args := m.Called(ctx, name, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *MockOrgService) Update(ctx context.Context, id uuid.UUID, name string) (*domain.Organization, error) {
	args := m.Called(ctx, id, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *MockOrgService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrgService) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*domain.OrganizationMember, error) {
	args := m.Called(ctx, orgID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrganizationMember), args.Error(1)
}

func setupOrgTestApp(mockSvc *MockOrgService, userID *uuid.UUID) *fiber.App {
	app := fiber.New()
	logger := zap.NewNop()

	if userID != nil {
		app.Use(testutil.TestUserMiddleware(*userID))
	}

	// ListOrganizations
	app.Get("/v1/organizations", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(uuid.UUID)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "User ID not found",
			})
		}

		orgs, err := mockSvc.ListByUser(c.Context(), userID)
		if err != nil {
			logger.Error("failed to list organizations", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to list organizations",
			})
		}

		return c.JSON(fiber.Map{
			"data": orgs,
		})
	})

	// GetOrganization
	app.Get("/v1/organizations/:orgId", func(c *fiber.Ctx) error {
		orgID, err := uuid.Parse(c.Params("orgId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid organization ID",
			})
		}

		org, err := mockSvc.Get(c.Context(), orgID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Organization not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to get organization",
			})
		}

		return c.JSON(org)
	})

	// GetOrganizationBySlug
	app.Get("/v1/organizations/slug/:slug", func(c *fiber.Ctx) error {
		slug := c.Params("slug")
		if slug == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Organization slug required",
			})
		}

		org, err := mockSvc.GetBySlug(c.Context(), slug)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Organization not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to get organization",
			})
		}

		return c.JSON(org)
	})

	// CreateOrganization
	app.Post("/v1/organizations", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(uuid.UUID)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "User ID not found",
			})
		}

		var input struct {
			Name string `json:"name"`
		}

		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body: " + err.Error(),
			})
		}

		if input.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "name is required",
			})
		}

		org, err := mockSvc.Create(c.Context(), input.Name, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to create organization",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(org)
	})

	// UpdateOrganization
	app.Patch("/v1/organizations/:orgId", func(c *fiber.Ctx) error {
		orgID, err := uuid.Parse(c.Params("orgId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid organization ID",
			})
		}

		var input struct {
			Name string `json:"name"`
		}

		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body: " + err.Error(),
			})
		}

		org, err := mockSvc.Update(c.Context(), orgID, input.Name)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Organization not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to update organization",
			})
		}

		return c.JSON(org)
	})

	// DeleteOrganization
	app.Delete("/v1/organizations/:orgId", func(c *fiber.Ctx) error {
		orgID, err := uuid.Parse(c.Params("orgId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid organization ID",
			})
		}

		if err := mockSvc.Delete(c.Context(), orgID); err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Organization not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to delete organization",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	// GetMember
	app.Get("/v1/organizations/:orgId/members/:userId", func(c *fiber.Ctx) error {
		orgID, err := uuid.Parse(c.Params("orgId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid organization ID",
			})
		}

		userID, err := uuid.Parse(c.Params("userId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid user ID",
			})
		}

		member, err := mockSvc.GetMember(c.Context(), orgID, userID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Member not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to get member",
			})
		}

		return c.JSON(member)
	})

	return app
}

// --- ListOrganizations Tests ---

func TestOrganizationsHandler_ListOrganizations(t *testing.T) {
	t.Run("successfully lists organizations", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		orgs := []domain.Organization{
			{ID: uuid.New(), Name: "Org 1", Slug: "org-1"},
			{ID: uuid.New(), Name: "Org 2", Slug: "org-2"},
		}

		mockSvc.On("ListByUser", mock.Anything, userID).Return(orgs, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data := result["data"].([]interface{})
		assert.Len(t, data, 2)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 401 for unauthenticated user", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		app := setupOrgTestApp(mockSvc, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 500 for service error", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		mockSvc.On("ListByUser", mock.Anything, userID).Return(nil, assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

// --- GetOrganization Tests ---

func TestOrganizationsHandler_GetOrganization(t *testing.T) {
	t.Run("successfully gets organization", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		orgID := uuid.New()
		org := &domain.Organization{
			ID:        orgID,
			Name:      "Test Org",
			Slug:      "test-org",
			CreatedAt: time.Now(),
		}

		mockSvc.On("Get", mock.Anything, orgID).Return(org, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/"+orgID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.Organization
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "Test Org", result.Name)
		assert.Equal(t, "test-org", result.Slug)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for invalid UUID", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/invalid-uuid", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 404 for non-existent organization", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		orgID := uuid.New()
		mockSvc.On("Get", mock.Anything, orgID).Return(nil, apperrors.NotFound("not found"))

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/"+orgID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

// --- GetOrganizationBySlug Tests ---

func TestOrganizationsHandler_GetOrganizationBySlug(t *testing.T) {
	t.Run("successfully gets organization by slug", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		org := &domain.Organization{
			ID:   uuid.New(),
			Name: "Test Org",
			Slug: "test-org",
		}

		mockSvc.On("GetBySlug", mock.Anything, "test-org").Return(org, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/slug/test-org", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.Organization
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "test-org", result.Slug)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 404 for non-existent slug", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		mockSvc.On("GetBySlug", mock.Anything, "nonexistent").Return(nil, apperrors.NotFound("not found"))

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/slug/nonexistent", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

// --- CreateOrganization Tests ---

func TestOrganizationsHandler_CreateOrganization(t *testing.T) {
	t.Run("successfully creates organization", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		expectedOrg := &domain.Organization{
			ID:   uuid.New(),
			Name: "New Org",
			Slug: "new-org",
		}

		mockSvc.On("Create", mock.Anything, "New Org", userID).Return(expectedOrg, nil)

		body, _ := json.Marshal(map[string]string{"name": "New Org"})
		req := httptest.NewRequest(http.MethodPost, "/v1/organizations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result domain.Organization
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "New Org", result.Name)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing name", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		body, _ := json.Marshal(map[string]string{})
		req := httptest.NewRequest(http.MethodPost, "/v1/organizations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "name is required", result["message"])
	})

	t.Run("returns 401 for unauthenticated user", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		app := setupOrgTestApp(mockSvc, nil)

		body, _ := json.Marshal(map[string]string{"name": "New Org"})
		req := httptest.NewRequest(http.MethodPost, "/v1/organizations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// --- UpdateOrganization Tests ---

func TestOrganizationsHandler_UpdateOrganization(t *testing.T) {
	t.Run("successfully updates organization", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		orgID := uuid.New()
		updatedOrg := &domain.Organization{
			ID:   orgID,
			Name: "Updated Name",
			Slug: "updated-name",
		}

		mockSvc.On("Update", mock.Anything, orgID, "Updated Name").Return(updatedOrg, nil)

		body, _ := json.Marshal(map[string]string{"name": "Updated Name"})
		req := httptest.NewRequest(http.MethodPatch, "/v1/organizations/"+orgID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.Organization
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "Updated Name", result.Name)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for invalid UUID", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		body, _ := json.Marshal(map[string]string{"name": "New Name"})
		req := httptest.NewRequest(http.MethodPatch, "/v1/organizations/invalid", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 404 for non-existent organization", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		orgID := uuid.New()
		mockSvc.On("Update", mock.Anything, orgID, mock.Anything).Return(nil, apperrors.NotFound("not found"))

		body, _ := json.Marshal(map[string]string{"name": "New Name"})
		req := httptest.NewRequest(http.MethodPatch, "/v1/organizations/"+orgID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

// --- DeleteOrganization Tests ---

func TestOrganizationsHandler_DeleteOrganization(t *testing.T) {
	t.Run("successfully deletes organization", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		orgID := uuid.New()
		mockSvc.On("Delete", mock.Anything, orgID).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/v1/organizations/"+orgID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for invalid UUID", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		req := httptest.NewRequest(http.MethodDelete, "/v1/organizations/invalid", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 404 for non-existent organization", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		orgID := uuid.New()
		mockSvc.On("Delete", mock.Anything, orgID).Return(apperrors.NotFound("not found"))

		req := httptest.NewRequest(http.MethodDelete, "/v1/organizations/"+orgID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

// --- GetMember Tests ---

func TestOrganizationsHandler_GetMember(t *testing.T) {
	t.Run("successfully gets member", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		authUserID := uuid.New()
		app := setupOrgTestApp(mockSvc, &authUserID)

		orgID := uuid.New()
		memberUserID := uuid.New()
		member := &domain.OrganizationMember{
			OrganizationID: orgID,
			UserID:         memberUserID,
			Role:           "admin",
		}

		mockSvc.On("GetMember", mock.Anything, orgID, memberUserID).Return(member, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/"+orgID.String()+"/members/"+memberUserID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.OrganizationMember
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, domain.OrgRoleAdmin, result.Role)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for invalid org ID", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		memberUserID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/invalid/members/"+memberUserID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 400 for invalid user ID", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		userID := uuid.New()
		app := setupOrgTestApp(mockSvc, &userID)

		orgID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/"+orgID.String()+"/members/invalid", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 404 for non-existent member", func(t *testing.T) {
		mockSvc := new(MockOrgService)
		authUserID := uuid.New()
		app := setupOrgTestApp(mockSvc, &authUserID)

		orgID := uuid.New()
		memberUserID := uuid.New()
		mockSvc.On("GetMember", mock.Anything, orgID, memberUserID).Return(nil, apperrors.NotFound("not found"))

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/"+orgID.String()+"/members/"+memberUserID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

func TestNewOrganizationsHandler(t *testing.T) {
	t.Run("creates handler with dependencies", func(t *testing.T) {
		logger := zap.NewNop()
		handler := NewOrganizationsHandler(nil, logger)

		require.NotNil(t, handler)
		assert.Equal(t, logger, handler.logger)
	})
}
