package handlers_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/solid-state-dan/twitter-backend/internal/handlers"
)

// assertCode checks that the response code matches the expected value.
func assertCode(t testing.TB, expectedCode, got int) {
	t.Helper()
	if expectedCode != got {
		t.Errorf("bad return code\nExpected %d\nGot %d", expectedCode, got)
	}
}

// assertUserResponse returns a clean assertion function for 200 OK responses
func assertUserResponse(expectedUsername string) func(*testing.T, *httptest.ResponseRecorder) {

	return func(t *testing.T, w *httptest.ResponseRecorder) {
		t.Helper()

		got := decodeJSON[handlers.UserResponse](t, w.Body)

		// Assertions
		if got.Username != expectedUsername {
			t.Errorf("bad user response\nExpected username %q\nGot %q", expectedUsername, got.Username)
		}
	}
}

// assertErrorResponse returns a clean assertion function for failure responses
func assertErrorResponse(expectedError error) func(*testing.T, *httptest.ResponseRecorder) {

	return func(t *testing.T, w *httptest.ResponseRecorder) {
		t.Helper()

		got := decodeJSON[handlers.ErrorResponse](t, w.Body)

		// Assertions
		if got.Error != handlers.GetClientErrorMessage(expectedError) {
			t.Errorf("bad client error response\nExpected %q\nGot %q", handlers.GetClientErrorMessage(expectedError), got.Error)
		}
	}
}

// decodeJSON decodes a response body into a target type.
func decodeJSON[T any](t *testing.T, body io.Reader) T {
	t.Helper()

	var data T
	if err := json.NewDecoder(body).Decode(&data); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	return data
}
