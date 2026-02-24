package domain

import (
	"time"

	"github.com/google/uuid"
)

// EmbeddingModel represents the model used for generating embeddings
type EmbeddingModel string

const (
	EmbeddingModelE5Small EmbeddingModel = "e5-small-v2"
	EmbeddingModelOpenAI  EmbeddingModel = "text-embedding-3-small"
)

// TraceEmbedding stores a vector embedding for a trace
type TraceEmbedding struct {
	ID        uuid.UUID      `json:"id"`
	TraceID   string         `json:"traceId"`
	ProjectID uuid.UUID      `json:"projectId"`
	Content   string         `json:"content"`
	Model     EmbeddingModel `json:"model"`
	Vector    []float32      `json:"vector"`
	CreatedAt time.Time      `json:"createdAt"`
}

// EmbeddingConfig configures the embedding pipeline
type EmbeddingConfig struct {
	Enabled       bool           `json:"enabled"`
	Model         EmbeddingModel `json:"model"`
	BatchSize     int            `json:"batchSize"`
	MaxTokens     int            `json:"maxTokens"`
	IndexInterval int            `json:"indexIntervalSeconds"`
}

// SimilarityResult represents a similarity search result
type SimilarityResult struct {
	TraceID string  `json:"traceId"`
	Score   float64 `json:"score"`
	Content string  `json:"content"`
}
