// Package handlers implements the HTTP request handlers for the router.
package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"

	"github.com/yudan-glitch/twitter-backend/internal/domain"
)

var (
	// ErrInvalidUsername indicates that the provided username violates validation constraints.
	ErrInvalidUsername = errors.New("invalid username")
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")
)

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

		// Decode user payload
		var incomingUserData UserRequest
		err := json.NewDecoder(r.Body).Decode(&incomingUserData)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err)
			return
		}

		// Payload Validation:
		// Username
		lenUsername := len(incomingUserData.Username)
		if lenUsername < 3 || lenUsername > 20 {
			respondWithError(w, http.StatusBadRequest, ErrInvalidUsername)
			return
		}

		// Email
		_, err = mail.ParseAddress(incomingUserData.Email)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, ErrInvalidEmail)
			return
		}

		// Password
		if len(incomingUserData.Password) == 0 {
			respondWithError(w, http.StatusBadRequest, ErrInvalidPassword)
			return
		}

		// Map UserRequest to domain User
		user := domain.User{
			Username:     incomingUserData.Username,
			Email:        incomingUserData.Email,
			PasswordHash: incomingUserData.Password,
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
	case errors.Is(err, ErrInvalidUsername):
		// I will eventually add more input constraints e.g. allowing only specific characters.
		return "The username must be between 3 and 20 characters long. Please provide a valid username."
	case errors.Is(err, ErrInvalidEmail):
		return "Invalid email. Please try again with a different email."

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
