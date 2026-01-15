package main

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// Databases holds all database connections
type Databases struct {
	Postgres    *database.PostgresDB
	ClickHouse  *database.ClickHouseDB
	Redis       *redis.Client
	Minio       *minio.Client
	AsynqClient *asynq.Client
}

// initDatabases initializes all database connections
func initDatabases(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*Databases, error) {
	dbs := &Databases{}

	// Initialize PostgreSQL
	pgDB, err := database.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}
	dbs.Postgres = pgDB

	// Initialize ClickHouse
	chDB, err := database.NewClickHouse(ctx, cfg.ClickHouse)
	if err != nil {
		dbs.Close()
		return nil, fmt.Errorf("failed to initialize ClickHouse: %w", err)
	}
	dbs.ClickHouse = chDB

	// Initialize Redis
	redisClient, err := initRedis(ctx, cfg)
	if err != nil {
		dbs.Close()
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}
	dbs.Redis = redisClient

	// Initialize MinIO (optional)
	minioClient, err := initMinio(cfg)
	if err != nil {
		logger.Warn("failed to initialize MinIO, file storage will be unavailable", zap.Error(err))
	}
	dbs.Minio = minioClient

	// Initialize Asynq client
	dbs.AsynqClient = asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	return dbs, nil
}

// Close closes all database connections
func (d *Databases) Close() {
	if d.Postgres != nil {
		d.Postgres.Close()
	}
	if d.ClickHouse != nil {
		_ = d.ClickHouse.Close()
	}
	if d.Redis != nil {
		d.Redis.Close()
	}
	if d.AsynqClient != nil {
		d.AsynqClient.Close()
	}
}

// initRedis initializes Redis client
func initRedis(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}

// initMinio initializes MinIO client
func initMinio(cfg *config.Config) (*minio.Client, error) {
	if cfg.MinIO.Endpoint == "" {
		return nil, nil
	}

	client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.MinIO.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.MinIO.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return client, nil
}
