package tokenizer_test

import (
	"testing"

	"github.com/mcpany/core/server/pkg/tokenizer"
)

type Tool struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

type ListToolsResult struct {
	Tools []*Tool
}

type ListToolsResultVal struct {
	Tools []Tool
}

func BenchmarkCountTokensInValue_StructPtr(b *testing.B) {
	tools := make([]*Tool, 100)
	for i := 0; i < 100; i++ {
		tools[i] = &Tool{
			Name:        "test_tool_" + string(rune(i)),
			Description: "This is a description for the test tool. It has some length to it.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"arg1": map[string]interface{}{"type": "string"},
				},
			},
		}
	}
	res := &ListToolsResult{Tools: tools}
	tok := tokenizer.NewSimpleTokenizer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tokenizer.CountTokensInValue(tok, res)
	}
}

func BenchmarkCountTokensInValue_StructVal(b *testing.B) {
	tools := make([]Tool, 100)
	for i := 0; i < 100; i++ {
		tools[i] = Tool{
			Name:        "test_tool_" + string(rune(i)),
			Description: "This is a description for the test tool. It has some length to it.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"arg1": map[string]interface{}{"type": "string"},
				},
			},
		}
	}
	res := &ListToolsResultVal{Tools: tools}
	tok := tokenizer.NewSimpleTokenizer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tokenizer.CountTokensInValue(tok, res)
	}
}
