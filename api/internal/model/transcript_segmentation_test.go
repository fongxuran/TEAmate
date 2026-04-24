package model

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func mustLoadMeetingInputFixture(t *testing.T) MeetingInput {
	t.Helper()
	fixturePath := filepath.Clean("../../../docs/fixtures/t-003/meeting_input.example.json")
	b, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture %s: %v", fixturePath, err)
	}
	var in MeetingInput
	if err := json.Unmarshal(b, &in); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	return in
}

func TestSegmentTranscript_DeterministicFixture(t *testing.T) {
	in := mustLoadMeetingInputFixture(t)

	opts := SegmentOptions{
		MaxTokens:            25,
		IncludeSpeakerLabels: true,
	}

	got1 := SegmentTranscript(in.Transcript, opts)
	got2 := SegmentTranscript(in.Transcript, opts)

	if !reflect.DeepEqual(got1, got2) {
		t.Fatalf("segmentation is not deterministic; got1 != got2")
	}
	if len(got1) == 0 {
		t.Fatalf("expected at least 1 segment")
	}

	want := []Segment{
		{
			SegmentID:    "seg-0-1",
			StartTurnIdx: 0,
			EndTurnIdx:   1,
			Text: strings.Join([]string{
				"Sam: Kicking off. Goal today: status, MVP decisions, then actions.",
				"Alex: Status: backend messages API is in place; next is transcript schema so segmentation can be deterministic.",
			}, "\n"),
		},
		{
			SegmentID:    "seg-2-3",
			StartTurnIdx: 2,
			EndTurnIdx:   3,
			Text: strings.Join([]string{
				"Priya: Frontend note: local proxy route works; we should avoid relying on npx for bootstrapping.",
				"Sam: Cool. Any blockers?",
			}, "\n"),
		},
		{
			SegmentID:    "seg-4-4",
			StartTurnIdx: 4,
			EndTurnIdx:   4,
			Text:         "Priya: No blockers, but we need a single canonical MeetingInput JSON so tests can reuse fixtures.",
		},
		{
			SegmentID:    "seg-5-5",
			StartTurnIdx: 5,
			EndTurnIdx:   5,
			Text:         "Alex: For MVP scope: realtime textbox stream is primary; pasted transcript is fallback; JSON upload is stretch but helps repeatability.",
		},
		{
			SegmentID:    "seg-6-7",
			StartTurnIdx: 6,
			EndTurnIdx:   7,
			Text: strings.Join([]string{
				"Sam: Decision: adopt v1 schema with schema_version field and RFC3339 timestamps.",
				"Sam: Action items: Alex drafts segmentation rules; Priya wires UI input to the schema.",
			}, "\n"),
		},
	}

	if !reflect.DeepEqual(got1, want) {
		t.Fatalf("unexpected segmentation output\n--- got ---\n%#v\n--- want ---\n%#v", got1, want)
	}

	for _, seg := range got1 {
		if seg.SegmentID == "" {
			t.Fatalf("segment_id: expected non-empty")
		}
		if strings.TrimSpace(seg.Text) == "" {
			t.Fatalf("segment %s text: expected non-empty", seg.SegmentID)
		}
		if seg.StartTurnIdx < 0 || seg.EndTurnIdx < seg.StartTurnIdx {
			t.Fatalf("segment %s indices: got %d..%d", seg.SegmentID, seg.StartTurnIdx, seg.EndTurnIdx)
		}
	}

	rendered := RenderSegmentsPlaintext(got1)
	if !strings.Contains(rendered, "--- seg-0-1 turns[0..1] ---") {
		t.Fatalf("renderer output missing boundary header; got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "Sam: Kicking off") {
		t.Fatalf("renderer output missing segment content; got:\n%s", rendered)
	}
}
