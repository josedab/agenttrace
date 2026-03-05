package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// TraceKBService manages the trace annotation knowledge base
type TraceKBService struct {
	logger *zap.Logger
	mu     sync.RWMutex
	entries map[uuid.UUID]*domain.KBEntry
}

// NewTraceKBService creates a new trace knowledge base service
func NewTraceKBService(logger *zap.Logger) *TraceKBService {
	return &TraceKBService{
		logger:  logger,
		entries: make(map[uuid.UUID]*domain.KBEntry),
	}
}

// ListEntries returns KB entries for a project with pagination
func (s *TraceKBService) ListEntries(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]domain.KBEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var all []domain.KBEntry
	for _, e := range s.entries {
		if e.ProjectID == projectID {
			all = append(all, *e)
		}
	}

	if offset >= len(all) {
		return []domain.KBEntry{}, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

// CreateEntry creates a new knowledge base entry with validation
func (s *TraceKBService) CreateEntry(ctx context.Context, projectID, userID uuid.UUID, input *domain.KBEntryInput) (*domain.KBEntry, error) {
	if input.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if input.Category == "" {
		return nil, fmt.Errorf("category is required")
	}
	if input.Description == "" {
		return nil, fmt.Errorf("description is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	entry := &domain.KBEntry{
		ID:            uuid.New(),
		ProjectID:     projectID,
		TraceID:       input.TraceID,
		ObservationID: input.ObservationID,
		Category:      input.Category,
		Title:         input.Title,
		Description:   input.Description,
		Tags:          input.Tags,
		RootCause:     input.RootCause,
		Pattern:       input.Pattern,
		Fix:           input.Fix,
		CreatedBy:     userID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if entry.Tags == nil {
		entry.Tags = []string{}
	}

	s.entries[entry.ID] = entry
	s.logger.Info("created KB entry",
		zap.String("id", entry.ID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("title", entry.Title),
	)
	return entry, nil
}

// GetEntry returns a single knowledge base entry
func (s *TraceKBService) GetEntry(ctx context.Context, entryID uuid.UUID) (*domain.KBEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.entries[entryID]
	if !ok {
		return nil, fmt.Errorf("KB entry not found: %s", entryID)
	}
	return entry, nil
}

// Search performs full-text search across KB entries using simple substring matching
func (s *TraceKBService) Search(ctx context.Context, projectID uuid.UUID, input *domain.KBSearchInput) (*domain.KBSearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := strings.ToLower(input.Query)
	var matched []domain.KBEntry

	for _, e := range s.entries {
		if e.ProjectID != projectID {
			continue
		}
		if input.Category != nil && e.Category != *input.Category {
			continue
		}

		// Substring match on title and description
		if query != "" &&
			!strings.Contains(strings.ToLower(e.Title), query) &&
			!strings.Contains(strings.ToLower(e.Description), query) {
			continue
		}

		// Tag filter
		if len(input.Tags) > 0 && !hasAnyTag(e.Tags, input.Tags) {
			continue
		}

		matched = append(matched, *e)
	}

	totalCount := int64(len(matched))

	// Apply pagination
	offset := input.Offset
	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	if offset >= len(matched) {
		matched = []domain.KBEntry{}
	} else {
		end := offset + limit
		if end > len(matched) {
			end = len(matched)
		}
		matched = matched[offset:end]
	}

	return &domain.KBSearchResult{
		Entries:    matched,
		TotalCount: totalCount,
		Query:      input.Query,
	}, nil
}

// GetSuggestions returns similar KB entries based on trace characteristics
func (s *TraceKBService) GetSuggestions(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.KBSuggestion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var suggestions []domain.KBSuggestion
	for _, e := range s.entries {
		if e.ProjectID != projectID {
			continue
		}

		relevance := 0.0
		reason := ""

		// Exact trace match gets highest relevance
		if e.TraceID == traceID {
			relevance = 1.0
			reason = "Direct annotation on this trace"
		} else if e.Category == "error_pattern" {
			relevance = 0.6
			reason = "Similar error pattern detected"
		} else if e.Category == "performance" {
			relevance = 0.4
			reason = "Related performance insight"
		} else {
			relevance = 0.3
			reason = "Potentially related knowledge"
		}

		suggestions = append(suggestions, domain.KBSuggestion{
			EntryID:   e.ID,
			Title:     e.Title,
			Category:  e.Category,
			Relevance: relevance,
			Reason:    reason,
		})
	}

	if suggestions == nil {
		suggestions = []domain.KBSuggestion{}
	}

	s.logger.Debug("generated KB suggestions",
		zap.String("projectId", projectID.String()),
		zap.String("traceId", traceID),
		zap.Int("count", len(suggestions)),
	)
	return suggestions, nil
}

// hasAnyTag checks if any of the filter tags appear in the entry tags
func hasAnyTag(entryTags, filterTags []string) bool {
	tagSet := make(map[string]struct{}, len(entryTags))
	for _, t := range entryTags {
		tagSet[t] = struct{}{}
	}
	for _, t := range filterTags {
		if _, ok := tagSet[t]; ok {
			return true
		}
	}
	return false
}
