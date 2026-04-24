package messages

import (
	"context"
	"database/sql"

	"teammate/internal/model"
)

// Repository provides access to message storage.
type Repository interface {
	Create(ctx context.Context, body string, binary []byte, fileName, contentType string, sizeBytes int64) (model.Message, error)
	List(ctx context.Context, limit, offset int, includeBinary bool) ([]model.Message, error)
}

// NewRepository returns a postgres-backed repository.
func NewRepository(db *sql.DB) Repository {
	return postgresRepository{db: db}
}
