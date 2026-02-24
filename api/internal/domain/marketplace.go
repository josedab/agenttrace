package domain

import (
	"time"

	"github.com/google/uuid"
)

// PackageType represents the type of marketplace package
type PackageType string

const (
	PackagePrompt    PackageType = "prompt"
	PackageGuardrail PackageType = "guardrail"
	PackageEvaluator PackageType = "evaluator"
	PackageBenchmark PackageType = "benchmark"
	PackageBundle    PackageType = "bundle"
)

// MarketplacePackage represents a package in the marketplace
type MarketplacePackage struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        PackageType `json:"type"`
	Version     string      `json:"version"`
	Author      string      `json:"author"`
	Tags        []string    `json:"tags"`
	Downloads   int         `json:"downloads"`
	Rating      float64     `json:"rating"`
	RatingCount int         `json:"ratingCount"`
	IsPublic    bool        `json:"isPublic"`
	Content     string      `json:"content"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

// MarketplaceSearch represents search parameters for marketplace packages
type MarketplaceSearch struct {
	Query  string       `json:"query,omitempty"`
	Type   *PackageType `json:"type,omitempty"`
	Tags   []string     `json:"tags,omitempty"`
	SortBy string       `json:"sortBy,omitempty"` // downloads, rating, newest
	Limit  int          `json:"limit,omitempty"`
	Offset int          `json:"offset,omitempty"`
}

// PackagePublishInput represents input for publishing a package
type PackagePublishInput struct {
	Name        string      `json:"name" validate:"required"`
	Description string      `json:"description"`
	Type        PackageType `json:"type" validate:"required"`
	Version     string      `json:"version"`
	Tags        []string    `json:"tags"`
	Content     string      `json:"content" validate:"required"`
	IsPublic    bool        `json:"isPublic"`
}

// PackageRating represents a user rating for a package
type PackageRating struct {
	PackageID uuid.UUID `json:"packageId"`
	UserID    uuid.UUID `json:"userId"`
	Score     int       `json:"score"` // 1-5
	Review    string    `json:"review,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}
