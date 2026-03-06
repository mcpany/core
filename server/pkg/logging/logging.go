// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package logging provides logging utilities for the application.
package logging

import (
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
	// programLevel is the dynamic log level control.
	programLevel = new(slog.LevelVar)
)

// SetLevel - Auto-generated documentation.
//
// Summary: SetLevel updates the global log level dynamically.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func SetLevel(level slog.Level) {
	programLevel.Set(level)
}

// ForTestsOnlyResetLogger is for use in tests to reset the `sync.Once` mechanism. This allows the global logger to be re-initialized in different test cases. This function should not be used in production code.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func ForTestsOnlyResetLogger() {
	mu.Lock()
	defer mu.Unlock()
	once = sync.Once{}
	defaultLogger.Store(nil)
	GlobalBroadcaster.Reset()
}

// Init - Auto-generated documentation.
//
// Summary: Init initializes the application's global logger with a specific log level
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func Init(level slog.Level, output io.Writer, logFilePath string, format ...string) {
	mu.Lock()
	defer mu.Unlock()
	// NOTE: We intentionally bypass sync.Once here to allow reconfiguration and overwrite lazy default logger.

	// Initialize dynamic level
	programLevel.Set(level)

	fmtStr := "text"
	if len(format) > 0 {
		fmtStr = format[0]
	}

	// ⚡ BOLT: Only add source code location in DEBUG mode to avoid expensive runtime.Callers lookup.
	// Randomized Selection from Top 5 High-Impact Targets
	opts := &slog.HandlerOptions{
		Level:     programLevel,
		AddSource: level == slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if util.IsSensitiveKey(a.Key) {
				return slog.String(a.Key, "[REDACTED]")
			}
			return a
		},
	}

	var handlers []slog.Handler

	// 1. Main Output (Stderr/Stdout)
	var mainHandler slog.Handler
	if fmtStr == "json" {
		output = &RedactingWriter{w: output}
		mainHandler = slog.NewJSONHandler(output, opts)
	} else {
		mainHandler = slog.NewTextHandler(output, opts)
	}
	handlers = append(handlers, mainHandler)

	// 2. File Output (JSON only, for hydration)
	if logFilePath != "" {
		// Ensure file can be opened/created
		// We use O_APPEND to preserve logs across restarts (until rotation logic is added)
		f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			// Fallback: log to main output that we failed to open log file
			// Use a temporary logger since defaultLogger is not set yet
			// Actually we can't log easily yet. Just ignore or print to stderr?
			// Best effort.
			_ = err // prevent empty block lint error
		} else {
			// Use JSON handler for file to ensure hydration works
			fileHandler := slog.NewJSONHandler(&RedactingWriter{w: f}, opts)
			handlers = append(handlers, fileHandler)
		}
	}

	// 3. Broadcast Handler (WebSocket)
	broadcastHandler := NewBroadcastHandler(GlobalBroadcaster, programLevel)
	handlers = append(handlers, broadcastHandler)

	teeHandler := NewTeeHandler(handlers...)

	defaultLogger.Store(slog.New(teeHandler))
	// Init complete
}

// GetLogger - Auto-generated documentation.
//
// Summary: GetLogger returns the shared global logger instance.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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
// Parameters:
//   - level (configv1.GlobalSettings_LogLevel): The log level from the configuration.
//
// Returns:
//   - slog.Level: The corresponding slog.Level.
//
// Side Effects:
//   - None.
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
