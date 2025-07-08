package service

import (
	"context"
	"log"

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
			log.Printf("failed to publish event %d to Redis: %v", event.ID, err)

			continue
		}

		// 発行済みとしてマーク
		if err := s.outboxRepo.MarkAsPublished(ctx, event.ID); err != nil {
			log.Printf("failed to mark event %d as published: %v", event.ID, err)

			continue
		}

		log.Printf("Published event %d to stream %s", event.ID, streamKey)
	}

	return nil
}
