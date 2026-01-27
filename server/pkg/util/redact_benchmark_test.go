// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/util"
)

func BenchmarkRedactSecrets(b *testing.B) {
	// Setup a large text with secrets
	// 100KB text
	textChunk := "This is a log line with some secrets hidden inside it. "
	var sb strings.Builder
	for i := 0; i < 2000; i++ {
		sb.WriteString(textChunk)
		if i%10 == 0 {
			sb.WriteString("SECRET_TOKEN_XYZ_123 ")
		}
		if i%15 == 0 {
			sb.WriteString("ANOTHER_SECRET_ABC_456 ")
		}
		if i%20 == 0 {
			sb.WriteString("OVERLAPPING_SECRET ")
		}
		if i%20 == 1 {
			sb.WriteString("OVERLAPPING_SECRET_LONG ")
		}
	}
	text := sb.String()

	secrets := []string{
		"SECRET_TOKEN_XYZ_123",
		"ANOTHER_SECRET_ABC_456",
		"OVERLAPPING_SECRET",
		"OVERLAPPING_SECRET_LONG",
		"NON_EXISTENT_SECRET",
		"SHORT",
	}

	// Add some random secrets that are not in the text
	for i := 0; i < 20; i++ {
		secrets = append(secrets, "RANDOM_SECRET_" + string(rune(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		util.RedactSecrets(text, secrets)
	}
}
