// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package logging provides logging utilities for the application.
package logging

import (
	"io"
	"log/slog"
	"os"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

var (
	mu            sync.Mutex
	once          sync.Once
	defaultLogger *slog.Logger
)

// ForTestsOnlyResetLogger is for use in tests to reset the `sync.Once`
// mechanism. This allows the global logger to be re-initialized in different
// test cases. This function should not be used in production code.
func ForTestsOnlyResetLogger() {
	mu.Lock()
	defer mu.Unlock()
	once = sync.Once{}
	defaultLogger = nil
}

// Init initializes the application's global logger with a specific log level
// and output destination. This function is designed to be called only once,
// typically at the start of the application, to ensure a consistent logging
// setup.
//
// Parameters:
//   - level: The minimum log level to be recorded (e.g., `slog.LevelInfo`).
//   - output: The `io.Writer` to which log entries will be written (e.g.,
//     `os.Stdout`).
//   - format: Optional format string ("json" or "text"). Defaults to "text".
func Init(level slog.Level, output io.Writer, format ...string) {
	mu.Lock()
	defer mu.Unlock()
	once.Do(func() {
		fmtStr := "text"
		if len(format) > 0 {
			fmtStr = format[0]
		}

		opts := &slog.HandlerOptions{
			Level:     level,
			AddSource: true,
		}

		var mainHandler slog.Handler
		if fmtStr == "json" {
			mainHandler = slog.NewJSONHandler(output, opts)
		} else {
			mainHandler = slog.NewTextHandler(output, opts)
		}

		broadcastHandler := NewBroadcastHandler(GlobalBroadcaster)
		teeHandler := NewTeeHandler(mainHandler, broadcastHandler)

		defaultLogger = slog.New(teeHandler)
	})
}

// GetLogger returns the shared global logger instance. If the logger has not yet
// been initialized through a call to `Init`, this function will initialize it
// with default settings: logging to `os.Stderr` at `slog.LevelInfo`.
//
// Returns:
//   - The global `*slog.Logger` instance.
func GetLogger() *slog.Logger {
	mu.Lock()
	defer mu.Unlock()
	once.Do(func() {
		defaultLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		}))
	})
	return defaultLogger
}

// ToSlogLevel converts a string log level to a slog.Level.
//
// Parameters:
//   - level: The log level from the configuration.
//
// Returns:
//   - The corresponding slog.Level.
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
