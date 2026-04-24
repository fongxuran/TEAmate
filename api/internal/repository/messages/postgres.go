package messages

import (
	"context"
	"database/sql"
	"fmt"

	pkgerrors "github.com/pkg/errors"

	"teammate/internal/model"
)

type postgresRepository struct {
	db *sql.DB
}

func (r postgresRepository) Create(ctx context.Context, body string) (model.Message, error) {
	const query = `
		INSERT INTO messages (body)
		VALUES ($1)
		RETURNING id, body, created_at
	`

	var msg model.Message
	if err := r.db.QueryRowContext(ctx, query, body).Scan(&msg.ID, &msg.Body, &msg.CreatedAt); err != nil {
		return model.Message{}, pkgerrors.WithStack(fmt.Errorf("insert message: %w", err))
	}

	return msg, nil
}

func (r postgresRepository) List(ctx context.Context, limit, offset int) ([]model.Message, error) {
	const query = `
		SELECT id, body, created_at
		FROM messages
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, pkgerrors.WithStack(fmt.Errorf("list messages: %w", err))
	}
	defer rows.Close()

	messages := make([]model.Message, 0)
	for rows.Next() {
		var msg model.Message
		if err := rows.Scan(&msg.ID, &msg.Body, &msg.CreatedAt); err != nil {
			return nil, pkgerrors.WithStack(fmt.Errorf("scan message: %w", err))
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.WithStack(fmt.Errorf("read messages: %w", err))
	}

	return messages, nil
}
