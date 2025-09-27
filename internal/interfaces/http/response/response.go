package response

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/jorge-j1m/hackspark_server/internal/pkg/common/errors"
)

// Response represents the standard API response structure
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// JSON sends a JSON response with the given status code and data
func JSON(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data == nil {
		return
	}

	resp := Response{
		Success: statusCode >= 200 && statusCode < 300,
		Message: message,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

// Error sends an error response with the appropriate status code
func Error(w http.ResponseWriter, err *errors.AppError) {
	// TODO: Check if it's internal error and if so, hide the error message

	// Determine HTTP status code
	statusCode := err.HTTPStatus
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := Response{
		Success: false,
		Message: err.Message,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error().Err(err).Msg("Failed to encode error response")
	}
}
