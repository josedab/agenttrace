package resolver

import (
	"context"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// ======== SUBSCRIPTION RESOLVERS ========

// TraceCreated subscribes to new traces
func (r *subscriptionResolver) TraceCreated(ctx context.Context, projectID uuid.UUID) (<-chan *domain.Trace, error) {
	// Create a channel for traces
	ch := make(chan *domain.Trace, 10)

	// Subscribe to real-time events
	// In a real implementation, this would use the realtime service
	go func() {
		<-ctx.Done()
		close(ch)
	}()

	return ch, nil
}

// TraceUpdated subscribes to trace updates
func (r *subscriptionResolver) TraceUpdated(ctx context.Context, projectID uuid.UUID) (<-chan *domain.Trace, error) {
	ch := make(chan *domain.Trace, 10)

	go func() {
		<-ctx.Done()
		close(ch)
	}()

	return ch, nil
}

// ObservationCreated subscribes to new observations
func (r *subscriptionResolver) ObservationCreated(ctx context.Context, projectID uuid.UUID) (<-chan *domain.Observation, error) {
	ch := make(chan *domain.Observation, 10)

	go func() {
		<-ctx.Done()
		close(ch)
	}()

	return ch, nil
}

// ScoreCreated subscribes to new scores
func (r *subscriptionResolver) ScoreCreated(ctx context.Context, projectID uuid.UUID) (<-chan *domain.Score, error) {
	ch := make(chan *domain.Score, 10)

	go func() {
		<-ctx.Done()
		close(ch)
	}()

	return ch, nil
}

// ======== HELPER TYPES ========

// subscriptionResolver implements Subscription resolvers
type subscriptionResolver struct {
	*Resolver
}

// Subscription returns the subscription resolver
func (r *Resolver) Subscription() SubscriptionResolver {
	return &subscriptionResolver{r}
}

// SubscriptionResolver interface
type SubscriptionResolver interface {
	TraceCreated(ctx context.Context, projectID uuid.UUID) (<-chan *domain.Trace, error)
	TraceUpdated(ctx context.Context, projectID uuid.UUID) (<-chan *domain.Trace, error)
	ObservationCreated(ctx context.Context, projectID uuid.UUID) (<-chan *domain.Observation, error)
	ScoreCreated(ctx context.Context, projectID uuid.UUID) (<-chan *domain.Score, error)
}
