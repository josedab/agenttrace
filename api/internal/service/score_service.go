package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// ScoreRepository defines score repository operations
type ScoreRepository interface {
	Create(ctx context.Context, score *domain.Score) error
	CreateBatch(ctx context.Context, scores []*domain.Score) error
	GetByID(ctx context.Context, projectID uuid.UUID, scoreID string) (*domain.Score, error)
	GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Score, error)
	GetByObservationID(ctx context.Context, projectID uuid.UUID, observationID string) ([]domain.Score, error)
	List(ctx context.Context, filter *domain.ScoreFilter, limit, offset int) (*domain.ScoreList, error)
	Update(ctx context.Context, score *domain.Score) error
	Delete(ctx context.Context, projectID, scoreID uuid.UUID) error
	GetStats(ctx context.Context, projectID uuid.UUID, name string) (*domain.ScoreStats, error)
	GetDistinctNames(ctx context.Context, projectID uuid.UUID) ([]string, error)
}

// ScoreService handles score operations
type ScoreService struct {
	scoreRepo       ScoreRepository
	traceRepo       TraceRepository
	observationRepo ObservationRepository
}

// NewScoreService creates a new score service
func NewScoreService(
	scoreRepo ScoreRepository,
	traceRepo TraceRepository,
	observationRepo ObservationRepository,
) *ScoreService {
	return &ScoreService{
		scoreRepo:       scoreRepo,
		traceRepo:       traceRepo,
		observationRepo: observationRepo,
	}
}

// Create creates a new score
func (s *ScoreService) Create(ctx context.Context, projectID uuid.UUID, input *domain.ScoreInput) (*domain.Score, error) {
	// Validate score (only if data type is explicitly set)
	if input.DataType != "" && !domain.ValidateScore(input.DataType, input.Value, input.StringValue, nil) {
		return nil, apperrors.Validation("invalid score value for data type")
	}

	// Verify trace exists
	_, err := s.traceRepo.GetByID(ctx, projectID, input.TraceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace: %w", err)
	}

	// Verify observation exists if provided
	if input.ObservationID != nil {
		_, err := s.observationRepo.GetByID(ctx, projectID, *input.ObservationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get observation: %w", err)
		}
	}

	// Handle optional comment
	var comment string
	if input.Comment != nil {
		comment = *input.Comment
	}

	now := time.Now()
	score := &domain.Score{
		ID:            uuid.New(),
		TraceID:       input.TraceID,
		ProjectID:     projectID,
		ObservationID: input.ObservationID,
		Name:          input.Name,
		Value:         input.Value,
		StringValue:   input.StringValue,
		DataType:      input.DataType,
		Source:        input.Source,
		Comment:       comment,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Set default source
	if score.Source == "" {
		score.Source = domain.ScoreSourceAPI
	}

	// Set default data type
	if score.DataType == "" {
		if input.Value != nil {
			score.DataType = domain.ScoreDataTypeNumeric
		} else if input.StringValue != nil {
			if *input.StringValue == "true" || *input.StringValue == "false" {
				score.DataType = domain.ScoreDataTypeBoolean
			} else {
				score.DataType = domain.ScoreDataTypeCategorical
			}
		}
	}

	if err := s.scoreRepo.Create(ctx, score); err != nil {
		return nil, fmt.Errorf("failed to create score: %w", err)
	}

	return score, nil
}

// CreateBatch creates multiple scores
func (s *ScoreService) CreateBatch(ctx context.Context, projectID uuid.UUID, inputs []*domain.ScoreInput) ([]*domain.Score, error) {
	now := time.Now()
	scores := make([]*domain.Score, 0, len(inputs))

	for _, input := range inputs {
		// Validate score (only if data type is explicitly set)
		if input.DataType != "" && !domain.ValidateScore(input.DataType, input.Value, input.StringValue, nil) {
			return nil, apperrors.Validation(fmt.Sprintf("invalid score for trace %s: invalid value for data type", input.TraceID))
		}

		// Handle optional comment
		var comment string
		if input.Comment != nil {
			comment = *input.Comment
		}

		score := &domain.Score{
			ID:            uuid.New(),
			TraceID:       input.TraceID,
			ProjectID:     projectID,
			ObservationID: input.ObservationID,
			Name:          input.Name,
			Value:         input.Value,
			StringValue:   input.StringValue,
			DataType:      input.DataType,
			Source:        input.Source,
			Comment:       comment,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if score.Source == "" {
			score.Source = domain.ScoreSourceAPI
		}

		scores = append(scores, score)
	}

	if err := s.scoreRepo.CreateBatch(ctx, scores); err != nil {
		return nil, fmt.Errorf("failed to create scores: %w", err)
	}

	return scores, nil
}

// Get retrieves a score by ID
func (s *ScoreService) Get(ctx context.Context, projectID uuid.UUID, scoreID string) (*domain.Score, error) {
	return s.scoreRepo.GetByID(ctx, projectID, scoreID)
}

// List retrieves scores with filtering
func (s *ScoreService) List(ctx context.Context, filter *domain.ScoreFilter, limit, offset int) (*domain.ScoreList, error) {
	return s.scoreRepo.List(ctx, filter, limit, offset)
}

// GetByTraceID retrieves scores for a trace
func (s *ScoreService) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Score, error) {
	return s.scoreRepo.GetByTraceID(ctx, projectID, traceID)
}

// GetByObservationID retrieves scores for an observation
func (s *ScoreService) GetByObservationID(ctx context.Context, projectID uuid.UUID, observationID string) ([]domain.Score, error) {
	return s.scoreRepo.GetByObservationID(ctx, projectID, observationID)
}

// Update updates an existing score
func (s *ScoreService) Update(ctx context.Context, projectID uuid.UUID, scoreID string, input *domain.ScoreInput) (*domain.Score, error) {
	score, err := s.scoreRepo.GetByID(ctx, projectID, scoreID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if input.Value != nil {
		score.Value = input.Value
	}
	if input.StringValue != nil {
		score.StringValue = input.StringValue
	}
	if input.Comment != nil {
		score.Comment = *input.Comment
	}
	score.UpdatedAt = time.Now()

	if err := s.scoreRepo.Update(ctx, score); err != nil {
		return nil, fmt.Errorf("failed to update score: %w", err)
	}

	return score, nil
}

// Delete deletes a score
func (s *ScoreService) Delete(ctx context.Context, projectID uuid.UUID, scoreID string) error {
	// Verify score exists
	score, err := s.scoreRepo.GetByID(ctx, projectID, scoreID)
	if err != nil {
		return err
	}

	return s.scoreRepo.Delete(ctx, projectID, score.ID)
}

// GetStats retrieves statistics for a score name
func (s *ScoreService) GetStats(ctx context.Context, projectID uuid.UUID, name string) (*domain.ScoreStats, error) {
	return s.scoreRepo.GetStats(ctx, projectID, name)
}

// GetScoreNames retrieves distinct score names for a project
func (s *ScoreService) GetScoreNames(ctx context.Context, projectID uuid.UUID) ([]string, error) {
	return s.scoreRepo.GetDistinctNames(ctx, projectID)
}

// SubmitFeedback submits user feedback as a score
func (s *ScoreService) SubmitFeedback(ctx context.Context, projectID uuid.UUID, traceID string, feedback *FeedbackInput) (*domain.Score, error) {
	input := &domain.ScoreInput{
		TraceID:  traceID,
		Name:     feedback.Name,
		Value:    feedback.Value,
		Source:   domain.ScoreSourceAnnotation,
		Comment:  feedback.Comment,
		DataType: feedback.DataType,
	}

	if input.Name == "" {
		input.Name = "user-feedback"
	}

	if input.DataType == "" {
		input.DataType = domain.ScoreDataTypeNumeric
	}

	return s.Create(ctx, projectID, input)
}

// FeedbackInput represents user feedback input
type FeedbackInput struct {
	Name     string            `json:"name"`
	Value    *float64          `json:"value"`
	DataType domain.ScoreDataType `json:"dataType"`
	Comment  *string           `json:"comment"`
}
