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
// FileAuditStore writes audit logs to a file or stdout.
//
// Summary: Audit store implementation that appends newline-delimited JSON (NDJSON) to a file or standard output.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
// Errors:
//   - triggers relevant error states on failure.
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
// Read implements the Store interface.
//
// Summary: Reads audit entries (Not implemented).
//
// Parameters:
// Close closes the file.
//
// Summary: Closes the underlying file handle if one exists.
//
// Returns:
//   - error: An error if closing the file fails.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Side Effects:
//   - Closes the file descriptor.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
func (s *FileAuditStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}
