package html

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"regexp"

	"axia-wiki/internal/domain"
	"axia-wiki/internal/middleware"

	"golang.org/x/oauth2"
)

type AuthHandler struct {
	authUsecase domain.AuthUsecase
	templates   map[string]*template.Template
	oauthConfig *oauth2.Config
}

func NewAuthHandler(authUsecase domain.AuthUsecase, oauthConfig *oauth2.Config) *AuthHandler {
	layoutFile := "web/templates/layout.html"
	pages := []string{"login.html", "register.html", "profile.html", "admin_users.html"}

	templates := make(map[string]*template.Template)
	for _, page := range pages {
		tmpl := template.Must(template.ParseFiles(layoutFile, "web/templates/"+page))
		templates[page] = tmpl
	}

	return &AuthHandler{
		authUsecase: authUsecase,
		templates:   templates,
		oauthConfig: oauthConfig,
	}
}

// LoginPage handles GET /login
func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	msg := ""
	if r.URL.Query().Get("registered") == "true" {
		msg = "Registration successful! Please sign in with your new account."
	}
	data := map[string]interface{}{
		"Error":          "",
		"SuccessMessage": msg,
	}
	h.templates["login.html"].ExecuteTemplate(w, "login.html", data)
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
		h.templates["login.html"].ExecuteTemplate(w, "login.html", data)
		return
	}

	// Đăng nhập thành công - Set cookie session
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user_id",
		Value:    user.ID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7, // 7 ngày
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user_role",
		Value:    user.Role,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "session_username",
		Value:    user.Username,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7,
		SameSite: http.SameSiteLaxMode,
	})

	log.Printf("✅ User '%s' logged in successfully (role: %s)", user.Username, user.Role)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// RegisterPage handles GET /register
func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Error":    "",
		"Username": "",
	}
	h.templates["register.html"].ExecuteTemplate(w, "register.html", data)
}

// RegisterSubmit handles POST /register
func (h *AuthHandler) RegisterSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	// 1. Kiểm tra mật khẩu khớp
	if password != confirmPassword {
		h.renderRegisterError(w, "Passwords do not match.", username)
		return
	}

	// 2. Kiểm tra độ dài mật khẩu (tối thiểu 8, tối đa 72)
	if len(password) < 8 {
		h.renderRegisterError(w, "Password must be at least 8 characters long.", username)
		return
	}
	if len(password) > 72 {
		h.renderRegisterError(w, "Password is too long (maximum 72 characters).", username)
		return
	}

	// 3. Kiểm tra định dạng tên đăng nhập (chỉ chữ không dấu, số, gạch dưới, từ 3-20 ký tự)
	rxUsername := regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
	if !rxUsername.MatchString(username) {
		h.renderRegisterError(w, "Invalid username. Only alphanumeric characters and underscores are allowed (3-20 characters).", username)
		return
	}

	// 4. Gọi Usecase đăng ký
	user, err := h.authUsecase.Register(r.Context(), username, password)
	if err != nil {
		h.renderRegisterError(w, "Registration failed: "+err.Error(), username)
		return
	}

	// 5. Đăng ký thành công - Điều hướng sang trang Đăng nhập và báo thành công
	log.Printf("✅ User '%s' registered successfully", user.Username)
	http.Redirect(w, r, "/login?registered=true", http.StatusSeeOther)
}

func (h *AuthHandler) renderRegisterError(w http.ResponseWriter, errMsg string, username string) {
	data := map[string]interface{}{
		"Error":    errMsg,
		"Username": username,
	}
	h.templates["register.html"].ExecuteTemplate(w, "register.html", data)
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

func (h *AuthHandler) render(w http.ResponseWriter, r *http.Request, page string, data interface{}) {
	tmpl, ok := h.templates[page]
	if !ok {
		http.Error(w, "Template not found: "+page, http.StatusInternalServerError)
		return
	}

	var title string
	if m, ok := data.(map[string]interface{}); ok {
		m["SessionUserID"] = middleware.GetSessionUserID(r)
		m["SessionUserRole"] = middleware.GetSessionUserRole(r)
		m["SessionUsername"] = middleware.GetSessionUsername(r)
		if t, ok := m["Title"].(string); ok {
			title = t
		}
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if title != "" {
			w.Write([]byte("<title>" + title + " - Axia Wiki</title>"))
		}
		if err := tmpl.ExecuteTemplate(w, "content", data); err != nil {
			http.Error(w, "Template render error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := tmpl.ExecuteTemplate(w, page, data); err != nil {
		http.Error(w, "Template render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ProfilePage handles GET /profile
func (h *AuthHandler) ProfilePage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetSessionUserID(r)
	if userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.authUsecase.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"Title": "Profile",
		"User":  user,
		"Error": "",
		"Success": "",
	}
	h.render(w, r, "profile.html", data)
}

// ChangePasswordSubmit handles POST /profile/change-password
func (h *AuthHandler) ChangePasswordSubmit(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetSessionUserID(r)
	if userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.authUsecase.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	oldPassword := r.FormValue("old_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	// 1. Kiểm tra mật khẩu khớp
	if newPassword != confirmPassword {
		h.renderProfileError(w, r, user, "New passwords do not match.", "")
		return
	}

	// 2. Kiểm tra độ dài mật khẩu (tối thiểu 8)
	if len(newPassword) < 8 {
		h.renderProfileError(w, r, user, "New password must be at least 8 characters long.", "")
		return
	}

	// 3. Thực hiện đổi mật khẩu
	if err := h.authUsecase.ChangePassword(r.Context(), userID, oldPassword, newPassword); err != nil {
		h.renderProfileError(w, r, user, err.Error(), "")
		return
	}

	h.renderProfileError(w, r, user, "", "Password changed successfully!")
}

func (h *AuthHandler) renderProfileError(w http.ResponseWriter, r *http.Request, user *domain.User, errMsg string, successMsg string) {
	data := map[string]interface{}{
		"Title":   "Profile",
		"User":    user,
		"Error":   errMsg,
		"Success": successMsg,
	}
	h.render(w, r, "profile.html", data)
}

// AdminUsersPage handles GET /admin/users
func (h *AuthHandler) AdminUsersPage(w http.ResponseWriter, r *http.Request) {
	users, err := h.authUsecase.GetAllUsers(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title": "User Management",
		"Users": users,
	}
	h.render(w, r, "admin_users.html", data)
}

// AdminUpdateUserRole handles POST /api/admin/users/role
func (h *AuthHandler) AdminUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	currentAdminID := middleware.GetSessionUserID(r)
	targetUserID := r.FormValue("user_id")
	newRole := r.FormValue("role")

	// Ràng buộc bảo mật: Không được tự hạ vai trò admin của chính mình
	if currentAdminID == targetUserID {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("<script>showToast('You cannot demote yourself.', 'error');</script>"))
		return
	}

	err := h.authUsecase.UpdateUserRole(r.Context(), targetUserID, newRole)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("<script>showToast('Failed to update role: " + err.Error() + "', 'error');</script>"))
		return
	}

	w.Write([]byte("<script>showToast('User role updated successfully.');</script>"))
}

// AdminDeleteUser handles POST /api/admin/users/delete
func (h *AuthHandler) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	currentAdminID := middleware.GetSessionUserID(r)
	targetUserID := r.FormValue("user_id")

	// Ràng buộc bảo mật: Không được tự xóa tài khoản của chính mình
	if currentAdminID == targetUserID {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("<script>showToast('You cannot delete your own account.', 'error');</script>"))
		return
	}

	err := h.authUsecase.DeleteUser(r.Context(), targetUserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("<script>showToast('Failed to delete user: " + err.Error() + "', 'error');</script>"))
		return
	}

	// Trả về lệnh Javascript để xóa hàng (row) tương ứng trên UI
	w.Write([]byte("<script>showToast('User deleted successfully.'); document.getElementById('user-row-" + targetUserID + "').remove();</script>"))
}

// GoogleLogin handles GET /auth/google
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.oauthConfig == nil {
		data := map[string]interface{}{
			"Error": "Google Authentication has not been configured by the system administrator.",
		}
		h.templates["login.html"].ExecuteTemplate(w, "login.html", data)
		return
	}

	state := "state-token-axia"
	url := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles GET /auth/google/callback
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.oauthConfig == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	state := r.FormValue("state")
	if state != "state-token-axia" {
		log.Printf("Invalid OAuth state: %s", state)
		data := map[string]interface{}{
			"Error": "Invalid security state returned from Google.",
		}
		h.templates["login.html"].ExecuteTemplate(w, "login.html", data)
		return
	}

	code := r.FormValue("code")
	token, err := h.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("OAuth Exchange failed: %v", err)
		data := map[string]interface{}{
			"Error": "Failed to exchange authorization code with Google.",
		}
		h.templates["login.html"].ExecuteTemplate(w, "login.html", data)
		return
	}

	client := h.oauthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("Failed to get Google user info: %v", err)
		data := map[string]interface{}{
			"Error": "Failed to retrieve user profile information from Google.",
		}
		h.templates["login.html"].ExecuteTemplate(w, "login.html", data)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Printf("Failed to parse Google user info: %v", err)
		data := map[string]interface{}{
			"Error": "Failed to parse user profile information from Google.",
		}
		h.templates["login.html"].ExecuteTemplate(w, "login.html", data)
		return
	}

	// Sign in or create user
	user, err := h.authUsecase.LoginOrCreateWithGoogle(r.Context(), userInfo.ID, userInfo.Email, userInfo.Name)
	if err != nil {
		log.Printf("Google Login/Create failed: %v", err)
		data := map[string]interface{}{
			"Error": "Database authentication failed: " + err.Error(),
		}
		h.templates["login.html"].ExecuteTemplate(w, "login.html", data)
		return
	}

	// Đăng nhập thành công - Set cookie session
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user_id",
		Value:    user.ID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user_role",
		Value:    user.Role,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "session_username",
		Value:    user.Username,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7,
		SameSite: http.SameSiteLaxMode,
	})

	log.Printf("✅ Google User '%s' logged in successfully (role: %s)", user.Username, user.Role)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
