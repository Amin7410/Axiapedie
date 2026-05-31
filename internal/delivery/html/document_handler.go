package html

import (
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"axia-wiki/internal/domain"
	"axia-wiki/internal/middleware"
	"axia-wiki/pkg/parser"
)

type DocumentHandler struct {
	usecase         domain.DocumentUsecase
	glossaryUsecase domain.GlossaryUsecase
	tagUsecase      domain.TagUsecase
	bookmarkUsecase domain.BookmarkUsecase
	templates       map[string]*template.Template
}

func NewDocumentHandler(usecase domain.DocumentUsecase, glossaryUsecase domain.GlossaryUsecase, tagUsecase domain.TagUsecase, bookmarkUsecase domain.BookmarkUsecase) *DocumentHandler {
	// Parse each page template SEPARATELY with the shared layout
	layoutFile := "web/templates/layout.html"
	pages := []string{"view.html", "editor.html", "search.html", "bookmarks.html"}

	templates := make(map[string]*template.Template)
	for _, page := range pages {
		tmpl := template.Must(template.ParseFiles(layoutFile, "web/templates/"+page))
		templates[page] = tmpl
	}

	return &DocumentHandler{
		usecase:         usecase,
		glossaryUsecase: glossaryUsecase,
		tagUsecase:      tagUsecase,
		bookmarkUsecase: bookmarkUsecase,
		templates:       templates,
	}
}

func (h *DocumentHandler) render(w http.ResponseWriter, r *http.Request, page string, data interface{}) {
	tmpl, ok := h.templates[page]
	if !ok {
		http.Error(w, "Template not found: "+page, http.StatusInternalServerError)
		return
	}

	// Inject session details into the template data map
	var title string
	if m, ok := data.(map[string]interface{}); ok {
		m["SessionUserID"] = middleware.GetSessionUserID(r)
		m["SessionUserRole"] = middleware.GetSessionUserRole(r)
		m["SessionUsername"] = middleware.GetSessionUsername(r)
		if t, ok := m["Title"].(string); ok {
			title = t
		}
		
		// Lấy danh sách bookmark của người dùng hiện tại
		if userID, ok := m["SessionUserID"].(string); ok && userID != "" {
			bookmarks, _ := h.bookmarkUsecase.GetByUser(r.Context(), userID)
			m["Bookmarks"] = bookmarks
		}
	}

	// If HTMX request, execute ONLY the "content" template definition
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

	// Execute the page file itself (which calls {{template "layout" .}})
	if err := tmpl.ExecuteTemplate(w, page, data); err != nil {
		http.Error(w, "Template render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// View handles GET /wiki/{title}
func (h *DocumentHandler) View(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimPrefix(r.URL.Path, "/wiki/")
	if title == "" || title == "/" {
		title = "Home"
	}

	doc, content, err := h.usecase.GetDocument(r.Context(), title)
	if err != nil && err.Error() != "document not found" {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var htmlContent string
	var breadcrumbs []map[string]interface{}
	description := "Axia Wiki - A modern, high-performance personal wiki and knowledge management system."

	if doc != nil {
		allDocs, _ := h.usecase.GetAll(r.Context())
		existingMap := make(map[string]bool)
		for _, ad := range allDocs {
			if !ad.IsFolder {
				existingMap[strings.ToLower(strings.TrimSpace(ad.Title))] = true
			}
		}
		
		existFn := func(targetTitle string) bool {
			return existingMap[strings.ToLower(strings.TrimSpace(targetTitle))]
		}

		if content != "" {
			parsed, err := parser.ParseToHTML(content, existFn)
			if err == nil {
				htmlContent = parsed
			} else {
				htmlContent = "<p class='text-red-500'>Error parsing markdown</p>"
			}

			if h.glossaryUsecase != nil {
				terms, _ := h.glossaryUsecase.GetAllTerms(r.Context())
				annotator := parser.NewGlossaryAnnotator(terms)
				htmlContent = annotator.AnnotateHTML(htmlContent)
			}
		}

		// Build breadcrumbs
		docMap := make(map[string]*domain.Document)
		for _, ad := range allDocs {
			docMap[ad.ID] = ad
		}

		curr := doc
		visited := make(map[string]bool)
		visited[curr.ID] = true
		for curr.ParentID != nil {
			pID := *curr.ParentID
			if pID == "" || pID == "root" || visited[pID] {
				break
			}
			visited[pID] = true
			parentDoc, exists := docMap[pID]
			if !exists {
				break
			}
			breadcrumbs = append([]map[string]interface{}{{
				"Title":    parentDoc.Title,
				"Link":     "/wiki/" + parentDoc.Title,
				"IsFolder": parentDoc.IsFolder,
			}}, breadcrumbs...)
			curr = parentDoc
		}

		// Add Position field for JSON-LD breadcrumbs schema
		for i, bc := range breadcrumbs {
			bc["Position"] = i + 2
		}

		// Make SEO description
		if doc.Subtitle != "" {
			description = doc.Subtitle
		} else if content != "" {
			description = makeDescription(content)
		}
	}

	data := map[string]interface{}{
		"Title":        title,
		"HTMLContent":  template.HTML(htmlContent),
		"IsLocked":     false,
		"IsHidden":     false,
		"IsBookmarked": false,
		"DocID":        "",
		"Subtitle":     "",
		"DoesNotExist": false,
		"Breadcrumbs":  breadcrumbs,
		"Description":  description,
	}
	if doc != nil {
		data["DocID"] = doc.ID
		data["Content"] = content
		data["Subtitle"] = doc.Subtitle
		data["IsLocked"] = doc.IsLocked
		data["IsHidden"] = doc.IsHidden
		data["LatestRevisionID"] = ""
		if doc.LatestRevisionID != nil {
			data["LatestRevisionID"] = *doc.LatestRevisionID
		}
		// Lấy danh sách tag của document
		tags, _ := h.tagUsecase.GetTagsByDocument(r.Context(), doc.ID)
		data["Tags"] = tags

		// AI/SEO: truyền ngày tạo, ngày cập nhật, và keywords vào template
		data["CreatedAt"] = doc.CreatedAt.Format("2006-01-02")
		data["UpdatedAt"] = doc.UpdatedAt.Format("2006-01-02")
		var kwParts []string
		for _, t := range tags {
			kwParts = append(kwParts, t.Name)
		}
		data["Keywords"] = strings.Join(kwParts, ", ")

		// Kiểm tra xem người dùng hiện tại có lưu bài này không
		if userID := middleware.GetSessionUserID(r); userID != "" {
			isBookmarked, _ := h.bookmarkUsecase.IsBookmarked(r.Context(), userID, doc.ID)
			data["IsBookmarked"] = isBookmarked
		}
	} else {
		data["DoesNotExist"] = true
		w.WriteHeader(http.StatusNotFound)
	}

	h.render(w, r, "view.html", data)
}

// Edit handles GET /editor/{title}
func (h *DocumentHandler) Edit(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimPrefix(r.URL.Path, "/editor/")
	if title == "" || title == "save" {
		http.Error(w, "Invalid title", http.StatusBadRequest)
		return
	}

	doc, content, err := h.usecase.GetDocument(r.Context(), title)
	if err != nil && err.Error() != "document not found" {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Chặn chỉnh sửa trang chủ (Home) đối với người dùng không phải admin
	if strings.ToLower(title) == "home" {
		userRole := middleware.GetSessionUserRole(r)
		if userRole != "admin" {
			http.Redirect(w, r, "/wiki/"+title, http.StatusSeeOther)
			return
		}
	}

	// Chặn mở editor cho trang bị khoá (nếu không phải admin)
	if doc != nil && doc.IsLocked {
		userRole := middleware.GetSessionUserRole(r)
		if userRole != "admin" {
			http.Redirect(w, r, "/wiki/"+title, http.StatusSeeOther)
			return
		}
	}

	data := map[string]interface{}{
		"Title":            title,
		"Content":          content,
		"LatestRevisionID": "",
		"IsLocked":         false,
		"Tags":             "",
		"DocID":            "",
		"Subtitle":         "",
	}
	if doc != nil {
		data["DocID"] = doc.ID
		data["Subtitle"] = doc.Subtitle
		data["IsLocked"] = doc.IsLocked
		if doc.LatestRevisionID != nil {
			data["LatestRevisionID"] = *doc.LatestRevisionID
		}
		// Lấy danh sách tag và nối lại thành chuỗi ngăn cách bằng dấu phẩy
		tags, _ := h.tagUsecase.GetTagsByDocument(r.Context(), doc.ID)
		var tagNames []string
		for _, t := range tags {
			tagNames = append(tagNames, t.Name)
		}
		data["Tags"] = strings.Join(tagNames, ", ")
	}

	h.render(w, r, "editor.html", data)
}

// Save handles POST /editor/save via HTMX
func (h *DocumentHandler) Save(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	baseRevID := r.FormValue("base_revision_id")
	comment := r.FormValue("comment")
	subtitle := r.FormValue("subtitle")
	
	parentIDStr := r.FormValue("parent_id")
	var parentID *string
	if parentIDStr != "" {
		parentID = &parentIDStr
	} else {
		// Default to Unsorted Bin for documents created outside explorer (e.g. from editor/WikiLinks)
		binID := "unsorted_bin_folder"
		parentID = &binID
	}

	// Lấy author ID thật từ session
	authorID := middleware.GetSessionUserID(r)
	if authorID == "" {
		w.Write([]byte("<script>showToast('Please sign in to save articles.', 'error');</script>"))
		return
	}

	// Chặn lưu trang chủ (Home) đối với người dùng không phải admin
	if strings.ToLower(title) == "home" {
		userRole := middleware.GetSessionUserRole(r)
		if userRole != "admin" {
			w.Write([]byte("<script>showToast('Only administrators can edit the Home page.', 'error');</script>"))
			return
		}
	}

	tagsStr := r.FormValue("tags")
	var tags []string
	if tagsStr != "" {
		rawTags := strings.Split(tagsStr, ",")
		for _, t := range rawTags {
			trimmed := strings.TrimSpace(t)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	doc, err := h.usecase.SaveDraft(r.Context(), title, subtitle, content, authorID, baseRevID, comment, parentID, tags)
	if err != nil {
		if err.Error() == "no changes detected" {
			w.Write([]byte("<script>showToast('No changes to save.', 'warning');</script>"))
			return
		}
		if err.Error() == "edit conflict: base revision does not match latest revision" {
			w.Write([]byte("<script>showToast('Edit conflict: someone else saved changes before you.', 'error');</script>"))
			return
		}
		w.Write([]byte("<script>showToast('Save failed: " + err.Error() + "', 'error');</script>"))
		return
	}

	w.Write([]byte("<script>showToast('Saved successfully.'); if(window.loadExplorerTree) window.loadExplorerTree();</script>"))
	if doc != nil && doc.LatestRevisionID != nil {
		w.Write([]byte("<input type='hidden' name='base_revision_id' id='base_revision_id' value='" + *doc.LatestRevisionID + "' hx-swap-oob='true'>"))
	}
}

// Search handles GET /search
func (h *DocumentHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	var results []*domain.Document
	var err error
	if query != "" {
		results, err = h.usecase.Search(r.Context(), query)
		if err != nil {
			// Log nhưng không crash, hiển thị kết quả rỗng
			results = []*domain.Document{}
		}
	}

	data := map[string]interface{}{
		"Query":   query,
		"Results": results,
	}

	h.render(w, r, "search.html", data)
}

// Bookmarks handles GET /bookmarks
func (h *DocumentHandler) Bookmarks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetSessionUserID(r)
	if userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	bookmarks, err := h.bookmarkUsecase.GetByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":     "Saved articles",
		"Bookmarks": bookmarks,
	}

	h.render(w, r, "bookmarks.html", data)
}

func makeDescription(markdown string) string {
	// 1. Loại bỏ wiki links: [[Page Title|Display Text]] -> Display Text, [[Title]] -> Title
	reWiki := regexp.MustCompile(`\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)
	text := reWiki.ReplaceAllStringFunc(markdown, func(m string) string {
		submatches := reWiki.FindStringSubmatch(m)
		if len(submatches) > 2 && submatches[2] != "" {
			return submatches[2]
		}
		if len(submatches) > 1 {
			return submatches[1]
		}
		return ""
	})

	// 2. Loại bỏ định dạng tiêu đề
	reHeaders := regexp.MustCompile(`(?m)^#+\s+(.*)$`)
	text = reHeaders.ReplaceAllString(text, "$1")

	// 3. Loại bỏ ký tự in đậm, in nghiêng
	reBoldItalic := regexp.MustCompile(`\*\*([^*]+)\*\*|__([^_]+)__| \*([^*]+)\*| _([^_]+)_`)
	text = reBoldItalic.ReplaceAllString(text, "$1$2$3$4")

	// 4. Loại bỏ các khối code inline
	reCode := regexp.MustCompile("`([^`]+)`")
	text = reCode.ReplaceAllString(text, "$1")

	// 5. Loại bỏ thẻ HTML thô nếu có
	reTags := regexp.MustCompile(`<[^>]*>`)
	text = reTags.ReplaceAllString(text, "")

	// Chuẩn hóa khoảng trắng
	text = strings.Join(strings.Fields(text), " ")

	// Giới hạn trong khoảng 155-160 ký tự
	if len(text) > 160 {
		cutIdx := 157
		if lastSpace := strings.LastIndex(text[:cutIdx], " "); lastSpace > 120 {
			cutIdx = lastSpace
		}
		text = text[:cutIdx] + "..."
	}

	return text
}

// Sitemap generates /sitemap.xml for AI crawlers and search engines
func (h *DocumentHandler) Sitemap(w http.ResponseWriter, r *http.Request) {
	docs, err := h.usecase.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get base URL from request
	scheme := "https"
	if r.TLS == nil {
		if fwd := r.Header.Get("X-Forwarded-Proto"); fwd != "" {
			scheme = fwd
		} else {
			scheme = "http"
		}
	}
	baseURL := scheme + "://" + r.Host

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n"))
	w.Write([]byte(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n"))

	// Home page
	w.Write([]byte("  <url>\n"))
	w.Write([]byte("    <loc>" + baseURL + "/wiki/Home</loc>\n"))
	w.Write([]byte("    <changefreq>daily</changefreq>\n"))
	w.Write([]byte("    <priority>1.0</priority>\n"))
	w.Write([]byte("  </url>\n"))

	for _, doc := range docs {
		if doc.IsFolder || doc.Title == "Home" {
			continue
		}
		w.Write([]byte("  <url>\n"))
		w.Write([]byte("    <loc>" + baseURL + "/wiki/" + strings.ReplaceAll(doc.Title, " ", "%20") + "</loc>\n"))
		w.Write([]byte("    <lastmod>" + doc.UpdatedAt.Format("2006-01-02") + "</lastmod>\n"))
		w.Write([]byte("    <changefreq>weekly</changefreq>\n"))
		w.Write([]byte("    <priority>0.8</priority>\n"))
		w.Write([]byte("  </url>\n"))
	}

	w.Write([]byte("</urlset>\n"))
}
