package crypto_test

import (
	"testing"

	"github.com/solid-state-dan/twitter-backend/internal/crypto"
)

func TestPasswordHashing(t *testing.T) {

	password := "1234"

	// 1. Test Hashing
	hash, err := crypto.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password\n%v", err)
	}

	if hash == password {
		t.Error("hash should not be equal to plain text password")
	}

	// 2. Test Verification (Correct Password)
	if !crypto.VerifyPassword(password, hash) {
		t.Error("CheckPasswordHash should return true for correct password")
	}

	// 3. Test Verification (Wrong Password)
	if crypto.VerifyPassword("wrong-password", hash) {
		t.Error("CheckPasswordHash should return false for incorrect password")
	}
}
