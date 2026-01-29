package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/mcpany/core/server/pkg/validation"
)

// FileAuditStore writes audit logs to a file or stdout.
type FileAuditStore struct {
	mu   sync.Mutex
	file *os.File
}

// NewFileAuditStore creates a new FileAuditStore.
//
// path is the path.
//
// Returns the result.
// Returns an error if the operation fails.
func NewFileAuditStore(path string) (*FileAuditStore, error) {
	var f *os.File
	var err error
	if path != "" {
		if err := validation.IsAllowedPath(path); err != nil {
			return nil, fmt.Errorf("audit log file path not allowed: %w", err)
		}
		f, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) //nolint:gosec
		if err != nil {
			return nil, fmt.Errorf("failed to open audit log file: %w", err)
		}
	}
	return &FileAuditStore{
		file: f,
	}, nil
}

// Write writes an audit entry to the file.
//
// _ is an unused parameter.
// entry is the entry.
//
// Returns an error if the operation fails.
func (s *FileAuditStore) Write(_ context.Context, entry Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var w *os.File
	if s.file != nil {
		w = s.file
	} else {
		w = os.Stdout
	}

	return json.NewEncoder(w).Encode(entry)
}

// Read implements the Store interface.
func (s *FileAuditStore) Read(_ context.Context, _ Filter) ([]Entry, error) {
	return nil, fmt.Errorf("read not implemented for file audit store")
}

// Close closes the file.
//
// Returns an error if the operation fails.
func (s *FileAuditStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}
