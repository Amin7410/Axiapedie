package domain

import (
	"context"
	"time"
)

// Document represents a wiki page or folder.
type Document struct {
	ID                  string
	Title               string
	Subtitle            string
	ParentID            *string
	IsFolder            bool
	IsLocked            bool
	IsHidden            bool
	PublishedRevisionID *string
	LatestRevisionID    *string
	ReviewStatus        string // 'draft', 'pending_review', 'published'
	SortOrder           int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Revision represents a specific version of a document.
type Revision struct {
	ID         string
	DocumentID string
	ParentID   *string
	AuthorID   string
	Comment    string
	CreatedAt  time.Time
}

// TextContent holds the heavy markdown content or compressed delta.
type TextContent struct {
	RevisionID  string
	ContentType string // 'full' or 'delta'
	Data        []byte // Raw markdown or gzipped delta
}

// DocumentRepository defines the interface for data access.
type DocumentRepository interface {
	GetByID(ctx context.Context, id string) (*Document, error)
	GetByTitle(ctx context.Context, title string) (*Document, error)
	Create(ctx context.Context, doc *Document) error
	Update(ctx context.Context, doc *Document) error
	GetAll(ctx context.Context) ([]*Document, error)
	Delete(ctx context.Context, id string) error
	
	// Revisions
	SaveRevision(ctx context.Context, rev *Revision, content *TextContent) error
	GetRevision(ctx context.Context, id string) (*Revision, *TextContent, error)
	// UpdateTextContent cập nhật nội dung text_contents (chuyển full -> delta)
	UpdateTextContent(ctx context.Context, tc *TextContent) error

	// Search
	Search(ctx context.Context, query string) ([]*Document, error)
	SearchWithTags(ctx context.Context, textQuery string, tags []string) ([]*Document, error)
}

// DocumentUsecase defines the business logic interface.
type DocumentUsecase interface {
	GetDocument(ctx context.Context, title string) (*Document, string, error) // Returns doc and its full HTML content
	SaveDraft(ctx context.Context, title, subtitle, content, authorID, baseRevID, comment string, parentID *string, tags []string) (*Document, error)
	Search(ctx context.Context, query string) ([]*Document, error)
	GetAll(ctx context.Context) ([]*Document, error)
	CreateFolder(ctx context.Context, title string, parentID *string) (*Document, error)
	Delete(ctx context.Context, id string) error
	Rename(ctx context.Context, id string, newTitle string) (*Document, error)
	SetLock(ctx context.Context, id string, locked bool) (*Document, error)
	SetHidden(ctx context.Context, id string, hidden bool) (*Document, error)
	Move(ctx context.Context, id string, parentID *string, targetID *string, position string) (*Document, error)
	BulkDelete(ctx context.Context, ids []string) error
	BulkSetLock(ctx context.Context, ids []string, locked bool) error
	BulkSetHidden(ctx context.Context, ids []string, hidden bool) error
}
