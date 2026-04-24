package messages

import (
	"context"
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
	rows := sqlmock.NewRows([]string{"id", "body", "created_at"}).AddRow(int64(1), "hello", createdAt)
	mock.ExpectQuery("INSERT INTO messages").WithArgs("hello").WillReturnRows(rows)

	msg, err := repo.Create(context.Background(), "hello")
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	if msg.ID != 1 {
		t.Fatalf("expected id 1, got %d", msg.ID)
	}
	if msg.Body != "hello" {
		t.Fatalf("expected body hello, got %s", msg.Body)
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
	rows := sqlmock.NewRows([]string{"id", "body", "created_at"}).AddRow(int64(1), "first", createdAt)
	mock.ExpectQuery("SELECT id, body, created_at FROM messages").WithArgs(2, 1).WillReturnRows(rows)

	messages, err := repo.List(context.Background(), 2, 1)
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
