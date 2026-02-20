package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// QueryServiceInterface defines the contract for trace and observation queries.
// Handlers should depend on this interface rather than *QueryService directly.
type QueryServiceInterface interface {
	GetTrace(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error)
	ListTraces(ctx context.Context, filter *domain.TraceFilter, limit, offset int) (*domain.TraceList, error)
	GetObservation(ctx context.Context, projectID uuid.UUID, observationID string) (*domain.Observation, error)
	ListObservations(ctx context.Context, filter *domain.ObservationFilter, limit, offset int) ([]domain.Observation, int64, error)
	GetObservationsByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Observation, error)
	GetObservationTree(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.ObservationTree, error)
	GetSessionTraces(ctx context.Context, projectID uuid.UUID, sessionID string) ([]domain.Trace, error)
	ListSessions(ctx context.Context, filter *domain.SessionFilter, limit, offset int) (*domain.SessionList, error)
	GetSession(ctx context.Context, projectID uuid.UUID, sessionID string) (*domain.Session, error)
	SetBookmark(ctx context.Context, projectID uuid.UUID, traceID string, bookmarked bool) error
	DeleteTrace(ctx context.Context, projectID uuid.UUID, traceID string) error
	UpdateTrace(ctx context.Context, projectID uuid.UUID, traceID string, input *domain.TraceUpdateInput) (*domain.Trace, error)
	GetTraceStats(ctx context.Context, filter *domain.TraceFilter) (*TraceStats, error)
	GetGenerationStats(ctx context.Context, projectID uuid.UUID, model *string) (*GenerationStats, error)
}

// IngestionServiceInterface defines the contract for trace and observation ingestion.
type IngestionServiceInterface interface {
	IngestTrace(ctx context.Context, projectID uuid.UUID, input *domain.TraceInput) (*domain.Trace, error)
	IngestObservation(ctx context.Context, projectID uuid.UUID, input *domain.ObservationInput) (*domain.Observation, error)
	IngestGeneration(ctx context.Context, projectID uuid.UUID, input *domain.GenerationInput) (*domain.Observation, error)
	IngestBatch(ctx context.Context, projectID uuid.UUID, batch *domain.IngestionBatch) error
	UpdateTrace(ctx context.Context, projectID uuid.UUID, traceID string, input *domain.TraceInput) (*domain.Trace, error)
	UpdateObservation(ctx context.Context, projectID uuid.UUID, obsID string, input *domain.ObservationInput) (*domain.Observation, error)
}

// AuthServiceInterface defines the contract for authentication and authorization.
type AuthServiceInterface interface {
	Register(ctx context.Context, input *domain.RegisterInput) (*domain.AuthResult, error)
	Login(ctx context.Context, input *domain.LoginInput) (*domain.AuthResult, error)
	LoginWithContext(ctx context.Context, input *domain.LoginInput, ipAddress, userAgent string) (*domain.AuthResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResult, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutWithContext(ctx context.Context, refreshToken string, userID uuid.UUID, userEmail string) error
	ValidateJWT(ctx context.Context, tokenString string) (*domain.JWTClaims, error)
	ValidateAPIKey(ctx context.Context, publicKey, secretKey string) (*uuid.UUID, error)
	ValidateAPIKeyPublicOnly(ctx context.Context, publicKey string) (*uuid.UUID, error)
	CreateAPIKey(ctx context.Context, projectID uuid.UUID, input *domain.APIKeyInput, userID uuid.UUID) (*domain.APIKeyCreateResult, error)
	CreateAPIKeyWithContext(ctx context.Context, projectID uuid.UUID, input *domain.APIKeyInput, userID uuid.UUID, userEmail string) (*domain.APIKeyCreateResult, error)
	DeleteAPIKey(ctx context.Context, id uuid.UUID) error
	DeleteAPIKeyWithContext(ctx context.Context, id uuid.UUID, actorID uuid.UUID, actorEmail string) error
	ListAPIKeys(ctx context.Context, projectID uuid.UUID) ([]domain.APIKey, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	CheckProjectAccess(ctx context.Context, projectID, userID uuid.UUID, requiredRole domain.OrgRole) error
	HandleOAuthCallback(ctx context.Context, input *domain.OAuthCallbackInput) (*domain.AuthResult, error)
	HandleOAuthCallbackWithContext(ctx context.Context, input *domain.OAuthCallbackInput, ipAddress, userAgent string) (*domain.AuthResult, error)
}

// ProjectServiceInterface defines the contract for project management.
type ProjectServiceInterface interface {
	Create(ctx context.Context, orgID uuid.UUID, input *ProjectInput, userID uuid.UUID) (*domain.Project, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	GetBySlug(ctx context.Context, orgID uuid.UUID, slug string) (*domain.Project, error)
	Update(ctx context.Context, id uuid.UUID, input *ProjectInput) (*domain.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]domain.Project, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Project, error)
	AddMember(ctx context.Context, projectID, userID uuid.UUID, role domain.OrgRole) error
	RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error
	CheckAccess(ctx context.Context, projectID, userID uuid.UUID, requiredRole domain.OrgRole) error
	GetUserRole(ctx context.Context, projectID, userID uuid.UUID) (*domain.OrgRole, error)
	ListAll(ctx context.Context, limit, offset int) ([]domain.Project, error)
}

// DatasetServiceInterface defines the contract for dataset management.
type DatasetServiceInterface interface {
	Create(ctx context.Context, projectID uuid.UUID, input *domain.DatasetInput) (*domain.Dataset, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.Dataset, error)
	GetByName(ctx context.Context, projectID uuid.UUID, name string) (*domain.Dataset, error)
	Update(ctx context.Context, id uuid.UUID, input *domain.DatasetInput) (*domain.Dataset, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *domain.DatasetFilter, limit, offset int) (*domain.DatasetList, error)
	AddItem(ctx context.Context, datasetID uuid.UUID, input *domain.DatasetItemInput) (*domain.DatasetItem, error)
	AddItemFromTrace(ctx context.Context, datasetID uuid.UUID, projectID uuid.UUID, traceID string, observationID *string) (*domain.DatasetItem, error)
	UpdateItem(ctx context.Context, itemID uuid.UUID, input *domain.DatasetItemUpdateInput) (*domain.DatasetItem, error)
	DeleteItem(ctx context.Context, itemID uuid.UUID) error
	ListItems(ctx context.Context, filter *domain.DatasetItemFilter, limit, offset int) ([]domain.DatasetItem, int64, error)
	CreateRun(ctx context.Context, datasetID uuid.UUID, input *domain.DatasetRunInput) (*domain.DatasetRun, error)
	GetRun(ctx context.Context, id uuid.UUID) (*domain.DatasetRun, error)
	GetRunByName(ctx context.Context, datasetID uuid.UUID, name string) (*domain.DatasetRun, error)
	ListRuns(ctx context.Context, datasetID uuid.UUID, limit, offset int) ([]domain.DatasetRun, int64, error)
	AddRunItem(ctx context.Context, runID uuid.UUID, input *domain.DatasetRunItemInput) (*domain.DatasetRunItem, error)
	AddRunItemsBatch(ctx context.Context, runID uuid.UUID, inputs []*domain.DatasetRunItemInput) ([]*domain.DatasetRunItem, error)
	GetRunResults(ctx context.Context, projectID uuid.UUID, runID uuid.UUID) (*domain.DatasetRunResults, error)
}

// Compile-time interface compliance checks
var (
	_ QueryServiceInterface     = (*QueryService)(nil)
	_ IngestionServiceInterface = (*IngestionService)(nil)
	_ AuthServiceInterface      = (*AuthService)(nil)
	_ ProjectServiceInterface   = (*ProjectService)(nil)
	_ DatasetServiceInterface   = (*DatasetService)(nil)
)
