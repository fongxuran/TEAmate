package model

import "time"

// Message represents a stored message.
type Message struct {
	ID          int64
	Body        string
	Binary      []byte
	FileName    string
	ContentType string
	SizeBytes   int64
	CreatedAt   time.Time
}
