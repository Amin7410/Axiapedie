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

// SeedHomePage tạo trang chủ mặc định giới thiệu hệ thống nếu chưa tồn tại
func SeedHomePage(db *sql.DB) error {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM documents WHERE title = 'Home'`).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	markdownContent := `# Chào mừng đến với Axiapedia! 👋

Axiapedia là hệ thống quản lý tri thức nội bộ và wiki cá nhân hiệu năng cao được thiết kế trực quan, hiện đại và bảo mật. Dưới đây là hướng dẫn nhanh giúp bạn làm quen và làm việc hiệu quả với hệ thống:

## 📁 1. Quản lý Thư mục & Tài liệu (File Explorer)
* **Xem cấu trúc cây**: Ở thanh bên trái (Sidebar), bạn sẽ thấy toàn bộ thư mục và tài liệu được sắp xếp phân cấp kiểu IDE.
* **Kéo & Thả (Drag & Drop)**: Bạn có thể kéo thả tài liệu để thay đổi vị trí hoặc chuyển chúng vào trong các thư mục con một cách dễ dàng.
* **Menu Chuột Phải**: Nhấp chuột phải vào bất kỳ mục nào trên cây thư mục để thực hiện các thao tác nhanh: **Đổi tên**, **Xóa**, **Khóa/Mở khóa** (Admin), hoặc **Ẩn/Hiện** (Admin).

## ✏️ 2. Trình soạn thảo Markdown chuyên nghiệp
Nhấp vào nút **Edit** ở góc phải bất kỳ bài viết nào để mở trình soạn thảo:
* **Thanh công cụ trực quan**: Hỗ trợ định dạng nhanh Chữ đậm, Chữ nghiêng, Tiêu đề (H2, H3), Danh sách, Trích dẫn, liên kết và bảng biểu.
* **Bảo toàn lịch sử soạn thảo**: Hỗ trợ đầy đủ phím tắt hoàn tác **Ctrl + Z** (Undo) và làm lại **Ctrl + Y** (Redo).
* **Tải lên hình ảnh**: Kéo và thả ảnh hoặc dán trực tiếp vào trình soạn thảo để tải lên hệ thống.

## 🔐 3. Phân quyền và Bảo mật
Hệ thống hỗ trợ 3 nhóm vai trò chính:
1. **Admin (Quản trị viên)**:
   * Có toàn quyền quản lý thành viên, thay đổi vai trò hoặc xóa tài khoản.
   * **Khóa tài liệu**: Tài liệu bị khóa sẽ chỉ có Admin sửa được (người khác chỉ đọc).
   * **Ẩn tài liệu**: Tài liệu bị ẩn sẽ hoàn toàn vô hình đối với người thường (chỉ Admin nhìn thấy và chỉnh sửa).
2. **Writer (Nhà viết bài)**: Có quyền viết mới, chỉnh sửa các tài liệu công khai chưa bị khóa.
3. **Reader (Người đọc)**: Chỉ có quyền đọc và tìm kiếm tài liệu.

## 🔍 4. Tìm kiếm thông minh
* Sử dụng thanh tìm kiếm ở đầu trang hoặc tổ hợp phím **Ctrl + K** để mở nhanh.
* **Tìm kiếm theo Hashtag**: Bạn có thể gắn thẻ bài viết bằng cách thêm hashtag (ví dụ: '#huongdan', '#quytrinh') ở ô Tags trong trình soạn thảo, sau đó gõ '#huongdan' vào ô Tìm kiếm để lọc nhanh toàn bộ bài viết liên quan.

---
*Trang chủ này được bảo vệ mặc định và chỉ quản trị viên (Admin) mới có quyền chỉnh sửa nội dung.*`

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert into documents
	_, err = tx.Exec(
		`INSERT INTO documents (id, title, subtitle, parent_id, is_folder, is_locked, is_hidden, published_revision_id, latest_revision_id, review_status, sort_order, created_at, updated_at) 
		 VALUES (?, ?, ?, NULL, 0, 1, 0, ?, ?, 'published', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		"home-page-doc", "Home", "Hướng dẫn bắt đầu sử dụng Axiapedia", "home-page-rev", "home-page-rev",
	)
	if err != nil {
		return err
	}

	// 2. Insert into revisions
	_, err = tx.Exec(
		`INSERT INTO revisions (id, document_id, parent_id, author_id, comment, created_at) 
		 VALUES (?, ?, NULL, ?, ?, CURRENT_TIMESTAMP)`,
		"home-page-rev", "home-page-doc", "admin-user", "Initial home page seed",
	)
	if err != nil {
		return err
	}

	// 3. Insert into text_contents
	_, err = tx.Exec(
		`INSERT INTO text_contents (revision_id, content_type, data) 
		 VALUES (?, ?, ?)`,
		"home-page-rev", "full", []byte(markdownContent),
	)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Println("✅ Home page seeded successfully.")
	return nil
}

