package tokenizer

import (
	"testing"
)

type BenchStruct struct {
	Name        string
	Description string
	Tags        []string
	Meta        map[string]string
	Count       int
	Active      bool
	Sub         *BenchStruct
}

func BenchmarkStructTokenization(b *testing.B) {
	t := NewSimpleTokenizer()

	deep := &BenchStruct{
		Name: "Deep",
		Sub: &BenchStruct{
			Name: "Deeper",
			Sub: &BenchStruct{
				Name: "Deepest",
			},
		},
	}

	s := BenchStruct{
		Name: "Test Struct",
		Description: "A description that is somewhat long to count tokens for.",
		Tags: []string{"tag1", "tag2", "tag3", "tag4", "tag5"},
		Meta: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
		Count: 12345,
		Active: true,
		Sub: deep,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CountTokensInValue(t, s)
	}
}
