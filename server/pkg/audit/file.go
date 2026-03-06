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

// FileAuditStore writes audit logs to a file or stdout. Summary: Audit store implementation that appends newline-delimited JSON (NDJSON) to a file or standard output.
//
// Summary: FileAuditStore writes audit logs to a file or stdout. Summary: Audit store implementation that appends newline-delimited JSON (NDJSON) to a file or standard output.
//
// Fields:
//   - Contains the configuration and state properties required for FileAuditStore functionality.
type FileAuditStore struct {
	mu   sync.Mutex
	file *os.File
	out  io.Writer
}

// NewFileAuditStore creates a new FileAuditStore.
//
// Summary: Initializes a new FileAuditStore.
//
// Parameters:
//   - path: string. The file path for the audit log (or empty for stdout).
//
// Returns:
//   - *FileAuditStore: The initialized store.
//   - error: An error if the path is invalid or file cannot be opened.
//
// Errors:
//   - Returns error if path validation fails.
//   - Returns error if file creation/opening fails.
//
// Side Effects:
//   - Opens (or creates) the specified file in append mode.
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
// Summary: Appends a JSON-marshaled audit entry to the configured output.
//
// Parameters:
//   - _: context.Context. Unused.
//   - entry: Entry. The audit entry to write.
//
// Returns:
//   - error: An error if writing fails.
//
// Side Effects:
//   - Writes data to the file or stdout.
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

// Read implements the Store interface. Summary: Reads audit entries (Not implemented). Parameters: - _: context.Context. Unused. - _: Filter. Unused. Returns: - []Entry: Nil. - error: Always returns "not implemented".
//
// Summary: Read implements the Store interface. Summary: Reads audit entries (Not implemented). Parameters: - _: context.Context. Unused. - _: Filter. Unused. Returns: - []Entry: Nil. - error: Always returns "not implemented".
//
// Parameters:
//   - _ (context.Context): The _ parameter used in the operation.
//   - _ (Filter): The _ parameter used in the operation.
//
// Returns:
//   - ([]Entry): The resulting []Entry object containing the requested data.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
func (s *FileAuditStore) Read(_ context.Context, _ Filter) ([]Entry, error) {
	return nil, fmt.Errorf("read not implemented for file audit store")
}

// Close closes the file. Summary: Closes the underlying file handle if one exists. Returns: - error: An error if closing the file fails. Side Effects: - Closes the file descriptor.
//
// Summary: Close closes the file. Summary: Closes the underlying file handle if one exists. Returns: - error: An error if closing the file fails. Side Effects: - Closes the file descriptor.
//
// Parameters:
//   - None.
//
// Returns:
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
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
