package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

const (
	defaultShareExpiry = 7 * 24 * time.Hour
	minShareExpiry     = 5 * time.Minute
	maxShareExpiry     = 30 * 24 * time.Hour
)

// ShareLinkRepository defines hashed token persistence.
type ShareLinkRepository interface {
	Create(ctx context.Context, link *domain.ShareLink) error
	GetByTokenHash(ctx context.Context, tokenHash []byte) (*domain.ShareLink, error)
	GetByID(ctx context.Context, projectID, linkID uuid.UUID) (*domain.ShareLink, error)
	Revoke(ctx context.Context, projectID, linkID uuid.UUID, revokedAt time.Time) error
}

// ShareTraceRepository verifies trace ownership.
type ShareTraceRepository interface {
	GetByID(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error)
}

// ShareReplayPlanReader verifies replay plan ownership.
type ShareReplayPlanReader interface {
	GetPlan(ctx context.Context, projectID, planID uuid.UUID) (*domain.ReplayPlan, error)
}

// ShareLinkService creates and resolves server-redacted read-only links.
type ShareLinkService struct {
	repository       ShareLinkRepository
	traceRepository  ShareTraceRepository
	timelineProvider ReplayTimelineProvider
	replayPlans      ShareReplayPlanReader
	redactor         *SensitiveDataRedactor
	publicBaseURL    string
	clock            func() time.Time
	random           func([]byte) (int, error)
}

// NewShareLinkService creates a share link service.
func NewShareLinkService(
	repository ShareLinkRepository,
	traceRepository ShareTraceRepository,
	timelineProvider ReplayTimelineProvider,
	replayPlans ShareReplayPlanReader,
	redactor *SensitiveDataRedactor,
	publicBaseURL string,
) *ShareLinkService {
	return &ShareLinkService{
		repository:       repository,
		traceRepository:  traceRepository,
		timelineProvider: timelineProvider,
		replayPlans:      replayPlans,
		redactor:         redactor,
		publicBaseURL:    strings.TrimRight(publicBaseURL, "/"),
		clock:            time.Now,
		random:           rand.Read,
	}
}

// Create creates a cryptographically random expiring token.
func (s *ShareLinkService) Create(
	ctx context.Context,
	projectID, actorID uuid.UUID,
	input domain.ShareLinkInput,
) (*domain.ShareLinkCreated, error) {
	if !input.ResourceType.IsValid() {
		return nil, apperrors.Validation("invalid share resource type")
	}
	if input.ResourceID == "" {
		return nil, apperrors.Validation("resourceId is required")
	}
	if err := s.verifyResource(ctx, projectID, input.ResourceType, input.ResourceID); err != nil {
		return nil, err
	}

	expiry := defaultShareExpiry
	if input.ExpiresInSeconds != 0 {
		expiry = time.Duration(input.ExpiresInSeconds) * time.Second
	}
	if expiry < minShareExpiry || expiry > maxShareExpiry {
		return nil, apperrors.Validation("share expiry must be between 5 minutes and 30 days")
	}

	tokenBytes := make([]byte, 32)
	if _, err := s.random(tokenBytes); err != nil {
		return nil, fmt.Errorf("generate share token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	tokenHash := sha256.Sum256([]byte(token))
	now := s.clock().UTC()

	link := &domain.ShareLink{
		ID:               uuid.New(),
		ProjectID:        projectID,
		ResourceType:     input.ResourceType,
		ResourceID:       input.ResourceID,
		TokenHash:        tokenHash[:],
		RedactionVersion: 1,
		ExpiresAt:        now.Add(expiry),
		CreatedBy:        actorID,
		CreatedAt:        now,
	}
	if err := s.repository.Create(ctx, link); err != nil {
		return nil, err
	}

	url := "/share/" + token
	if s.publicBaseURL != "" {
		url = s.publicBaseURL + url
	}
	return &domain.ShareLinkCreated{
		ShareLink: *link,
		Token:     token,
		URL:       url,
	}, nil
}

// Resolve validates token state and returns a server-redacted view.
func (s *ShareLinkService) Resolve(
	ctx context.Context,
	token string,
) (*domain.SharedResourceView, error) {
	if len(token) < 32 || len(token) > 128 {
		return nil, apperrors.NotFound("share link")
	}
	tokenHash := sha256.Sum256([]byte(token))
	link, err := s.repository.GetByTokenHash(ctx, tokenHash[:])
	if err != nil {
		return nil, err
	}
	now := s.clock().UTC()
	if link.RevokedAt != nil || !now.Before(link.ExpiresAt) {
		return nil, apperrors.NotFound("share link")
	}

	view := &domain.SharedResourceView{
		ResourceType: link.ResourceType,
		ExpiresAt:    link.ExpiresAt,
	}
	switch link.ResourceType {
	case domain.ShareResourceTrace:
		timeline, err := s.timelineProvider.GetTimelineForTrace(
			ctx,
			link.ProjectID,
			link.ResourceID,
		)
		if err != nil {
			return nil, apperrors.NotFound("shared trace")
		}
		view.Trace = s.redactTimeline(timeline)
	case domain.ShareResourceReplayPlan:
		planID, err := uuid.Parse(link.ResourceID)
		if err != nil {
			return nil, apperrors.NotFound("shared replay plan")
		}
		plan, err := s.replayPlans.GetPlan(ctx, link.ProjectID, planID)
		if err != nil {
			return nil, apperrors.NotFound("shared replay plan")
		}
		shared := &domain.SharedReplayPlanView{
			PlanID:       plan.ID,
			TraceID:      plan.TraceID,
			Status:       plan.Status,
			Capabilities: plan.Capabilities,
		}
		if plan.Result != nil {
			comparison := plan.Result.Comparison
			comparison.Notes = redactStrings(s.redactor, comparison.Notes)
			shared.Comparison = &comparison
		}
		view.ReplayPlan = shared
	}
	return view, nil
}

// Revoke revokes a link only within its project.
func (s *ShareLinkService) Revoke(
	ctx context.Context,
	projectID, linkID uuid.UUID,
) error {
	if _, err := s.repository.GetByID(ctx, projectID, linkID); err != nil {
		return err
	}
	return s.repository.Revoke(ctx, projectID, linkID, s.clock().UTC())
}

func (s *ShareLinkService) verifyResource(
	ctx context.Context,
	projectID uuid.UUID,
	resourceType domain.ShareResourceType,
	resourceID string,
) error {
	switch resourceType {
	case domain.ShareResourceTrace:
		_, err := s.traceRepository.GetByID(ctx, projectID, resourceID)
		return err
	case domain.ShareResourceReplayPlan:
		planID, err := uuid.Parse(resourceID)
		if err != nil {
			return apperrors.Validation("invalid replay plan ID")
		}
		_, err = s.replayPlans.GetPlan(ctx, projectID, planID)
		return err
	default:
		return apperrors.Validation("invalid share resource type")
	}
}

func (s *ShareLinkService) redactTimeline(
	timeline *domain.ReplayTimeline,
) *domain.SharedTraceView {
	events := make([]domain.SharedReplayEvent, 0, len(timeline.Events))
	for _, event := range timeline.Events {
		title := event.Title
		switch event.Type {
		case domain.ReplayEventFileOperation:
			title = "File operation"
		case domain.ReplayEventTerminalCmd:
			title = "Terminal command"
		case domain.ReplayEventToolCall:
			title = "Tool call"
		case domain.ReplayEventGitOperation:
			title = "Git operation"
		case domain.ReplayEventUserInput:
			title = "User input"
		case domain.ReplayEventLLMCall,
			domain.ReplayEventCheckpoint,
			domain.ReplayEventAgentThought,
			domain.ReplayEventError:
			// These titles are useful after deterministic sensitive-data redaction.
		}
		events = append(events, domain.SharedReplayEvent{
			Type:       event.Type,
			Timestamp:  event.Timestamp,
			DurationMs: event.Duration,
			Title:      s.redactor.RedactText(title),
			Status:     s.redactor.RedactText(event.Status),
			Model:      s.redactor.RedactText(event.Data.Model),
			Tokens:     event.Data.TokensInput + event.Data.TokensOutput,
		})
	}
	return &domain.SharedTraceView{
		TraceID:    timeline.TraceID.String(),
		Name:       s.redactor.RedactText(timeline.TraceName),
		StartTime:  timeline.StartTime,
		EndTime:    timeline.EndTime,
		DurationMs: timeline.Duration,
		Summary:    timeline.Summary,
		Events:     events,
	}
}

func redactStrings(redactor *SensitiveDataRedactor, values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, redactor.RedactText(value))
	}
	return result
}
