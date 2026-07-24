package mock_test

import (
	"testing"
	"time"

	"github.com/yudan-glitch/twitter-backend/internal/domain"
	"github.com/yudan-glitch/twitter-backend/internal/storage/mock"
)

func TestMockUserStore_GetUser(t *testing.T) {

	// Define in-memory db for testing
	store := &mock.MockUserStore{
		Users: map[string]domain.User{
			"alice": {
				ID:           1,
				Username:     "alice",
				Email:        "alice@example.com",
				PasswordHash: "secret_hash",
				CreatedAt:    time.Now(),
			},
		},
	}

	// Define multiple test cases (Table-Driven Test)
	tests := []struct {
		name           string
		lookupUsername string
		expectedUser   domain.User
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
			expectedUser: domain.User{
				ID:           1,
				Username:     "alice",
				Email:        "alice@example.com",
				PasswordHash: "secret_hash",
			},
			expectedErr: nil,
		},
	}

	// Run each test
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stored, err := store.GetUser(tc.lookupUsername)
			if err != tc.expectedErr {
				t.Fatalf("bad return error\nExpected %v\nGot %v", tc.expectedErr, err)
			}

			if err == nil {
				assertUserMatches(t, &tc.expectedUser, &stored)
			}
		})
	}

}

func TestMockUserStore_CreateUser(t *testing.T) {

	// Define in-memory db for testing
	store := &mock.MockUserStore{
		Users: map[string]domain.User{
			"unique": {
				ID:           1,
				Username:     "unique",
				Email:        "unique@example.com",
				PasswordHash: "secret_hash",
				CreatedAt:    time.Now(),
			},
		},
	}

	// Define test cases (Table-Driven Test)
	tests := []struct {
		name          string
		input         domain.User
		expectedError error
	}{
		{
			name: "(1) Duplicate Username",
			input: domain.User{
				Username:     "unique",
				Email:        "different@example.com",
				PasswordHash: "secret_hash",
			},
			expectedError: domain.ErrUsernameTaken,
		},
		{
			name: "(2) Duplicate Email",
			input: domain.User{
				Username:     "different",
				Email:        "unique@example.com",
				PasswordHash: "secret_hash",
			},
			expectedError: domain.ErrEmailTaken,
		},
		{
			name: "(3) Valid input registration",
			input: domain.User{
				Username:     "different",
				Email:        "different@example.com",
				PasswordHash: "secret_hash",
			},
			expectedError: nil,
		},
	}

	// Run each test
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := store.CreateUser(&tc.input)
			if err != tc.expectedError {
				t.Fatalf("create user: bad return error\nExpected %v\nGot %v", tc.expectedError, err)
			}
			if err == nil {
				stored, err := store.GetUser(tc.input.Username)
				if err != nil {
					t.Errorf("get user: bad return error\nExpected %v\nGot %v", nil, err)
				}
				assertUserMatches(t, &tc.input, &stored)
			}
		})
	}
}

// Keep It Local (Simplest & Recommended for TDD)
// Copied the assertUserMatches function from /internal/storage/postgres/user_store_test.go.
// Why: Test files ending in _test.go are completely ignored by the production compiler.
// Having the same helper function in both postgres_test and mock_test is not considered
// "bad duplication" because they belong to completely independent test suites that never interact.
// It keeps the mock tests entirely self-contained.
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
