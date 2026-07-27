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
				Username: "maya_02",
				Email:    "maya@example.com",
				Password: "123456",
			},
			expectedCode: http.StatusCreated,
			assert:       assertUserResponse("maya_02"),
		},
		{
			name: "(2a) Duplicate: username",
			userPayload: handlers.UserRequest{
				Username: "unique",
				Email:    "different@example.com",
				Password: "123456",
			},
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(domain.ErrUsernameTaken),
		},
		{
			name: "(2b) Duplicate: email",
			userPayload: handlers.UserRequest{
				Username: "different",
				Email:    "unique@example.com",
				Password: "123456",
			},
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(domain.ErrEmailTaken),
		},
		{
			name: "(3a) Invalid username: too short",
			userPayload: handlers.UserRequest{
				Username: "usr",
				Email:    "valid@example.com",
				Password: "123456",
			},
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidUsername),
		},
		{
			name: "(3b) Invalid username: too long",
			userPayload: handlers.UserRequest{
				Username: "usernametooloooong",
				Email:    "valid@example.com",
				Password: "123456",
			},
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidUsername),
		},
		{
			name: "(3c) Invalid username: regex violation",
			userPayload: handlers.UserRequest{
				Username: "_user_",
				Email:    "valid@example.com",
				Password: "123456",
			},
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidUsername),
		},
		{
			name: "(4a) Invalid email: invalid syntax",
			userPayload: handlers.UserRequest{
				Username: "valid",
				Email:    "validexampledotcom",
				Password: "123456",
			},
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidEmail),
		},
		{
			name: "(4b) Invalid email: too long",
			userPayload: handlers.UserRequest{
				Username: "valid",
				Email: "qmvpeimvqpnvepqnveqienvoasjdpwmfoqijeoiqjefoiwqoifjowqjfowqjfoiwqjfo" +
					"iqjwfoijqwofijqwoijfoiwqjfoiwqjfoijwfpqiwnpwqingpqingpiqnegpqinepinqeog" +
					"iewqeinvoqinefjqeinfoqenfoqinfiqnwfoiqneoineqfnqonfnwqinfvoqienvoqiqwmc" +
					"monqiwnvoiqnvoiqnevoneqvoinqeoinoivqnnennvqie@example.com",
				Password: "123456",
			},
			expectedCode: http.StatusBadRequest,
			assert:       assertErrorResponse(handlers.ErrInvalidEmail),
		},
		{
			name: "(5) Invalid password: too short",
			userPayload: handlers.UserRequest{
				Username: "valid",
				Email:    "valide@example.com",
				Password: "12345",
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
						PasswordHash: "dummy_bcrypt_hash",
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

			// 1. Call Create User Handler
			handler := handlers.HandleCreateUser(&mockStore)
			handler.ServeHTTP(w, r)

			// 2. Assertions
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
