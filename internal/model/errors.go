package model

import "errors"

var (
	// ErrInvalidName is returned when user name is empty or invalid.
	ErrInvalidName = errors.New("name is required")
	// ErrInvalidEmail is returned when user email is empty or invalid.
	ErrInvalidEmail = errors.New("email is required")
	// ErrUserNotFound is returned when user is not found in database.
	ErrUserNotFound = errors.New("user not found")
)
