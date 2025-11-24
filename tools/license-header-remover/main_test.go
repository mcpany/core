package main

import (
	"regexp"
	"testing"
)

func TestRefineEndIndex(t *testing.T) {
	// Setup regexes as they are global in main.go
	spdxRegex = regexp.MustCompile(`SPDX-License-Identifier`)
	limitationsRegex = regexp.MustCompile(`limitations under the License`)

	tests := []struct {
		name     string
		lines    []string
		start    int
		end      int
		expected int
	}{
		{
			name: "No markers",
			lines: []string{
				"// Copyright",
				"// Some text",
			},
			start:    0,
			end:      1,
			expected: 1, // Fallback to end
		},
		{
			name: "With SPDX",
			lines: []string{
				"// Copyright",
				"// SPDX-License-Identifier: Apache-2.0",
				"// Trailing comment",
			},
			start:    0,
			end:      2,
			expected: 1, // Stop at SPDX
		},
		{
			name: "With Limitations",
			lines: []string{
				"// Copyright",
				"// limitations under the License.",
				"// Trailing comment",
			},
			start:    0,
			end:      2,
			expected: 1, // Stop at Limitations
		},
		{
			name: "With Both, SPDX last",
			lines: []string{
				"// Copyright",
				"// limitations under the License.",
				"//",
				"// SPDX-License-Identifier: Apache-2.0",
				"// Trailing",
			},
			start:    0,
			end:      4,
			expected: 3, // Stop at SPDX
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := refineEndIndex(tt.lines, tt.start, tt.end)
			if got != tt.expected {
				t.Errorf("refineEndIndex() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFindBlockComment(t *testing.T) {
	tests := []struct {
		name      string
		lines     []string
		idx       int
		wantStart int
		wantEnd   int
	}{
		{
			name:      "Simple block",
			lines:     []string{"/*", " * Copyright", " */", "code"},
			idx:       1,
			wantStart: 0,
			wantEnd:   2,
		},
		{
			name:      "No start comment",
			lines:     []string{" * Copyright", " */", "code"},
			idx:       0,
			wantStart: -1,
			wantEnd:   -1,
		},
		{
			name:      "No end comment",
			lines:     []string{"/*", " * Copyright", "code"},
			idx:       1,
			wantStart: -1,
			wantEnd:   -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := findBlockComment(tt.lines, tt.idx)
			if gotStart != tt.wantStart || gotEnd != tt.wantEnd {
				t.Errorf("findBlockComment() = (%v, %v), want (%v, %v)", gotStart, gotEnd, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestIsHeaderBlock(t *testing.T) {
	shebangRegex = regexp.MustCompile(`^#!`)

	tests := []struct {
		name     string
		lines    []string
		startIdx int
		want     bool
	}{
		{
			name:     "Top of file",
			lines:    []string{"// Copyright"},
			startIdx: 0,
			want:     true,
		},
		{
			name:     "After shebang",
			lines:    []string{"#!/bin/bash", "# Copyright"},
			startIdx: 1,
			want:     true,
		},
		{
			name:     "After empty lines",
			lines:    []string{"", "", "// Copyright"},
			startIdx: 2,
			want:     true,
		},
		{
			name:     "After code",
			lines:    []string{"package main", "", "// Copyright"},
			startIdx: 2,
			want:     false,
		},
		{
			name:     "After comments",
			lines:    []string{"// build tag", "", "// Copyright"},
			startIdx: 2,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isHeaderBlock(tt.lines, tt.startIdx); got != tt.want {
				t.Errorf("isHeaderBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSourceFile(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"Go file", "main.go", true},
		{"Python file", "script.py", true},
		{"Shell script", "run.sh", true},
		{"YAML file", "config.yaml", true},
		{"YML file", "config.yml", true},
		{"Proto file", "service.proto", true},
		{"Makefile", "Makefile", true},
		{"Dockerfile", "Dockerfile", true},
		{"Text file", "notes.txt", false},
		{"No extension", "README", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSourceFile(tt.path); got != tt.want {
				t.Errorf("isSourceFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindBlock(t *testing.T) {
	shebangRegex = regexp.MustCompile(`^#!`)

	tests := []struct {
		name      string
		lines     []string
		idx       int
		prefix    string
		wantStart int
		wantEnd   int
	}{
		{
			name:      "Simple block",
			lines:     []string{"// L1", "// L2", "code"},
			idx:       0,
			prefix:    "//",
			wantStart: 0,
			wantEnd:   1,
		},
		{
			name:      "Middle of block",
			lines:     []string{"// L1", "// L2", "// L3"},
			idx:       1,
			prefix:    "//",
			wantStart: 0,
			wantEnd:   2,
		},
		{
			name:      "With shebang",
			lines:     []string{"#!/bin/sh", "# Copyright", "# License"},
			idx:       1,
			prefix:    "#",
			wantStart: 1, // Should stop at shebang (scan up)
			wantEnd:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := findBlock(tt.lines, tt.idx, tt.prefix)
			if gotStart != tt.wantStart || gotEnd != tt.wantEnd {
				t.Errorf("findBlock() = (%v, %v), want (%v, %v)", gotStart, gotEnd, tt.wantStart, tt.wantEnd)
			}
		})
	}
}
