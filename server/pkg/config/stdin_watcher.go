// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// StdinWatcher reads configuration updates from an io.Reader (typically stdin).
// It expects a stream of Newline Delimited JSON (NDJSON) objects, where each object
// is a complete McpAnyServerConfig.
type StdinWatcher struct {
	reader io.Reader
}

// NewStdinWatcher creates a new StdinWatcher.
func NewStdinWatcher(r io.Reader) *StdinWatcher {
	return &StdinWatcher{
		reader: r,
	}
}

// Watch starts reading from the input stream using a line-based scanner.
// For each valid configuration object read, it calls the updateFunc.
// It blocks until the stream is closed.
func (w *StdinWatcher) Watch(updateFunc func(*configv1.McpAnyServerConfig)) {
	scanner := bufio.NewScanner(w.reader)
	// Increase buffer size to handle large config files on a single line (up to 10MB)
	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Bytes()
		// Skip empty lines
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}

		var cfg configv1.McpAnyServerConfig
		// Use protojson for better protobuf support
		if err := protojson.Unmarshal(line, &cfg); err != nil {
			// Try standard json if protojson fails (though protojson is preferred)
			if err2 := json.Unmarshal(line, &cfg); err2 != nil {
				log.Printf("Error decoding config line from stdin: %v", err)
				continue
			}
		}

		log.Println("Received configuration update from stdin")
		updateFunc(&cfg)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from stdin: %v", err)
	}
}
