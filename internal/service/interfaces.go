// Package service provides business logic layer implementations.
package service

import (
	"context"

	"github.com/jnst/transactional-outbox-pattern/internal/model"
)

// UserService defines business logic methods for user management.
type UserService interface {
	CreateUser(ctx context.Context, params *model.CreateUserParams) (*model.User, error)
	GetUser(ctx context.Context, id int64) (*model.User, error)
}

// OutboxService defines business logic methods for outbox event processing.
type OutboxService interface {
	ProcessUnpublishedEvents(ctx context.Context, limit int) error
}
