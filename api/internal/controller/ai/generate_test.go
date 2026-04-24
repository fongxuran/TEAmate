package ai

import (
	"context"
	"errors"
	"testing"

	"teammate/internal/gateway/anthropic"
)

type mockGateway struct {
	lastInput anthropic.GenerateTextInput
	output    anthropic.GenerateTextOutput
	err       error
}

func (m *mockGateway) GenerateText(ctx context.Context, input anthropic.GenerateTextInput) (anthropic.GenerateTextOutput, error) {
	m.lastInput = input
	if m.err != nil {
		return anthropic.GenerateTextOutput{}, m.err
	}
	return m.output, nil
}

func TestGenerateTextValidatesPrompt(t *testing.T) {
	ctrl := New(&mockGateway{}, Options{DefaultModel: "claude"})
	_, err := ctrl.GenerateText(context.Background(), GenerateTextInput{Prompt: "  "})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestGenerateTextUsesDefaults(t *testing.T) {
	mg := &mockGateway{
		output: anthropic.GenerateTextOutput{
			Text: "done",
			Raw: anthropic.MessageResponse{
				Usage: anthropic.Usage{InputTokens: 5, OutputTokens: 7},
			},
		},
	}
	ctrl := New(mg, Options{DefaultModel: "claude", DefaultMaxTokens: 128})

	out, err := ctrl.GenerateText(context.Background(), GenerateTextInput{Prompt: "hi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Text != "done" {
		t.Fatalf("unexpected output: %s", out.Text)
	}
	if mg.lastInput.Model != "claude" {
		t.Fatalf("unexpected model: %s", mg.lastInput.Model)
	}
	if mg.lastInput.MaxTokens != 128 {
		t.Fatalf("unexpected max tokens: %d", mg.lastInput.MaxTokens)
	}
}

func TestGenerateTextPropagatesErrors(t *testing.T) {
	mg := &mockGateway{err: errors.New("boom")}
	ctrl := New(mg, Options{DefaultModel: "claude", DefaultMaxTokens: 64})

	_, err := ctrl.GenerateText(context.Background(), GenerateTextInput{Prompt: "hi"})
	if err == nil {
		t.Fatalf("expected error")
	}
}
