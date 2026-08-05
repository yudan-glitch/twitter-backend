package mock

import (
	"errors"
	"time"

	"github.com/solid-state-dan/twitter-backend/internal/domain"
)

type MockSessionStore struct {
	// Add fields here later if I want to "spy" on calls
	FailCreate bool // A flag to trigger an error

	// Optional: allows to customize TTL in tests
	// Also, under the hood, time.Duration is an int64,
	// so its default value is 0
	DefaultTTL time.Duration
}

func (m *MockSessionStore) CreateSession(userID int64) (*domain.Session, error) {

	if m.FailCreate {
		return nil, errors.New("database connection failed")
	}

	ttl := m.DefaultTTL
	if ttl == 0 {
		ttl = 24 * time.Hour // Default for tests
	}

	// For now, return a dummy session so the test can proceed
	return &domain.Session{
		ID:        "mock-session-id",
		UserID:    userID,
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}
