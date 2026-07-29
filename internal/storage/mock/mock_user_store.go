package mock

import (
	"time"

	"github.com/yudan-glitch/twitter-backend/internal/auth"
	"github.com/yudan-glitch/twitter-backend/internal/domain"
	"github.com/yudan-glitch/twitter-backend/internal/handlers"
)

// Note: This uses the core domain.User entity layout instead of handlers.UserResponse

// MockUserStore provides an in-memory database mock for unit tests.
type MockUserStore struct {
	Users map[string]domain.User
}

// GetUser looks up the user in the mock map, returning a 404 error if missing.
func (m *MockUserStore) GetUser(name string) (*domain.User, error) {
	user, exist := m.Users[name]
	if !exist {
		return &domain.User{}, domain.ErrUserNotFound
	}
	return &user, nil
}

func (m *MockUserStore) CreateUser(user *domain.User) error {

	// If a user with the provided name already exists:
	_, exist := m.Users[user.Username]
	if exist {
		return domain.ErrUsernameTaken
	}

	// If it doesn't, check if the email is already taken
	for _, usr := range m.Users {
		if usr.Email == user.Email {
			return domain.ErrEmailTaken
		}
	}

	// Simulate database identity column index generation increment
	user.ID = int64(len(m.Users) + 1)

	user.CreatedAt = time.Now()

	// Save the copy directly into the in-memory key-value dictionary
	m.Users[user.Username] = *user
	return nil
}

func (m *MockUserStore) VerifyCredentials(email, password string) (userId int64, err error) {

	var targetUser *domain.User

	// Check for user with provided email
	for _, user := range m.Users {
		if user.Email == email {
			targetUser = &user
			break
		}
	}

	// No users with provided email
	if targetUser == nil {
		return 0, handlers.ErrInvalidLoginCredentials
	}

	// Verify the passwords match
	if !auth.VerifyPassword(password, targetUser.PasswordHash) {
		return 0, handlers.ErrInvalidLoginCredentials
	}

	// Login Success
	return targetUser.ID, nil
}
