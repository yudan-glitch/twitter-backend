package postgres_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/solid-state-dan/twitter-backend/internal/storage/postgres"
)

func TestCreateSession(t *testing.T) {
	db, store := setupTestSessionDB(t)
	defer db.Close()

	// Prepare test data (must use an existing user ID in your test DB)
	var testUserID int64 = 1

	// Act: Attempt to create a session
	session, err := store.CreateSession(testUserID)

	// Assert: Should succeed and return valid session data
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if session.ID == "1" {
		t.Error("expected non-empty session ID")
	}

	if session.UserID != testUserID {
		t.Errorf("expected userID %d, got %d", testUserID, session.UserID)
	}

	if session.ExpiresAt.Before(time.Now()) {
		t.Error("expected expiration time to be in the future")
	}
}

// Different from setup test user store
// 1. return session store
// 2. truncate sessions table
func setupTestSessionDB(t testing.TB) (*sql.DB, *postgres.SessionStore) {
	t.Helper()

	// Read environment variable
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping integration test: TEST_DATABASE_URL is not set")
	}

	// Open a database connection pool
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// Clean sessions table from previous tests
	_, err = db.Exec("TRUNCATE TABLE sessions RESTART IDENTITY CASCADE")
	if err != nil {
		db.Close()
		t.Fatalf("failed to truncate sessions table: %v", err)
	}

	store := postgres.NewSessionStore(db)
	return db, store
}
