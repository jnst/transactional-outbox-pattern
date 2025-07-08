// Package model defines domain models and data structures.
package model

import "time"

// User represents a user entity.
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUserParams represents parameters for creating a new user.
type CreateUserParams struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Validate validates the create user parameters.
func (p *CreateUserParams) Validate() error {
	if p.Name == "" {
		return ErrInvalidName
	}

	if p.Email == "" {
		return ErrInvalidEmail
	}

	return nil
}

// EventAction represents the type of event action.
type EventAction string

const (
	// EventActionUserCreated represents the user creation event action.
	EventActionUserCreated EventAction = "user_created"
)

// UserCreatedEvent represents the payload for user creation events.
type UserCreatedEvent struct {
	UserID int64       `json:"user_id"`
	Name   string      `json:"name"`
	Email  string      `json:"email"`
	Action EventAction `json:"action"`
}
