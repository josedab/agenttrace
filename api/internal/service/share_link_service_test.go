package service

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

type shareLinkRepositoryStub struct {
	links map[uuid.UUID]*domain.ShareLink
	hash  map[string]uuid.UUID
}

func newShareLinkRepositoryStub() *shareLinkRepositoryStub {
	return &shareLinkRepositoryStub{
		links: map[uuid.UUID]*domain.ShareLink{},
		hash:  map[string]uuid.UUID{},
	}
}

func (r *shareLinkRepositoryStub) Create(_ context.Context, link *domain.ShareLink) error {
	copy := *link
	copy.TokenHash = append([]byte(nil), link.TokenHash...)
	r.links[link.ID] = &copy
	r.hash[base64.RawURLEncoding.EncodeToString(link.TokenHash)] = link.ID
	return nil
}

func (r *shareLinkRepositoryStub) GetByTokenHash(
	_ context.Context,
	tokenHash []byte,
) (*domain.ShareLink, error) {
	id, ok := r.hash[base64.RawURLEncoding.EncodeToString(tokenHash)]
	if !ok {
		return nil, apperrors.NotFound("share link")
	}
	copy := *r.links[id]
	return &copy, nil
}

func (r *shareLinkRepositoryStub) GetByID(
	_ context.Context,
	projectID, linkID uuid.UUID,
) (*domain.ShareLink, error) {
	link, ok := r.links[linkID]
	if !ok || link.ProjectID != projectID {
		return nil, apperrors.NotFound("share link")
	}
	copy := *link
	return &copy, nil
}

func (r *shareLinkRepositoryStub) Revoke(
	_ context.Context,
	projectID, linkID uuid.UUID,
	revokedAt time.Time,
) error {
	link, err := r.GetByID(context.Background(), projectID, linkID)
	if err != nil {
		return err
	}
	link.RevokedAt = &revokedAt
	r.links[linkID] = link
	return nil
}

type shareTraceRepositoryStub struct {
	projectID uuid.UUID
	traceID   string
}

func (r *shareTraceRepositoryStub) GetByID(
	_ context.Context,
	projectID uuid.UUID,
	traceID string,
) (*domain.Trace, error) {
	if projectID != r.projectID || !sameTraceID(traceID, r.traceID) {
		return nil, apperrors.NotFound("trace")
	}
	return &domain.Trace{ID: r.traceID, ProjectID: r.projectID}, nil
}

type shareReplayPlanReaderStub struct {
	plan *domain.ReplayPlan
}

func (r *shareReplayPlanReaderStub) GetPlan(
	_ context.Context,
	projectID, planID uuid.UUID,
) (*domain.ReplayPlan, error) {
	if r.plan == nil || r.plan.ProjectID != projectID || r.plan.ID != planID {
		return nil, apperrors.NotFound("replay plan")
	}
	return r.plan, nil
}

func newShareLinkTestService(
	projectID uuid.UUID,
	traceID string,
	repository *shareLinkRepositoryStub,
) *ShareLinkService {
	timeline := &domain.ReplayTimeline{
		TraceID:   uuid.MustParse(traceID),
		TraceName: "deploy sk-at-supersecret",
		StartTime: time.Date(2026, 7, 25, 9, 0, 0, 0, time.UTC),
		Events: []domain.ReplayEvent{
			{
				ID:        "file-1",
				Type:      domain.ReplayEventFileOperation,
				Timestamp: time.Date(2026, 7, 25, 9, 1, 0, 0, time.UTC),
				Title:     "/private/repository/src/secret.go",
				Status:    "success",
				Data: domain.ReplayEventData{
					FilePath: "/private/repository/src/secret.go",
					Diff:     "+ password = secret",
				},
			},
			{
				ID:        "llm-1",
				Type:      domain.ReplayEventLLMCall,
				Timestamp: time.Date(2026, 7, 25, 9, 2, 0, 0, time.UTC),
				Title:     "Email user@example.com",
				Status:    "success",
				Data: domain.ReplayEventData{
					Model:        "gpt-4.1",
					Input:        "secret prompt",
					Output:       "secret output",
					TokensInput:  10,
					TokensOutput: 5,
				},
			},
		},
		Summary: domain.ReplaySummary{TotalEvents: 2, LLMCalls: 1, FileOperations: 1},
	}
	service := NewShareLinkService(
		repository,
		&shareTraceRepositoryStub{projectID: projectID, traceID: traceID},
		&replayTimelineProviderStub{timeline: timeline},
		&shareReplayPlanReaderStub{},
		NewSensitiveDataRedactor(),
		"https://app.example.com",
	)
	service.clock = func() time.Time {
		return time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	}
	return service
}

func TestShareLinkCreatesRandomHashedTokens(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New().String()
	repository := newShareLinkRepositoryStub()
	service := newShareLinkTestService(projectID, traceID, repository)

	first, err := service.Create(context.Background(), projectID, uuid.New(), domain.ShareLinkInput{
		ResourceType: domain.ShareResourceTrace,
		ResourceID:   traceID,
	})
	require.NoError(t, err)
	second, err := service.Create(context.Background(), projectID, uuid.New(), domain.ShareLinkInput{
		ResourceType: domain.ShareResourceTrace,
		ResourceID:   traceID,
	})
	require.NoError(t, err)

	assert.NotEqual(t, first.Token, second.Token)
	assert.Len(t, first.Token, 43)
	assert.Contains(t, first.URL, "/share/")
	assert.NotEqual(t, []byte(first.Token), repository.links[first.ID].TokenHash)
	assert.Len(t, repository.links[first.ID].TokenHash, 32)
}

func TestShareLinkResolveRedactsSourceAndSecrets(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New().String()
	repository := newShareLinkRepositoryStub()
	service := newShareLinkTestService(projectID, traceID, repository)
	created, err := service.Create(context.Background(), projectID, uuid.New(), domain.ShareLinkInput{
		ResourceType: domain.ShareResourceTrace,
		ResourceID:   traceID,
	})
	require.NoError(t, err)

	view, err := service.Resolve(context.Background(), created.Token)

	require.NoError(t, err)
	require.NotNil(t, view.Trace)
	assert.Contains(t, view.Trace.Name, "[REDACTED:api-key]")
	require.Len(t, view.Trace.Events, 2)
	assert.Equal(t, "File operation", view.Trace.Events[0].Title)
	assert.NotContains(t, view.Trace.Events[0].Title, "secret.go")
	assert.Contains(t, view.Trace.Events[1].Title, "[REDACTED:email]")
}

func TestShareLinkExpiryAndRevocation(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New().String()
	repository := newShareLinkRepositoryStub()
	service := newShareLinkTestService(projectID, traceID, repository)
	created, err := service.Create(context.Background(), projectID, uuid.New(), domain.ShareLinkInput{
		ResourceType:     domain.ShareResourceTrace,
		ResourceID:       traceID,
		ExpiresInSeconds: int64((10 * time.Minute).Seconds()),
	})
	require.NoError(t, err)

	require.NoError(t, service.Revoke(context.Background(), projectID, created.ID))
	_, err = service.Resolve(context.Background(), created.Token)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	expiring, err := service.Create(context.Background(), projectID, uuid.New(), domain.ShareLinkInput{
		ResourceType:     domain.ShareResourceTrace,
		ResourceID:       traceID,
		ExpiresInSeconds: int64((10 * time.Minute).Seconds()),
	})
	require.NoError(t, err)
	service.clock = func() time.Time { return expiring.ExpiresAt }
	_, err = service.Resolve(context.Background(), expiring.Token)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestShareLinkRejectsCrossProjectResource(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New().String()
	service := newShareLinkTestService(projectID, traceID, newShareLinkRepositoryStub())

	_, err := service.Create(context.Background(), uuid.New(), uuid.New(), domain.ShareLinkInput{
		ResourceType: domain.ShareResourceTrace,
		ResourceID:   traceID,
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestSensitiveDataRedactorIsDeterministic(t *testing.T) {
	redactor := NewSensitiveDataRedactor()
	value := "Bearer abcdefghijk user@example.com sk-at-secretvalue"

	first := redactor.RedactText(value)
	second := redactor.RedactText(value)

	assert.Equal(t, first, second)
	assert.NotContains(t, first, "abcdefghijk")
	assert.NotContains(t, first, "user@example.com")
	assert.NotContains(t, first, "secretvalue")
}
