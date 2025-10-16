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

package util

import (
	"testing"

	"github.com/mcpxy/core/pkg/consts"
)

func TestSanitizeOperationID(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no sanitization needed",
			input: "simple-id_123",
			want:  "simple-id_123",
		},
		{
			name:  "sanitization of one sequence",
			input: "id with spaces",
			want:  "id_b858cb_with_b858cb_spaces",
		},
		{
			name:  "sanitization of multiple sequences",
			input: "a!!!b$$$c",
			want:  "a_0ab831__0ab831__0ab831_b_3cdf29__3cdf29__3cdf29_c",
		},
		{
			name:  "no sanitization needed with slash",
			input: "a/b",
			want:  "a/b",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SanitizeOperationID(tc.input)
			if got != tc.want {
				t.Errorf("SanitizeOperationID(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestGenerateToolID(t *testing.T) {
	cases := []struct {
		name       string
		serviceKey string
		toolName   string
		want       string
		wantErr    bool
	}{
		{
			name:       "valid tool name",
			serviceKey: "service",
			toolName:   "tool",
			want:       "service" + consts.ToolNameServiceSeparator + "tool",
		},
		{
			name:       "empty service key",
			serviceKey: "",
			toolName:   "tool",
			want:       "tool",
		},
		{
			name:     "empty tool name",
			toolName: "",
			wantErr:  true,
		},
		{
			name:     "invalid tool name",
			toolName: "tool!",
			wantErr:  true,
		},
		{
			name:     "fully qualified tool name",
			toolName: "service" + consts.ToolNameServiceSeparator + "tool",
			want:     "service" + consts.ToolNameServiceSeparator + "tool",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GenerateToolID(tc.serviceKey, tc.toolName)

			if (err != nil) != tc.wantErr {
				t.Fatalf("GenerateToolID() error = %v, wantErr %v", err, tc.wantErr)
			}

			if got != tc.want {
				t.Errorf("GenerateToolID() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseToolName(t *testing.T) {
	cases := []struct {
		name           string
		toolName       string
		wantService    string
		wantBareTool   string
		wantErr        bool
	}{
		{
			name:         "valid tool name",
			toolName:     "service" + consts.ToolNameServiceSeparator + "tool",
			wantService:  "service",
			wantBareTool: "tool",
		},
		{
			name:         "no service",
			toolName:     "tool",
			wantService:  "",
			wantBareTool: "tool",
		},
		{
			name:         "empty tool name",
			toolName:     "",
			wantService:  "",
			wantBareTool: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotService, gotBareTool, err := ParseToolName(tc.toolName)

			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseToolName() error = %v, wantErr %v", err, tc.wantErr)
			}

			if gotService != tc.wantService {
				t.Errorf("ParseToolName() gotService = %v, want %v", gotService, tc.wantService)
			}

			if gotBareTool != tc.wantBareTool {
				t.Errorf("ParseToolName() gotBareTool = %v, want %v", gotBareTool, tc.wantBareTool)
			}
		})
	}
}

func TestGenerateServiceKey(t *testing.T) {
	cases := []struct {
		name      string
		serviceID string
		want      string
		wantErr   bool
	}{
		{
			name:      "valid service id",
			serviceID: "service",
			want:      "service",
		},
		{
			name:      "empty service id",
			serviceID: "",
			wantErr:   true,
		},
		{
			name:      "invalid service id",
			serviceID: "service!",
			wantErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GenerateServiceKey(tc.serviceID)

			if (err != nil) != tc.wantErr {
				t.Fatalf("GenerateServiceKey() error = %v, wantErr %v", err, tc.wantErr)
			}

			if got != tc.want {
				t.Errorf("GenerateServiceKey() = %v, want %v", got, tc.want)
			}
		})
	}
}