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

func BenchmarkTransformer_JoinStrings(b *testing.B) {
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

func BenchmarkTransformer_Join(b *testing.B) {
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

func BenchmarkTransformer_Transform(b *testing.B) {
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

func BenchmarkTextParser_ParseJSON(b *testing.B) {
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

func BenchmarkTextParser_ParseXML(b *testing.B) {
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

func BenchmarkTextParser_ParseText(b *testing.B) {
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

func BenchmarkTextParser_ParseJQ(b *testing.B) {
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
