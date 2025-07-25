// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type OutboxEvent struct {
	ID          int64            `json:"id"`
	AggregateID string           `json:"aggregateId"`
	EventType   string           `json:"eventType"`
	Payload     []byte           `json:"payload"`
	CreatedAt   pgtype.Timestamp `json:"createdAt"`
	PublishedAt pgtype.Timestamp `json:"publishedAt"`
}

type User struct {
	ID        int64            `json:"id"`
	Email     string           `json:"email"`
	Name      string           `json:"name"`
	CreatedAt pgtype.Timestamp `json:"createdAt"`
}
