package domain

type contextKey string

const (
	ContextUserIDKey   contextKey = "user_id"
	ContextUserRoleKey contextKey = "user_role"
	ContextUsernameKey contextKey = "username"
	ContextCSRFTokenKey contextKey = "csrf_token"
)
