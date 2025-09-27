package users

import (
	"net/http"

	"github.com/jorge-j1m/hackspark_server/ent"
	log "github.com/jorge-j1m/hackspark_server/internal/infrastructure/logger"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/response"
	"github.com/jorge-j1m/hackspark_server/internal/pkg/common/errors"
)

// UserResponse represents the filtered user data returned in API responses
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Status    string `json:"status"`
}

// convertUserToResponse converts ent.User to UserResponse, filtering out sensitive data
func convertUserToResponse(user *ent.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		Status:    string(user.AccountStatus),
	}
}

func (u *UsersHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value(log.UserCtxKey).(*ent.User)
	if !ok || user == nil {
		log.Debug(ctx).Msg("Failed to get user from context")
		response.Error(w, errors.ErrUserNotFound)
		return
	}

	response.JSON(w, http.StatusOK, "User fetched successfully", convertUserToResponse(user))
}
