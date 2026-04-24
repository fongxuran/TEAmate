package notion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"teammate/internal/model"
)

func TestClient_Status_ConfiguredAndDryRun(t *testing.T) {
	c := NewClient(Config{APIKey: "k", DatabaseID: "db", TitleProperty: "Name", DryRun: true})
	st := c.Status()
	if !st.Configured {
		t.Fatalf("expected configured")
	}
	if !st.DryRun {
		t.Fatalf("expected dry run")
	}
}

func TestClient_CreateTask_NotConfigured(t *testing.T) {
	c := NewClient(Config{})
	_, err := c.CreateTask(context.Background(), model.TicketDraft{Title: "x", SourceActionItemID: "ai"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != ErrNotConfigured {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestClient_CreateTask_DryRun_NoNetwork(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"x","url":"u"}`))
	}))
	defer srv.Close()

	cfg := Config{APIKey: "k", DatabaseID: "db", TitleProperty: "Name", DryRun: true, BaseURL: srv.URL}
	c := NewClient(cfg)
	ref, err := c.CreateTask(context.Background(), model.TicketDraft{Title: "Hello", SourceActionItemID: "ai"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !ref.DryRun {
		t.Fatalf("expected dry run ref")
	}
	if called {
		t.Fatalf("expected no HTTP call in dry run")
	}
}

func TestClient_CreateTask_CreatesPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method: got %s", r.Method)
		}
		if r.URL.Path != "/v1/pages" {
			t.Fatalf("path: got %s", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Fatalf("missing bearer auth")
		}
		if r.Header.Get("Notion-Version") == "" {
			t.Fatalf("missing notion version")
		}

		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		props, ok := req["properties"].(map[string]any)
		if !ok {
			t.Fatalf("properties missing or wrong type")
		}
		if _, ok := props["Name"]; !ok {
			t.Fatalf("missing Name property")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"page-1","url":"https://notion.so/page-1"}`))
	}))
	defer srv.Close()

	cfg := Config{APIKey: "k", DatabaseID: "db", TitleProperty: "Name", DryRun: false, BaseURL: srv.URL}
	c := NewClient(cfg)

	ref, err := c.CreateTask(context.Background(), model.TicketDraft{Title: "Ship it", Description: "Do thing", SourceActionItemID: "ai-1"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if ref.DryRun {
		t.Fatalf("expected non-dry-run")
	}
	if ref.ID != "page-1" {
		t.Fatalf("id: got %q", ref.ID)
	}
	if ref.URL == "" {
		t.Fatalf("expected url")
	}
}
