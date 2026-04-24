package model

import "time"

// AgendaItem is a canonical agenda entry for a meeting.
//
// v1 JSON schema:
// - id: required, opaque string
// - title: required
// - description: optional
// - keywords: optional
type AgendaItem struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description *string  `json:"description,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
}

// TranscriptTurn is a single speaker turn.
//
// v1 JSON schema:
// - timestamp: optional RFC3339
// - speaker: optional
// - text: required
type TranscriptTurn struct {
	Timestamp *time.Time `json:"timestamp,omitempty"`
	Speaker   *string    `json:"speaker,omitempty"`
	Text      string     `json:"text"`
}

// Transcript is a turn-based transcript.
// meeting_id and meeting_name are both optional, but callers should provide at least one.
type Transcript struct {
	MeetingID   *string          `json:"meeting_id,omitempty"`
	MeetingName *string          `json:"meeting_name,omitempty"`
	Turns       []TranscriptTurn `json:"turns"`
}

// RealtimeMessage represents a realtime textbox event sent by clients.
// Exactly one of TextDelta or Text should be set.
//
// For MVP, TextDelta is the primary append-only mode.
type RealtimeMessage struct {
	ClientID  string    `json:"client_id"`
	Timestamp time.Time `json:"timestamp"`
	TextDelta *string   `json:"text_delta,omitempty"`
	Text      *string   `json:"text,omitempty"`
	Author    *string   `json:"author,omitempty"`
}

// MeetingInput is the recommended wrapper payload for downstream processing.
// It packages Agenda + Transcript in a repeatable, versioned format.
//
// v1 requires SchemaVersion == "v1".
type MeetingInput struct {
	SchemaVersion string       `json:"schema_version"`
	Source        *string      `json:"source,omitempty"`
	Agenda        []AgendaItem `json:"agenda"`
	Transcript    Transcript   `json:"transcript"`
}
