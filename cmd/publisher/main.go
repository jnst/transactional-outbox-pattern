// Package main provides the outbox publisher that polls unpublished events and publishes them to Redis Streams.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/rueidis"

	"github.com/jnst/transactional-outbox-pattern/internal/config"
	"github.com/jnst/transactional-outbox-pattern/internal/logger"
	"github.com/jnst/transactional-outbox-pattern/internal/repository"
	"github.com/jnst/transactional-outbox-pattern/internal/service"
)

const (
	signalBufferSize = 1
	exitCode         = 1
)

func setupDatabase(cfg *config.Config) (*pgxpool.Pool, error) {
	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	return dbPool, nil
}

func setupPublisherRedisClient(cfg *config.Config) (rueidis.Client, error) {
	redisClient, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{cfg.RedisAddr},
	})
	if err != nil {
		return nil, err
	}

	return redisClient, nil
}

func setupPublisherSignalHandling() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, signalBufferSize)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("shutdown signal received, stopping publisher")
		cancel()
	}()

	return ctx, cancel
}

func runPublisherLoop(
	ctx context.Context,
	outboxService service.OutboxService,
	pollInterval time.Duration,
	batchSize int,
) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("publisher stopped")
			return
		case <-ticker.C:
			if err := outboxService.ProcessUnpublishedEvents(ctx, batchSize); err != nil {
				slog.Error("error processing outbox events", slog.String("error", err.Error()))
			}
		}
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(exitCode)
	}

	// ログ設定
	loggerInstance := logger.Setup(cfg.LogLevel)
	slog.SetDefault(loggerInstance)

	dbPool, err := setupDatabase(cfg)
	if err != nil {
		slog.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(exitCode)
	}
	defer dbPool.Close()

	redisClient, err := setupPublisherRedisClient(cfg)
	if err != nil {
		slog.Error("failed to connect to Redis", slog.String("error", err.Error()))
		return
	}
	defer redisClient.Close()

	outboxRepo := repository.NewOutboxRepositoryImpl(dbPool)
	outboxService := service.NewOutboxServiceImpl(outboxRepo, redisClient)

	ctx, cancel := setupPublisherSignalHandling()
	defer cancel()

	slog.Info("starting outbox publisher",
		slog.String("service", "publisher"),
		slog.Duration("poll_interval", cfg.PublisherPollInterval),
		slog.Int("batch_size", cfg.PublisherBatchSize),
	)

	runPublisherLoop(ctx, outboxService, cfg.PublisherPollInterval, cfg.PublisherBatchSize)
}
