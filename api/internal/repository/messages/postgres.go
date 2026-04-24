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

func (r postgresRepository) Create(ctx context.Context, body string, binary []byte, fileName, contentType string, sizeBytes int64) (model.Message, error) {
	const query = `
		INSERT INTO messages (body, binary, file_name, content_type, size_bytes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, body, binary, file_name, content_type, size_bytes, created_at
	`

	bodyValue := sql.NullString{String: body, Valid: body != ""}
	fileNameValue := sql.NullString{String: fileName, Valid: fileName != ""}
	contentTypeValue := sql.NullString{String: contentType, Valid: contentType != ""}
	sizeValue := sql.NullInt64{Int64: sizeBytes, Valid: sizeBytes > 0}

	var msg model.Message
	var dbBody sql.NullString
	var dbFileName sql.NullString
	var dbContentType sql.NullString
	var dbSize sql.NullInt64
	if err := r.db.QueryRowContext(ctx, query, bodyValue, binary, fileNameValue, contentTypeValue, sizeValue).
		Scan(&msg.ID, &dbBody, &msg.Binary, &dbFileName, &dbContentType, &dbSize, &msg.CreatedAt); err != nil {
		return model.Message{}, pkgerrors.WithStack(fmt.Errorf("insert message: %w", err))
	}

	if dbBody.Valid {
		msg.Body = dbBody.String
	}
	if dbFileName.Valid {
		msg.FileName = dbFileName.String
	}
	if dbContentType.Valid {
		msg.ContentType = dbContentType.String
	}
	if dbSize.Valid {
		msg.SizeBytes = dbSize.Int64
	}

	return msg, nil
}

func (r postgresRepository) List(ctx context.Context, limit, offset int, includeBinary bool) ([]model.Message, error) {
	query := `
		SELECT id, body, file_name, content_type, size_bytes, created_at
		FROM messages
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	if includeBinary {
		query = `
			SELECT id, body, binary, file_name, content_type, size_bytes, created_at
			FROM messages
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
	}

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, pkgerrors.WithStack(fmt.Errorf("list messages: %w", err))
	}
	defer rows.Close()

	messages := make([]model.Message, 0)
	for rows.Next() {
		var msg model.Message
		var dbBody sql.NullString
		var dbFileName sql.NullString
		var dbContentType sql.NullString
		var dbSize sql.NullInt64
		if includeBinary {
			if err := rows.Scan(&msg.ID, &dbBody, &msg.Binary, &dbFileName, &dbContentType, &dbSize, &msg.CreatedAt); err != nil {
				return nil, pkgerrors.WithStack(fmt.Errorf("scan message: %w", err))
			}
		} else {
			if err := rows.Scan(&msg.ID, &dbBody, &dbFileName, &dbContentType, &dbSize, &msg.CreatedAt); err != nil {
				return nil, pkgerrors.WithStack(fmt.Errorf("scan message: %w", err))
			}
		}

		if dbBody.Valid {
			msg.Body = dbBody.String
		}
		if dbFileName.Valid {
			msg.FileName = dbFileName.String
		}
		if dbContentType.Valid {
			msg.ContentType = dbContentType.String
		}
		if dbSize.Valid {
			msg.SizeBytes = dbSize.Int64
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.WithStack(fmt.Errorf("read messages: %w", err))
	}

	return messages, nil
}
