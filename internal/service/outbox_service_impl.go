package service

import (
	"context"
	"log/slog"

	"github.com/redis/rueidis"

	"github.com/jnst/transactional-outbox-pattern/internal/repository"
)

// OutboxServiceImpl implements OutboxService for processing outbox events.
type OutboxServiceImpl struct {
	outboxRepo  repository.OutboxRepository
	redisClient rueidis.Client
}

// NewOutboxServiceImpl creates a new OutboxService implementation.
func NewOutboxServiceImpl(outboxRepo repository.OutboxRepository, redisClient rueidis.Client) OutboxService {
	return &OutboxServiceImpl{
		outboxRepo:  outboxRepo,
		redisClient: redisClient,
	}
}

// ProcessUnpublishedEvents processes unpublished outbox events.
func (s *OutboxServiceImpl) ProcessUnpublishedEvents(ctx context.Context, limit int) error {
	events, err := s.outboxRepo.GetUnpublishedEvents(ctx, limit)
	if err != nil {
		return err
	}

	for _, event := range events {
		// Redis Streamsにイベントを発行
		streamKey := "user:events"
		cmd := s.redisClient.B().Xadd().Key(streamKey).Id("*").
			FieldValue().FieldValue("event_type", event.EventType).
			FieldValue("aggregate_id", event.AggregateID).
			FieldValue("payload", string(event.Payload)).
			Build()

		if err := s.redisClient.Do(ctx, cmd).Error(); err != nil {
			slog.Error("failed to publish event to Redis",
				slog.Int64("event_id", event.ID),
				slog.String("stream", streamKey),
				slog.String("error", err.Error()),
			)

			continue
		}

		// 発行済みとしてマーク
		if err := s.outboxRepo.MarkAsPublished(ctx, event.ID); err != nil {
			slog.Error("failed to mark event as published",
				slog.Int64("event_id", event.ID),
				slog.String("error", err.Error()),
			)

			continue
		}

		slog.Info("event published successfully",
			slog.Int64("event_id", event.ID),
			slog.String("stream", streamKey),
			slog.String("aggregate_id", event.AggregateID),
			slog.String("event_type", event.EventType),
		)
	}

	return nil
}
