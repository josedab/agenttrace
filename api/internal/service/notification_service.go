package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/circuitbreaker"
)

// NotificationService handles sending notifications via webhooks
type NotificationService struct {
	logger           *zap.Logger
	httpClient       *http.Client
	dashboardURL     string
	cbRegistry       *circuitbreaker.Registry
}

// NewNotificationService creates a new notification service
func NewNotificationService(logger *zap.Logger, dashboardURL string) *NotificationService {
	registry := circuitbreaker.NewRegistry()

	return &NotificationService{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		dashboardURL: dashboardURL,
		cbRegistry:   registry,
	}
}

// getCircuitBreakerForHost returns a circuit breaker for the given webhook URL's host
func (s *NotificationService) getCircuitBreakerForHost(webhookURL string) *circuitbreaker.CircuitBreaker {
	parsedURL, err := url.Parse(webhookURL)
	if err != nil {
		// Fallback to a default circuit breaker if URL parsing fails
		return s.cbRegistry.Get("webhook:default", s.webhookCircuitBreakerConfig("default"))
	}

	host := parsedURL.Host
	return s.cbRegistry.Get("webhook:"+host, s.webhookCircuitBreakerConfig(host))
}

// webhookCircuitBreakerConfig returns circuit breaker configuration for webhooks
func (s *NotificationService) webhookCircuitBreakerConfig(name string) circuitbreaker.Config {
	return circuitbreaker.Config{
		Name:                "webhook:" + name,
		MaxFailures:         3,                // Open after 3 consecutive failures
		Timeout:             60 * time.Second, // Try again after 1 minute
		MaxHalfOpenRequests: 1,
		OnStateChange: func(cbName string, from, to circuitbreaker.State) {
			s.logger.Info("webhook circuit breaker state changed",
				zap.String("circuit_breaker", cbName),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	}
}

// SendNotification sends a notification to a webhook
func (s *NotificationService) SendNotification(
	ctx context.Context,
	webhook *domain.Webhook,
	eventType domain.EventType,
	data map[string]any,
) (*domain.WebhookDelivery, error) {
	delivery := &domain.WebhookDelivery{
		ID:         uuid.New(),
		WebhookID:  webhook.ID,
		EventType:  eventType,
		CreatedAt:  time.Now(),
		RetryCount: 0,
	}

	start := time.Now()

	// Build the appropriate message format based on webhook type
	var payload []byte
	var err error

	switch webhook.Type {
	case domain.WebhookTypeSlack:
		payload, err = s.buildSlackPayload(eventType, data)
	case domain.WebhookTypeDiscord:
		payload, err = s.buildDiscordPayload(eventType, data)
	case domain.WebhookTypeMSTeams:
		payload, err = s.buildMSTeamsPayload(eventType, data)
	case domain.WebhookTypePagerDuty:
		payload, err = s.buildPagerDutyPayload(eventType, data)
	default:
		payload, err = s.buildGenericPayload(eventType, data)
	}

	if err != nil {
		delivery.Success = false
		delivery.Error = fmt.Sprintf("failed to build payload: %v", err)
		return delivery, err
	}

	delivery.Payload = string(payload)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewReader(payload))
	if err != nil {
		delivery.Success = false
		delivery.Error = fmt.Sprintf("failed to create request: %v", err)
		return delivery, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AgentTrace-Webhook/1.0")

	// Add custom headers
	for key, value := range webhook.Headers {
		req.Header.Set(key, value)
	}

	// Add signature if secret is configured
	if webhook.Secret != "" {
		signature := s.computeSignature(payload, webhook.Secret)
		req.Header.Set("X-AgentTrace-Signature", signature)
	}

	// Get circuit breaker for this webhook's host
	cb := s.getCircuitBreakerForHost(webhook.URL)

	// Send the request with circuit breaker protection
	err = cb.Execute(ctx, func() error {
		resp, httpErr := s.httpClient.Do(req)
		if httpErr != nil {
			delivery.Success = false
			delivery.Error = fmt.Sprintf("request failed: %v", httpErr)
			delivery.Duration = time.Since(start).Milliseconds()
			return httpErr
		}
		defer resp.Body.Close()

		delivery.Duration = time.Since(start).Milliseconds()
		delivery.StatusCode = resp.StatusCode

		// Read response body
		body, _ := io.ReadAll(resp.Body)
		delivery.Response = string(body)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			delivery.Success = true
			return nil
		}

		delivery.Success = false
		delivery.Error = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	})

	// Check if the error was due to circuit breaker being open
	if err == circuitbreaker.ErrCircuitOpen {
		delivery.Success = false
		delivery.Error = "circuit breaker open: webhook endpoint temporarily unavailable"
		s.logger.Warn("webhook call blocked by circuit breaker",
			zap.String("webhook_url", webhook.URL),
			zap.String("webhook_id", webhook.ID.String()),
		)
	} else if err == circuitbreaker.ErrTooManyRequests {
		delivery.Success = false
		delivery.Error = "circuit breaker half-open: too many concurrent requests"
	}

	return delivery, err
}

// computeSignature computes HMAC-SHA256 signature
func (s *NotificationService) computeSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// buildSlackPayload builds a Slack message payload
func (s *NotificationService) buildSlackPayload(eventType domain.EventType, data map[string]any) ([]byte, error) {
	var color, title, text string

	switch eventType {
	case domain.EventTypeTraceError:
		color = "#dc3545" // red
		title = "Trace Error Alert"
		text = s.formatTraceErrorMessage(data)
	case domain.EventTypeTraceCostThreshold:
		color = "#ffc107" // yellow
		title = "Cost Threshold Exceeded"
		text = s.formatCostThresholdMessage(data)
	case domain.EventTypeTraceLatencyThreshold:
		color = "#ffc107" // yellow
		title = "Latency Threshold Exceeded"
		text = s.formatLatencyThresholdMessage(data)
	case domain.EventTypeDailyCostReport:
		color = "#17a2b8" // blue
		title = "Daily Cost Report"
		text = s.formatDailyCostMessage(data)
	case domain.EventTypeEvalFailed:
		color = "#dc3545" // red
		title = "Evaluation Failed"
		text = s.formatEvalFailedMessage(data)
	case domain.EventTypeEvalScoreLow:
		color = "#ffc107" // yellow
		title = "Low Evaluation Score"
		text = s.formatLowScoreMessage(data)
	case domain.EventTypeAnomalyDetected:
		color = "#6f42c1" // purple
		title = "Anomaly Detected"
		text = s.formatAnomalyMessage(data)
	default:
		color = "#6c757d" // gray
		title = "AgentTrace Notification"
		text = fmt.Sprintf("Event: %s", eventType)
	}

	// Build Slack attachment format
	msg := domain.SlackMessage{
		Attachments: []domain.SlackAttachment{
			{
				Color:  color,
				Title:  title,
				Text:   text,
				Footer: "AgentTrace",
				Timestamp: time.Now().Unix(),
			},
		},
	}

	// Add fields if available
	if traceID, ok := data["traceId"].(string); ok {
		msg.Attachments[0].TitleLink = fmt.Sprintf("%s/traces/%s", s.dashboardURL, traceID)
		msg.Attachments[0].Fields = append(msg.Attachments[0].Fields, domain.SlackField{
			Title: "Trace ID",
			Value: traceID,
			Short: true,
		})
	}

	if projectName, ok := data["projectName"].(string); ok {
		msg.Attachments[0].Fields = append(msg.Attachments[0].Fields, domain.SlackField{
			Title: "Project",
			Value: projectName,
			Short: true,
		})
	}

	return json.Marshal(msg)
}

// buildDiscordPayload builds a Discord webhook payload
func (s *NotificationService) buildDiscordPayload(eventType domain.EventType, data map[string]any) ([]byte, error) {
	var color int
	var title, description string

	switch eventType {
	case domain.EventTypeTraceError:
		color = 0xdc3545 // red
		title = "Trace Error Alert"
		description = s.formatTraceErrorMessage(data)
	case domain.EventTypeTraceCostThreshold:
		color = 0xffc107 // yellow
		title = "Cost Threshold Exceeded"
		description = s.formatCostThresholdMessage(data)
	case domain.EventTypeTraceLatencyThreshold:
		color = 0xffc107 // yellow
		title = "Latency Threshold Exceeded"
		description = s.formatLatencyThresholdMessage(data)
	case domain.EventTypeDailyCostReport:
		color = 0x17a2b8 // blue
		title = "Daily Cost Report"
		description = s.formatDailyCostMessage(data)
	case domain.EventTypeEvalFailed:
		color = 0xdc3545 // red
		title = "Evaluation Failed"
		description = s.formatEvalFailedMessage(data)
	case domain.EventTypeEvalScoreLow:
		color = 0xffc107 // yellow
		title = "Low Evaluation Score"
		description = s.formatLowScoreMessage(data)
	case domain.EventTypeAnomalyDetected:
		color = 0x6f42c1 // purple
		title = "Anomaly Detected"
		description = s.formatAnomalyMessage(data)
	default:
		color = 0x6c757d // gray
		title = "AgentTrace Notification"
		description = fmt.Sprintf("Event: %s", eventType)
	}

	embed := domain.DiscordEmbed{
		Title:       title,
		Description: description,
		Color:       color,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &domain.DiscordEmbedFooter{
			Text: "AgentTrace",
		},
	}

	// Add fields
	if traceID, ok := data["traceId"].(string); ok {
		embed.URL = fmt.Sprintf("%s/traces/%s", s.dashboardURL, traceID)
		embed.Fields = append(embed.Fields, domain.DiscordEmbedField{
			Name:   "Trace ID",
			Value:  traceID,
			Inline: true,
		})
	}

	if projectName, ok := data["projectName"].(string); ok {
		embed.Fields = append(embed.Fields, domain.DiscordEmbedField{
			Name:   "Project",
			Value:  projectName,
			Inline: true,
		})
	}

	msg := domain.DiscordMessage{
		Username:  "AgentTrace",
		Embeds:    []domain.DiscordEmbed{embed},
	}

	return json.Marshal(msg)
}

// buildMSTeamsPayload builds a Microsoft Teams payload
func (s *NotificationService) buildMSTeamsPayload(eventType domain.EventType, data map[string]any) ([]byte, error) {
	var color, title, text string

	switch eventType {
	case domain.EventTypeTraceError:
		color = "dc3545"
		title = "Trace Error Alert"
		text = s.formatTraceErrorMessage(data)
	case domain.EventTypeTraceCostThreshold:
		color = "ffc107"
		title = "Cost Threshold Exceeded"
		text = s.formatCostThresholdMessage(data)
	default:
		color = "6c757d"
		title = "AgentTrace Notification"
		text = fmt.Sprintf("Event: %s", eventType)
	}

	// Microsoft Teams Adaptive Card format
	card := map[string]any{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"themeColor": color,
		"summary":    title,
		"sections": []map[string]any{
			{
				"activityTitle": title,
				"text":          text,
			},
		},
	}

	// Add link if trace ID available
	if traceID, ok := data["traceId"].(string); ok {
		card["potentialAction"] = []map[string]any{
			{
				"@type": "OpenUri",
				"name":  "View Trace",
				"targets": []map[string]string{
					{"os": "default", "uri": fmt.Sprintf("%s/traces/%s", s.dashboardURL, traceID)},
				},
			},
		}
	}

	return json.Marshal(card)
}

// buildPagerDutyPayload builds a PagerDuty Events API v2 payload
func (s *NotificationService) buildPagerDutyPayload(eventType domain.EventType, data map[string]any) ([]byte, error) {
	severity := "warning"
	summary := string(eventType)

	switch eventType {
	case domain.EventTypeTraceError, domain.EventTypeEvalFailed:
		severity = "error"
	case domain.EventTypeAnomalyDetected:
		severity = "critical"
	}

	if traceName, ok := data["traceName"].(string); ok {
		summary = fmt.Sprintf("%s: %s", eventType, traceName)
	}

	payload := map[string]any{
		"routing_key":  "", // Will be set from webhook URL or config
		"event_action": "trigger",
		"payload": map[string]any{
			"summary":  summary,
			"severity": severity,
			"source":   "agenttrace",
			"custom_details": data,
		},
	}

	if traceID, ok := data["traceId"].(string); ok {
		payload["dedup_key"] = traceID
		payload["links"] = []map[string]string{
			{"href": fmt.Sprintf("%s/traces/%s", s.dashboardURL, traceID), "text": "View in AgentTrace"},
		}
	}

	return json.Marshal(payload)
}

// buildGenericPayload builds a generic JSON payload
func (s *NotificationService) buildGenericPayload(eventType domain.EventType, data map[string]any) ([]byte, error) {
	payload := domain.NotificationPayload{
		ID:        uuid.New().String(),
		EventType: eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	if projectID, ok := data["projectId"].(string); ok {
		payload.ProjectID = projectID
	}

	return json.Marshal(payload)
}

// Message formatting helpers

func (s *NotificationService) formatTraceErrorMessage(data map[string]any) string {
	traceName := getString(data, "traceName", "Unknown")
	errorMsg := getString(data, "error", "An error occurred")
	return fmt.Sprintf("Trace '%s' failed with error:\n```%s```", traceName, errorMsg)
}

func (s *NotificationService) formatCostThresholdMessage(data map[string]any) string {
	traceName := getString(data, "traceName", "Unknown")
	cost := getFloat(data, "cost", 0)
	threshold := getFloat(data, "threshold", 0)
	return fmt.Sprintf("Trace '%s' cost $%.4f exceeded threshold of $%.4f", traceName, cost, threshold)
}

func (s *NotificationService) formatLatencyThresholdMessage(data map[string]any) string {
	traceName := getString(data, "traceName", "Unknown")
	latency := getFloat(data, "latencyMs", 0)
	threshold := getFloat(data, "threshold", 0)
	return fmt.Sprintf("Trace '%s' latency %.0fms exceeded threshold of %.0fms", traceName, latency, threshold)
}

func (s *NotificationService) formatDailyCostMessage(data map[string]any) string {
	totalCost := getFloat(data, "totalCost", 0)
	traceCount := getInt(data, "traceCount", 0)
	date := getString(data, "date", time.Now().Format("2006-01-02"))
	return fmt.Sprintf("Daily summary for %s:\n- Total Cost: $%.2f\n- Total Traces: %d", date, totalCost, traceCount)
}

func (s *NotificationService) formatEvalFailedMessage(data map[string]any) string {
	evalName := getString(data, "evaluatorName", "Unknown")
	traceName := getString(data, "traceName", "Unknown")
	errorMsg := getString(data, "error", "Evaluation failed")
	return fmt.Sprintf("Evaluator '%s' failed on trace '%s':\n```%s```", evalName, traceName, errorMsg)
}

func (s *NotificationService) formatLowScoreMessage(data map[string]any) string {
	scoreName := getString(data, "scoreName", "Unknown")
	score := getFloat(data, "score", 0)
	threshold := getFloat(data, "threshold", 0)
	traceName := getString(data, "traceName", "Unknown")
	return fmt.Sprintf("Score '%s' = %.2f (below threshold %.2f) on trace '%s'", scoreName, score, threshold, traceName)
}

func (s *NotificationService) formatAnomalyMessage(data map[string]any) string {
	anomalyType := getString(data, "anomalyType", "Unknown")
	description := getString(data, "description", "Anomaly detected")
	return fmt.Sprintf("Anomaly detected: %s\n%s", anomalyType, description)
}

// Helper functions

func getString(data map[string]any, key string, defaultVal string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return defaultVal
}

func getFloat(data map[string]any, key string, defaultVal float64) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}
	if val, ok := data[key].(int); ok {
		return float64(val)
	}
	return defaultVal
}

func getInt(data map[string]any, key string, defaultVal int) int {
	if val, ok := data[key].(int); ok {
		return val
	}
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	return defaultVal
}

// TestWebhook sends a test notification to verify webhook configuration
func (s *NotificationService) TestWebhook(ctx context.Context, webhook *domain.Webhook) (*domain.WebhookDelivery, error) {
	testData := map[string]any{
		"message":     "This is a test notification from AgentTrace",
		"projectId":   webhook.ProjectID.String(),
		"webhookId":   webhook.ID.String(),
		"webhookName": webhook.Name,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	return s.SendNotification(ctx, webhook, domain.EventType("test"), testData)
}
