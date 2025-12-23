// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// FileAuditStore writes audit logs to a file or stdout.
type FileAuditStore struct {
	mu       sync.Mutex
	file     *os.File
	filePath string
	lastHash string
}

// NewFileAuditStore creates a new FileAuditStore.
func NewFileAuditStore(path string) (*FileAuditStore, error) {
	lastHash := ""
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			// Read file to get last hash
			f, err := os.Open(path)
			if err != nil {
				return nil, fmt.Errorf("failed to open audit log for reading: %w", err)
			}
			r := bufio.NewReader(f)
			for {
				line, err := r.ReadBytes('\n')
				if len(line) > 0 {
					var entry AuditEntry
					if err := json.Unmarshal(line, &entry); err == nil {
						lastHash = entry.Hash
					}
				}
				if err != nil {
					break
				}
			}
			_ = f.Close()
		}
	}

	var f *os.File
	var err error
	if path != "" {
		f, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) //nolint:gosec
		if err != nil {
			return nil, fmt.Errorf("failed to open audit log file: %w", err)
		}
	}
	return &FileAuditStore{
		file:     f,
		filePath: path,
		lastHash: lastHash,
	}, nil
}

// Write writes an audit entry to the file.
func (s *FileAuditStore) Write(entry AuditEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry.PreviousHash = s.lastHash
	entry.Hash = ComputeAuditHash(entry, s.lastHash)
	s.lastHash = entry.Hash

	var w *os.File
	if s.file != nil {
		w = s.file
	} else {
		w = os.Stdout
	}

	return json.NewEncoder(w).Encode(entry)
}

// Verify checks the integrity of the audit logs.
func (s *FileAuditStore) Verify() (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.filePath == "" {
		// Can't verify stdout
		return true, nil
	}

	f, err := os.Open(s.filePath)
	if err != nil {
		return false, fmt.Errorf("failed to open file for verification: %w", err)
	}
	defer func() { _ = f.Close() }()

	r := bufio.NewReader(f)
	var prevHash string
	lineNum := 0

	for {
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			lineNum++
			var entry AuditEntry
			if err := json.Unmarshal(line, &entry); err != nil {
				return false, fmt.Errorf("line %d: invalid json: %w", lineNum, err)
			}

			if entry.PreviousHash != prevHash {
				return false, fmt.Errorf("line %d: prev_hash mismatch (expected %q, got %q)", lineNum, prevHash, entry.PreviousHash)
			}

			calculatedHash := ComputeAuditHash(entry, prevHash)
			if calculatedHash != entry.Hash {
				return false, fmt.Errorf("line %d: hash mismatch (calculated %q, got %q)", lineNum, calculatedHash, entry.Hash)
			}

			prevHash = entry.Hash
		}
		if err != nil {
			if err != os.ErrClosed && err.Error() != "EOF" {
				// return false, fmt.Errorf("read error: %w", err)
			}
			break
		}
	}

	return true, nil
}

// Close closes the file.
func (s *FileAuditStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}
