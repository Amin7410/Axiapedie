package api

import (
	"encoding/json"
	"net/http"

	"axia-wiki/internal/domain"
	"axia-wiki/internal/middleware"
)

type BookmarkAPIHandler struct {
	bookmarkUsecase domain.BookmarkUsecase
}

func NewBookmarkAPIHandler(u domain.BookmarkUsecase) *BookmarkAPIHandler {
	return &BookmarkAPIHandler{bookmarkUsecase: u}
}

// Toggle handles POST /api/v1/bookmarks/toggle
func (h *BookmarkAPIHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetSessionUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	docID := r.URL.Query().Get("document_id")
	if docID == "" {
		http.Error(w, "document_id query param is required", http.StatusBadRequest)
		return
	}

	isBookmarked, err := h.bookmarkUsecase.IsBookmarked(r.Context(), userID, docID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isBookmarked {
		err = h.bookmarkUsecase.Remove(r.Context(), userID, docID)
	} else {
		err = h.bookmarkUsecase.Add(r.Context(), userID, docID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newStatus := !isBookmarked

	// If HTMX request, return the updated HTML button snippet directly!
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Trả về HTML cho nút Bookmark mới
		if newStatus {
			w.Write([]byte(`
				<button hx-post="/api/v1/bookmarks/toggle?document_id=` + docID + `" hx-swap="outerHTML" 
				        class="border border-blue-200 bg-blue-50 text-blue-600 hover:bg-blue-100 px-3 py-1.5 rounded-lg flex items-center gap-1 text-xs font-semibold shadow-sm transition" title="Remove bookmark">
					⭐ Saved
				</button>
			`))
		} else {
			w.Write([]byte(`
				<button hx-post="/api/v1/bookmarks/toggle?document_id=` + docID + `" hx-swap="outerHTML" 
				        class="border border-gray-300 hover:bg-gray-50 text-gray-600 px-3 py-1.5 rounded-lg flex items-center gap-1 text-xs font-semibold shadow-sm transition" title="Save article">
					☆ Save
				</button>
			`))
		}
		return
	}

	// Otherwise, return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"bookmarked": newStatus,
	})
}
