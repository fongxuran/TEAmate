package messages

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"nhooyr.io/websocket"

	"teammate/internal/model"
	"teammate/internal/repository/messages"
)

const (
	defaultLimit  = 50
	defaultOffset = 0
	maxBinarySize = 15 * 1024 * 1024
	maxMetaSize   = 64 * 1024
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
		r.Get("/", h.readMessages)
		r.Get("/ws", h.uploadMessageWS)
	})
}

type uploadMessageMetadata struct {
	Body        string `json:"body"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

type messageResponse struct {
	ID           int64   `json:"id"`
	Body         string  `json:"body,omitempty"`
	FileName     string  `json:"file_name,omitempty"`
	ContentType  string  `json:"content_type,omitempty"`
	SizeBytes    int64   `json:"size_bytes,omitempty"`
	BinaryBase64 *string `json:"binary_base64,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

type listMessagesResponse struct {
	Messages []messageResponse `json:"messages"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h Handler) uploadMessageWS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusInternalError, "internal error")

	msgType, reader, err := conn.Reader(ctx)
	if err != nil {
		return
	}
	if msgType != websocket.MessageText {
		writeWSError(ctx, conn, websocket.StatusUnsupportedData, "metadata must be JSON text")
		return
	}

	metaBytes, err := readLimited(reader, maxMetaSize)
	if err != nil {
		writeWSError(ctx, conn, websocket.StatusMessageTooLarge, err.Error())
		return
	}

	var meta uploadMessageMetadata
	decoder := json.NewDecoder(bytes.NewReader(metaBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&meta); err != nil {
		writeWSError(ctx, conn, websocket.StatusInvalidFramePayloadData, "invalid metadata")
		return
	}

	if meta.SizeBytes < 0 {
		writeWSError(ctx, conn, websocket.StatusInvalidFramePayloadData, "size_bytes must be non-negative")
		return
	}
	if meta.SizeBytes > maxBinarySize {
		writeWSError(ctx, conn, websocket.StatusMessageTooLarge, "binary payload exceeds 15mb limit")
		return
	}

	body := strings.TrimSpace(meta.Body)
	var binary []byte
	if meta.SizeBytes > 0 {
		msgType, reader, err = conn.Reader(ctx)
		if err != nil {
			return
		}
		if msgType != websocket.MessageBinary {
			writeWSError(ctx, conn, websocket.StatusUnsupportedData, "binary payload must be a binary frame")
			return
		}

		binary, err = readLimited(reader, maxBinarySize)
		if err != nil {
			writeWSError(ctx, conn, websocket.StatusMessageTooLarge, err.Error())
			return
		}
		if int64(len(binary)) != meta.SizeBytes {
			writeWSError(ctx, conn, websocket.StatusInvalidFramePayloadData, "size_bytes does not match payload size")
			return
		}
	}

	if body == "" && len(binary) == 0 {
		writeWSError(ctx, conn, websocket.StatusInvalidFramePayloadData, "body or binary payload is required")
		return
	}

	msg, err := h.repo.Create(ctx, body, binary, strings.TrimSpace(meta.FileName), strings.TrimSpace(meta.ContentType), meta.SizeBytes)
	if err != nil {
		writeWSError(ctx, conn, websocket.StatusInternalError, "failed to store message")
		return
	}

	response := toMessageResponse(msg, false)
	payload, err := json.Marshal(response)
	if err != nil {
		writeWSError(ctx, conn, websocket.StatusInternalError, "failed to encode response")
		return
	}
	if err := conn.Write(ctx, websocket.MessageText, payload); err != nil {
		return
	}

	_ = conn.Close(websocket.StatusNormalClosure, "")
}

func (h Handler) readMessages(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	includeBinary, err := parseIncludeBinary(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	messages, err := h.repo.List(r.Context(), limit, offset, includeBinary)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read messages")
		return
	}

	response := listMessagesResponse{
		Messages: toMessageResponses(messages, includeBinary),
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

func parseIncludeBinary(r *http.Request) (bool, error) {
	value := r.URL.Query().Get("include_binary")
	if value == "" || value == "false" {
		return false, nil
	}
	if value == "true" {
		return true, nil
	}
	return false, fmtError("include_binary must be true or false")
}

func toMessageResponse(msg model.Message, includeBinary bool) messageResponse {
	var binaryBase64 *string
	if includeBinary && len(msg.Binary) > 0 {
		encoded := base64.StdEncoding.EncodeToString(msg.Binary)
		binaryBase64 = &encoded
	}

	return messageResponse{
		ID:           msg.ID,
		Body:         msg.Body,
		FileName:     msg.FileName,
		ContentType:  msg.ContentType,
		SizeBytes:    msg.SizeBytes,
		BinaryBase64: binaryBase64,
		CreatedAt:    msg.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toMessageResponses(messages []model.Message, includeBinary bool) []messageResponse {
	responses := make([]messageResponse, 0, len(messages))
	for _, msg := range messages {
		responses = append(responses, toMessageResponse(msg, includeBinary))
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

func writeWSError(ctx context.Context, conn *websocket.Conn, status websocket.StatusCode, message string) {
	payload, err := json.Marshal(errorResponse{Error: message})
	if err == nil {
		_ = conn.Write(ctx, websocket.MessageText, payload)
	}
	_ = conn.Close(status, message)
}

func readLimited(reader io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmtError("payload exceeds limit")
	}
	return data, nil
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
