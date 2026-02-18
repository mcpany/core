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
// Summary: Initializes a file-based audit store.
//
// Parameters:
//   - path: string. The file path (or empty for stdout).
//
// Returns:
//   - *FileAuditStore: The initialized store.
//   - error: Error if file opening fails.
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
// Summary: Writes an audit entry to the file/stdout.
//
// Parameters:
//   - ctx: context.Context. The context.
//   - entry: Entry. The audit entry.
//
// Returns:
//   - error: Error if writing fails.
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
// Summary: Reads audit entries (not implemented for file store yet).
//
// Parameters:
//   - ctx: context.Context. The context.
//   - filter: Filter. The filter.
//
// Returns:
//   - []Entry: Nil.
//   - error: Error indicating not implemented.
func (s *FileAuditStore) Read(_ context.Context, _ Filter) ([]Entry, error) {
	return nil, fmt.Errorf("read not implemented for file audit store")
}

// Close closes the file.
//
// Summary: Closes the file handle.
//
// Returns:
//   - error: Error if closing fails.
func (s *FileAuditStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}
