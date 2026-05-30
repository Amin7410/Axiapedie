package html

import (
	"fmt"
	"net/http"
	"strings"

	"axia-wiki/internal/domain"
)

type GlossaryHandler struct {
	usecase domain.GlossaryUsecase
}

func NewGlossaryHandler(u domain.GlossaryUsecase) *GlossaryHandler {
	return &GlossaryHandler{usecase: u}
}

// Tooltip handles GET /ui/glossary/tooltip/{id}
func (h *GlossaryHandler) Tooltip(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/ui/glossary/tooltip/")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	term, err := h.usecase.GetTooltipInfo(r.Context(), id)
	if err != nil || term == nil {
		w.Write([]byte(`<div class="p-2 text-sm text-gray-500">Term not found.</div>`))
		return
	}

	docLink := ""
	if term.DocumentID != nil {
		// In MVP we don't have document titles from DocumentID easily available here without joining,
		// so we just link to a generic search or wiki view if needed.
		// For simplicity, we assume term.Term is exactly the document title if there's a link.
		docLink = fmt.Sprintf(`<a href="/wiki/%s" class="text-xs text-blue-300 mt-2 block">View details &rarr;</a>`, term.Term)
	}

	html := fmt.Sprintf(`
		<div class="p-3 bg-gray-800 text-white rounded shadow-lg">
			<h4 class="font-bold text-blue-400">%s</h4>
			<p class="text-sm mt-1">%s</p>
			%s
		</div>
	`, term.Term, term.Definition, docLink)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
