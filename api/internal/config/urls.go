package config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func applyPostgresURL(cfg *PostgresConfig, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return fmt.Errorf("unsupported scheme %q", parsed.Scheme)
	}
	if parsed.Hostname() == "" {
		return fmt.Errorf("host is required")
	}

	cfg.Host = parsed.Hostname()
	if parsed.Port() != "" {
		port, err := strconv.Atoi(parsed.Port())
		if err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
		cfg.Port = port
	}
	if parsed.User != nil {
		cfg.User = parsed.User.Username()
		if password, ok := parsed.User.Password(); ok {
			cfg.Password = password
		}
	}
	if database := strings.TrimPrefix(parsed.Path, "/"); database != "" {
		cfg.Database = database
	}
	if sslMode := parsed.Query().Get("sslmode"); sslMode != "" {
		cfg.SSLMode = sslMode
	}
	return nil
}

func applyClickHouseURL(cfg *ClickHouseConfig, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if parsed.Scheme != "clickhouse" {
		return fmt.Errorf("unsupported scheme %q", parsed.Scheme)
	}
	if parsed.Hostname() == "" {
		return fmt.Errorf("host is required")
	}

	cfg.Host = parsed.Hostname()
	if parsed.Port() != "" {
		port, err := strconv.Atoi(parsed.Port())
		if err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
		cfg.Port = port
	}

	query := parsed.Query()
	if parsed.User != nil {
		cfg.User = parsed.User.Username()
		if password, ok := parsed.User.Password(); ok {
			cfg.Password = password
		}
	}
	if username := query.Get("username"); username != "" {
		cfg.User = username
	}
	if password := query.Get("password"); password != "" {
		cfg.Password = password
	}
	if database := query.Get("database"); database != "" {
		cfg.Database = database
	} else if database := strings.TrimPrefix(parsed.Path, "/"); database != "" {
		cfg.Database = database
	}
	return nil
}

func applyRedisURL(cfg *RedisConfig, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if parsed.Scheme != "redis" {
		return fmt.Errorf("unsupported scheme %q", parsed.Scheme)
	}
	if parsed.Hostname() == "" {
		return fmt.Errorf("host is required")
	}

	cfg.Enabled = true
	cfg.Host = parsed.Hostname()
	if parsed.Port() != "" {
		port, err := strconv.Atoi(parsed.Port())
		if err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
		cfg.Port = port
	}
	if parsed.User != nil {
		if password, ok := parsed.User.Password(); ok {
			cfg.Password = password
		} else if username := parsed.User.Username(); username != "" {
			cfg.Password = username
		}
	}
	if database := strings.TrimPrefix(parsed.Path, "/"); database != "" {
		db, err := strconv.Atoi(database)
		if err != nil {
			return fmt.Errorf("invalid database: %w", err)
		}
		cfg.DB = db
	}
	return nil
}
