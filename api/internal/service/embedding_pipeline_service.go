package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// EmbeddingProvider defines the interface for generating text embeddings
type EmbeddingProvider interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
	Dimension() int
}

// EmbeddingRepository defines the interface for persisting embeddings
type EmbeddingRepository interface {
	SaveEmbedding(ctx context.Context, record *EmbeddingRecord) error
	FindSimilar(ctx context.Context, projectID uuid.UUID, embedding []float32, limit int, threshold float64) ([]SimilarityResult, error)
	DeleteByTraceID(ctx context.Context, traceID uuid.UUID) error
	CountByProject(ctx context.Context, projectID uuid.UUID) (int64, error)
}

// EmbeddingRecord represents a stored embedding with metadata
type EmbeddingRecord struct {
	ID          uuid.UUID `json:"id"`
	ProjectID   uuid.UUID `json:"projectId"`
	TraceID     uuid.UUID `json:"traceId"`
	ContentType string    `json:"contentType"` // trace, observation, generation
	ContentHash string    `json:"contentHash"`
	Text        string    `json:"text"`
	Embedding   []float32 `json:"embedding"`
	Metadata    map[string]interface{} `json:"metadata"`
	IndexedAt   time.Time `json:"indexedAt"`
}

// SimilarityResult represents a search result with similarity score
type SimilarityResult struct {
	TraceID     uuid.UUID              `json:"traceId"`
	ContentType string                 `json:"contentType"`
	Score       float64                `json:"score"`
	Text        string                 `json:"text"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// LocalEmbeddingProvider implements a simple TF-IDF-like embedding without external deps
type LocalEmbeddingProvider struct {
	dimension int
	vocab     map[string]int
}

// NewLocalEmbeddingProvider creates a local embedding provider
func NewLocalEmbeddingProvider() *LocalEmbeddingProvider {
	return &LocalEmbeddingProvider{
		dimension: 128,
		vocab:     make(map[string]int),
	}
}

// Embed generates a hash-based embedding for a text
func (p *LocalEmbeddingProvider) Embed(_ context.Context, text string) ([]float32, error) {
	return p.hashEmbed(text), nil
}

// EmbedBatch generates embeddings for multiple texts
func (p *LocalEmbeddingProvider) EmbedBatch(_ context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		results[i] = p.hashEmbed(text)
	}
	return results, nil
}

// Dimension returns the embedding dimension
func (p *LocalEmbeddingProvider) Dimension() int {
	return p.dimension
}

// hashEmbed creates a deterministic embedding via feature hashing
func (p *LocalEmbeddingProvider) hashEmbed(text string) []float32 {
	embedding := make([]float32, p.dimension)
	tokens := tokenize(text)

	for _, token := range tokens {
		hash := sha256.Sum256([]byte(token))
		idx := int(hash[0])<<8 | int(hash[1])
		idx = idx % p.dimension
		sign := float32(1.0)
		if hash[2]%2 == 0 {
			sign = -1.0
		}
		embedding[idx] += sign
	}

	// L2 normalize
	var norm float64
	for _, v := range embedding {
		norm += float64(v * v)
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range embedding {
			embedding[i] = float32(float64(embedding[i]) / norm)
		}
	}

	return embedding
}

// tokenize splits text into lowercase tokens
func tokenize(text string) []string {
	text = strings.ToLower(text)
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_')
	})
	// Also add bigrams for better semantic capture
	var tokens []string
	tokens = append(tokens, words...)
	for i := 0; i < len(words)-1; i++ {
		tokens = append(tokens, words[i]+"_"+words[i+1])
	}
	return tokens
}

// cosineSimilarity computes similarity between two vectors
func pipelineCosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}

// contentHash generates a deterministic hash for content deduplication
func contentHash(text string) string {
	h := sha256.Sum256([]byte(text))
	return hex.EncodeToString(h[:])
}

// EmbeddingPipelineService orchestrates the embedding pipeline
type EmbeddingPipelineService struct {
	logger   *zap.Logger
	provider EmbeddingProvider
	// In-memory index for search (production would use pgvector)
	index []EmbeddingRecord
}

// NewEmbeddingPipelineService creates a new embedding pipeline
func NewEmbeddingPipelineService(logger *zap.Logger) *EmbeddingPipelineService {
	return &EmbeddingPipelineService{
		logger:   logger,
		provider: NewLocalEmbeddingProvider(),
		index:    make([]EmbeddingRecord, 0),
	}
}

// IndexTrace embeds and indexes a trace's content
func (s *EmbeddingPipelineService) IndexTrace(
	ctx context.Context,
	projectID uuid.UUID,
	traceID uuid.UUID,
	traceName string,
	input interface{},
	output interface{},
) error {
	// Build searchable text from trace content
	text := buildTraceText(traceName, input, output)
	if text == "" {
		return nil
	}

	hash := contentHash(text)

	// Check for duplicate
	for _, record := range s.index {
		if record.TraceID == traceID && record.ContentHash == hash {
			return nil // Already indexed
		}
	}

	embedding, err := s.provider.Embed(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	record := EmbeddingRecord{
		ID:          uuid.New(),
		ProjectID:   projectID,
		TraceID:     traceID,
		ContentType: "trace",
		ContentHash: hash,
		Text:        truncateText(text, 1000),
		Embedding:   embedding,
		Metadata: map[string]interface{}{
			"traceName": traceName,
		},
		IndexedAt: time.Now(),
	}

	s.index = append(s.index, record)

	s.logger.Debug("indexed trace",
		zap.String("traceId", traceID.String()),
		zap.Int("textLength", len(text)),
	)

	return nil
}

// IndexObservation embeds an observation
func (s *EmbeddingPipelineService) IndexObservation(
	ctx context.Context,
	projectID uuid.UUID,
	traceID uuid.UUID,
	observationID uuid.UUID,
	name string,
	input interface{},
	output interface{},
) error {
	text := buildTraceText(name, input, output)
	if text == "" {
		return nil
	}

	embedding, err := s.provider.Embed(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	record := EmbeddingRecord{
		ID:          uuid.New(),
		ProjectID:   projectID,
		TraceID:     traceID,
		ContentType: "observation",
		ContentHash: contentHash(text),
		Text:        truncateText(text, 1000),
		Embedding:   embedding,
		Metadata: map[string]interface{}{
			"observationId": observationID.String(),
			"name":          name,
		},
		IndexedAt: time.Now(),
	}

	s.index = append(s.index, record)
	return nil
}

// Search finds traces similar to a natural language query
func (s *EmbeddingPipelineService) Search(
	ctx context.Context,
	projectID uuid.UUID,
	query string,
	limit int,
	minScore float64,
) ([]domain.SemanticTraceSearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if minScore <= 0 {
		minScore = 0.1
	}

	queryEmbedding, err := s.provider.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Score all indexed records for this project
	type scoredResult struct {
		record EmbeddingRecord
		score  float64
	}

	var scored []scoredResult
	for _, record := range s.index {
		if record.ProjectID != projectID {
			continue
		}
		score := pipelineCosineSimilarity(queryEmbedding, record.Embedding)
		if score >= minScore {
			scored = append(scored, scoredResult{record, score})
		}
	}

	// Sort by score descending
	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Limit results
	if len(scored) > limit {
		scored = scored[:limit]
	}

	// Convert to domain results
	results := make([]domain.SemanticTraceSearchResult, 0, len(scored))
	for _, s := range scored {
		traceName := ""
		if name, ok := s.record.Metadata["traceName"].(string); ok {
			traceName = name
		}

		result := domain.SemanticTraceSearchResult{
			ID:        s.record.ID,
			Type:      domain.SearchResultType(s.record.ContentType),
			Score:     s.score,
			Snippet:   truncateText(s.record.Text, 200),
			TraceID:   s.record.TraceID,
			TraceName: traceName,
			Timestamp: s.record.IndexedAt,
			Metadata:  s.record.Metadata,
		}
		results = append(results, result)
	}

	return results, nil
}

// ClusterTraces groups similar traces into clusters
func (s *EmbeddingPipelineService) ClusterTraces(
	ctx context.Context,
	projectID uuid.UUID,
	numClusters int,
) ([]domain.TraceCluster, error) {
	if numClusters <= 0 {
		numClusters = 5
	}

	// Collect project records
	var projectRecords []EmbeddingRecord
	for _, r := range s.index {
		if r.ProjectID == projectID {
			projectRecords = append(projectRecords, r)
		}
	}

	if len(projectRecords) == 0 {
		return []domain.TraceCluster{}, nil
	}

	// Simple k-means-inspired clustering using greedy assignment
	if numClusters > len(projectRecords) {
		numClusters = len(projectRecords)
	}

	// Pick initial centroids (first N distinct records)
	centroids := make([][]float32, numClusters)
	for i := 0; i < numClusters; i++ {
		centroids[i] = projectRecords[i].Embedding
	}

	// Assign each record to nearest centroid
	assignments := make([]int, len(projectRecords))
	for i, record := range projectRecords {
		bestCluster := 0
		bestScore := -1.0
		for c, centroid := range centroids {
			score := pipelineCosineSimilarity(record.Embedding, centroid)
			if score > bestScore {
				bestScore = score
				bestCluster = c
			}
		}
		assignments[i] = bestCluster
	}

	// Build cluster objects
	clusterMap := make(map[int][]EmbeddingRecord)
	for i, cluster := range assignments {
		clusterMap[cluster] = append(clusterMap[cluster], projectRecords[i])
	}

	var clusters []domain.TraceCluster
	for clusterIdx, records := range clusterMap {
		if len(records) == 0 {
			continue
		}

		// Extract common patterns from trace names
		nameFreq := make(map[string]int)
		var traceIDs []uuid.UUID
		for _, r := range records {
			if name, ok := r.Metadata["traceName"].(string); ok {
				nameFreq[name]++
			}
			traceIDs = append(traceIDs, r.TraceID)
		}

		var patterns []string
		for name, count := range nameFreq {
			if count > 1 {
				patterns = append(patterns, name)
			}
		}

		label := fmt.Sprintf("Cluster %d", clusterIdx+1)
		if len(patterns) > 0 {
			label = patterns[0]
		}

		cluster := domain.TraceCluster{
			ID:                     uuid.New(),
			Label:                  label,
			Description:            fmt.Sprintf("Group of %d similar traces", len(records)),
			TraceCount:             len(records),
			CommonPatterns:         patterns,
			RepresentativeTraceIDs: traceIDs,
		}
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

// GetIndexStats returns statistics about the embedding index
func (s *EmbeddingPipelineService) GetIndexStats(projectID uuid.UUID) map[string]interface{} {
	var total, traces, observations int
	for _, r := range s.index {
		if r.ProjectID == projectID {
			total++
			switch r.ContentType {
			case "trace":
				traces++
			case "observation":
				observations++
			}
		}
	}

	return map[string]interface{}{
		"totalIndexed":        total,
		"tracesIndexed":       traces,
		"observationsIndexed": observations,
		"embeddingDimension":  s.provider.Dimension(),
		"embeddingModel":      "local-hash-128d",
	}
}

// buildTraceText concatenates trace content into searchable text
func buildTraceText(name string, input interface{}, output interface{}) string {
	var parts []string
	if name != "" {
		parts = append(parts, name)
	}
	if input != nil {
		parts = append(parts, fmt.Sprintf("%v", input))
	}
	if output != nil {
		parts = append(parts, fmt.Sprintf("%v", output))
	}
	return strings.Join(parts, " ")
}

// truncateText truncates text to maxLen characters
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
