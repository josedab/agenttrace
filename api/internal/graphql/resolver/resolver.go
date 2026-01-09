package resolver

import (
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/service"
)

// Resolver holds all service dependencies for GraphQL resolvers
type Resolver struct {
	logger          *zap.Logger
	queryService    *service.QueryService
	ingestionService *service.IngestionService
	scoreService    *service.ScoreService
	promptService   *service.PromptService
	datasetService  *service.DatasetService
	evalService     *service.EvalService
	authService     *service.AuthService
	orgService      *service.OrgService
	projectService  *service.ProjectService
	costService     *service.CostService
}

// NewResolver creates a new resolver with all service dependencies
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
		logger:           logger,
		queryService:     queryService,
		ingestionService: ingestionService,
		scoreService:     scoreService,
		promptService:    promptService,
		datasetService:   datasetService,
		evalService:      evalService,
		authService:      authService,
		orgService:       orgService,
		projectService:   projectService,
		costService:      costService,
	}
}

// ContextKey type for context keys
type ContextKey string

const (
	// ContextKeyProjectID holds the project ID in context
	ContextKeyProjectID ContextKey = "projectID"
	// ContextKeyUserID holds the user ID in context
	ContextKeyUserID ContextKey = "userID"
)

// GetProjectIDFromContext extracts project ID from context
func GetProjectIDFromContext(ctx interface{}) (uuid.UUID, bool) {
	// This would be implemented based on how context is passed in GraphQL
	// For now, return empty - actual implementation depends on middleware setup
	return uuid.UUID{}, false
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx interface{}) (uuid.UUID, bool) {
	return uuid.UUID{}, false
}
