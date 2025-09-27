package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/jorge-j1m/hackspark_server/ent"
	session_ent "github.com/jorge-j1m/hackspark_server/ent/session"
	user_ent "github.com/jorge-j1m/hackspark_server/ent/user"
	log "github.com/jorge-j1m/hackspark_server/internal/infrastructure/logger"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/response"
	"github.com/jorge-j1m/hackspark_server/internal/pkg/common/errors"
	"go.jetify.com/typeid/v2"
)

// AuthMiddleware provides session-based authentication middleware
type AuthMiddleware struct {
	client *ent.Client
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(client *ent.Client) *AuthMiddleware {
	return &AuthMiddleware{
		client: client,
	}
}

// Authenticate middleware validates session and adds the user to the request context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, err := m.GetUserFromRequest(ctx, r)
		if err != nil {
			log.Debug(ctx).Err(err).Msg("Failed to get user from request")
			response.Error(w, errors.ErrUserNotFound)
			return
		}

		// Add the user to the context
		ctx = context.WithValue(ctx, log.UserCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) GetUserFromRequest(ctx context.Context, r *http.Request) (*ent.User, error) {
	// Extract the session ID from the header
	sessionID, err := GetSessionFromHeaders(ctx, r.Header)
	if err != nil {
		log.Debug(ctx).Err(err).Msg("Failed to extract session ID from request")
		return nil, errors.ErrSessionInvalid
	}

	// Get the user from the session
	user, err := m.getUserFromSession(ctx, sessionID)
	if err != nil {
		log.Debug(ctx).Err(err).Msg("Failed to get user from session")
		return nil, errors.ErrUserNotFound
	}

	// Check if the user is active
	if user.AccountStatus == user_ent.AccountStatusSuspended {
		log.Debug(ctx).
			Str("user_id", user.ID).
			Str("status", string(user.AccountStatus)).
			Msg("User account is suspended")
		return nil, errors.ErrAccountSuspended
	}

	return user, nil
}

func GetSessionFromHeaders(ctx context.Context, h http.Header) (string, error) {
	// Extract the session ID from the Authorization header
	authHeader := h.Get("Authorization")

	if authHeader == "" {
		log.Debug(ctx).Msg("Missing Authorization header")
		return "", errors.ErrSessionInvalid
	}

	// Check if the header is properly formatted (Bearer <session_id>)
	const prefix = "Bearer "
	if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
		return "", errors.ErrSessionInvalid
	}

	// Extract the session ID without the "Bearer " prefix
	headerValue := authHeader[len(prefix):]
	sessionId, err := typeid.Parse(headerValue)
	if err != nil {
		log.Debug(ctx).Err(err).Msg("Failed to parse session ID")
		return "", errors.ErrSessionInvalid
	}

	return sessionId.String(), nil
}

// GetUserIDFromContext extracts the user ID from the authenticated request context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	user, ok := ctx.Value(log.UserCtxKey).(*ent.User)
	if !ok || user == nil {
		return "", errors.ErrUserNotFound
	}
	return user.ID, nil
}

// getUserFromSession
func (m *AuthMiddleware) getUserFromSession(ctx context.Context, sessionID string) (*ent.User, error) {
	// Query the session by its ID, ensuring it has not expired.
	// Then, traverse the graph to the owner (the user).
	// The `Only` method ensures that exactly one user is returned,
	// or returns a friendly error (ent.NotFoundError) if the session is invalid.
	return m.client.Session.Query().
		Where(
			session_ent.ID(sessionID),
			session_ent.ExpiresAtGT(time.Now()),
		).
		QueryUser(). // Traverse to the User edge.
		Only(ctx)
}

// OptionalAuth middleware attempts to authenticate but allows requests to proceed even if authentication fails
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, err := m.GetUserFromRequest(ctx, r)
		if err != nil {
			log.Debug(ctx).Err(err).Msg("Failed to get user from request")
			next.ServeHTTP(w, r)
			return
		}

		// Add the user to the context
		ctx = context.WithValue(ctx, log.UserCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
