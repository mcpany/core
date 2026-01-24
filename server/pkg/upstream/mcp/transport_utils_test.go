// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestTailBuffer(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		writes   []string
		expected string
	}{
		{
			name:     "Write within limit",
			limit:    10,
			writes:   []string{"hello"},
			expected: "hello",
		},
		{
			name:     "Write exceeds limit",
			limit:    5,
			writes:   []string{"hello world"},
			expected: "world",
		},
		{
			name:     "Multiple writes exceeding limit",
			limit:    5,
			writes:   []string{"hello", " ", "world"},
			expected: "world",
		},
		{
			name:     "Exact limit",
			limit:    5,
			writes:   []string{"hello"},
			expected: "hello",
		},
		{
			name:     "Empty write",
			limit:    5,
			writes:   []string{""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tb := &tailBuffer{limit: tt.limit}
			for _, w := range tt.writes {
				_, err := tb.Write([]byte(w))
				if err != nil {
					t.Fatalf("Write failed: %v", err)
				}
			}

			got := tb.String()
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestSlogWriter(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Remove time and level for easier testing
			if a.Key == slog.TimeKey || a.Key == slog.LevelKey {
				return slog.Attr{}
			}
			return a
		},
	})
	logger := slog.New(handler)
	writer := &slogWriter{log: logger, level: slog.LevelInfo}

	input := "line1\nline2"
	_, err := writer.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	// slog text handler format: msg="line1"\nmsg="line2"\n
	if !strings.Contains(output, "msg=line1") {
		t.Errorf("expected output to contain line1, got: %s", output)
	}
	if !strings.Contains(output, "msg=line2") {
		t.Errorf("expected output to contain line2, got: %s", output)
	}
}
