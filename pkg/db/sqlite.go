package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3" // CGO SQLite driver
	_ "github.com/tursodatabase/libsql-client-go/libsql" // Pure Go LibSQL driver
)

// IsCloudDSN returns true if the DSN points to a remote Turso/LibSQL database.
func IsCloudDSN(dsn string) bool {
	return strings.HasPrefix(dsn, "libsql://") || strings.HasPrefix(dsn, "http://") || strings.HasPrefix(dsn, "https://")
}

// NewSQLiteDB creates a connection to the SQLite database.
func NewSQLiteDB(dsn string) (*sql.DB, error) {
	if dsn == "" {
		dsn = "data/wiki.db"
	}

	var driverName string
	var connString string

	// If DSN is a remote Turso/LibSQL database URL
	if IsCloudDSN(dsn) {
		driverName = "libsql"
		connString = dsn
	} else {
		driverName = "sqlite3"
		connString = fmt.Sprintf("file:%s?cache=shared&mode=rwc&_fk=1&_journal_mode=WAL&_busy_timeout=5000", dsn)
	}
	
	db, err := sql.Open(driverName, connString)
	if err != nil {
		return nil, err
	}

	// Giới hạn kết nối tối đa cho SQLite cục bộ để tránh lỗi 'database is locked' khi có ghi đồng thời
	if !IsCloudDSN(dsn) {
		db.SetMaxOpenConns(1)
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Tự động thêm cột sort_order nếu chưa tồn tại
	_, _ = db.Exec("ALTER TABLE documents ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0")
	// Tự động thêm cột google_id cho chức năng đăng nhập Google
	_, _ = db.Exec("ALTER TABLE users ADD COLUMN google_id TEXT")
	_, _ = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id)")

	// Tự động dọn dẹp và xây dựng lại chỉ mục FTS5 sạch sẽ dựa trên các bản Live hiện hành
	if !IsCloudDSN(dsn) {
		log.Println("Optimizing and rebuilding FTS5 search index to clean up stale revisions...")
		_, _ = db.Exec("DELETE FROM documents_fts;")
		_, _ = db.Exec(`
			INSERT INTO documents_fts (document_id, title, content)
			SELECT d.id, d.title, CAST(tc.data AS TEXT)
			FROM documents d
			JOIN revisions r ON d.published_revision_id = r.id
			JOIN text_contents tc ON r.id = tc.revision_id;
		`)
	}

	return db, nil
}

// Migrate loads the schema from the given file and executes it.
// For LibSQL/Turso compatibility, statements are executed one at a time.
func Migrate(db *sql.DB, schemaPath string) error {
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("could not read schema file: %w", err)
	}

	// Nếu là local SQLite, chạy toàn bộ file cùng lúc để hỗ trợ trigger chứa dấu chấm phẩy
	if !strings.Contains(schemaPath, "schema_turso.sql") {
		if _, err := db.Exec(string(schema)); err != nil {
			return fmt.Errorf("could not execute schema: %w", err)
		}
		return nil
	}

	// Nếu là Turso/LibSQL Cloud, phân tách và chạy từng câu lệnh một
	statements := strings.Split(string(schema), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		// Skip comment-only lines
		lines := strings.Split(stmt, "\n")
		hasCode := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
				hasCode = true
				break
			}
		}
		if !hasCode {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("could not execute statement: %w\nStatement: %.100s", err, stmt)
		}
	}

	return nil
}
