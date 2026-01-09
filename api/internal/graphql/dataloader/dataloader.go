package dataloader

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ContextKey is the key type for context values
type ContextKey string

const (
	// LoadersKey is the context key for dataloaders
	LoadersKey ContextKey = "dataloaders"
)

// Loaders holds all dataloaders
type Loaders struct {
	// User loaders
	UserByID *UserLoader

	// Organization loaders
	OrganizationByID *OrganizationLoader

	// Project loaders
	ProjectByID *ProjectLoader

	// Trace loaders
	TraceByID           *TraceLoader
	ObservationsByTrace *ObservationsLoader
	ScoresByTrace       *ScoresLoader

	// Prompt loaders
	PromptByID          *PromptLoader
	PromptVersionsByID  *PromptVersionsLoader

	// Dataset loaders
	DatasetByID       *DatasetLoader
	DatasetItemsByID  *DatasetItemsLoader
	DatasetRunsByID   *DatasetRunsLoader

	// Evaluator loaders
	EvaluatorByID *EvaluatorLoader
}

// NewLoaders creates new dataloaders
func NewLoaders(
	queryService *service.QueryService,
	authService *service.AuthService,
	orgService *service.OrgService,
	projectService *service.ProjectService,
	promptService *service.PromptService,
	datasetService *service.DatasetService,
	evalService *service.EvalService,
	scoreService *service.ScoreService,
) *Loaders {
	return &Loaders{
		UserByID:            NewUserLoader(authService),
		OrganizationByID:    NewOrganizationLoader(orgService),
		ProjectByID:         NewProjectLoader(projectService),
		TraceByID:           NewTraceLoader(queryService),
		ObservationsByTrace: NewObservationsLoader(queryService),
		ScoresByTrace:       NewScoresLoader(scoreService),
		PromptByID:          NewPromptLoader(promptService),
		PromptVersionsByID:  NewPromptVersionsLoader(promptService),
		DatasetByID:         NewDatasetLoader(datasetService),
		DatasetItemsByID:    NewDatasetItemsLoader(datasetService),
		DatasetRunsByID:     NewDatasetRunsLoader(datasetService),
		EvaluatorByID:       NewEvaluatorLoader(evalService),
	}
}

// For retrieves dataloaders from context
func For(ctx context.Context) *Loaders {
	return ctx.Value(LoadersKey).(*Loaders)
}

// Middleware is middleware to inject dataloaders into context
func Middleware(loaders *Loaders) func(ctx context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, LoadersKey, loaders)
	}
}

// ======== GENERIC DATALOADER ========

// Result is a generic result type
type Result[T any] struct {
	Data  T
	Error error
}

// Loader is a generic dataloader
type Loader[K comparable, V any] struct {
	fetch    func(keys []K) (map[K]V, error)
	wait     time.Duration
	maxBatch int

	mu      sync.Mutex
	batch   []K
	results map[K]chan Result[V]
	timer   *time.Timer
}

// NewLoader creates a new dataloader
func NewLoader[K comparable, V any](fetch func(keys []K) (map[K]V, error)) *Loader[K, V] {
	return &Loader[K, V]{
		fetch:    fetch,
		wait:     2 * time.Millisecond,
		maxBatch: 100,
		results:  make(map[K]chan Result[V]),
	}
}

// Load loads a single value
func (l *Loader[K, V]) Load(ctx context.Context, key K) (V, error) {
	l.mu.Lock()

	// Check if already in batch
	if ch, ok := l.results[key]; ok {
		l.mu.Unlock()
		result := <-ch
		return result.Data, result.Error
	}

	// Add to batch
	l.batch = append(l.batch, key)
	ch := make(chan Result[V], 1)
	l.results[key] = ch

	// Start timer if first in batch
	if len(l.batch) == 1 {
		l.timer = time.AfterFunc(l.wait, func() {
			l.dispatch()
		})
	}

	// Dispatch if batch is full
	if len(l.batch) >= l.maxBatch {
		l.timer.Stop()
		go l.dispatch()
	}

	l.mu.Unlock()

	// Wait for result
	select {
	case result := <-ch:
		return result.Data, result.Error
	case <-ctx.Done():
		var zero V
		return zero, ctx.Err()
	}
}

// LoadAll loads multiple values
func (l *Loader[K, V]) LoadAll(ctx context.Context, keys []K) ([]V, error) {
	results := make([]V, len(keys))
	var firstErr error

	for i, key := range keys {
		result, err := l.Load(ctx, key)
		if err != nil && firstErr == nil {
			firstErr = err
		}
		results[i] = result
	}

	return results, firstErr
}

func (l *Loader[K, V]) dispatch() {
	l.mu.Lock()
	batch := l.batch
	results := l.results
	l.batch = nil
	l.results = make(map[K]chan Result[V])
	l.mu.Unlock()

	if len(batch) == 0 {
		return
	}

	// Fetch all at once
	data, err := l.fetch(batch)

	// Send results
	for _, key := range batch {
		ch := results[key]
		if err != nil {
			ch <- Result[V]{Error: err}
		} else if val, ok := data[key]; ok {
			ch <- Result[V]{Data: val}
		} else {
			var zero V
			ch <- Result[V]{Data: zero}
		}
		close(ch)
	}
}

// ======== SPECIFIC LOADERS ========

// UserLoader loads users by ID
type UserLoader struct {
	*Loader[uuid.UUID, *domain.User]
}

// NewUserLoader creates a user loader
func NewUserLoader(authService *service.AuthService) *UserLoader {
	return &UserLoader{
		Loader: NewLoader(func(keys []uuid.UUID) (map[uuid.UUID]*domain.User, error) {
			result := make(map[uuid.UUID]*domain.User, len(keys))
			for _, id := range keys {
				user, err := authService.GetUserByID(context.Background(), id)
				if err == nil {
					result[id] = user
				}
			}
			return result, nil
		}),
	}
}

// OrganizationLoader loads organizations by ID
type OrganizationLoader struct {
	*Loader[uuid.UUID, *domain.Organization]
}

// NewOrganizationLoader creates an organization loader
func NewOrganizationLoader(orgService *service.OrgService) *OrganizationLoader {
	return &OrganizationLoader{
		Loader: NewLoader(func(keys []uuid.UUID) (map[uuid.UUID]*domain.Organization, error) {
			result := make(map[uuid.UUID]*domain.Organization, len(keys))
			for _, id := range keys {
				org, err := orgService.Get(context.Background(), id)
				if err == nil {
					result[id] = org
				}
			}
			return result, nil
		}),
	}
}

// ProjectLoader loads projects by ID
type ProjectLoader struct {
	*Loader[uuid.UUID, *domain.Project]
}

// NewProjectLoader creates a project loader
func NewProjectLoader(projectService *service.ProjectService) *ProjectLoader {
	return &ProjectLoader{
		Loader: NewLoader(func(keys []uuid.UUID) (map[uuid.UUID]*domain.Project, error) {
			result := make(map[uuid.UUID]*domain.Project, len(keys))
			for _, id := range keys {
				project, err := projectService.Get(context.Background(), id)
				if err == nil {
					result[id] = project
				}
			}
			return result, nil
		}),
	}
}

// TraceLoader loads traces by ID
type TraceLoader struct {
	*Loader[string, *domain.Trace]
}

// NewTraceLoader creates a trace loader
func NewTraceLoader(queryService *service.QueryService) *TraceLoader {
	return &TraceLoader{
		Loader: NewLoader(func(keys []string) (map[string]*domain.Trace, error) {
			result := make(map[string]*domain.Trace, len(keys))
			for _, id := range keys {
				// Note: This needs project ID in real implementation
				trace, err := queryService.GetTrace(context.Background(), uuid.Nil, id)
				if err == nil {
					result[id] = trace
				}
			}
			return result, nil
		}),
	}
}

// ObservationsLoader loads observations by trace ID
type ObservationsLoader struct {
	*Loader[string, []*domain.Observation]
}

// NewObservationsLoader creates an observations loader
func NewObservationsLoader(queryService *service.QueryService) *ObservationsLoader {
	return &ObservationsLoader{
		Loader: NewLoader(func(keys []string) (map[string][]*domain.Observation, error) {
			result := make(map[string][]*domain.Observation, len(keys))
			for _, traceID := range keys {
				observations, err := queryService.GetObservationsByTraceID(context.Background(), uuid.Nil, traceID)
				if err == nil {
					// Convert []domain.Observation to []*domain.Observation
					ptrs := make([]*domain.Observation, len(observations))
					for i := range observations {
						ptrs[i] = &observations[i]
					}
					result[traceID] = ptrs
				}
			}
			return result, nil
		}),
	}
}

// ScoresLoader loads scores by trace ID
type ScoresLoader struct {
	*Loader[string, []*domain.Score]
}

// NewScoresLoader creates a scores loader
func NewScoresLoader(scoreService *service.ScoreService) *ScoresLoader {
	return &ScoresLoader{
		Loader: NewLoader(func(keys []string) (map[string][]*domain.Score, error) {
			result := make(map[string][]*domain.Score, len(keys))
			for _, traceID := range keys {
				scores, err := scoreService.GetByTraceID(context.Background(), uuid.Nil, traceID)
				if err == nil {
					// Convert []domain.Score to []*domain.Score
					ptrs := make([]*domain.Score, len(scores))
					for i := range scores {
						ptrs[i] = &scores[i]
					}
					result[traceID] = ptrs
				}
			}
			return result, nil
		}),
	}
}

// PromptLoader loads prompts by ID
type PromptLoader struct {
	*Loader[uuid.UUID, *domain.Prompt]
}

// NewPromptLoader creates a prompt loader
func NewPromptLoader(promptService *service.PromptService) *PromptLoader {
	return &PromptLoader{
		Loader: NewLoader(func(keys []uuid.UUID) (map[uuid.UUID]*domain.Prompt, error) {
			result := make(map[uuid.UUID]*domain.Prompt, len(keys))
			for _, id := range keys {
				prompt, err := promptService.Get(context.Background(), id)
				if err == nil {
					result[id] = prompt
				}
			}
			return result, nil
		}),
	}
}

// PromptVersionsLoader loads prompt versions by prompt ID
type PromptVersionsLoader struct {
	*Loader[uuid.UUID, []*domain.PromptVersion]
}

// NewPromptVersionsLoader creates a prompt versions loader
func NewPromptVersionsLoader(promptService *service.PromptService) *PromptVersionsLoader {
	return &PromptVersionsLoader{
		Loader: NewLoader(func(keys []uuid.UUID) (map[uuid.UUID][]*domain.PromptVersion, error) {
			result := make(map[uuid.UUID][]*domain.PromptVersion, len(keys))
			for _, promptID := range keys {
				versions, err := promptService.ListVersions(context.Background(), promptID)
				if err == nil {
					// Convert []domain.PromptVersion to []*domain.PromptVersion
					ptrs := make([]*domain.PromptVersion, len(versions))
					for i := range versions {
						ptrs[i] = &versions[i]
					}
					result[promptID] = ptrs
				}
			}
			return result, nil
		}),
	}
}

// DatasetLoader loads datasets by ID
type DatasetLoader struct {
	*Loader[uuid.UUID, *domain.Dataset]
}

// NewDatasetLoader creates a dataset loader
func NewDatasetLoader(datasetService *service.DatasetService) *DatasetLoader {
	return &DatasetLoader{
		Loader: NewLoader(func(keys []uuid.UUID) (map[uuid.UUID]*domain.Dataset, error) {
			result := make(map[uuid.UUID]*domain.Dataset, len(keys))
			for _, id := range keys {
				dataset, err := datasetService.Get(context.Background(), id)
				if err == nil {
					result[id] = dataset
				}
			}
			return result, nil
		}),
	}
}

// DatasetItemsLoader loads dataset items by dataset ID
type DatasetItemsLoader struct {
	*Loader[uuid.UUID, []*domain.DatasetItem]
}

// NewDatasetItemsLoader creates a dataset items loader
func NewDatasetItemsLoader(datasetService *service.DatasetService) *DatasetItemsLoader {
	return &DatasetItemsLoader{
		Loader: NewLoader(func(keys []uuid.UUID) (map[uuid.UUID][]*domain.DatasetItem, error) {
			result := make(map[uuid.UUID][]*domain.DatasetItem, len(keys))
			for _, datasetID := range keys {
				filter := &domain.DatasetItemFilter{DatasetID: datasetID}
				items, _, err := datasetService.ListItems(context.Background(), filter, 1000, 0)
				if err == nil {
					// Convert []domain.DatasetItem to []*domain.DatasetItem
					ptrs := make([]*domain.DatasetItem, len(items))
					for i := range items {
						ptrs[i] = &items[i]
					}
					result[datasetID] = ptrs
				}
			}
			return result, nil
		}),
	}
}

// DatasetRunsLoader loads dataset runs by dataset ID
type DatasetRunsLoader struct {
	*Loader[uuid.UUID, []*domain.DatasetRun]
}

// NewDatasetRunsLoader creates a dataset runs loader
func NewDatasetRunsLoader(datasetService *service.DatasetService) *DatasetRunsLoader {
	return &DatasetRunsLoader{
		Loader: NewLoader(func(keys []uuid.UUID) (map[uuid.UUID][]*domain.DatasetRun, error) {
			result := make(map[uuid.UUID][]*domain.DatasetRun, len(keys))
			for _, datasetID := range keys {
				runs, _, err := datasetService.ListRuns(context.Background(), datasetID, 1000, 0)
				if err == nil {
					// Convert []domain.DatasetRun to []*domain.DatasetRun
					ptrs := make([]*domain.DatasetRun, len(runs))
					for i := range runs {
						ptrs[i] = &runs[i]
					}
					result[datasetID] = ptrs
				}
			}
			return result, nil
		}),
	}
}

// EvaluatorLoader loads evaluators by ID
type EvaluatorLoader struct {
	*Loader[uuid.UUID, *domain.Evaluator]
}

// NewEvaluatorLoader creates an evaluator loader
func NewEvaluatorLoader(evalService *service.EvalService) *EvaluatorLoader {
	return &EvaluatorLoader{
		Loader: NewLoader(func(keys []uuid.UUID) (map[uuid.UUID]*domain.Evaluator, error) {
			result := make(map[uuid.UUID]*domain.Evaluator, len(keys))
			for _, id := range keys {
				evaluator, err := evalService.Get(context.Background(), id)
				if err == nil {
					result[id] = evaluator
				}
			}
			return result, nil
		}),
	}
}
