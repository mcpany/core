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
type FileAuditStore struct {
	mu   sync.Mutex
	file *os.File
	out  io.Writer
}

// NewFileAuditStore creates a new FileAuditStore.
//
// Summary: Creates a new FileAuditStore instance.
//
// Parameters:
//   - path (string): The file path to write audit logs to. If empty, writes to stdout.
//
// Returns:
//   - *FileAuditStore: The created store instance.
//   - error: An error if the path is invalid or file opening fails.
//
// Errors:
//   - Returns error if path is not allowed.
//   - Returns error if file cannot be opened.
//
// Side Effects:
//   - Opens the specified file for appending.
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
//
// Summary: Writes an audit entry to the file or stdout.
//
// Parameters:
//   - _ (context.Context): Unused context.
//   - entry (Entry): The audit entry to write.
//
// Returns:
//   - error: An error if writing fails.
//
// Errors:
//   - Returns error if JSON marshaling fails.
//   - Returns error if file write fails.
//
// Side Effects:
//   - Writes JSON-encoded entry to the output.
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
//
// Summary: Reads audit entries (Not Implemented).
//
// Parameters:
//   - _ (context.Context): Unused context.
//   - _ (Filter): Unused filter.
//
// Returns:
//   - []Entry: Nil.
//   - error: Error indicating not implemented.
//
// Errors:
//   - Returns "read not implemented for file audit store".
func (s *FileAuditStore) Read(_ context.Context, _ Filter) ([]Entry, error) {
	return nil, fmt.Errorf("read not implemented for file audit store")
}

// Close closes the file.
//
// Summary: Closes the underlying file.
//
// Returns:
//   - error: An error if closing fails.
//
// Side Effects:
//   - Closes the file descriptor if one exists.
func (s *FileAuditStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}
