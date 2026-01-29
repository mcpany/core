package audit

import (
	"context"
	"encoding/json"
	"time"
)

// Entry represents a single audit log entry.
type Entry struct {
	Timestamp  time.Time       `json:"timestamp"`
	ToolName   string          `json:"tool_name"`
	UserID     string          `json:"user_id,omitempty"`
	ProfileID  string          `json:"profile_id,omitempty"`
	Arguments  json.RawMessage `json:"arguments,omitempty"`
	Result     any             `json:"result,omitempty"`
	Error      string          `json:"error,omitempty"`
	Duration   string          `json:"duration"`
	DurationMs int64           `json:"duration_ms"`
}

// Filter defines the filters for reading audit logs.
type Filter struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	ToolName  string     `json:"tool_name,omitempty"`
	UserID    string     `json:"user_id,omitempty"`
	ProfileID string     `json:"profile_id,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// Store defines the interface for audit log storage.
type Store interface {
	// Write writes an audit entry to the store.
	//
	// ctx is the context for the request.
	// entry is the entry.
	//
	// Returns an error if the operation fails.
	Write(ctx context.Context, entry Entry) error
	// Read reads audit entries from the store based on the filter.
	//
	// ctx is the context for the request.
	// filter is the filter to apply.
	//
	// Returns the entries and an error if the operation fails.
	Read(ctx context.Context, filter Filter) ([]Entry, error)
	// Close closes the store.
	//
	// Returns an error if the operation fails.
	Close() error
}
