package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"teammate/internal/connector"
	"teammate/internal/model"
)

// ErrNotConfigured is returned when the Notion connector is not configured.
var ErrNotConfigured = errors.New("notion connector not configured")

// Config configures the Notion connector.
//
// Required for live create:
// - APIKey
// - DatabaseID
// - TitleProperty (defaults to "Name")
//
// Safety:
// - DryRun defaults to true
// - even when configured, DryRun prevents any external side-effects.
type Config struct {
	APIKey        string
	DatabaseID    string
	TitleProperty string
	NotionVersion string
	DryRun        bool

	BaseURL     string
	HTTPClient  *http.Client
	UserAgent   string
	MaxChildren int
}

// Status is returned by the integration status endpoint.
type Status struct {
	Configured bool   `json:"configured"`
	DryRun     bool   `json:"dry_run"`
	DatabaseID string `json:"database_id,omitempty"`
}

// LoadConfigFromEnv builds a connector config from environment variables.
//
// Env vars:
// - NOTION_API_KEY
// - NOTION_DATABASE_ID
// - NOTION_TITLE_PROPERTY (optional; default: Name)
// - NOTION_VERSION (optional; default: 2026-03-11)
// - NOTION_DRY_RUN (optional; default: true)
//
// Optional for testing:
// - NOTION_BASE_URL
func LoadConfigFromEnv() Config {
	cfg := Config{
		APIKey:        strings.TrimSpace(os.Getenv("NOTION_API_KEY")),
		DatabaseID:    strings.TrimSpace(os.Getenv("NOTION_DATABASE_ID")),
		TitleProperty: strings.TrimSpace(os.Getenv("NOTION_TITLE_PROPERTY")),
		NotionVersion: strings.TrimSpace(os.Getenv("NOTION_VERSION")),
		DryRun:        envBool("NOTION_DRY_RUN", true),
		BaseURL:       strings.TrimSpace(os.Getenv("NOTION_BASE_URL")),
		UserAgent:     strings.TrimSpace(os.Getenv("NOTION_USER_AGENT")),
		MaxChildren:   100,
	}

	if cfg.TitleProperty == "" {
		cfg.TitleProperty = "Name"
	}
	if cfg.NotionVersion == "" {
		// per Notion docs: latest version at time of writing
		cfg.NotionVersion = "2026-03-11"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.notion.com"
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "TEAmate-MVP (optional Notion connector)"
	}

	return cfg
}

func (c Config) configured() bool {
	return c.APIKey != "" && c.DatabaseID != "" && c.TitleProperty != ""
}

// Client creates Notion pages from ticket drafts.
type Client struct {
	cfg Config
}

// NewClient returns a Notion connector client.
func NewClient(cfg Config) *Client {
	// Apply the same defaults as LoadConfigFromEnv so programmatic construction behaves well.
	if strings.TrimSpace(cfg.TitleProperty) == "" {
		cfg.TitleProperty = "Name"
	}
	if strings.TrimSpace(cfg.NotionVersion) == "" {
		cfg.NotionVersion = "2026-03-11"
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = "https://api.notion.com"
	}
	if strings.TrimSpace(cfg.UserAgent) == "" {
		cfg.UserAgent = "TEAmate-MVP (optional Notion connector)"
	}

	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	if cfg.MaxChildren <= 0 {
		cfg.MaxChildren = 100
	}
	return &Client{cfg: cfg}
}

func (c *Client) Status() Status {
	st := Status{Configured: c.cfg.configured(), DryRun: c.cfg.DryRun}
	if c.cfg.DatabaseID != "" {
		st.DatabaseID = c.cfg.DatabaseID
	}
	return st
}

// CreateTask creates a Notion page under the configured database.
//
// In DryRun mode it returns a TaskRef with DryRun=true and does not perform
// any external requests.
func (c *Client) CreateTask(ctx context.Context, draft model.TicketDraft) (connector.TaskRef, error) {
	if !c.cfg.configured() {
		return connector.TaskRef{}, ErrNotConfigured
	}
	if c.cfg.DryRun {
		return connector.TaskRef{DryRun: true}, nil
	}

	payload, err := buildCreatePageRequest(c.cfg, draft)
	if err != nil {
		return connector.TaskRef{}, err
	}

	url := strings.TrimRight(c.cfg.BaseURL, "/") + "/v1/pages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return connector.TaskRef{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Notion-Version", c.cfg.NotionVersion)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.cfg.UserAgent != "" {
		req.Header.Set("User-Agent", c.cfg.UserAgent)
	}

	res, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return connector.TaskRef{}, fmt.Errorf("notion request failed: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20)) // 1MB
	if err != nil {
		return connector.TaskRef{}, fmt.Errorf("read notion response: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		msg := strings.TrimSpace(string(body))
		var ne notionError
		if err := json.Unmarshal(body, &ne); err == nil && ne.Message != "" {
			msg = ne.Message
		}
		if msg == "" {
			msg = "notion request failed"
		}
		return connector.TaskRef{}, fmt.Errorf("notion create page: status %d: %s", res.StatusCode, msg)
	}

	var created createPageResponse
	if err := json.Unmarshal(body, &created); err != nil {
		return connector.TaskRef{}, fmt.Errorf("decode notion response: %w", err)
	}

	return connector.TaskRef{ID: created.ID, URL: created.URL, DryRun: false}, nil
}

type notionError struct {
	Object  string `json:"object"`
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type createPageResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func buildCreatePageRequest(cfg Config, draft model.TicketDraft) ([]byte, error) {
	title := strings.TrimSpace(draft.Title)
	if title == "" {
		return nil, fmt.Errorf("draft title is required")
	}

	children := make([]any, 0, 8)
	if desc := strings.TrimSpace(draft.Description); desc != "" {
		children = append(children, paragraphBlock(desc))
	}

	// Add minimal, human-readable metadata.
	if len(draft.Labels) > 0 {
		children = append(children, paragraphBlock("Labels: "+strings.Join(draft.Labels, ", ")))
	}
	children = append(children, paragraphBlock("Source action item: "+draft.SourceActionItemID))
	if draft.SourceMeetingName != nil && strings.TrimSpace(*draft.SourceMeetingName) != "" {
		children = append(children, paragraphBlock("Meeting: "+strings.TrimSpace(*draft.SourceMeetingName)))
	}
	if draft.SourceMeetingID != nil && strings.TrimSpace(*draft.SourceMeetingID) != "" {
		children = append(children, paragraphBlock("Meeting ID: "+strings.TrimSpace(*draft.SourceMeetingID)))
	}
	if len(draft.SourceSegmentIDs) > 0 {
		children = append(children, paragraphBlock("Segments: "+strings.Join(draft.SourceSegmentIDs, ", ")))
	}

	if cfg.MaxChildren > 0 && len(children) > cfg.MaxChildren {
		children = children[:cfg.MaxChildren]
	}

	req := map[string]any{
		"parent": map[string]any{
			"database_id": cfg.DatabaseID,
		},
		"properties": map[string]any{
			cfg.TitleProperty: map[string]any{
				"title": []any{map[string]any{"text": map[string]any{"content": title}}},
			},
		},
	}
	if len(children) > 0 {
		req["children"] = children
	}

	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal notion create page: %w", err)
	}
	return b, nil
}

func paragraphBlock(content string) map[string]any {
	return map[string]any{
		"object": "block",
		"type":   "paragraph",
		"paragraph": map[string]any{
			"rich_text": []any{map[string]any{
				"type": "text",
				"text": map[string]any{"content": content},
			}},
		},
	}
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
