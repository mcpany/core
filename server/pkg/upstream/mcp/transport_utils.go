// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"bufio"
	"context"
	"log/slog"
	"strings"
	"sync"
)

// slogWriter implements the io.Writer interface, allowing it to be used as a
// destination for log output. It writes each line of the input to a slog.Logger.
type slogWriter struct {
	log   *slog.Logger
	level slog.Level
}

// Write takes a byte slice, scans it for lines, and logs each line
// individually using the configured slog.Logger and level.
func (s *slogWriter) Write(p []byte) (n int, err error) {
	scanner := bufio.NewScanner(strings.NewReader(string(p)))
	for scanner.Scan() {
		s.log.Log(context.Background(), s.level, scanner.Text())
	}
	return len(p), nil
}

// tailBuffer is a thread-safe buffer that keeps the last N bytes written to it.
type tailBuffer struct {
	buf   []byte
	limit int
	mu    sync.Mutex
}

// Write writes data to the buffer, maintaining the size limit.
//
// p is the p.
//
// Returns the result.
// Returns an error if the operation fails.
func (b *tailBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buf = append(b.buf, p...)
	if len(b.buf) > b.limit {
		// Keep the last 'limit' bytes
		b.buf = b.buf[len(b.buf)-b.limit:]
	}
	return len(p), nil
}

// String returns the buffered data as a string.
//
// Returns the result.
func (b *tailBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.buf)
}
