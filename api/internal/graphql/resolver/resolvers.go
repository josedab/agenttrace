package resolver

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/agenttrace/agenttrace/api"
	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/graphql/generated"
	"github.com/agenttrace/agenttrace/api/internal/graphql/model"
	"github.com/agenttrace/agenttrace/api/internal/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ContextKey is the key type for context values
type ContextKey string

const (
	// ContextKeyProjectID is the context key for project ID
	ContextKeyProjectID ContextKey = "projectID"
	// ContextKeyUserID is the context key for user ID
	ContextKeyUserID ContextKey = "userID"
	// ContextKeyOrgID is the context key for organization ID
	ContextKeyOrgID ContextKey = "orgID"
)

// Resolver holds all service dependencies for GraphQL resolvers
type Resolver struct {
	Logger           *zap.Logger
	QueryService     *service.QueryService
	IngestionService *service.IngestionService
	ScoreService     *service.ScoreService
	PromptService    *service.PromptService
	DatasetService   *service.DatasetService
	EvalService      *service.EvalService
	AuthService      *service.AuthService
	OrgService       *service.OrgService
	ProjectService   *service.ProjectService
	CostService      *service.CostService
}

// NewResolver creates a new resolver with all dependencies
func NewResolver(
	logger *zap.Logger,
	queryService *service.QueryService,
	ingestionService *service.IngestionService,
	scoreService *service.ScoreService,
	promptService *service.PromptService,
	datasetService *service.DatasetService,
	evalService *service.EvalService,
	authService *service.AuthService,
	orgService *service.OrgService,
	projectService *service.ProjectService,
	costService *service.CostService,
) *Resolver {
	return &Resolver{
		Logger:           logger.Named("graphql"),
		QueryService:     queryService,
		IngestionService: ingestionService,
		ScoreService:     scoreService,
		PromptService:    promptService,
		DatasetService:   datasetService,
		EvalService:      evalService,
		AuthService:      authService,
		OrgService:       orgService,
		ProjectService:   projectService,
		CostService:      costService,
	}
}

// Helper functions

// getProjectID extracts project ID from context
func getProjectID(ctx context.Context) (uuid.UUID, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok || projectID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("project ID not found in context")
	}
	return projectID, nil
}

// getUserID extracts user ID from context
func getUserID(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(ContextKeyUserID).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// encodeCursor encodes an offset as a cursor
func encodeCursor(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

// decodeCursor decodes a cursor to an offset
func decodeCursor(cursor string) (int, error) {
	data, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

// getLimit returns the limit with default
func getLimit(limit *int) int {
	if limit == nil || *limit <= 0 {
		return 20
	}
	if *limit > 100 {
		return 100
	}
	return *limit
}

// getOffset calculates offset from cursor
func getOffset(cursor *string) int {
	if cursor == nil || *cursor == "" {
		return 0
	}
	offset, err := decodeCursor(*cursor)
	if err != nil {
		return 0
	}
	return offset
}

// parseJSONString parses a JSON string into map[string]any
func parseJSONString(s string) (map[string]any, error) {
	if s == "" {
		return nil, nil
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// parseJSONStringPtr parses a JSON string pointer into map[string]any
func parseJSONStringPtr(s *string) (map[string]any, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	return parseJSONString(*s)
}

// ======== APIKey Resolver ========

// DisplayKey is the resolver for the displayKey field.
func (r *aPIKeyResolver) DisplayKey(ctx context.Context, obj *domain.APIKey) (string, error) {
	// Display key shows the secret key preview
	if obj.SecretKeyPreview != "" {
		return obj.SecretKeyPreview, nil
	}
	// Fallback to public key if preview not available
	if len(obj.PublicKey) < 8 {
		return "****", nil
	}
	return obj.PublicKey[:4] + "...", nil
}

// ======== AnnotationQueue Resolver ========

// ScoreConfig is the resolver for the scoreConfig field.
func (r *annotationQueueResolver) ScoreConfig(ctx context.Context, obj *domain.AnnotationQueue) (map[string]any, error) {
	return parseJSONString(obj.ScoreConfig)
}

// ======== AnnotationQueueItem Resolver ========

// Status is the resolver for the status field.
func (r *annotationQueueItemResolver) Status(ctx context.Context, obj *domain.AnnotationQueueItem) (api.AnnotationStatus, error) {
	return api.AnnotationStatus(obj.Status), nil
}

// AssignedUserID is the resolver for the assignedUserId field.
func (r *annotationQueueItemResolver) AssignedUserID(ctx context.Context, obj *domain.AnnotationQueueItem) (*uuid.UUID, error) {
	// CompletedBy is used as the assigned user ID
	return obj.CompletedBy, nil
}

// Trace is the resolver for the trace field.
func (r *annotationQueueItemResolver) Trace(ctx context.Context, obj *domain.AnnotationQueueItem) (*domain.Trace, error) {
	if obj.TraceID == "" {
		return nil, nil
	}
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}
	return r.QueryService.GetTrace(ctx, projectID, obj.TraceID)
}

// Observation is the resolver for the observation field.
func (r *annotationQueueItemResolver) Observation(ctx context.Context, obj *domain.AnnotationQueueItem) (*domain.Observation, error) {
	if obj.ObservationID == nil || *obj.ObservationID == "" {
		return nil, nil
	}
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}
	return r.QueryService.GetObservation(ctx, projectID, *obj.ObservationID)
}

// ======== Dataset Resolver ========

// Metadata is the resolver for the metadata field.
func (r *datasetResolver) Metadata(ctx context.Context, obj *domain.Dataset) (map[string]any, error) {
	return parseJSONString(obj.Metadata)
}

// Items is the resolver for the items field.
func (r *datasetResolver) Items(ctx context.Context, obj *domain.Dataset, limit *int, offset *int) (*model.DatasetItemConnection, error) {
	limitVal := getLimit(limit)
	offsetVal := 0
	if offset != nil {
		offsetVal = *offset
	}

	filter := &domain.DatasetItemFilter{DatasetID: obj.ID}
	items, total, err := r.DatasetService.ListItems(ctx, filter, limitVal, offsetVal)
	if err != nil {
		return nil, err
	}

	edges := make([]*model.DatasetItemEdge, len(items))
	for i := range items {
		edges[i] = &model.DatasetItemEdge{
			Node:   &items[i],
			Cursor: encodeCursor(offsetVal + i),
		}
	}

	hasNext := int64(offsetVal+len(items)) < total
	hasPrev := offsetVal > 0

	return &model.DatasetItemConnection{
		Edges:      edges,
		TotalCount: int(total),
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrev,
		},
	}, nil
}

// ======== DatasetItem Resolver ========

// Input is the resolver for the input field.
func (r *datasetItemResolver) Input(ctx context.Context, obj *domain.DatasetItem) (map[string]any, error) {
	return parseJSONString(obj.Input)
}

// ExpectedOutput is the resolver for the expectedOutput field.
func (r *datasetItemResolver) ExpectedOutput(ctx context.Context, obj *domain.DatasetItem) (map[string]any, error) {
	return parseJSONStringPtr(obj.ExpectedOutput)
}

// Metadata is the resolver for the metadata field.
func (r *datasetItemResolver) Metadata(ctx context.Context, obj *domain.DatasetItem) (map[string]any, error) {
	return parseJSONString(obj.Metadata)
}

// ======== DatasetRun Resolver ========

// Metadata is the resolver for the metadata field.
func (r *datasetRunResolver) Metadata(ctx context.Context, obj *domain.DatasetRun) (map[string]any, error) {
	return parseJSONString(obj.Metadata)
}

// ======== DatasetRunItemWithDiff Resolver ========

// RunItem is the resolver for the runItem field.
func (r *datasetRunItemWithDiffResolver) RunItem(ctx context.Context, obj *domain.DatasetRunItemWithDiff) (*domain.DatasetRunItem, error) {
	return &obj.DatasetRunItem, nil
}

// ExpectedOutput is the resolver for the expectedOutput field.
func (r *datasetRunItemWithDiffResolver) ExpectedOutput(ctx context.Context, obj *domain.DatasetRunItemWithDiff) (map[string]any, error) {
	return parseJSONStringPtr(obj.ExpectedOutput)
}

// ActualOutput is the resolver for the actualOutput field.
func (r *datasetRunItemWithDiffResolver) ActualOutput(ctx context.Context, obj *domain.DatasetRunItemWithDiff) (map[string]any, error) {
	return parseJSONStringPtr(obj.ActualOutput)
}

// ======== Evaluator Resolver ========

// Config is the resolver for the config field.
func (r *evaluatorResolver) Config(ctx context.Context, obj *domain.Evaluator) (map[string]any, error) {
	return parseJSONString(obj.Config)
}

// TargetFilter is the resolver for the targetFilter field.
func (r *evaluatorResolver) TargetFilter(ctx context.Context, obj *domain.Evaluator) (map[string]any, error) {
	return parseJSONString(obj.TargetFilter)
}

// CreatedByID is the resolver for the createdById field.
func (r *evaluatorResolver) CreatedByID(ctx context.Context, obj *domain.Evaluator) (*uuid.UUID, error) {
	return obj.CreatedBy, nil
}

// ======== EvaluatorTemplate Resolver ========

// Type is the resolver for the type field.
func (r *evaluatorTemplateResolver) Type(ctx context.Context, obj *domain.EvaluatorTemplate) (domain.EvaluatorType, error) {
	// EvaluatorTemplates are LLM-based by default
	return domain.EvaluatorTypeLLM, nil
}

// Config is the resolver for the config field.
func (r *evaluatorTemplateResolver) Config(ctx context.Context, obj *domain.EvaluatorTemplate) (map[string]any, error) {
	return parseJSONString(obj.Config)
}

// ======== Mutation Resolver ========

// derefString safely dereferences a string pointer
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// derefBool safely dereferences a bool pointer
func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// uuidToStringPtr converts a *uuid.UUID to *string
func uuidToStringPtr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

// mapToProjectSettings converts map[string]interface{} to *domain.ProjectSettings
func mapToProjectSettings(m map[string]interface{}) *domain.ProjectSettings {
	if m == nil {
		return nil
	}
	settings := &domain.ProjectSettings{}
	if v, ok := m["enableEvaluations"].(bool); ok {
		settings.EnableEvaluations = v
	}
	if v, ok := m["enablePrompts"].(bool); ok {
		settings.EnablePrompts = v
	}
	if v, ok := m["enableDatasets"].(bool); ok {
		settings.EnableDatasets = v
	}
	if v, ok := m["enableExports"].(bool); ok {
		settings.EnableExports = v
	}
	if v, ok := m["defaultRetentionDays"].(int); ok {
		settings.DefaultRetentionDays = v
	} else if v, ok := m["defaultRetentionDays"].(float64); ok {
		settings.DefaultRetentionDays = int(v)
	}
	return settings
}

// stringValue safely dereferences a string pointer (alias for derefString)
func stringValue(s *string) string {
	return derefString(s)
}

// boolValue safely dereferences a bool pointer with a default value
func boolValue(b *bool, defaultVal bool) bool {
	if b == nil {
		return defaultVal
	}
	return *b
}

// timeValue safely dereferences a time pointer with a default value
func timeValue(t *time.Time, defaultVal time.Time) time.Time {
	if t == nil {
		return defaultVal
	}
	return *t
}

// timeValuePtr returns the time pointer as-is (identity function)
func timeValuePtr(t *time.Time) *time.Time {
	return t
}

// generateTraceID generates a random 32-character hex trace ID
func generateTraceID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%032x", b)
}

// generateSpanID generates a random 16-character hex span ID
func generateSpanID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%016x", b)
}

// slugify converts a string to a URL-friendly slug
func slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)
	// Replace spaces and underscores with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	s = result.String()
	// Remove consecutive hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	// Trim hyphens from edges
	s = strings.Trim(s, "-")
	return s
}

// CreateTrace is the resolver for the createTrace field.
func (r *mutationResolver) CreateTrace(ctx context.Context, input model.CreateTraceInput) (*domain.Trace, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	traceInput := &domain.TraceInput{
		ID:        derefString(input.ID),
		Name:      derefString(input.Name),
		Timestamp: input.Timestamp,
		Input:     input.Input,
		Output:    input.Output,
		Metadata:  input.Metadata,
		Tags:      input.Tags,
		UserID:    derefString(input.UserId),
		SessionID: derefString(input.SessionId),
		Release:   derefString(input.Release),
		Version:   derefString(input.Version),
		Public:    derefBool(input.Public),
	}

	return r.IngestionService.IngestTrace(ctx, projectID, traceInput)
}

// UpdateTrace is the resolver for the updateTrace field.
func (r *mutationResolver) UpdateTrace(ctx context.Context, id string, input model.UpdateTraceInput) (*domain.Trace, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	updateInput := &domain.TraceUpdateInput{
		Name:          input.Name,
		UserID:        input.UserId,
		Level:         input.Level,
		StatusMessage: input.StatusMessage,
		Public:        input.Public,
		Tags:          input.Tags,
	}

	return r.QueryService.UpdateTrace(ctx, projectID, id, updateInput)
}

// DeleteTrace is the resolver for the deleteTrace field.
func (r *mutationResolver) DeleteTrace(ctx context.Context, id string) (bool, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return false, err
	}

	err = r.QueryService.DeleteTrace(ctx, projectID, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateSpan is the resolver for the createSpan field.
func (r *mutationResolver) CreateSpan(ctx context.Context, input model.CreateObservationInput) (*domain.Observation, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	obsType := domain.ObservationTypeSpan
	obsInput := &domain.ObservationInput{
		ID:                  input.ID,
		TraceID:             &input.TraceId,
		ParentObservationID: input.ParentObservationId,
		Type:                &obsType,
		Name:                input.Name,
		StartTime:           input.StartTime,
		EndTime:             input.EndTime,
		Input:               input.Input,
		Output:              input.Output,
		Metadata:            input.Metadata,
		Level:               input.Level,
		StatusMessage:       input.StatusMessage,
	}

	return r.IngestionService.IngestObservation(ctx, projectID, obsInput)
}

// CreateGeneration is the resolver for the createGeneration field.
func (r *mutationResolver) CreateGeneration(ctx context.Context, input model.CreateGenerationInput) (*domain.Observation, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	var usage *domain.UsageDetailsInput
	if input.Usage != nil {
		usage = &domain.UsageDetailsInput{}
		if input.Usage.PromptTokens != nil {
			val := int64(*input.Usage.PromptTokens)
			usage.InputTokens = &val
		}
		if input.Usage.CompletionTokens != nil {
			val := int64(*input.Usage.CompletionTokens)
			usage.OutputTokens = &val
		}
		if input.Usage.TotalTokens != nil {
			val := int64(*input.Usage.TotalTokens)
			usage.TotalTokens = &val
		}
	}

	genInput := &domain.GenerationInput{
		ObservationInput: domain.ObservationInput{
			ID:                  input.ID,
			TraceID:             &input.TraceId,
			ParentObservationID: input.ParentObservationId,
			Name:                input.Name,
			StartTime:           input.StartTime,
			EndTime:             input.EndTime,
			Input:               input.Input,
			Output:              input.Output,
			Metadata:            input.Metadata,
			Level:               input.Level,
			StatusMessage:       input.StatusMessage,
		},
		Model:           derefString(input.Model),
		ModelParameters: input.ModelParameters,
		Usage:           usage,
	}

	return r.IngestionService.IngestGeneration(ctx, projectID, genInput)
}

// CreateEvent is the resolver for the createEvent field.
func (r *mutationResolver) CreateEvent(ctx context.Context, input model.CreateObservationInput) (*domain.Observation, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	obsType := domain.ObservationTypeEvent
	obsInput := &domain.ObservationInput{
		ID:                  input.ID,
		TraceID:             &input.TraceId,
		ParentObservationID: input.ParentObservationId,
		Type:                &obsType,
		Name:                input.Name,
		StartTime:           input.StartTime,
		EndTime:             input.EndTime,
		Input:               input.Input,
		Output:              input.Output,
		Metadata:            input.Metadata,
		Level:               input.Level,
		StatusMessage:       input.StatusMessage,
	}

	return r.IngestionService.IngestObservation(ctx, projectID, obsInput)
}

// UpdateObservation is the resolver for the updateObservation field.
func (r *mutationResolver) UpdateObservation(ctx context.Context, id string, input model.UpdateObservationInput) (*domain.Observation, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	var usage *domain.UsageDetailsInput
	if input.Usage != nil {
		usage = &domain.UsageDetailsInput{}
		if input.Usage.PromptTokens != nil {
			val := int64(*input.Usage.PromptTokens)
			usage.InputTokens = &val
		}
		if input.Usage.CompletionTokens != nil {
			val := int64(*input.Usage.CompletionTokens)
			usage.OutputTokens = &val
		}
		if input.Usage.TotalTokens != nil {
			val := int64(*input.Usage.TotalTokens)
			usage.TotalTokens = &val
		}
	}

	obsInput := &domain.ObservationInput{
		Name:            input.Name,
		EndTime:         input.EndTime,
		Input:           input.Input,
		Output:          input.Output,
		Metadata:        input.Metadata,
		Level:           input.Level,
		StatusMessage:   input.StatusMessage,
		Model:           input.Model,
		ModelParameters: input.ModelParameters,
		Usage:           usage,
	}

	return r.IngestionService.UpdateObservation(ctx, projectID, id, obsInput)
}

// CreateScore is the resolver for the createScore field.
func (r *mutationResolver) CreateScore(ctx context.Context, input model.CreateScoreInput) (*domain.Score, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	var dataType domain.ScoreDataType
	if input.DataType != nil {
		dataType = *input.DataType
	}
	var source domain.ScoreSource
	if input.Source != nil {
		source = *input.Source
	}

	scoreInput := &domain.ScoreInput{
		TraceID:       input.TraceId,
		ObservationID: input.ObservationId,
		Name:          input.Name,
		Value:         input.Value,
		StringValue:   input.StringValue,
		DataType:      dataType,
		Source:        source,
		Comment:       input.Comment,
	}

	return r.ScoreService.Create(ctx, projectID, scoreInput)
}

// UpdateScore is the resolver for the updateScore field.
func (r *mutationResolver) UpdateScore(ctx context.Context, id string, input model.UpdateScoreInput) (*domain.Score, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	scoreInput := &domain.ScoreInput{
		Value:       input.Value,
		StringValue: input.StringValue,
		Comment:     input.Comment,
	}

	return r.ScoreService.Update(ctx, projectID, id, scoreInput)
}

// DeleteScore is the resolver for the deleteScore field.
func (r *mutationResolver) DeleteScore(ctx context.Context, id string) (bool, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return false, err
	}

	if r.ScoreService == nil {
		return false, fmt.Errorf("score service not configured")
	}

	err = r.ScoreService.Delete(ctx, projectID, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreatePrompt is the resolver for the createPrompt field.
func (r *mutationResolver) CreatePrompt(ctx context.Context, input model.CreatePromptInput) (*domain.Prompt, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}
	userID, _ := getUserID(ctx) // Optional - may not be set for API key auth

	// Build content from prompt or messages
	var content string
	if input.Prompt != nil {
		content = *input.Prompt
	} else if len(input.Messages) > 0 {
		// Serialize messages to JSON for storage
		messagesJSON, err := json.Marshal(input.Messages)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize messages: %w", err)
		}
		content = string(messagesJSON)
	}

	promptInput := &domain.PromptInput{
		Name:        input.Name,
		Type:        input.Type,
		Description: input.Description,
		Content:     content,
		Config:      input.Config,
		Labels:      input.Labels,
	}

	return r.PromptService.Create(ctx, projectID, promptInput, userID)
}

// UpdatePrompt is the resolver for the updatePrompt field.
func (r *mutationResolver) UpdatePrompt(ctx context.Context, name string, input model.UpdatePromptInput) (*domain.Prompt, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	// Get prompt by name to get its ID
	prompt, err := r.PromptService.GetByName(ctx, projectID, name)
	if err != nil {
		return nil, err
	}

	// Build content from prompt or messages
	var content string
	if input.Prompt != nil {
		content = *input.Prompt
	} else if len(input.Messages) > 0 {
		messagesJSON, err := json.Marshal(input.Messages)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize messages: %w", err)
		}
		content = string(messagesJSON)
	}

	promptInput := &domain.PromptInput{
		Name:        name, // Keep the same name
		Description: input.Description,
		Content:     content,
		Config:      input.Config,
		Labels:      input.Labels,
	}

	return r.PromptService.Update(ctx, prompt.ID, promptInput)
}

// DeletePrompt is the resolver for the deletePrompt field.
func (r *mutationResolver) DeletePrompt(ctx context.Context, name string) (bool, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return false, err
	}

	// Get prompt by name to get its ID
	prompt, err := r.PromptService.GetByName(ctx, projectID, name)
	if err != nil {
		return false, err
	}

	err = r.PromptService.Delete(ctx, prompt.ID)
	if err != nil {
		return false, err
	}
	return true, nil
}

// SetPromptLabel is the resolver for the setPromptLabel field.
func (r *mutationResolver) SetPromptLabel(ctx context.Context, name string, version int, label string) (bool, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return false, err
	}

	// Get prompt by name to get its ID
	prompt, err := r.PromptService.GetByName(ctx, projectID, name)
	if err != nil {
		return false, err
	}

	err = r.PromptService.SetVersionLabel(ctx, prompt.ID, version, label, true)
	if err != nil {
		return false, err
	}
	return true, nil
}

// RemovePromptLabel is the resolver for the removePromptLabel field.
func (r *mutationResolver) RemovePromptLabel(ctx context.Context, name string, label string) (bool, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return false, err
	}

	// Get prompt by name to get its ID
	prompt, err := r.PromptService.GetByName(ctx, projectID, name)
	if err != nil {
		return false, err
	}

	// SetVersionLabel with version=0 and add=false to remove label from all versions
	err = r.PromptService.SetVersionLabel(ctx, prompt.ID, 0, label, false)
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateDataset is the resolver for the createDataset field.
func (r *mutationResolver) CreateDataset(ctx context.Context, input model.CreateDatasetInput) (*domain.Dataset, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	datasetInput := &domain.DatasetInput{
		Name:        input.Name,
		Description: derefString(input.Description),
		Metadata:    input.Metadata,
	}

	return r.DatasetService.Create(ctx, projectID, datasetInput)
}

// UpdateDataset is the resolver for the updateDataset field.
func (r *mutationResolver) UpdateDataset(ctx context.Context, id uuid.UUID, input model.UpdateDatasetInput) (*domain.Dataset, error) {
	datasetInput := &domain.DatasetInput{
		Name:        derefString(input.Name),
		Description: derefString(input.Description),
		Metadata:    input.Metadata,
	}

	return r.DatasetService.Update(ctx, id, datasetInput)
}

// DeleteDataset is the resolver for the deleteDataset field.
func (r *mutationResolver) DeleteDataset(ctx context.Context, id uuid.UUID) (bool, error) {
	err := r.DatasetService.Delete(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateDatasetItem is the resolver for the createDatasetItem field.
func (r *mutationResolver) CreateDatasetItem(ctx context.Context, datasetID uuid.UUID, input model.CreateDatasetItemInput) (*domain.DatasetItem, error) {
	itemInput := &domain.DatasetItemInput{
		Input:               input.Input,
		ExpectedOutput:      input.ExpectedOutput,
		Metadata:            input.Metadata,
		SourceTraceID:       input.SourceTraceId,
		SourceObservationID: input.SourceObservationId,
	}

	return r.DatasetService.AddItem(ctx, datasetID, itemInput)
}

// UpdateDatasetItem is the resolver for the updateDatasetItem field.
func (r *mutationResolver) UpdateDatasetItem(ctx context.Context, id uuid.UUID, input model.UpdateDatasetItemInput) (*domain.DatasetItem, error) {
	updateInput := &domain.DatasetItemUpdateInput{
		Input:          input.Input,
		ExpectedOutput: input.ExpectedOutput,
		Metadata:       input.Metadata,
		Status:         input.Status,
	}

	return r.DatasetService.UpdateItem(ctx, id, updateInput)
}

// DeleteDatasetItem is the resolver for the deleteDatasetItem field.
func (r *mutationResolver) DeleteDatasetItem(ctx context.Context, id uuid.UUID) (bool, error) {
	err := r.DatasetService.DeleteItem(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateDatasetRun is the resolver for the createDatasetRun field.
func (r *mutationResolver) CreateDatasetRun(ctx context.Context, datasetID uuid.UUID, input model.CreateDatasetRunInput) (*domain.DatasetRun, error) {
	runInput := &domain.DatasetRunInput{
		Name:        input.Name,
		Description: derefString(input.Description),
		Metadata:    input.Metadata,
	}

	return r.DatasetService.CreateRun(ctx, datasetID, runInput)
}

// AddDatasetRunItem is the resolver for the addDatasetRunItem field.
func (r *mutationResolver) AddDatasetRunItem(ctx context.Context, runID uuid.UUID, input model.AddDatasetRunItemInput) (*domain.DatasetRunItem, error) {
	itemInput := &domain.DatasetRunItemInput{
		DatasetItemID: input.DatasetItemId.String(),
		TraceID:       derefString(input.TraceId),
		ObservationID: input.ObservationId,
	}

	return r.DatasetService.AddRunItem(ctx, runID, itemInput)
}

// CreateEvaluator is the resolver for the createEvaluator field.
func (r *mutationResolver) CreateEvaluator(ctx context.Context, input model.CreateEvaluatorInput) (*domain.Evaluator, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	evalType := input.Type
	scoreDataType := input.ScoreDataType
	evalInput := &domain.EvaluatorInput{
		Name:            input.Name,
		Description:     input.Description,
		Type:            &evalType,
		ScoreName:       input.ScoreName,
		ScoreDataType:   &scoreDataType,
		ScoreCategories: input.ScoreCategories,
		PromptTemplate:  derefString(input.PromptTemplate),
		Variables:       input.Variables,
		Config:          input.Config,
		TargetFilter:    input.TargetFilter,
		SamplingRate:    input.SamplingRate,
		Enabled:         input.Enabled,
		TemplateID:      uuidToStringPtr(input.TemplateId),
	}

	return r.EvalService.Create(ctx, projectID, evalInput, userID)
}

// UpdateEvaluator is the resolver for the updateEvaluator field.
func (r *mutationResolver) UpdateEvaluator(ctx context.Context, id uuid.UUID, input model.UpdateEvaluatorInput) (*domain.Evaluator, error) {
	updateInput := &domain.EvaluatorUpdateInput{
		Name:            input.Name,
		Description:     input.Description,
		PromptTemplate:  input.PromptTemplate,
		Variables:       input.Variables,
		ScoreCategories: input.ScoreCategories,
		Config:          input.Config,
		TargetFilter:    input.TargetFilter,
		SamplingRate:    input.SamplingRate,
		Enabled:         input.IsActive,
	}

	return r.EvalService.Update(ctx, id, updateInput)
}

// DeleteEvaluator is the resolver for the deleteEvaluator field.
func (r *mutationResolver) DeleteEvaluator(ctx context.Context, id uuid.UUID) (bool, error) {
	err := r.EvalService.Delete(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateOrganization is the resolver for the createOrganization field.
func (r *mutationResolver) CreateOrganization(ctx context.Context, name string) (*domain.Organization, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	return r.OrgService.Create(ctx, name, userID)
}

// UpdateOrganization is the resolver for the updateOrganization field.
func (r *mutationResolver) UpdateOrganization(ctx context.Context, id uuid.UUID, name string) (*domain.Organization, error) {
	return r.OrgService.Update(ctx, id, name)
}

// DeleteOrganization is the resolver for the deleteOrganization field.
func (r *mutationResolver) DeleteOrganization(ctx context.Context, id uuid.UUID) (bool, error) {
	err := r.OrgService.Delete(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateProject is the resolver for the createProject field.
func (r *mutationResolver) CreateProject(ctx context.Context, input model.CreateProjectInput) (*domain.Project, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	projectInput := &service.ProjectInput{
		Name:            input.Name,
		Description:     derefString(input.Description),
		Settings:        mapToProjectSettings(input.Settings),
		RetentionDays:   input.RetentionDays,
		RateLimitPerMin: input.RateLimitPerMin,
	}

	return r.ProjectService.Create(ctx, input.OrganizationId, projectInput, userID)
}

// UpdateProject is the resolver for the updateProject field.
func (r *mutationResolver) UpdateProject(ctx context.Context, id uuid.UUID, input model.UpdateProjectInput) (*domain.Project, error) {
	updateInput := &service.ProjectInput{
		Name:            derefString(input.Name),
		Description:     derefString(input.Description),
		Settings:        mapToProjectSettings(input.Settings),
		RetentionDays:   input.RetentionDays,
		RateLimitPerMin: input.RateLimitPerMin,
	}

	return r.ProjectService.Update(ctx, id, updateInput)
}

// DeleteProject is the resolver for the deleteProject field.
func (r *mutationResolver) DeleteProject(ctx context.Context, id uuid.UUID) (bool, error) {
	err := r.ProjectService.Delete(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateAPIKey is the resolver for the createAPIKey field.
func (r *mutationResolver) CreateAPIKey(ctx context.Context, input model.CreateAPIKeyInput) (*model.APIKeyWithSecret, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	apiKeyInput := &domain.APIKeyInput{
		Name:      input.Name,
		Scopes:    input.Scopes,
		ExpiresAt: input.ExpiresAt,
	}

	result, err := r.AuthService.CreateAPIKey(ctx, projectID, apiKeyInput, userID)
	if err != nil {
		return nil, err
	}

	displayKey := result.SecretKey
	if len(displayKey) > 12 {
		displayKey = displayKey[:8] + "..." + displayKey[len(displayKey)-4:]
	}

	return &model.APIKeyWithSecret{
		ID:         result.APIKey.ID,
		Name:       result.APIKey.Name,
		Key:        result.SecretKey,
		DisplayKey: displayKey,
		Scopes:     result.APIKey.Scopes,
		ExpiresAt:  result.APIKey.ExpiresAt,
		CreatedAt:  result.APIKey.CreatedAt,
	}, nil
}

// DeleteAPIKey is the resolver for the deleteAPIKey field.
func (r *mutationResolver) DeleteAPIKey(ctx context.Context, id uuid.UUID) (bool, error) {
	err := r.AuthService.DeleteAPIKey(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ======== Observation Resolver ========

// Input is the resolver for the input field.
func (r *observationResolver) Input(ctx context.Context, obj *domain.Observation) (map[string]any, error) {
	return parseJSONString(obj.Input)
}

// Output is the resolver for the output field.
func (r *observationResolver) Output(ctx context.Context, obj *domain.Observation) (map[string]any, error) {
	return parseJSONString(obj.Output)
}

// Metadata is the resolver for the metadata field.
func (r *observationResolver) Metadata(ctx context.Context, obj *domain.Observation) (map[string]any, error) {
	return parseJSONString(obj.Metadata)
}

// ModelParameters is the resolver for the modelParameters field.
func (r *observationResolver) ModelParameters(ctx context.Context, obj *domain.Observation) (map[string]any, error) {
	return parseJSONString(obj.ModelParameters)
}

// PromptTokens is the resolver for the promptTokens field.
func (r *observationResolver) PromptTokens(ctx context.Context, obj *domain.Observation) (*int, error) {
	tokens := obj.UsageDetails.InputTokens
	if tokens == 0 {
		return nil, nil
	}
	intTokens := int(tokens)
	return &intTokens, nil
}

// CompletionTokens is the resolver for the completionTokens field.
func (r *observationResolver) CompletionTokens(ctx context.Context, obj *domain.Observation) (*int, error) {
	tokens := obj.UsageDetails.OutputTokens
	if tokens == 0 {
		return nil, nil
	}
	intTokens := int(tokens)
	return &intTokens, nil
}

// TotalTokens is the resolver for the totalTokens field.
func (r *observationResolver) TotalTokens(ctx context.Context, obj *domain.Observation) (*int, error) {
	tokens := obj.UsageDetails.TotalTokens
	if tokens == 0 {
		return nil, nil
	}
	intTokens := int(tokens)
	return &intTokens, nil
}

// Latency is the resolver for the latency field.
func (r *observationResolver) Latency(ctx context.Context, obj *domain.Observation) (*float64, error) {
	if obj.DurationMs == 0 {
		return nil, nil
	}
	latency := obj.DurationMs
	return &latency, nil
}

// Cost is the resolver for the cost field.
func (r *observationResolver) Cost(ctx context.Context, obj *domain.Observation) (*float64, error) {
	if obj.CostDetails.TotalCost == 0 {
		return nil, nil
	}
	cost := obj.CostDetails.TotalCost
	return &cost, nil
}

// ======== Project Resolver ========

// Settings is the resolver for the settings field.
func (r *projectResolver) Settings(ctx context.Context, obj *domain.Project) (map[string]any, error) {
	return parseJSONString(obj.Settings)
}

// ======== Prompt Resolver ========

// IsActive is the resolver for the isActive field.
func (r *promptResolver) IsActive(ctx context.Context, obj *domain.Prompt) (bool, error) {
	// A prompt is considered active if it has a latest version
	return obj.LatestVersion != nil, nil
}

// CreatedByID is the resolver for the createdById field.
func (r *promptResolver) CreatedByID(ctx context.Context, obj *domain.Prompt) (*uuid.UUID, error) {
	// Get from latest version if available
	if obj.LatestVersion != nil && obj.LatestVersion.CreatedBy != nil {
		return obj.LatestVersion.CreatedBy, nil
	}
	return nil, nil
}

// Version is the resolver for the version field.
func (r *promptResolver) Version(ctx context.Context, obj *domain.Prompt) (*domain.PromptVersion, error) {
	if obj.LatestVersion == nil {
		return nil, nil
	}
	return obj.LatestVersion, nil
}

// Labels is the resolver for the labels field.
func (r *promptResolver) Labels(ctx context.Context, obj *domain.Prompt) ([]string, error) {
	// Get labels from latest version
	if obj.LatestVersion != nil {
		return obj.LatestVersion.Labels, nil
	}
	return nil, nil
}

// ======== PromptVersion Resolver ========

// Prompt is the resolver for the prompt field.
func (r *promptVersionResolver) Prompt(ctx context.Context, obj *domain.PromptVersion) (*string, error) {
	if obj.Content == "" {
		return nil, nil
	}
	return &obj.Content, nil
}

// Messages is the resolver for the messages field.
func (r *promptVersionResolver) Messages(ctx context.Context, obj *domain.PromptVersion) ([]api.PromptMessage, error) {
	// Try to parse Content as JSON array of messages
	if obj.Content == "" {
		return nil, nil
	}
	var messages []api.PromptMessage
	if err := json.Unmarshal([]byte(obj.Content), &messages); err != nil {
		// Content is not a messages array, return nil
		return nil, nil
	}
	return messages, nil
}

// Config is the resolver for the config field.
func (r *promptVersionResolver) Config(ctx context.Context, obj *domain.PromptVersion) (map[string]any, error) {
	return parseJSONString(obj.Config)
}

// CreatedByID is the resolver for the createdById field.
func (r *promptVersionResolver) CreatedByID(ctx context.Context, obj *domain.PromptVersion) (*uuid.UUID, error) {
	return obj.CreatedBy, nil
}

// ======== Query Resolver ========

// Trace is the resolver for the trace field.
func (r *queryResolver) Trace(ctx context.Context, id string) (*domain.Trace, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}
	return r.QueryService.GetTrace(ctx, projectID, id)
}

// Traces is the resolver for the traces field.
func (r *queryResolver) Traces(ctx context.Context, input model.TracesInput) (*model.TraceConnection, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	limit := getLimit(input.Limit)
	offset := getOffset(input.Cursor)

	filter := &domain.TraceFilter{
		ProjectID: projectID,
		UserID:    input.UserID,
		SessionID: input.SessionID,
		Name:      input.Name,
		Tags:      input.Tags,
		FromTime:  input.FromTimestamp,
		ToTime:    input.ToTimestamp,
		Version:   input.Version,
		Release:   input.Release,
	}

	list, err := r.QueryService.ListTraces(ctx, filter, limit, offset)
	if err != nil {
		return nil, err
	}

	edges := make([]*model.TraceEdge, len(list.Traces))
	for i := range list.Traces {
		edges[i] = &model.TraceEdge{
			Node:   &list.Traces[i],
			Cursor: encodeCursor(offset + i),
		}
	}

	hasNext := int64(offset+len(list.Traces)) < list.TotalCount
	hasPrev := offset > 0

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &model.TraceConnection{
		Edges:      edges,
		TotalCount: int(list.TotalCount),
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrev,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// Observation is the resolver for the observation field.
func (r *queryResolver) Observation(ctx context.Context, id string) (*domain.Observation, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}
	return r.QueryService.GetObservation(ctx, projectID, id)
}

// Observations is the resolver for the observations field.
func (r *queryResolver) Observations(ctx context.Context, input model.ObservationsInput) (*model.ObservationConnection, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	limit := getLimit(input.Limit)
	offset := getOffset(input.Cursor)

	filter := &domain.ObservationFilter{
		ProjectID:           projectID,
		TraceID:             input.TraceID,
		ParentObservationID: input.ParentObservationID,
		Type:                input.Type,
		Name:                input.Name,
	}

	observations, total, err := r.QueryService.ListObservations(ctx, filter, limit, offset)
	if err != nil {
		return nil, err
	}

	edges := make([]*model.ObservationEdge, len(observations))
	for i := range observations {
		edges[i] = &model.ObservationEdge{
			Node:   &observations[i],
			Cursor: encodeCursor(offset + i),
		}
	}

	hasNext := int64(offset+len(observations)) < total
	hasPrev := offset > 0

	return &model.ObservationConnection{
		Edges:      edges,
		TotalCount: int(total),
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrev,
		},
	}, nil
}

// Score is the resolver for the score field.
func (r *queryResolver) Score(ctx context.Context, id string) (*domain.Score, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}
	return r.ScoreService.Get(ctx, projectID, id)
}

// Scores is the resolver for the scores field.
func (r *queryResolver) Scores(ctx context.Context, input model.ScoresInput) (*model.ScoreConnection, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	limit := getLimit(input.Limit)
	offset := getOffset(input.Cursor)

	filter := &domain.ScoreFilter{
		ProjectID:     projectID,
		TraceID:       input.TraceID,
		ObservationID: input.ObservationID,
		Name:          input.Name,
		Source:        input.Source,
	}

	list, err := r.ScoreService.List(ctx, filter, limit, offset)
	if err != nil {
		return nil, err
	}

	edges := make([]*model.ScoreEdge, len(list.Scores))
	for i := range list.Scores {
		edges[i] = &model.ScoreEdge{
			Node:   &list.Scores[i],
			Cursor: encodeCursor(offset + i),
		}
	}

	hasNext := int64(offset+len(list.Scores)) < list.TotalCount
	hasPrev := offset > 0

	return &model.ScoreConnection{
		Edges:      edges,
		TotalCount: int(list.TotalCount),
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrev,
		},
	}, nil
}

// Session is the resolver for the session field.
func (r *queryResolver) Session(ctx context.Context, id string) (*model.Session, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	session, err := r.QueryService.GetSession(ctx, projectID, id)
	if err != nil {
		return nil, err
	}

	// Convert []domain.Trace to []*domain.Trace
	traces := make([]*domain.Trace, len(session.Traces))
	for i := range session.Traces {
		traces[i] = &session.Traces[i]
	}

	// TotalCost needs to be a pointer
	var totalCost *float64
	if session.TotalCost > 0 {
		totalCost = &session.TotalCost
	}

	return &model.Session{
		ID:            session.ID,
		ProjectID:     session.ProjectID,
		CreatedAt:     session.CreatedAt,
		TraceCount:    int(session.TraceCount),
		TotalDuration: nil, // Not available in domain
		TotalCost:     totalCost,
		Traces:        traces,
	}, nil
}

// Sessions is the resolver for the sessions field.
func (r *queryResolver) Sessions(ctx context.Context, input model.SessionsInput) (*model.SessionConnection, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	limit := getLimit(input.Limit)
	offset := getOffset(input.Cursor)

	filter := &domain.SessionFilter{
		ProjectID: projectID,
		FromTime:  input.FromTimestamp,
		ToTime:    input.ToTimestamp,
	}

	list, err := r.QueryService.ListSessions(ctx, filter, limit, offset)
	if err != nil {
		return nil, err
	}

	edges := make([]*model.SessionEdge, len(list.Sessions))
	for i, s := range list.Sessions {
		var totalCost *float64
		if s.TotalCost > 0 {
			cost := s.TotalCost
			totalCost = &cost
		}
		edges[i] = &model.SessionEdge{
			Node: &model.Session{
				ID:            s.ID,
				ProjectID:     s.ProjectID,
				CreatedAt:     s.CreatedAt,
				TraceCount:    int(s.TraceCount),
				TotalDuration: nil, // Not available in domain
				TotalCost:     totalCost,
			},
			Cursor: encodeCursor(offset + i),
		}
	}

	hasNext := int64(offset+len(list.Sessions)) < list.TotalCount
	hasPrev := offset > 0

	return &model.SessionConnection{
		Edges:      edges,
		TotalCount: int(list.TotalCount),
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrev,
		},
	}, nil
}

// Prompt is the resolver for the prompt field.
func (r *queryResolver) Prompt(ctx context.Context, name string, version *int, label *string) (*domain.Prompt, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	// Call appropriate method based on parameters
	if version != nil {
		return r.PromptService.GetByNameAndVersion(ctx, projectID, name, *version)
	}
	if label != nil {
		return r.PromptService.GetByNameAndLabel(ctx, projectID, name, *label)
	}
	return r.PromptService.GetByName(ctx, projectID, name)
}

// Prompts is the resolver for the prompts field.
func (r *queryResolver) Prompts(ctx context.Context, input model.PromptsInput) (*model.PromptConnection, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	limit := getLimit(input.Limit)
	offset := getOffset(input.Cursor)

	filter := &domain.PromptFilter{
		ProjectID: projectID,
		Name:      input.Name,
		Label:     input.Label,
		Tags:      input.Tags,
	}

	list, err := r.PromptService.List(ctx, filter, limit, offset)
	if err != nil {
		return nil, err
	}

	edges := make([]*model.PromptEdge, len(list.Prompts))
	for i := range list.Prompts {
		edges[i] = &model.PromptEdge{
			Node:   &list.Prompts[i],
			Cursor: encodeCursor(offset + i),
		}
	}

	hasNext := int64(offset+len(list.Prompts)) < list.TotalCount
	hasPrev := offset > 0

	return &model.PromptConnection{
		Edges:      edges,
		TotalCount: int(list.TotalCount),
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrev,
		},
	}, nil
}

// Dataset is the resolver for the dataset field.
func (r *queryResolver) Dataset(ctx context.Context, id uuid.UUID) (*domain.Dataset, error) {
	return r.DatasetService.Get(ctx, id)
}

// Datasets is the resolver for the datasets field.
func (r *queryResolver) Datasets(ctx context.Context, input model.DatasetsInput) (*model.DatasetConnection, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	limit := getLimit(input.Limit)
	offset := getOffset(input.Cursor)

	filter := &domain.DatasetFilter{
		ProjectID: projectID,
		Name:      input.Name,
	}

	list, err := r.DatasetService.List(ctx, filter, limit, offset)
	if err != nil {
		return nil, err
	}

	edges := make([]*model.DatasetEdge, len(list.Datasets))
	for i := range list.Datasets {
		edges[i] = &model.DatasetEdge{
			Node:   &list.Datasets[i],
			Cursor: encodeCursor(offset + i),
		}
	}

	hasNext := int64(offset+len(list.Datasets)) < list.TotalCount
	hasPrev := offset > 0

	return &model.DatasetConnection{
		Edges:      edges,
		TotalCount: int(list.TotalCount),
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrev,
		},
	}, nil
}

// DatasetByName is the resolver for the datasetByName field.
func (r *queryResolver) DatasetByName(ctx context.Context, name string) (*domain.Dataset, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}
	return r.DatasetService.GetByName(ctx, projectID, name)
}

// Evaluator is the resolver for the evaluator field.
func (r *queryResolver) Evaluator(ctx context.Context, id uuid.UUID) (*domain.Evaluator, error) {
	return r.EvalService.Get(ctx, id)
}

// Evaluators is the resolver for the evaluators field.
func (r *queryResolver) Evaluators(ctx context.Context, input model.EvaluatorsInput) (*model.EvaluatorConnection, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	limit := getLimit(input.Limit)
	offset := getOffset(input.Cursor)

	filter := &domain.EvaluatorFilter{
		ProjectID: projectID,
		Type:      input.Type,
		Enabled:   input.Enabled,
	}

	list, err := r.EvalService.List(ctx, filter, limit, offset)
	if err != nil {
		return nil, err
	}

	edges := make([]*model.EvaluatorEdge, len(list.Evaluators))
	for i := range list.Evaluators {
		edges[i] = &model.EvaluatorEdge{
			Node:   &list.Evaluators[i],
			Cursor: encodeCursor(offset + i),
		}
	}

	hasNext := int64(offset+len(list.Evaluators)) < list.TotalCount
	hasPrev := offset > 0

	return &model.EvaluatorConnection{
		Edges:      edges,
		TotalCount: int(list.TotalCount),
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrev,
		},
	}, nil
}

// EvaluatorTemplates is the resolver for the evaluatorTemplates field.
func (r *queryResolver) EvaluatorTemplates(ctx context.Context) ([]domain.EvaluatorTemplate, error) {
	return r.EvalService.ListTemplates(ctx)
}

// Organization is the resolver for the organization field.
func (r *queryResolver) Organization(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	return r.OrgService.Get(ctx, id)
}

// Organizations is the resolver for the organizations field.
func (r *queryResolver) Organizations(ctx context.Context) ([]domain.Organization, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}
	return r.OrgService.ListByUser(ctx, userID)
}

// Project is the resolver for the project field.
func (r *queryResolver) Project(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	return r.ProjectService.Get(ctx, id)
}

// Projects is the resolver for the projects field.
func (r *queryResolver) Projects(ctx context.Context, organizationID *uuid.UUID) ([]domain.Project, error) {
	if organizationID != nil {
		return r.ProjectService.ListByOrganization(ctx, *organizationID)
	}
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}
	return r.ProjectService.ListByUser(ctx, userID)
}

// Me is the resolver for the me field.
func (r *queryResolver) Me(ctx context.Context) (*domain.User, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}
	return r.AuthService.GetUserByID(ctx, userID)
}

// Metrics is the resolver for the metrics field.
func (r *queryResolver) Metrics(ctx context.Context, input model.MetricsInput) (*model.Metrics, error) {
	projectID, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	filter := &domain.TraceFilter{
		ProjectID: projectID,
		UserID:    input.UserId,
		SessionID: input.SessionId,
		Name:      input.Name,
		Tags:      input.Tags,
		FromTime:  &input.FromTimestamp,
		ToTime:    &input.ToTimestamp,
	}

	stats, err := r.QueryService.GetTraceStats(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Get generation stats for model usage
	genStats, err := r.QueryService.GetGenerationStats(ctx, projectID, nil)
	if err != nil {
		return nil, err
	}

	var modelUsage []*model.ModelUsage
	for _, ms := range genStats.ByModel {
		modelUsage = append(modelUsage, &model.ModelUsage{
			Model:            ms.Model,
			Count:            int(ms.Count),
			PromptTokens:     int(ms.TotalInputTokens),
			CompletionTokens: int(ms.TotalOutputTokens),
			TotalTokens:      int(ms.TotalInputTokens + ms.TotalOutputTokens),
			Cost:             ms.TotalCost,
		})
	}

	return &model.Metrics{
		TraceCount:       int(stats.TotalCount),
		ObservationCount: int(genStats.TotalCount),
		TotalCost:        stats.TotalCost,
		TotalTokens:      int(stats.TotalTokens),
		AvgLatency:       &stats.AvgDuration,
		ModelUsage:       modelUsage,
	}, nil
}

// DailyCosts is the resolver for the dailyCosts field.
func (r *queryResolver) DailyCosts(ctx context.Context, input model.DailyCostsInput) ([]model.DailyCost, error) {
	// Daily costs query not yet implemented in cost service
	// Return empty list for now
	_, err := getProjectID(ctx)
	if err != nil {
		return nil, err
	}

	return []model.DailyCost{}, nil
}

// ======== Score Resolver ========

// ID is the resolver for the id field.
func (r *scoreResolver) ID(ctx context.Context, obj *domain.Score) (string, error) {
	return obj.ID.String(), nil
}

// Timestamp is the resolver for the timestamp field.
func (r *scoreResolver) Timestamp(ctx context.Context, obj *domain.Score) (*time.Time, error) {
	if obj.CreatedAt.IsZero() {
		return nil, nil
	}
	return &obj.CreatedAt, nil
}

// ======== Subscription Resolver ========

// TraceCreated is the resolver for the traceCreated field.
func (r *subscriptionResolver) TraceCreated(ctx context.Context, projectID uuid.UUID) (<-chan *domain.Trace, error) {
	// Subscriptions require a realtime service - return error for now
	return nil, fmt.Errorf("subscriptions not yet implemented")
}

// TraceUpdated is the resolver for the traceUpdated field.
func (r *subscriptionResolver) TraceUpdated(ctx context.Context, projectID uuid.UUID) (<-chan *domain.Trace, error) {
	return nil, fmt.Errorf("subscriptions not yet implemented")
}

// ObservationCreated is the resolver for the observationCreated field.
func (r *subscriptionResolver) ObservationCreated(ctx context.Context, projectID uuid.UUID) (<-chan *domain.Observation, error) {
	return nil, fmt.Errorf("subscriptions not yet implemented")
}

// ScoreCreated is the resolver for the scoreCreated field.
func (r *subscriptionResolver) ScoreCreated(ctx context.Context, projectID uuid.UUID) (<-chan *domain.Score, error) {
	return nil, fmt.Errorf("subscriptions not yet implemented")
}

// ======== Trace Resolver ========

// Timestamp is the resolver for the timestamp field.
func (r *traceResolver) Timestamp(ctx context.Context, obj *domain.Trace) (*time.Time, error) {
	if obj.StartTime.IsZero() {
		return nil, nil
	}
	return &obj.StartTime, nil
}

// Input is the resolver for the input field.
func (r *traceResolver) Input(ctx context.Context, obj *domain.Trace) (map[string]any, error) {
	return parseJSONString(obj.Input)
}

// Output is the resolver for the output field.
func (r *traceResolver) Output(ctx context.Context, obj *domain.Trace) (map[string]any, error) {
	return parseJSONString(obj.Output)
}

// Metadata is the resolver for the metadata field.
func (r *traceResolver) Metadata(ctx context.Context, obj *domain.Trace) (map[string]any, error) {
	return parseJSONString(obj.Metadata)
}

// Latency is the resolver for the latency field.
func (r *traceResolver) Latency(ctx context.Context, obj *domain.Trace) (*float64, error) {
	if obj.DurationMs == 0 {
		return nil, nil
	}
	latency := obj.DurationMs
	return &latency, nil
}

// TokenUsage is the resolver for the tokenUsage field.
func (r *traceResolver) TokenUsage(ctx context.Context, obj *domain.Trace) (*model.TokenUsage, error) {
	if obj.TotalTokens == 0 {
		return nil, nil
	}
	return &model.TokenUsage{
		TotalTokens: int(obj.TotalTokens),
	}, nil
}

// ======== Factory Methods ========

// APIKey returns generated.APIKeyResolver implementation.
func (r *Resolver) APIKey() generated.APIKeyResolver { return &aPIKeyResolver{r} }

// AnnotationQueue returns generated.AnnotationQueueResolver implementation.
func (r *Resolver) AnnotationQueue() generated.AnnotationQueueResolver {
	return &annotationQueueResolver{r}
}

// AnnotationQueueItem returns generated.AnnotationQueueItemResolver implementation.
func (r *Resolver) AnnotationQueueItem() generated.AnnotationQueueItemResolver {
	return &annotationQueueItemResolver{r}
}

// Dataset returns generated.DatasetResolver implementation.
func (r *Resolver) Dataset() generated.DatasetResolver { return &datasetResolver{r} }

// DatasetItem returns generated.DatasetItemResolver implementation.
func (r *Resolver) DatasetItem() generated.DatasetItemResolver { return &datasetItemResolver{r} }

// DatasetRun returns generated.DatasetRunResolver implementation.
func (r *Resolver) DatasetRun() generated.DatasetRunResolver { return &datasetRunResolver{r} }

// DatasetRunItemWithDiff returns generated.DatasetRunItemWithDiffResolver implementation.
func (r *Resolver) DatasetRunItemWithDiff() generated.DatasetRunItemWithDiffResolver {
	return &datasetRunItemWithDiffResolver{r}
}

// Evaluator returns generated.EvaluatorResolver implementation.
func (r *Resolver) Evaluator() generated.EvaluatorResolver { return &evaluatorResolver{r} }

// EvaluatorTemplate returns generated.EvaluatorTemplateResolver implementation.
func (r *Resolver) EvaluatorTemplate() generated.EvaluatorTemplateResolver {
	return &evaluatorTemplateResolver{r}
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Observation returns generated.ObservationResolver implementation.
func (r *Resolver) Observation() generated.ObservationResolver { return &observationResolver{r} }

// Project returns generated.ProjectResolver implementation.
func (r *Resolver) Project() generated.ProjectResolver { return &projectResolver{r} }

// Prompt returns generated.PromptResolver implementation.
func (r *Resolver) Prompt() generated.PromptResolver { return &promptResolver{r} }

// PromptVersion returns generated.PromptVersionResolver implementation.
func (r *Resolver) PromptVersion() generated.PromptVersionResolver { return &promptVersionResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Score returns generated.ScoreResolver implementation.
func (r *Resolver) Score() generated.ScoreResolver { return &scoreResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

// Trace returns generated.TraceResolver implementation.
func (r *Resolver) Trace() generated.TraceResolver { return &traceResolver{r} }

// UpdateEvaluatorInput returns generated.UpdateEvaluatorInputResolver implementation.
func (r *Resolver) UpdateEvaluatorInput() generated.UpdateEvaluatorInputResolver {
	return &updateEvaluatorInputResolver{r}
}

// ======== Wrapper Types ========

type aPIKeyResolver struct{ *Resolver }
type annotationQueueResolver struct{ *Resolver }
type annotationQueueItemResolver struct{ *Resolver }
type datasetResolver struct{ *Resolver }
type datasetItemResolver struct{ *Resolver }
type datasetRunResolver struct{ *Resolver }
type datasetRunItemWithDiffResolver struct{ *Resolver }
type evaluatorResolver struct{ *Resolver }
type evaluatorTemplateResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type observationResolver struct{ *Resolver }
type projectResolver struct{ *Resolver }
type promptResolver struct{ *Resolver }
type promptVersionResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type scoreResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
type traceResolver struct{ *Resolver }
type updateEvaluatorInputResolver struct{ *Resolver }

// ======== UpdateEvaluatorInput Resolver ========

// Enabled maps the isActive field to enabled
func (r *updateEvaluatorInputResolver) Enabled(ctx context.Context, obj *model.UpdateEvaluatorInput, data *bool) error {
	obj.IsActive = data
	return nil
}
