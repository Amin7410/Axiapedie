package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3" // CGO SQLite driver
	_ "github.com/tursodatabase/libsql-client-go/libsql" // Pure Go LibSQL driver
)

// NewSQLiteDB creates a connection to the SQLite database.
func NewSQLiteDB(dsn string) (*sql.DB, error) {
	if dsn == "" {
		dsn = "data/wiki.db"
	}

	var driverName string
	var connString string

	// If DSN is a remote Turso/LibSQL database URL
	if strings.HasPrefix(dsn, "libsql://") || strings.HasPrefix(dsn, "http://") || strings.HasPrefix(dsn, "https://") {
		driverName = "libsql"
		connString = dsn
	} else {
		driverName = "sqlite3"
		connString = fmt.Sprintf("file:%s?cache=shared&mode=rwc&_fk=1&_journal_mode=WAL", dsn)
	}
	
	db, err := sql.Open(driverName, connString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Tự động thêm cột sort_order nếu chưa tồn tại
	_, _ = db.Exec("ALTER TABLE documents ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0")

	return db, nil
}

// Migrate loads the schema from the given file and executes it.
func Migrate(db *sql.DB, schemaPath string) error {
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("could not read schema file: %w", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("could not execute schema: %w", err)
	}

	return nil
}
