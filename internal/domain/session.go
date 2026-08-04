package domain

import "time"

type Session struct {
	ID        string
	UserID    int64
	ExpiresAt time.Time
	CreatedAt time.Time
}

type SessionStore interface {
	CreateSession(userID int64) (*Session, error)
}
