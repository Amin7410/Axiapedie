package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
)

func CasbinAuthzMiddleware(enforcer *casbin.Enforcer, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Lấy role thực từ Session, mặc định là guest nếu chưa đăng nhập
		userRole := GetSessionUserRole(r)
		if userRole == "" {
			userRole = "guest"
		}

		obj := r.URL.Path
		act := "read"
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			act = "write"
		}

		// Ánh xạ các đường dẫn API và editor vào Casbin objects
		casbinObj := obj
		if strings.HasPrefix(obj, "/wiki/") {
			casbinObj = "/wiki/*"
		} else if strings.HasPrefix(obj, "/editor/") {
			casbinObj = "/editor/*"
		} else if obj == "/bookmarks" {
			// Yêu cầu đăng nhập đối với trang Bookmarks riêng
			if userRole == "guest" {
				if r.Header.Get("HX-Request") == "true" {
					w.Header().Set("HX-Redirect", "/login")
					w.WriteHeader(http.StatusOK)
					return
				}
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			casbinObj = "/wiki/*"
			act = "read"
		} else if strings.HasPrefix(obj, "/api/v1/explorer/tree") || strings.HasPrefix(obj, "/api/v1/tags") {
			// API đọc danh sách thư mục hoặc gợi ý tag được ánh xạ sang quyền read của wiki
			casbinObj = "/wiki/*"
			act = "read"
		} else if strings.HasPrefix(obj, "/admin/") || strings.HasPrefix(obj, "/api/admin/") {
			// Các tài nguyên và API quản trị hệ thống
			casbinObj = "/admin/*"
			act = "write"
			if r.Method == http.MethodGet {
				act = "read"
			}
		} else if strings.HasPrefix(obj, "/profile") {
			if userRole == "guest" {
				if r.Header.Get("HX-Request") == "true" {
					w.Header().Set("HX-Redirect", "/login")
					w.WriteHeader(http.StatusOK)
					return
				}
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
			return
		} else if strings.HasPrefix(obj, "/api/") {
			// Các API thay đổi dữ liệu được ánh xạ sang quyền write của editor
			casbinObj = "/editor/*"
			act = "write"
		}

		// Bypass cho trang chủ, tìm kiếm, file tĩnh và các tài nguyên public
		if obj == "/" || obj == "/search" || obj == "/robots.txt" ||
			strings.HasPrefix(obj, "/static/") ||
			strings.HasPrefix(obj, "/uploads/") ||
			strings.HasPrefix(obj, "/ui/") {
			next.ServeHTTP(w, r)
			return
		}

		allowed, err := enforcer.Enforce(userRole, casbinObj, act)
		if err != nil {
			log.Printf("Casbin enforce error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !allowed {
			// Nếu là khách chưa đăng nhập thì chuyển hướng tới trang login
			if userRole == "guest" {
				if r.Header.Get("HX-Request") == "true" {
					w.Header().Set("HX-Redirect", "/login")
					w.WriteHeader(http.StatusOK)
					return
				}
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			http.Error(w, "403 Forbidden - You don't have permission to perform this action", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}
