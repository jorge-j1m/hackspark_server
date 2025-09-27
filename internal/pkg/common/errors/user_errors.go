package errors

var (
	// CRUD
	ErrUserCreationFailed = NewInternalError("Failed to create user")
	ErrUserNotFound       = NewNotFoundError("User not found")

	// Auth
	ErrInvalidSignupData = NewBadRequestError("Invalid signup data provided")
	ErrInvalidLoginData  = NewBadRequestError("Invalid login data")
	ErrLoginFailed       = NewInternalError("Failed to login")
	ErrInvalidEmail      = NewBadRequestError("Email is required")
	// Used for password format validation
	ErrInvalidPassword = NewBadRequestError("Password is required")
	// Used for password comparison
	ErrWrongPassword = NewAuthenticationError("Wrong password")

	// Access
	ErrNoPermission     = NewForbiddenError("user does not have permission")
	ErrUserInactive     = NewForbiddenError("user account is not active")
	ErrAccountInactive  = NewAuthorizationError("User account is inactive")
	ErrAccountSuspended = NewForbiddenError("Account suspended")
)
