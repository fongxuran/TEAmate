package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"teammate/internal/model"
)

// TicketDraftFormats configures which export artifacts to generate.
// If all fields are false, ExportTicketDrafts defaults to JSON + Markdown.
type TicketDraftFormats struct {
	JSON     bool
	Markdown bool
	CSV      bool
}

// TicketDraftExport is the JSON on-disk wrapper format for exported ticket drafts.
// This is meant to be stable and human-inspectable.
type TicketDraftExport struct {
	SchemaVersion string              `json:"schema_version"`
	GeneratedAt   time.Time           `json:"generated_at"`
	MeetingID     *string             `json:"meeting_id,omitempty"`
	MeetingName   *string             `json:"meeting_name,omitempty"`
	Drafts        []model.TicketDraft `json:"drafts"`
}

// TicketDraftExportPaths returns the file paths that were written.
// A path may be empty when the corresponding format is disabled.
type TicketDraftExportPaths struct {
	JSONPath     string
	MarkdownPath string
	CSVPath      string
}

// ExportTicketDrafts converts action items into ticket drafts and exports them to disk.
//
// outDir defaults to "out" when empty.
func ExportTicketDrafts(outDir string, meeting model.MeetingInput, actionItems []model.ActionItem, formats TicketDraftFormats) (TicketDraftExportPaths, error) {
	if outDir == "" {
		outDir = "out"
	}
	if !formats.JSON && !formats.Markdown && !formats.CSV {
		formats.JSON = true
		formats.Markdown = true
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return TicketDraftExportPaths{}, fmt.Errorf("create out dir %s: %w", outDir, err)
	}

	drafts := model.TicketDraftsFromActionItems(meeting, actionItems)
	exp := TicketDraftExport{
		SchemaVersion: "v1",
		GeneratedAt:   time.Now().UTC(),
		MeetingID:     meeting.Transcript.MeetingID,
		MeetingName:   meeting.Transcript.MeetingName,
		Drafts:        drafts,
	}

	paths := TicketDraftExportPaths{}

	if formats.JSON {
		p := filepath.Join(outDir, "ticket_drafts.json")
		if err := writeTicketDraftsJSON(p, exp); err != nil {
			return TicketDraftExportPaths{}, err
		}
		paths.JSONPath = p
	}
	if formats.Markdown {
		p := filepath.Join(outDir, "ticket_drafts.md")
		if err := writeTicketDraftsMarkdown(p, exp); err != nil {
			return TicketDraftExportPaths{}, err
		}
		paths.MarkdownPath = p
	}
	if formats.CSV {
		p := filepath.Join(outDir, "ticket_drafts.csv")
		if err := writeTicketDraftsCSV(p, exp); err != nil {
			return TicketDraftExportPaths{}, err
		}
		paths.CSVPath = p
	}

	return paths, nil
}

func writeTicketDraftsJSON(path string, exp TicketDraftExport) error {
	b, err := json.MarshalIndent(exp, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal ticket drafts json: %w", err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func writeTicketDraftsMarkdown(path string, exp TicketDraftExport) error {
	content := RenderTicketDraftsMarkdown(exp)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// RenderTicketDraftsMarkdown renders the markdown export content in-memory.
func RenderTicketDraftsMarkdown(exp TicketDraftExport) string {
	var b strings.Builder
	b.WriteString("# Ticket Drafts\n\n")
	b.WriteString("- schema_version: ")
	b.WriteString(exp.SchemaVersion)
	b.WriteString("\n")
	b.WriteString("- generated_at: ")
	b.WriteString(exp.GeneratedAt.Format(time.RFC3339))
	b.WriteString("\n")
	if exp.MeetingID != nil {
		b.WriteString("- meeting_id: ")
		b.WriteString(*exp.MeetingID)
		b.WriteString("\n")
	}
	if exp.MeetingName != nil {
		b.WriteString("- meeting_name: ")
		b.WriteString(*exp.MeetingName)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if len(exp.Drafts) == 0 {
		b.WriteString("_No ticket drafts produced._\n")
		return b.String()
	}

	for i, d := range exp.Drafts {
		b.WriteString("## ")
		b.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, strings.TrimSpace(d.Title)))

		if d.Owner != nil && strings.TrimSpace(*d.Owner) != "" {
			b.WriteString("- owner: ")
			b.WriteString(strings.TrimSpace(*d.Owner))
			b.WriteString("\n")
		}
		if d.Priority != nil && strings.TrimSpace(*d.Priority) != "" {
			b.WriteString("- priority: ")
			b.WriteString(strings.TrimSpace(*d.Priority))
			b.WriteString("\n")
		}
		if len(d.Labels) > 0 {
			b.WriteString("- labels: ")
			b.WriteString(strings.Join(d.Labels, ", "))
			b.WriteString("\n")
		}
		b.WriteString("\n")

		if strings.TrimSpace(d.Description) != "" {
			b.WriteString(d.Description)
			b.WriteString("\n\n")
		}

		b.WriteString("**Traceability**\n\n")
		b.WriteString("- source_action_item_id: ")
		b.WriteString(d.SourceActionItemID)
		b.WriteString("\n")
		if d.SourceMeetingID != nil {
			b.WriteString("- source_meeting_id: ")
			b.WriteString(*d.SourceMeetingID)
			b.WriteString("\n")
		}
		if d.SourceMeetingName != nil {
			b.WriteString("- source_meeting_name: ")
			b.WriteString(*d.SourceMeetingName)
			b.WriteString("\n")
		}
		if len(d.SourceSegmentIDs) > 0 {
			b.WriteString("- source_segment_ids: ")
			b.WriteString(strings.Join(d.SourceSegmentIDs, ", "))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

// RenderTicketDraftsCSV renders a CSV payload (including header) in-memory.
func RenderTicketDraftsCSV(exp TicketDraftExport) (string, error) {
	var b strings.Builder
	w := csv.NewWriter(&b)

	header := []string{
		"title",
		"description",
		"labels",
		"owner",
		"priority",
		"source_action_item_id",
		"source_meeting_id",
		"source_meeting_name",
		"source_segment_ids",
	}
	if err := w.Write(header); err != nil {
		return "", fmt.Errorf("write csv header: %w", err)
	}

	for _, d := range exp.Drafts {
		owner := ""
		if d.Owner != nil {
			owner = *d.Owner
		}
		priority := ""
		if d.Priority != nil {
			priority = *d.Priority
		}
		meetingID := ""
		if d.SourceMeetingID != nil {
			meetingID = *d.SourceMeetingID
		}
		meetingName := ""
		if d.SourceMeetingName != nil {
			meetingName = *d.SourceMeetingName
		}

		rec := []string{
			d.Title,
			d.Description,
			strings.Join(d.Labels, ";"),
			owner,
			priority,
			d.SourceActionItemID,
			meetingID,
			meetingName,
			strings.Join(d.SourceSegmentIDs, ";"),
		}
		if err := w.Write(rec); err != nil {
			return "", fmt.Errorf("write csv record: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return "", fmt.Errorf("flush csv: %w", err)
	}
	return b.String(), nil
}

func writeTicketDraftsCSV(path string, exp TicketDraftExport) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)

	header := []string{
		"title",
		"description",
		"labels",
		"owner",
		"priority",
		"source_action_item_id",
		"source_meeting_id",
		"source_meeting_name",
		"source_segment_ids",
	}
	if err := w.Write(header); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	for _, d := range exp.Drafts {
		owner := ""
		if d.Owner != nil {
			owner = *d.Owner
		}
		priority := ""
		if d.Priority != nil {
			priority = *d.Priority
		}
		meetingID := ""
		if d.SourceMeetingID != nil {
			meetingID = *d.SourceMeetingID
		}
		meetingName := ""
		if d.SourceMeetingName != nil {
			meetingName = *d.SourceMeetingName
		}

		rec := []string{
			d.Title,
			d.Description,
			strings.Join(d.Labels, ";"),
			owner,
			priority,
			d.SourceActionItemID,
			meetingID,
			meetingName,
			strings.Join(d.SourceSegmentIDs, ";"),
		}
		if err := w.Write(rec); err != nil {
			return fmt.Errorf("write csv record: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("flush csv: %w", err)
	}
	return nil
}
