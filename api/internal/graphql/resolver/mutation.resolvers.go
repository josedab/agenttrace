package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/graphql/model"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ======== TRACE MUTATIONS ========

// CreateTrace creates a new trace
func (r *mutationResolver) CreateTrace(ctx context.Context, input model.CreateTraceInput) (*domain.Trace, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("project ID not found in context")
	}

	// Convert inputs to TraceInput for ingestion service
	traceInput := &domain.TraceInput{
		Tags: input.Tags,
	}

	// Use provided ID or let service generate one
	if input.ID != nil {
		traceInput.ID = *input.ID
	}
	if input.Name != nil {
		traceInput.Name = *input.Name
	}
	if input.UserId != nil {
		traceInput.UserID = *input.UserId
	}
	if input.SessionId != nil {
		traceInput.SessionID = *input.SessionId
	}
	if input.Release != nil {
		traceInput.Release = *input.Release
	}
	if input.Version != nil {
		traceInput.Version = *input.Version
	}
	if input.Public != nil {
		traceInput.Public = *input.Public
	}

	// Handle timestamp
	if input.Timestamp != nil {
		traceInput.Timestamp = input.Timestamp
	}

	// Pass map values directly (TraceInput.Input/Output/Metadata are any)
	if input.Input != nil {
		traceInput.Input = input.Input
	}
	if input.Output != nil {
		traceInput.Output = input.Output
	}
	if input.Metadata != nil {
		traceInput.Metadata = input.Metadata
	}

	trace, err := r.ingestionService.IngestTrace(ctx, projectID, traceInput)
	if err != nil {
		r.logger.Error("failed to create trace", zap.Error(err))
		return nil, err
	}

	return trace, nil
}

// UpdateTrace updates an existing trace
func (r *mutationResolver) UpdateTrace(ctx context.Context, id string, input model.UpdateTraceInput) (*domain.Trace, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("project ID not found in context")
	}

	// Convert to TraceInput for update
	traceInput := &domain.TraceInput{
		Tags: input.Tags,
	}

	if input.Name != nil {
		traceInput.Name = *input.Name
	}
	if input.UserId != nil {
		traceInput.UserID = *input.UserId
	}
	if input.Public != nil {
		traceInput.Public = *input.Public
	}
	if input.Level != nil {
		traceInput.Level = *input.Level
	}
	if input.StatusMessage != nil {
		traceInput.StatusMessage = *input.StatusMessage
	}

	// Pass map values directly
	if input.Input != nil {
		traceInput.Input = input.Input
	}
	if input.Output != nil {
		traceInput.Output = input.Output
	}
	if input.Metadata != nil {
		traceInput.Metadata = input.Metadata
	}

	trace, err := r.ingestionService.UpdateTrace(ctx, projectID, id, traceInput)
	if err != nil {
		r.logger.Error("failed to update trace", zap.Error(err))
		return nil, err
	}

	return trace, nil
}

// DeleteTrace deletes a trace
// Note: IngestionService does not support trace deletion yet
func (r *mutationResolver) DeleteTrace(ctx context.Context, id string) (bool, error) {
	_, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return false, fmt.Errorf("project ID not found in context")
	}

	// TODO: Implement trace deletion when the service supports it
	return false, fmt.Errorf("trace deletion is not yet supported")
}

// ======== OBSERVATION MUTATIONS ========

// CreateSpan creates a new span observation
func (r *mutationResolver) CreateSpan(ctx context.Context, input model.CreateObservationInput) (*domain.Observation, error) {
	return r.createObservation(ctx, input, domain.ObservationTypeSpan)
}

// CreateGeneration creates a new generation observation
func (r *mutationResolver) CreateGeneration(ctx context.Context, input model.CreateGenerationInput) (*domain.Observation, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("project ID not found in context")
	}

	traceID := input.TraceId
	obsType := domain.ObservationTypeGeneration

	// Build ObservationInput
	obsInput := &domain.ObservationInput{
		ID:                  input.ID,
		TraceID:             &traceID,
		ParentObservationID: input.ParentObservationId,
		Type:                &obsType,
		Name:                input.Name,
		Level:               input.Level,
		StatusMessage:       input.StatusMessage,
		StartTime:           input.StartTime,
		EndTime:             input.EndTime,
		Model:               input.Model,
		Version:             input.Version,
		Input:               input.Input,
		Output:              input.Output,
		Metadata:            input.Metadata,
		ModelParameters:     input.ModelParameters,
		Usage:               input.Usage,
	}

	// Handle prompt ID conversion
	if input.PromptId != nil {
		pidStr := input.PromptId.String()
		obsInput.PromptID = &pidStr
	}

	obs, err := r.ingestionService.IngestObservation(ctx, projectID, obsInput)
	if err != nil {
		r.logger.Error("failed to create generation", zap.Error(err))
		return nil, err
	}

	return obs, nil
}

// CreateEvent creates a new event observation
func (r *mutationResolver) CreateEvent(ctx context.Context, input model.CreateObservationInput) (*domain.Observation, error) {
	return r.createObservation(ctx, input, domain.ObservationTypeEvent)
}

// createObservation is a helper for creating span and event observations
func (r *mutationResolver) createObservation(ctx context.Context, input model.CreateObservationInput, obsType domain.ObservationType) (*domain.Observation, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("project ID not found in context")
	}

	traceID := input.TraceId

	// Build ObservationInput
	obsInput := &domain.ObservationInput{
		ID:                  input.ID,
		TraceID:             &traceID,
		ParentObservationID: input.ParentObservationId,
		Type:                &obsType,
		Name:                input.Name,
		Level:               input.Level,
		StatusMessage:       input.StatusMessage,
		StartTime:           input.StartTime,
		EndTime:             input.EndTime,
		Version:             input.Version,
		Input:               input.Input,
		Output:              input.Output,
		Metadata:            input.Metadata,
	}

	obs, err := r.ingestionService.IngestObservation(ctx, projectID, obsInput)
	if err != nil {
		r.logger.Error("failed to create observation", zap.Error(err))
		return nil, err
	}

	return obs, nil
}

// UpdateObservation updates an existing observation
func (r *mutationResolver) UpdateObservation(ctx context.Context, id string, input model.UpdateObservationInput) (*domain.Observation, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("project ID not found in context")
	}

	// Build observation input from update input
	obsInput := &domain.ObservationInput{
		Name:          input.Name,
		EndTime:       input.EndTime,
		Level:         input.Level,
		StatusMessage: input.StatusMessage,
		Model:         input.Model,
	}

	// Convert map inputs to any type
	if input.Input != nil {
		obsInput.Input = input.Input
	}
	if input.Output != nil {
		obsInput.Output = input.Output
	}
	if input.Metadata != nil {
		obsInput.Metadata = input.Metadata
	}
	if input.ModelParameters != nil {
		obsInput.ModelParameters = input.ModelParameters
	}

	// Convert usage input
	if input.Usage != nil {
		usage := make(map[string]interface{})
		if input.Usage.PromptTokens != nil {
			usage["input"] = *input.Usage.PromptTokens
		}
		if input.Usage.CompletionTokens != nil {
			usage["output"] = *input.Usage.CompletionTokens
		}
		if input.Usage.TotalTokens != nil {
			usage["total"] = *input.Usage.TotalTokens
		}
		if len(usage) > 0 {
			obsInput.Usage = usage
		}
	}

	obs, err := r.ingestionService.UpdateObservation(ctx, projectID, id, obsInput)
	if err != nil {
		r.logger.Error("failed to update observation", zap.Error(err))
		return nil, err
	}

	return obs, nil
}

// ======== SCORE MUTATIONS ========

// CreateScore creates a new score
func (r *mutationResolver) CreateScore(ctx context.Context, input model.CreateScoreInput) (*domain.Score, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("project ID not found in context")
	}

	scoreInput := &domain.ScoreInput{
		TraceID: input.TraceId,
		Name:    input.Name,
		Source:  domain.ScoreSourceAPI,
	}

	if input.ObservationId != nil {
		scoreInput.ObservationID = input.ObservationId
	}
	if input.Value != nil {
		scoreInput.Value = input.Value
	}
	if input.StringValue != nil {
		scoreInput.StringValue = input.StringValue
	}
	if input.DataType != nil {
		scoreInput.DataType = *input.DataType
	} else {
		if input.Value != nil {
			scoreInput.DataType = domain.ScoreDataTypeNumeric
		} else if input.StringValue != nil {
			scoreInput.DataType = domain.ScoreDataTypeCategorical
		}
	}
	if input.Source != nil {
		scoreInput.Source = *input.Source
	}
	if input.Comment != nil {
		scoreInput.Comment = input.Comment
	}

	score, err := r.scoreService.Create(ctx, projectID, scoreInput)
	if err != nil {
		r.logger.Error("failed to create score", zap.Error(err))
		return nil, err
	}

	return score, nil
}

// UpdateScore updates an existing score
func (r *mutationResolver) UpdateScore(ctx context.Context, id string, input model.UpdateScoreInput) (*domain.Score, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("project ID not found in context")
	}

	scoreInput := &domain.ScoreInput{}
	if input.Value != nil {
		scoreInput.Value = input.Value
	}
	if input.StringValue != nil {
		scoreInput.StringValue = input.StringValue
	}
	if input.Comment != nil {
		scoreInput.Comment = input.Comment
	}

	score, err := r.scoreService.Update(ctx, projectID, id, scoreInput)
	if err != nil {
		r.logger.Error("failed to update score", zap.Error(err))
		return nil, err
	}

	return score, nil
}

// DeleteScore deletes a score
func (r *mutationResolver) DeleteScore(ctx context.Context, id string) (bool, error) {
	_, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return false, fmt.Errorf("project ID not found in context")
	}

	// Score deletion is not yet supported
	return false, fmt.Errorf("score deletion is not yet supported")
}

// ======== PROMPT MUTATIONS ========

// CreatePrompt creates a new prompt
func (r *mutationResolver) CreatePrompt(ctx context.Context, input model.CreatePromptInput) (*domain.Prompt, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("project ID not found in context")
	}
	userID, _ := ctx.Value(ContextKeyUserID).(uuid.UUID)

	// Build content from prompt or messages
	var content string
	if input.Prompt != nil {
		content = *input.Prompt
	} else if len(input.Messages) > 0 {
		// Convert messages to JSON content
		if data, err := json.Marshal(input.Messages); err == nil {
			content = string(data)
		}
	}

	promptInput := &domain.PromptInput{
		Name:    input.Name,
		Type:    input.Type,
		Content: content,
	}

	if input.Description != nil {
		promptInput.Description = input.Description
	}
	if input.Config != nil {
		promptInput.Config = input.Config
	}
	if input.Labels != nil {
		promptInput.Labels = input.Labels
	}

	prompt, err := r.promptService.Create(ctx, projectID, promptInput, userID)
	if err != nil {
		r.logger.Error("failed to create prompt", zap.Error(err))
		return nil, err
	}

	return prompt, nil
}

// UpdatePrompt updates a prompt (creates new version)
func (r *mutationResolver) UpdatePrompt(ctx context.Context, name string, input model.UpdatePromptInput) (*domain.Prompt, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("project ID not found in context")
	}
	userID, _ := ctx.Value(ContextKeyUserID).(uuid.UUID)

	prompt, err := r.promptService.GetByName(ctx, projectID, name)
	if err != nil {
		return nil, err
	}

	// Create new version if prompt/messages are provided
	if input.Prompt != nil || len(input.Messages) > 0 {
		var content string
		if input.Prompt != nil {
			content = *input.Prompt
		} else if len(input.Messages) > 0 {
			if data, err := json.Marshal(input.Messages); err == nil {
				content = string(data)
			}
		}

		versionInput := &domain.PromptVersionInput{
			Content: content,
		}
		if input.Config != nil {
			versionInput.Config = input.Config
		}
		if input.Labels != nil {
			versionInput.Labels = input.Labels
		}
		if _, err := r.promptService.CreateVersion(ctx, prompt.ID, versionInput, userID); err != nil {
			r.logger.Error("failed to create prompt version", zap.Error(err))
			return nil, err
		}
	}

	// Update prompt metadata if description changed
	if input.Description != nil {
		updateInput := &domain.PromptInput{
			Name:        prompt.Name,
			Description: input.Description,
		}
		if _, err := r.promptService.Update(ctx, prompt.ID, updateInput); err != nil {
			r.logger.Error("failed to update prompt", zap.Error(err))
			return nil, err
		}
	}

	// Refresh prompt
	return r.promptService.Get(ctx, prompt.ID)
}

// DeletePrompt deletes a prompt
func (r *mutationResolver) DeletePrompt(ctx context.Context, name string) (bool, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return false, fmt.Errorf("project ID not found in context")
	}

	prompt, err := r.promptService.GetByName(ctx, projectID, name)
	if err != nil {
		return false, err
	}

	if err := r.promptService.Delete(ctx, prompt.ID); err != nil {
		r.logger.Error("failed to delete prompt", zap.Error(err))
		return false, err
	}

	return true, nil
}

// SetPromptLabel sets a label on a prompt version
func (r *mutationResolver) SetPromptLabel(ctx context.Context, name string, version int, label string) (bool, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return false, fmt.Errorf("project ID not found in context")
	}

	prompt, err := r.promptService.GetByName(ctx, projectID, name)
	if err != nil {
		return false, err
	}

	if err := r.promptService.SetVersionLabel(ctx, prompt.ID, version, label, true); err != nil {
		r.logger.Error("failed to set prompt label", zap.Error(err))
		return false, err
	}

	return true, nil
}

// RemovePromptLabel removes a label from a prompt
func (r *mutationResolver) RemovePromptLabel(ctx context.Context, name string, label string) (bool, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return false, fmt.Errorf("project ID not found in context")
	}

	prompt, err := r.promptService.GetByName(ctx, projectID, name)
	if err != nil {
		return false, err
	}

	// Remove label from all versions (pass 0 to remove from all, add=false)
	if err := r.promptService.SetVersionLabel(ctx, prompt.ID, 0, label, false); err != nil {
		r.logger.Error("failed to remove prompt label", zap.Error(err))
		return false, err
	}

	return true, nil
}

// ======== DATASET MUTATIONS ========

// CreateDataset creates a new dataset
func (r *mutationResolver) CreateDataset(ctx context.Context, input model.CreateDatasetInput) (*domain.Dataset, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("project ID not found in context")
	}

	datasetInput := &domain.DatasetInput{
		Name: input.Name,
	}

	if input.Description != nil {
		datasetInput.Description = *input.Description
	}
	if input.Metadata != nil {
		datasetInput.Metadata = input.Metadata
	}

	dataset, err := r.datasetService.Create(ctx, projectID, datasetInput)
	if err != nil {
		r.logger.Error("failed to create dataset", zap.Error(err))
		return nil, err
	}

	return dataset, nil
}

// UpdateDataset updates a dataset
func (r *mutationResolver) UpdateDataset(ctx context.Context, id uuid.UUID, input model.UpdateDatasetInput) (*domain.Dataset, error) {
	// Get existing dataset
	dataset, err := r.datasetService.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Build update input
	updateInput := &domain.DatasetInput{
		Name:        dataset.Name,
		Description: dataset.Description,
	}

	if input.Name != nil {
		updateInput.Name = *input.Name
	}
	if input.Description != nil {
		updateInput.Description = *input.Description
	}
	if input.Metadata != nil {
		updateInput.Metadata = input.Metadata
	}

	updatedDataset, err := r.datasetService.Update(ctx, id, updateInput)
	if err != nil {
		r.logger.Error("failed to update dataset", zap.Error(err))
		return nil, err
	}

	return updatedDataset, nil
}

// DeleteDataset deletes a dataset
func (r *mutationResolver) DeleteDataset(ctx context.Context, id uuid.UUID) (bool, error) {
	if err := r.datasetService.Delete(ctx, id); err != nil {
		r.logger.Error("failed to delete dataset", zap.Error(err))
		return false, err
	}

	return true, nil
}

// CreateDatasetItem creates a new dataset item
func (r *mutationResolver) CreateDatasetItem(ctx context.Context, datasetID uuid.UUID, input model.CreateDatasetItemInput) (*domain.DatasetItem, error) {
	itemInput := &domain.DatasetItemInput{
		Input:               input.Input,
		ExpectedOutput:      input.ExpectedOutput,
		Metadata:            input.Metadata,
		SourceTraceID:       input.SourceTraceId,
		SourceObservationID: input.SourceObservationId,
	}

	item, err := r.datasetService.AddItem(ctx, datasetID, itemInput)
	if err != nil {
		r.logger.Error("failed to create dataset item", zap.Error(err))
		return nil, err
	}

	return item, nil
}

// UpdateDatasetItem updates a dataset item
func (r *mutationResolver) UpdateDatasetItem(ctx context.Context, id uuid.UUID, input model.UpdateDatasetItemInput) (*domain.DatasetItem, error) {
	updateInput := &domain.DatasetItemUpdateInput{
		Input:          input.Input,
		ExpectedOutput: input.ExpectedOutput,
		Metadata:       input.Metadata,
		Status:         input.Status,
	}

	item, err := r.datasetService.UpdateItem(ctx, id, updateInput)
	if err != nil {
		r.logger.Error("failed to update dataset item", zap.Error(err))
		return nil, err
	}

	return item, nil
}

// DeleteDatasetItem deletes a dataset item
func (r *mutationResolver) DeleteDatasetItem(ctx context.Context, id uuid.UUID) (bool, error) {
	if err := r.datasetService.DeleteItem(ctx, id); err != nil {
		r.logger.Error("failed to delete dataset item", zap.Error(err))
		return false, err
	}

	return true, nil
}

// CreateDatasetRun creates a new dataset run
func (r *mutationResolver) CreateDatasetRun(ctx context.Context, datasetID uuid.UUID, input model.CreateDatasetRunInput) (*domain.DatasetRun, error) {
	runInput := &domain.DatasetRunInput{
		Name: input.Name,
	}

	if input.Description != nil {
		runInput.Description = *input.Description
	}
	if input.Metadata != nil {
		runInput.Metadata = input.Metadata
	}

	run, err := r.datasetService.CreateRun(ctx, datasetID, runInput)
	if err != nil {
		r.logger.Error("failed to create dataset run", zap.Error(err))
		return nil, err
	}

	return run, nil
}

// AddDatasetRunItem adds an item result to a dataset run
func (r *mutationResolver) AddDatasetRunItem(ctx context.Context, runID uuid.UUID, input model.AddDatasetRunItemInput) (*domain.DatasetRunItem, error) {
	runItemInput := &domain.DatasetRunItemInput{
		DatasetItemID: input.DatasetItemId.String(),
		ObservationID: input.ObservationId,
	}

	if input.TraceId != nil {
		runItemInput.TraceID = *input.TraceId
	}

	runItem, err := r.datasetService.AddRunItem(ctx, runID, runItemInput)
	if err != nil {
		r.logger.Error("failed to add dataset run item", zap.Error(err))
		return nil, err
	}

	return runItem, nil
}

// ======== EVALUATOR MUTATIONS ========

// CreateEvaluator creates a new evaluator
func (r *mutationResolver) CreateEvaluator(ctx context.Context, input model.CreateEvaluatorInput) (*domain.Evaluator, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("project ID not found in context")
	}
	userID, _ := ctx.Value(ContextKeyUserID).(uuid.UUID)

	evalInput := &domain.EvaluatorInput{
		Name:      input.Name,
		Type:      &input.Type,
		ScoreName: input.Name, // Use name as score name by default
	}

	if input.Description != nil {
		evalInput.Description = input.Description
	}
	if input.Config != nil {
		evalInput.Config = input.Config
	}
	if input.SamplingRate != nil {
		evalInput.SamplingRate = input.SamplingRate
	}
	if input.TargetFilter != nil {
		evalInput.TargetFilter = input.TargetFilter
	}

	evaluator, err := r.evalService.Create(ctx, projectID, evalInput, userID)
	if err != nil {
		r.logger.Error("failed to create evaluator", zap.Error(err))
		return nil, err
	}

	return evaluator, nil
}

// UpdateEvaluator updates an evaluator
func (r *mutationResolver) UpdateEvaluator(ctx context.Context, id uuid.UUID, input model.UpdateEvaluatorInput) (*domain.Evaluator, error) {
	updateInput := &domain.EvaluatorUpdateInput{
		Name:        input.Name,
		Description: input.Description,
		Enabled:     input.IsActive,
	}

	if input.Config != nil {
		updateInput.Config = input.Config
	}
	if input.SamplingRate != nil {
		updateInput.SamplingRate = input.SamplingRate
	}
	if input.TargetFilter != nil {
		updateInput.TargetFilter = input.TargetFilter
	}

	evaluator, err := r.evalService.Update(ctx, id, updateInput)
	if err != nil {
		r.logger.Error("failed to update evaluator", zap.Error(err))
		return nil, err
	}

	return evaluator, nil
}

// DeleteEvaluator deletes an evaluator
func (r *mutationResolver) DeleteEvaluator(ctx context.Context, id uuid.UUID) (bool, error) {
	if err := r.evalService.Delete(ctx, id); err != nil {
		r.logger.Error("failed to delete evaluator", zap.Error(err))
		return false, err
	}

	return true, nil
}

// ======== ORGANIZATION MUTATIONS ========

// CreateOrganization creates a new organization
func (r *mutationResolver) CreateOrganization(ctx context.Context, name string) (*domain.Organization, error) {
	userID, ok := ctx.Value(ContextKeyUserID).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("user ID not found in context")
	}

	org, err := r.orgService.Create(ctx, name, userID)
	if err != nil {
		r.logger.Error("failed to create organization", zap.Error(err))
		return nil, err
	}

	return org, nil
}

// UpdateOrganization updates an organization
func (r *mutationResolver) UpdateOrganization(ctx context.Context, id uuid.UUID, name string) (*domain.Organization, error) {
	org, err := r.orgService.Update(ctx, id, name)
	if err != nil {
		r.logger.Error("failed to update organization", zap.Error(err))
		return nil, err
	}

	return org, nil
}

// DeleteOrganization deletes an organization
func (r *mutationResolver) DeleteOrganization(ctx context.Context, id uuid.UUID) (bool, error) {
	if err := r.orgService.Delete(ctx, id); err != nil {
		r.logger.Error("failed to delete organization", zap.Error(err))
		return false, err
	}

	return true, nil
}

// ======== PROJECT MUTATIONS ========

// CreateProject creates a new project
func (r *mutationResolver) CreateProject(ctx context.Context, input model.CreateProjectInput) (*domain.Project, error) {
	userID, _ := ctx.Value(ContextKeyUserID).(uuid.UUID)

	projectInput := &service.ProjectInput{
		Name: input.Name,
	}

	if input.Description != nil {
		projectInput.Description = *input.Description
	}
	if input.RetentionDays != nil {
		projectInput.RetentionDays = input.RetentionDays
	}
	if input.RateLimitPerMin != nil {
		projectInput.RateLimitPerMin = input.RateLimitPerMin
	}

	project, err := r.projectService.Create(ctx, input.OrganizationId, projectInput, userID)
	if err != nil {
		r.logger.Error("failed to create project", zap.Error(err))
		return nil, err
	}

	return project, nil
}

// UpdateProject updates a project
func (r *mutationResolver) UpdateProject(ctx context.Context, id uuid.UUID, input model.UpdateProjectInput) (*domain.Project, error) {
	// Get existing project to build update
	project, err := r.projectService.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	updateInput := &service.ProjectInput{
		Name:        project.Name,
		Description: project.Description,
	}

	if input.Name != nil {
		updateInput.Name = *input.Name
	}
	if input.Description != nil {
		updateInput.Description = *input.Description
	}
	if input.RetentionDays != nil {
		updateInput.RetentionDays = input.RetentionDays
	}
	if input.RateLimitPerMin != nil {
		updateInput.RateLimitPerMin = input.RateLimitPerMin
	}

	updatedProject, err := r.projectService.Update(ctx, id, updateInput)
	if err != nil {
		r.logger.Error("failed to update project", zap.Error(err))
		return nil, err
	}

	return updatedProject, nil
}

// DeleteProject deletes a project
func (r *mutationResolver) DeleteProject(ctx context.Context, id uuid.UUID) (bool, error) {
	if err := r.projectService.Delete(ctx, id); err != nil {
		r.logger.Error("failed to delete project", zap.Error(err))
		return false, err
	}

	return true, nil
}

// ======== API KEY MUTATIONS ========

// CreateAPIKey creates a new API key
func (r *mutationResolver) CreateAPIKey(ctx context.Context, input model.CreateAPIKeyInput) (*model.APIKeyWithSecret, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("project ID not found in context")
	}
	userID, _ := ctx.Value(ContextKeyUserID).(uuid.UUID)

	apiKeyInput := &domain.APIKeyInput{
		Name:      input.Name,
		Scopes:    input.Scopes,
		ExpiresAt: input.ExpiresAt,
	}

	result, err := r.authService.CreateAPIKey(ctx, projectID, apiKeyInput, userID)
	if err != nil {
		r.logger.Error("failed to create API key", zap.Error(err))
		return nil, err
	}

	return &model.APIKeyWithSecret{
		ID:         result.APIKey.ID,
		Name:       result.APIKey.Name,
		Key:        result.SecretKey,
		DisplayKey: result.APIKey.SecretKeyPreview,
		Scopes:     result.APIKey.Scopes,
		ExpiresAt:  result.APIKey.ExpiresAt,
		CreatedAt:  result.APIKey.CreatedAt,
	}, nil
}

// DeleteAPIKey deletes an API key
func (r *mutationResolver) DeleteAPIKey(ctx context.Context, id uuid.UUID) (bool, error) {
	if err := r.authService.DeleteAPIKey(ctx, id); err != nil {
		r.logger.Error("failed to delete API key", zap.Error(err))
		return false, err
	}

	return true, nil
}

// ======== HELPER TYPES ========

// mutationResolver implements Mutation resolvers
type mutationResolver struct {
	*Resolver
}

// Mutation returns the mutation resolver
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// MutationResolver interface
type MutationResolver interface {
	CreateTrace(ctx context.Context, input model.CreateTraceInput) (*domain.Trace, error)
	UpdateTrace(ctx context.Context, id string, input model.UpdateTraceInput) (*domain.Trace, error)
	DeleteTrace(ctx context.Context, id string) (bool, error)
	CreateSpan(ctx context.Context, input model.CreateObservationInput) (*domain.Observation, error)
	CreateGeneration(ctx context.Context, input model.CreateGenerationInput) (*domain.Observation, error)
	CreateEvent(ctx context.Context, input model.CreateObservationInput) (*domain.Observation, error)
	UpdateObservation(ctx context.Context, id string, input model.UpdateObservationInput) (*domain.Observation, error)
	CreateScore(ctx context.Context, input model.CreateScoreInput) (*domain.Score, error)
	UpdateScore(ctx context.Context, id string, input model.UpdateScoreInput) (*domain.Score, error)
	DeleteScore(ctx context.Context, id string) (bool, error)
	CreatePrompt(ctx context.Context, input model.CreatePromptInput) (*domain.Prompt, error)
	UpdatePrompt(ctx context.Context, name string, input model.UpdatePromptInput) (*domain.Prompt, error)
	DeletePrompt(ctx context.Context, name string) (bool, error)
	SetPromptLabel(ctx context.Context, name string, version int, label string) (bool, error)
	RemovePromptLabel(ctx context.Context, name string, label string) (bool, error)
	CreateDataset(ctx context.Context, input model.CreateDatasetInput) (*domain.Dataset, error)
	UpdateDataset(ctx context.Context, id uuid.UUID, input model.UpdateDatasetInput) (*domain.Dataset, error)
	DeleteDataset(ctx context.Context, id uuid.UUID) (bool, error)
	CreateDatasetItem(ctx context.Context, datasetID uuid.UUID, input model.CreateDatasetItemInput) (*domain.DatasetItem, error)
	UpdateDatasetItem(ctx context.Context, id uuid.UUID, input model.UpdateDatasetItemInput) (*domain.DatasetItem, error)
	DeleteDatasetItem(ctx context.Context, id uuid.UUID) (bool, error)
	CreateDatasetRun(ctx context.Context, datasetID uuid.UUID, input model.CreateDatasetRunInput) (*domain.DatasetRun, error)
	AddDatasetRunItem(ctx context.Context, runID uuid.UUID, input model.AddDatasetRunItemInput) (*domain.DatasetRunItem, error)
	CreateEvaluator(ctx context.Context, input model.CreateEvaluatorInput) (*domain.Evaluator, error)
	UpdateEvaluator(ctx context.Context, id uuid.UUID, input model.UpdateEvaluatorInput) (*domain.Evaluator, error)
	DeleteEvaluator(ctx context.Context, id uuid.UUID) (bool, error)
	CreateOrganization(ctx context.Context, name string) (*domain.Organization, error)
	UpdateOrganization(ctx context.Context, id uuid.UUID, name string) (*domain.Organization, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) (bool, error)
	CreateProject(ctx context.Context, input model.CreateProjectInput) (*domain.Project, error)
	UpdateProject(ctx context.Context, id uuid.UUID, input model.UpdateProjectInput) (*domain.Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) (bool, error)
	CreateAPIKey(ctx context.Context, input model.CreateAPIKeyInput) (*model.APIKeyWithSecret, error)
	DeleteAPIKey(ctx context.Context, id uuid.UUID) (bool, error)
}

// ======== HELPER FUNCTIONS ========

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func boolValue(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

func timeValue(t *time.Time, def time.Time) time.Time {
	if t == nil {
		return def
	}
	return *t
}

func timeValuePtr(t *time.Time) *time.Time {
	return t
}

func generateTraceID() string {
	return fmt.Sprintf("%032x", uuid.New())
}

func generateSpanID() string {
	id := uuid.New()
	return fmt.Sprintf("%016x", id[:8])
}

func slugify(s string) string {
	// Simple slugify - lowercase and replace spaces with dashes
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result = append(result, c+'a'-'A')
		} else if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			result = append(result, c)
		} else if c == ' ' || c == '-' || c == '_' {
			if len(result) > 0 && result[len(result)-1] != '-' {
				result = append(result, '-')
			}
		}
	}
	return string(result)
}

