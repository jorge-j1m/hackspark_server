package auth

import (
	"encoding/json"
	"net/http"

	log "github.com/jorge-j1m/hackspark_server/internal/infrastructure/logger"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/response"
	"github.com/jorge-j1m/hackspark_server/internal/pkg/common/errors"
)

type SignUpRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error(r.Context()).Err(err).Msg("Failed to decode request body")
		response.Error(w, errors.ErrInvalidSignupData)
		return
	}

	user, err := h.client.User.Create().
		SetFirstName(req.FirstName).
		SetLastName(req.LastName).
		SetUsername(req.Username).
		SetEmail(req.Email).
		SetPassword(req.Password).
		Save(r.Context())
	if err != nil {
		log.Error(r.Context()).Err(err).Msg("Failed to create user")
		response.Error(w, errors.ErrUserCreationFailed)
		return
	}

	log.Info(r.Context()).Msgf("User created successfully: %s", user.ID)
	response.JSON(w, http.StatusCreated, "User created successfully", map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.FirstName + " " + user.LastName,
	})
}
