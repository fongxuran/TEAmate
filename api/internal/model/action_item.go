package model

import "time"

// ActionItem is a structured follow-up that can be converted into a ticket draft.
//
// This schema is defined by docs/ticket/T-008.
//
// action_item_id should be stable within a single extraction run so downstream
// artifacts (like ticket drafts) can reference it.
type ActionItem struct {
	ActionItemID     string     `json:"action_item_id"`
	Title            string     `json:"title"`
	Description      *string    `json:"description,omitempty"`
	Owner            *string    `json:"owner,omitempty"`
	DueDate          *time.Time `json:"due_date,omitempty"`
	SourceSegmentIDs []string   `json:"source_segment_ids,omitempty"`
	Confidence       float64    `json:"confidence"`
}
