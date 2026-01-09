package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/pkg/logger"
)

// ClickHouseDB wraps a ClickHouse connection
type ClickHouseDB struct {
	Conn driver.Conn
}

// NewClickHouse creates a new ClickHouse connection
func NewClickHouse(ctx context.Context, cfg config.ClickHouseConfig) (*ClickHouseDB, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.User,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:          10 * time.Second,
		MaxOpenConns:         25,
		MaxIdleConns:         5,
		ConnMaxLifetime:      time.Hour,
		ConnOpenStrategy:     clickhouse.ConnOpenInOrder,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10 * 1024 * 1024, // 10MB
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	// Test connection
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	logger.Info("connected to ClickHouse",
		zap.String("host", cfg.Host),
		zap.String("database", cfg.Database),
	)

	return &ClickHouseDB{Conn: conn}, nil
}

// NewClickHouseWithTLS creates a new ClickHouse connection with TLS
func NewClickHouseWithTLS(ctx context.Context, cfg config.ClickHouseConfig, tlsConfig *tls.Config) (*ClickHouseDB, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.User,
			Password: cfg.Password,
		},
		TLS: tlsConfig,
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:      10 * time.Second,
		MaxOpenConns:     25,
		MaxIdleConns:     5,
		ConnMaxLifetime:  time.Hour,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection with TLS: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	return &ClickHouseDB{Conn: conn}, nil
}

// Close closes the connection
func (db *ClickHouseDB) Close() error {
	if db.Conn != nil {
		return db.Conn.Close()
	}
	return nil
}

// PrepareBatch prepares a batch insert
func (db *ClickHouseDB) PrepareBatch(ctx context.Context, query string) (driver.Batch, error) {
	return db.Conn.PrepareBatch(ctx, query)
}

// Exec executes a query
func (db *ClickHouseDB) Exec(ctx context.Context, query string, args ...interface{}) error {
	return db.Conn.Exec(ctx, query, args...)
}

// Select executes a select query and scans results into dest
func (db *ClickHouseDB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return db.Conn.Select(ctx, dest, query, args...)
}

// QueryRow executes a query that returns a single row
func (db *ClickHouseDB) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	return db.Conn.QueryRow(ctx, query, args...)
}

// Query executes a query
func (db *ClickHouseDB) Query(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	return db.Conn.Query(ctx, query, args...)
}

// AsyncInsert performs an asynchronous insert
func (db *ClickHouseDB) AsyncInsert(ctx context.Context, query string, wait bool, args ...interface{}) error {
	return db.Conn.AsyncInsert(ctx, query, wait, args...)
}

// BatchInsertTraces performs batch insert of traces
func (db *ClickHouseDB) BatchInsertTraces(ctx context.Context, traces []map[string]interface{}) error {
	if len(traces) == 0 {
		return nil
	}

	batch, err := db.PrepareBatch(ctx, `
		INSERT INTO traces (
			id, project_id, name, user_id, session_id, release, version,
			tags, metadata, public, bookmarked, start_time, end_time,
			input, output, level, status_message, git_commit_sha,
			git_branch, git_repo_url, created_at, updated_at
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, trace := range traces {
		if err := batch.Append(
			trace["id"],
			trace["project_id"],
			trace["name"],
			trace["user_id"],
			trace["session_id"],
			trace["release"],
			trace["version"],
			trace["tags"],
			trace["metadata"],
			trace["public"],
			trace["bookmarked"],
			trace["start_time"],
			trace["end_time"],
			trace["input"],
			trace["output"],
			trace["level"],
			trace["status_message"],
			trace["git_commit_sha"],
			trace["git_branch"],
			trace["git_repo_url"],
			trace["created_at"],
			trace["updated_at"],
		); err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}
