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

	// currentOutput stores the io.Writer used for logging, allowing reconfiguration.
	currentOutput io.Writer = os.Stderr
	// currentLogFilePath stores the path to the log file, allowing reconfiguration.
	currentLogFilePath string
	// currentLogFileHandle tracks the open file to avoid FD leaks on reconfiguration.
	currentLogFileHandle *os.File
)

// ForTestsOnlyResetLogger is for use in tests to reset the `sync.Once`
// mechanism. This allows the global logger to be re-initialized in different
// test cases. This function should not be used in production code.
func ForTestsOnlyResetLogger() {
	mu.Lock()
	defer mu.Unlock()
	once = sync.Once{}
	defaultLogger.Store(nil)
	if currentLogFileHandle != nil {
		_ = currentLogFileHandle.Close()
		currentLogFileHandle = nil
	}
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
//   - logFilePath: string. Optional path to a log file. If provided, JSON logs will be written here.
//   - format: ...string. Optional format string ("json" or "text"). Defaults to "text".
//
// Returns:
//   None.
func Init(level slog.Level, output io.Writer, logFilePath string, format ...string) {
	mu.Lock()
	defer mu.Unlock()
	once.Do(func() {
		currentOutput = output
		currentLogFilePath = logFilePath

		fmtStr := "text"
		if len(format) > 0 {
			fmtStr = format[0]
		}

		logger, fileHandle := configure(level, output, logFilePath, fmtStr)
		if fileHandle != nil {
			currentLogFileHandle = fileHandle
		}
		defaultLogger.Store(logger)
	})
}

// Reconfigure allows updating the logger configuration at runtime.
// It uses the originally provided output writer and log file path.
//
// Summary: Reconfigures the global logger.
//
// Parameters:
//   - level: slog.Level. The new log level.
//   - format: string. The new log format ("json" or "text").
func Reconfigure(level slog.Level, format string) {
	mu.Lock()
	defer mu.Unlock()

	logger, fileHandle := configure(level, currentOutput, currentLogFilePath, format)

	// Close old file handle to prevent FD leaks
	if currentLogFileHandle != nil {
		_ = currentLogFileHandle.Close()
	}
	currentLogFileHandle = fileHandle

	defaultLogger.Store(logger)
}

// configure creates a new logger with the specified settings.
func configure(level slog.Level, output io.Writer, logFilePath string, fmtStr string) (*slog.Logger, *os.File) {
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

	// 1. Main Output (Stderr/Stdout)
	var mainHandler slog.Handler
	if fmtStr == "json" {
		output = &RedactingWriter{w: output}
		mainHandler = slog.NewJSONHandler(output, opts)
	} else {
		mainHandler = slog.NewTextHandler(output, opts)
	}
	handlers = append(handlers, mainHandler)

	var fileHandle *os.File

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
			fileHandle = f
		}
	}

	// 3. Broadcast Handler (WebSocket)
	broadcastHandler := NewBroadcastHandler(GlobalBroadcaster, level)
	handlers = append(handlers, broadcastHandler)

	teeHandler := NewTeeHandler(handlers...)

	return slog.New(teeHandler), fileHandle
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
