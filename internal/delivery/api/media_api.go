package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"axia-wiki/internal/domain"
	"axia-wiki/internal/middleware"
)

type MediaAPIHandler struct {
	usecase domain.MediaUsecase
}

func NewMediaAPIHandler(usecase domain.MediaUsecase) *MediaAPIHandler {
	return &MediaAPIHandler{usecase: usecase}
}

// Upload handles POST /api/v1/media/upload
func (h *MediaAPIHandler) Upload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Method not allowed"})
		return
	}

	// Parse multipart form, 10 MB limit
	r.ParseMultipartForm(10 << 20)

	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Failed to read file from request"})
		return
	}
	defer file.Close()

	uploaderID := middleware.GetSessionUserID(r)
	if uploaderID == "" {
		uploaderID = "admin-user"
	}

	media, err := h.usecase.UploadFile(r.Context(), file, header.Filename, uploaderID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: err.Error()})
		return
	}

	// Trả về JSON chuẩn của thiết kế
	json.NewEncoder(w).Encode(JSONResponse{
		Status: "success",
		Data: map[string]interface{}{
			"media_id":         media.ID,
			"url":              media.FilePath,
			"markdown_snippet": fmt.Sprintf("![%s](%s)", media.OriginalName, media.FilePath),
		},
	})
}
