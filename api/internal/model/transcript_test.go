package model

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMeetingInputFixture_UnmarshalRoundTrip(t *testing.T) {
	// This fixture is defined by docs/ticket/T-003.
	fixturePath := filepath.Clean("../../../docs/fixtures/t-003/meeting_input.example.json")

	b, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture %s: %v", fixturePath, err)
	}

	var in MeetingInput
	if err := json.Unmarshal(b, &in); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	if in.SchemaVersion != "v1" {
		t.Fatalf("schema_version: got %q want %q", in.SchemaVersion, "v1")
	}
	if len(in.Agenda) == 0 {
		t.Fatalf("agenda: expected at least 1 item")
	}
	if len(in.Transcript.Turns) == 0 {
		t.Fatalf("transcript.turns: expected at least 1 turn")
	}
	if in.Transcript.Turns[0].Text == "" {
		t.Fatalf("first turn text: expected non-empty")
	}

	// Round-trip: ensure it can be marshaled and parsed back.
	outBytes, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out MeetingInput
	if err := json.Unmarshal(outBytes, &out); err != nil {
		t.Fatalf("unmarshal marshaled: %v", err)
	}

	if out.SchemaVersion != "v1" {
		t.Fatalf("schema_version after roundtrip: got %q want %q", out.SchemaVersion, "v1")
	}
	if len(out.Transcript.Turns) != len(in.Transcript.Turns) {
		t.Fatalf("turn count after roundtrip: got %d want %d", len(out.Transcript.Turns), len(in.Transcript.Turns))
	}
}
