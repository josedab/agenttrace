package domain

import (
	"time"

	"github.com/google/uuid"
)

// TicketProvider represents a supported external ticket provider
type TicketProvider string

const (
	TicketProviderGitHub TicketProvider = "GITHUB"
	TicketProviderJira   TicketProvider = "JIRA"
	TicketProviderLinear TicketProvider = "LINEAR"
)

// IsValid checks if the ticket provider is valid
func (p TicketProvider) IsValid() bool {
	switch p {
	case TicketProviderGitHub, TicketProviderJira, TicketProviderLinear:
		return true
	}
	return false
}

// TicketIntegration represents a configured ticket provider integration for a project
type TicketIntegration struct {
	ID        uuid.UUID            `json:"id"`
	ProjectID uuid.UUID            `json:"projectId"`
	Provider  TicketProvider       `json:"provider"`
	Config    TicketProviderConfig `json:"config"`
	Enabled   bool                 `json:"enabled"`
	CreatedAt time.Time            `json:"createdAt"`
	UpdatedAt time.Time            `json:"updatedAt"`
}

// TicketProviderConfig holds provider-specific configuration for ticket creation
type TicketProviderConfig struct {
	APIURL          string   `json:"apiUrl"`
	APIToken        string   `json:"apiToken"`
	ProjectKey      string   `json:"projectKey"`
	DefaultLabels   []string `json:"defaultLabels,omitempty"`
	DefaultAssignee string   `json:"defaultAssignee,omitempty"`
}

// TicketTemplate represents a rendered ticket ready for preview or creation
type TicketTemplate struct {
	Provider TicketProvider `json:"provider"`
	Title    string         `json:"title"`
	Body     string         `json:"body"`
	Labels   []string       `json:"labels,omitempty"`
	Priority string         `json:"priority,omitempty"`
}

// TicketCreateInput represents input for creating a ticket from a trace
type TicketCreateInput struct {
	TraceID     string         `json:"traceId"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Provider    TicketProvider `json:"provider"`
	Priority    string         `json:"priority"` // critical, high, medium, low
	Labels      []string       `json:"labels,omitempty"`
}

// TicketResult represents the result of a successfully created ticket
type TicketResult struct {
	ID        string         `json:"id"`
	URL       string         `json:"url"`
	Provider  TicketProvider `json:"provider"`
	CreatedAt time.Time      `json:"createdAt"`
}
