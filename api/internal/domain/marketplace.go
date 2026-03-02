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

// PackageVersion represents a specific version of a marketplace package
type PackageVersion struct {
	Version     string    `json:"version"`
	Content     string    `json:"content"`
	Changelog   string    `json:"changelog,omitempty"`
	Downloads   int       `json:"downloads"`
	PublishedAt time.Time `json:"publishedAt"`
}

// PackageInstall represents an installation record
type PackageInstall struct {
	PackageID   uuid.UUID `json:"packageId"`
	ProjectID   uuid.UUID `json:"projectId"`
	Version     string    `json:"version"`
	InstalledAt time.Time `json:"installedAt"`
}

// MarketplaceCategory represents a marketplace category
type MarketplaceCategory struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Count       int    `json:"count"`
	Icon        string `json:"icon,omitempty"`
}

// StarterKit represents a curated starter kit for common AI patterns
type StarterKit struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Pattern     string      `json:"pattern"` // "rag", "coding_agent", "chatbot", "data_pipeline"
	Packages    []uuid.UUID `json:"packages"`
	Installs    int         `json:"installs"`
	CreatedAt   time.Time   `json:"createdAt"`
}

// RevenueShare represents revenue sharing configuration for a publisher
type RevenueShare struct {
	PublisherID uuid.UUID `json:"publisherId"`
	PackageID   uuid.UUID `json:"packageId"`
	SharePercent float64  `json:"sharePercent"` // publisher's share (e.g., 70%)
	TotalRevenue float64  `json:"totalRevenue"`
	PaidOut      float64  `json:"paidOut"`
}

