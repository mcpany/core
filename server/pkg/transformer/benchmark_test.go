// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"testing"
)

// BenchmarkTextTemplate_Render ...
// Summary: BenchmarkTextTemplate_Render
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// BenchmarkTransformer_JoinStrings ...
// Summary: BenchmarkTransformer_JoinStrings
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	b.ReportAllocs()
	t := NewTransformer()
	// Create a large list
	listSize := 1000
	items := make([]any, listSize)
	for i := 0; i < listSize; i++ {
		items[i] = "item"
	}

	templateStr := `{{join "," .items}}`
	data := map[string]any{
		"items": items,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := t.Transform(templateStr, data)
		if err != nil {
			b.Fatalf("failed to transform: %v", err)
		}
	}
}

// BenchmarkTransformer_Join ...
// Summary: BenchmarkTransformer_Join
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	b.ReportAllocs()
	t := NewTransformer()
	// Create a large list
	listSize := 1000
	items := make([]any, listSize)
	for i := 0; i < listSize; i++ {
		items[i] = i
	}

	templateStr := `{{join "," .items}}`
	data := map[string]any{
		"items": items,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := t.Transform(templateStr, data)
		if err != nil {
			b.Fatalf("failed to transform: %v", err)
		}
	}
}

// BenchmarkTransformer_Transform ...
// Summary: BenchmarkTransformer_Transform
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t := NewTransformer()
	templateStr := "Hello, {{.name}}! You are {{.age}} years old. This is a {{.test}} of {{.performance}}."
	data := map[string]any{
		"name":        "World",
		"age":         99,
		"test":        "benchmark",
		"performance": "optimization",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := t.Transform(templateStr, data)
		if err != nil {
			b.Fatalf("failed to transform: %v", err)
		}
	}
}

// BenchmarkTextParser_ParseJSON ...
// Summary: BenchmarkTextParser_ParseJSON
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	parser := NewTextParser()
	jsonInput := []byte(`{"person": {"name": "test", "age": 123}}`)
	config := map[string]string{
		"name": `{.person.name}`,
		"age":  `{.person.age}`,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse("json", jsonInput, config, "")
		if err != nil {
			b.Fatalf("failed to parse: %v", err)
		}
	}
}

// BenchmarkTextParser_ParseXML ...
// Summary: BenchmarkTextParser_ParseXML
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	parser := NewTextParser()
	xmlInput := []byte(`<root><name>test</name><value>123</value></root>`)
	config := map[string]string{
		"name":  `//name`,
		"value": `//value`,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse("xml", xmlInput, config, "")
		if err != nil {
			b.Fatalf("failed to parse: %v", err)
		}
	}
}

// BenchmarkTextParser_ParseText ...
// Summary: BenchmarkTextParser_ParseText
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	parser := NewTextParser()
	textInput := []byte(`User ID: 12345, Name: John Doe`)
	config := map[string]string{
		"userId": `User ID: (\d+)`,
		"name":   `Name: ([\w\s]+)`,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse("text", textInput, config, "")
		if err != nil {
			b.Fatalf("failed to parse: %v", err)
		}
	}
}

// BenchmarkTextParser_ParseJQ ...
// Summary: BenchmarkTextParser_ParseJQ
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	parser := NewTextParser()
	jsonInput := []byte(`{"users": [{"name": "Alice"}, {"name": "Bob"}]}`)
	query := `{names: [.users[].name]}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse("jq", jsonInput, nil, query)
		if err != nil {
			b.Fatalf("failed to parse: %v", err)
		}
	}
}

// BenchmarkJoinStringsDirect ...
// Summary: BenchmarkJoinStringsDirect
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t := NewTransformer()
	listSize := 1000
	items := make([]string, listSize)
	for i := 0; i < listSize; i++ {
		items[i] = "item"
	}
	templateStr := `{{join "," .items}}`
	data := map[string]any{"items": items}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := t.Transform(templateStr, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkJoinIntsDirect ...
// Summary: BenchmarkJoinIntsDirect
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t := NewTransformer()
	listSize := 1000
	items := make([]int, listSize)
	for i := 0; i < listSize; i++ {
		items[i] = i
	}
	templateStr := `{{join "," .items}}`
	data := map[string]any{"items": items}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := t.Transform(templateStr, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
