// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"math"
	"reflect"
	"strconv"
	"testing"
)

func TestCountTokensInValue_FastPathConsistency(t *testing.T) {
	st := NewSimpleTokenizer()
	wt := NewWordTokenizer()

	tests := []struct {
		name  string
		input interface{}
	}{
		{"Zero", 0},
		{"Small positive", 123},
		{"Small negative", -123},
		{"MaxInt", math.MaxInt},
		{"MinInt", math.MinInt},
		{"MaxInt64", int64(math.MaxInt64)},
		{"MinInt64", int64(math.MinInt64)},
		{"Bool true", true},
		{"Bool false", false},
		{"Nil", nil},
		{"Float64", 3.14159},
		{"Float64_Sci", 1.23e10},
		{"Slice_Int", []int{1, 2, 3, -4, math.MaxInt}},
		{"Slice_Int64", []int64{1, -2, math.MaxInt64}},
		{"Slice_Bool", []bool{true, false, true}},
		{"Slice_Float64", []float64{1.1, -2.2, 3.3e-5}},
		{"Slice_String", []string{"hello", "world", "this is a longer string"}},
		{"Map_String_String", map[string]string{"key1": "value1", "key2": "value2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/Simple", func(t *testing.T) {
			got, err := CountTokensInValue(st, tt.input)
			if err != nil {
				t.Fatalf("CountTokensInValue failed: %v", err)
			}

			// We need to mirror the logic from SimpleTokenizer
			// For basic types, it defaults to length/4 for string-like reps, but we use fast paths.
			// For consistency test, we can just ensure it doesn't panic and returns > 0 where expected.
			if got < 1 && tt.input != nil && tt.input != "" {
				// slices might evaluate to 0 if empty
				if reflect.TypeOf(tt.input).Kind() == reflect.Slice && reflect.ValueOf(tt.input).Len() == 0 {
					return
				}
				// maps might evaluate to 0 if empty
				if reflect.TypeOf(tt.input).Kind() == reflect.Map && reflect.ValueOf(tt.input).Len() == 0 {
					return
				}
				// Otherwise should be >= 1
				t.Errorf("Expected >= 1 tokens, got %d", got)
			}

			// For specific types, we can assert exactly against simpleTokenize*
			var want int
			switch v := tt.input.(type) {
			case int:
				want = simpleTokenizeInt(v)
			case int64:
				want = simpleTokenizeInt64(v)
			case bool:
				want = 1
			case nil:
				want = 1
			case float64:
				c, _ := st.CountTokens(strconv.FormatFloat(v, 'g', -1, 64))
				want = c
			case []int:
				for _, x := range v {
					want += simpleTokenizeInt(x)
				}
			case []int64:
				for _, x := range v {
					want += simpleTokenizeInt64(x)
				}
			case []bool:
				want = len(v)
			case []float64:
				for _, x := range v {
					c, _ := st.CountTokens(strconv.FormatFloat(x, 'g', -1, 64))
					want += c
				}
			case []string:
				for _, x := range v {
					c, _ := st.CountTokens(x)
					want += c
				}
			case map[string]string:
				for k, v := range v {
					kc, _ := st.CountTokens(k)
					vc, _ := st.CountTokens(v)
					want += kc + vc
				}
			}

			if got != want {
				t.Errorf("Mismatch for %v: got %d, want %d", tt.input, got, want)
			}
		})

		t.Run(tt.name+"/Word", func(t *testing.T) {
			got, err := CountTokensInValue(wt, tt.input)
			if err != nil {
				t.Fatalf("CountTokensInValue failed: %v", err)
			}

			primCount := int(wt.Factor)
			if primCount < 1 {
				primCount = 1
			}

			var want int
			switch v := tt.input.(type) {
			case int, int64, float64, bool, nil:
				want = primCount
			case []int:
				want = int(float64(len(v)) * wt.Factor)
				if want < 1 && len(v) > 0 {
					want = 1
				}
			case []int64:
				want = int(float64(len(v)) * wt.Factor)
				if want < 1 && len(v) > 0 {
					want = 1
				}
			case []float64:
				want = int(float64(len(v)) * wt.Factor)
				if want < 1 && len(v) > 0 {
					want = 1
				}
			case []bool:
				want = int(float64(len(v)) * wt.Factor)
				if want < 1 && len(v) > 0 {
					want = 1
				}
			case []string:
				var words int
				for _, x := range v {
					words += countWords(x)
				}
				want = int(float64(words) * wt.Factor)
				if want < 1 && words > 0 {
					want = 1
				}
			case map[string]string:
				var words int
				for k, v := range v {
					words += countWords(k)
					words += countWords(v)
				}
				want = int(float64(words) * wt.Factor)
				if want < 1 && words > 0 {
					want = 1
				}
			}

			if got != want {
				t.Errorf("Mismatch for %v: got %d, want %d", tt.input, got, want)
			}
		})
	}
}

func TestCountSliceInterfaceSimple(t *testing.T) {
	st := NewSimpleTokenizer()

	tests := []struct {
		name    string
		input   []interface{}
		want    int
		wantErr bool
	}{
		{"empty", []interface{}{}, 0, false},
		{"string", []interface{}{"hello", "world"}, 2, false}, // 5/4 + 5/4 = 1+1=2 tokens, simple heuristic
		{"int", []interface{}{123, 456}, 2, false},
		{"float_int", []interface{}{123.0}, 1, false},
		{"float_frac", []interface{}{3.14159}, 1, false}, // 7 chars / 4 = 1 token
		{"bool_nil", []interface{}{true, false, nil}, 3, false},
		{"nested_slice", []interface{}{[]interface{}{"hello"}}, 1, false},
		{"unsupported_fallback", []interface{}{complex(1, 2)}, 1, false}, // Uses countTokensInValueRecursive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := countSliceInterfaceSimple(st, tt.input, make(map[uintptr]bool))
			if (err != nil) != tt.wantErr {
				t.Errorf("countSliceInterfaceSimple() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("countSliceInterfaceSimple() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountSliceInterfaceSimple_Cycle(t *testing.T) {
	st := NewSimpleTokenizer()

	s3 := make([]interface{}, 1)
	s3[0] = s3

	_, err := countSliceInterfaceSimple(st, s3, make(map[uintptr]bool))
	if err == nil {
		t.Errorf("Expected cycle error, got nil")
	}
}

func TestSimpleTokenizeInt64(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  int
	}{
		{"small_positive", 123456, 1},
		{"small_negative", -123456, 1},
		{"edge_10", 10, 1},
		{"edge_100", 100, 1},
		{"edge_1000", 1000, 1},
		{"edge_10000", 10000, 1},
		{"edge_100000", 100000, 1},
		{"edge_1000000", 1000000, 1},
		{"edge_10000000", 10000000, 2},                       // 8 chars / 4 = 2 tokens
		{"edge_100000000", 100000000, 2},                     // 9 chars
		{"edge_1000000000", 1000000000, 2},                   // 10 chars -> 10/4 = 2
		{"edge_10000000000", 10000000000, 2},                 // 11 chars -> 11/4 = 2
		{"edge_100000000000", 100000000000, 3},               // 12 chars -> 12/4 = 3
		{"edge_1000000000000", 1000000000000, 3},             // 13 chars
		{"edge_10000000000000", 10000000000000, 3},           // 14 chars
		{"edge_100000000000000", 100000000000000, 3},         // 15 chars
		{"edge_1000000000000000", 1000000000000000, 4},       // 16 chars -> 16/4 = 4
		{"edge_10000000000000000", 10000000000000000, 4},     // 17 chars
		{"edge_100000000000000000", 100000000000000000, 4},   // 18 chars
		{"edge_1000000000000000000", 1000000000000000000, 4}, // 19 chars -> 19/4 = 4
		{"min_int64", -9223372036854775808, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := simpleTokenizeInt64(tt.input); got != tt.want {
				t.Errorf("simpleTokenizeInt64(%d) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSimpleTokenizeInt64_FullCoverage(t *testing.T) {
	// Full coverage for switch cases.
	tests := []struct {
		input int64
		want  int
	}{
		{9, 1},
		{99, 1},
		{999, 1},
		{9999, 1},
		{99999, 1},
		{999999, 1},
		{9999999, 1},
		{99999999, 2},
		{999999999, 2},
		{9999999999, 2},
		{99999999999, 2},
		{999999999999, 3},
		{9999999999999, 3},
		{99999999999999, 3},
		{999999999999999, 3},
		{9999999999999999, 4},
		{99999999999999999, 4},
		{999999999999999999, 4},
	}

	for _, tc := range tests {
		got := simpleTokenizeInt64(tc.input)
		if got != tc.want {
			t.Errorf("simpleTokenizeInt64(%d) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestCountTokensReflectStruct_AllKinds(t *testing.T) {
	st := NewSimpleTokenizer()

	type MyStruct struct {
		StringField string
		BoolTrue    bool
		BoolFalse   bool
		IntField    int
		InterfaceF  interface{}
		_           string
	}

	val := reflect.ValueOf(MyStruct{
		StringField: "hello world",
		BoolTrue:    true,
		BoolFalse:   false,
		IntField:    123,
		InterfaceF:  "interface string",
	})

	got, err := countTokensReflectStruct(st, val, make(map[uintptr]bool))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := 9
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestCountTokensReflectSlice(t *testing.T) {
	st := NewSimpleTokenizer()

	type MyStruct struct {
		Name string
	}

	tests := []struct {
		name    string
		input   interface{}
		want    int
		wantErr bool
	}{
		{"[]string", []string{"hello", "world"}, 2, false},
		{"[]int", []int{1, 2, 3}, 3, false},
		{"[]MyStruct", []MyStruct{{"A"}, {"B"}}, 2, false}, // 1 token each
		{"[]bool", []bool{true, false}, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := reflect.ValueOf(tt.input)
			got, err := countTokensReflectSlice(st, val, make(map[uintptr]bool))
			if (err != nil) != tt.wantErr {
				t.Errorf("countTokensReflectSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("countTokensReflectSlice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountTokensReflectMap(t *testing.T) {
	st := NewSimpleTokenizer()

	tests := []struct {
		name    string
		input   interface{}
		want    int
		wantErr bool
	}{
		{"map[string]string", map[string]string{"k": "v"}, 2, false}, // key(1) + value(1)
		{"map[int]int", map[int]int{1: 2, 3: 4}, 4, false},
		{"map[string]bool", map[string]bool{"k1": true, "k2": false}, 4, false},
		{"map[string]interface{}", map[string]interface{}{"k1": "hello"}, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := reflect.ValueOf(tt.input)
			got, err := countTokensReflectMap(st, val, make(map[uintptr]bool))
			if (err != nil) != tt.wantErr {
				t.Errorf("countTokensReflectMap() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("countTokensReflectMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountSliceInterfaceRaw(t *testing.T) {

	tests := []struct {
		name    string
		input   []interface{}
		want    int
		wantErr bool
	}{
		{"empty", []interface{}{}, 0, false},
		{"strings", []interface{}{"hello", "world"}, 2, false},
		{"mixed", []interface{}{"hello", 123, true, nil}, 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := countSliceInterfaceRaw(&rawWordCounter{}, tt.input, make(map[uintptr]bool))
			if (err != nil) != tt.wantErr {
				t.Errorf("countSliceInterfaceRaw() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("countSliceInterfaceRaw() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountSliceInterfaceRaw_Cycle(t *testing.T) {

	s := make([]interface{}, 1)
	s[0] = s

	_, err := countSliceInterfaceRaw(&rawWordCounter{}, s, make(map[uintptr]bool))
	if err == nil {
		t.Errorf("Expected cycle error, got nil")
	}
}

func TestCountWordsInValueFast_Extra(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  int
		ok    bool
	}{
		{"map[string]int", map[string]int{"k1": 1}, 2, true},
		{"map[string]int64", map[string]int64{"k1": 1}, 2, true},
		{"map[string]float64", map[string]float64{"k1": 1.1}, 2, true},
		{"map[string]bool", map[string]bool{"k1": true}, 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := countWordsInValueFast(tt.input)
			if ok != tt.ok {
				t.Errorf("countWordsInValueFast() ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Errorf("countWordsInValueFast() got = %v, want %v", got, tt.want)
			}
		})
	}
}
