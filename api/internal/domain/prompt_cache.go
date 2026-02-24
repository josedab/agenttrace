package domain

import (
	"time"

	"github.com/google/uuid"
)

// CacheSegmentType represents the type of cacheable segment
type CacheSegmentType string

const (
	CacheSegmentSystemPrompt  CacheSegmentType = "system_prompt"
	CacheSegmentFewShot       CacheSegmentType = "few_shot"
	CacheSegmentStaticContext CacheSegmentType = "static_context"
)

// CacheStrategy represents the caching strategy
type CacheStrategy string

const (
	CacheStrategyHash     CacheStrategy = "hash"
	CacheStrategySemantic CacheStrategy = "semantic"
)

// CacheableSegment represents a segment of a prompt that can be cached
type CacheableSegment struct {
	PromptName     string           `json:"promptName"`
	SegmentType    CacheSegmentType `json:"segmentType"`
	TokenCount     int              `json:"tokenCount"`
	Frequency      int              `json:"frequency"`
	CacheHitRate   float64          `json:"cacheHitRate"`
	MonthlySavings float64          `json:"monthlySavings"`
}

// CacheAnalysis represents the analysis of prompt caching opportunities
type CacheAnalysis struct {
	ProjectID              uuid.UUID         `json:"projectId"`
	TotalPrompts           int               `json:"totalPrompts"`
	CacheableSegments      int               `json:"cacheableSegments"`
	EstimatedSavingsPct    float64           `json:"estimatedSavingsPct"`
	EstimatedMonthlySavings float64          `json:"estimatedMonthlySavings"`
	Segments               []CacheableSegment `json:"segments"`
}

// CacheConfig represents the configuration for prompt caching
type CacheConfig struct {
	ID                uuid.UUID     `json:"id"`
	ProjectID         uuid.UUID     `json:"projectId"`
	Enabled           bool          `json:"enabled"`
	Strategy          CacheStrategy `json:"strategy"`
	TTLSeconds        int           `json:"ttlSeconds"`
	MaxEntries        int           `json:"maxEntries"`
	InvalidateOnDrift bool          `json:"invalidateOnDrift"`
}

// CacheStats represents runtime statistics for the prompt cache
type CacheStats struct {
	ProjectID   uuid.UUID `json:"projectId"`
	HitCount    int64     `json:"hitCount"`
	MissCount   int64     `json:"missCount"`
	HitRate     float64   `json:"hitRate"`
	TotalSaved  float64   `json:"totalSaved"`
	AvgLookupMs float64   `json:"avgLookupMs"`
	Entries     int       `json:"entries"`
}

// CacheConfigInput is the input for updating cache configuration
type CacheConfigInput struct {
	Enabled           *bool          `json:"enabled,omitempty"`
	Strategy          *CacheStrategy `json:"strategy,omitempty"`
	TTLSeconds        *int           `json:"ttlSeconds,omitempty"`
	MaxEntries        *int           `json:"maxEntries,omitempty"`
	InvalidateOnDrift *bool          `json:"invalidateOnDrift,omitempty"`
}

// CacheInvalidation represents a cache invalidation event
type CacheInvalidation struct {
	ProjectID    uuid.UUID `json:"projectId"`
	EntriesCleared int     `json:"entriesCleared"`
	Timestamp    time.Time `json:"timestamp"`
}
