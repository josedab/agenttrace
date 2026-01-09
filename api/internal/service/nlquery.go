package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// NLQueryService handles natural language to structured query conversion
type NLQueryService struct {
	config       *config.Config
	queryService *QueryService
	logger       *zap.Logger
	httpClient   *http.Client
}

// NewNLQueryService creates a new natural language query service
func NewNLQueryService(
	cfg *config.Config,
	queryService *QueryService,
	logger *zap.Logger,
) *NLQueryService {
	return &NLQueryService{
		config:       cfg,
		queryService: queryService,
		logger:       logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NLQueryRequest represents a natural language query request
type NLQueryRequest struct {
	Query string `json:"query"`
}

// NLQueryResponse represents the response from a natural language query
type NLQueryResponse struct {
	Query           string            `json:"query"`
	InterpretedAs   string            `json:"interpretedAs"`
	Filter          *domain.TraceFilter `json:"filter"`
	Traces          *domain.TraceList `json:"traces,omitempty"`
	Suggestions     []string          `json:"suggestions,omitempty"`
	ExecutionTimeMs int64             `json:"executionTimeMs"`
}

// ParsedQuery represents the LLM's interpretation of a natural language query
type ParsedQuery struct {
	Interpretation string             `json:"interpretation"`
	Filter         ParsedFilter       `json:"filter"`
	Suggestions    []string           `json:"suggestions"`
}

// ParsedFilter represents filter parameters extracted by the LLM
type ParsedFilter struct {
	Name        *string `json:"name,omitempty"`
	UserID      *string `json:"userId,omitempty"`
	SessionID   *string `json:"sessionId,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Level       *string `json:"level,omitempty"`
	HasError    *bool   `json:"hasError,omitempty"`
	MinCost     *float64 `json:"minCost,omitempty"`
	MaxCost     *float64 `json:"maxCost,omitempty"`
	MinDuration *float64 `json:"minDurationMs,omitempty"`
	MaxDuration *float64 `json:"maxDurationMs,omitempty"`
	FromTime    *string `json:"fromTime,omitempty"`
	ToTime      *string `json:"toTime,omitempty"`
	Search      *string `json:"search,omitempty"`
	GitBranch   *string `json:"gitBranch,omitempty"`
	GitCommit   *string `json:"gitCommitSha,omitempty"`
}

// QueryTraces processes a natural language query and returns matching traces
func (s *NLQueryService) QueryTraces(
	ctx context.Context,
	projectID uuid.UUID,
	query string,
	limit int,
) (*NLQueryResponse, error) {
	startTime := time.Now()

	// Parse the natural language query using LLM
	parsed, err := s.parseQuery(ctx, query)
	if err != nil {
		s.logger.Warn("failed to parse NL query, using fallback search",
			zap.String("query", query),
			zap.Error(err),
		)
		// Fallback to basic search
		parsed = &ParsedQuery{
			Interpretation: "Text search for: " + query,
			Filter: ParsedFilter{
				Search: &query,
			},
		}
	}

	// Convert parsed filter to domain filter
	filter := s.convertFilter(parsed.Filter, projectID)

	// Execute the query
	traces, err := s.queryService.ListTraces(ctx, filter, limit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &NLQueryResponse{
		Query:           query,
		InterpretedAs:   parsed.Interpretation,
		Filter:          filter,
		Traces:          traces,
		Suggestions:     parsed.Suggestions,
		ExecutionTimeMs: time.Since(startTime).Milliseconds(),
	}, nil
}

// parseQuery uses an LLM to convert natural language to structured query
func (s *NLQueryService) parseQuery(ctx context.Context, query string) (*ParsedQuery, error) {
	if s.config.Eval.APIKey == "" {
		return nil, fmt.Errorf("LLM API key not configured")
	}

	systemPrompt := `You are a query parser for an AI agent observability platform called AgentTrace.
Your job is to convert natural language queries about traces into structured filters.

Available filter fields:
- name: Trace name (string, partial match)
- userId: User ID (exact match)
- sessionId: Session ID (exact match)
- tags: List of tags (array)
- level: Log level (DEBUG, INFO, WARNING, ERROR)
- hasError: Whether trace has errors (boolean)
- minCost: Minimum total cost in USD (float)
- maxCost: Maximum total cost in USD (float)
- minDurationMs: Minimum duration in milliseconds (float)
- maxDurationMs: Maximum duration in milliseconds (float)
- fromTime: Start of time range (ISO 8601 format)
- toTime: End of time range (ISO 8601 format)
- search: Full-text search query (string)
- gitBranch: Git branch name (string)
- gitCommitSha: Git commit SHA (string)

Time references like "today", "yesterday", "last week", "last 24 hours" should be converted to ISO 8601 dates.
Current time reference: ` + time.Now().Format(time.RFC3339) + `

Respond with a JSON object containing:
1. "interpretation": A human-readable description of what the query means
2. "filter": An object with the filter fields to apply (only include fields that are relevant)
3. "suggestions": Array of related queries the user might want to try

Example input: "Show me failed traces from yesterday with high cost"
Example output:
{
  "interpretation": "Traces with errors from yesterday that cost more than average",
  "filter": {
    "hasError": true,
    "fromTime": "2024-01-08T00:00:00Z",
    "toTime": "2024-01-09T00:00:00Z",
    "minCost": 0.01
  },
  "suggestions": [
    "Show me the most expensive traces this week",
    "Which users had failed traces yesterday?"
  ]
}`

	userPrompt := fmt.Sprintf("Convert this query to a structured filter:\n\n%s", query)

	// Call OpenAI API
	response, err := s.callOpenAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Parse the response
	var parsed ParsedQuery
	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		// Try to extract JSON from response
		start := strings.Index(response, "{")
		end := strings.LastIndex(response, "}")
		if start >= 0 && end > start {
			jsonStr := response[start : end+1]
			if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
				return nil, fmt.Errorf("failed to parse LLM response: %w", err)
			}
		} else {
			return nil, fmt.Errorf("invalid LLM response format")
		}
	}

	return &parsed, nil
}

// callOpenAI makes a call to the OpenAI API
func (s *NLQueryService) callOpenAI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	requestBody := map[string]interface{}{
		"model": s.config.Eval.DefaultModel,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.1,
		"max_tokens": 1000,
		"response_format": map[string]string{"type": "json_object"},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.Eval.APIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return result.Choices[0].Message.Content, nil
}

// convertFilter converts a ParsedFilter to a domain.TraceFilter
func (s *NLQueryService) convertFilter(parsed ParsedFilter, projectID uuid.UUID) *domain.TraceFilter {
	filter := &domain.TraceFilter{
		ProjectID: projectID,
		Name:      parsed.Name,
		UserID:    parsed.UserID,
		SessionID: parsed.SessionID,
		Tags:      parsed.Tags,
		HasError:  parsed.HasError,
		MinCost:   parsed.MinCost,
		MaxCost:   parsed.MaxCost,
		MinDuration: parsed.MinDuration,
		MaxDuration: parsed.MaxDuration,
		Search:    parsed.Search,
		GitBranch: parsed.GitBranch,
		GitCommitSha: parsed.GitCommit,
	}

	// Convert level string to Level type
	if parsed.Level != nil {
		level := domain.Level(*parsed.Level)
		filter.Level = &level
	}

	// Parse time strings
	if parsed.FromTime != nil {
		if t, err := time.Parse(time.RFC3339, *parsed.FromTime); err == nil {
			filter.FromTime = &t
		}
	}

	if parsed.ToTime != nil {
		if t, err := time.Parse(time.RFC3339, *parsed.ToTime); err == nil {
			filter.ToTime = &t
		}
	}

	return filter
}

// GetQueryExamples returns example natural language queries
func (s *NLQueryService) GetQueryExamples() []string {
	return []string{
		"Show me traces with errors from the last 24 hours",
		"Find expensive traces that cost more than $0.10",
		"Traces from user john@example.com",
		"Show slow traces taking more than 5 seconds",
		"Recent traces on the main branch",
		"Traces tagged with 'production' from this week",
		"Failed agent runs from yesterday",
		"Show me traces with the name 'chat-completion'",
	}
}
