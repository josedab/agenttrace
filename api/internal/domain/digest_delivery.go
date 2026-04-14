package domain

import "github.com/google/uuid"

// TeamDigestDeliveryInput selects configured project webhooks.
type TeamDigestDeliveryInput struct {
	WebhookIDs []uuid.UUID `json:"webhookIds"`
	Window     string      `json:"window,omitempty"`
}

// TeamDigestDeliveryResult reports per-channel delivery outcomes.
// A request that reached the sending stage always returns this honest summary,
// including mixed success and failure, instead of a single aggregate error.
type TeamDigestDeliveryResult struct {
	Deliveries []WebhookDelivery `json:"deliveries"`
	Succeeded  int               `json:"succeeded"`
	Failed     int               `json:"failed"`
	// Duplicates counts webhooks that already received this digest or have the
	// same atomic delivery claim in progress.
	Duplicates int `json:"duplicates"`
	// TriggerUpdateFailures counts successful deliveries whose bookkeeping
	// update failed; the digest was still delivered.
	TriggerUpdateFailures int `json:"triggerUpdateFailures,omitempty"`
	// DeliveryKey identifies the digest payload used for duplicate suppression.
	DeliveryKey string `json:"deliveryKey"`
}
