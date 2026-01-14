// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBug_PascalCaseRedaction(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		shouldRedact bool
	}{
		// The bug was: "AuthId" (starts with "Auth" which is "auth" key) was skipped
		// because logic thought "Auth" + "I" (upper) meant it was like "AUTHORITY".
		{"AuthId", `{"AuthId": "123"}`, true},
		{"AuthID", `{"AuthID": "123"}`, true}, // "Auth" (mixed) + "I" (upper) -> Redact
		{"AUTHID", `{"AUTHID": "123"}`, false}, // "AUTH" (upper) + "I" (upper) -> Treated as upper case word (like AUTHORITY), so skipped.

		{"MyAuthId", `{"MyAuthId": "123"}`, true}, // "Auth" inside.
		{"AUTHORITY", `{"AUTHORITY": "public"}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := RedactJSON([]byte(tt.input))
			outStr := string(output)
			if tt.shouldRedact {
				assert.Contains(t, outStr, `[REDACTED]`, "Expected redaction for %s", tt.name)
			} else {
				assert.NotContains(t, outStr, `[REDACTED]`, "Expected NO redaction for %s", tt.name)
			}
		})
	}
}
