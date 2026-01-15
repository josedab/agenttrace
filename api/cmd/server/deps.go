package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/graphql/resolver"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
)

// Dependencies holds all application dependencies
type Dependencies struct {
	Config *config.Config
	Logger *zap.Logger

	// Grouped dependencies
	Databases    *Databases
	Repositories *Repositories
	Services     *Services
	Handlers     *Handlers

	// Middleware
	AuthMiddleware      *middleware.AuthMiddleware
	RateLimitMiddleware *middleware.RateLimitMiddleware
	CSRFMiddleware      *middleware.CSRFMiddleware

	// GraphQL resolver
	Resolver *resolver.Resolver
}

// initDependencies initializes all dependencies
func initDependencies(cfg *config.Config, logger *zap.Logger) (*Dependencies, error) {
	ctx := context.Background()

	// Initialize databases
	dbs, err := initDatabases(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	repos := initRepositories(dbs, logger)

	// Initialize services
	svcs := initServices(cfg, logger, repos)

	// Initialize handlers
	handlers := initHandlers(
		logger,
		svcs,
		repos,
		dbs.Postgres,
		dbs.ClickHouse,
		dbs.Redis,
		dbs.AsynqClient,
		"0.1.0",
	)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(svcs.Auth)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(dbs.Redis)
	csrfMiddleware := middleware.NewCSRFMiddlewareWithConfig(middleware.CSRFConfig{
		Enabled:        cfg.Server.CSRFEnabled,
		CookieSecure:   cfg.Server.SecureCookies,
		CookieSameSite: "Strict",
	})

	// Initialize GraphQL resolver
	gqlResolver := resolver.NewResolver(
		logger,
		svcs.Query,
		svcs.Ingestion,
		svcs.Score,
		svcs.Prompt,
		svcs.Dataset,
		svcs.Eval,
		svcs.Auth,
		svcs.Org,
		svcs.Project,
		svcs.Cost,
	)

	return &Dependencies{
		Config:              cfg,
		Logger:              logger,
		Databases:           dbs,
		Repositories:        repos,
		Services:            svcs,
		Handlers:            handlers,
		AuthMiddleware:      authMiddleware,
		RateLimitMiddleware: rateLimitMiddleware,
		CSRFMiddleware:      csrfMiddleware,
		Resolver:            gqlResolver,
	}, nil
}

// Close closes all dependencies
func (d *Dependencies) Close() {
	if d.Databases != nil {
		d.Databases.Close()
	}
}
