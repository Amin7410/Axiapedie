package api

import (
	"encoding/json"
	"net/http"

	"axia-wiki/internal/domain"
	"axia-wiki/internal/middleware"
)

type TagAPIHandler struct {
	tagUsecase domain.TagUsecase
}

// NewTagAPIHandler tạo một handler xử lý API liên quan đến Tag
func NewTagAPIHandler(u domain.TagUsecase) *TagAPIHandler {
	return &TagAPIHandler{tagUsecase: u}
}

// ListTags handles GET /api/v1/tags
func (h *TagAPIHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Method not allowed"})
		return
	}

	tags, err := h.tagUsecase.GetAllTags(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   tags,
	})
}

// HandleTags dispatches requests on /api/v1/tags to ListTags or CreateTag
func (h *TagAPIHandler) HandleTags(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.CreateTag(w, r)
	} else {
		h.ListTags(w, r)
	}
}

type CreateTagRequest struct {
	Name string `json:"name"`
}

// CreateTag handles POST /api/v1/tags
func (h *TagAPIHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Method not allowed"})
		return
	}

	// Verify user role is admin
	role := middleware.GetSessionUserRole(r)
	if role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Only admins can create tags"})
		return
	}

	var req CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Invalid request payload"})
		return
	}

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Tag name cannot be empty"})
		return
	}

	err := h.tagUsecase.CreateTag(r.Context(), req.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Tag created successfully",
	})
}
