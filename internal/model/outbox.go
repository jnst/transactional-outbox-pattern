package model

import "time"

// OutboxEvent represents an outbox event for reliable message delivery.
type OutboxEvent struct {
	ID          int64      `json:"id"`
	AggregateID string     `json:"aggregate_id"`
	EventType   string     `json:"event_type"`
	Payload     []byte     `json:"payload"`
	CreatedAt   time.Time  `json:"created_at"`
	PublishedAt *time.Time `json:"published_at"`
}

// CreateOutboxEventParams represents parameters for creating a new outbox event.
type CreateOutboxEventParams struct {
	AggregateID string
	EventType   string
	Payload     []byte
}
