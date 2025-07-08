package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jnst/transactional-outbox-pattern/internal/db"
	"github.com/jnst/transactional-outbox-pattern/internal/model"
)

// OutboxRepositoryImpl implements OutboxRepository using PostgreSQL.
type OutboxRepositoryImpl struct {
	db   *db.Queries
	pool *pgxpool.Pool
}

// NewOutboxRepositoryImpl creates a new OutboxRepository implementation.
func NewOutboxRepositoryImpl(pool *pgxpool.Pool) OutboxRepository {
	return &OutboxRepositoryImpl{
		db:   db.New(pool),
		pool: pool,
	}
}

// CreateEvent creates a new outbox event.
func (r *OutboxRepositoryImpl) CreateEvent(
	ctx context.Context, params *model.CreateOutboxEventParams,
) (*model.OutboxEvent, error) {
	dbEvent, err := r.db.CreateOutboxEvent(ctx, &db.CreateOutboxEventParams{
		AggregateID: params.AggregateID,
		EventType:   params.EventType,
		Payload:     params.Payload,
	})
	if err != nil {
		return nil, err
	}

	var publishedAt *time.Time
	if dbEvent.PublishedAt.Valid {
		publishedAt = &dbEvent.PublishedAt.Time
	}

	return &model.OutboxEvent{
		ID:          dbEvent.ID,
		AggregateID: dbEvent.AggregateID,
		EventType:   dbEvent.EventType,
		Payload:     dbEvent.Payload,
		CreatedAt:   dbEvent.CreatedAt.Time,
		PublishedAt: publishedAt,
	}, nil
}

// GetUnpublishedEvents retrieves unpublished outbox events.
func (r *OutboxRepositoryImpl) GetUnpublishedEvents(ctx context.Context, limit int) ([]*model.OutboxEvent, error) {
	dbEvents, err := r.db.GetUnpublishedEvents(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	events := make([]*model.OutboxEvent, len(dbEvents))

	for i, dbEvent := range dbEvents {
		var publishedAt *time.Time
		if dbEvent.PublishedAt.Valid {
			publishedAt = &dbEvent.PublishedAt.Time
		}

		events[i] = &model.OutboxEvent{
			ID:          dbEvent.ID,
			AggregateID: dbEvent.AggregateID,
			EventType:   dbEvent.EventType,
			Payload:     dbEvent.Payload,
			CreatedAt:   dbEvent.CreatedAt.Time,
			PublishedAt: publishedAt,
		}
	}

	return events, nil
}

// MarkAsPublished marks an outbox event as published.
func (r *OutboxRepositoryImpl) MarkAsPublished(ctx context.Context, id int64) error {
	return r.db.MarkEventAsPublished(ctx, id)
}
