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
//
// Summary: Audit store implementation that appends newline-delimited JSON (NDJSON) to a file or standard output.
type FileAuditStore struct {
	mu   sync.Mutex
	file *os.File
	out  io.Writer
}

// NewFileAuditStore provides newfileauditstore functionality.
//
// Summary: NewFileAuditStore.
//
// Parameters.
//   - path: The parameter.
//   - error: The parameter.
//
// Returns.
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

// Write provides write functionality.
//
// Summary: Write.
//
// Parameters.
//   - _: The parameter.
//   - entry: The parameter.
//
// Returns.
//   - result: The result.
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

// Read provides read functionality.
//
// Summary: Read.
//
// Parameters.
//   - _: The parameter.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *FileAuditStore) Read(_ context.Context, _ Filter) ([]Entry, error) {
	return nil, fmt.Errorf("read not implemented for file audit store")
}

// Close provides close functionality.
//
// Summary: Close.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *FileAuditStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}
