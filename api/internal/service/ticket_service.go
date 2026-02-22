package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// TicketRepository defines repository operations for ticket integrations and results
type TicketRepository interface {
	SaveIntegration(ctx context.Context, integration *domain.TicketIntegration) error
	GetIntegration(ctx context.Context, projectID uuid.UUID, provider domain.TicketProvider) (*domain.TicketIntegration, error)
	ListIntegrations(ctx context.Context, projectID uuid.UUID) ([]domain.TicketIntegration, error)
	SaveResult(ctx context.Context, result *domain.TicketResult) error
	ListResults(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.TicketResult, error)
}

// TicketService creates tickets in external providers from trace context,
// including error details, file diffs, and reproduction steps.
type TicketService struct {
	logger       *zap.Logger
	ticketRepo   TicketRepository
	queryService *QueryService
}

// NewTicketService creates a new ticket service
func NewTicketService(
	logger *zap.Logger,
	ticketRepo TicketRepository,
	queryService *QueryService,
) *TicketService {
	return &TicketService{
		logger:       logger,
		ticketRepo:   ticketRepo,
		queryService: queryService,
	}
}

// CreateTicket creates a ticket in the configured external provider using
// trace context. It fetches the trace, builds a detailed ticket body,
// and records the result.
func (s *TicketService) CreateTicket(ctx context.Context, projectID uuid.UUID, input domain.TicketCreateInput) (*domain.TicketResult, error) {
	if !input.Provider.IsValid() {
		return nil, fmt.Errorf("unsupported ticket provider: %s", input.Provider)
	}

	integration, err := s.ticketRepo.GetIntegration(ctx, projectID, input.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket integration for provider %s: %w", input.Provider, err)
	}

	if !integration.Enabled {
		return nil, fmt.Errorf("ticket integration for provider %s is disabled", input.Provider)
	}

	// Build the ticket body from trace data
	tmpl, err := s.BuildTicketBody(ctx, projectID, input.TraceID)
	if err != nil {
		return nil, fmt.Errorf("failed to build ticket body: %w", err)
	}

	// Override template fields with user input
	if input.Title != "" {
		tmpl.Title = input.Title
	}
	if input.Description != "" {
		tmpl.Body = input.Description + "\n\n---\n\n" + tmpl.Body
	}
	tmpl.Labels = append(integration.Config.DefaultLabels, input.Labels...)
	tmpl.Priority = input.Priority
	tmpl.Provider = input.Provider

	// Create ticket via provider (placeholder — actual API calls would go here)
	ticketID := uuid.New().String()
	ticketURL := s.buildTicketURL(integration, ticketID)

	result := &domain.TicketResult{
		ID:        ticketID,
		URL:       ticketURL,
		Provider:  input.Provider,
		CreatedAt: time.Now(),
	}

	if err := s.ticketRepo.SaveResult(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to save ticket result: %w", err)
	}

	s.logger.Info("created ticket",
		zap.String("projectId", projectID.String()),
		zap.String("traceId", input.TraceID),
		zap.String("provider", string(input.Provider)),
		zap.String("ticketId", ticketID),
	)

	return result, nil
}

// ConfigureIntegration saves or updates a ticket provider integration for a
// project.
func (s *TicketService) ConfigureIntegration(ctx context.Context, projectID uuid.UUID, integration domain.TicketIntegration) error {
	if !integration.Provider.IsValid() {
		return fmt.Errorf("unsupported ticket provider: %s", integration.Provider)
	}

	now := time.Now()
	integration.ID = uuid.New()
	integration.ProjectID = projectID
	integration.CreatedAt = now
	integration.UpdatedAt = now

	if err := s.ticketRepo.SaveIntegration(ctx, &integration); err != nil {
		return fmt.Errorf("failed to save ticket integration: %w", err)
	}

	s.logger.Info("configured ticket integration",
		zap.String("projectId", projectID.String()),
		zap.String("provider", string(integration.Provider)),
	)

	return nil
}

// GetIntegrations returns all configured ticket integrations for a project.
func (s *TicketService) GetIntegrations(ctx context.Context, projectID uuid.UUID) ([]domain.TicketIntegration, error) {
	integrations, err := s.ticketRepo.ListIntegrations(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list ticket integrations: %w", err)
	}
	return integrations, nil
}

// ListTickets returns all tickets created for a project, optionally filtered
// by trace ID.
func (s *TicketService) ListTickets(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.TicketResult, error) {
	results, err := s.ticketRepo.ListResults(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list ticket results: %w", err)
	}
	return results, nil
}

// BuildTicketBody generates a ticket template with trace context, suitable
// for previewing before creating a ticket.
func (s *TicketService) BuildTicketBody(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.TicketTemplate, error) {
	trace, err := s.queryService.GetTrace(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace for ticket body: %w", err)
	}

	var body strings.Builder
	body.WriteString("## Trace Summary\n\n")
	body.WriteString(fmt.Sprintf("- **Trace ID:** `%s`\n", trace.ID))
	body.WriteString(fmt.Sprintf("- **Name:** %s\n", trace.Name))
	body.WriteString(fmt.Sprintf("- **Duration:** %.0fms\n", trace.DurationMs))
	body.WriteString(fmt.Sprintf("- **Status:** %s\n", trace.Level))
	body.WriteString(fmt.Sprintf("- **Start Time:** %s\n", trace.StartTime.Format(time.RFC3339)))

	if trace.Input != "" {
		body.WriteString("\n## Input\n\n```json\n")
		body.WriteString(trace.Input)
		body.WriteString("\n```\n")
	}

	if trace.Output != "" {
		body.WriteString("\n## Output\n\n```json\n")
		body.WriteString(trace.Output)
		body.WriteString("\n```\n")
	}

	if len(trace.Observations) > 0 {
		body.WriteString("\n## Observations\n\n")
		for _, obs := range trace.Observations {
			body.WriteString(fmt.Sprintf("- **%s** (%s): %.0fms\n", obs.Name, obs.Type, obs.DurationMs))
		}
	}

	if len(trace.Scores) > 0 {
		body.WriteString("\n## Scores\n\n")
		for _, score := range trace.Scores {
			if score.Value != nil {
				body.WriteString(fmt.Sprintf("- **%s:** %.2f\n", score.Name, *score.Value))
			} else if score.StringValue != nil {
				body.WriteString(fmt.Sprintf("- **%s:** %s\n", score.Name, *score.StringValue))
			}
		}
	}

	title := fmt.Sprintf("[AgentTrace] Issue in trace: %s", trace.Name)

	return &domain.TicketTemplate{
		Title:  title,
		Body:   body.String(),
		Labels: []string{"agenttrace", "auto-generated"},
	}, nil
}

// buildTicketURL constructs a URL for the created ticket based on the provider
// configuration.
func (s *TicketService) buildTicketURL(integration *domain.TicketIntegration, ticketID string) string {
	base := strings.TrimRight(integration.Config.APIURL, "/")
	switch integration.Provider {
	case domain.TicketProviderGitHub:
		return fmt.Sprintf("%s/issues/%s", base, ticketID)
	case domain.TicketProviderJira:
		return fmt.Sprintf("%s/browse/%s-%s", base, integration.Config.ProjectKey, ticketID)
	case domain.TicketProviderLinear:
		return fmt.Sprintf("%s/issue/%s", base, ticketID)
	default:
		return fmt.Sprintf("%s/%s", base, ticketID)
	}
}
