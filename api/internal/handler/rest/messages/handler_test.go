package messages

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"teammate/internal/model"
)

type fakeRepository struct {
	createFn func(ctx context.Context, body string) (model.Message, error)
	listFn   func(ctx context.Context, limit, offset int) ([]model.Message, error)
}

func (f fakeRepository) Create(ctx context.Context, body string) (model.Message, error) {
	return f.createFn(ctx, body)
}

func (f fakeRepository) List(ctx context.Context, limit, offset int) ([]model.Message, error) {
	return f.listFn(ctx, limit, offset)
}

func TestUploadMessage(t *testing.T) {
	createdAt := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	repo := fakeRepository{
		createFn: func(ctx context.Context, body string) (model.Message, error) {
			return model.Message{ID: 1, Body: body, CreatedAt: createdAt}, nil
		},
		listFn: func(ctx context.Context, limit, offset int) ([]model.Message, error) {
			return nil, nil
		},
	}

	handler := New(repo)

	payload, err := json.Marshal(uploadMessageRequest{Body: "hello"})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewReader(payload))
	rec := httptest.NewRecorder()

	handler.uploadMessage(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
}

func TestReadMessages(t *testing.T) {
	createdAt := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	repo := fakeRepository{
		createFn: func(ctx context.Context, body string) (model.Message, error) {
			return model.Message{}, nil
		},
		listFn: func(ctx context.Context, limit, offset int) ([]model.Message, error) {
			if limit != 10 || offset != 5 {
				t.Fatalf("expected limit 10 offset 5, got %d %d", limit, offset)
			}
			return []model.Message{{ID: 1, Body: "hello", CreatedAt: createdAt}}, nil
		},
	}

	handler := New(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/messages?limit=10&offset=5", nil)
	rec := httptest.NewRecorder()

	handler.readMessages(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
