package postgres_test

import (
	"database/sql"
	"errors"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/yudan-glitch/twitter-backend/internal/domain"
	"github.com/yudan-glitch/twitter-backend/internal/storage/postgres"
)

func TestPostgreSQLUserStore_GetUser(t *testing.T) {

	db, store := setupTestDB(t)
	// Good practice to close it once done
	defer db.Close()

	tests := []struct {
		name           string
		lookupUsername string
		expectedUser   domain.User
		expectedErr    error
	}{
		{
			name:           "Request non-existing user",
			lookupUsername: "unknown",
			expectedErr:    domain.ErrUserNotFound,
		},
		{
			name:           "Request existing user",
			lookupUsername: "alice",
			expectedUser: domain.User{
				Username:     "alice",
				Email:        "alice@example.com",
				PasswordHash: "1234",
			},
			expectedErr: nil,
		},
	}

	// Seed a real row into PostgreSQL
	// log.Println("--- DEBUG: About to insert Alice ---")
	_, err := db.Exec("INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)", "alice", "alice@example.com", "1234")
	if err != nil {
		t.Fatalf("failed to seed test user\n%v", err)
	}
	// // log.Println("--- DEBUG: Alice successfully inserted ---")

	// // DEBUG
	// var count int
	// err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	// if err != nil {
	// 	t.Fatalf("failed to query count: %v", err)
	// }
	// log.Printf("--- REAL-TIME DB CHECK: There are currently %d users in the table", count)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stored, err := store.GetUser(tc.lookupUsername)
			if err != tc.expectedErr {
				t.Fatalf("Expected %v\nGot %v", tc.expectedErr, err)
			}

			if err == nil {
				assertUserMatches(t, &tc.expectedUser, &stored)
			}
		})
	}
}

func TestPostgreSQLUserStore_CreateUser_Success(t *testing.T) {

	db, store := setupTestDB(t)
	// Good practice to close it once done
	defer db.Close()

	// Make it a pointer here
	input := &domain.User{
		// ID and CreatedAt should be automatically set (NOT TRUE!!!!, it gives default values)
		Username:     "new_user",
		Email:        "newmail@example.com",
		PasswordHash: "password",
	}

	err := store.CreateUser(input)
	if err != nil {
		t.Fatal("expecting nil, got error")
	}

	stored, err := store.GetUser(input.Username)
	if err != nil {
		t.Fatalf("created user insertion failed\n%v", err)
	}

	assertUserMatches(t, input, &stored)
}

func TestPostgreSQLUserStore_CreateUser_Duplicate(t *testing.T) {
	db, store := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec("INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)", "alice", "alice@example.com", "1234")
	if err != nil {
		t.Fatalf("failed to seed test user\n%v", err)
	}

	// Real-World Application Note
	// If the goal is to show both errors to a client in an API response simultaneously
	// (e.g., "Username is taken" AND "Email is taken"), that logic cannot live in the
	// low-level database storage layer. Instead, handle that later up in the HTTP Handler
	// layer by performing rapid, non-blocking check queries before attempting the insert,
	// or by prioritizing the username message return rule.

	t.Run("Duplicate username", func(t *testing.T) {
		input := &domain.User{
			Username:     "alice",
			Email:        "aliceeee@example.com",
			PasswordHash: "password",
		}
		err = store.CreateUser(input)

		if err == nil {
			t.Errorf("Expected error, got nil\n%v", err)
		}

		if !errors.Is(err, domain.ErrUsernameTaken) {
			t.Errorf("Expected %v\nGot %v", domain.ErrUsernameTaken, err)
		}
	})

	t.Run("Duplicate email", func(t *testing.T) {
		input := &domain.User{
			Username:     "aliceeee",
			Email:        "alice@example.com",
			PasswordHash: "password",
		}
		err = store.CreateUser(input)

		if err == nil {
			t.Errorf("Expected error, got nil\n%v", err)
		}

		if !errors.Is(err, domain.ErrEmailTaken) {
			t.Errorf("Expected %v\nGot %v", domain.ErrEmailTaken, err)
		}
	})

}

func setupTestDB(t testing.TB) (*sql.DB, *postgres.PostgreSQLUserStore) {
	t.Helper()

	// STEP. 1
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

	_, err = db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	if err != nil {
		db.Close()
		t.Fatalf("failed to truncate users table: %v", err)
	}

	store := postgres.NewPostgreSQLUserStore(db)
	return db, store
}

func assertUserMatches(t testing.TB, expected, got *domain.User) {
	t.Helper()

	// ID Validation
	if expected.ID != 0 {
		// If the test specified an exact ID, they must match perfectly
		if got.ID != expected.ID {
			t.Errorf("mismatch on field ID: expected %d, got %d", expected.ID, got.ID)
		}
	} else {
		// If the test left ID as 0, just make sure the database populated something
		if got.ID == 0 {
			t.Error("expected database to assign a non-zero serial ID")
		}
	}

	// Username Matching
	if expected.Username != got.Username {
		t.Errorf("Expected username %q\nGot %q", expected.Username, got.Username)
	}

	// Email Matching
	if expected.Email != got.Email {
		t.Errorf("Expected email %q\nGot %q", expected.Email, got.Email)
	}

	// Password Hash Matching
	if expected.PasswordHash != got.PasswordHash {
		t.Errorf("Expected password hash %q\nGot %q", expected.PasswordHash, got.PasswordHash)
	}

	// Created At Validation
	// Dynamic validation rule: If the test defines a timestamp, match them exactly.
	// Otherwise, simply assert that the database returned a non-zero time entry.
	if !expected.CreatedAt.IsZero() {
		if !got.CreatedAt.Equal(expected.CreatedAt) {
			t.Errorf("mismatch on field CreatedAt: expected %v, got %v", expected.CreatedAt, got.CreatedAt)
		}

		// If PostgreSQL successfully saved the timestamp, got.CreatedAt will contain a real time value (e.g., 2026-07-22...)
	} else if got.CreatedAt.IsZero() {
		t.Error("expected stored created_at timestamp to be populated, got a zero time value")
	}
}
