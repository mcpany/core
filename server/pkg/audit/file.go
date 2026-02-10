// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"bufio"
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
	path string
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
		f, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open audit log file: %w", err)
		}
	}
	return &FileAuditStore{
		file: f,
		out:  os.Stdout,
		path: path,
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

	var w io.Writer
	if s.file != nil {
		w = s.file
	} else {
		w = s.out
	}

	return json.NewEncoder(w).Encode(entry)
}

// Read implements the Store interface.
func (s *FileAuditStore) Read(_ context.Context, filter Filter) ([]Entry, error) {
	if s.path == "" {
		// If writing to stdout, we can't read back history easily.
		// Return empty list or error? Empty list is safer for UI.
		return []Entry{}, nil
	}

	// Open file for reading
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, fmt.Errorf("failed to open audit log for reading: %w", err)
	}
	defer func() { _ = f.Close() }()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	// Increase buffer for large log lines (default is 64k)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 1MB max line

	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue // Skip malformed lines
		}

		// Apply filters
		if filter.ToolName != "" && entry.ToolName != filter.ToolName {
			continue
		}
		if filter.UserID != "" && entry.UserID != filter.UserID {
			continue
		}
		if filter.ProfileID != "" && entry.ProfileID != filter.ProfileID {
			continue
		}
		if filter.StartTime != nil && entry.Timestamp.Before(*filter.StartTime) {
			continue
		}
		if filter.EndTime != nil && entry.Timestamp.After(*filter.EndTime) {
			continue
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan audit log: %w", err)
	}

	// Reverse (Newest First)
	for i := len(entries)/2 - 1; i >= 0; i-- {
		opp := len(entries) - 1 - i
		entries[i], entries[opp] = entries[opp], entries[i]
	}

	// Pagination
	start := filter.Offset
	if start >= len(entries) {
		return []Entry{}, nil
	}
	end := len(entries)
	if filter.Limit > 0 {
		end = start + filter.Limit
		if end > len(entries) {
			end = len(entries)
		}
	}

	return entries[start:end], nil
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
