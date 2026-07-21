package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/yudan-glitch/twitter-backend/internal/domain"
)

type PostgreSQLUserStore struct {
	queries *Queries
}

func NewPostgreSQLUserStore(db *sql.DB) *PostgreSQLUserStore {
	return &PostgreSQLUserStore{
		queries: New(db),
	}
}

func (s *PostgreSQLUserStore) GetUser(name string) (domain.User, error) {
	// sqlc-generated queries require a context parameter
	ctx := context.Background()

	// Run the actual database query
	dbUser, err := s.queries.GetUser(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, domain.ErrInternalServer
	}

	// 2. Map the database model directly to the pure domain entity
	return domain.User{
		ID:        dbUser.ID,
		Username:  dbUser.Username,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt,
	}, nil
}
