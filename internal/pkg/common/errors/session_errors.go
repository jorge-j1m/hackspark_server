package errors

// Session-related errors
var (
	// ErrSessionGeneration is returned when session generation fails
	ErrSessionGeneration = NewInternalError("failed to generate session")

	// ErrSessionExpired is returned when a session has expired
	ErrSessionExpired = NewAuthenticationError("session has expired")

	// ErrSessionRefreshFailed is returned when a session refresh fails
	ErrSessionRefreshFailed = NewInternalError("failed to refresh session")

	// ErrSessionInvalid is returned when a session is invalid
	ErrSessionInvalid = NewAuthenticationError("invalid session")

	// ErrSessionInvalidated is returned when a session has been invalidated
	ErrSessionInvalidated = NewForbiddenError("session has been invalidated")

	// ErrSessionInvalidationFailed is returned when a session invalidation fails
	ErrSessionInvalidationFailed = NewInternalError("failed to invalidate session")

	// ErrSessionCleanupFailed is returned when a session cleanup fails
	ErrSessionCleanupFailed = NewInternalError("failed to cleanup sessions")

	// ErrSessionCreateFailed is returned when a session cannot be created
	ErrSessionCreateFailed = NewInternalError("failed to create session")

	// ErrSessionValidationFailed is returned when a session validation fails
	ErrSessionValidationFailed = NewAuthenticationError("session validation failed")

	// ErrSessionNotFound is returned when no session is provided
	ErrSessionNotFound = NewNotFoundError("session not found")

	// ErrInvalidCredentials is returned when the provided credentials are invalid
	ErrInvalidCredentials = NewAuthenticationError("invalid credentials")

	// Note: The following errors are defined in user_errors.go:
	// - ErrUserNotFound
	// - ErrNoPermission
	// - ErrAccountSuspended
)
