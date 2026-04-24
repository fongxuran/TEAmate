package messages

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"nhooyr.io/websocket"

	"teammate/internal/model"
)

type fakeRepository struct {
	createFn func(ctx context.Context, body string, binary []byte, fileName, contentType string, sizeBytes int64) (model.Message, error)
	listFn   func(ctx context.Context, limit, offset int, includeBinary bool) ([]model.Message, error)
}

func (f fakeRepository) Create(ctx context.Context, body string, binary []byte, fileName, contentType string, sizeBytes int64) (model.Message, error) {
	return f.createFn(ctx, body, binary, fileName, contentType, sizeBytes)
}

func (f fakeRepository) List(ctx context.Context, limit, offset int, includeBinary bool) ([]model.Message, error) {
	return f.listFn(ctx, limit, offset, includeBinary)
}

func TestUploadMessageWebsocket(t *testing.T) {
	createdAt := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	repo := fakeRepository{
		createFn: func(ctx context.Context, body string, binary []byte, fileName, contentType string, sizeBytes int64) (model.Message, error) {
			if body != "hello" {
				t.Fatalf("expected body hello, got %s", body)
			}
			if string(binary) != "bin" {
				t.Fatalf("expected binary bin, got %s", string(binary))
			}
			if fileName != "file.txt" {
				t.Fatalf("expected file name file.txt, got %s", fileName)
			}
			if contentType != "text/plain" {
				t.Fatalf("expected content type text/plain, got %s", contentType)
			}
			if sizeBytes != 3 {
				t.Fatalf("expected size 3, got %d", sizeBytes)
			}
			return model.Message{ID: 1, Body: body, FileName: fileName, ContentType: contentType, SizeBytes: sizeBytes, CreatedAt: createdAt}, nil
		},
		listFn: func(ctx context.Context, limit, offset int, includeBinary bool) ([]model.Message, error) {
			return nil, nil
		},
	}

	handler := New(repo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	srv := httptest.NewServer(r)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/messages/ws"
	ctx := context.Background()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	meta := uploadMessageMetadata{
		Body:        "hello",
		FileName:    "file.txt",
		ContentType: "text/plain",
		SizeBytes:   3,
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}

	if err := conn.Write(ctx, websocket.MessageText, metaBytes); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	if err := conn.Write(ctx, websocket.MessageBinary, []byte("bin")); err != nil {
		t.Fatalf("write binary: %v", err)
	}

	_, payload, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	var resp messageResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.ID != 1 {
		t.Fatalf("expected id 1, got %d", resp.ID)
	}
}

func TestReadMessages(t *testing.T) {
	createdAt := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	repo := fakeRepository{
		createFn: func(ctx context.Context, body string, binary []byte, fileName, contentType string, sizeBytes int64) (model.Message, error) {
			return model.Message{}, nil
		},
		listFn: func(ctx context.Context, limit, offset int, includeBinary bool) ([]model.Message, error) {
			if limit != 10 || offset != 5 {
				t.Fatalf("expected limit 10 offset 5, got %d %d", limit, offset)
			}
			if !includeBinary {
				t.Fatalf("expected includeBinary true")
			}
			return []model.Message{{ID: 1, Body: "hello", Binary: []byte("bin"), SizeBytes: 3, CreatedAt: createdAt}}, nil
		},
	}

	handler := New(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/messages?limit=10&offset=5&include_binary=true", nil)
	rec := httptest.NewRecorder()

	handler.readMessages(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var response listMessagesResponse
	if err := json.NewDecoder(bytes.NewReader(rec.Body.Bytes())).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(response.Messages))
	}
	if response.Messages[0].BinaryBase64 == nil {
		t.Fatalf("expected binary_base64 to be set")
	}
}
