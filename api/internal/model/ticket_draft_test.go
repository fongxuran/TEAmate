package model

import "testing"

func TestTicketDraftFromActionItem_TraceabilityAndCopying(t *testing.T) {
	meetingID := "m-123"
	meetingName := "Weekly"
	meeting := MeetingInput{
		SchemaVersion: "v1",
		Transcript: Transcript{
			MeetingID:   &meetingID,
			MeetingName: &meetingName,
		},
	}

	owner := "Alex"
	desc := "Do the thing"
	ai := ActionItem{
		ActionItemID:     "ai-1",
		Title:            "Implement exporter",
		Description:      &desc,
		Owner:            &owner,
		SourceSegmentIDs: []string{"seg-1-2", "seg-3-4"},
		Confidence:       0.9,
	}

	d := TicketDraftFromActionItem(meeting, ai)
	if d.Title != ai.Title {
		t.Fatalf("title: got %q want %q", d.Title, ai.Title)
	}
	if d.Description != desc {
		t.Fatalf("description: got %q want %q", d.Description, desc)
	}
	if d.Owner == nil || *d.Owner != owner {
		t.Fatalf("owner: got %#v want %q", d.Owner, owner)
	}
	if d.SourceActionItemID != ai.ActionItemID {
		t.Fatalf("source_action_item_id: got %q want %q", d.SourceActionItemID, ai.ActionItemID)
	}
	if d.SourceMeetingID == nil || *d.SourceMeetingID != meetingID {
		t.Fatalf("source_meeting_id: got %#v want %q", d.SourceMeetingID, meetingID)
	}
	if d.SourceMeetingName == nil || *d.SourceMeetingName != meetingName {
		t.Fatalf("source_meeting_name: got %#v want %q", d.SourceMeetingName, meetingName)
	}
	if len(d.SourceSegmentIDs) != len(ai.SourceSegmentIDs) {
		t.Fatalf("source_segment_ids len: got %d want %d", len(d.SourceSegmentIDs), len(ai.SourceSegmentIDs))
	}
	// Ensure it is a copy (not the same underlying slice).
	if len(d.SourceSegmentIDs) > 0 {
		orig := ai.SourceSegmentIDs[0]
		d.SourceSegmentIDs[0] = "mutated"
		if ai.SourceSegmentIDs[0] != orig {
			t.Fatalf("expected source segment ids to be copied, but action item mutated")
		}
	}
}
