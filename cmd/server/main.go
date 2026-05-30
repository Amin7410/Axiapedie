package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"axia-wiki/internal/delivery/api"
	"axia-wiki/internal/delivery/html"
	"axia-wiki/internal/middleware"
	"axia-wiki/internal/repository/sqlite"
	"axia-wiki/internal/repository/storage"
	"axia-wiki/internal/usecase"
	"axia-wiki/pkg/db"

	"github.com/casbin/casbin/v2"
)

func main() {
	// Initialize Database Connection (SQLite)
	log.Println("Initializing database...")
	dbDSN := os.Getenv("DB_DSN")
	database, err := db.NewSQLiteDB(dbDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Auto-migrate schema
	log.Println("Running auto-migration...")
	schemaPath := "db/migrations/schema.sql"
	if db.IsCloudDSN(dbDSN) {
		schemaPath = "db/migrations/schema_turso.sql"
		log.Println("Using Turso-compatible schema (no PRAGMA/FTS5/TRIGGER)")
	}
	if err := db.Migrate(database, schemaPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	// Add subtitle column if it does not exist (for compatibility with older DB files)
	_, _ = database.Exec("ALTER TABLE documents ADD COLUMN subtitle TEXT DEFAULT '';")
	log.Println("Database initialized successfully.")

	// Seed tài khoản admin mặc định
	if err := db.SeedDefaultAdmin(database); err != nil {
		log.Printf("Warning: Failed to seed admin user: %v", err)
	}

	// Seed tài khoản test (writer, reader) để kiểm tra phân quyền
	if err := db.SeedTestAccounts(database); err != nil {
		log.Printf("Warning: Failed to seed test accounts: %v", err)
	}

	// Seed thư mục Unsorted Bin
	if err := db.SeedUnsortedBin(database); err != nil {
		log.Printf("Warning: Failed to seed unsorted bin folder: %v", err)
	}


	docRepo := sqlite.NewDocumentRepository(database)
	userRepo := sqlite.NewUserRepository(database)
	mediaRepo := sqlite.NewMediaRepository(database)
	tagRepo := sqlite.NewTagRepository(database)
	bookmarkRepo := sqlite.NewBookmarkRepository(database)

	// Initialize Storage Service (Infrastructure Adapter)
	storageService := storage.NewLocalStorageService("./uploads")

	// Initialize Usecases
	docUsecase := usecase.NewDocumentUsecase(docRepo, tagRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo)
	mediaUsecase := usecase.NewMediaUsecase(mediaRepo, storageService)
	glossaryRepo := sqlite.NewGlossaryRepository(database)
	glossaryUsecase := usecase.NewGlossaryUsecase(glossaryRepo)
	tagUsecase := usecase.NewTagUsecase(tagRepo)
	bookmarkUsecase := usecase.NewBookmarkUsecase(bookmarkRepo)

	// Initialize HTTP Handlers
	docHandler := html.NewDocumentHandler(docUsecase, glossaryUsecase, tagUsecase, bookmarkUsecase)
	apiHandler := api.NewDocumentAPIHandler(docUsecase)
	authHandler := html.NewAuthHandler(authUsecase)
	explorerAPIHandler := api.NewExplorerAPIHandler(docUsecase)
	mediaAPIHandler := api.NewMediaAPIHandler(mediaUsecase)
	glossaryHandler := html.NewGlossaryHandler(glossaryUsecase)
	tagAPIHandler := api.NewTagAPIHandler(tagUsecase)
	bookmarkAPIHandler := api.NewBookmarkAPIHandler(bookmarkUsecase)

	// Initialize Casbin
	enforcer, err := casbin.NewEnforcer("db/casbin/model.conf", "db/casbin/policy.csv")
	if err != nil {
		log.Fatalf("Failed to initialize Casbin enforcer: %v", err)
	}

	// Khởi tạo router
	mux := http.NewServeMux()

	// Phục vụ các file tĩnh (CSS, JS, Images)
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Tĩnh hóa thư mục uploads để truy cập ảnh
	uploadsFs := http.FileServer(http.Dir("./uploads"))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", uploadsFs))

	// Auth Routes (KHÔNG cần Casbin, KHÔNG cần session)
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.LoginSubmit(w, r)
		} else {
			authHandler.LoginPage(w, r)
		}
	})
	mux.HandleFunc("/logout", authHandler.Logout)

	// Routes cho Wiki - Bọc bằng CasbinAuthzMiddleware
	mux.HandleFunc("/", middleware.CasbinAuthzMiddleware(enforcer, docHandler.View))
	mux.HandleFunc("/wiki/", middleware.CasbinAuthzMiddleware(enforcer, docHandler.View))
	mux.HandleFunc("/editor/save", middleware.CasbinAuthzMiddleware(enforcer, docHandler.Save))
	mux.HandleFunc("/editor/", middleware.CasbinAuthzMiddleware(enforcer, docHandler.Edit))
	mux.HandleFunc("/search", middleware.CasbinAuthzMiddleware(enforcer, docHandler.Search))
	mux.HandleFunc("/bookmarks", middleware.CasbinAuthzMiddleware(enforcer, docHandler.Bookmarks))
	mux.HandleFunc("/ui/glossary/tooltip/", middleware.CasbinAuthzMiddleware(enforcer, glossaryHandler.Tooltip))

	// RESTful JSON APIs
	mux.HandleFunc("/api/v1/documents/save", middleware.CasbinAuthzMiddleware(enforcer, apiHandler.SaveDocument))
	mux.HandleFunc("/api/v1/media/upload", middleware.CasbinAuthzMiddleware(enforcer, mediaAPIHandler.Upload))
	mux.HandleFunc("/api/v1/explorer/tree", middleware.CasbinAuthzMiddleware(enforcer, explorerAPIHandler.GetTree))
	mux.HandleFunc("/api/v1/explorer/create", middleware.CasbinAuthzMiddleware(enforcer, explorerAPIHandler.CreateNode))
	mux.HandleFunc("/api/v1/explorer/rename", middleware.CasbinAuthzMiddleware(enforcer, explorerAPIHandler.RenameNode))
	mux.HandleFunc("/api/v1/explorer/delete", middleware.CasbinAuthzMiddleware(enforcer, explorerAPIHandler.DeleteNode))
	mux.HandleFunc("/api/v1/explorer/lock", middleware.CasbinAuthzMiddleware(enforcer, explorerAPIHandler.LockNode))
	mux.HandleFunc("/api/v1/explorer/report", middleware.CasbinAuthzMiddleware(enforcer, explorerAPIHandler.ReportNode))
	mux.HandleFunc("/api/v1/explorer/move", middleware.CasbinAuthzMiddleware(enforcer, explorerAPIHandler.MoveNode))
	mux.HandleFunc("/api/v1/tags", middleware.CasbinAuthzMiddleware(enforcer, tagAPIHandler.HandleTags))
	mux.HandleFunc("/api/v1/bookmarks/toggle", middleware.CasbinAuthzMiddleware(enforcer, bookmarkAPIHandler.Toggle))

	// API Ping test cho HTMX
	mux.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<span class='text-green-600 font-medium'>Pong! HTMX connection is working smoothly.</span>"))
	})

	// Bọc toàn bộ router bằng Session Middleware
	handler := middleware.SessionMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	log.Printf("Server is starting on %s...", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
