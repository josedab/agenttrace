package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ImportHandler handles import endpoints
type ImportHandler struct {
	datasetService *service.DatasetService
	promptService  *service.PromptService
	logger         *zap.Logger
}

// NewImportHandler creates a new import handler
func NewImportHandler(
	datasetService *service.DatasetService,
	promptService *service.PromptService,
	logger *zap.Logger,
) *ImportHandler {
	return &ImportHandler{
		datasetService: datasetService,
		promptService:  promptService,
		logger:         logger,
	}
}

// ImportDatasetRequest represents a request to import a dataset
type ImportDatasetRequest struct {
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	Items       []DatasetItemImport  `json:"items"`
	Metadata    map[string]any       `json:"metadata,omitempty"`
}

// DatasetItemImport represents an imported dataset item
type DatasetItemImport struct {
	Input          any            `json:"input"`
	ExpectedOutput any            `json:"expectedOutput,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

// ImportPromptRequest represents a request to import a prompt
type ImportPromptRequest struct {
	Name        string         `json:"name"`
	Content     string         `json:"content"`
	Description *string        `json:"description,omitempty"`
	Config      map[string]any `json:"config,omitempty"`
	Labels      []string       `json:"labels,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
}

// ImportResponse represents an import response
type ImportResponse struct {
	ID           string `json:"id"`
	ImportedCount int    `json:"importedCount"`
	Message      string `json:"message,omitempty"`
}

// ImportDataset handles POST /v1/import/dataset
func (h *ImportHandler) ImportDataset(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var request ImportDatasetRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// Validate request
	if request.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "name is required",
		})
	}

	if len(request.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "at least one item is required",
		})
	}

	// Create dataset
	datasetInput := &domain.DatasetInput{
		Name:        request.Name,
		Description: request.Description,
		Metadata:    request.Metadata,
	}

	dataset, err := h.datasetService.Create(c.Context(), projectID, datasetInput)
	if err != nil {
		h.logger.Error("failed to create dataset for import", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create dataset",
		})
	}

	// Add items
	importedCount := 0
	for _, item := range request.Items {
		itemInput := &domain.DatasetItemInput{
			Input:          item.Input,
			ExpectedOutput: item.ExpectedOutput,
			Metadata:       item.Metadata,
		}

		_, err := h.datasetService.AddItem(c.Context(), dataset.ID, itemInput)
		if err != nil {
			h.logger.Warn("failed to import dataset item",
				zap.Error(err),
				zap.String("dataset_id", dataset.ID.String()),
			)
			continue
		}
		importedCount++
	}

	h.logger.Info("dataset imported",
		zap.String("dataset_id", dataset.ID.String()),
		zap.Int("imported_count", importedCount),
		zap.Int("total_items", len(request.Items)),
	)

	return c.Status(fiber.StatusCreated).JSON(ImportResponse{
		ID:           dataset.ID.String(),
		ImportedCount: importedCount,
		Message:      fmt.Sprintf("Successfully imported %d items", importedCount),
	})
}

// ImportDatasetCSV handles POST /v1/import/dataset/csv
func (h *ImportHandler) ImportDatasetCSV(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	// Get dataset name from query
	datasetName := c.Query("name")
	if datasetName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "name query parameter is required",
		})
	}

	// Get CSV file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "file is required",
		})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Failed to open file",
		})
	}
	defer f.Close()

	// Parse CSV
	reader := csv.NewReader(f)
	headers, err := reader.Read()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Failed to read CSV headers",
		})
	}

	// Find column indices
	inputIdx := -1
	outputIdx := -1
	metadataIdx := -1

	for i, h := range headers {
		switch strings.ToLower(h) {
		case "input":
			inputIdx = i
		case "expected_output", "expectedoutput", "output":
			outputIdx = i
		case "metadata":
			metadataIdx = i
		}
	}

	if inputIdx == -1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "CSV must have an 'input' column",
		})
	}

	// Create dataset
	dataset, err := h.datasetService.Create(c.Context(), projectID, &domain.DatasetInput{
		Name: datasetName,
	})
	if err != nil {
		h.logger.Error("failed to create dataset for CSV import", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create dataset",
		})
	}

	// Import rows
	importedCount := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		var input, output any
		var metadata map[string]any

		// Get input
		if inputIdx < len(row) {
			input = row[inputIdx]
		}

		// Get expected output
		if outputIdx >= 0 && outputIdx < len(row) {
			output = row[outputIdx]
		}

		// Get metadata
		if metadataIdx >= 0 && metadataIdx < len(row) && row[metadataIdx] != "" {
			_ = json.Unmarshal([]byte(row[metadataIdx]), &metadata)
		}

		itemInput := &domain.DatasetItemInput{
			Input:          input,
			ExpectedOutput: output,
			Metadata:       metadata,
		}

		_, err = h.datasetService.AddItem(c.Context(), dataset.ID, itemInput)
		if err != nil {
			continue
		}
		importedCount++
	}

	h.logger.Info("CSV dataset imported",
		zap.String("dataset_id", dataset.ID.String()),
		zap.Int("imported_count", importedCount),
	)

	return c.Status(fiber.StatusCreated).JSON(ImportResponse{
		ID:           dataset.ID.String(),
		ImportedCount: importedCount,
		Message:      fmt.Sprintf("Successfully imported %d items from CSV", importedCount),
	})
}

// ImportPrompt handles POST /v1/import/prompt
func (h *ImportHandler) ImportPrompt(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var request ImportPromptRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// Validate request
	if request.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "name is required",
		})
	}

	if request.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "content is required",
		})
	}

	// Get user ID (for API key auth, use a system user ID)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		userID = uuid.Nil
	}

	// Create prompt
	promptInput := &domain.PromptInput{
		Name:        request.Name,
		Content:     request.Content,
		Description: request.Description,
		Config:      request.Config,
		Labels:      request.Labels,
		Tags:        request.Tags,
	}

	prompt, err := h.promptService.Create(c.Context(), projectID, promptInput, userID)
	if err != nil {
		h.logger.Error("failed to create prompt for import", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create prompt",
		})
	}

	h.logger.Info("prompt imported",
		zap.String("prompt_id", prompt.ID.String()),
		zap.String("prompt_name", prompt.Name),
	)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":      prompt.ID.String(),
		"name":    prompt.Name,
		"message": "Prompt imported successfully",
	})
}

// ImportOpenAIFinetune handles POST /v1/import/dataset/openai-finetune
// Imports a dataset from OpenAI fine-tune JSONL format
func (h *ImportHandler) ImportOpenAIFinetune(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	datasetName := c.Query("name")
	if datasetName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "name query parameter is required",
		})
	}

	// Get JSONL file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "file is required",
		})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Failed to open file",
		})
	}
	defer f.Close()

	// Create dataset
	dataset, err := h.datasetService.Create(c.Context(), projectID, &domain.DatasetInput{
		Name: datasetName,
		Metadata: map[string]any{
			"format": "openai_finetune",
		},
	})
	if err != nil {
		h.logger.Error("failed to create dataset for OpenAI import", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create dataset",
		})
	}

	// Read file content
	content, err := io.ReadAll(f)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Failed to read file",
		})
	}

	// Parse JSONL
	lines := strings.Split(string(content), "\n")
	importedCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var record struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}

		if err := json.Unmarshal([]byte(line), &record); err != nil {
			continue
		}

		// Extract input (user messages) and output (assistant messages)
		var input, output string
		var systemPrompt string

		for _, msg := range record.Messages {
			switch msg.Role {
			case "system":
				systemPrompt = msg.Content
			case "user":
				input = msg.Content
			case "assistant":
				output = msg.Content
			}
		}

		if input == "" {
			continue
		}

		metadata := map[string]any{}
		if systemPrompt != "" {
			metadata["system"] = systemPrompt
		}

		itemInput := &domain.DatasetItemInput{
			Input:          input,
			ExpectedOutput: output,
			Metadata:       metadata,
		}

		_, err = h.datasetService.AddItem(c.Context(), dataset.ID, itemInput)
		if err != nil {
			continue
		}
		importedCount++
	}

	h.logger.Info("OpenAI fine-tune dataset imported",
		zap.String("dataset_id", dataset.ID.String()),
		zap.Int("imported_count", importedCount),
	)

	return c.Status(fiber.StatusCreated).JSON(ImportResponse{
		ID:           dataset.ID.String(),
		ImportedCount: importedCount,
		Message:      fmt.Sprintf("Successfully imported %d items from OpenAI fine-tune format", importedCount),
	})
}

// RegisterRoutes registers import routes
func (h *ImportHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())

	v1.Post("/import/dataset", h.ImportDataset)
	v1.Post("/import/dataset/csv", h.ImportDatasetCSV)
	v1.Post("/import/dataset/openai-finetune", h.ImportOpenAIFinetune)
	v1.Post("/import/prompt", h.ImportPrompt)
}
