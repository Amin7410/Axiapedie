-- LƯỢC ĐỒ CƠ SỞ DỮ LIỆU TỔNG HỢP CHO HỆ THỐNG WIKI
-- Áp dụng cho SQLite3
-- Tuân thủ Clean Architecture: DB Schema phản ánh cấu trúc Entities.

PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL; -- Write-Ahead Logging để tăng tốc độ ghi/đọc đồng thời
PRAGMA synchronous = NORMAL;

-- ==========================================
-- 1. ENTITY: USERS & AUTHORIZATION
-- ==========================================
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'reader', -- 'guest', 'reader', 'writer', 'admin'
    google_id TEXT UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    action TEXT NOT NULL,
    target_id TEXT,
    details TEXT, -- JSON format
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- ==========================================
-- 2. ENTITY: DOCUMENTS & REVISIONS (Bản nháp & Lịch sử)
-- ==========================================
CREATE TABLE IF NOT EXISTS documents (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL UNIQUE,
    parent_id TEXT,
    subtitle TEXT DEFAULT '',
    is_folder INTEGER NOT NULL DEFAULT 0, -- 0: file, 1: folder
    is_locked INTEGER NOT NULL DEFAULT 0,  -- 0: unlocked, 1: locked
    published_revision_id TEXT, -- Bản đang Live (hiển thị cho Reader)
    latest_revision_id TEXT,    -- Bản Nháp mới nhất (hiển thị cho Writer/Admin)
    review_status TEXT DEFAULT 'draft', -- 'draft', 'pending_review', 'published'
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES documents(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_docs_title ON documents(title);
CREATE INDEX IF NOT EXISTS idx_docs_parent_id ON documents(parent_id);

CREATE TABLE IF NOT EXISTS tags (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS document_tags (
    document_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    PRIMARY KEY (document_id, tag_id),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_doc_tags_tag_id ON document_tags(tag_id);

CREATE TABLE IF NOT EXISTS user_bookmarks (
    user_id TEXT NOT NULL,
    document_id TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, document_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_user_bookmarks_user_id ON user_bookmarks(user_id);

CREATE TABLE IF NOT EXISTS revisions (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL,
    parent_id TEXT,            -- Trỏ tới revision cũ để dựng Delta Tree
    author_id TEXT NOT NULL,
    comment TEXT,              -- Tóm tắt chỉnh sửa
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES revisions(id),
    FOREIGN KEY (author_id) REFERENCES users(id)
);

-- Bảng chứa nội dung nặng (Tách biệt khỏi documents để tăng tốc query)
CREATE TABLE IF NOT EXISTS text_contents (
    revision_id TEXT PRIMARY KEY,
    content_type TEXT NOT NULL, -- 'full' (bản mới nhất) hoặc 'delta' (bản nén khác biệt)
    data BLOB NOT NULL,         -- Nội dung Markdown gốc hoặc file zip chứa Delta
    FOREIGN KEY (revision_id) REFERENCES revisions(id) ON DELETE CASCADE
);

-- ==========================================
-- 3. ENTITY: MEDIA & GARBAGE COLLECTION
-- ==========================================
CREATE TABLE IF NOT EXISTS media (
    id TEXT PRIMARY KEY,
    filename TEXT NOT NULL UNIQUE,
    original_name TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    file_path TEXT NOT NULL,
    uploaded_by TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (uploaded_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS document_media (
    document_id TEXT NOT NULL,
    media_id TEXT NOT NULL,
    PRIMARY KEY (document_id, media_id),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_doc_media_media_id ON document_media(media_id);

-- ==========================================
-- 4. ENTITY: WIKI FEATURES (Links, Templates, Glossary)
-- ==========================================
CREATE TABLE IF NOT EXISTS document_links (
    source_id TEXT NOT NULL,
    target_title TEXT NOT NULL,
    target_id TEXT, -- Có thể NULL nếu là "Liên kết đỏ"
    PRIMARY KEY (source_id, target_title),
    FOREIGN KEY (source_id) REFERENCES documents(id) ON DELETE CASCADE,
    FOREIGN KEY (target_id) REFERENCES documents(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_links_target_title ON document_links(target_title);

CREATE TABLE IF NOT EXISTS templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    html_layout TEXT NOT NULL, -- Go HTML Template code
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS glossary_terms (
    id TEXT PRIMARY KEY,
    term TEXT NOT NULL UNIQUE,
    definition TEXT NOT NULL,
    document_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS glossary_aliases (
    term_id TEXT NOT NULL,
    alias TEXT NOT NULL UNIQUE,
    PRIMARY KEY (term_id, alias),
    FOREIGN KEY (term_id) REFERENCES glossary_terms(id) ON DELETE CASCADE
);

-- ==========================================
-- 5. PERFORMANCE & SEARCH CACHE
-- ==========================================
CREATE TABLE IF NOT EXISTS parser_cache (
    revision_id TEXT PRIMARY KEY,
    html_content TEXT NOT NULL,
    toc_json TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (revision_id) REFERENCES revisions(id) ON DELETE CASCADE
);

-- Bảng ảo FTS5 (Full-Text Search 5) hỗ trợ tìm kiếm Tiếng Việt (bỏ dấu)
CREATE VIRTUAL TABLE IF NOT EXISTS documents_fts USING fts5(
    document_id UNINDEXED,
    title,
    content,
    tokenize="unicode61 remove_diacritics 2"
);

-- ==========================================
-- 6. SQLITE TRIGGERS (Tự động hóa đồng bộ FTS)
-- ==========================================
-- Tự động thêm nội dung vào bảng FTS khi lưu một Revision Full Text mới
CREATE TRIGGER IF NOT EXISTS trg_sync_fts_insert AFTER INSERT ON text_contents
WHEN new.content_type = 'full'
BEGIN
    INSERT INTO documents_fts (document_id, title, content)
    VALUES (
        (SELECT document_id FROM revisions WHERE id = new.revision_id),
        (SELECT title FROM documents WHERE id = (SELECT document_id FROM revisions WHERE id = new.revision_id)),
        CAST(new.data AS TEXT)
    );
END;

-- Tự động cập nhật tiêu đề FTS khi đổi tên tài liệu
CREATE TRIGGER IF NOT EXISTS trg_sync_fts_update_title AFTER UPDATE OF title ON documents
BEGIN
    UPDATE documents_fts 
    SET title = new.title
    WHERE document_id = new.id;
END;

-- Tự động xóa khỏi FTS khi tài liệu bị xóa
CREATE TRIGGER IF NOT EXISTS trg_sync_fts_delete AFTER DELETE ON documents
BEGIN
    DELETE FROM documents_fts WHERE document_id = old.id;
END;
