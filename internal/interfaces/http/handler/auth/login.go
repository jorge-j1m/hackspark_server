package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jorge-j1m/hackspark_server/ent"
	user_ent "github.com/jorge-j1m/hackspark_server/ent/user"
	log "github.com/jorge-j1m/hackspark_server/internal/infrastructure/logger"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/response"
	"github.com/jorge-j1m/hackspark_server/internal/pkg/common/errors"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

func (l LoginRequest) Validate() error {
	// DoS protection - check for oversized inputs
	if len(l.Email) > 254 {
		return errors.ErrInvalidEmail
	}
	if len(l.Password) > 1000 {
		return errors.ErrInvalidPassword
	}

	if !isValidEmail(l.Email) {
		return errors.ErrInvalidEmail
	}
	if !isValidPassword(l.Password) {
		return errors.ErrInvalidPassword
	}
	return nil
}

type LoginSuccessData struct {
	Id        string `json:"id"`        // User ID
	SessionID string `json:"sessionId"` // Session ID
	Email     string `json:"email"`     // User Email
	FirstName string `json:"firstName"` // User First Name
	LastName  string `json:"lastName"`  // User Last Name
	Username  string `json:"username"`  // User Username
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var loginData LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
		log.Error(ctx).Err(err).Msg("Failed to decode request body")
		response.Error(w, errors.ErrInvalidLoginData)
		return
	}

	if err := loginData.Validate(); err != nil {
		log.Error(ctx).Err(err).Msg("Invalid login data")
		response.Error(w, errors.ErrInvalidLoginData)
		return
	}

	// Get user by email
	user, err := h.client.User.Query().Where(user_ent.Email(loginData.Email)).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Error(ctx).Err(err).Msg("Failed to find user by email")
			response.Error(w, errors.ErrUserNotFound)
			return
		}
		log.Error(ctx).Err(err).Msg("Failed to find user by email")
		response.Error(w, errors.ErrLoginFailed)
		return
	}

	// Check account suspended or not
	if user.AccountStatus == user_ent.AccountStatusSuspended {
		log.Error(ctx).Msg("User account is suspended")
		response.Error(w, errors.ErrAccountSuspended)
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
		log.Error(ctx).Err(err).Msg("Failed to compare password")
		response.Error(w, errors.ErrWrongPassword)
		return
	}

	// Create temp_session
	temp_session := h.client.Session.Create().SetUser(user).SetIPAddress(r.RemoteAddr).SetUserAgent(r.UserAgent())
	if loginData.Remember {
		temp_session.SetExpiresAt(time.Now().Add(30 * 24 * time.Hour))
	}

	session, err := temp_session.Save(ctx)
	if err != nil {
		log.Error(ctx).Err(err).Msg("Failed to create session")
		response.Error(w, errors.ErrSessionCreateFailed)
		return
	}

	// Update last login
	if _, err := h.client.User.UpdateOneID(user.ID).SetLastLoginAt(time.Now()).Save(ctx); err != nil {
		log.Error(ctx).Err(err).Msg("Failed to update last login")
		// Won't fail the login since it's not critical
	}

	log.Info(ctx).Msgf("User logged in successfully: %s", user.Email)
	response.JSON(w, http.StatusOK, "User logged in successfully", LoginSuccessData{
		Id:        user.ID,
		SessionID: session.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
	})
}
