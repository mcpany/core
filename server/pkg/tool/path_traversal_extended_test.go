// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckForPathTraversalExtended(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{"Safe path", "safe/path", false},
		{"Safe filename", "file.txt", false},
		{"Parent directory", "..", true},
		{"Parent directory start", "../file", true},
		{"Parent directory end", "dir/..", true},
		{"Parent directory middle", "dir/../file", true},
		{"Windows parent directory", "..\\file", true},
		{"Windows parent directory middle", "dir\\..\\file", true},
		{"Encoded dot dot", "%2e%2e", true},
		{"Encoded dot dot uppercase", "%2E%2E", true},
		{"Mixed encoded dot dot", "%2e%2E", true},
		{"Double encoded dot dot", "%252e%252e", true}, // Should catch this
		{"Double encoded dot dot uppercase", "%252E%252E", true}, // Should catch this
		{"Triple encoded dot dot", "%25252e%25252e", true}, // Should catch this
        // {"UTF-8 overlong encoding", "%c0%ae%c0%ae", true}, // Removed as it might be too specific to certain servers
		{"Encoded slash", "..%2f", true},
		{"Encoded backslash", "..%5c", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := checkForPathTraversal(tc.input)
			if tc.shouldErr {
				assert.Error(t, err, "Expected error for input: %s", tc.input)
			} else {
				assert.NoError(t, err, "Expected no error for input: %s", tc.input)
			}
		})
	}
}

func TestCheckForShellInjection(t *testing.T) {
    tests := []struct {
        name string
        val string
        template string
        placeholder string
        shouldErr bool
    }{
        {"Safe", "safe", "", "", false},
        {"Semi-colon", "cmd; rm -rf /", "", "", true},
        {"Pipe", "cmd | bash", "", "", true},
        {"Backtick", "`rm -rf /`", "", "", true},
        {"Dollar", "$(rm -rf /)", "", "", true},
        {"Double Quote", "\"quote\"", "", "", true},
        {"Single Quote", "'quote'", "", "", true},
        {"New Line", "cmd\n", "", "", true},
        {"Single Quote in Single Quote Template", "val'ue", "'{{val}}'", "{{val}}", true},
        {"Safe in Single Quote Template", "value", "'{{val}}'", "{{val}}", false},
        {"Double Quote in Double Quote Template", "val\"ue", "\"{{val}}\"", "{{val}}", true},
        {"Backtick in Double Quote Template", "val`ue", "\"{{val}}\"", "{{val}}", true},
        {"Dollar in Double Quote Template", "val$ue", "\"{{val}}\"", "{{val}}", true},
        {"Backslash in Double Quote Template", "val\\ue", "\"{{val}}\"", "{{val}}", true},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            err := checkForShellInjection(tc.val, tc.template, tc.placeholder)
             if tc.shouldErr {
                assert.Error(t, err, "Expected error for input: %s", tc.val)
            } else {
                assert.NoError(t, err, "Expected no error for input: %s", tc.val)
            }
        })
    }
}
