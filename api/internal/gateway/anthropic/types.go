package anthropic

// GenerateTextInput holds parameters for an Anthropic messages request.
type GenerateTextInput struct {
	Prompt    string
	Model     string
	MaxTokens int
	System    string
}

// GenerateTextOutput holds the parsed response text plus the raw response.
type GenerateTextOutput struct {
	Text string
	Raw  MessageResponse
}

// MessageRequest is the Anthropic messages API request payload.
type MessageRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
}

// MessageResponse is a subset of the Anthropic messages API response.
type MessageResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence"`
	Usage        Usage          `json:"usage"`
}

// Message represents an input message.
type Message struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// ContentBlock is a single message content block.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Usage is the token usage returned by Anthropic.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ErrorResponse is the error envelope from Anthropic.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail holds error metadata from Anthropic.
type ErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
