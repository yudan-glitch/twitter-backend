package mock

import (
	"github.com/yudan-glitch/twitter-backend/internal/domain"
)

// Note: This uses the core domain.User entity layout instead of handlers.UserResponse

// MockUserStore provides an in-memory database mock for unit tests.
type MockUserStore struct {
	Users map[string]domain.User
}

// GetUser looks up the user in the mock map, returning a 404 error if missing.
func (m *MockUserStore) GetUser(name string) (domain.User, error) {
	user, exist := m.Users[name]
	if !exist {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}
