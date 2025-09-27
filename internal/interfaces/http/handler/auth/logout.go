package auth

import (
	"net/http"

	log "github.com/jorge-j1m/hackspark_server/internal/infrastructure/logger"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/middleware"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/response"
	"github.com/jorge-j1m/hackspark_server/internal/pkg/common/errors"
)

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract the session ID from the Authorization header
	sessionID, err := middleware.GetSessionFromHeaders(ctx, r.Header)
	if err != nil {
		log.Debug(ctx).Err(err).Msg("Failed to extract session ID from request")
		response.Error(w, errors.ErrSessionNotFound)
		return
	}

	// Invalidate the session (delete it)
	err = h.client.Session.DeleteOneID(sessionID).Exec(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Something went wrong logging out user")
		response.Error(w, errors.ErrSessionInvalid)
		return
	}

	log.Info(ctx).Msgf("User logged out successfully: %s", sessionID)
	response.JSON(w, http.StatusOK, "User logged out successfully", nil)
}
