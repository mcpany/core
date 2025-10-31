/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logging

import (
	"io"
	"log/slog"
	"os"
	"sync"
)

var (
	once          sync.Once
	defaultLogger *slog.Logger
)

// ForTestsOnlyResetLogger is for use in tests to reset the `sync.Once`
// mechanism. This allows the global logger to be re-initialized in different
// test cases. This function should not be used in production code.
func ForTestsOnlyResetLogger() {
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
func Init(level slog.Level, output io.Writer) {
	once.Do(func() {
		defaultLogger = slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{
			Level:     level,
			AddSource: true,
		}))
	})
}

// GetLogger returns the shared global logger instance. If the logger has not yet
// been initialized through a call to `Init`, this function will initialize it
// with default settings: logging to `os.Stderr` at `slog.LevelInfo`.
//
// Returns the global `*slog.Logger` instance.
func GetLogger() *slog.Logger {
	once.Do(func() {
		defaultLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		}))
	})
	return defaultLogger
}
