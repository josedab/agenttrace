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
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/service"
	"github.com/agenttrace/agenttrace/api/internal/testutil"
)

// MockProjectService mocks the project service for testing.
type MockProjectService struct {
	mock.Mock
}

func (m *MockProjectService) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Project, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Project), args.Error(1)
}

func (m *MockProjectService) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]domain.Project, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Project), args.Error(1)
}

func (m *MockProjectService) Get(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *MockProjectService) Create(ctx context.Context, orgID uuid.UUID, input *service.ProjectInput, userID uuid.UUID) (*domain.Project, error) {
	args := m.Called(ctx, orgID, input, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *MockProjectService) Update(ctx context.Context, id uuid.UUID, input *service.ProjectInput) (*domain.Project, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *MockProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectService) AddMember(ctx context.Context, projectID, userID uuid.UUID, role domain.OrgRole) error {
	args := m.Called(ctx, projectID, userID, role)
	return args.Error(0)
}

func (m *MockProjectService) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	args := m.Called(ctx, projectID, userID)
	return args.Error(0)
}

func (m *MockProjectService) GetUserRole(ctx context.Context, projectID, userID uuid.UUID) (*domain.OrgRole, error) {
	args := m.Called(ctx, projectID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrgRole), args.Error(1)
}

func setupProjectsTestApp(mockSvc *MockProjectService, userID *uuid.UUID) *fiber.App {
	app := fiber.New()
	logger := zap.NewNop()

	if userID != nil {
		app.Use(testutil.TestUserMiddleware(*userID))
	}

	// ListProjects
	app.Get("/v1/projects", func(c *fiber.Ctx) error {
		uid, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "User ID not found",
			})
		}

		orgIDStr := c.Query("organizationId")
		var projects []domain.Project
		var err error

		if orgIDStr != "" {
			orgID, parseErr := uuid.Parse(orgIDStr)
			if parseErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Bad Request",
					"message": "Invalid organization ID",
				})
			}
			projects, err = mockSvc.ListByOrganization(c.Context(), orgID)
		} else {
			projects, err = mockSvc.ListByUser(c.Context(), uid)
		}

		if err != nil {
			logger.Error("failed to list projects", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to list projects",
			})
		}

		return c.JSON(fiber.Map{
			"data": projects,
		})
	})

	// GetProject
	app.Get("/v1/projects/:projectId", func(c *fiber.Ctx) error {
		projectID, err := uuid.Parse(c.Params("projectId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid project ID",
			})
		}

		project, err := mockSvc.Get(c.Context(), projectID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Project not found",
				})
			}
			logger.Error("failed to get project", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to get project",
			})
		}

		return c.JSON(project)
	})

	// CreateProject
	app.Post("/v1/projects", func(c *fiber.Ctx) error {
		uid, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "User ID not found",
			})
		}

		var input struct {
			OrganizationID string `json:"organizationId"`
			Name           string `json:"name"`
			Description    string `json:"description,omitempty"`
		}

		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body: " + err.Error(),
			})
		}

		if input.OrganizationID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "organizationId is required",
			})
		}

		orgID, err := uuid.Parse(input.OrganizationID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid organization ID",
			})
		}

		if input.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "name is required",
			})
		}

		projectInput := &service.ProjectInput{
			Name:        input.Name,
			Description: input.Description,
		}

		project, err := mockSvc.Create(c.Context(), orgID, projectInput, uid)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Organization not found",
				})
			}
			logger.Error("failed to create project", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to create project",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(project)
	})

	// UpdateProject
	app.Patch("/v1/projects/:projectId", func(c *fiber.Ctx) error {
		projectID, err := uuid.Parse(c.Params("projectId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid project ID",
			})
		}

		var input service.ProjectInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body: " + err.Error(),
			})
		}

		project, err := mockSvc.Update(c.Context(), projectID, &input)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Project not found",
				})
			}
			logger.Error("failed to update project", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to update project",
			})
		}

		return c.JSON(project)
	})

	// DeleteProject
	app.Delete("/v1/projects/:projectId", func(c *fiber.Ctx) error {
		projectID, err := uuid.Parse(c.Params("projectId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid project ID",
			})
		}

		if err := mockSvc.Delete(c.Context(), projectID); err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Project not found",
				})
			}
			logger.Error("failed to delete project", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to delete project",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	// GetUserRole
	app.Get("/v1/projects/:projectId/role", func(c *fiber.Ctx) error {
		projectID, err := uuid.Parse(c.Params("projectId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid project ID",
			})
		}

		uid, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "User ID not found",
			})
		}

		role, err := mockSvc.GetUserRole(c.Context(), projectID, uid)
		if err != nil {
			logger.Error("failed to get user role", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to get user role",
			})
		}

		if role == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "User not a member of this project",
			})
		}

		return c.JSON(fiber.Map{
			"role": role,
		})
	})

	return app
}

// --- ListProjects Tests ---

func TestProjectsHandler_ListProjects(t *testing.T) {
	t.Run("successfully lists projects for user", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		orgID := uuid.New()
		expectedProjects := []domain.Project{
			{
				ID:             uuid.New(),
				Name:           "Project 1",
				OrganizationID: orgID,
				CreatedAt:      time.Now(),
			},
			{
				ID:             uuid.New(),
				Name:           "Project 2",
				OrganizationID: orgID,
				CreatedAt:      time.Now(),
			},
		}

		mockSvc.On("ListByUser", mock.Anything, userID).Return(expectedProjects, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/projects", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].([]interface{})
		assert.Len(t, data, 2)

		mockSvc.AssertExpectations(t)
	})

	t.Run("successfully lists projects by organization", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		orgID := uuid.New()
		expectedProjects := []domain.Project{
			{
				ID:             uuid.New(),
				Name:           "Org Project",
				OrganizationID: orgID,
				CreatedAt:      time.Now(),
			},
		}

		mockSvc.On("ListByOrganization", mock.Anything, orgID).Return(expectedProjects, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/projects?organizationId="+orgID.String(), nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for invalid organization ID", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		req := httptest.NewRequest(http.MethodGet, "/v1/projects?organizationId=invalid", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 401 when user ID not found", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		app := setupProjectsTestApp(mockSvc, nil) // No user ID

		req := httptest.NewRequest(http.MethodGet, "/v1/projects", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		mockSvc.On("ListByUser", mock.Anything, userID).Return(nil, assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/v1/projects", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

// --- GetProject Tests ---

func TestProjectsHandler_GetProject(t *testing.T) {
	t.Run("successfully gets project", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		projectID := uuid.New()
		expectedProject := &domain.Project{
			ID:             projectID,
			Name:           "Test Project",
			OrganizationID: uuid.New(),
			CreatedAt:      time.Now(),
		}

		mockSvc.On("Get", mock.Anything, projectID).Return(expectedProject, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/projects/"+projectID.String(), nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.Project
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "Test Project", result.Name)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for invalid project ID", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		req := httptest.NewRequest(http.MethodGet, "/v1/projects/invalid-uuid", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 404 when project not found", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		projectID := uuid.New()
		mockSvc.On("Get", mock.Anything, projectID).Return(nil, apperrors.NotFound("project not found"))

		req := httptest.NewRequest(http.MethodGet, "/v1/projects/"+projectID.String(), nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

// --- CreateProject Tests ---

func TestProjectsHandler_CreateProject(t *testing.T) {
	t.Run("successfully creates project", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		orgID := uuid.New()
		projectID := uuid.New()
		expectedProject := &domain.Project{
			ID:             projectID,
			Name:           "New Project",
			OrganizationID: orgID,
			CreatedAt:      time.Now(),
		}

		mockSvc.On("Create", mock.Anything, orgID, mock.MatchedBy(func(input *service.ProjectInput) bool {
			return input.Name == "New Project"
		}), userID).Return(expectedProject, nil)

		body, _ := json.Marshal(map[string]string{
			"organizationId": orgID.String(),
			"name":           "New Project",
			"description":    "A new project",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/projects", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result domain.Project
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "New Project", result.Name)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing organizationId", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		body, _ := json.Marshal(map[string]string{
			"name": "New Project",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/projects", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "organizationId is required", result["message"])
	})

	t.Run("returns 400 for missing name", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		body, _ := json.Marshal(map[string]string{
			"organizationId": uuid.New().String(),
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/projects", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "name is required", result["message"])
	})

	t.Run("returns 401 when user ID not found", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		app := setupProjectsTestApp(mockSvc, nil) // No user ID

		body, _ := json.Marshal(map[string]string{
			"organizationId": uuid.New().String(),
			"name":           "New Project",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/projects", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 404 when organization not found", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		orgID := uuid.New()
		mockSvc.On("Create", mock.Anything, orgID, mock.Anything, userID).
			Return(nil, apperrors.NotFound("organization not found"))

		body, _ := json.Marshal(map[string]string{
			"organizationId": orgID.String(),
			"name":           "New Project",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/projects", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

// --- UpdateProject Tests ---

func TestProjectsHandler_UpdateProject(t *testing.T) {
	t.Run("successfully updates project", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		projectID := uuid.New()
		expectedProject := &domain.Project{
			ID:        projectID,
			Name:      "Updated Project",
			UpdatedAt: time.Now(),
		}

		mockSvc.On("Update", mock.Anything, projectID, mock.Anything).Return(expectedProject, nil)

		body, _ := json.Marshal(map[string]string{
			"name": "Updated Project",
		})
		req := httptest.NewRequest(http.MethodPatch, "/v1/projects/"+projectID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 404 when project not found", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		projectID := uuid.New()
		mockSvc.On("Update", mock.Anything, projectID, mock.Anything).
			Return(nil, apperrors.NotFound("project not found"))

		body, _ := json.Marshal(map[string]string{
			"name": "Updated Project",
		})
		req := httptest.NewRequest(http.MethodPatch, "/v1/projects/"+projectID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

// --- DeleteProject Tests ---

func TestProjectsHandler_DeleteProject(t *testing.T) {
	t.Run("successfully deletes project", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		projectID := uuid.New()
		mockSvc.On("Delete", mock.Anything, projectID).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/v1/projects/"+projectID.String(), nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 404 when project not found", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		projectID := uuid.New()
		mockSvc.On("Delete", mock.Anything, projectID).Return(apperrors.NotFound("project not found"))

		req := httptest.NewRequest(http.MethodDelete, "/v1/projects/"+projectID.String(), nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

// --- GetUserRole Tests ---

func TestProjectsHandler_GetUserRole(t *testing.T) {
	t.Run("successfully gets user role", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		projectID := uuid.New()
		role := domain.OrgRoleAdmin
		mockSvc.On("GetUserRole", mock.Anything, projectID, userID).Return(&role, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/projects/"+projectID.String()+"/role", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, string(domain.OrgRoleAdmin), result["role"])

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 404 when user not a member", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		userID := uuid.New()
		app := setupProjectsTestApp(mockSvc, &userID)

		projectID := uuid.New()
		mockSvc.On("GetUserRole", mock.Anything, projectID, userID).Return(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/projects/"+projectID.String()+"/role", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 401 when user ID not found", func(t *testing.T) {
		mockSvc := new(MockProjectService)
		app := setupProjectsTestApp(mockSvc, nil) // No user ID

		projectID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/v1/projects/"+projectID.String()+"/role", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
