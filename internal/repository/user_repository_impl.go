package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jnst/transactional-outbox-pattern/internal/db"
	"github.com/jnst/transactional-outbox-pattern/internal/model"
)

// UserRepositoryImpl implements UserRepository using PostgreSQL.
type UserRepositoryImpl struct {
	db   *db.Queries
	pool *pgxpool.Pool
}

// NewUserRepositoryImpl creates a new UserRepository implementation.
func NewUserRepositoryImpl(pool *pgxpool.Pool) UserRepository {
	return &UserRepositoryImpl{
		db:   db.New(pool),
		pool: pool,
	}
}

// Create creates a new user.
func (r *UserRepositoryImpl) Create(ctx context.Context, params *model.CreateUserParams) (*model.User, error) {
	dbUser, err := r.db.CreateUser(ctx, &db.CreateUserParams{
		Name:  params.Name,
		Email: params.Email,
	})
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:        dbUser.ID,
		Name:      dbUser.Name,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt.Time,
	}, nil
}

// GetByID retrieves a user by ID.
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id int64) (*model.User, error) {
	dbUser, err := r.db.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:        dbUser.ID,
		Name:      dbUser.Name,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt.Time,
	}, nil
}

// GetByEmail retrieves a user by email.
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	dbUser, err := r.db.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:        dbUser.ID,
		Name:      dbUser.Name,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt.Time,
	}, nil
}
