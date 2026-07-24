package domain

import (
	"errors"
	"time"
)

// Core business User model (No HTTP/JSON specifics)
type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

// Core business errors
var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUsernameTaken = errors.New("username already taken")
	ErrEmailTaken    = errors.New("email already registered")
)

// Create a blueprint (an interface) of what the database should do
// (e.g., "Give me a user by name").
// We use interfaces to decouple the HTTP handlers from the
// database. This lets us swap out a real database for a mock database
// during testing without changing a single line of the handler's logic:

// UserStore provides access to user data (implemented by both the real DB and testing mocks).
type UserStore interface {
	GetUser(name string) (User, error)
	CreateUser(user *User) error
}
