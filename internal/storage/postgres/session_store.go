package postgres

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/solid-state-dan/twitter-backend/internal/domain"
)

type SessionStore struct {
	db *sql.DB
}

func NewSessionStore(db *sql.DB) *SessionStore {
	return &SessionStore{db: db}
}

// CreateSession generates a new session and persists it in PostgreSQL.
func (s *SessionStore) CreateSession(userID int64) (*domain.Session, error) {
	// 1. Generate a cryptographically secure random token (32 bytes = 64 hex characters)
	sessionID, err := generateToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// 2. Set default session duration (24 hours)
	createdAt := time.Now()
	expiresAt := createdAt.Add(24 * time.Hour)

	// 3. Insert session into PostgreSQL
	query := `
		INSERT INTO session (id, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = s.db.Exec(query, sessionID, userID, expiresAt, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert session into database: %w", err)
	}

	// 4. Return the newly created session
	return &domain.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
	}, nil
}

// generateToken creates a secure hex string of the specified byte length.
func generateToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
