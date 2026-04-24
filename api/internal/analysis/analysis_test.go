package analysis

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"teammate/internal/model"
)

func mustLoadMeetingInputFixture(t *testing.T) model.MeetingInput {
	t.Helper()
	fixturePath := filepath.Clean("../../../docs/fixtures/t-003/meeting_input.example.json")
	b, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture %s: %v", fixturePath, err)
	}
	var in model.MeetingInput
	if err := json.Unmarshal(b, &in); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	return in
}

func TestParseTranscriptText_SpeakerParsing(t *testing.T) {
	input := strings.Join([]string{
		"Sam: Welcome everyone.",
		"Quick update on scope.",
		"Priya: UI work is moving.",
	}, "\n")

	got := ParseTranscriptText(input)
	if len(got.Turns) != 3 {
		t.Fatalf("turn count: got %d want %d", len(got.Turns), 3)
	}
	if got.Turns[0].Speaker == nil || *got.Turns[0].Speaker != "Sam" {
		t.Fatalf("turn 0 speaker: got %#v want %q", got.Turns[0].Speaker, "Sam")
	}
	if got.Turns[0].Text != "Welcome everyone." {
		t.Fatalf("turn 0 text: got %q", got.Turns[0].Text)
	}
	if got.Turns[1].Speaker != nil {
		t.Fatalf("turn 1 speaker: expected nil, got %#v", got.Turns[1].Speaker)
	}
	if got.Turns[1].Text != "Quick update on scope." {
		t.Fatalf("turn 1 text: got %q", got.Turns[1].Text)
	}
	if got.Turns[2].Speaker == nil || *got.Turns[2].Speaker != "Priya" {
		t.Fatalf("turn 2 speaker: got %#v want %q", got.Turns[2].Speaker, "Priya")
	}
}

func TestAnalyze_DriftScoringShape(t *testing.T) {
	meeting := mustLoadMeetingInputFixture(t)
	cfg := DefaultConfig()

	res1 := Analyze(meeting, cfg, nil)
	res2 := Analyze(meeting, cfg, nil)

	if res1.SchemaVersion != "v1" {
		t.Fatalf("schema_version: got %q want %q", res1.SchemaVersion, "v1")
	}
	if len(res1.Segments) == 0 {
		t.Fatalf("segments: expected at least 1")
	}
	if !reflect.DeepEqual(res1.Segments, res2.Segments) {
		t.Fatalf("drift scoring is not deterministic")
	}

	for i, seg := range res1.Segments {
		if seg.Segment.SegmentID == "" {
			t.Fatalf("segment %d id: expected non-empty", i)
		}
		if seg.BestScore < 0 || seg.BestScore > 1 {
			t.Fatalf("segment %d best_score: got %v", i, seg.BestScore)
		}
		if seg.BestAgendaItemID == nil || *seg.BestAgendaItemID == "" {
			t.Fatalf("segment %d best_agenda_item_id: expected non-empty", i)
		}
		if seg.BestAgendaTitle == nil || *seg.BestAgendaTitle == "" {
			t.Fatalf("segment %d best_agenda_title: expected non-empty", i)
		}
		if seg.FeedbackOverride != nil {
			t.Fatalf("segment %d feedback_override: expected nil", i)
		}
	}
}

func TestAnalyze_ActionItemSchema(t *testing.T) {
	meeting := model.MeetingInput{
		SchemaVersion: "v1",
		Transcript: model.Transcript{
			Turns: []model.TranscriptTurn{
				{Text: "Action: draft the rollout plan."},
				{Text: "Action item: publish the changelog."},
			},
		},
	}
	res := Analyze(meeting, DefaultConfig(), nil)

	if len(res.ActionItems) < 2 {
		t.Fatalf("action_items: got %d want >= 2", len(res.ActionItems))
	}

	for i, item := range res.ActionItems {
		if strings.TrimSpace(item.ActionItemID) == "" {
			t.Fatalf("action_item %d id: expected non-empty", i)
		}
		if strings.TrimSpace(item.Title) == "" {
			t.Fatalf("action_item %d title: expected non-empty", i)
		}
		if item.Confidence <= 0 || item.Confidence > 1 {
			t.Fatalf("action_item %d confidence: got %v", i, item.Confidence)
		}

		b, err := json.Marshal(item)
		if err != nil {
			t.Fatalf("action_item %d marshal: %v", i, err)
		}
		var payload map[string]any
		if err := json.Unmarshal(b, &payload); err != nil {
			t.Fatalf("action_item %d unmarshal: %v", i, err)
		}
		if _, ok := payload["action_item_id"]; !ok {
			t.Fatalf("action_item %d missing action_item_id", i)
		}
		if _, ok := payload["title"]; !ok {
			t.Fatalf("action_item %d missing title", i)
		}
		if _, ok := payload["confidence"]; !ok {
			t.Fatalf("action_item %d missing confidence", i)
		}
	}
}

func TestAnalyze_SmokeFixture(t *testing.T) {
	meeting := mustLoadMeetingInputFixture(t)
	res := Analyze(meeting, DefaultConfig(), nil)

	if res.SchemaVersion != "v1" {
		t.Fatalf("schema_version: got %q want %q", res.SchemaVersion, "v1")
	}
	if res.GeneratedAt.IsZero() {
		t.Fatalf("generated_at: expected non-zero")
	}
	if strings.TrimSpace(res.Summary) == "" {
		t.Fatalf("summary: expected non-empty")
	}
	if len(res.Segments) == 0 {
		t.Fatalf("segments: expected at least 1")
	}
	if len(res.TicketDrafts) != len(res.ActionItems) {
		t.Fatalf("ticket_drafts: got %d want %d", len(res.TicketDrafts), len(res.ActionItems))
	}

	ids := make(map[string]struct{}, len(res.ActionItems))
	for _, item := range res.ActionItems {
		ids[item.ActionItemID] = struct{}{}
	}
	for i, draft := range res.TicketDrafts {
		if _, ok := ids[draft.SourceActionItemID]; !ok {
			t.Fatalf("draft %d source_action_item_id: %q not found", i, draft.SourceActionItemID)
		}
	}
}
