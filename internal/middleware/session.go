package middleware

import (
	"context"
	"net/http"

	"axia-wiki/internal/domain"
)

// SessionMiddleware kiểm tra cookie session và gắn user info vào context
func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := ""
		userRole := "guest"
		username := ""

		cookie, err := r.Cookie("session_user_id")
		if err == nil && cookie.Value != "" {
			userID = cookie.Value
		}

		roleCookie, err := r.Cookie("session_user_role")
		if err == nil && roleCookie.Value != "" {
			userRole = roleCookie.Value
		}

		userCookie, err := r.Cookie("session_username")
		if err == nil && userCookie.Value != "" {
			username = userCookie.Value
		}

		ctx := context.WithValue(r.Context(), domain.ContextUserIDKey, userID)
		ctx = context.WithValue(ctx, domain.ContextUserRoleKey, userRole)
		ctx = context.WithValue(ctx, domain.ContextUsernameKey, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetSessionUserID lấy user ID từ context
func GetSessionUserID(r *http.Request) string {
	if val, ok := r.Context().Value(domain.ContextUserIDKey).(string); ok {
		return val
	}
	return ""
}

// GetSessionUserRole lấy role từ context
func GetSessionUserRole(r *http.Request) string {
	if val, ok := r.Context().Value(domain.ContextUserRoleKey).(string); ok {
		return val
	}
	return "guest"
}

// GetSessionUsername lấy username từ context
func GetSessionUsername(r *http.Request) string {
	if val, ok := r.Context().Value(domain.ContextUsernameKey).(string); ok {
		return val
	}
	return ""
}

