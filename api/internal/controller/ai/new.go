package ai

import (
	"context"
	"strings"

	"teammate/internal/gateway/anthropic"
)

// Gateway defines the Anthropic gateway dependency used by the controller.
type Gateway interface {
	GenerateText(ctx context.Context, input anthropic.GenerateTextInput) (anthropic.GenerateTextOutput, error)
}

// Controller represents the specification of this pkg.
type Controller interface {
	GenerateText(ctx context.Context, input GenerateTextInput) (GenerateTextOutput, error)
}

// Options configures defaults for the AI controller.
type Options struct {
	DefaultModel     string
	DefaultMaxTokens int
}

// New initializes a new Controller instance and returns it.
func New(gateway Gateway, opts Options) Controller {
	return impl{
		gateway:          gateway,
		defaultModel:     strings.TrimSpace(opts.DefaultModel),
		defaultMaxTokens: opts.DefaultMaxTokens,
	}
}

type impl struct {
	gateway          Gateway
	defaultModel     string
	defaultMaxTokens int
}
