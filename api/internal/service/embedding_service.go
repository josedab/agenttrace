package service

import (
	"context"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// EmbeddingService manages trace embeddings for semantic search
type EmbeddingService struct {
	logger     *zap.Logger
	embeddings []domain.TraceEmbedding
	mu         sync.RWMutex
	config     domain.EmbeddingConfig
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(logger *zap.Logger) *EmbeddingService {
	return &EmbeddingService{
		logger: logger,
		config: domain.EmbeddingConfig{
			Enabled:       true,
			Model:         domain.EmbeddingModelE5Small,
			BatchSize:     32,
			MaxTokens:     512,
			IndexInterval: 60,
		},
	}
}

// IndexTrace creates an embedding for a trace
func (s *EmbeddingService) IndexTrace(ctx context.Context, projectID uuid.UUID, traceID, content string) error {
	vector := s.generateSimpleEmbedding(content)

	embedding := domain.TraceEmbedding{
		ID:        uuid.New(),
		TraceID:   traceID,
		ProjectID: projectID,
		Content:   content,
		Model:     s.config.Model,
		Vector:    vector,
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	s.embeddings = append(s.embeddings, embedding)
	s.mu.Unlock()

	return nil
}

// SearchSimilar finds traces similar to the query
func (s *EmbeddingService) SearchSimilar(ctx context.Context, projectID uuid.UUID, query string, limit int) ([]domain.SimilarityResult, error) {
	queryVector := s.generateSimpleEmbedding(query)

	s.mu.RLock()
	defer s.mu.RUnlock()

	type scored struct {
		traceID string
		content string
		score   float64
	}

	var results []scored
	for _, emb := range s.embeddings {
		if emb.ProjectID != projectID {
			continue
		}
		score := cosineSimilarity(queryVector, emb.Vector)
		results = append(results, scored{
			traceID: emb.TraceID,
			content: emb.Content,
			score:   score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	var output []domain.SimilarityResult
	for _, r := range results {
		output = append(output, domain.SimilarityResult{
			TraceID: r.traceID,
			Score:   r.score,
			Content: r.content,
		})
	}
	return output, nil
}

// GetConfig returns the current embedding configuration
func (s *EmbeddingService) GetConfig() domain.EmbeddingConfig {
	return s.config
}

// GetIndexStats returns statistics about the embedding index
func (s *EmbeddingService) GetIndexStats(projectID uuid.UUID) map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, emb := range s.embeddings {
		if emb.ProjectID == projectID {
			count++
		}
	}

	return map[string]any{
		"totalEmbeddings": count,
		"model":           string(s.config.Model),
		"enabled":         s.config.Enabled,
	}
}

// Simple bag-of-words embedding for in-memory use (production would use real model)
func (s *EmbeddingService) generateSimpleEmbedding(text string) []float32 {
	const dims = 128
	vector := make([]float32, dims)
	words := strings.Fields(strings.ToLower(text))
	for _, word := range words {
		hash := uint32(0)
		for _, c := range word {
			hash = hash*31 + uint32(c)
		}
		idx := hash % uint32(dims)
		vector[idx] += 1.0
	}
	// Normalize
	var norm float32
	for _, v := range vector {
		norm += v * v
	}
	if norm > 0 {
		norm = float32(math.Sqrt(float64(norm)))
		for i := range vector {
			vector[i] /= norm
		}
	}
	return vector
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
