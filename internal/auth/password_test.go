package auth_test

import (
	"testing"

	"github.com/yudan-glitch/twitter-backend/internal/auth"
)

func TestPasswordHashing(t *testing.T) {

	password := "1234"

	// 1. Test Hashing
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password\n%v", err)
	}

	if hash == password {
		t.Error("hash should not be equal to plain text password")
	}

	// 2. Test Verification (Correct Password)
	if !auth.CheckPasswordHash(password, hash) {
		t.Error("CheckPasswordHash should return true for correct password")
	}

	// 3. Test Verification (Wrong Password)
	if auth.CheckPasswordHash("wrong-password", hash) {
		t.Error("CheckPasswordHash should return false for incorrect password")
	}
}
