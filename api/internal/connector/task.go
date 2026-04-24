package connector

import (
	"context"

	"teammate/internal/model"
)

// TaskRef is a generic reference to a created task/page in an external system.
//
// For MVP, this is used by the Notion connector (T-010).
type TaskRef struct {
	ID     string `json:"id,omitempty"`
	URL    string `json:"url,omitempty"`
	DryRun bool   `json:"dry_run"`
}

// TaskCreator is an abstraction over external task systems (e.g. Notion).
//
// It intentionally accepts a TicketDraft so draft-to-task mapping is centralized
// in the connector implementation.
type TaskCreator interface {
	CreateTask(ctx context.Context, draft model.TicketDraft) (TaskRef, error)
}