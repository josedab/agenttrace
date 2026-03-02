package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// MarketplaceService manages the agent marketplace
type MarketplaceService struct {
	logger   *zap.Logger
	mu       sync.RWMutex
	packages map[uuid.UUID]*domain.MarketplacePackage
	ratings  []domain.PackageRating
}

// NewMarketplaceService creates a new marketplace service with pre-seeded packages
func NewMarketplaceService(logger *zap.Logger) *MarketplaceService {
	svc := &MarketplaceService{
		logger:   logger,
		packages: make(map[uuid.UUID]*domain.MarketplacePackage),
		ratings:  []domain.PackageRating{},
	}
	svc.seedPackages()
	return svc
}

func (s *MarketplaceService) seedPackages() {
	seeds := []domain.MarketplacePackage{
		{
			ID:          uuid.New(),
			Name:        "Safety Guardrail Suite",
			Description: "Comprehensive guardrails for LLM output safety including toxicity, PII detection, and hallucination checks.",
			Type:        domain.PackageGuardrail,
			Version:     "1.2.0",
			Author:      "agenttrace-team",
			Tags:        []string{"safety", "guardrail", "pii", "toxicity"},
			Downloads:   1842,
			Rating:      4.7,
			RatingCount: 124,
			IsPublic:    true,
			Content:     "rules:\n  - name: toxicity_check\n    threshold: 0.8\n  - name: pii_detection\n    enabled: true\n  - name: hallucination_guard\n    confidence: 0.9\n",
			CreatedAt:   time.Now().Add(-90 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-7 * 24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Name:        "RAG Quality Evaluator",
			Description: "Evaluate retrieval-augmented generation quality with relevance, faithfulness, and context precision metrics.",
			Type:        domain.PackageEvaluator,
			Version:     "2.0.1",
			Author:      "eval-labs",
			Tags:        []string{"rag", "evaluation", "quality", "retrieval"},
			Downloads:   3201,
			Rating:      4.9,
			RatingCount: 256,
			IsPublic:    true,
			Content:     "evaluators:\n  - name: relevance\n    weight: 0.4\n  - name: faithfulness\n    weight: 0.35\n  - name: context_precision\n    weight: 0.25\n",
			CreatedAt:   time.Now().Add(-120 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-3 * 24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Name:        "Chain-of-Thought Prompt Pack",
			Description: "Curated prompt templates for chain-of-thought reasoning across coding, math, and analysis tasks.",
			Type:        domain.PackagePrompt,
			Version:     "1.0.0",
			Author:      "prompt-engineers",
			Tags:        []string{"prompt", "cot", "reasoning", "coding"},
			Downloads:   987,
			Rating:      4.5,
			RatingCount: 78,
			IsPublic:    true,
			Content:     "prompts:\n  - name: code_review_cot\n    template: \"Think step by step...\"\n  - name: math_reasoning\n    template: \"Let's solve this systematically...\"\n",
			CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-14 * 24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Name:        "Agent Benchmark Suite",
			Description: "Standard benchmarks for evaluating coding agents including SWE-bench, HumanEval, and custom task suites.",
			Type:        domain.PackageBenchmark,
			Version:     "3.1.0",
			Author:      "bench-team",
			Tags:        []string{"benchmark", "swe-bench", "humaneval", "coding"},
			Downloads:   2456,
			Rating:      4.8,
			RatingCount: 189,
			IsPublic:    true,
			Content:     "benchmarks:\n  - name: swe_bench_lite\n    tasks: 300\n  - name: humaneval_plus\n    tasks: 164\n  - name: custom_agent_tasks\n    tasks: 50\n",
			CreatedAt:   time.Now().Add(-150 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * 24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Name:        "Production Agent Bundle",
			Description: "All-in-one bundle with guardrails, evaluators, and monitoring for production agent deployments.",
			Type:        domain.PackageBundle,
			Version:     "1.5.0",
			Author:      "agenttrace-team",
			Tags:        []string{"bundle", "production", "monitoring", "guardrails"},
			Downloads:   1567,
			Rating:      4.6,
			RatingCount: 95,
			IsPublic:    true,
			Content:     "bundle:\n  includes:\n    - safety-guardrail-suite@1.2.0\n    - rag-quality-evaluator@2.0.1\n  monitoring:\n    alerts: true\n    dashboards: true\n",
			CreatedAt:   time.Now().Add(-45 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-5 * 24 * time.Hour),
		},
	}

	for i := range seeds {
		s.packages[seeds[i].ID] = &seeds[i]
	}
}

// PublishPackage publishes a new package to the marketplace
func (s *MarketplaceService) PublishPackage(ctx context.Context, input *domain.PackagePublishInput) (*domain.MarketplacePackage, error) {
	now := time.Now()
	version := input.Version
	if version == "" {
		version = "1.0.0"
	}

	pkg := &domain.MarketplacePackage{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
		Type:        input.Type,
		Version:     version,
		Author:      "current-user",
		Tags:        input.Tags,
		Downloads:   0,
		Rating:      0,
		RatingCount: 0,
		IsPublic:    input.IsPublic,
		Content:     input.Content,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.mu.Lock()
	s.packages[pkg.ID] = pkg
	s.mu.Unlock()

	s.logger.Info("published marketplace package",
		zap.String("packageId", pkg.ID.String()),
		zap.String("name", pkg.Name),
		zap.String("type", string(pkg.Type)),
	)

	return pkg, nil
}

// SearchPackages searches and filters marketplace packages
func (s *MarketplaceService) SearchPackages(ctx context.Context, search *domain.MarketplaceSearch) ([]domain.MarketplacePackage, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []domain.MarketplacePackage
	for _, pkg := range s.packages {
		if !pkg.IsPublic {
			continue
		}
		if search.Type != nil && pkg.Type != *search.Type {
			continue
		}
		if search.Query != "" {
			query := strings.ToLower(search.Query)
			if !strings.Contains(strings.ToLower(pkg.Name), query) &&
				!strings.Contains(strings.ToLower(pkg.Description), query) {
				continue
			}
		}
		if len(search.Tags) > 0 {
			matched := false
			for _, searchTag := range search.Tags {
				for _, pkgTag := range pkg.Tags {
					if strings.EqualFold(searchTag, pkgTag) {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				continue
			}
		}
		results = append(results, *pkg)
	}

	// Sort results
	switch search.SortBy {
	case "downloads":
		sort.Slice(results, func(i, j int) bool { return results[i].Downloads > results[j].Downloads })
	case "rating":
		sort.Slice(results, func(i, j int) bool { return results[i].Rating > results[j].Rating })
	case "newest":
		sort.Slice(results, func(i, j int) bool { return results[i].CreatedAt.After(results[j].CreatedAt) })
	default:
		sort.Slice(results, func(i, j int) bool { return results[i].Downloads > results[j].Downloads })
	}

	total := len(results)

	// Apply pagination
	limit := search.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := search.Offset
	if offset > len(results) {
		return []domain.MarketplacePackage{}, total
	}
	end := offset + limit
	if end > len(results) {
		end = len(results)
	}
	results = results[offset:end]

	return results, total
}

// GetPackage returns a single package by ID
func (s *MarketplaceService) GetPackage(ctx context.Context, packageID uuid.UUID) (*domain.MarketplacePackage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pkg, exists := s.packages[packageID]
	if !exists {
		return nil, fmt.Errorf("package not found")
	}
	return pkg, nil
}

// InstallPackage increments the download count for a package
func (s *MarketplaceService) InstallPackage(ctx context.Context, packageID uuid.UUID) (*domain.MarketplacePackage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pkg, exists := s.packages[packageID]
	if !exists {
		return nil, fmt.Errorf("package not found")
	}

	pkg.Downloads++
	pkg.UpdatedAt = time.Now()

	s.logger.Info("installed marketplace package",
		zap.String("packageId", packageID.String()),
		zap.String("name", pkg.Name),
		zap.Int("downloads", pkg.Downloads),
	)

	return pkg, nil
}

// RatePackage adds a rating to a package
func (s *MarketplaceService) RatePackage(ctx context.Context, packageID uuid.UUID, score int, review string) (*domain.MarketplacePackage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pkg, exists := s.packages[packageID]
	if !exists {
		return nil, fmt.Errorf("package not found")
	}

	if score < 1 || score > 5 {
		return nil, fmt.Errorf("score must be between 1 and 5")
	}

	rating := domain.PackageRating{
		PackageID: packageID,
		UserID:    uuid.New(),
		Score:     score,
		Review:    review,
		CreatedAt: time.Now(),
	}
	s.ratings = append(s.ratings, rating)

	// Recalculate average rating
	totalScore := pkg.Rating * float64(pkg.RatingCount)
	pkg.RatingCount++
	pkg.Rating = (totalScore + float64(score)) / float64(pkg.RatingCount)
	pkg.UpdatedAt = time.Now()

	s.logger.Info("rated marketplace package",
		zap.String("packageId", packageID.String()),
		zap.Int("score", score),
		zap.Float64("newRating", pkg.Rating),
	)

	return pkg, nil
}

// GetFeatured returns featured marketplace packages
func (s *MarketplaceService) GetFeatured(ctx context.Context) []domain.MarketplacePackage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var all []domain.MarketplacePackage
	for _, pkg := range s.packages {
		if pkg.IsPublic {
			all = append(all, *pkg)
		}
	}

	// Sort by a combination of downloads and rating
	sort.Slice(all, func(i, j int) bool {
		scoreI := float64(all[i].Downloads)*0.3 + all[i].Rating*100
		scoreJ := float64(all[j].Downloads)*0.3 + all[j].Rating*100
		return scoreI > scoreJ
	})

	if len(all) > 5 {
		all = all[:5]
	}

	return all
}

// GetCategories returns marketplace categories with package counts
func (s *MarketplaceService) GetCategories(ctx context.Context) []domain.MarketplaceCategory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := make(map[domain.PackageType]int)
	for _, pkg := range s.packages {
		if pkg.IsPublic {
			counts[pkg.Type]++
		}
	}

	return []domain.MarketplaceCategory{
		{Name: "Prompts", Description: "Prompt templates and chains", Count: counts[domain.PackagePrompt], Icon: "message-square"},
		{Name: "Guardrails", Description: "Safety and compliance guardrails", Count: counts[domain.PackageGuardrail], Icon: "shield"},
		{Name: "Evaluators", Description: "Quality evaluation templates", Count: counts[domain.PackageEvaluator], Icon: "check-circle"},
		{Name: "Benchmarks", Description: "Agent benchmark suites", Count: counts[domain.PackageBenchmark], Icon: "bar-chart"},
		{Name: "Bundles", Description: "All-in-one packages", Count: counts[domain.PackageBundle], Icon: "package"},
	}
}

// GetStarterKits returns curated starter kits
func (s *MarketplaceService) GetStarterKits(ctx context.Context) []domain.StarterKit {
	return []domain.StarterKit{
		{
			ID:          uuid.New(),
			Name:        "RAG Agent Starter",
			Description: "Everything you need to monitor and evaluate a retrieval-augmented generation agent",
			Pattern:     "rag",
			Installs:    342,
			CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Name:        "Coding Agent Kit",
			Description: "Guardrails, evaluators, and benchmarks for coding assistants",
			Pattern:     "coding_agent",
			Installs:    567,
			CreatedAt:   time.Now().Add(-45 * 24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Name:        "Chatbot Production Kit",
			Description: "Production-ready monitoring, safety, and quality for customer-facing chatbots",
			Pattern:     "chatbot",
			Installs:    891,
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Name:        "Data Pipeline Agent Kit",
			Description: "Observability and cost optimization for data processing AI agents",
			Pattern:     "data_pipeline",
			Installs:    234,
			CreatedAt:   time.Now().Add(-20 * 24 * time.Hour),
		},
	}
}

// GetReviews returns ratings and reviews for a package
func (s *MarketplaceService) GetReviews(ctx context.Context, packageID uuid.UUID) []domain.PackageRating {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var reviews []domain.PackageRating
	for _, r := range s.ratings {
		if r.PackageID == packageID {
			reviews = append(reviews, r)
		}
	}
	return reviews
}
