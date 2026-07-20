package handlers

// ErrorResponse represents the standardized JSON structure for all API errors.
type ErrorResponse struct {
	Error string `json:"error"`
}
