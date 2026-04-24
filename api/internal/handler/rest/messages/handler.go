package messages

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"teammate/internal/model"
	"teammate/internal/repository/messages"
)

const (
	defaultLimit  = 50
	defaultOffset = 0
)

// Handler handles message HTTP requests.
type Handler struct {
	repo messages.Repository
}

// New returns a new message handler.
func New(repo messages.Repository) Handler {
	return Handler{repo: repo}
}

// RegisterRoutes mounts message routes under the provided router.
func (h Handler) RegisterRoutes(r chi.Router) {
	r.Route("/messages", func(r chi.Router) {
		r.Post("/", h.uploadMessage)
		r.Get("/", h.readMessages)
	})
}

type uploadMessageRequest struct {
	Body string `json:"body"`
}

type messageResponse struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

type listMessagesResponse struct {
	Messages []messageResponse `json:"messages"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h Handler) uploadMessage(w http.ResponseWriter, r *http.Request) {
	var req uploadMessageRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	body := strings.TrimSpace(req.Body)
	if body == "" {
		writeError(w, http.StatusBadRequest, "body is required")
		return
	}

	msg, err := h.repo.Create(r.Context(), body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store message")
		return
	}

	writeJSON(w, http.StatusCreated, toMessageResponse(msg))
}

func (h Handler) readMessages(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	messages, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read messages")
		return
	}

	response := listMessagesResponse{
		Messages: toMessageResponses(messages),
		Limit:    limit,
		Offset:   offset,
	}

	writeJSON(w, http.StatusOK, response)
}

func parsePagination(r *http.Request) (int, int, error) {
	limit := defaultLimit
	offset := defaultOffset

	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			return 0, 0, fmtError("limit must be a positive integer")
		}
		limit = parsed
	}

	if value := r.URL.Query().Get("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			return 0, 0, fmtError("offset must be a non-negative integer")
		}
		offset = parsed
	}

	return limit, offset, nil
}

func toMessageResponse(msg model.Message) messageResponse {
	return messageResponse{
		ID:        msg.ID,
		Body:      msg.Body,
		CreatedAt: msg.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toMessageResponses(messages []model.Message) []messageResponse {
	responses := make([]messageResponse, 0, len(messages))
	for _, msg := range messages {
		responses = append(responses, toMessageResponse(msg))
	}
	return responses
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func fmtError(message string) error {
	return &requestError{message: message}
}

type requestError struct {
	message string
}

func (e *requestError) Error() string {
	return e.message
}
