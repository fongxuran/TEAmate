package model

import "time"

// Message represents a stored message.
type Message struct {
	ID        int64
	Body      string
	CreatedAt time.Time
}
