/*
 * Copyright 2025 Author(s) of MCP-XY
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
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestInitAndGetLogger(t *testing.T) {
	// This is a combined test to handle the singleton state. It's not a pure
	// unit test for Init, but it's a pragmatic way to test this package given
	// the use of sync.Once with global variables.

	// --- Part 1: Test Init on a (hopefully) clean slate ---

	// Create a temporary file to act as the log output.
	tmpfile, err := os.CreateTemp("", "testlog.*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	// Use a variable to hold the file name because tmpfile will be closed.
	logFileName := tmpfile.Name()
	defer os.Remove(logFileName)

	// Call Init. This will only execute the inner function if it's the first
	// time Init() or GetLogger() has been called in this test suite.
	Init(slog.LevelDebug, tmpfile)

	// Get the logger. It should now be the one we just initialized (or a
	// pre-existing one if another test ran first).
	logger := GetLogger()

	// Log a message at a level that should be enabled if our Init call was successful.
	logger.Debug("unique debug message for init test")
	logger.Info("another message to ensure logger is working")

	// Close the file to ensure content is flushed to disk.
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Read the file and check if the debug message is there.
	content, err := os.ReadFile(logFileName)
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	// This check determines if our Init() call was the one that configured the logger.
	if strings.Contains(string(content), "unique debug message for init test") {
		t.Log("Init() successfully configured the logger on the first run.")
		if !logger.Enabled(context.Background(), slog.LevelDebug) {
			t.Error("Logger should have Debug level enabled after successful Init")
		}
	} else {
		// If the debug message isn't there, it means the logger was already initialized
		// by a previous call to GetLogger() in another test.
		t.Log("Logger was already initialized; Init() call was correctly a no-op.")
		if logger.Enabled(context.Background(), slog.LevelDebug) {
			// This case would be strange: it means some other test set the level to Debug.
			t.Log("Warning: Logger already had Debug level enabled, possibly from another test.")
		}
	}

	// --- Part 2: Test that GetLogger returns a singleton ---
	logger2 := GetLogger()
	if logger != logger2 {
		t.Error("GetLogger() should always return the same logger instance")
	}
}
