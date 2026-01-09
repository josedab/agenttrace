package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// EvaluatorRepository defines evaluator repository operations
type EvaluatorRepository interface {
	Create(ctx context.Context, eval *domain.Evaluator) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Evaluator, error)
	GetByName(ctx context.Context, projectID uuid.UUID, name string) (*domain.Evaluator, error)
	Update(ctx context.Context, eval *domain.Evaluator) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *domain.EvaluatorFilter, limit, offset int) (*domain.EvaluatorList, error)
	ListEnabled(ctx context.Context, projectID uuid.UUID) ([]domain.Evaluator, error)
	NameExists(ctx context.Context, projectID uuid.UUID, name string) (bool, error)

	// Template operations
	GetTemplate(ctx context.Context, id uuid.UUID) (*domain.EvaluatorTemplate, error)
	GetTemplateByName(ctx context.Context, name string) (*domain.EvaluatorTemplate, error)
	ListTemplates(ctx context.Context) ([]domain.EvaluatorTemplate, error)

	// Job operations
	CreateJob(ctx context.Context, job *domain.EvaluationJob) error
	GetJobByID(ctx context.Context, id uuid.UUID) (*domain.EvaluationJob, error)
	UpdateJob(ctx context.Context, job *domain.EvaluationJob) error
	ListPendingJobs(ctx context.Context, limit int) ([]domain.EvaluationJob, error)
	JobExists(ctx context.Context, evaluatorID uuid.UUID, traceID string, observationID *string) (bool, error)
	GetEvalCount(ctx context.Context, evaluatorID uuid.UUID) (int64, error)

	// Annotation queue operations
	CreateAnnotationQueue(ctx context.Context, queue *domain.AnnotationQueue) error
	GetAnnotationQueueByID(ctx context.Context, id uuid.UUID) (*domain.AnnotationQueue, error)
	UpdateAnnotationQueue(ctx context.Context, queue *domain.AnnotationQueue) error
	DeleteAnnotationQueue(ctx context.Context, id uuid.UUID) error
	ListAnnotationQueues(ctx context.Context, projectID uuid.UUID) ([]domain.AnnotationQueue, error)
	GetNextAnnotationItem(ctx context.Context, queueID uuid.UUID) (*domain.AnnotationQueueItem, error)
	CompleteAnnotationItem(ctx context.Context, itemID uuid.UUID, userID uuid.UUID) error
	GetAnnotationQueueStats(ctx context.Context, queueID uuid.UUID) (pending, completed int64, err error)
}

// EvalService handles evaluator operations
type EvalService struct {
	evalRepo     EvaluatorRepository
	scoreService *ScoreService
}

// NewEvalService creates a new eval service
func NewEvalService(evalRepo EvaluatorRepository, scoreService *ScoreService) *EvalService {
	return &EvalService{
		evalRepo:     evalRepo,
		scoreService: scoreService,
	}
}

// Create creates a new evaluator
func (s *EvalService) Create(ctx context.Context, projectID uuid.UUID, input *domain.EvaluatorInput, userID uuid.UUID) (*domain.Evaluator, error) {
	// Check if name exists
	exists, err := s.evalRepo.NameExists(ctx, projectID, input.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check name: %w", err)
	}
	if exists {
		return nil, apperrors.Validation("evaluator name already exists")
	}

	now := time.Now()

	eval := &domain.Evaluator{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      input.Name,
		ScoreName: input.ScoreName,
		CreatedBy: &userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Set defaults
	if input.Type != nil {
		eval.Type = *input.Type
	} else {
		eval.Type = domain.EvaluatorTypeLLMAsJudge
	}

	if input.Description != nil {
		eval.Description = *input.Description
	}

	if input.SamplingRate != nil {
		eval.SamplingRate = *input.SamplingRate
	} else {
		eval.SamplingRate = 1.0
	}

	if input.ScoreDataType != nil {
		eval.ScoreDataType = *input.ScoreDataType
	} else {
		eval.ScoreDataType = domain.ScoreDataTypeNumeric
	}

	if input.Enabled != nil {
		eval.Enabled = *input.Enabled
	} else {
		eval.Enabled = true
	}

	// Apply template if specified
	if input.TemplateID != nil {
		templateID, err := uuid.Parse(*input.TemplateID)
		if err != nil {
			return nil, apperrors.Validation("invalid template ID")
		}

		template, err := s.evalRepo.GetTemplate(ctx, templateID)
		if err != nil {
			return nil, err
		}

		eval.PromptTemplate = template.PromptTemplate
		eval.Variables = template.Variables
		eval.ScoreDataType = template.ScoreDataType
		eval.ScoreCategories = template.ScoreCategories
		if template.Config != "" {
			eval.Config = template.Config
		}
	} else {
		eval.PromptTemplate = input.PromptTemplate
		eval.Variables = input.Variables
		eval.ScoreCategories = input.ScoreCategories
	}

	// Marshal config
	if input.Config != nil {
		configBytes, err := json.Marshal(input.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		eval.Config = string(configBytes)
	}

	// Marshal target filter
	if input.TargetFilter != nil {
		filterBytes, err := json.Marshal(input.TargetFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal target filter: %w", err)
		}
		eval.TargetFilter = string(filterBytes)
	}

	if err := s.evalRepo.Create(ctx, eval); err != nil {
		return nil, fmt.Errorf("failed to create evaluator: %w", err)
	}

	return eval, nil
}

// Get retrieves an evaluator by ID
func (s *EvalService) Get(ctx context.Context, id uuid.UUID) (*domain.Evaluator, error) {
	eval, err := s.evalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load stats
	evalCount, err := s.evalRepo.GetEvalCount(ctx, id)
	if err == nil {
		eval.EvalCount = evalCount
	}

	return eval, nil
}

// Update updates an evaluator
func (s *EvalService) Update(ctx context.Context, id uuid.UUID, input *domain.EvaluatorUpdateInput) (*domain.Evaluator, error) {
	eval, err := s.evalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil && *input.Name != eval.Name {
		exists, err := s.evalRepo.NameExists(ctx, eval.ProjectID, *input.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check name: %w", err)
		}
		if exists {
			return nil, apperrors.Validation("evaluator name already exists")
		}
		eval.Name = *input.Name
	}

	if input.Description != nil {
		eval.Description = *input.Description
	}
	if input.PromptTemplate != nil {
		eval.PromptTemplate = *input.PromptTemplate
	}
	if len(input.Variables) > 0 {
		eval.Variables = input.Variables
	}
	if input.SamplingRate != nil {
		eval.SamplingRate = *input.SamplingRate
	}
	if len(input.ScoreCategories) > 0 {
		eval.ScoreCategories = input.ScoreCategories
	}
	if input.Enabled != nil {
		eval.Enabled = *input.Enabled
	}
	if input.Config != nil {
		configBytes, _ := json.Marshal(input.Config)
		eval.Config = string(configBytes)
	}
	if input.TargetFilter != nil {
		filterBytes, _ := json.Marshal(input.TargetFilter)
		eval.TargetFilter = string(filterBytes)
	}

	eval.UpdatedAt = time.Now()

	if err := s.evalRepo.Update(ctx, eval); err != nil {
		return nil, fmt.Errorf("failed to update evaluator: %w", err)
	}

	return eval, nil
}

// Delete deletes an evaluator
func (s *EvalService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.evalRepo.Delete(ctx, id)
}

// List retrieves evaluators with filtering
func (s *EvalService) List(ctx context.Context, filter *domain.EvaluatorFilter, limit, offset int) (*domain.EvaluatorList, error) {
	return s.evalRepo.List(ctx, filter, limit, offset)
}

// ListTemplates retrieves all evaluator templates
func (s *EvalService) ListTemplates(ctx context.Context) ([]domain.EvaluatorTemplate, error) {
	return s.evalRepo.ListTemplates(ctx)
}

// TriggerForTrace triggers evaluators for a new trace
func (s *EvalService) TriggerForTrace(ctx context.Context, projectID uuid.UUID, trace *domain.Trace) error {
	evaluators, err := s.evalRepo.ListEnabled(ctx, projectID)
	if err != nil {
		return err
	}

	for _, eval := range evaluators {
		// Check sampling rate
		if eval.SamplingRate < 1.0 && rand.Float64() > eval.SamplingRate {
			continue
		}

		// Parse target filter
		var targetFilter *domain.TargetFilter
		if eval.TargetFilter != "" {
			targetFilter = &domain.TargetFilter{}
			if err := json.Unmarshal([]byte(eval.TargetFilter), targetFilter); err != nil {
				continue
			}
		}

		// Check if trace matches filter
		if targetFilter != nil && !targetFilter.MatchesTrace(trace) {
			continue
		}

		// Check if job already exists
		exists, err := s.evalRepo.JobExists(ctx, eval.ID, trace.ID, nil)
		if err != nil || exists {
			continue
		}

		// Create evaluation job
		job := &domain.EvaluationJob{
			ID:          uuid.New(),
			EvaluatorID: eval.ID,
			TraceID:     trace.ID,
			Status:      domain.JobStatusPending,
			ScheduledAt: time.Now(),
			CreatedAt:   time.Now(),
		}

		if err := s.evalRepo.CreateJob(ctx, job); err != nil {
			continue
		}
	}

	return nil
}

// TriggerForObservation triggers evaluators for a new observation
func (s *EvalService) TriggerForObservation(ctx context.Context, projectID uuid.UUID, obs *domain.Observation) error {
	evaluators, err := s.evalRepo.ListEnabled(ctx, projectID)
	if err != nil {
		return err
	}

	for _, eval := range evaluators {
		// Check sampling rate
		if eval.SamplingRate < 1.0 && rand.Float64() > eval.SamplingRate {
			continue
		}

		// Parse target filter
		var targetFilter *domain.TargetFilter
		if eval.TargetFilter != "" {
			targetFilter = &domain.TargetFilter{}
			if err := json.Unmarshal([]byte(eval.TargetFilter), targetFilter); err != nil {
				continue
			}
		}

		// Check if observation matches filter
		if targetFilter != nil && !targetFilter.MatchesObservation(obs) {
			continue
		}

		// Check if job already exists
		exists, err := s.evalRepo.JobExists(ctx, eval.ID, obs.TraceID, &obs.ID)
		if err != nil || exists {
			continue
		}

		// Create evaluation job
		job := &domain.EvaluationJob{
			ID:            uuid.New(),
			EvaluatorID:   eval.ID,
			TraceID:       obs.TraceID,
			ObservationID: &obs.ID,
			Status:        domain.JobStatusPending,
			ScheduledAt:   time.Now(),
			CreatedAt:     time.Now(),
		}

		if err := s.evalRepo.CreateJob(ctx, job); err != nil {
			continue
		}
	}

	return nil
}

// ExecuteInput represents input for executing an evaluator
type ExecuteInput struct {
	TraceID       string  `json:"traceId"`
	ObservationID *string `json:"observationId,omitempty"`
}

// ExecuteResult represents the result of an evaluation execution
type ExecuteResult struct {
	JobID   uuid.UUID      `json:"jobId"`
	Status  string         `json:"status"`
	Score   *domain.Score  `json:"score,omitempty"`
	Message string         `json:"message,omitempty"`
}

// Execute runs an evaluator on a specific trace or observation
func (s *EvalService) Execute(ctx context.Context, evaluatorID uuid.UUID, input *ExecuteInput) (*ExecuteResult, error) {
	// Get evaluator
	eval, err := s.evalRepo.GetByID(ctx, evaluatorID)
	if err != nil {
		return nil, err
	}

	// Check if job already exists
	exists, err := s.evalRepo.JobExists(ctx, eval.ID, input.TraceID, input.ObservationID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing job: %w", err)
	}
	if exists {
		return &ExecuteResult{
			Status:  "skipped",
			Message: "Evaluation already exists for this trace/observation",
		}, nil
	}

	// Create evaluation job with immediate priority
	job := &domain.EvaluationJob{
		ID:          uuid.New(),
		EvaluatorID: eval.ID,
		TraceID:     input.TraceID,
		Status:      domain.JobStatusPending,
		ScheduledAt: time.Now(),
		CreatedAt:   time.Now(),
	}

	if input.ObservationID != nil {
		job.ObservationID = input.ObservationID
	}

	if err := s.evalRepo.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create evaluation job: %w", err)
	}

	return &ExecuteResult{
		JobID:   job.ID,
		Status:  "pending",
		Message: "Evaluation job created and queued for processing",
	}, nil
}

// GetJobStatus retrieves the status of an evaluation job
func (s *EvalService) GetJobStatus(ctx context.Context, jobID uuid.UUID) (*domain.EvaluationJob, error) {
	return s.evalRepo.GetJobByID(ctx, jobID)
}

// CreateAnnotationQueue creates a new annotation queue
func (s *EvalService) CreateAnnotationQueue(ctx context.Context, projectID uuid.UUID, name, description, scoreName string, scoreConfig map[string]any) (*domain.AnnotationQueue, error) {
	now := time.Now()

	queue := &domain.AnnotationQueue{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        name,
		Description: description,
		ScoreName:   scoreName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if scoreConfig != nil {
		configBytes, _ := json.Marshal(scoreConfig)
		queue.ScoreConfig = string(configBytes)
	}

	if err := s.evalRepo.CreateAnnotationQueue(ctx, queue); err != nil {
		return nil, fmt.Errorf("failed to create annotation queue: %w", err)
	}

	return queue, nil
}

// GetAnnotationQueue retrieves an annotation queue by ID
func (s *EvalService) GetAnnotationQueue(ctx context.Context, id uuid.UUID) (*domain.AnnotationQueue, error) {
	queue, err := s.evalRepo.GetAnnotationQueueByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load stats
	pending, completed, err := s.evalRepo.GetAnnotationQueueStats(ctx, id)
	if err == nil {
		queue.PendingCount = pending
		queue.CompletedCount = completed
	}

	return queue, nil
}

// ListAnnotationQueues retrieves annotation queues for a project
func (s *EvalService) ListAnnotationQueues(ctx context.Context, projectID uuid.UUID) ([]domain.AnnotationQueue, error) {
	queues, err := s.evalRepo.ListAnnotationQueues(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Load stats for each queue
	for i := range queues {
		pending, completed, err := s.evalRepo.GetAnnotationQueueStats(ctx, queues[i].ID)
		if err == nil {
			queues[i].PendingCount = pending
			queues[i].CompletedCount = completed
		}
	}

	return queues, nil
}

// GetNextAnnotationItem retrieves the next item to annotate
func (s *EvalService) GetNextAnnotationItem(ctx context.Context, queueID uuid.UUID) (*domain.AnnotationQueueItem, error) {
	return s.evalRepo.GetNextAnnotationItem(ctx, queueID)
}

// CompleteAnnotation completes an annotation with a score
func (s *EvalService) CompleteAnnotation(ctx context.Context, projectID uuid.UUID, queueID uuid.UUID, itemID uuid.UUID, userID uuid.UUID, value *float64, stringValue *string, comment *string) error {
	// Get queue for score name
	queue, err := s.evalRepo.GetAnnotationQueueByID(ctx, queueID)
	if err != nil {
		return err
	}

	// Get item for trace ID
	item, err := s.evalRepo.GetNextAnnotationItem(ctx, queueID)
	if err != nil {
		return err
	}

	// Create score
	scoreInput := &domain.ScoreInput{
		TraceID:       item.TraceID,
		ObservationID: item.ObservationID,
		Name:          queue.ScoreName,
		Value:         value,
		StringValue:   stringValue,
		Source:        domain.ScoreSourceAnnotation,
		Comment:       comment,
	}

	if _, err := s.scoreService.Create(ctx, projectID, scoreInput); err != nil {
		return fmt.Errorf("failed to create score: %w", err)
	}

	// Mark item as completed
	return s.evalRepo.CompleteAnnotationItem(ctx, itemID, userID)
}
