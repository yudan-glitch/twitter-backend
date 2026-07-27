// Package handlers implements the HTTP request handlers for the router.
package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"github.com/yudan-glitch/twitter-backend/internal/auth"
	"github.com/yudan-glitch/twitter-backend/internal/domain"
)

var (
	ErrInvalidUsername = errors.New("invalid username")
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")

	// ErrPayloadTooLarge is returned when the request body exceeds the limit.
	ErrPayloadTooLarge = errors.New("request body too large")

	// ErrUnknownFields is returned when the JSON contains fields not in the DTO.
	ErrUnknownFields = errors.New("request body contains unknown fields")
)

// Regex Breakdown:
// ^[a-zA-Z0-9]      -> Must start with 1 alphanumeric character
// [a-zA-Z0-9_]{2,13} -> Middle chunk can include underscores, length 2 to 13
// [a-zA-Z0-9]$      -> Must end with 1 alphanumeric character
// Total length: 1 + (2 to 13) + 1 = 4 to 15 characters
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_]{2,13}[a-zA-Z0-9]$`)

func isValidUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

func isValidEmail(email string) bool {
	// Too Long (The maximum length defined by RFC 5321 is 254 characters)
	if len(email) > 254 {
		return false
	}
	// Invalid Syntax
	_, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return true
}

// UserResponse is the clean, public Data Transfer Object (DTO).
// It only includes fields that are 100% safe to send over the network.
// HTTP JSON payload shapes live here.
type UserResponse struct {
	Username string `json:"username"`
}

type UserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

const maxRequestBodySize = 1024 * 1024 // 1MB

// HandleGetSpecificUser takes a 'store' (database layer) and returns an HTTP
// handler function that fetches a single user by name.
func HandleGetSpecificUser(store domain.UserStore) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Grab the username from the URL path (e.g., /users/{name}).
		name := r.PathValue("name")

		// 2. Validate username length constraints before querying.
		if len(name) < 3 || len(name) > 20 {
			respondWithError(w, http.StatusBadRequest, ErrInvalidUsername)
			return
		}

		// 3. Ask the database to go fetch the user.
		// Note: When testing, this calls the mock.
		//       In production, it calls the real database.
		user, err := store.GetUser(name)

		if err != nil {

			// If user doesn't exist, return 404 Not Found.
			if errors.Is(err, domain.ErrUserNotFound) {
				respondWithError(w, http.StatusNotFound, domain.ErrUserNotFound)
				return
			}

			// Else it's a database error. Log the real technical error internally...
			log.Printf("Database failure: %v", err)
			// ...but hide the raw error from the public client.
			respondWithError(w, http.StatusInternalServerError, err)
			return
		}

		// Success! Map the user to UserResponse and send it back with 200 OK.
		userResp := UserResponse{Username: user.Username}
		respondWithJSON(w, http.StatusOK, userResp)
	}
}

// HandleCreateUser takes a 'store' (database layer) and returns an HTTP
// handler function that creates a single user by user payload.
func HandleCreateUser(store domain.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Limit the size of the request body to prevent DOS
		r.Body = http.MaxBytesReader(w, r.Body, int64(maxRequestBodySize))

		// Decode user payload
		var incomingUserData UserRequest
		decoder := json.NewDecoder(r.Body)
		// Prevent DOS via unknown large fields
		decoder.DisallowUnknownFields()

		err := decoder.Decode(&incomingUserData)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				// Internal: http.MaxBytesError -> Client: ErrPayloadTooLarge
				respondWithError(w, http.StatusRequestEntityTooLarge, ErrPayloadTooLarge)
				return
			}

			// Check if error is due to DisallowUnknownFields
			if strings.Contains(err.Error(), "unknown field") {
				respondWithError(w, http.StatusBadRequest, ErrUnknownFields)
				return
			}

			respondWithError(w, http.StatusBadRequest, err)
			return
		}

		// Payload Validation:
		// Username (4-15 characters)
		if !isValidUsername(incomingUserData.Username) {
			respondWithError(w, http.StatusBadRequest, ErrInvalidUsername)
			return
		}

		// Email
		if !isValidEmail(incomingUserData.Email) {
			respondWithError(w, http.StatusBadRequest, ErrInvalidEmail)
			return
		}

		// Password too short (at least 6 characters)
		if len(incomingUserData.Password) < 6 {
			respondWithError(w, http.StatusBadRequest, ErrInvalidPassword)
			return
		}

		// Hash password before creating domain User
		hashedPassword, err := auth.HashPassword(incomingUserData.Password)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err)
			return
		}

		// Map UserRequest to domain User
		user := domain.User{
			Username:     incomingUserData.Username,
			Email:        incomingUserData.Email,
			PasswordHash: hashedPassword,
		}

		// Create the user
		// Note: already handled errors at the storage layer (so if the username
		// or email is already taken, it will be checked there, not here)
		err = store.CreateUser(&user)

		if err != nil {
			switch err {
			case domain.ErrUsernameTaken, domain.ErrEmailTaken:
				respondWithError(w, http.StatusBadRequest, err)
				return
			default:
				respondWithError(w, http.StatusInternalServerError, err)
				return
			}
		}
		// Success! Map the user to UserResponse and send it back with 201 CREATED.
		userResp := UserResponse{
			Username: user.Username,
		}
		respondWithJSON(w, http.StatusCreated, userResp)
	}
}

// respondWithError builds a standardized JSON error message and writes it to the response stream.
func respondWithError(w http.ResponseWriter, statusCode int, err error) {
	// Set headers and encode the actual struct to JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error: GetClientErrorMessage(err),
	}

	json.NewEncoder(w).Encode(response)
}

// respondWithJSON converts the provided data structure as JSON and writes it to the response stream.
func respondWithJSON(w http.ResponseWriter, statusCode int, response any) {
	// Set headers and encode the actual struct to JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)

}

// GetClientErrorMessage translates internal backend errors to user-friendly strings.
func GetClientErrorMessage(err error) string {
	switch {
	// User Request Fields
	case errors.Is(err, ErrInvalidUsername):
		return "The username must be between 3 and 20 characters long. Please provide a valid username."
	case errors.Is(err, ErrInvalidEmail):
		return "Invalid email. Please try again with a different email."
	case errors.Is(err, ErrInvalidPassword):
		return "Password must contain at least 6 characters. Please try again with a different password."

	// Payload
	case errors.Is(err, ErrPayloadTooLarge):
		return "The request payload is too large. Max size is 1MB."
	case errors.Is(err, ErrUnknownFields):
		return "The request contains fields that are not allowed."

	// Domain
	case errors.Is(err, domain.ErrUserNotFound):
		return "We couldn't find an account matching that username. Please verify the spelling and try again."
	case errors.Is(err, domain.ErrUsernameTaken):
		return "Username already taken. Please try again with a different username."
	case errors.Is(err, domain.ErrEmailTaken):
		return "Email already taken. Please try again with a different email."

	default:
		return "An unexpected internal server error occurred. Please try again later."
	}
}
