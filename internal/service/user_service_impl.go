package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jnst/transactional-outbox-pattern/internal/model"
	"github.com/jnst/transactional-outbox-pattern/internal/repository"
)

// UserServiceImpl implements UserService for user management business logic.
type UserServiceImpl struct {
	userRepo       repository.UserRepository
	outboxRepo     repository.OutboxRepository
	transactionMgr repository.TransactionManager
}

// NewUserServiceImpl creates a new UserService implementation.
func NewUserServiceImpl(
	userRepo repository.UserRepository,
	outboxRepo repository.OutboxRepository,
	transactionMgr repository.TransactionManager,
) UserService {
	return &UserServiceImpl{
		userRepo:       userRepo,
		outboxRepo:     outboxRepo,
		transactionMgr: transactionMgr,
	}
}

// CreateUser creates a new user and publishes an outbox event.
func (s *UserServiceImpl) CreateUser(ctx context.Context, params *model.CreateUserParams) (*model.User, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	var createdUser *model.User

	err := s.transactionMgr.WithTransaction(ctx, func(ctx context.Context) error {
		user, err := s.userRepo.Create(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		createdUser = user

		return s.createOutboxEvent(ctx, user)
	})

	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

// GetUser retrieves a user by ID.
func (s *UserServiceImpl) GetUser(ctx context.Context, id int64) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (*UserServiceImpl) createUserEventPayload(user *model.User) ([]byte, error) {
	event := model.UserCreatedEvent{
		UserID: user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Action: model.EventActionUserCreated,
	}

	payloadJSON, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event payload: %w", err)
	}

	return payloadJSON, nil
}

func (s *UserServiceImpl) createOutboxEvent(ctx context.Context, user *model.User) error {
	payloadJSON, err := s.createUserEventPayload(user)
	if err != nil {
		return err
	}

	_, err = s.outboxRepo.CreateEvent(ctx, &model.CreateOutboxEventParams{
		AggregateID: fmt.Sprintf("user_%d", user.ID),
		EventType:   string(model.EventActionUserCreated),
		Payload:     payloadJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to create outbox event: %w", err)
	}

	return nil
}
