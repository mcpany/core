/*
 * Copyright 2025 Author(s) of MCPX
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

package tool

import (
	"testing"

	"github.com/mcpxy/mcpx/pkg/consts"
)

func TestParseToolName(t *testing.T) {
	testCases := []struct {
		name          string
		toolName      string
		wantNamespace string
		wantMethod    string
		wantErr       bool
	}{
		{
			name:          "Valid tool name",
			toolName:      "namespace" + consts.ToolNameServiceSeparator + "method",
			wantNamespace: "namespace",
			wantMethod:    "method",
			wantErr:       false,
		},
		{
			name:       "Invalid tool name - no section",
			toolName:   "namespacemethod",
			wantMethod: "namespacemethod",
			wantErr:    false,
		},
		{
			name:          "Invalid tool name - too many sections",
			toolName:      "namespace" + consts.ToolNameServiceSeparator + "method" + consts.ToolNameServiceSeparator + "extra",
			wantNamespace: "namespace",
			wantMethod:    "method" + consts.ToolNameServiceSeparator + "extra",
			wantErr:       false,
		},
		{
			name:     "Invalid tool name - empty",
			toolName: "",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			namespace, method, err := ParseToolName(tc.toolName)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseToolName() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr {
				if namespace != tc.wantNamespace {
					t.Errorf("ParseToolName() namespace = %v, want %v", namespace, tc.wantNamespace)
				}
				if method != tc.wantMethod {
					t.Errorf("ParseToolName() method = %v, want %v", method, tc.wantMethod)
				}
			}
		})
	}
}
