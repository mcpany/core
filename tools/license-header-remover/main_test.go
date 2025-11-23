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
