package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"axia-wiki/internal/domain"
)

var cookieSecret []byte

func init() {
	secretPath := filepath.Join("data", ".secret_key")
	// Đảm bảo thư mục data tồn tại
	_ = os.MkdirAll("data", 0755)

	data, err := os.ReadFile(secretPath)
	if err == nil && len(data) == 32 {
		cookieSecret = data
		return
	}

	// Tạo key bí mật ngẫu nhiên và lưu lại
	cookieSecret = make([]byte, 32)
	_, _ = rand.Read(cookieSecret)
	_ = os.WriteFile(secretPath, cookieSecret, 0600)
}

// SignValue ký giá trị bằng HMAC-SHA256
func SignValue(value string) string {
	h := hmac.New(sha256.New, cookieSecret)
	h.Write([]byte(value))
	sig := hex.EncodeToString(h.Sum(nil))
	return value + "." + sig
}

// VerifyValue xác minh chữ ký của giá trị
func VerifyValue(signedValue string) (string, bool) {
	parts := strings.SplitN(signedValue, ".", 2)
	if len(parts) != 2 {
		return "", false
	}
	val, sig := parts[0], parts[1]

	h := hmac.New(sha256.New, cookieSecret)
	h.Write([]byte(val))
	expectedSig := hex.EncodeToString(h.Sum(nil))

	if hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return val, true
	}
	return "", false
}

// GenerateCSRFToken tạo CSRF token ngẫu nhiên
func GenerateCSRFToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// SessionMiddleware kiểm tra cookie session và gắn user info vào context, đồng thời bảo vệ CSRF
func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Kiểm tra CSRF Token cho các yêu cầu thay đổi dữ liệu
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			cookieToken, err := r.Cookie("csrf_token")
			if err != nil || cookieToken.Value == "" {
				http.Error(w, "Forbidden - CSRF token missing", http.StatusForbidden)
				return
			}

			requestToken := r.Header.Get("X-CSRF-Token")
			if requestToken == "" {
				requestToken = r.FormValue("csrf_token")
			}

			if requestToken == "" || requestToken != cookieToken.Value {
				http.Error(w, "Forbidden - CSRF token mismatch", http.StatusForbidden)
				return
			}
		}

		// 2. Thiết lập hoặc lấy CSRF token cho context và cookie
		csrfToken := ""
		csrfCookie, err := r.Cookie("csrf_token")
		if err != nil || csrfCookie.Value == "" {
			csrfToken = GenerateCSRFToken()
			http.SetCookie(w, &http.Cookie{
				Name:     "csrf_token",
				Value:    csrfToken,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
		} else {
			csrfToken = csrfCookie.Value
		}

		// 3. Giải mã và xác thực thông tin Session từ Cookies có chữ ký bảo mật
		userID := ""
		userRole := "guest"
		username := ""

		cookie, err := r.Cookie("session_user_id")
		if err == nil && cookie.Value != "" {
			if val, ok := VerifyValue(cookie.Value); ok {
				userID = val
			}
		}

		roleCookie, err := r.Cookie("session_user_role")
		if err == nil && roleCookie.Value != "" {
			if val, ok := VerifyValue(roleCookie.Value); ok {
				userRole = val
			}
		}

		userCookie, err := r.Cookie("session_username")
		if err == nil && userCookie.Value != "" {
			if val, ok := VerifyValue(userCookie.Value); ok {
				username = val
			}
		}

		ctx := context.WithValue(r.Context(), domain.ContextUserIDKey, userID)
		ctx = context.WithValue(ctx, domain.ContextUserRoleKey, userRole)
		ctx = context.WithValue(ctx, domain.ContextUsernameKey, username)
		ctx = context.WithValue(ctx, domain.ContextCSRFTokenKey, csrfToken)
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

// GetCSRFToken lấy CSRF token từ context
func GetCSRFToken(r *http.Request) string {
	if val, ok := r.Context().Value(domain.ContextCSRFTokenKey).(string); ok {
		return val
	}
	return ""
}


