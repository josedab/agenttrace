package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// digestDuplicateWindow suppresses an immediate repeat of the same digest, which
// is what a double-clicked "send digest" action produces.
const digestDuplicateWindow = 5 * time.Minute

// DigestWebhookRepository defines project-scoped webhook and audit persistence.
type DigestWebhookRepository interface {
	GetByProjectID(
		ctx context.Context,
		projectID, webhookID uuid.UUID,
	) (*domain.Webhook, error)
	CreateDelivery(ctx context.Context, delivery *domain.WebhookDelivery) error
	// ClaimDelivery atomically reserves one logical delivery across API
	// instances. A false result means another request owns the active claim.
	ClaimDelivery(
		ctx context.Context,
		webhookID uuid.UUID,
		deliveryKey string,
		claimedAt, expiresAt time.Time,
	) (bool, error)
	// FindRecentDelivery returns a delivery already recorded for the same
	// delivery key so an immediate retry is not sent twice.
	FindRecentDelivery(
		ctx context.Context,
		webhookID uuid.UUID,
		deliveryKey string,
		since time.Time,
	) (*domain.WebhookDelivery, error)
	UpdateLastTriggered(ctx context.Context, id uuid.UUID) error
}

// DigestNotificationSender sends SSRF-safe webhook notifications.
type DigestNotificationSender interface {
	SendNotification(
		ctx context.Context,
		webhook *domain.Webhook,
		eventType domain.EventType,
		data map[string]any,
	) (*domain.WebhookDelivery, error)
}

// OutcomeDigestProvider supplies canonical team digest content.
type OutcomeDigestProvider interface {
	GetDigest(
		ctx context.Context,
		projectID uuid.UUID,
		from, to time.Time,
	) (*domain.OutcomeDigest, error)
}

// TeamDigestDeliveryService delivers one canonical digest across configured channels.
type TeamDigestDeliveryService struct {
	webhooks      DigestWebhookRepository
	notifications DigestNotificationSender
	outcomes      OutcomeDigestProvider
	guard         OutboundGuard
	clock         func() time.Time
	validateURL   func(context.Context, string) error
}

// NewTeamDigestDeliveryService creates a digest delivery service.
// The outbound guard and per-webhook URL validation run before any send, so a
// privacy or SSRF rejection never leaves a partially delivered digest behind.
func NewTeamDigestDeliveryService(
	webhooks DigestWebhookRepository,
	notifications DigestNotificationSender,
	outcomes OutcomeDigestProvider,
	guards ...OutboundGuard,
) *TeamDigestDeliveryService {
	var guard OutboundGuard
	if len(guards) > 0 {
		guard = guards[0]
	}
	return &TeamDigestDeliveryService{
		webhooks:      webhooks,
		notifications: notifications,
		outcomes:      outcomes,
		guard:         guard,
		clock:         time.Now,
		validateURL:   ValidateWebhookURL,
	}
}

// Deliver sends a digest only to explicitly selected, enabled project webhooks.
// Every requested webhook is validated before the first send; once sending has
// started the result is reported per webhook instead of aborting midway.
func (s *TeamDigestDeliveryService) Deliver(
	ctx context.Context,
	projectID uuid.UUID,
	input domain.TeamDigestDeliveryInput,
) (*domain.TeamDigestDeliveryResult, error) {
	if err := RequireOutbound(s.guard, EgressWebhooks); err != nil {
		return nil, err
	}
	if len(input.WebhookIDs) == 0 {
		return nil, apperrors.Validation("at least one webhookId is required")
	}
	if len(input.WebhookIDs) > 20 {
		return nil, apperrors.Validation("no more than 20 webhookIds may be delivered at once")
	}

	from, to, err := outcomeReportWindow(s.clock().UTC(), input.Window)
	if err != nil {
		return nil, err
	}
	webhooks, err := s.validateWebhooks(ctx, projectID, input.WebhookIDs)
	if err != nil {
		return nil, err
	}

	digest, err := s.outcomes.GetDigest(ctx, projectID, from, to)
	if err != nil {
		return nil, err
	}
	markdown := RenderOutcomeDigestMarkdown(digest)
	deliveryKey := teamDigestDeliveryKey(projectID, input.Window, markdown)
	data := map[string]any{
		"title":       digest.Title,
		"summary":     digest.Summary,
		"highlights":  digest.Highlights,
		"attention":   digest.Attention,
		"markdown":    markdown,
		"deliveryKey": deliveryKey,
	}

	result := &domain.TeamDigestDeliveryResult{
		Deliveries:  []domain.WebhookDelivery{},
		DeliveryKey: deliveryKey,
	}
	for _, webhook := range webhooks {
		delivery := s.deliverToWebhook(ctx, webhook, deliveryKey, data, result)
		result.Deliveries = append(result.Deliveries, *delivery)
	}
	return result, nil
}

// validateWebhooks resolves every requested webhook before any side effect.
// A single invalid webhook rejects the whole request, so callers never discover
// a rejected channel after other channels were already notified.
func (s *TeamDigestDeliveryService) validateWebhooks(
	ctx context.Context,
	projectID uuid.UUID,
	webhookIDs []uuid.UUID,
) ([]*domain.Webhook, error) {
	seen := make(map[uuid.UUID]struct{}, len(webhookIDs))
	webhooks := make([]*domain.Webhook, 0, len(webhookIDs))
	for _, webhookID := range webhookIDs {
		if _, duplicate := seen[webhookID]; duplicate {
			continue
		}
		seen[webhookID] = struct{}{}

		webhook, err := s.webhooks.GetByProjectID(ctx, projectID, webhookID)
		if err != nil {
			return nil, err
		}
		if !webhook.IsEnabled {
			return nil, apperrors.Validation(
				fmt.Sprintf("webhook %s is disabled", webhookID),
			)
		}
		switch webhook.Type {
		case domain.WebhookTypeSlack, domain.WebhookTypeDiscord, domain.WebhookTypeGeneric:
		default:
			return nil, apperrors.Validation(
				fmt.Sprintf("webhook %s does not support team digests", webhookID),
			)
		}
		// Resolve every destination before the first delivery. The transport
		// repeats the same check when connecting to protect against DNS rebinding.
		if err := s.validateURL(ctx, webhook.URL); err != nil {
			return nil, apperrors.Validation(
				fmt.Sprintf("webhook %s has an unusable destination: %s", webhookID, err.Error()),
			)
		}
		webhooks = append(webhooks, webhook)
	}
	return webhooks, nil
}

// deliverToWebhook sends one digest and records the outcome. It never returns an
// error: after the first send the caller receives an honest partial result.
func (s *TeamDigestDeliveryService) deliverToWebhook(
	ctx context.Context,
	webhook *domain.Webhook,
	deliveryKey string,
	data map[string]any,
	result *domain.TeamDigestDeliveryResult,
) *domain.WebhookDelivery {
	now := s.clock().UTC()
	previous, lookupErr := s.webhooks.FindRecentDelivery(
		ctx,
		webhook.ID,
		deliveryKey,
		now.Add(-digestDuplicateWindow),
	)
	if lookupErr != nil {
		delivery := &domain.WebhookDelivery{
			ID:          uuid.New(),
			WebhookID:   webhook.ID,
			EventType:   domain.EventTypeTeamDigest,
			Success:     false,
			Error:       "delivery duplicate check failed",
			DeliveryKey: deliveryKey,
			CreatedAt:   now,
		}
		result.Failed++
		if err := s.webhooks.CreateDelivery(ctx, delivery); err != nil {
			delivery.Error = "delivery could not be recorded"
		}
		return delivery
	}
	if previous != nil {
		result.Duplicates++
		if previous.Success {
			result.Succeeded++
		} else {
			result.Failed++
		}
		return previous
	}

	claimed, claimErr := s.webhooks.ClaimDelivery(
		ctx,
		webhook.ID,
		deliveryKey,
		now,
		now.Add(digestDuplicateWindow),
	)
	if claimErr != nil {
		return s.recordDigestSafetyFailure(
			ctx,
			webhook.ID,
			deliveryKey,
			now,
			"delivery duplicate claim failed",
			result,
		)
	}
	if !claimed {
		// The winning request may have completed between our first lookup and
		// the failed claim. Return its record when available; otherwise report
		// the in-flight duplicate without sending a second request.
		previous, lookupErr = s.webhooks.FindRecentDelivery(
			ctx,
			webhook.ID,
			deliveryKey,
			now.Add(-digestDuplicateWindow),
		)
		if lookupErr != nil {
			return s.recordDigestSafetyFailure(
				ctx,
				webhook.ID,
				deliveryKey,
				now,
				"delivery duplicate check failed",
				result,
			)
		}
		result.Duplicates++
		if previous != nil {
			if previous.Success {
				result.Succeeded++
			} else {
				result.Failed++
			}
			return previous
		}
		return &domain.WebhookDelivery{
			ID:          uuid.New(),
			WebhookID:   webhook.ID,
			EventType:   domain.EventTypeTeamDigest,
			Success:     false,
			Error:       "matching delivery is already in progress",
			DeliveryKey: deliveryKey,
			CreatedAt:   now,
		}
	}

	delivery, sendErr := s.notifications.SendNotification(
		ctx,
		webhook,
		domain.EventTypeTeamDigest,
		data,
	)
	if delivery == nil {
		delivery = &domain.WebhookDelivery{
			ID:        uuid.New(),
			WebhookID: webhook.ID,
			EventType: domain.EventTypeTeamDigest,
			CreatedAt: now,
		}
	}
	delivery.DeliveryKey = deliveryKey
	if sendErr != nil {
		delivery.Success = false
		if delivery.Error == "" {
			delivery.Error = safeDigestFailure(sendErr)
		}
	}

	if err := s.webhooks.CreateDelivery(ctx, delivery); err != nil {
		result.Failed++
		delivery.Success = false
		delivery.Error = "delivery could not be recorded"
		return delivery
	}
	if !delivery.Success {
		result.Failed++
		return delivery
	}

	result.Succeeded++
	if err := s.webhooks.UpdateLastTriggered(ctx, webhook.ID); err != nil {
		// The digest was delivered; a bookkeeping failure must not turn a
		// successful send into a reported failure.
		result.TriggerUpdateFailures++
	}
	return delivery
}

func (s *TeamDigestDeliveryService) recordDigestSafetyFailure(
	ctx context.Context,
	webhookID uuid.UUID,
	deliveryKey string,
	now time.Time,
	message string,
	result *domain.TeamDigestDeliveryResult,
) *domain.WebhookDelivery {
	delivery := &domain.WebhookDelivery{
		ID:          uuid.New(),
		WebhookID:   webhookID,
		EventType:   domain.EventTypeTeamDigest,
		Success:     false,
		Error:       message,
		DeliveryKey: deliveryKey,
		CreatedAt:   now,
	}
	result.Failed++
	if err := s.webhooks.CreateDelivery(ctx, delivery); err != nil {
		delivery.Error = "delivery could not be recorded"
	}
	return delivery
}

// teamDigestDeliveryKey identifies one digest for one project window. Exact
// request timestamps are deliberately excluded so a double-click seconds later
// produces the same key when the rendered digest has not changed.
func teamDigestDeliveryKey(projectID uuid.UUID, window, markdown string) string {
	if window == "" {
		window = "7d"
	}
	checksum := sha256.Sum256([]byte(
		projectID.String() + "|" + window + "|" + markdown,
	))
	return hex.EncodeToString(checksum[:])[:32]
}

// safeDigestFailure keeps webhook internals out of stored delivery records.
func safeDigestFailure(err error) string {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		return appErr.Message
	}
	return "delivery failed"
}
