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

// FileAuditStore represents the public FileAuditStore entity.
//
// Summary: Defines the structured data model representing a audit store.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type FileAuditStore struct {
	mu   sync.Mutex
	file *os.File
	out  io.Writer
}

// NewFileAuditStore serves as a public interface for interacting with NewFileAuditStore.
//
// Summary: Constructs and returns an initialized file audit store ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Write serves as a public interface for interacting with Write.
//
// Summary: Write the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Read serves as a public interface for interacting with Read.
//
// Summary: Read the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *FileAuditStore) Read(_ context.Context, _ Filter) ([]Entry, error) {
	return nil, fmt.Errorf("read not implemented for file audit store")
}

// Close serves as a public interface for interacting with Close.
//
// Summary: Close the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *FileAuditStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}
