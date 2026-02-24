package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestNewStreamingService(t *testing.T) {
	realtime := NewRealtimeService()
	svc := NewStreamingService(zap.NewNop(), realtime)
	assert.NotNil(t, svc)
}

func TestStreamingService_GetOrCreateStream(t *testing.T) {
	realtime := NewRealtimeService()
	svc := NewStreamingService(zap.NewNop(), realtime)

	traceID := uuid.New()
	projectID := uuid.New()

	t.Run("creates new stream", func(t *testing.T) {
		stream := svc.GetOrCreateStream(traceID, projectID)
		require.NotNil(t, stream)
		assert.Equal(t, traceID, stream.TraceID)
		assert.Equal(t, projectID, stream.ProjectID)
		assert.NotNil(t, stream.Metrics)
	})

	t.Run("returns existing stream", func(t *testing.T) {
		stream1 := svc.GetOrCreateStream(traceID, projectID)
		stream2 := svc.GetOrCreateStream(traceID, projectID)
		assert.Equal(t, stream1, stream2)
	})
}

func TestStreamingService_PublishAndSubscribe(t *testing.T) {
	realtime := NewRealtimeService()
	svc := NewStreamingService(zap.NewNop(), realtime)

	traceID := uuid.New()
	projectID := uuid.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("subscriber receives activities", func(t *testing.T) {
		filter := domain.StreamSubscription{FollowMode: true}
		sub := svc.SubscribeToTrace(ctx, traceID, projectID, filter)
		require.NotNil(t, sub)

		activity := domain.StreamActivity{
			ID:        uuid.New().String(),
			TraceID:   traceID,
			Type:      domain.StreamEventObservationStart,
			Title:     "Test observation",
			Timestamp: time.Now(),
			Status:    "running",
		}

		svc.PublishActivity(ctx, activity)

		select {
		case received := <-sub.Channel:
			assert.Equal(t, activity.Title, received.Title)
			assert.Equal(t, activity.Type, received.Type)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for activity")
		}
	})
}

func TestStreamingService_LiveMetrics(t *testing.T) {
	realtime := NewRealtimeService()
	svc := NewStreamingService(zap.NewNop(), realtime)

	traceID := uuid.New()
	projectID := uuid.New()
	ctx := context.Background()

	t.Run("returns nil for unknown trace", func(t *testing.T) {
		metrics := svc.GetLiveMetrics(uuid.New())
		assert.Nil(t, metrics)
	})

	t.Run("updates metrics on observation events", func(t *testing.T) {
		svc.GetOrCreateStream(traceID, projectID)

		svc.PublishActivity(ctx, domain.StreamActivity{
			ID:      uuid.New().String(),
			TraceID: traceID,
			Type:    domain.StreamEventObservationStart,
			Status:  "running",
		})

		metrics := svc.GetLiveMetrics(traceID)
		require.NotNil(t, metrics)
		assert.Equal(t, 1, metrics.ActiveSpans)

		svc.PublishActivity(ctx, domain.StreamActivity{
			ID:      uuid.New().String(),
			TraceID: traceID,
			Type:    domain.StreamEventObservationEnd,
			Status:  "completed",
			Metadata: map[string]any{
				"tokens": float64(100),
				"cost":   float64(0.005),
			},
		})

		metrics = svc.GetLiveMetrics(traceID)
		assert.Equal(t, 0, metrics.ActiveSpans)
		assert.Equal(t, 1, metrics.CompletedSpans)
		assert.Equal(t, 100, metrics.TotalTokens)
		assert.InDelta(t, 0.005, metrics.TotalCost, 0.0001)
	})

	t.Run("tracks file and terminal events", func(t *testing.T) {
		svc.PublishActivity(ctx, domain.StreamActivity{
			ID: uuid.New().String(), TraceID: traceID,
			Type: domain.StreamEventFileChange, Status: "completed",
		})
		svc.PublishActivity(ctx, domain.StreamActivity{
			ID: uuid.New().String(), TraceID: traceID,
			Type: domain.StreamEventTerminalOutput, Status: "completed",
		})
		svc.PublishActivity(ctx, domain.StreamActivity{
			ID: uuid.New().String(), TraceID: traceID,
			Type: domain.StreamEventErrorOccurred, Status: "error",
		})

		metrics := svc.GetLiveMetrics(traceID)
		assert.Equal(t, 1, metrics.FilesModified)
		assert.Equal(t, 1, metrics.TerminalCommands)
		assert.Equal(t, 1, metrics.ErrorCount)
	})
}

func TestStreamingService_RecentActivities(t *testing.T) {
	realtime := NewRealtimeService()
	svc := NewStreamingService(zap.NewNop(), realtime)

	traceID := uuid.New()
	projectID := uuid.New()
	ctx := context.Background()

	svc.GetOrCreateStream(traceID, projectID)

	for i := 0; i < 10; i++ {
		svc.PublishActivity(ctx, domain.StreamActivity{
			ID: uuid.New().String(), TraceID: traceID,
			Type: domain.StreamEventTraceActivity, Title: "Activity",
			Status: "completed",
		})
	}

	t.Run("returns limited activities", func(t *testing.T) {
		activities := svc.GetRecentActivities(traceID, 5)
		assert.Len(t, activities, 5)
	})

	t.Run("returns all if limit exceeds count", func(t *testing.T) {
		activities := svc.GetRecentActivities(traceID, 100)
		assert.Len(t, activities, 10)
	})

	t.Run("returns nil for unknown trace", func(t *testing.T) {
		activities := svc.GetRecentActivities(uuid.New(), 10)
		assert.Nil(t, activities)
	})
}

func TestStreamingService_Intervention(t *testing.T) {
	realtime := NewRealtimeService()
	svc := NewStreamingService(zap.NewNop(), realtime)

	traceID := uuid.New()
	projectID := uuid.New()
	ctx := context.Background()

	svc.GetOrCreateStream(traceID, projectID)

	t.Run("creates and retrieves intervention", func(t *testing.T) {
		req := domain.InterventionRequest{
			ID:        uuid.New(),
			TraceID:   traceID,
			ProjectID: projectID,
			Action:    domain.InterventionPause,
			Message:   "Please pause",
			CreatedAt: time.Now(),
			Status:    "pending",
		}

		err := svc.RequestIntervention(ctx, req)
		require.NoError(t, err)

		pending := svc.GetPendingInterventions(traceID)
		assert.Len(t, pending, 1)
		assert.Equal(t, domain.InterventionPause, pending[0].Action)
	})

	t.Run("acknowledges intervention", func(t *testing.T) {
		pending := svc.GetPendingInterventions(traceID)
		require.Len(t, pending, 1)

		err := svc.AcknowledgeIntervention(traceID, pending[0].ID)
		require.NoError(t, err)

		pending = svc.GetPendingInterventions(traceID)
		assert.Len(t, pending, 0)
	})
}

func TestStreamingService_ActiveStreams(t *testing.T) {
	realtime := NewRealtimeService()
	svc := NewStreamingService(zap.NewNop(), realtime)

	projectID := uuid.New()

	svc.GetOrCreateStream(uuid.New(), projectID)
	svc.GetOrCreateStream(uuid.New(), projectID)
	svc.GetOrCreateStream(uuid.New(), uuid.New()) // different project

	streams := svc.GetActiveStreams(projectID)
	assert.Len(t, streams, 2)
}

func TestStreamingService_CleanupStream(t *testing.T) {
	realtime := NewRealtimeService()
	svc := NewStreamingService(zap.NewNop(), realtime)

	traceID := uuid.New()
	projectID := uuid.New()

	svc.GetOrCreateStream(traceID, projectID)
	assert.NotNil(t, svc.GetLiveMetrics(traceID))

	svc.CleanupStream(traceID)
	assert.Nil(t, svc.GetLiveMetrics(traceID))
}
