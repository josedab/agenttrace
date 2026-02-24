package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

const (
	TypeWebhookDelivery = "webhook:delivery"
	MaxWebhookRetries   = 3
)

// WebhookDeliveryPayload is the payload for webhook delivery tasks
type WebhookDeliveryPayload struct {
	RuleID       uuid.UUID         `json:"ruleId"`
	ProjectID    uuid.UUID         `json:"projectId"`
	Action       string            `json:"action"`
	ActionConfig map[string]string `json:"actionConfig"`
	EventData    map[string]any    `json:"eventData"`
}

// WebhookDeliveryWorker handles async webhook deliveries
type WebhookDeliveryWorker struct {
	logger     *zap.Logger
	httpClient *http.Client
}

// NewWebhookDeliveryWorker creates a new webhook delivery worker
func NewWebhookDeliveryWorker(logger *zap.Logger) *WebhookDeliveryWorker {
	return &WebhookDeliveryWorker{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessTask handles webhook delivery tasks
func (w *WebhookDeliveryWorker) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload WebhookDeliveryPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal webhook payload: %w", err)
	}

	w.logger.Info("processing webhook delivery",
		zap.String("ruleId", payload.RuleID.String()),
		zap.String("action", payload.Action),
	)

	var err error
	switch domain.WebhookRuleAction(payload.Action) {
	case domain.ActionSlack:
		err = w.deliverSlack(ctx, payload)
	case domain.ActionPagerDuty:
		err = w.deliverPagerDuty(ctx, payload)
	case domain.ActionCustom:
		err = w.deliverCustomWebhook(ctx, payload)
	default:
		err = w.deliverCustomWebhook(ctx, payload)
	}

	if err != nil {
		w.logger.Error("webhook delivery failed",
			zap.String("ruleId", payload.RuleID.String()),
			zap.Error(err),
		)
		return err
	}

	w.logger.Info("webhook delivered successfully",
		zap.String("ruleId", payload.RuleID.String()),
	)
	return nil
}

func (w *WebhookDeliveryWorker) deliverSlack(ctx context.Context, payload WebhookDeliveryPayload) error {
	webhookURL := payload.ActionConfig["webhookUrl"]
	if webhookURL == "" {
		return fmt.Errorf("slack webhookUrl not configured")
	}

	slackPayload := map[string]any{
		"text": fmt.Sprintf("🔔 AgentTrace Alert: %s", payload.Action),
		"blocks": []map[string]any{
			{
				"type": "section",
				"text": map[string]string{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*AgentTrace Alert*\n*Rule:* %s\n*Event:* %v",
						payload.RuleID.String(), payload.EventData),
				},
			},
		},
	}

	return w.postJSON(ctx, webhookURL, slackPayload)
}

func (w *WebhookDeliveryWorker) deliverPagerDuty(ctx context.Context, payload WebhookDeliveryPayload) error {
	routingKey := payload.ActionConfig["routingKey"]
	if routingKey == "" {
		return fmt.Errorf("pagerduty routingKey not configured")
	}

	pdPayload := map[string]any{
		"routing_key":  routingKey,
		"event_action": "trigger",
		"payload": map[string]any{
			"summary":        fmt.Sprintf("AgentTrace: %s triggered", payload.Action),
			"source":         "agenttrace",
			"severity":       "warning",
			"timestamp":      time.Now().Format(time.RFC3339),
			"custom_details": payload.EventData,
		},
	}

	return w.postJSON(ctx, "https://events.pagerduty.com/v2/enqueue", pdPayload)
}

func (w *WebhookDeliveryWorker) deliverCustomWebhook(ctx context.Context, payload WebhookDeliveryPayload) error {
	url := payload.ActionConfig["url"]
	if url == "" {
		url = payload.ActionConfig["webhookUrl"]
	}
	if url == "" {
		return fmt.Errorf("webhook URL not configured")
	}

	webhookPayload := map[string]any{
		"event":     payload.Action,
		"ruleId":    payload.RuleID.String(),
		"projectId": payload.ProjectID.String(),
		"data":      payload.EventData,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return w.postJSON(ctx, url, webhookPayload)
}

func (w *WebhookDeliveryWorker) postJSON(ctx context.Context, url string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AgentTrace-Webhook/1.0")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// NewWebhookDeliveryTask creates a new webhook delivery task
func NewWebhookDeliveryTask(payload WebhookDeliveryPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeWebhookDelivery, data,
		asynq.MaxRetry(MaxWebhookRetries),
		asynq.Queue("notifications"),
		asynq.Timeout(60*time.Second),
	), nil
}
