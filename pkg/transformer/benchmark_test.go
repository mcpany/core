// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"testing"
)

func BenchmarkTextTemplate_Render(b *testing.B) {
	templateString := "Hello, {{name}}! You are {{age}} years old. This is a {{test}} of {{performance}}."
	tpl, err := NewTemplate(templateString, "{{", "}}")
	if err != nil {
		b.Fatalf("failed to create template: %v", err)
	}

	params := map[string]any{
		"name":        "World",
		"age":         99,
		"test":        "benchmark",
		"performance": "optimization",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tpl.Render(params)
		if err != nil {
			b.Fatalf("failed to render: %v", err)
		}
	}
}
