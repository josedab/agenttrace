package service

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

type digestWebhookRepositoryStub struct {
	mu            sync.Mutex
	projectID     uuid.UUID
	webhooks      map[uuid.UUID]*domain.Webhook
	deliveries    []domain.WebhookDelivery
	triggered     []uuid.UUID
	createErrors  map[uuid.UUID]error
	triggerErrors map[uuid.UUID]error
	lookupErrors  map[uuid.UUID]error
	lookupWindow  bool
	claims        map[string]time.Time
}

func (r *digestWebhookRepositoryStub) GetByProjectID(
	_ context.Context,
	projectID, webhookID uuid.UUID,
) (*domain.Webhook, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	webhook, ok := r.webhooks[webhookID]
	if !ok || projectID != r.projectID || webhook.ProjectID != projectID {
		return nil, apperrors.NotFound("webhook")
	}
	return webhook, nil
}

func (r *digestWebhookRepositoryStub) CreateDelivery(
	_ context.Context,
	delivery *domain.WebhookDelivery,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err, failing := r.createErrors[delivery.WebhookID]; failing {
		return err
	}
	r.deliveries = append(r.deliveries, *delivery)
	return nil
}

func (r *digestWebhookRepositoryStub) ClaimDelivery(
	_ context.Context,
	webhookID uuid.UUID,
	deliveryKey string,
	claimedAt, expiresAt time.Time,
) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.claims == nil {
		r.claims = map[string]time.Time{}
	}
	key := webhookID.String() + ":" + deliveryKey
	if expiry, exists := r.claims[key]; exists && expiry.After(claimedAt) {
		return false, nil
	}
	r.claims[key] = expiresAt
	return true, nil
}

func (r *digestWebhookRepositoryStub) FindRecentDelivery(
	_ context.Context,
	webhookID uuid.UUID,
	deliveryKey string,
	_ time.Time,
) (*domain.WebhookDelivery, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err, failing := r.lookupErrors[webhookID]; failing {
		return nil, err
	}
	if !r.lookupWindow || deliveryKey == "" {
		return nil, nil
	}
	for index := len(r.deliveries) - 1; index >= 0; index-- {
		delivery := r.deliveries[index]
		if delivery.WebhookID == webhookID && delivery.DeliveryKey == deliveryKey {
			return &delivery, nil
		}
	}
	return nil, nil
}

func (r *digestWebhookRepositoryStub) UpdateLastTriggered(
	_ context.Context,
	id uuid.UUID,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err, failing := r.triggerErrors[id]; failing {
		return err
	}
	r.triggered = append(r.triggered, id)
	return nil
}

type digestNotificationSenderStub struct {
	mu       sync.Mutex
	data     map[string]any
	sent     []uuid.UUID
	failures map[uuid.UUID]error
}

func (s *digestNotificationSenderStub) SendNotification(
	_ context.Context,
	webhook *domain.Webhook,
	eventType domain.EventType,
	data map[string]any,
) (*domain.WebhookDelivery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = data
	s.sent = append(s.sent, webhook.ID)
	if err, failing := s.failures[webhook.ID]; failing {
		return &domain.WebhookDelivery{
			ID:        uuid.New(),
			WebhookID: webhook.ID,
			EventType: eventType,
			Success:   false,
			CreatedAt: time.Now(),
		}, err
	}
	return &domain.WebhookDelivery{
		ID:        uuid.New(),
		WebhookID: webhook.ID,
		EventType: eventType,
		Success:   true,
		CreatedAt: time.Now(),
	}, nil
}

type digestProviderStub struct {
	digest *domain.OutcomeDigest
}

func (p *digestProviderStub) GetDigest(
	_ context.Context,
	_ uuid.UUID,
	_, _ time.Time,
) (*domain.OutcomeDigest, error) {
	return p.digest, nil
}

func TestTeamDigestDeliveryUsesProjectWebhooks(t *testing.T) {
	projectID := uuid.New()
	webhookID := uuid.New()
	repository := &digestWebhookRepositoryStub{
		projectID: projectID,
		webhooks: map[uuid.UUID]*domain.Webhook{
			webhookID: {
				ID:        webhookID,
				ProjectID: projectID,
				Type:      domain.WebhookTypeSlack,
				URL:       "https://8.8.8.8/digest",
				IsEnabled: true,
			},
		},
	}
	sender := &digestNotificationSenderStub{}
	service := NewTeamDigestDeliveryService(
		repository,
		sender,
		&digestProviderStub{digest: &domain.OutcomeDigest{
			ProjectID: projectID,
			Period: domain.OutcomePeriod{
				From: time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
				To:   time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
			},
			Title:      "Agent outcome digest",
			Summary:    "Eight successful runs.",
			Highlights: []string{"CI passed"},
		}},
	)

	result, err := service.Deliver(
		context.Background(),
		projectID,
		domain.TeamDigestDeliveryInput{
			WebhookIDs: []uuid.UUID{webhookID, webhookID},
			Window:     "7d",
		},
	)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Succeeded)
	assert.Zero(t, result.Failed)
	assert.Len(t, repository.deliveries, 1)
	assert.Equal(t, domain.EventTypeTeamDigest, repository.deliveries[0].EventType)
	assert.Contains(t, sender.data["markdown"], "Agent outcome digest")
	assert.Equal(t, []uuid.UUID{webhookID}, repository.triggered)
}

func TestTeamDigestDeliveryRejectsDisabledOrCrossProjectWebhook(t *testing.T) {
	projectID := uuid.New()
	webhookID := uuid.New()
	repository := &digestWebhookRepositoryStub{
		projectID: projectID,
		webhooks: map[uuid.UUID]*domain.Webhook{
			webhookID: {
				ID:        webhookID,
				ProjectID: projectID,
				Type:      domain.WebhookTypeDiscord,
				URL:       "https://8.8.8.8/digest",
				IsEnabled: false,
			},
		},
	}
	service := NewTeamDigestDeliveryService(
		repository,
		&digestNotificationSenderStub{},
		&digestProviderStub{digest: &domain.OutcomeDigest{}},
	)

	_, err := service.Deliver(
		context.Background(),
		projectID,
		domain.TeamDigestDeliveryInput{WebhookIDs: []uuid.UUID{webhookID}},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")

	_, err = service.Deliver(
		context.Background(),
		uuid.New(),
		domain.TeamDigestDeliveryInput{WebhookIDs: []uuid.UUID{webhookID}},
	)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestNotificationTeamDigestPayloads(t *testing.T) {
	service := NewNotificationService(nil, "")
	data := map[string]any{
		"title":    "Agent outcome digest",
		"summary":  "Summary",
		"markdown": "## Agent outcome digest\n\nSummary",
	}

	slackPayload, err := service.buildSlackPayload(domain.EventTypeTeamDigest, data)
	require.NoError(t, err)
	var slack domain.SlackMessage
	require.NoError(t, json.Unmarshal(slackPayload, &slack))
	require.Len(t, slack.Attachments, 1)
	assert.Equal(t, "Agent outcome digest", slack.Attachments[0].Title)
	assert.Contains(t, slack.Attachments[0].Text, "Summary")

	discordPayload, err := service.buildDiscordPayload(domain.EventTypeTeamDigest, data)
	require.NoError(t, err)
	var discord domain.DiscordMessage
	require.NoError(t, json.Unmarshal(discordPayload, &discord))
	require.Len(t, discord.Embeds, 1)
	assert.Equal(t, "Agent outcome digest", discord.Embeds[0].Title)
}

func newDigestTestFixture(
	projectID uuid.UUID,
	webhooks map[uuid.UUID]*domain.Webhook,
) (*digestWebhookRepositoryStub, *digestNotificationSenderStub, *digestProviderStub) {
	repository := &digestWebhookRepositoryStub{
		projectID:     projectID,
		webhooks:      webhooks,
		createErrors:  map[uuid.UUID]error{},
		triggerErrors: map[uuid.UUID]error{},
		lookupErrors:  map[uuid.UUID]error{},
		claims:        map[string]time.Time{},
	}
	sender := &digestNotificationSenderStub{failures: map[uuid.UUID]error{}}
	provider := &digestProviderStub{digest: &domain.OutcomeDigest{
		ProjectID: projectID,
		Period: domain.OutcomePeriod{
			From: time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
		},
		Title:   "Agent outcome digest",
		Summary: "Eight successful runs.",
	}}
	return repository, sender, provider
}

func digestWebhook(projectID uuid.UUID, webhookType domain.WebhookType) *domain.Webhook {
	return &domain.Webhook{
		ID:        uuid.New(),
		ProjectID: projectID,
		Type:      webhookType,
		URL:       "https://8.8.8.8/digest",
		IsEnabled: true,
	}
}

func TestTeamDigestValidatesEveryWebhookBeforeAnySend(t *testing.T) {
	projectID := uuid.New()
	valid := digestWebhook(projectID, domain.WebhookTypeSlack)
	disabled := digestWebhook(projectID, domain.WebhookTypeSlack)
	disabled.IsEnabled = false
	repository, sender, provider := newDigestTestFixture(projectID, map[uuid.UUID]*domain.Webhook{
		valid.ID:    valid,
		disabled.ID: disabled,
	})
	service := NewTeamDigestDeliveryService(repository, sender, provider, AllowAllOutbound())

	_, err := service.Deliver(context.Background(), projectID, domain.TeamDigestDeliveryInput{
		WebhookIDs: []uuid.UUID{valid.ID, disabled.ID},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
	// The valid webhook must not have been notified before the request failed.
	assert.Empty(t, sender.sent)
	assert.Empty(t, repository.deliveries)
}

func TestTeamDigestRejectsUnusableDestinationBeforeAnySend(t *testing.T) {
	projectID := uuid.New()
	valid := digestWebhook(projectID, domain.WebhookTypeSlack)
	internal := digestWebhook(projectID, domain.WebhookTypeGeneric)
	internal.URL = "https://127.0.0.1/internal"
	repository, sender, provider := newDigestTestFixture(projectID, map[uuid.UUID]*domain.Webhook{
		valid.ID:    valid,
		internal.ID: internal,
	})
	service := NewTeamDigestDeliveryService(repository, sender, provider, AllowAllOutbound())

	_, err := service.Deliver(context.Background(), projectID, domain.TeamDigestDeliveryInput{
		WebhookIDs: []uuid.UUID{valid.ID, internal.ID},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unusable destination")
	assert.Empty(t, sender.sent)
	assert.Empty(t, repository.deliveries)
}

func TestTeamDigestFailsFastInNoEgressMode(t *testing.T) {
	projectID := uuid.New()
	webhook := digestWebhook(projectID, domain.WebhookTypeSlack)
	repository, sender, provider := newDigestTestFixture(projectID, map[uuid.UUID]*domain.Webhook{
		webhook.ID: webhook,
	})
	service := NewTeamDigestDeliveryService(
		repository,
		sender,
		provider,
		NewEgressPolicy(true, true),
	)

	_, err := service.Deliver(context.Background(), projectID, domain.TeamDigestDeliveryInput{
		WebhookIDs: []uuid.UUID{webhook.ID},
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsUnprocessable(err))
	assert.Contains(t, err.Error(), "no-egress")
	assert.Empty(t, sender.sent)
	assert.Empty(t, repository.deliveries)
}

func TestTeamDigestReportsMixedResultsWithoutAborting(t *testing.T) {
	projectID := uuid.New()
	first := digestWebhook(projectID, domain.WebhookTypeSlack)
	failing := digestWebhook(projectID, domain.WebhookTypeDiscord)
	last := digestWebhook(projectID, domain.WebhookTypeGeneric)
	repository, sender, provider := newDigestTestFixture(projectID, map[uuid.UUID]*domain.Webhook{
		first.ID:   first,
		failing.ID: failing,
		last.ID:    last,
	})
	sender.failures[failing.ID] = errors.New("channel returned 500")
	service := NewTeamDigestDeliveryService(repository, sender, provider, AllowAllOutbound())

	result, err := service.Deliver(context.Background(), projectID, domain.TeamDigestDeliveryInput{
		WebhookIDs: []uuid.UUID{first.ID, failing.ID, last.ID},
	})

	require.NoError(t, err)
	assert.Equal(t, 2, result.Succeeded)
	assert.Equal(t, 1, result.Failed)
	require.Len(t, result.Deliveries, 3)
	// A failure in the middle must not stop the remaining channels.
	assert.Len(t, sender.sent, 3)
	assert.Len(t, repository.deliveries, 3)
	assert.ElementsMatch(t, []uuid.UUID{first.ID, last.ID}, repository.triggered)

	for _, delivery := range result.Deliveries {
		assert.NotEmpty(t, delivery.DeliveryKey)
		if delivery.WebhookID == failing.ID {
			assert.False(t, delivery.Success)
			assert.NotEmpty(t, delivery.Error)
		}
	}
	assert.NotEmpty(t, result.DeliveryKey)
}

func TestTeamDigestKeepsSuccessWhenBookkeepingFails(t *testing.T) {
	projectID := uuid.New()
	webhook := digestWebhook(projectID, domain.WebhookTypeSlack)
	repository, sender, provider := newDigestTestFixture(projectID, map[uuid.UUID]*domain.Webhook{
		webhook.ID: webhook,
	})
	repository.triggerErrors[webhook.ID] = errors.New("timestamp update failed")
	service := NewTeamDigestDeliveryService(repository, sender, provider, AllowAllOutbound())

	result, err := service.Deliver(context.Background(), projectID, domain.TeamDigestDeliveryInput{
		WebhookIDs: []uuid.UUID{webhook.ID},
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Succeeded)
	assert.Zero(t, result.Failed)
	assert.Equal(t, 1, result.TriggerUpdateFailures)
}

func TestTeamDigestSuppressesImmediateDuplicateDelivery(t *testing.T) {
	projectID := uuid.New()
	webhook := digestWebhook(projectID, domain.WebhookTypeSlack)
	repository, sender, provider := newDigestTestFixture(projectID, map[uuid.UUID]*domain.Webhook{
		webhook.ID: webhook,
	})
	repository.lookupWindow = true
	service := NewTeamDigestDeliveryService(repository, sender, provider, AllowAllOutbound())
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	service.clock = func() time.Time { return now }

	first, err := service.Deliver(context.Background(), projectID, domain.TeamDigestDeliveryInput{
		WebhookIDs: []uuid.UUID{webhook.ID},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, first.Succeeded)
	assert.Zero(t, first.Duplicates)

	// A double-clicked send reuses the recorded delivery instead of sending again.
	now = now.Add(3 * time.Second)
	second, err := service.Deliver(context.Background(), projectID, domain.TeamDigestDeliveryInput{
		WebhookIDs: []uuid.UUID{webhook.ID},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, second.Duplicates)
	assert.Equal(t, 1, second.Succeeded)
	assert.Len(t, sender.sent, 1)
	assert.Len(t, repository.deliveries, 1)
	assert.Equal(t, first.DeliveryKey, second.DeliveryKey)
}

func TestTeamDigestRecordsDeliveryPersistenceFailureAsFailure(t *testing.T) {
	projectID := uuid.New()
	first := digestWebhook(projectID, domain.WebhookTypeSlack)
	second := digestWebhook(projectID, domain.WebhookTypeGeneric)
	repository, sender, provider := newDigestTestFixture(projectID, map[uuid.UUID]*domain.Webhook{
		first.ID:  first,
		second.ID: second,
	})
	repository.createErrors[first.ID] = errors.New("delivery table unavailable")
	service := NewTeamDigestDeliveryService(repository, sender, provider, AllowAllOutbound())

	result, err := service.Deliver(context.Background(), projectID, domain.TeamDigestDeliveryInput{
		WebhookIDs: []uuid.UUID{first.ID, second.ID},
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Succeeded)
	assert.Equal(t, 1, result.Failed)
	assert.Len(t, sender.sent, 2)
	assert.Equal(t, []uuid.UUID{second.ID}, repository.triggered)
}

func TestTeamDigestDoesNotSendWhenDuplicateLookupFails(t *testing.T) {
	projectID := uuid.New()
	first := digestWebhook(projectID, domain.WebhookTypeSlack)
	second := digestWebhook(projectID, domain.WebhookTypeGeneric)
	repository, sender, provider := newDigestTestFixture(projectID, map[uuid.UUID]*domain.Webhook{
		first.ID:  first,
		second.ID: second,
	})
	repository.lookupErrors[first.ID] = errors.New("delivery lookup unavailable")
	service := NewTeamDigestDeliveryService(repository, sender, provider, AllowAllOutbound())

	result, err := service.Deliver(context.Background(), projectID, domain.TeamDigestDeliveryInput{
		WebhookIDs: []uuid.UUID{first.ID, second.ID},
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Succeeded)
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t, []uuid.UUID{second.ID}, sender.sent)
	require.Len(t, result.Deliveries, 2)
	assert.Equal(t, "delivery duplicate check failed", result.Deliveries[0].Error)
	assert.Len(t, repository.deliveries, 2)
}

type blockingDigestSender struct {
	mu      sync.Mutex
	started chan struct{}
	release chan struct{}
	once    sync.Once
	sent    []uuid.UUID
}

func (s *blockingDigestSender) SendNotification(
	_ context.Context,
	webhook *domain.Webhook,
	eventType domain.EventType,
	_ map[string]any,
) (*domain.WebhookDelivery, error) {
	s.mu.Lock()
	s.sent = append(s.sent, webhook.ID)
	s.mu.Unlock()
	s.once.Do(func() { close(s.started) })
	<-s.release
	return &domain.WebhookDelivery{
		ID:        uuid.New(),
		WebhookID: webhook.ID,
		EventType: eventType,
		Success:   true,
		CreatedAt: time.Now(),
	}, nil
}

func (s *blockingDigestSender) sendCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sent)
}

func TestTeamDigestAtomicClaimSuppressesConcurrentDuplicateSend(t *testing.T) {
	projectID := uuid.New()
	webhook := digestWebhook(projectID, domain.WebhookTypeSlack)
	repository, _, provider := newDigestTestFixture(projectID, map[uuid.UUID]*domain.Webhook{
		webhook.ID: webhook,
	})
	repository.lookupWindow = true
	sender := &blockingDigestSender{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	service := NewTeamDigestDeliveryService(repository, sender, provider, AllowAllOutbound())
	service.clock = func() time.Time {
		return time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	}
	input := domain.TeamDigestDeliveryInput{WebhookIDs: []uuid.UUID{webhook.ID}}

	firstResult := make(chan *domain.TeamDigestDeliveryResult, 1)
	firstError := make(chan error, 1)
	go func() {
		result, err := service.Deliver(context.Background(), projectID, input)
		firstResult <- result
		firstError <- err
	}()
	<-sender.started

	second, err := service.Deliver(context.Background(), projectID, input)
	require.NoError(t, err)
	assert.Equal(t, 1, second.Duplicates)
	assert.Zero(t, second.Succeeded)
	assert.Zero(t, second.Failed)
	require.Len(t, second.Deliveries, 1)
	assert.Equal(t, "matching delivery is already in progress", second.Deliveries[0].Error)
	assert.Equal(t, 1, sender.sendCount())

	close(sender.release)
	require.NoError(t, <-firstError)
	first := <-firstResult
	require.NotNil(t, first)
	assert.Equal(t, 1, first.Succeeded)
	assert.Equal(t, 1, sender.sendCount())
}
