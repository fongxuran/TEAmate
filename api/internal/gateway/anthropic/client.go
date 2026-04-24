package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.anthropic.com"
	messagesPath   = "/v1/messages"
	defaultTimeout = 30 * time.Second
)

// Client defines the Anthropic gateway interface.
type Client interface {
	GenerateText(ctx context.Context, input GenerateTextInput) (GenerateTextOutput, error)
}

// Config configures the Anthropic gateway client.
type Config struct {
	BaseURL      string
	APIKey       string
	Version      string
	DefaultModel string
	Timeout      time.Duration
	HTTPClient   *http.Client
}

// NewClient constructs a new Anthropic client.
func NewClient(cfg Config) (Client, error) {
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		return nil, errors.New("anthropic api key is required")
	}

	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout == 0 {
			timeout = defaultTimeout
		}
		httpClient = &http.Client{Timeout: timeout}
	}

	return &client{
		baseURL:      baseURL,
		apiKey:       apiKey,
		version:      strings.TrimSpace(cfg.Version),
		defaultModel: strings.TrimSpace(cfg.DefaultModel),
		httpClient:   httpClient,
	}, nil
}

type client struct {
	baseURL      string
	apiKey       string
	version      string
	defaultModel string
	httpClient   *http.Client
}

// GenerateText sends a non-streaming message request to Anthropic and returns the first text block.
func (c *client) GenerateText(ctx context.Context, input GenerateTextInput) (GenerateTextOutput, error) {
	prompt := strings.TrimSpace(input.Prompt)
	if prompt == "" {
		return GenerateTextOutput{}, errors.New("prompt is required")
	}

	model := strings.TrimSpace(input.Model)
	if model == "" {
		model = c.defaultModel
	}
	if model == "" {
		return GenerateTextOutput{}, errors.New("model is required")
	}
	if input.MaxTokens <= 0 {
		return GenerateTextOutput{}, errors.New("max tokens must be positive")
	}

	payload := MessageRequest{
		Model:     model,
		MaxTokens: input.MaxTokens,
		System:    strings.TrimSpace(input.System),
		Messages: []Message{
			{
				Role: "user",
				Content: []ContentBlock{
					{Type: "text", Text: prompt},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return GenerateTextOutput{}, fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + messagesPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return GenerateTextOutput{}, fmt.Errorf("build request: %w", err)
	}
	c.addHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GenerateTextOutput{}, fmt.Errorf("send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return GenerateTextOutput{}, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return GenerateTextOutput{}, parseErrorResponse(resp.StatusCode, respBody)
	}

	var parsed MessageResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return GenerateTextOutput{}, fmt.Errorf("decode response: %w", err)
	}

	text, ok := extractText(parsed.Content)
	if !ok {
		return GenerateTextOutput{}, errors.New("no text content returned")
	}

	return GenerateTextOutput{Text: text, Raw: parsed}, nil
}

func (c *client) addHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	if c.version != "" {
		req.Header.Set("anthropic-version", c.version)
	}
}

func extractText(blocks []ContentBlock) (string, bool) {
	for _, block := range blocks {
		if block.Type == "text" && strings.TrimSpace(block.Text) != "" {
			return block.Text, true
		}
	}
	return "", false
}

func parseErrorResponse(statusCode int, body []byte) error {
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil {
		if errResp.Error.Message != "" {
			return fmt.Errorf("anthropic error (%d): %s (%s)", statusCode, errResp.Error.Message, errResp.Error.Type)
		}
	}

	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		trimmed = http.StatusText(statusCode)
	}
	return fmt.Errorf("anthropic error (%d): %s", statusCode, trimmed)
}
