package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/pkg/logger"
)

// PostgresDB wraps a PostgreSQL connection pool
type PostgresDB struct {
	Pool *pgxpool.Pool
}

// NewPostgres creates a new PostgreSQL connection pool
func NewPostgres(ctx context.Context, cfg config.PostgresConfig) (*PostgresDB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	// Add logging for slow queries in development
	poolConfig.ConnConfig.Tracer = &queryTracer{}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	logger.Info("connected to PostgreSQL",
		zap.String("host", cfg.Host),
		zap.String("database", cfg.Database),
		zap.Int32("max_conns", cfg.MaxConns),
	)

	return &PostgresDB{Pool: pool}, nil
}

// Close closes the connection pool
func (db *PostgresDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// BeginTx starts a new transaction
func (db *PostgresDB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return db.Pool.Begin(ctx)
}

// BeginTxWithOptions starts a new transaction with options
func (db *PostgresDB) BeginTxWithOptions(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return db.Pool.BeginTx(ctx, opts)
}

// queryTracer implements pgx.QueryTracer for logging
type queryTracer struct{}

type queryStartKey struct{}
type querySQLKey struct{}

func (t *queryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx = context.WithValue(ctx, queryStartKey{}, time.Now())
	ctx = context.WithValue(ctx, querySQLKey{}, data.SQL)
	return ctx
}

func (t *queryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	start, ok := ctx.Value(queryStartKey{}).(time.Time)
	if !ok {
		return
	}

	duration := time.Since(start)

	// Log slow queries (> 100ms)
	if duration > 100*time.Millisecond {
		sql, _ := ctx.Value(querySQLKey{}).(string)
		logger.Warn("slow query detected",
			zap.Int64("duration_ms", duration.Milliseconds()),
			zap.String("sql", truncateSQL(sql, 200)),
		)
	}
}

func truncateSQL(sql string, maxLen int) string {
	if len(sql) <= maxLen {
		return sql
	}
	return sql[:maxLen] + "..."
}

// Transaction executes a function within a transaction
func Transaction(ctx context.Context, db *PostgresDB, fn func(tx pgx.Tx) error) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			logger.Error("failed to rollback transaction",
				zap.Error(rbErr),
			)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
