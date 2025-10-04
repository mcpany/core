/*
 * Copyright 2025 Author(s) of MCPXY
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
	"log/slog"
	"os"
	"sync"
)

var (
	once          sync.Once
	defaultLogger *slog.Logger
)

// Init initializes the application's global logger with a specific log level
// and output destination. This function is designed to be called only once at
// the start of the application to ensure a consistent logging setup.
//
// level specifies the minimum log level to be recorded.
// output is the file to which log entries will be written.
func Init(level slog.Level, output *os.File) {
	once.Do(func() {
		defaultLogger = slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{
			Level:     level,
			AddSource: true,
		}))
	})
}

// GetLogger returns the shared global logger instance. If the logger has not yet
// been initialized through a call to Init, this function will initialize it with
// default settings: logging to os.Stderr at slog.LevelInfo.
func GetLogger() *slog.Logger {
	once.Do(func() {
		defaultLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		}))
	})
	return defaultLogger
}
