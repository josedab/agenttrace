package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// WorkflowSimulatorHandler handles workflow simulator HTTP requests
type WorkflowSimulatorHandler struct {
	logger  *zap.Logger
	service *service.WorkflowSimulatorService
}

// NewWorkflowSimulatorHandler creates a new workflow simulator handler
func NewWorkflowSimulatorHandler(svc *service.WorkflowSimulatorService, logger *zap.Logger) *WorkflowSimulatorHandler {
	return &WorkflowSimulatorHandler{
		logger:  logger,
		service: svc,
	}
}

// ListWorkflows handles GET /api/public/workflow-simulator/workflows
// @Summary List workflows
// @Description List all workflow definitions for a project
// @Tags workflow-simulator
// @Accept json
// @Produce json
// @Success 200 {object} domain.WorkflowList
// @Failure 401 {object} map[string]string
// @Router /api/public/workflow-simulator/workflows [get]
func (h *WorkflowSimulatorHandler) ListWorkflows(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	result, err := h.service.ListWorkflows(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list workflows", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list workflows"})
	}

	return c.JSON(result)
}

// CreateWorkflow handles POST /api/public/workflow-simulator/workflows
// @Summary Create workflow
// @Description Create a new workflow definition
// @Tags workflow-simulator
// @Accept json
// @Produce json
// @Param workflow body domain.WorkflowDefinitionInput true "Workflow definition"
// @Success 201 {object} domain.WorkflowDefinition
// @Failure 400 {object} map[string]string
// @Router /api/public/workflow-simulator/workflows [post]
func (h *WorkflowSimulatorHandler) CreateWorkflow(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.WorkflowDefinitionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	workflow, err := h.service.CreateWorkflow(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to create workflow", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create workflow"})
	}

	return c.Status(fiber.StatusCreated).JSON(workflow)
}

// GetWorkflow handles GET /api/public/workflow-simulator/workflows/:workflowId
// @Summary Get workflow
// @Description Get a specific workflow definition by ID
// @Tags workflow-simulator
// @Accept json
// @Produce json
// @Param workflowId path string true "Workflow ID"
// @Success 200 {object} domain.WorkflowDefinition
// @Failure 400 {object} map[string]string
// @Router /api/public/workflow-simulator/workflows/{workflowId} [get]
func (h *WorkflowSimulatorHandler) GetWorkflow(c *fiber.Ctx) error {
	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid workflow ID"})
	}

	workflow, err := h.service.GetWorkflow(c.Context(), workflowID)
	if err != nil {
		h.logger.Error("failed to get workflow", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get workflow"})
	}

	return c.JSON(workflow)
}

// UpdateWorkflow handles PUT /api/public/workflow-simulator/workflows/:workflowId
// @Summary Update workflow
// @Description Update a workflow definition
// @Tags workflow-simulator
// @Accept json
// @Produce json
// @Param workflowId path string true "Workflow ID"
// @Param workflow body domain.WorkflowDefinitionInput true "Updated workflow definition"
// @Success 200 {object} domain.WorkflowDefinition
// @Failure 400 {object} map[string]string
// @Router /api/public/workflow-simulator/workflows/{workflowId} [put]
func (h *WorkflowSimulatorHandler) UpdateWorkflow(c *fiber.Ctx) error {
	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid workflow ID"})
	}

	var input domain.WorkflowDefinitionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	workflow, err := h.service.UpdateWorkflow(c.Context(), workflowID, input)
	if err != nil {
		h.logger.Error("failed to update workflow", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update workflow"})
	}

	return c.JSON(workflow)
}

// DeleteWorkflow handles DELETE /api/public/workflow-simulator/workflows/:workflowId
// @Summary Delete workflow
// @Description Delete a workflow definition
// @Tags workflow-simulator
// @Accept json
// @Produce json
// @Param workflowId path string true "Workflow ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Router /api/public/workflow-simulator/workflows/{workflowId} [delete]
func (h *WorkflowSimulatorHandler) DeleteWorkflow(c *fiber.Ctx) error {
	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid workflow ID"})
	}

	if err := h.service.DeleteWorkflow(c.Context(), workflowID); err != nil {
		h.logger.Error("failed to delete workflow", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete workflow"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RunSimulation handles POST /api/public/workflow-simulator/simulations
// @Summary Run simulation
// @Description Run a workflow simulation
// @Tags workflow-simulator
// @Accept json
// @Produce json
// @Param simulation body domain.SimulationInput true "Simulation input"
// @Success 201 {object} domain.WorkflowSimulation
// @Failure 400 {object} map[string]string
// @Router /api/public/workflow-simulator/simulations [post]
func (h *WorkflowSimulatorHandler) RunSimulation(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.SimulationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	simulation, err := h.service.RunSimulation(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to run simulation", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to run simulation"})
	}

	return c.Status(fiber.StatusCreated).JSON(simulation)
}

// GetSimulation handles GET /api/public/workflow-simulator/simulations/:simulationId
// @Summary Get simulation
// @Description Get a specific simulation by ID
// @Tags workflow-simulator
// @Accept json
// @Produce json
// @Param simulationId path string true "Simulation ID"
// @Success 200 {object} domain.WorkflowSimulation
// @Failure 400 {object} map[string]string
// @Router /api/public/workflow-simulator/simulations/{simulationId} [get]
func (h *WorkflowSimulatorHandler) GetSimulation(c *fiber.Ctx) error {
	simulationID, err := uuid.Parse(c.Params("simulationId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid simulation ID"})
	}

	simulation, err := h.service.GetSimulation(c.Context(), simulationID)
	if err != nil {
		h.logger.Error("failed to get simulation", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get simulation"})
	}

	return c.JSON(simulation)
}

// ListSimulations handles GET /api/public/workflow-simulator/simulations
// @Summary List simulations
// @Description List simulations for a workflow
// @Tags workflow-simulator
// @Accept json
// @Produce json
// @Param workflowId query string true "Workflow ID"
// @Success 200 {object} domain.SimulationList
// @Failure 400 {object} map[string]string
// @Router /api/public/workflow-simulator/simulations [get]
func (h *WorkflowSimulatorHandler) ListSimulations(c *fiber.Ctx) error {
	workflowIDStr := c.Query("workflowId")
	if workflowIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "workflowId query parameter is required"})
	}

	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid workflow ID"})
	}

	result, err := h.service.ListSimulations(c.Context(), workflowID)
	if err != nil {
		h.logger.Error("failed to list simulations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list simulations"})
	}

	return c.JSON(result)
}

// ValidateWorkflow handles POST /api/public/workflow-simulator/workflows/validate
// @Summary Validate workflow
// @Description Validate a workflow definition
// @Tags workflow-simulator
// @Accept json
// @Produce json
// @Param workflow body domain.WorkflowDefinitionInput true "Workflow definition to validate"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/public/workflow-simulator/workflows/validate [post]
func (h *WorkflowSimulatorHandler) ValidateWorkflow(c *fiber.Ctx) error {
	var input domain.WorkflowDefinitionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	def := &domain.WorkflowDefinition{
		Name:  input.Name,
		Nodes: input.Nodes,
		Edges: input.Edges,
	}

	errors := h.service.ValidateWorkflow(def)

	return c.JSON(fiber.Map{
		"valid":  len(errors) == 0,
		"errors": errors,
	})
}
