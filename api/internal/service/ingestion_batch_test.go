package service

import (
	"context"
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestIngestBatch_MarshalFailure(t *testing.T) {
	// json.Marshal fails on values like math.NaN() or channels
	t.Run("trace metadata marshal failure returns error", func(t *testing.T) {
		svc := &IngestionService{}
		batch := &domain.IngestionBatch{
			Traces: []*domain.TraceInput{
				{
					Name:     "test-trace",
					Metadata: map[string]any{"bad": math.NaN()},
				},
			},
		}
		err := svc.IngestBatch(context.Background(), uuid.New(), batch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal trace metadata")
	})

	t.Run("observation metadata marshal failure returns error", func(t *testing.T) {
		svc := &IngestionService{}
		obsType := domain.ObservationTypeSpan
		batch := &domain.IngestionBatch{
			Observations: []*domain.ObservationInput{
				{
					Type:     &obsType,
					Metadata: map[string]any{"bad": math.NaN()},
				},
			},
		}
		err := svc.IngestBatch(context.Background(), uuid.New(), batch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal observation metadata")
	})

	t.Run("observation input marshal failure returns error", func(t *testing.T) {
		svc := &IngestionService{}
		obsType := domain.ObservationTypeSpan
		batch := &domain.IngestionBatch{
			Observations: []*domain.ObservationInput{
				{
					Type:  &obsType,
					Input: math.NaN(),
				},
			},
		}
		err := svc.IngestBatch(context.Background(), uuid.New(), batch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal observation input")
	})

	t.Run("generation metadata marshal failure returns error", func(t *testing.T) {
		svc := &IngestionService{}
		batch := &domain.IngestionBatch{
			Generations: []*domain.GenerationInput{
				{
					ObservationInput: domain.ObservationInput{
						Metadata: map[string]any{"bad": math.NaN()},
					},
				},
			},
		}
		err := svc.IngestBatch(context.Background(), uuid.New(), batch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal generation metadata")
	})

	t.Run("generation model params marshal failure returns error", func(t *testing.T) {
		svc := &IngestionService{}
		batch := &domain.IngestionBatch{
			Generations: []*domain.GenerationInput{
				{
					ModelParameters: map[string]any{"bad": math.NaN()},
				},
			},
		}
		err := svc.IngestBatch(context.Background(), uuid.New(), batch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal model parameters")
	})
}
