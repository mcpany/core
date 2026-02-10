// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package logging provides logging utilities for the application.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
)

var (
	mu            sync.Mutex
	once          sync.Once
	defaultLogger atomic.Pointer[slog.Logger]
)

// ForTestsOnlyResetLogger is for use in tests to reset the `sync.Once`
// mechanism. This allows the global logger to be re-initialized in different
// test cases. This function should not be used in production code.
func ForTestsOnlyResetLogger() {
	mu.Lock()
	defer mu.Unlock()
	once = sync.Once{}
	defaultLogger.Store(nil)
}

// Init initializes the application's global logger with a specific log level
// and output destination. This function is designed to be called only once,
// typically at the start of the application, to ensure a consistent logging
// setup.
//
// Summary: Initializes the global logger singleton.
//
// Parameters:
//   - level: slog.Level. The minimum log level to be recorded (e.g., `slog.LevelInfo`).
//   - output: io.Writer. The `io.Writer` to which log entries will be written (e.g., `os.Stdout`).
//   - logFile: string. Optional path to a log file. If provided, logs will be written to this file in JSON format.
//   - format: ...string. Optional format string ("json" or "text"). Defaults to "text".
//
// Returns:
//   None.
func Init(level slog.Level, output io.Writer, logFile string, format ...string) {
	mu.Lock()
	defer mu.Unlock()
	once.Do(func() {
		fmtStr := "text"
		if len(format) > 0 {
			fmtStr = format[0]
		}

		// ⚡ BOLT: Only add source code location in DEBUG mode to avoid expensive runtime.Callers lookup.
		// Randomized Selection from Top 5 High-Impact Targets
		opts := &slog.HandlerOptions{
			Level:     level,
			AddSource: level == slog.LevelDebug,
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				if util.IsSensitiveKey(a.Key) {
					return slog.String(a.Key, "[REDACTED]")
				}
				return a
			},
		}

		var handlers []slog.Handler

		// 1. Console Handler (User Preference)
		var consoleHandler slog.Handler
		if fmtStr == "json" {
			output = &RedactingWriter{w: output}
			consoleHandler = slog.NewJSONHandler(output, opts)
		} else {
			consoleHandler = slog.NewTextHandler(output, opts)
		}
		handlers = append(handlers, consoleHandler)

		// 2. File Handler (Always JSON, Optional)
		if logFile != "" {
			f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				// Fallback: Just print to stderr since logger isn't ready
				fmt.Fprintf(os.Stderr, "Failed to open log file %s: %v\n", logFile, err)
			} else {
				// We use a separate JSON handler for the file to ensure it's always machine-readable for hydration
				fileOutput := &RedactingWriter{w: f}
				fileHandler := slog.NewJSONHandler(fileOutput, opts)
				handlers = append(handlers, fileHandler)
			}
		}

		// 3. Broadcast Handler (WebSocket)
		broadcastHandler := NewBroadcastHandler(GlobalBroadcaster, level)
		handlers = append(handlers, broadcastHandler)

		// Combine all handlers
		teeHandler := NewTeeHandler(handlers...)

		defaultLogger.Store(slog.New(teeHandler))
	})
}

// GetLogger returns the shared global logger instance. If the logger has not yet
// been initialized through a call to `Init`, this function will initialize it
// with default settings: logging to `os.Stderr` at `slog.LevelInfo`.
//
// Summary: Retrieves the global logger instance.
//
// Returns:
//   - *slog.Logger: The global `*slog.Logger` instance.
func GetLogger() *slog.Logger {
	// ⚡ Bolt Optimization: Fast path to avoid lock contention on every log call.
	// Atomic load is much cheaper than mutex lock.
	if l := defaultLogger.Load(); l != nil {
		return l
	}

	mu.Lock()
	defer mu.Unlock()
	once.Do(func() {
		defaultLogger.Store(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			// ⚡ BOLT: Defaults to INFO, so AddSource is false by default.
			AddSource: false,
		})))
	})
	return defaultLogger.Load()
}

// ToSlogLevel converts a string log level to a slog.Level.
//
// Summary: Converts protobuf log level to slog level.
//
// Parameters:
//   - level: configv1.GlobalSettings_LogLevel. The log level from the configuration.
//
// Returns:
//   - slog.Level: The corresponding slog.Level.
func ToSlogLevel(level configv1.GlobalSettings_LogLevel) slog.Level {
	switch level {
	case configv1.GlobalSettings_LOG_LEVEL_DEBUG:
		return slog.LevelDebug
	case configv1.GlobalSettings_LOG_LEVEL_INFO:
		return slog.LevelInfo
	case configv1.GlobalSettings_LOG_LEVEL_WARN:
		return slog.LevelWarn
	case configv1.GlobalSettings_LOG_LEVEL_ERROR:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
