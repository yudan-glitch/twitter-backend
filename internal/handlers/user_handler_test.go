// Using a separate test package enforces black-box testing and keeps TDD clean.

// Package handlers_test implements external integration and unit tests for the handlers package.
package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yudan-glitch/twitter-backend/internal/auth"
	"github.com/yudan-glitch/twitter-backend/internal/domain"
	"github.com/yudan-glitch/twitter-backend/internal/handlers"
	"github.com/yudan-glitch/twitter-backend/internal/storage/mock"
)

const apiVersion = "/api/v1"

// --- Get User ---

func TestHandleGetSpecificUser(t *testing.T) {

	// 1. Define in-memory db for testing.
	mockStore := &mock.MockUserStore{
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
			handler := handlers.HandleGetSpecificUser(mockStore)
			handler.ServeHTTP(w, r) // The handler gets exectued here

			// Assertions
			assertCode(t, tc.expectedCode, w.Code)
			tc.assert(t, w)
		})
	}
}

// --- Create User ---

type userPayloadBuilder struct {
	payload handlers.RegisterRequest
}

func newUserPayloadBuilder() *userPayloadBuilder {
	return &userPayloadBuilder{
		payload: handlers.RegisterRequest{
			Username: "user_02",
			Email:    "user@mail.com",
			Password: "123abc",
		},
	}
}

func (b *userPayloadBuilder) withUsername(u string) *userPayloadBuilder {
	b.payload.Username = u
	return b
}
func (b *userPayloadBuilder) withEmail(e string) *userPayloadBuilder {
	b.payload.Email = e
	return b
}
func (b *userPayloadBuilder) withPassword(p string) *userPayloadBuilder {
	b.payload.Password = p
	return b
}
func (b *userPayloadBuilder) build() handlers.RegisterRequest {
	return b.payload
}

func TestHandleCreateUser(t *testing.T) {

	// Define multiple test cases (Table-Driven Test)
	tests := []struct {
		name         string
		userPayload  handlers.RegisterRequest
		expectedCode int
		assert       func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:         "(1) Valid Input Registration",
			userPayload:  newUserPayloadBuilder().build(),
			expectedCode: http.StatusCreated,
			assert:       assertUserResponse("user_02"),
		},
		{
			name:         "(2a) Duplicate: username",
			userPayload:  newUserPayloadBuilder().withUsername("unique").build(),
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(domain.ErrUsernameTaken),
		},
		{
			name:         "(2b) Duplicate: email",
			userPayload:  newUserPayloadBuilder().withEmail("unique@mail.com").build(),
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(domain.ErrEmailTaken),
		},
		{
			name:         "(3a) Invalid username: too short",
			userPayload:  newUserPayloadBuilder().withUsername("usr").build(),
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidUsername),
		},
		{
			name:         "(3b) Invalid username: too long",
			userPayload:  newUserPayloadBuilder().withUsername("usernametoolongxx").build(),
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidUsername),
		},
		{
			name:         "(3c) Invalid username: regex violation",
			userPayload:  newUserPayloadBuilder().withUsername("_user_").build(),
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidUsername),
		},
		{
			name:         "(4a) Invalid email: invalid syntax",
			userPayload:  newUserPayloadBuilder().withEmail("validexampledotcom").build(),
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidEmail),
		},
		{
			name:         "(4b) Invalid email: too long",
			userPayload:  newUserPayloadBuilder().withEmail("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@://bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.com").build(),
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidEmail),
		},
		{
			name:         "(5) Invalid password: too short",
			userPayload:  newUserPayloadBuilder().withPassword("abc12").build(),
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidPassword),
		},
		{
			name:         "(6a) Empty field: username",
			userPayload:  newUserPayloadBuilder().withUsername("").build(),
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidUsername),
		},
		{
			name:         "(6b) Empty field: email",
			userPayload:  newUserPayloadBuilder().withEmail("").build(),
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidEmail),
		},
		{
			name:         "(6c) Empty field: password",
			userPayload:  newUserPayloadBuilder().withPassword("").build(),
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
			mockStore := &mock.MockUserStore{
				Users: map[string]domain.User{
					"unique": {
						ID:           1,
						Username:     "unique",
						Email:        "unique@mail.com",
						PasswordHash: "dummy_bcrypt_hash",
						CreatedAt:    time.Now(),
					},
				},
			}

			// Prepare Server Response
			w := httptest.NewRecorder()

			// Simulate Browser Request
			requestBodyBuffer := encodePayload(t, tc.userPayload)
			r := httptest.NewRequest(http.MethodPost, route, requestBodyBuffer)
			r.Header.Set("Content-Type", "application/json")

			// Execute request and record response
			handlers.HandleCreateUser(mockStore).ServeHTTP(w, r)

			// Assertions
			assertCode(t, tc.expectedCode, w.Code)
			tc.assert(t, w)
		})
	}

}

func TestHandleCreateUser_PayloadLimits(t *testing.T) {
	route := apiVersion + "/users"
	mockStore := &mock.MockUserStore{}

	t.Run("Reject massive payloads", func(t *testing.T) {
		// Create a body that starts as valid JSON but is very long
		// This forces the decoder to keep reading until it hits the limit.
		hugeValue := strings.Repeat("a", 1024*1024+100)
		bodyString := `{"username":"` + hugeValue + `"}`
		hugeBody := strings.NewReader(bodyString)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, route, hugeBody)

		handlers.HandleCreateUser(mockStore).ServeHTTP(w, r)

		assertCode(t, http.StatusRequestEntityTooLarge, w.Code)
		assertErrorResponse(handlers.ErrPayloadTooLarge)(t, w)
	})

	t.Run("Reject unknown JSON fields", func(t *testing.T) {
		body := bytes.NewBufferString(`{"username":"bob","email":"b@b.com","password":"password123","hacker_field":"inject"}`)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, route, body)

		handlers.HandleCreateUser(mockStore).ServeHTTP(w, r)

		assertCode(t, http.StatusBadRequest, w.Code)
		assertErrorResponse(handlers.ErrUnknownFields)(t, w)
	})
}

// --- LOGIN ---
func TestHandleLogin(t *testing.T) {

	tests := []struct {
		name         string
		loginPayload handlers.LoginRequest
		expectedCode int
		mockFail     bool // tell the mock to fail for a specific test
	}{
		{
			name: "Unknown Email",
			loginPayload: handlers.LoginRequest{
				Email:    "unknown@mail.com",
				Password: "password", // Doesn't matter, the account doesn't exist anyway.
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "Existing Email, Wrong Password",
			loginPayload: handlers.LoginRequest{
				Email:    "user@mail.com",
				Password: "wrong_password",
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "Existing Email, Right Password",
			loginPayload: handlers.LoginRequest{
				Email:    "user@mail.com",
				Password: "right_password",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Session Store Error",
			loginPayload: handlers.LoginRequest{
				Email:    "user@mail.com",
				Password: "right_password",
			},
			expectedCode: http.StatusInternalServerError,
			mockFail:     true,
		},
	}

	mockPasword := "right_password"
	hashedMockPassword, _ := auth.HashPassword(mockPasword)

	mockStore := &mock.MockUserStore{
		Users: map[string]domain.User{
			"user": {
				ID:           1,
				Username:     "user",
				Email:        "user@mail.com",
				PasswordHash: hashedMockPassword,
				CreatedAt:    time.Now(),
			},
		},
	}

	mockSessionStore := &mock.MockSessionStore{}

	route := apiVersion + "/login"

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			// Set the failure state (I might have to move this test on a new
			// function, I don't like this current approach)
			mockSessionStore.FailCreate = tc.mockFail

			// Prepare server response
			w := httptest.NewRecorder()

			// Simulate browser request
			requestBodyBuffer := encodePayload(t, tc.loginPayload)
			r := httptest.NewRequest(http.MethodPost, route, requestBodyBuffer)
			r.Header.Set("Content-Type", "application/json")

			// Execute request
			handlers.HandleLoginRequest(mockStore, mockSessionStore).ServeHTTP(w, r)

			// Assertions
			assertCode(t, tc.expectedCode, w.Code)

			// If Login Successful, i MUST have a session cookie
			if tc.expectedCode == http.StatusOK {
				assertCookie(t, w, "session_id")
			}
		})
	}
}

func assertCookie(t *testing.T, w *httptest.ResponseRecorder, name string) {
	t.Helper()

	cookies := w.Result().Cookies()

	for _, c := range cookies {
		if c.Name == name {
			// Check security attributes
			if c.Value == "" {
				t.Errorf("cookie %s found but value is empty", name)
			}
			if !c.HttpOnly {
				t.Errorf("cookie %s should be HttpOnly", name)
			}
			if !c.Secure {
				t.Errorf("cookie %s should be Secure", name)
			}
			if c.SameSite != http.SameSiteLaxMode {
				t.Errorf("cookie %s should have SameSite=Lax", name)
			}
			return
		}
	}
	t.Error("expected session cookie 'session_id' but it was not found")
}

func encodePayload(t *testing.T, payload any) *bytes.Buffer {
	t.Helper()

	// Create a buffer to hold the encoded JSON bytes
	requestBodyBuffer := new(bytes.Buffer)

	// Encode the data into the buffer
	err := json.NewEncoder(requestBodyBuffer).Encode(payload)
	if err != nil {
		t.Fatalf("error encoding user payload\n%v", err)
	}

	return requestBodyBuffer
}
