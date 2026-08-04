package postgres_test

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/yudan-glitch/twitter-backend/internal/crypto"
	"github.com/yudan-glitch/twitter-backend/internal/domain"
	"github.com/yudan-glitch/twitter-backend/internal/storage/postgres"
)

func TestPostgreSQLUserStore_GetUser(t *testing.T) {

	db, store := setupTestUserDB(t)
	// Good practice to close it once done
	defer db.Close()

	tests := []struct {
		name           string
		lookupUsername string
		expectedUser   *domain.User
		expectedErr    error
	}{
		{
			// Start from lazy case
			name:           "(1) Request non-existing user",
			lookupUsername: "unknown",
			expectedErr:    domain.ErrUserNotFound,
		},
		{
			name:           "(2) Request existing user",
			lookupUsername: "alice",
			expectedUser: &domain.User{
				ID:           1,
				Username:     "alice",
				Email:        "alice@example.com",
				PasswordHash: "secret_hash",
				// CreatedAt doesn't really matter here.
				// "created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP" automatically
				// sets a time when the row gets inserted. So this value will never be zero.
			},
			expectedErr: nil,
		},
	}

	// Seed a real row into PostgreSQL
	_, err := db.Exec("INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)", "alice", "alice@example.com", "secret_hash")
	if err != nil {
		t.Fatalf("failed to seed test user\n%v", err)
	}

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
				t.Fatalf("bad return error\nExpected %v\nGot %v", tc.expectedErr, err)
			}

			if err == nil {
				assertUserMatches(t, tc.expectedUser, stored)
			}
		})
	}
}

func TestPostgreSQLUserStore_CreateUser_Success(t *testing.T) {

	db, store := setupTestUserDB(t)
	// Good practice to close it once done
	defer db.Close()

	// Define db for testing
	_, err := db.Exec("INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)", "unique", "unique@example.com", "secret_hash")
	if err != nil {
		t.Fatalf("failed to seed test user\n%v", err)
	}

	// Define test cases (Table-Driven Test)
	tests := []struct {
		name        string
		inputUser   domain.User
		expectedErr error
	}{
		{
			name: "(1) Duplicate Username",
			inputUser: domain.User{
				Username:     "unique",
				Email:        "different@example.com",
				PasswordHash: "secret_hash",
			},
			expectedErr: domain.ErrUsernameTaken,
		},
		{
			name: "(2) Duplicate Email",
			inputUser: domain.User{
				Username:     "different",
				Email:        "unique@example.com",
				PasswordHash: "secret_hash",
			},
			expectedErr: domain.ErrEmailTaken,
		},
		{
			name: "(3) Valid input registration",
			inputUser: domain.User{
				Username:     "different",
				Email:        "different@example.com",
				PasswordHash: "secret_hash",
			},
			expectedErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := store.CreateUser(&tc.inputUser)
			if err != tc.expectedErr {
				t.Fatalf("create user: bad return error\nExpected %v\nGot %v", tc.expectedErr, err)
			}

			// Check if user was registered correctly
			if err == nil {
				stored, err := store.GetUser(tc.inputUser.Username)
				if err != nil {
					t.Fatalf("get user: bad return error\nExpecting nil\nGot %v", err)
				}

				assertUserMatches(t, &tc.inputUser, stored)
			}
		})
	}
}

func TestPostgreSQLUserStore_VerifyCredentials(t *testing.T) {
	db, store := setupTestUserDB(t)
	defer db.Close()

	// Define db for testing
	hashedPassword, err := crypto.HashPassword("correct_password")
	if err != nil {
		t.Fatalf("error hashing password\n%v", err)
	}
	_, err = db.Exec("INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)", "verify_me", "verify@mail.com", hashedPassword)
	if err != nil {
		t.Fatalf("failed to seed test user\n%v", err)
	}

	tests := []struct {
		name          string
		inputEmail    string
		inputPassword string
		expectedErr   error
	}{
		{
			name:          "Unknown Email",
			inputEmail:    "unknown@mail.com",
			inputPassword: "any_password",
			expectedErr:   domain.ErrUserNotFound,
		},
		{
			name:          "Valid Credentials",
			inputEmail:    "verify@mail.com",
			inputPassword: "correct_password",
			expectedErr:   nil,
		},
		{
			name:          "Wrong Password",
			inputEmail:    "verify@mail.com",
			inputPassword: "wrong_password",
			expectedErr:   domain.ErrInvalidLoginCredentials,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userID, err := store.VerifyCredentials(tc.inputEmail, tc.inputPassword)
			if err != tc.expectedErr {
				t.Fatalf("bad error response:\nExpected: %v\nGot: %v", tc.expectedErr, err)
			}

			if err == nil && userID == 0 {
				t.Error("expected non-zero userID for successful verification")
			}
		})
	}
}

func setupTestUserDB(t testing.TB) (*sql.DB, *postgres.PostgreSQLUserStore) {
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

	// Clean users table from previous tests
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
			t.Errorf("mismatch on field ID\nExpected %d\nGot %d", expected.ID, got.ID)
		}
	} else {
		// If the test left ID as 0, just make sure the database populated something
		if got.ID == 0 {
			t.Error("expected database to assign a non-zero serial ID")
		}
	}

	// Username Matching
	if expected.Username != got.Username {
		t.Errorf("mismatch on field Username\nExpected %q\nGot %q", expected.Username, got.Username)
	}

	// Email Matching
	if expected.Email != got.Email {
		t.Errorf("mismatch of field Email\nExpected%q\nGot %q", expected.Email, got.Email)
	}

	// Password Hash Matching
	if expected.PasswordHash != got.PasswordHash {
		t.Errorf("mismatch on field PasswordHash\nExpected %q\nGot %q", expected.PasswordHash, got.PasswordHash)
	}

	// Created At Validation
	// Assert that the database returned a non-zero time entry.
	if got.CreatedAt.IsZero() {
		t.Error("expected stored created_at timestamp to be populated, got a zero time value")
	}
}
