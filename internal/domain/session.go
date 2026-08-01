package domain

import "time"

type Session struct {
	ID        string
	UserID    int64
	ExpiresAt time.Time
	// need to add createdAt
}

type SessionStore interface {
	CreateSession(userID int64) (*Session, error)
}
