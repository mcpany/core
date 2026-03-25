// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/mcpany/core/server/pkg/validation"
)

// FileAuditStore writes audit logs to a file or stdout.
// Summary: Audit store implementation that appends newline-delimited JSON (NDJSON) to a file or standard output.
type FileAuditStore struct {
	mu   sync.Mutex
	file *os.File
	out  io.Writer
}

// NewFileAuditStore creates a new FileAuditStore.
// Summary: Initializes a new FileAuditStore at the specified path, or uses stdout if path is empty.
// Parameters:
//   - path (string): The file path for the audit log.
//
// Returns:
//   - *FileAuditStore: The initialized audit store.
//   - error: An error if the path is not allowed or the file cannot be opened.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewFileAuditStore(path string) (*FileAuditStore, error) {
	var f *os.File
	var err error
	if path != "" {
		if err := validation.IsAllowedPath(path); err != nil {
			return nil, fmt.Errorf("audit log file path not allowed: %w", err)
		}
		f, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open audit log file: %w", err)
		}
	}
	return &FileAuditStore{
		file: f,
		out:  os.Stdout,
	}, nil
}

// Write writes an audit entry to the file.
// Summary: Appends a JSON-marshaled audit entry to the configured output file or stdout.
// Parameters:
//   - ctx (context.Context): The context for the request (unused).
//   - entry (Entry): The audit entry to be persisted.
//
// Returns:
//   - error: An error if marshaling fails or the write operation encounters an I/O error.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *FileAuditStore) Write(_ context.Context, entry Entry) error {
	// ⚡ BOLT: Serialize JSON outside the lock to reduce critical section duration.
	// Randomized Selection from Top 5 High-Impact Targets
	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	// json.NewEncoder.Encode appends a newline, so we must add it here too.
	b = append(b, '\n')

	s.mu.Lock()
	defer s.mu.Unlock()

	var w io.Writer
	if s.file != nil {
		w = s.file
	} else {
		w = s.out
	}

	_, err = w.Write(b)
	return err
}

// Read implements the Store interface.
// Summary: Not implemented for FileAuditStore.
// Parameters:
//   - ctx (context.Context): The context for the request (unused).
//   - filter (Filter): The filter criteria (unused).
//
// Returns:
//   - []Entry: Always returns nil.
//   - error: Always returns a "not implemented" error.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *FileAuditStore) Read(_ context.Context, _ Filter) ([]Entry, error) {
	return nil, fmt.Errorf("read not implemented for file audit store")
}

// Close closes the file.
// Summary: Closes the underlying file handle if it was opened during initialization.
// Returns:
//   - error: An error if the file handle fails to close.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *FileAuditStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}
