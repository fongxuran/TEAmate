package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateTextSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != messagesPath {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("x-api-key"); got != "test-key" {
			t.Fatalf("unexpected api key header: %s", got)
		}
		if got := r.Header.Get("anthropic-version"); got != "2023-06-01" {
			t.Fatalf("unexpected version header: %s", got)
		}

		var req MessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "claude-test" {
			t.Fatalf("unexpected model: %s", req.Model)
		}
		if req.MaxTokens != 32 {
			t.Fatalf("unexpected max tokens: %d", req.MaxTokens)
		}
		if req.System != "system" {
			t.Fatalf("unexpected system: %s", req.System)
		}
		if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
			t.Fatalf("unexpected messages: %#v", req.Messages)
		}

		resp := MessageResponse{
			ID:   "msg_1",
			Type: "message",
			Role: "assistant",
			Content: []ContentBlock{
				{Type: "text", Text: "Hello there"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(Config{
		BaseURL:      server.URL,
		APIKey:       "test-key",
		Version:      "2023-06-01",
		DefaultModel: "claude-test",
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	output, err := client.GenerateText(context.Background(), GenerateTextInput{
		Prompt:    "Hello",
		MaxTokens: 32,
		System:    "system",
	})
	if err != nil {
		t.Fatalf("generate text: %v", err)
	}
	if output.Text != "Hello there" {
		t.Fatalf("unexpected output text: %s", output.Text)
	}
}

func TestGenerateTextErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: ErrorDetail{Type: "invalid_request_error", Message: "bad request"}})
	}))
	defer server.Close()

	client, err := NewClient(Config{
		BaseURL:      server.URL,
		APIKey:       "test-key",
		DefaultModel: "claude-test",
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.GenerateText(context.Background(), GenerateTextInput{
		Prompt:    "Hello",
		MaxTokens: 32,
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateTextPlainError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server failed"))
	}))
	defer server.Close()

	client, err := NewClient(Config{
		BaseURL:      server.URL,
		APIKey:       "test-key",
		DefaultModel: "claude-test",
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.GenerateText(context.Background(), GenerateTextInput{
		Prompt:    "Hello",
		MaxTokens: 32,
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "server failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}
