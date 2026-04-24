package model

// TicketDraft is a locally-exportable representation of work derived from a meeting.
//
// This schema is defined by docs/ticket/T-009.
//
// Note: "labels" is intentionally shaped to map cleanly into downstream systems
// (e.g., Notion pages in T-010).
type TicketDraft struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Labels      []string `json:"labels,omitempty"`
	Owner       *string  `json:"owner,omitempty"`
	Priority    *string  `json:"priority,omitempty"`

	// Traceability back to extraction and meeting context.
	SourceActionItemID string   `json:"source_action_item_id"`
	SourceMeetingID    *string  `json:"source_meeting_id,omitempty"`
	SourceMeetingName  *string  `json:"source_meeting_name,omitempty"`
	SourceSegmentIDs   []string `json:"source_segment_ids,omitempty"`
}

// TicketDraftFromActionItem converts an ActionItem into a TicketDraft while preserving
// traceability back to the meeting and source segments.
func TicketDraftFromActionItem(meeting MeetingInput, ai ActionItem) TicketDraft {
	d := ""
	if ai.Description != nil {
		d = *ai.Description
	}

	return TicketDraft{
		Title:              ai.Title,
		Description:        d,
		Owner:              ai.Owner,
		SourceActionItemID: ai.ActionItemID,
		SourceMeetingID:    meeting.Transcript.MeetingID,
		SourceMeetingName:  meeting.Transcript.MeetingName,
		SourceSegmentIDs:   append([]string(nil), ai.SourceSegmentIDs...),
	}
}

// TicketDraftsFromActionItems converts a slice of ActionItems into TicketDrafts.
func TicketDraftsFromActionItems(meeting MeetingInput, items []ActionItem) []TicketDraft {
	if len(items) == 0 {
		return nil
	}
	out := make([]TicketDraft, 0, len(items))
	for _, ai := range items {
		out = append(out, TicketDraftFromActionItem(meeting, ai))
	}
	return out
}
