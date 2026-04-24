package export

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"teammate/internal/model"
)

func TestExportTicketDrafts_WritesJSONMarkdownCSV(t *testing.T) {
	meetingID := "m-2026"
	meetingName := "Sync"
	meeting := model.MeetingInput{
		SchemaVersion: "v1",
		Transcript: model.Transcript{
			MeetingID:   &meetingID,
			MeetingName: &meetingName,
		},
	}

	desc1 := "Add JSON + MD exporters"
	owner1 := "Priya"
	items := []model.ActionItem{
		{
			ActionItemID:     "ai-1",
			Title:            "Export ticket drafts",
			Description:      &desc1,
			Owner:            &owner1,
			SourceSegmentIDs: []string{"seg-0-1"},
			Confidence:       0.8,
		},
		{
			ActionItemID:     "ai-2",
			Title:            "Add CSV exporter",
			SourceSegmentIDs: []string{"seg-2-3", "seg-4-5"},
			Confidence:       0.6,
		},
	}

	outDir := t.TempDir()
	paths, err := ExportTicketDrafts(outDir, meeting, items, TicketDraftFormats{JSON: true, Markdown: true, CSV: true})
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	if paths.JSONPath == "" || paths.MarkdownPath == "" || paths.CSVPath == "" {
		t.Fatalf("expected all paths to be set; got %#v", paths)
	}

	// JSON
	b, err := os.ReadFile(paths.JSONPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var exp TicketDraftExport
	if err := json.Unmarshal(b, &exp); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if exp.SchemaVersion != "v1" {
		t.Fatalf("schema_version: got %q want %q", exp.SchemaVersion, "v1")
	}
	if exp.MeetingID == nil || *exp.MeetingID != meetingID {
		t.Fatalf("meeting_id: got %#v want %q", exp.MeetingID, meetingID)
	}
	if len(exp.Drafts) != len(items) {
		t.Fatalf("draft count: got %d want %d", len(exp.Drafts), len(items))
	}
	if exp.Drafts[0].SourceActionItemID != "ai-1" {
		t.Fatalf("first draft source_action_item_id: got %q", exp.Drafts[0].SourceActionItemID)
	}
	if exp.Drafts[0].SourceMeetingID == nil || *exp.Drafts[0].SourceMeetingID != meetingID {
		t.Fatalf("first draft source_meeting_id: got %#v want %q", exp.Drafts[0].SourceMeetingID, meetingID)
	}
	if got := strings.Join(exp.Drafts[1].SourceSegmentIDs, ","); got != "seg-2-3,seg-4-5" {
		t.Fatalf("second draft source_segment_ids: got %q", got)
	}

	// Markdown
	md, err := os.ReadFile(paths.MarkdownPath)
	if err != nil {
		t.Fatalf("read md: %v", err)
	}
	mds := string(md)
	if !strings.Contains(mds, "# Ticket Drafts") {
		t.Fatalf("md missing header")
	}
	if !strings.Contains(mds, "source_action_item_id: ai-1") {
		t.Fatalf("md missing traceability")
	}
	if !strings.Contains(mds, "meeting_id: "+meetingID) {
		t.Fatalf("md missing meeting_id")
	}

	// CSV
	f, err := os.Open(paths.CSVPath)
	if err != nil {
		t.Fatalf("open csv: %v", err)
	}
	defer f.Close()
	cr := csv.NewReader(f)
	recs, err := cr.ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if len(recs) != len(items)+1 {
		t.Fatalf("csv row count: got %d want %d", len(recs), len(items)+1)
	}
	if recs[0][0] != "title" {
		t.Fatalf("csv header first col: got %q", recs[0][0])
	}
	if recs[1][0] != "Export ticket drafts" {
		t.Fatalf("csv first record title: got %q", recs[1][0])
	}
	if recs[1][5] != "ai-1" {
		t.Fatalf("csv source_action_item_id: got %q", recs[1][5])
	}

	// Ensure file names are stable (handy for local use).
	if filepath.Base(paths.JSONPath) != "ticket_drafts.json" {
		t.Fatalf("json filename: got %q", filepath.Base(paths.JSONPath))
	}
}

func TestExportTicketDrafts_DefaultFormatsAreJSONAndMarkdown(t *testing.T) {
	meeting := model.MeetingInput{SchemaVersion: "v1"}
	outDir := t.TempDir()
	paths, err := ExportTicketDrafts(outDir, meeting, nil, TicketDraftFormats{})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if paths.JSONPath == "" || paths.MarkdownPath == "" {
		t.Fatalf("expected json+markdown by default; got %#v", paths)
	}
	if paths.CSVPath != "" {
		t.Fatalf("expected csv to be empty by default; got %#v", paths)
	}
}
