package mvp

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"teammate/internal/export"
	"teammate/internal/realtime"
)

type Handler struct {
	hub *realtime.Hub
}

func New(hub *realtime.Hub) Handler {
	return Handler{hub: hub}
}

func (h Handler) RegisterRoutes(r chi.Router) {
	r.Route("/meeting", func(r chi.Router) {
		r.Get("/state", h.getState)
		r.Post("/reset", h.postReset)
		r.Post("/analyze", h.postAnalyze)
	})
	// Exports are derived from the last computed analysis.
	r.Route("/exports", func(r chi.Router) {
		r.Post("/ticket-drafts", h.postTicketDraftExport)
	})
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h Handler) getState(w http.ResponseWriter, r *http.Request) {
	s := h.hub.GetSession(r.URL.Query().Get("session"))
	b, err := s.MarshalSnapshot()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to encode state"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func (h Handler) postReset(w http.ResponseWriter, r *http.Request) {
	s := h.hub.GetSession(r.URL.Query().Get("session"))
	s.Reset()
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h Handler) postAnalyze(w http.ResponseWriter, r *http.Request) {
	s := h.hub.GetSession(r.URL.Query().Get("session"))
	res := s.Recompute()
	writeJSON(w, http.StatusOK, res)
}

type ticketDraftExportResponse struct {
	SchemaVersion string                 `json:"schema_version"`
	GeneratedAt   string                 `json:"generated_at"`
	Drafts        any                    `json:"drafts"`
	Markdown      string                 `json:"markdown"`
	CSV           string                 `json:"csv"`
}

func (h Handler) postTicketDraftExport(w http.ResponseWriter, r *http.Request) {
	s := h.hub.GetSession(r.URL.Query().Get("session"))
	_, _, _, _, res := s.Snapshot()

	exp := export.TicketDraftExport{
		SchemaVersion: "v1",
		GeneratedAt:   res.GeneratedAt,
		MeetingID:     res.Transcript.MeetingID,
		MeetingName:   res.Transcript.MeetingName,
		Drafts:        res.TicketDrafts,
	}
	md := export.RenderTicketDraftsMarkdown(exp)
	csv, err := export.RenderTicketDraftsCSV(exp)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to render csv"})
		return
	}

	generatedAt := res.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}
	writeJSON(w, http.StatusOK, ticketDraftExportResponse{
		SchemaVersion: "v1",
		GeneratedAt:   generatedAt.UTC().Format(time.RFC3339),
		Drafts:        res.TicketDrafts,
		Markdown:      md,
		CSV:           csv,
	})
}
