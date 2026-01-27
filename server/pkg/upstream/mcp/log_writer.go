// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
)

// LogWriter is an io.Writer that logs each line to an slog.Logger.
// It attempts to detect the log level from the content.
// It buffers input until a newline is found to ensure complete lines are logged.
type LogWriter struct {
	logger *slog.Logger
	mu     sync.Mutex
	buf    []byte
}

// NewLogWriter creates a new LogWriter.
func NewLogWriter(logger *slog.Logger) *LogWriter {
	return &LogWriter{
		logger: logger,
		buf:    make([]byte, 0, 1024),
	}
}

// Write implements io.Writer.
func (w *LogWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buf = append(w.buf, p...)

	// Process complete lines
	for {
		idx := bytes.IndexByte(w.buf, '\n')
		if idx == -1 {
			break
		}

		line := string(w.buf[:idx]) // Exclude newline
		w.buf = w.buf[idx+1:]       // Advance buffer

		// Log the line
		// Trim CR if present (Windows line endings)
		line = strings.TrimSuffix(line, "\r")

		level := w.detectLevel(line)
		w.logger.Log(context.Background(), level, line)
	}

	return len(p), nil
}

// Close flushes any remaining data in the buffer as a final log line.
func (w *LogWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.buf) > 0 {
		line := string(w.buf)
		// Trim CR if present
		line = strings.TrimSuffix(line, "\r")
		level := w.detectLevel(line)
		w.logger.Log(context.Background(), level, line)
		w.buf = w.buf[:0]
	}
	return nil
}

func (w *LogWriter) detectLevel(line string) slog.Level {
	lower := strings.ToLower(line)

	// Check for explicit prefixes or bracketed levels
	// We check case-insensitive match for common prefixes
	trimmed := strings.TrimSpace(lower)

	// Check for [LEVEL] or LEVEL: patterns
	if strings.HasPrefix(trimmed, "debug:") || strings.HasPrefix(trimmed, "[debug]") {
		return slog.LevelDebug
	}
	if strings.HasPrefix(trimmed, "info:") || strings.HasPrefix(trimmed, "[info]") {
		return slog.LevelInfo
	}
	if strings.HasPrefix(trimmed, "warn:") || strings.HasPrefix(trimmed, "warning:") || strings.HasPrefix(trimmed, "[warn]") || strings.HasPrefix(trimmed, "[warning]") {
		return slog.LevelWarn
	}
	if strings.HasPrefix(trimmed, "error:") || strings.HasPrefix(trimmed, "[error]") {
		return slog.LevelError
	}

	// Heuristic for content keywords if no explicit prefix found
	if strings.Contains(lower, "error") ||
		strings.Contains(lower, "fail") ||
		strings.Contains(lower, "fatal") ||
		strings.Contains(lower, "panic") ||
		strings.Contains(lower, "exception") {
		return slog.LevelError
	}

	if strings.Contains(lower, "warn") {
		return slog.LevelWarn
	}

	if strings.Contains(lower, "debug") {
		return slog.LevelDebug
	}

	// Default to INFO for stderr to avoid noise, as many tools log normal activity to stderr
	return slog.LevelInfo
}
