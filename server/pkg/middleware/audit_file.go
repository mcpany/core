// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// FileAuditStore writes audit logs to a file or stdout.
type FileAuditStore struct {
	mu           sync.Mutex
	file         *os.File
	lastHash     string
	path         string // Store path for reading back
	reopenNeeded bool   // For log rotation support (not fully implemented but good practice)
}

// NewFileAuditStore creates a new FileAuditStore.
func NewFileAuditStore(path string) (*FileAuditStore, error) {
	var f *os.File
	var err error
	var lastHash string

	if path != "" {
		f, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644) //nolint:gosec
		if err != nil {
			return nil, fmt.Errorf("failed to open audit log file: %w", err)
		}

		// Read last line to get last hash
		lastHash, err = readLastHash(path)
		if err != nil {
			// Warn but don't fail, maybe file is corrupted or empty
			fmt.Fprintf(os.Stderr, "Warning: failed to read last hash from audit file: %v\n", err)
		}
	}

	return &FileAuditStore{
		file:     f,
		path:     path,
		lastHash: lastHash,
	}, nil
}

// readLastHash reads the last line of the file and extracts the hash.
// It handles large lines by reading from the end backwards in chunks.
func readLastHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer func() { _ = f.Close() }()

	stat, err := f.Stat()
	if err != nil {
		return "", err
	}
	size := stat.Size()
	if size == 0 {
		return "", nil
	}

	// We'll search for the last newline from the end.
	// We read in chunks of 4KB from the end. If we find a newline, we have the start of the last line.
	// If the last character is a newline (expected), we skip it and look for the previous newline.

	const chunkSize = 4096
	var lastLineData []byte

	// Start position (end of file)
	pos := size

	// We need to find the start of the last line.
	// The file might end with a newline.

	// Read last byte first to check for trailing newline
	b := make([]byte, 1)
	_, err = f.ReadAt(b, size-1)
	if err == nil && b[0] == '\n' {
		pos--
	}

	foundNewline := false

	for pos > 0 {
		readSize := int64(chunkSize)
		if pos < readSize {
			readSize = pos
		}

		pos -= readSize
		buf := make([]byte, readSize)
		_, err := f.ReadAt(buf, pos)
		if err != nil {
			return "", err
		}

		// Scan backwards for newline
		for i := len(buf) - 1; i >= 0; i-- {
			if buf[i] == '\n' {
				// Found start of last line
				startOfLine := pos + int64(i) + 1
				// Read from startOfLine to end (ignoring trailing newline logic handled by 'pos' initially but we need full content)

				// We need to read everything from startOfLine to the end of actual content (ignoring the very last byte if it was \n)
				// Actually, simpler: just Seek and ReadAll from startOfLine
				_, err := f.Seek(startOfLine, 0)
				if err != nil {
					return "", err
				}

				// Read until end
				reader := bufio.NewReader(f)
				line, err := reader.ReadBytes('\n')
				if err != nil && err != io.EOF {
					return "", err
				}
				// Trim trailing newline from the read line if present
				if len(line) > 0 && line[len(line)-1] == '\n' {
					line = line[:len(line)-1]
				}
				lastLineData = line
				foundNewline = true
				break
			}
		}
		if foundNewline {
			break
		}

		// If we didn't find a newline in this chunk, it means the last line is longer than this chunk + previous chunks.
		// We continue loop to read previous chunk.
	}

	if !foundNewline {
		// We reached start of file without finding a newline (except possibly at very end).
		// The whole file is one line.
		_, err := f.Seek(0, 0)
		if err != nil {
			return "", err
		}
		// Read entire file (careful with memory, but if we are here, we decided to support it)
		// Or limit it? Assuming audit logs fit in memory for now if it's a single entry.
		// Use ReadBytes to get the line safely.
		reader := bufio.NewReader(f)
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return "", err
		}
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		lastLineData = line
	}

	if len(lastLineData) == 0 {
		return "", nil
	}

	var entry AuditEntry
	if err := json.Unmarshal(lastLineData, &entry); err != nil {
		// Could be partial line or corrupted.
		return "", fmt.Errorf("failed to unmarshal last line: %w", err)
	}

	return entry.Hash, nil
}

// Write writes an audit entry to the file.
func (s *FileAuditStore) Write(_ context.Context, entry AuditEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Compute hash using current state
	entry.PrevHash = s.lastHash

	// Format args/result for hashing (needs to match what will be written)
	argsStr := ""
	if len(entry.Arguments) > 0 {
		argsStr = string(entry.Arguments)
	}

	resultStr := "{}"
	if entry.Result != nil {
		b, err := json.Marshal(entry.Result)
		if err == nil {
			resultStr = string(b)
		}
	}

	tsStr := entry.Timestamp.Format(time.RFC3339Nano)

	newHash := computeHash(tsStr, entry.ToolName, entry.UserID, entry.ProfileID, argsStr, resultStr, entry.Error, entry.DurationMs, entry.PrevHash)
	entry.Hash = newHash

	var w *os.File
	if s.file != nil {
		w = s.file
	} else {
		w = os.Stdout
	}

	// Write to disk
	if err := json.NewEncoder(w).Encode(entry); err != nil {
		return err
	}

	// Only update state if write succeeded
	s.lastHash = newHash

	return nil
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
