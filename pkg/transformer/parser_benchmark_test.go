// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"testing"
)

func BenchmarkTextParser_Parse_Text(b *testing.B) {
	parser := NewTextParser()
	input := []byte("The quick brown fox jumps over the lazy dog. Order ID: 12345")
	config := map[string]string{
		"order_id": "Order ID: (\\d+)",
		"animal":   "(fox)",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse("text", input, config)
		if err != nil {
			b.Fatalf("failed to parse: %v", err)
		}
	}
}

func BenchmarkTextParser_Parse_JSON(b *testing.B) {
	parser := NewTextParser()
	input := []byte(`{"store": {"book": [{"category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95}]}}`)
	config := map[string]string{
		"author": "$.store.book[0].author",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse("json", input, config)
		if err != nil {
			b.Fatalf("failed to parse: %v", err)
		}
	}
}

func BenchmarkTextParser_Parse_XML(b *testing.B) {
	parser := NewTextParser()
	input := []byte(`<bookstore><book category="cooking"><title lang="en">Everyday Italian</title><author>Giada De Laurentiis</author><year>2005</year><price>30.00</price></book></bookstore>`)
	config := map[string]string{
		"author": "//book/author",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse("xml", input, config)
		if err != nil {
			b.Fatalf("failed to parse: %v", err)
		}
	}
}
