package postgres_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/yudan-glitch/twitter-backend/internal/domain"
	"github.com/yudan-glitch/twitter-backend/internal/storage/postgres"
)

func TestPostgreSQLUserStore_GetUser(t *testing.T) {

	// Write just enough test code to establish a database connection
	// and create the store instance. Will not query anything yet.

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
	// Good practice to close it once done
	defer db.Close()

	store := postgres.NewPostgreSQLUserStore(db)

	tests := []struct {
		name          string
		username      string
		expectedEmail string
		expectedErr   error
	}{
		{
			name:        "Request non-existing user",
			username:    "unknown",
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name:          "Request existing user",
			username:      "alice",
			expectedEmail: "alice@example.com",
			expectedErr:   nil,
		},
	}

	// Clean the table first to guarantee isolation
	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("failed to clear table: %v", err)
	}

	// Seed a real row into PostgreSQL
	// log.Println("--- DEBUG: About to insert Alice ---")
	_, err = db.Exec("INSERT INTO users (username, email) VALUES ($1, $2)", "alice", "alice@example.com")
	if err != nil {
		t.Fatalf("failed to seed test user: %v", err)
	}
	// // log.Println("--- DEBUG: Alice successfully inserted ---")

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query count: %v", err)
	}
	log.Printf("--- REAL-TIME DB CHECK: There are currently %d users in the table", count)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			user, err := store.GetUser(tc.username)
			if err != tc.expectedErr {
				t.Fatalf("Expected %v\nGot %v", tc.expectedErr, err)
			}

			if err == nil {
				// 4. Assertions
				if user.Username != tc.username {
					t.Errorf("Expected %q\nGot %q", tc.username, user.Username)
				}
				if user.Email != tc.expectedEmail {
					t.Errorf("Expected %q\nGot %q", tc.expectedEmail, user.Email)
				}
				// Assert that the timestamp column maps accurately to the domain entity
				if user.CreatedAt.IsZero() {
					t.Error("Expected created_at timestamp to be populated, got a zero time value")
				}
			}
		})
	}
}
