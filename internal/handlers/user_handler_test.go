// Using a separate test package enforces black-box testing and keeps TDD clean.

// Package handlers_test implements external integration and unit tests for the handlers package.
package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yudan-glitch/twitter-backend/internal/domain"
	"github.com/yudan-glitch/twitter-backend/internal/handlers"
	"github.com/yudan-glitch/twitter-backend/internal/storage/mock"
)

const apiVersion = "/api/v1"

func TestHandleGetSpecificUser(t *testing.T) {

	// 1. Define in-memory db for testing.
	mockStore := mock.MockUserStore{
		Users: map[string]domain.User{
			"alice":  {Username: "alice"},
			"jessie": {Username: "jessie"},
		},
	}

	// 2. Define multiple test cases (Table-Driven Test).
	tests := []struct {
		name         string
		username     string
		expectedCode int
		assert       func(t *testing.T, w *httptest.ResponseRecorder) // Pass a custom body assertion for each case
	}{
		{
			name:         "1a) Request existing user",
			username:     "alice",
			expectedCode: http.StatusOK,
			assert:       assertUserResponse("alice"),
		},
		{
			name:         "1b) Request existing user",
			username:     "jessie",
			expectedCode: http.StatusOK,
			assert:       assertUserResponse("jessie"),
		},
		{
			name:         "2) Request unknown user",
			username:     "unknown",
			expectedCode: http.StatusNotFound,
			assert:       assertErrorResponse(domain.ErrUserNotFound),
		},

		// I previously had a "3) Empty request" test case, so the target would've
		// been "/users/", but the way Go works, in production the router bypasses
		// it and this specfic case will never actually get triggered.
		// So I changed it to input too short/long instead. (I plan to add
		// more input validations later on).

		{ // Router allows these inputs, but our business logic rejects it!
			name:         "3a) Username too short",
			username:     "a",
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidUsername),
		},
		{
			name:         "3b) Username too long",
			username:     "aanovunqeovnqoienoieqvnoeqnvoqeivoqenv",
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidUsername),
		},
	}

	// 3. Run each test.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			// Construct full versioned API path
			route := apiVersion + "/users/" + tc.username

			// Simulate browser request and server response
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, route, nil)
			r.SetPathValue("name", tc.username)

			// Inject the mock into the handler
			handler := handlers.HandleGetSpecificUser(&mockStore)
			handler.ServeHTTP(w, r) // The handler gets exectued here

			// Assertions
			assertCode(t, tc.expectedCode, w.Code)
			tc.assert(t, w)
		})
	}
}
