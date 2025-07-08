// Package repository provides data access interfaces and implementations.
package repository

import (
	"context"

	"github.com/jnst/transactional-outbox-pattern/internal/model"
)

// UserRepository defines methods for user data access.
type UserRepository interface {
	Create(ctx context.Context, params *model.CreateUserParams) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

// OutboxRepository defines methods for outbox event data access.
type OutboxRepository interface {
	CreateEvent(ctx context.Context, params *model.CreateOutboxEventParams) (*model.OutboxEvent, error)
	GetUnpublishedEvents(ctx context.Context, limit int) ([]*model.OutboxEvent, error)
	MarkAsPublished(ctx context.Context, id int64) error
}

// TransactionManager defines methods for database transaction management.
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
