package mock_test

import (
	"testing"
	"time"

	"github.com/yudan-glitch/twitter-backend/internal/domain"
	"github.com/yudan-glitch/twitter-backend/internal/storage/mock"
)

func TestMockUserStore_GetUser(t *testing.T) {
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
				PasswordHash: "secret_hash",
			},
			expectedErr: nil,
		},
	}

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

func TestMockUserStore_CreateUser_Success(t *testing.T) {

	store := &mock.MockUserStore{
		Users: map[string]domain.User{
			"mock_tester": {
				ID:           1,
				Username:     "mock_tester",
				Email:        "mock@example.com",
				PasswordHash: "secret_hash",
				CreatedAt:    time.Now(),
			},
		},
	}

	t.Run("Duplicate username", func(t *testing.T) {
		input := &domain.User{
			Username:     "mock_tester",
			Email:        "tester@example.com",
			PasswordHash: "secret_hash",
		}

		err := store.CreateUser(input)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		if err != domain.ErrUsernameTaken {
			t.Errorf("Expected %v\nGot %v", domain.ErrUsernameTaken, err)
		}
	})

	t.Run("Duplicate email", func(t *testing.T) {
		input := &domain.User{
			Username:     "tester",
			Email:        "mock@example.com",
			PasswordHash: "secret_hash",
		}

		err := store.CreateUser(input)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		if err != domain.ErrEmailTaken {
			t.Errorf("Expected %v\nGot %v", domain.ErrEmailTaken, err)
		}
	})

}

func TestMockUserStore_CreateUser_Duplicate(t *testing.T) {

	// Initialize a completely blank mock store instance
	store := &mock.MockUserStore{
		Users: make(map[string]domain.User),
	}

	input := &domain.User{
		Username:     "mock_tester",
		Email:        "mock@example.com",
		PasswordHash: "secret_hash",
	}

	err := store.CreateUser(input)
	if err != nil {
		t.Fatalf("expected mock to save user without error, got %v", err)
	}

	stored, err := store.GetUser(input.Username)
	if err != nil {
		t.Fatalf("created user insertion failed\n%v", err)
	}

	assertUserMatches(t, input, &stored)
}

// Keep It Local (Simplest & Recommended for TDD)
// Copy the assertUserMatches function directly into internal/storage/mock/user_mock_test.go.
// Why: Test files ending in _test.go are completely ignored by the production compiler.
// Having the same helper function in both postgres_test and mock_test is not considered
// "bad duplication" because they belong to completely independent test suites that never interact.
// It keeps your mock tests entirely self-contained.
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
