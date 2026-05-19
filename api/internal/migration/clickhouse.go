// Package migration runs and inspects AgentTrace database migrations.
package migration

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// ClickHouseSchemaState describes whether migration history is present.
type ClickHouseSchemaState string

const (
	// ClickHouseSchemaEmpty indicates no managed tables or history exist.
	ClickHouseSchemaEmpty ClickHouseSchemaState = "empty"
	// ClickHouseSchemaLegacy indicates all legacy tables exist without history.
	ClickHouseSchemaLegacy ClickHouseSchemaState = "legacy"
	// ClickHouseSchemaManaged indicates schema_migrations is present.
	ClickHouseSchemaManaged ClickHouseSchemaState = "managed"
)

// NormalizeClickHouseURL converts supported DSN forms to golang-migrate format.
func NormalizeClickHouseURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "clickhouse" {
		return "", fmt.Errorf("unsupported scheme %q; use clickhouse://", parsed.Scheme)
	}

	query := parsed.Query()
	host := parsed.Hostname()
	if host == "" {
		return "", fmt.Errorf("host is required")
	}

	secure := query.Get("secure") == "true" || query.Get("secure") == "1"
	port := parsed.Port()
	if port == "" {
		if secure {
			port = "9440"
		} else {
			port = "9000"
		}
	}

	username := query.Get("username")
	password := query.Get("password")
	if parsed.User != nil {
		if username == "" {
			username = parsed.User.Username()
		}
		if password == "" {
			if value, ok := parsed.User.Password(); ok {
				password = value
			}
		}
	}
	if username == "" {
		username = "default"
	}

	database := query.Get("database")
	if database == "" {
		database = strings.TrimPrefix(parsed.Path, "/")
	}
	if database == "" {
		database = "default"
	}

	query.Set("username", username)
	query.Set("password", password)
	query.Set("database", database)
	query.Set("x-multi-statement", "true")

	normalized := &url.URL{
		Scheme:   "clickhouse",
		Host:     net.JoinHostPort(host, port),
		RawQuery: query.Encode(),
	}
	return normalized.String(), nil
}

// InspectClickHouseSchema checks whether a database is empty, legacy, or managed.
func InspectClickHouseSchema(ctx context.Context, rawURL string) (state ClickHouseSchemaState, err error) {
	normalizedURL, err := NormalizeClickHouseURL(rawURL)
	if err != nil {
		return "", err
	}
	options, err := clickHouseOptions(normalizedURL)
	if err != nil {
		return "", err
	}

	conn, err := clickhouse.Open(options)
	if err != nil {
		return "", fmt.Errorf("open ClickHouse connection: %w", err)
	}
	defer func() {
		err = errors.Join(err, conn.Close())
	}()

	hasHistory, err := tableExists(ctx, conn, "schema_migrations")
	if err != nil {
		return "", fmt.Errorf("check migration history: %w", err)
	}
	if hasHistory {
		return ClickHouseSchemaManaged, nil
	}

	hasFirst, err := tableExists(ctx, conn, "traces")
	if err != nil {
		return "", fmt.Errorf("check traces table: %w", err)
	}
	hasLast, err := tableExists(ctx, conn, "ci_runs")
	if err != nil {
		return "", fmt.Errorf("check ci_runs table: %w", err)
	}

	switch {
	case hasFirst && hasLast:
		return ClickHouseSchemaLegacy, nil
	case !hasFirst && !hasLast:
		return ClickHouseSchemaEmpty, nil
	default:
		return "", fmt.Errorf("partial legacy ClickHouse schema detected; repair it before running migrations")
	}
}

func clickHouseOptions(rawURL string) (*clickhouse.Options, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "clickhouse" {
		return nil, fmt.Errorf("unsupported scheme %q", parsed.Scheme)
	}

	query := parsed.Query()
	host := parsed.Hostname()
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}

	secure := query.Get("secure") == "true" || query.Get("secure") == "1"
	port := parsed.Port()
	if port == "" {
		if secure {
			port = "9440"
		} else {
			port = "9000"
		}
	}

	options := &clickhouse.Options{
		Addr: []string{net.JoinHostPort(host, port)},
		Auth: clickhouse.Auth{
			Database: query.Get("database"),
			Username: query.Get("username"),
			Password: query.Get("password"),
		},
		DialTimeout: 10 * time.Second,
	}
	if secure {
		options.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	return options, nil
}

func tableExists(ctx context.Context, conn clickhouse.Conn, table string) (bool, error) {
	var exists bool
	err := conn.QueryRow(
		ctx,
		"SELECT count() > 0 FROM system.tables WHERE database = currentDatabase() AND name = ?",
		table,
	).Scan(&exists)
	return exists, err
}
