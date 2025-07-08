// Package main provides the outbox publisher that polls unpublished events and publishes them to Redis Streams.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/rueidis"

	"github.com/jnst/transactional-outbox-pattern/internal/config"
	"github.com/jnst/transactional-outbox-pattern/internal/repository"
	"github.com/jnst/transactional-outbox-pattern/internal/service"
)

const (
	signalBufferSize = 1
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
		log.Println("Shutdown signal received, stopping publisher...")
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
			log.Println("Publisher stopped")
			return
		case <-ticker.C:
			if err := outboxService.ProcessUnpublishedEvents(ctx, batchSize); err != nil {
				log.Printf("Error processing outbox events: %v", err)
			}
		}
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	dbPool, err := setupDatabase(cfg)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer dbPool.Close()

	redisClient, err := setupPublisherRedisClient(cfg)
	if err != nil {
		log.Printf("failed to connect to Redis: %v", err)
		return
	}
	defer redisClient.Close()

	outboxRepo := repository.NewOutboxRepositoryImpl(dbPool)
	outboxService := service.NewOutboxServiceImpl(outboxRepo, redisClient)

	ctx, cancel := setupPublisherSignalHandling()
	defer cancel()

	log.Printf(
		"Starting outbox publisher (poll interval: %v, batch size: %d)...",
		cfg.PublisherPollInterval,
		cfg.PublisherBatchSize,
	)

	runPublisherLoop(ctx, outboxService, cfg.PublisherPollInterval, cfg.PublisherBatchSize)
}
