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

func BenchmarkPrimitiveSliceTokenization(b *testing.B) {
	t := NewSimpleTokenizer()

	// Create a large slice of ints
	ints := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		ints[i] = i * 12345
	}

	// Create a large slice of floats
	floats := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		floats[i] = float64(i) * 1.2345
	}

	// Create a large slice of floats that are actually integers
	floatInts := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		floatInts[i] = float64(i * 12345)
	}

	b.Run("IntSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = CountTokensInValue(t, ints)
		}
	})

	b.Run("FloatSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = CountTokensInValue(t, floats)
		}
	})

	b.Run("FloatIntSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = CountTokensInValue(t, floatInts)
		}
	})
}
