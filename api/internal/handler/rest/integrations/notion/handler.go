package notion

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	notionconnector "teammate/internal/connector/notion"
	"teammate/internal/model"
)

// Handler exposes REST endpoints for the Notion integration.
//
// Routes (mounted under /api by cmd/serverd):
// - GET  /integrations/notion/status
// - POST /integrations/notion/pages
//
// Safety: creation is controlled by the connector's DryRun mode, which defaults
// to true via env (NOTION_DRY_RUN).
type Handler struct {
	notion *notionconnector.Client
}

func New(notion *notionconnector.Client) Handler {
	return Handler{notion: notion}
}

func (h Handler) RegisterRoutes(r chi.Router) {
	r.Route("/integrations/notion", func(r chi.Router) {
		r.Get("/status", h.status)
		r.Post("/pages", h.createPage)
	})
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h Handler) status(w http.ResponseWriter, r *http.Request) {
	st := notionconnector.Status{Configured: false, DryRun: true}
	if h.notion != nil {
		st = h.notion.Status()
	}
	writeJSON(w, http.StatusOK, st)
}

func (h Handler) createPage(w http.ResponseWriter, r *http.Request) {
	if h.notion == nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "notion connector not configured"})
		return
	}

	var draft model.TicketDraft
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&draft); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid ticket draft"})
		return
	}

	ref, err := h.notion.CreateTask(r.Context(), draft)
	if err != nil {
		status := http.StatusBadRequest
		if err != notionconnector.ErrNotConfigured {
			status = http.StatusBadGateway
		}
		writeJSON(w, status, errorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, ref)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
