package domain

import (
	"errors"
	"time"
)

// Core business User model (No HTTP/JSON specifics)
type User struct {
	ID        int64
	Username  string
	Email     string
	CreatedAt time.Time
}

// Core business errors
var (
	ErrUserNotFound   = errors.New("user not found")
	ErrInternalServer = errors.New("internal server error")
)

// Create a blueprint (an interface) of what the database should do
// (e.g., "Give me a user by name").
// We use interfaces to decouple the HTTP handlers from the
// database. This lets us swap out a real database for a mock database
// during testing without changing a single line of the handler's logic:

// UserStore provides access to user data (implemented by both the real DB and testing mocks).
type UserStore interface {
	GetUser(name string) (User, error)
}
