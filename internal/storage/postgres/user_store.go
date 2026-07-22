package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
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
		ID:           dbUser.ID,
		Username:     dbUser.Username,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
		CreatedAt:    dbUser.CreatedAt,
	}, nil
}

func (s *PostgreSQLUserStore) CreateUser(user *domain.User) error {
	ctx := context.Background()

	arg := CreateUserParams{
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}
	dbUser, err := s.queries.CreateUser(ctx, arg)
	if err != nil {
		// Catch driver-specific error constraints
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == "23505" {
			if strings.Contains(pqErr.Message, "users_username_key") {
				return domain.ErrUsernameTaken
			}
			if strings.Contains(pqErr.Message, "users_email_key") {
				return domain.ErrEmailTaken
			}
		}
		return err
	}

	user.ID = dbUser.ID

	// Why map CreatedAt if the database will set it up automatically (because of
	// DEFAULT NOW() in the migration file)?
	// If the database creates the timestamp correctly on disk, why do we need to
	// write user.CreatedAt = dbUser. CreatedAt inside our creation function? We do
	// it as a convenience for the rest of the application. If an HTTP handler
	// creates a user, it often wants to return the full, completed user object
	// to the client immediately—including the fresh ID and the real registration
	// timestamp—without being forced to run a slow, second database query (GetUser)
	// right after the insert.Explicitly copying dbUser. CreatedAt onto the user
	// pointer makes sure the in-memory Go object matches the on-disk database
	// row perfectly before the function finishes.
	user.CreatedAt = dbUser.CreatedAt
	return nil
}
