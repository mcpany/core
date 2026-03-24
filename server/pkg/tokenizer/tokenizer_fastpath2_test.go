package tokenizer

import (
	"testing"
	"math"
)

func TestCountTokensInValueFastPathMissingCoverage(t *testing.T) {
	st := NewSimpleTokenizer()
	wt := NewWordTokenizer()

	tests := []struct {
		name  string
		input interface{}
	}{
		{"map[string]int", map[string]int{"a": 1, "b": 2}},
		{"map[string]int64", map[string]int64{"a": 1, "b": 2}},
		{"map[string]float64", map[string]float64{"a": 1.1, "b": 2.2, "c": float64(3)}},
		{"map[string]bool", map[string]bool{"a": true, "b": false}},
		{"[]byte", []byte("hello world")},
		{"[]byte empty", []byte("")},
		{"[]byte long", []byte("hello world hello world hello world")},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/Simple", func(t *testing.T) {
			_, err := CountTokensInValue(st, tt.input)
			if err != nil {
				t.Fatalf("CountTokensInValue failed: %v", err)
			}
		})
		t.Run(tt.name+"/Word", func(t *testing.T) {
			_, err := CountTokensInValue(wt, tt.input)
			if err != nil {
				t.Fatalf("CountTokensInValue failed: %v", err)
			}
		})
	}

	// For simpleTokenizeInt64 missing paths
	t.Run("simpleTokenizeInt64 limits", func(t *testing.T) {
		CountTokensInValue(st, []interface{}{
			int64(math.MinInt64),
			int64(10),
			int64(100),
			int64(1000),
			int64(10000),
			int64(100000),
			int64(1000000),
			int64(10000000),
			int64(100000000),
			int64(1000000000),
			int64(10000000000),
			int64(100000000000),
			int64(1000000000000),
			int64(10000000000000),
			int64(100000000000000),
			int64(1000000000000000),
			int64(10000000000000000),
			int64(100000000000000000),
			int64(1000000000000000000),
			int64(math.MaxInt64),
		})
	})

	t.Run("countSliceInterface", func(t *testing.T) {
		sliceWithRecur := []interface{}{
			[]interface{}{1, 2},
			map[string]interface{}{"a": 1},
			"hello", float64(1.1), float64(1), int(1), int64(1), true, nil,
		}
		CountTokensInValue(st, sliceWithRecur)
		CountTokensInValue(wt, sliceWithRecur)
	})
}

func TestReflectStructures(t *testing.T) {
	st := NewSimpleTokenizer()
	wt := NewWordTokenizer()

	type CustomStruct struct {
		StringField string
		BoolField   bool
		IntField    int
	}

	s := CustomStruct{
		StringField: "hello",
		BoolField:   true,
		IntField:    123,
	}

	arr := [2]CustomStruct{s, s}

	// Test reflection slices
	CountTokensInValue(st, arr)
	CountTokensInValue(wt, arr)

	CountTokensInValue(st, s)
	CountTokensInValue(wt, s)

	type CycleStruct struct {
		Name string
		Next *CycleStruct
	}

	c1 := &CycleStruct{Name: "c1"}
	c2 := &CycleStruct{Name: "c2"}
	c1.Next = c2
	c2.Next = c1

	CountTokensInValue(st, c1)
	CountTokensInValue(wt, c1)

	m := map[interface{}]interface{}{
		"key": "value",
		123:   true,
		s:     s,
	}
	CountTokensInValue(st, m)
	CountTokensInValue(wt, m)
}
