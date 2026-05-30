package api

import (
	"encoding/json"
	"net/http"

	"axia-wiki/internal/domain"
)

type DocumentAPIHandler struct {
	usecase domain.DocumentUsecase
}

func NewDocumentAPIHandler(usecase domain.DocumentUsecase) *DocumentAPIHandler {
	return &DocumentAPIHandler{
		usecase: usecase,
	}
}

type SaveDocumentRequest struct {
	Title          string `json:"title"`
	Subtitle       string `json:"subtitle"`
	BaseRevisionID string `json:"base_revision_id"`
	Content        string `json:"content"`
	Comment        string `json:"comment"`
}

type JSONResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// SaveDocument handles POST /api/v1/documents/save
func (h *DocumentAPIHandler) SaveDocument(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Method not allowed"})
		return
	}

	var req SaveDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Invalid request payload"})
		return
	}

	authorID := "api-user"
	doc, err := h.usecase.SaveDraft(r.Context(), req.Title, req.Subtitle, req.Content, authorID, req.BaseRevisionID, req.Comment, nil, nil)
	if err != nil {
		if err.Error() == "edit conflict: base revision does not match latest revision" {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Conflict detected"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Failed to save document"})
		return
	}

	json.NewEncoder(w).Encode(JSONResponse{
		Status: "success",
		Data: map[string]string{
			"new_revision_id": *doc.LatestRevisionID,
			"title":           doc.Title,
		},
	})
}
