package domain

import (
	"time"

	"github.com/google/uuid"
)

// WebhookType represents the type of webhook integration
type WebhookType string

const (
	WebhookTypeSlack     WebhookType = "slack"
	WebhookTypeDiscord   WebhookType = "discord"
	WebhookTypeMSTeams   WebhookType = "msteams"
	WebhookTypePagerDuty WebhookType = "pagerduty"
	WebhookTypeGeneric   WebhookType = "generic"
)

// EventType represents the type of event that can trigger notifications
type EventType string

const (
	EventTypeTraceError       EventType = "trace.error"
	EventTypeTraceCostThreshold EventType = "trace.cost_threshold"
	EventTypeTraceLatencyThreshold EventType = "trace.latency_threshold"
	EventTypeDailyCostReport  EventType = "daily.cost_report"
	EventTypeEvalFailed       EventType = "eval.failed"
	EventTypeEvalScoreLow     EventType = "eval.score_low"
	EventTypeAnomalyDetected  EventType = "anomaly.detected"
)

// Webhook represents a notification webhook configuration
type Webhook struct {
	ID           uuid.UUID    `json:"id"`
	ProjectID    uuid.UUID    `json:"projectId"`
	Type         WebhookType  `json:"type"`
	Name         string       `json:"name"`
	URL          string       `json:"url"`
	Secret       string       `json:"secret,omitempty"` // For signature verification
	Events       []EventType  `json:"events"`
	IsEnabled    bool         `json:"isEnabled"`
	Headers      map[string]string `json:"headers,omitempty"`

	// Thresholds for threshold-based events
	CostThreshold    *float64 `json:"costThreshold,omitempty"`    // USD per trace
	LatencyThreshold *int64   `json:"latencyThreshold,omitempty"` // milliseconds
	ScoreThreshold   *float64 `json:"scoreThreshold,omitempty"`   // 0-1 range

	// Rate limiting
	RateLimitPerHour *int `json:"rateLimitPerHour,omitempty"`

	// Audit fields
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	LastTriggeredAt *time.Time `json:"lastTriggeredAt,omitempty"`
	SuccessCount int64      `json:"successCount"`
	FailureCount int64      `json:"failureCount"`
}

// WebhookInput represents input for creating a webhook
type WebhookInput struct {
	Type             WebhookType       `json:"type" validate:"required,oneof=slack discord msteams pagerduty generic"`
	Name             string            `json:"name" validate:"required,min=1,max=100"`
	URL              string            `json:"url" validate:"required,url"`
	Secret           string            `json:"secret,omitempty"`
	Events           []EventType       `json:"events" validate:"required,min=1"`
	IsEnabled        bool              `json:"isEnabled"`
	Headers          map[string]string `json:"headers,omitempty"`
	CostThreshold    *float64          `json:"costThreshold,omitempty"`
	LatencyThreshold *int64            `json:"latencyThreshold,omitempty"`
	ScoreThreshold   *float64          `json:"scoreThreshold,omitempty"`
	RateLimitPerHour *int              `json:"rateLimitPerHour,omitempty"`
}

// WebhookUpdateInput represents input for updating a webhook
type WebhookUpdateInput struct {
	Name             *string           `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	URL              *string           `json:"url,omitempty" validate:"omitempty,url"`
	Secret           *string           `json:"secret,omitempty"`
	Events           []EventType       `json:"events,omitempty"`
	IsEnabled        *bool             `json:"isEnabled,omitempty"`
	Headers          map[string]string `json:"headers,omitempty"`
	CostThreshold    *float64          `json:"costThreshold,omitempty"`
	LatencyThreshold *int64            `json:"latencyThreshold,omitempty"`
	ScoreThreshold   *float64          `json:"scoreThreshold,omitempty"`
	RateLimitPerHour *int              `json:"rateLimitPerHour,omitempty"`
}

// NotificationPayload represents the data sent to a webhook
type NotificationPayload struct {
	ID        string         `json:"id"`
	EventType EventType      `json:"eventType"`
	Timestamp time.Time      `json:"timestamp"`
	ProjectID string         `json:"projectId"`
	Data      map[string]any `json:"data"`
}

// SlackMessage represents a Slack message payload
type SlackMessage struct {
	Text        string            `json:"text,omitempty"`
	Blocks      []SlackBlock      `json:"blocks,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackBlock represents a Slack Block Kit block
type SlackBlock struct {
	Type     string     `json:"type"`
	Text     *SlackText `json:"text,omitempty"`
	Fields   []SlackText `json:"fields,omitempty"`
	Accessory any       `json:"accessory,omitempty"`
}

// SlackText represents text content in a Slack block
type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// SlackAttachment represents a Slack attachment
type SlackAttachment struct {
	Color     string       `json:"color,omitempty"`
	Title     string       `json:"title,omitempty"`
	TitleLink string       `json:"title_link,omitempty"`
	Text      string       `json:"text,omitempty"`
	Fields    []SlackField `json:"fields,omitempty"`
	Footer    string       `json:"footer,omitempty"`
	Timestamp int64        `json:"ts,omitempty"`
}

// SlackField represents a field in a Slack attachment
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// DiscordMessage represents a Discord webhook message payload
type DiscordMessage struct {
	Content   string          `json:"content,omitempty"`
	Username  string          `json:"username,omitempty"`
	AvatarURL string          `json:"avatar_url,omitempty"`
	Embeds    []DiscordEmbed  `json:"embeds,omitempty"`
}

// DiscordEmbed represents a Discord embed
type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	URL         string              `json:"url,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
	Footer      *DiscordEmbedFooter `json:"footer,omitempty"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
}

// DiscordEmbedFooter represents the footer of a Discord embed
type DiscordEmbedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

// DiscordEmbedField represents a field in a Discord embed
type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// WebhookDelivery represents a delivery attempt for a webhook
type WebhookDelivery struct {
	ID          uuid.UUID `json:"id"`
	WebhookID   uuid.UUID `json:"webhookId"`
	EventType   EventType `json:"eventType"`
	Payload     string    `json:"payload"`
	StatusCode  int       `json:"statusCode"`
	Response    string    `json:"response,omitempty"`
	Success     bool      `json:"success"`
	Error       string    `json:"error,omitempty"`
	Duration    int64     `json:"duration"` // milliseconds
	RetryCount  int       `json:"retryCount"`
	CreatedAt   time.Time `json:"createdAt"`
}

// WebhookFilter represents filter options for querying webhooks
type WebhookFilter struct {
	ProjectID uuid.UUID
	Type      *WebhookType
	IsEnabled *bool
	EventType *EventType
}

// WebhookList represents a paginated list of webhooks
type WebhookList struct {
	Webhooks   []Webhook `json:"webhooks"`
	TotalCount int64     `json:"totalCount"`
	HasMore    bool      `json:"hasMore"`
}

// WebhookDeliveryFilter represents filter options for querying deliveries
type WebhookDeliveryFilter struct {
	WebhookID uuid.UUID
	EventType *EventType
	Success   *bool
}

// WebhookDeliveryList represents a paginated list of deliveries
type WebhookDeliveryList struct {
	Deliveries []WebhookDelivery `json:"deliveries"`
	TotalCount int64             `json:"totalCount"`
	HasMore    bool              `json:"hasMore"`
}
