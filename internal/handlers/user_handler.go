// Package handlers implements the HTTP request handlers for the router.
package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/yudan-glitch/twitter-backend/internal/domain"
)

// ErrInvalidUsername indicates that the provided username violates validation constraints.
var ErrInvalidUsername = errors.New("invalid username")

// UserResponse is the clean, public Data Transfer Object (DTO).
// It only includes fields that are 100% safe to send over the network.
// HTTP JSON payload shapes live here.
type UserResponse struct {
	Username string `json:"username"`
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
			respondWithError(w, http.StatusInternalServerError, domain.ErrInternalServer)
			return
		}

		// Success! Map the user to UserResponse and send it back with 200 OK.
		userResp := UserResponse{Username: user.Username}
		respondWithJSON(w, http.StatusOK, userResp)
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
	case errors.Is(err, domain.ErrUserNotFound):
		return "We couldn't find an account matching that username. Please verify the spelling and try again."
	default:
		return "An unexpected internal server error occurred. Please try again later."
	}
}
