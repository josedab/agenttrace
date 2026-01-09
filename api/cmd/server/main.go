package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/graphql/generated"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	var logger *zap.Logger
	if cfg.Server.Env == "production" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	// Initialize dependencies
	deps, err := initDependencies(cfg, logger)
	if err != nil {
		logger.Fatal("failed to initialize dependencies", zap.Error(err))
	}
	defer deps.Close()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               "AgentTrace API",
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           120 * time.Second,
		DisableStartupMessage: cfg.Server.Env == "production",
		ErrorHandler:          errorHandler(logger),
	})

	// Apply global middleware
	app.Use(middleware.RequestID())

	loggerMiddleware := middleware.NewLoggerMiddleware(middleware.DefaultLoggerConfig(logger))
	app.Use(loggerMiddleware.Handler())

	recoverMiddleware := middleware.NewRecoverMiddleware(middleware.DefaultRecoverConfig(logger))
	app.Use(recoverMiddleware.Handler())

	corsMiddleware := middleware.NewCORSMiddleware(middleware.DefaultCORSConfig())
	app.Use(corsMiddleware.Handler())

	// Metrics middleware
	metricsMiddleware := middleware.NewMetricsMiddleware(middleware.DefaultMetricsConfig())
	app.Use(metricsMiddleware.Handler())

	// Register routes
	registerRoutes(app, deps)

	// Setup GraphQL
	setupGraphQL(app, deps, cfg)

	// Start server
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		logger.Info("starting server", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("server shutdown error", zap.Error(err))
	}

	logger.Info("server stopped")
}

// setupGraphQL sets up GraphQL handlers
func setupGraphQL(app *fiber.App, deps *Dependencies, cfg *config.Config) {
	// Create GraphQL server
	srv := handler.NewDefaultServer(
		generated.NewExecutableSchema(generated.Config{
			Resolvers: deps.Resolver,
		}),
	)

	// GraphQL endpoint
	app.All("/graphql", adaptor.HTTPHandler(srv))

	// GraphQL playground (only in development)
	if cfg.Server.Env != "production" {
		app.Get("/playground", adaptor.HTTPHandlerFunc(playground.Handler("AgentTrace GraphQL", "/graphql")))
	}
}

// errorHandler creates a custom error handler
func errorHandler(logger *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Default to 500 Internal Server Error
		code := fiber.StatusInternalServerError
		message := "Internal Server Error"

		// Check if it's a Fiber error
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			message = e.Message
		}

		// Log error
		logger.Error("request error",
			zap.Int("status", code),
			zap.String("error", err.Error()),
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
		)

		// Return JSON error response
		return c.Status(code).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    code,
				"message": message,
			},
		})
	}
}
