package messages

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("new sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	createdAt := time.Now().UTC()
	rows := sqlmock.NewRows([]string{"id", "body", "binary", "file_name", "content_type", "size_bytes", "created_at"}).
		AddRow(int64(1), sql.NullString{String: "hello", Valid: true}, []byte("bin"), sql.NullString{String: "file.txt", Valid: true}, sql.NullString{String: "text/plain", Valid: true}, sql.NullInt64{Int64: 3, Valid: true}, createdAt)
	mock.ExpectQuery("INSERT INTO messages").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(rows)

	msg, err := repo.Create(context.Background(), "hello", []byte("bin"), "file.txt", "text/plain", 3)
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	if msg.ID != 1 {
		t.Fatalf("expected id 1, got %d", msg.ID)
	}
	if msg.Body != "hello" {
		t.Fatalf("expected body hello, got %s", msg.Body)
	}
	if string(msg.Binary) != "bin" {
		t.Fatalf("expected binary bin, got %s", string(msg.Binary))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRepositoryList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("new sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	createdAt := time.Now().UTC()
	rows := sqlmock.NewRows([]string{"id", "body", "file_name", "content_type", "size_bytes", "created_at"}).
		AddRow(int64(1), sql.NullString{String: "first", Valid: true}, sql.NullString{String: "", Valid: false}, sql.NullString{String: "", Valid: false}, sql.NullInt64{Int64: 0, Valid: false}, createdAt)
	mock.ExpectQuery("SELECT id, body, file_name, content_type, size_bytes, created_at FROM messages").WithArgs(2, 1).WillReturnRows(rows)

	messages, err := repo.List(context.Background(), 2, 1, false)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	if messages[0].Body != "first" {
		t.Fatalf("expected body first, got %s", messages[0].Body)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
