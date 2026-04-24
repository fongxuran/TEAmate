package notion

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	notionconnector "teammate/internal/connector/notion"
	"teammate/internal/model"
)

func TestStatus_DefaultsWhenNilClient(t *testing.T) {
	h := New(nil)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/integrations/notion/status", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}

	var st notionconnector.Status
	if err := json.NewDecoder(bytes.NewReader(rec.Body.Bytes())).Decode(&st); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if st.Configured {
		t.Fatalf("expected not configured")
	}
	if !st.DryRun {
		t.Fatalf("expected dry_run true")
	}
}

func TestCreatePage_DryRun_OK(t *testing.T) {
	cfg := notionconnector.Config{APIKey: "k", DatabaseID: "db", TitleProperty: "Name", DryRun: true}
	client := notionconnector.NewClient(cfg)
	h := New(client)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	draft := model.TicketDraft{Title: "Hello", Description: "World", SourceActionItemID: "ai-1"}
	b, _ := json.Marshal(draft)

	req := httptest.NewRequest(http.MethodPost, "/integrations/notion/pages", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%s", rec.Code, rec.Body.String())
	}

	var ref struct {
		DryRun bool `json:"dry_run"`
	}
	if err := json.NewDecoder(bytes.NewReader(rec.Body.Bytes())).Decode(&ref); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !ref.DryRun {
		t.Fatalf("expected dry_run true")
	}
}

func TestCreatePage_InvalidJSON(t *testing.T) {
	cfg := notionconnector.Config{APIKey: "k", DatabaseID: "db", TitleProperty: "Name", DryRun: true}
	client := notionconnector.NewClient(cfg)
	h := New(client)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/integrations/notion/pages", bytes.NewReader([]byte("{")))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d", rec.Code)
	}
}
