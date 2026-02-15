package tokenizer

import (
	"math"
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
		{"Float64", 123.456},
		{"Float64 Sci", 1.23e10},

		// Slices
		{"Slice Int", []int{1, 2, 3, 10000, -123}},
		{"Slice Int64", []int64{1, 2, 3, 10000000000, -123}},
		{"Slice Bool", []bool{true, false, true}},
		{"Slice Float64", []float64{1.1, 2.2, 3.3, 1.23e5}},
		{"Slice String", []string{"hello", "world", "fast", "path"}},

		// Maps
		{"Map String String", map[string]string{"key1": "val1", "key2": "val2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/Simple", func(t *testing.T) {
			got, err := CountTokensInValue(st, tt.input)
			if err != nil {
				t.Fatalf("CountTokensInValue failed: %v", err)
			}

			// Validate correctness
			// We can compare against manual calculation or reflect-based fallback (if we could force it)
			// But since we can't easily force fallback, let's verify logic.

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
			if primCount < 1 { primCount = 1 }

			var want int
			switch v := tt.input.(type) {
			case int, int64, float64, bool, nil:
				want = primCount
			case []int:
				want = int(float64(len(v)) * wt.Factor)
				if want < 1 && len(v) > 0 { want = 1 }
			case []int64:
				want = int(float64(len(v)) * wt.Factor)
				if want < 1 && len(v) > 0 { want = 1 }
			case []float64:
				want = int(float64(len(v)) * wt.Factor)
				if want < 1 && len(v) > 0 { want = 1 }
			case []bool:
				want = int(float64(len(v)) * wt.Factor)
				if want < 1 && len(v) > 0 { want = 1 }
			case []string:
				var words int
				for _, x := range v {
					words += countWords(x)
				}
				want = int(float64(words) * wt.Factor)
				if want < 1 && words > 0 { want = 1 }
			case map[string]string:
				var words int
				for k, v := range v {
					words += countWords(k)
					words += countWords(v)
				}
				want = int(float64(words) * wt.Factor)
				if want < 1 && words > 0 { want = 1 }
			}

			if got != want {
				t.Errorf("Mismatch for %v: got %d, want %d", tt.input, got, want)
			}
		})
	}
}
