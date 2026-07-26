// Using a separate test package enforces black-box testing and keeps TDD clean.

// Package handlers_test implements external integration and unit tests for the handlers package.
package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestHandleCreateUser(t *testing.T) {

	// Define multiple test cases (Table-Driven Test)
	tests := []struct {
		name         string
		userPayload  handlers.UserRequest
		expectedCode int
		assert       func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "(1) Valid Input Registration",
			// Prepare the payload data (DTO)
			userPayload: handlers.UserRequest{
				Username: "maya",
				Email:    "maya@example.com",
				Password: "1234",
			},
			expectedCode: http.StatusCreated,
			assert:       assertUserResponse("maya"),
		},
		{
			name: "(2) Duplicate username",
			userPayload: handlers.UserRequest{
				Username: "unique",
				Email:    "different@example.com",
				Password: "1234",
			},
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(domain.ErrUsernameTaken),
		},
		{
			name: "(3) Duplicate email",
			userPayload: handlers.UserRequest{
				Username: "different",
				Email:    "unique@example.com",
				Password: "1234",
			},
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(domain.ErrEmailTaken),
		},
		{
			name: "(4) Invalid username",
			userPayload: handlers.UserRequest{
				Username: "a",
				Email:    "aaaaaaaaaa@example.com",
				Password: "1234",
			},
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidUsername),
		},
		{
			name: "(5) Invalid email",
			userPayload: handlers.UserRequest{
				Username: "valid",
				Email:    "validexample.com",
				Password: "1234",
			},
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidEmail),
		},
		{
			name: "(5) Invalid password",
			userPayload: handlers.UserRequest{
				Username: "valid",
				Email:    "valide@example.com",
				Password: "",
			},
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidPassword),
		},
	}

	// Contstruct full versioned API path
	route := apiVersion + "/users"

	// Run each test
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			// Define in-memory db for testing **per subtest**
			mockStore := mock.MockUserStore{
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

			// ---------------
			// --- REQUEST ---
			// ---------------
			// 1. Create a buffer to hold the encoded JSON bytes
			requestBodyBuffer := new(bytes.Buffer)

			// 2. Encode the data into the buffer
			err := json.NewEncoder(requestBodyBuffer).Encode(tc.userPayload)
			if err != nil {
				t.Fatalf("error encoding user payload: %v", err)
			}

			// 3. Simulate browser request and server response
			w := httptest.NewRecorder()
			// Pass the buffer (which is an io.Reader) into the request
			r := httptest.NewRequest(http.MethodPost, route, requestBodyBuffer)
			r.Header.Set("Content-Type", "application/json")

			// ----------------
			// --- RESPONSE ---
			// ----------------
			// // Try fetch the user
			// // Note: not wrong, but it adds noise and duplicated knowledge.
			// //       Handler test should focus on HTTP response; store setup
			// //       can be implicit in the seed map

			// _, err = mockStore.GetUser(tc.userPayload.Username)
			// if err != tc.expectedGetUserErr {
			// 	t.Fatalf("get user: bad return error\nExpected %v\nGot %v", tc.expectedGetUserErr, err)
			// }

			// 1. Call Create User Handler
			handler := handlers.HandleCreateUser(&mockStore)
			handler.ServeHTTP(w, r)

			// 2. Assertions
			assertCode(t, tc.expectedCode, w.Code)
			tc.assert(t, w)
		})
	}

}
