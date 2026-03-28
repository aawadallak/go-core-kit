// Package outbox provides outbox functionality.
package outbox

import "time"

type EntryStatus string

const (
	EntryStatusPending   EntryStatus = "pending"
	EntryStatusSent      EntryStatus = "sent"
	EntryStatusFailed    EntryStatus = "failed"
	EntryStatusExhausted EntryStatus = "exhausted"
)

type Entry struct {
	ID            uint        `gorm:"primaryKey"`
	EventID       string      `gorm:"column:event_id"`
	EventName     string      `gorm:"column:event_name;not null"`
	EventVersion  int         `gorm:"column:event_version;not null;default:1"`
	Payload       []byte      `gorm:"column:payload;type:jsonb;not null"`
	CorrelationID string      `gorm:"column:correlation_id"`
	RequestID     string      `gorm:"column:request_id"`
	TraceID       string      `gorm:"column:trace_id"`
	SpanID        string      `gorm:"column:span_id"`
	Status        EntryStatus `gorm:"column:status;not null;default:pending"`
	RetryCount    int         `gorm:"column:retry_count;not null;default:0"`
	LastError     string      `gorm:"column:last_error"`
	CreatedAt     time.Time   `gorm:"column:created_at;not null"`
	ProcessedAt   *time.Time  `gorm:"column:processed_at"`
}

func (Entry) TableName() string {
	return "tbl_event_outbox"
}
