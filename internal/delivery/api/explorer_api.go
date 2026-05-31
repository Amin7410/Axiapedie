package api

import (
	"encoding/json"
	"net/http"

	"axia-wiki/internal/domain"
	"axia-wiki/internal/middleware"
)

type ExplorerAPIHandler struct {
	docUsecase domain.DocumentUsecase
}

func NewExplorerAPIHandler(docUsecase domain.DocumentUsecase) *ExplorerAPIHandler {
	return &ExplorerAPIHandler{
		docUsecase: docUsecase,
	}
}

type ExplorerNode struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	ParentID  *string         `json:"parent_id"`
	IsFolder  bool            `json:"is_folder"`
	IsLocked  bool            `json:"is_locked"`
	IsHidden  bool            `json:"is_hidden"`
	SortOrder int             `json:"sort_order"`
	Children  []*ExplorerNode `json:"children"`
}

type CreateNodeRequest struct {
	Title    string  `json:"title"`
	ParentID *string `json:"parent_id"`
	IsFolder bool    `json:"is_folder"`
}

// GetTree handles GET /api/v1/explorer/tree
func (h *ExplorerAPIHandler) GetTree(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	docs, err := h.docUsecase.GetAll(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: err.Error()})
		return
	}

	// Build the tree hierarchy from flat list of docs
	nodesMap := make(map[string]*ExplorerNode)
	var rootNodes []*ExplorerNode

	// 1. Create a node for each document
	for _, doc := range docs {
		nodesMap[doc.ID] = &ExplorerNode{
			ID:        doc.ID,
			Title:     doc.Title,
			ParentID:  doc.ParentID,
			IsFolder:  doc.IsFolder,
			IsLocked:  doc.IsLocked,
			IsHidden:  doc.IsHidden,
			SortOrder: doc.SortOrder,
			Children:  []*ExplorerNode{},
		}
	}

	// 2. Associate children with their parents
	for _, doc := range docs {
		node := nodesMap[doc.ID]
		if doc.ParentID == nil || *doc.ParentID == "" {
			rootNodes = append(rootNodes, node)
		} else {
			parentNode, exists := nodesMap[*doc.ParentID]
			if exists {
				parentNode.Children = append(parentNode.Children, node)
			} else {
				// Fallback to root if parent not found
				rootNodes = append(rootNodes, node)
			}
		}
	}

	json.NewEncoder(w).Encode(JSONResponse{
		Status: "success",
		Data:   rootNodes,
	})
}

// CreateNode handles POST /api/v1/explorer/create
func (h *ExplorerAPIHandler) CreateNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Method not allowed"})
		return
	}

	var req CreateNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Invalid request payload"})
		return
	}

	if req.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Title cannot be empty"})
		return
	}

	authorID := middleware.GetSessionUserID(r)
	if authorID == "" {
		authorID = "admin-user" // Fallback
	}

	var doc *domain.Document
	var err error

	if req.IsFolder {
		doc, err = h.docUsecase.CreateFolder(r.Context(), req.Title, req.ParentID)
	} else {
		// Kiểm tra xem tên bài viết đã tồn tại chưa để tránh lỗi Edit Conflict
		existingDoc, _, errGet := h.docUsecase.GetDocument(r.Context(), req.Title)
		if errGet == nil && existingDoc != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "a document or folder with this name already exists"})
			return
		}
		// Create blank document draft
		doc, err = h.docUsecase.SaveDraft(r.Context(), req.Title, "", "", authorID, "", "Created via Sidebar Explorer", req.ParentID, nil)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(JSONResponse{
		Status: "success",
		Data:   doc,
	})
}

type RenameNodeRequest struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type DeleteNodeRequest struct {
	ID  string   `json:"id"`
	IDs []string `json:"ids"`
}

type LockNodeRequest struct {
	ID     string   `json:"id"`
	IDs    []string `json:"ids"`
	Locked bool     `json:"locked"`
}

type HideNodeRequest struct {
	ID     string   `json:"id"`
	IDs    []string `json:"ids"`
	Hidden bool     `json:"hidden"`
}

type ReportNodeRequest struct {
	ID     string   `json:"id"`
	IDs    []string `json:"ids"`
	Reason string   `json:"reason"`
}

type MoveNodeRequest struct {
	ID       string   `json:"id"`
	IDs      []string `json:"ids"`
	ParentID *string  `json:"parent_id"`
	TargetID *string  `json:"target_id"`
	Position string   `json:"position"`
}

// RenameNode handles POST /api/v1/explorer/rename
func (h *ExplorerAPIHandler) RenameNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Method not allowed"})
		return
	}

	var req RenameNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Invalid request payload"})
		return
	}

	doc, err := h.docUsecase.Rename(r.Context(), req.ID, req.Title)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(JSONResponse{
		Status: "success",
		Data:   doc,
	})
}

// DeleteNode handles POST /api/v1/explorer/delete
func (h *ExplorerAPIHandler) DeleteNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Method not allowed"})
		return
	}

	var req DeleteNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Invalid request payload"})
		return
	}

	ids := req.IDs
	if len(ids) == 0 && req.ID != "" {
		ids = []string{req.ID}
	}

	err := h.docUsecase.BulkDelete(r.Context(), ids)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(JSONResponse{
		Status: "success",
	})
}

// LockNode handles POST /api/v1/explorer/lock
func (h *ExplorerAPIHandler) LockNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Method not allowed"})
		return
	}

	var req LockNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Invalid request payload"})
		return
	}

	ids := req.IDs
	if len(ids) == 0 && req.ID != "" {
		ids = []string{req.ID}
	}

	err := h.docUsecase.BulkSetLock(r.Context(), ids, req.Locked)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(JSONResponse{
		Status: "success",
	})
}

// HideNode handles POST /api/v1/explorer/hide
func (h *ExplorerAPIHandler) HideNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Method not allowed"})
		return
	}

	var req HideNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Invalid request payload"})
		return
	}

	ids := req.IDs
	if len(ids) == 0 && req.ID != "" {
		ids = []string{req.ID}
	}

	err := h.docUsecase.BulkSetHidden(r.Context(), ids, req.Hidden)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(JSONResponse{
		Status: "success",
	})
}

// ReportNode handles POST /api/v1/explorer/report
func (h *ExplorerAPIHandler) ReportNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Method not allowed"})
		return
	}

	var req ReportNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Invalid request payload"})
		return
	}

	ids := req.IDs
	if len(ids) == 0 && req.ID != "" {
		ids = []string{req.ID}
	}

	for _, id := range ids {
		println("🚨 [DOCUMENT REPORT] Node ID: " + id + ", Reason: " + req.Reason)
	}

	json.NewEncoder(w).Encode(JSONResponse{
		Status:  "success",
		Message: "Thank you for your report. Moderators will review this document soon.",
	})
}

// MoveNode handles POST /api/v1/explorer/move
func (h *ExplorerAPIHandler) MoveNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Method not allowed"})
		return
	}

	var req MoveNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: "Invalid request payload"})
		return
	}

	ids := req.IDs
	if len(ids) == 0 && req.ID != "" {
		ids = []string{req.ID}
	}

	var lastDoc *domain.Document
	for _, id := range ids {
		doc, err := h.docUsecase.Move(r.Context(), id, req.ParentID, req.TargetID, req.Position)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: err.Error()})
			return
		}
		lastDoc = doc

		// Để chèn liên tiếp các item kéo thả hàng loạt xếp liền nhau:
		if req.Position == "after" || req.Position == "before" {
			req.Position = "after"
			newTargetID := doc.ID
			req.TargetID = &newTargetID
		}
	}

	json.NewEncoder(w).Encode(JSONResponse{
		Status: "success",
		Data:   lastDoc,
	})
}
