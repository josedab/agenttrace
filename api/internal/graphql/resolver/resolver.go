package resolver

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/agenttrace/agenttrace/api/internal/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ContextKey is the key type for context values
type ContextKey string

const (
	// ContextKeyProjectID is the context key for project ID
	ContextKeyProjectID ContextKey = "projectID"
	// ContextKeyUserID is the context key for user ID
	ContextKeyUserID ContextKey = "userID"
	// ContextKeyOrgID is the context key for organization ID
	ContextKeyOrgID ContextKey = "orgID"
)

// Resolver holds all service dependencies for GraphQL resolvers
type Resolver struct {
	Logger           *zap.Logger
	QueryService     *service.QueryService
	IngestionService *service.IngestionService
	ScoreService     *service.ScoreService
	PromptService    *service.PromptService
	DatasetService   *service.DatasetService
	EvalService      *service.EvalService
	AuthService      *service.AuthService
	OrgService       *service.OrgService
	ProjectService   *service.ProjectService
	CostService      *service.CostService
}

// NewResolver creates a new resolver with all dependencies
func NewResolver(
	logger *zap.Logger,
	queryService *service.QueryService,
	ingestionService *service.IngestionService,
	scoreService *service.ScoreService,
	promptService *service.PromptService,
	datasetService *service.DatasetService,
	evalService *service.EvalService,
	authService *service.AuthService,
	orgService *service.OrgService,
	projectService *service.ProjectService,
	costService *service.CostService,
) *Resolver {
	return &Resolver{
		Logger:           logger.Named("graphql"),
		QueryService:     queryService,
		IngestionService: ingestionService,
		ScoreService:     scoreService,
		PromptService:    promptService,
		DatasetService:   datasetService,
		EvalService:      evalService,
		AuthService:      authService,
		OrgService:       orgService,
		ProjectService:   projectService,
		CostService:      costService,
	}
}

// Helper functions

// getProjectID extracts project ID from context
func getProjectID(ctx context.Context) (uuid.UUID, error) {
	projectID, ok := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	if !ok || projectID == uuid.Nil {
		return uuid.Nil, ErrProjectIDNotFound
	}
	return projectID, nil
}

// getUserID extracts user ID from context
func getUserID(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(ContextKeyUserID).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return uuid.Nil, ErrUserIDNotFound
	}
	return userID, nil
}

// encodeCursor encodes an offset as a cursor
func encodeCursor(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

// decodeCursor decodes a cursor to an offset
func decodeCursor(cursor string) (int, error) {
	data, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

// getLimit returns the limit with default
func getLimit(limit *int) int {
	if limit == nil || *limit <= 0 {
		return 20
	}
	if *limit > 100 {
		return 100
	}
	return *limit
}

// getOffset calculates offset from cursor
func getOffset(cursor *string) int {
	if cursor == nil || *cursor == "" {
		return 0
	}
	offset, err := decodeCursor(*cursor)
	if err != nil {
		return 0
	}
	return offset
}

// parseJSONString parses a JSON string into map[string]any
func parseJSONString(s string) (map[string]any, error) {
	if s == "" {
		return nil, nil
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// parseJSONStringPtr parses a JSON string pointer into map[string]any
func parseJSONStringPtr(s *string) (map[string]any, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	return parseJSONString(*s)
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// generateTraceID generates a random 32-character hex trace ID
func generateTraceID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%032x", b)
}

// generateSpanID generates a random 16-character hex span ID
func generateSpanID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%016x", b)
}

// slugify converts a string to a URL-friendly slug
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	s = result.String()
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	return s
}

// timeValue safely dereferences a time pointer with a default value
func timeValue(t *time.Time, defaultVal time.Time) time.Time {
	if t == nil {
		return defaultVal
	}
	return *t
}

// timeValuePtr returns the time pointer as-is (identity function)
func timeValuePtr(t *time.Time) *time.Time {
	return t
}
