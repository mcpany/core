package logs

import (
	"database/sql"
	"encoding/json"

	"github.com/mcpany/core/server/pkg/logging"
)

// LogStore handles persistence of logs to SQLite.
type LogStore struct {
	db *sql.DB
}

// NewStore creates a new LogStore.
func NewStore(db *sql.DB) *LogStore {
	return &LogStore{db: db}
}

// Init creates the logs table if it doesn't exist.
func (s *LogStore) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS logs (
		id TEXT PRIMARY KEY,
		timestamp DATETIME,
		level TEXT,
		source TEXT,
		message TEXT,
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);
	`
	_, err := s.db.Exec(query)
	return err
}

// InsertBytes inserts a log entry from JSON bytes.
func (s *LogStore) InsertBytes(data []byte) error {
	var entry logging.LogEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return err
	}
	return s.Insert(entry)
}

// Insert inserts a log entry.
func (s *LogStore) Insert(entry logging.LogEntry) error {
	metadata, _ := json.Marshal(entry.Metadata)

	// Keep only last 100k logs to prevent DB explosion?
	// Or implement cleanup job?
	// For now, simple insert.

	_, err := s.db.Exec("INSERT INTO logs (id, timestamp, level, source, message, metadata) VALUES (?, ?, ?, ?, ?, ?)",
		entry.ID, entry.Timestamp, entry.Level, entry.Source, entry.Message, string(metadata))
	return err
}

// Query fetches logs with pagination and filtering.
func (s *LogStore) Query(limit int, offset int) ([]logging.LogEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}

	rows, err := s.db.Query("SELECT id, timestamp, level, source, message, metadata FROM logs ORDER BY timestamp DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []logging.LogEntry
	for rows.Next() {
		var l logging.LogEntry
		var metaStr string
		var tsStr string
		if err := rows.Scan(&l.ID, &tsStr, &l.Level, &l.Source, &l.Message, &metaStr); err != nil {
			return nil, err
		}

		// Ensure timestamp matches format
		l.Timestamp = tsStr
		// Try to normalize if it was stored as standard SQL datetime which might differ from RFC3339
		// But we store what we get (RFC3339 string usually) into TEXT/DATETIME column.
		// SQLite DATETIME is flexible.

		if metaStr != "" {
			_ = json.Unmarshal([]byte(metaStr), &l.Metadata)
		}
		logs = append(logs, l)
	}

	// Reverse to return chronological order [oldest ... latest]
	// This makes it easier for UI to prepend/append.
	// But usually "Get History" implies "Latest".
	// If UI calls getLogs(), it likely wants to show them.
	// If I return [Newest...Oldest], UI has to reverse.
	// If I return [Oldest...Newest], UI just appends?
	// But we queried DESC. So we have Newest first.
	// Let's reverse it so it's chronological.
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	return logs, nil
}

// Cleanup removes old logs (e.g. keep last 100k).
func (s *LogStore) Cleanup() error {
	// Simple retention policy: Keep last 7 days or 100k records
	// For MVP, just keep last 100k
	_, err := s.db.Exec(`DELETE FROM logs WHERE id NOT IN (SELECT id FROM logs ORDER BY timestamp DESC LIMIT 100000)`)
	return err
}
