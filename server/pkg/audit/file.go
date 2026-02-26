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
// Summary: Creates a new FileAuditStore.
//
// Parameters:
//   - path (string): The path to the log file. If empty, logs to stdout.
//
// Returns:
//   - *FileAuditStore: The initialized store.
//   - error: An error if the file cannot be opened.
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
// Summary: Writes an audit log entry to the file or stdout.
//
// Parameters:
//   - _ (context.Context): The context (unused).
//   - entry (Entry): The audit entry.
//
// Returns:
//   - error: An error if the write operation fails.
//
// Side Effects:
//   - Writes JSON data to the file/stdout.
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
// Summary: Reads audit logs (Not implemented).
//
// Parameters:
//   - _ (context.Context): The context.
//   - _ (Filter): The filter criteria.
//
// Returns:
//   - []Entry: The list of entries (nil).
//   - error: An error indicating this method is not implemented.
//
// Side Effects:
//   - None.
func (s *FileAuditStore) Read(_ context.Context, _ Filter) ([]Entry, error) {
	return nil, fmt.Errorf("read not implemented for file audit store")
}

// Close closes the file.
//
// Summary: Closes the underlying file.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if closing fails.
//
// Side Effects:
//   - Closes the file descriptor.
func (s *FileAuditStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}
