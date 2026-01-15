// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package check

import (
	"context"
	"strings"
	"testing"
)

func TestValidateFile(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		wantErr     bool
		wantErrors  int
		errContains string
		checkLine   int
	}{
		{
			name:       "Valid Config",
			file:       "testdata/valid.yaml",
			wantErr:    false,
			wantErrors: 0,
		},
		{
			name:        "Invalid Schema",
			file:        "testdata/invalid_schema.yaml",
			wantErr:     false, // ValidateFile returns nil error but non-empty results for schema violations
			wantErrors:  1,
			errContains: "additionalProperties 'port' not allowed",
			checkLine:   8,
		},
		{
			name:        "Invalid Syntax",
			file:        "testdata/invalid_syntax.txt",
			wantErr:     true,
			errContains: "failed to parse YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := ValidateFile(context.Background(), tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
			}

			if len(results) != tt.wantErrors {
				t.Errorf("ValidateFile() returned %d errors, want %d", len(results), tt.wantErrors)
			}

			if len(results) > 0 && tt.errContains != "" {
				found := false
				for _, r := range results {
					if strings.Contains(r.Message, tt.errContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected validation errors to contain %q, got %v", tt.errContains, results)
				}
			}

			if tt.checkLine > 0 && len(results) > 0 {
				if results[0].Line != tt.checkLine {
					t.Errorf("expected error at line %d, got %d", tt.checkLine, results[0].Line)
				}
			}
		})
	}
}
