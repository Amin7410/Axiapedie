package html

import (
	"html/template"
	"log"
	"net/http"

	"axia-wiki/internal/domain"
)

type AuthHandler struct {
	authUsecase domain.AuthUsecase
	tmpl        *template.Template
}

func NewAuthHandler(authUsecase domain.AuthUsecase) *AuthHandler {
	tmpl := template.Must(template.ParseFiles("web/templates/layout.html", "web/templates/login.html"))
	return &AuthHandler{
		authUsecase: authUsecase,
		tmpl:        tmpl,
	}
}

// LoginPage handles GET /login
func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Error": "",
	}
	h.tmpl.ExecuteTemplate(w, "login.html", data)
}

// LoginSubmit handles POST /login
func (h *AuthHandler) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := h.authUsecase.Login(r.Context(), username, password)
	if err != nil {
		log.Printf("Login failed for user '%s': %v", username, err)
		data := map[string]interface{}{
			"Error": "Invalid username or password.",
		}
		h.tmpl.ExecuteTemplate(w, "login.html", data)
		return
	}

	// Đăng nhập thành công - Set cookie session
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user_id",
		Value:    user.ID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7, // 7 ngày
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user_role",
		Value:    user.Role,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "session_username",
		Value:    user.Username,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7,
	})

	log.Printf("✅ User '%s' logged in successfully (role: %s)", user.Username, user.Role)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout handles GET /logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Xóa cookies
	http.SetCookie(w, &http.Cookie{
		Name:   "session_user_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "session_user_role",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "session_username",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
