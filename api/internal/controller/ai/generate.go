package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"teammate/internal/gateway/anthropic"
)

// GenerateTextInput holds inputs for text generation.
type GenerateTextInput struct {
	Prompt    string
	Model     string
	MaxTokens int
	System    string
}

// GenerateTextOutput is the controller output for text generation.
type GenerateTextOutput struct {
	Text  string
	Usage Usage
}

// Usage provides token usage details.
type Usage struct {
	InputTokens  int
	OutputTokens int
}

// GenerateText validates input and calls Anthropic to generate text.
func (i impl) GenerateText(ctx context.Context, input GenerateTextInput) (GenerateTextOutput, error) {
	if i.gateway == nil {
		return GenerateTextOutput{}, errors.New("anthropic gateway is not configured")
	}

	prompt := strings.TrimSpace(input.Prompt)
	if prompt == "" {
		return GenerateTextOutput{}, errors.New("prompt is required")
	}

	model := strings.TrimSpace(input.Model)
	if model == "" {
		model = i.defaultModel
	}
	if model == "" {
		return GenerateTextOutput{}, errors.New("model is required")
	}

	maxTokens := input.MaxTokens
	if maxTokens <= 0 {
		maxTokens = i.defaultMaxTokens
	}
	if maxTokens <= 0 {
		return GenerateTextOutput{}, errors.New("max tokens must be positive")
	}

	resp, err := i.gateway.GenerateText(ctx, anthropic.GenerateTextInput{
		Prompt:    prompt,
		Model:     model,
		MaxTokens: maxTokens,
		System:    strings.TrimSpace(input.System),
	})
	if err != nil {
		return GenerateTextOutput{}, fmt.Errorf("generate text: %w", err)
	}

	return GenerateTextOutput{
		Text: resp.Text,
		Usage: Usage{
			InputTokens:  resp.Raw.Usage.InputTokens,
			OutputTokens: resp.Raw.Usage.OutputTokens,
		},
	}, nil
}
