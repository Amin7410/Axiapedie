package db

import (
	"database/sql"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// SeedDefaultAdmin tạo tài khoản admin mặc định nếu chưa tồn tại
func SeedDefaultAdmin(db *sql.DB) error {
	// Kiểm tra xem đã có user admin chưa
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE username = 'admin'`).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("Admin user already exists, skipping seed.")
		return nil
	}

	// Tạo password hash cho tài khoản admin mặc định
	// Mật khẩu mặc định: admin123 (Người dùng nên đổi ngay sau khi đăng nhập)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT INTO users (id, username, password_hash, role) VALUES (?, ?, ?, ?)`,
		"admin-user", "admin", string(passwordHash), "admin",
	)
	if err != nil {
		return err
	}

	log.Println("✅ Default admin user created (username: admin, password: admin123)")
	return nil
}

// SeedTestAccounts tạo các tài khoản test để kiểm tra phân quyền
func SeedTestAccounts(db *sql.DB) error {
	accounts := []struct {
		ID       string
		Username string
		Password string
		Role     string
	}{
		{"writer-user", "writer", "writer123", "writer"},
		{"reader-user", "reader", "reader123", "reader"},
	}

	for _, acc := range accounts {
		var count int
		err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE username = ?`, acc.Username).Scan(&count)
		if err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(acc.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		_, err = db.Exec(
			`INSERT INTO users (id, username, password_hash, role) VALUES (?, ?, ?, ?)`,
			acc.ID, acc.Username, string(passwordHash), acc.Role,
		)
		if err != nil {
			return err
		}

		log.Printf("✅ Test account created (username: %s, password: %s, role: %s)", acc.Username, acc.Password, acc.Role)
	}

	return nil
}

// SeedUnsortedBin tạo thư mục Unsorted Bin đặc biệt nếu chưa tồn tại
func SeedUnsortedBin(db *sql.DB) error {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM documents WHERE id = 'unsorted_bin_folder'`).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	_, err = db.Exec(
		`INSERT INTO documents (id, title, parent_id, is_folder, is_locked, review_status, created_at, updated_at) 
		 VALUES (?, ?, NULL, 1, 1, 'published', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		"unsorted_bin_folder", "Unsorted Bin",
	)
	if err != nil {
		return err
	}

	log.Println("✅ Unsorted Bin folder initialized.")
	return nil
}

