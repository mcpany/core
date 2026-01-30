// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package tokenizer provides interfaces and implementations for counting tokens in text.
package tokenizer

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"unicode"
	"unicode/utf8"
)

// Tokenizer defines the interface for counting tokens in a given text.
type Tokenizer interface {
	// CountTokens estimates or calculates the number of tokens in the input text.
	//
	// text is the text.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	CountTokens(text string) (int, error)
}

// SimpleTokenizer implements a character-based heuristic.
// Logic: ~4 characters per token.
type SimpleTokenizer struct{}

// NewSimpleTokenizer creates a new SimpleTokenizer.
//
// Returns the result.
func NewSimpleTokenizer() *SimpleTokenizer {
	return &SimpleTokenizer{}
}

// CountTokens counts tokens in text using the simple heuristic.
//
// text is the text.
//
// Returns the result.
// Returns an error if the operation fails.
func (t *SimpleTokenizer) CountTokens(text string) (int, error) {
	if len(text) == 0 {
		return 0, nil
	}
	count := len(text) / 4
	if count < 1 {
		count = 1
	}
	return count, nil
}

// WordTokenizer implements a word-based heuristic.
// Logic: Count words (split by space) and multiply by a factor (e.g. 1.3) to account for subwords/punctuation.
type WordTokenizer struct {
	Factor float64
}

// NewWordTokenizer creates a new WordTokenizer with a default factor of 1.3.
//
// Returns the result.
func NewWordTokenizer() *WordTokenizer {
	return &WordTokenizer{Factor: 1.3}
}

// CountTokens counts tokens in text using the word-based heuristic.
//
// text is the text.
//
// Returns the result.
// Returns an error if the operation fails.
func (t *WordTokenizer) CountTokens(text string) (int, error) {
	if len(text) == 0 {
		return 0, nil
	}

	wordCount := countWords(text)

	count := int(float64(wordCount) * t.Factor)
	if count < 1 && len(text) > 0 {
		count = 1
	}
	return count, nil
}

func countWords(text string) int {
	// Count words by iterating through the string and counting transitions
	// from whitespace to non-whitespace. This avoids allocating a slice of strings.
	wordCount := 0
	inWord := false

	// Iterate by bytes for performance optimization.
	// For ASCII characters (< 128), we can check directly without decoding runes.
	i := 0
	n := len(text)
	for i < n {
		c := text[i]
		if c < utf8.RuneSelf {
			// ASCII fast path
			// Optimization: Check c > ' ' (most common case) first to avoid complex boolean logic
			if c > ' ' {
				if !inWord {
					inWord = true
					wordCount++
				}
				// OPTIMIZATION: Skip subsequent non-whitespace ASCII characters
				// This avoids checking inWord and c > ' ' repeatedly for each character in a word.
				for i+1 < n {
					c2 := text[i+1]
					if c2 <= ' ' || c2 >= utf8.RuneSelf {
						break
					}
					i++
				}
			} else {
				// c <= ' '
				// unicode.IsSpace for ASCII includes '\t', '\n', '\v', '\f', '\r', ' '.
				if c == ' ' || (c >= '\t' && c <= '\r') {
					inWord = false
					// OPTIMIZATION: Skip subsequent whitespace
					for i+1 < n {
						c2 := text[i+1]
						if c2 != ' ' && (c2 < '\t' || c2 > '\r') {
							break
						}
						i++
					}
				} else if !inWord {
					// Control characters that are not whitespace (e.g. \x00) start a word
					inWord = true
					wordCount++
				}
			}
			i++
		} else {
			// Multibyte character, decode rune
			r, w := utf8.DecodeRuneInString(text[i:])
			if unicode.IsSpace(r) {
				inWord = false
			} else if !inWord {
				inWord = true
				wordCount++
			}
			i += w
		}
	}
	return wordCount
}

// CountTokensInValue recursively counts tokens in arbitrary structures.
//
// t is the t.
// v is the v.
//
// Returns the result.
// Returns an error if the operation fails.
func CountTokensInValue(t Tokenizer, v interface{}) (int, error) {
	// OPTIMIZATION: Handle common primitive types and simple collections
	// without allocating the 'visited' map. This significantly improves performance
	// for simple inputs (strings, ints, etc.) which are common in metrics and logging.

	if st, ok := t.(*SimpleTokenizer); ok && st != nil {
		if c, handled, err := countTokensInValueSimpleFast(st, v); handled {
			return c, err
		}
	} else if wt, ok := t.(*WordTokenizer); ok && wt != nil {
		if c, handled := countTokensInValueWordFast(wt, v); handled {
			return c, nil
		}
	} else {
		// Generic fallback for other Tokenizer implementations
		// We can still handle string safely
		if str, ok := v.(string); ok {
			return t.CountTokens(str)
		}
	}

	visited := visitedPool.Get().(map[uintptr]bool)
	c, err := countTokensInValueRecursive(t, v, visited)

	// Ensure map is cleared before putting back
	clear(visited)
	visitedPool.Put(visited)

	return c, err
}

// rawWordCounter implements the recursiveTokenizer interface but counts raw words instead of tokens.
type rawWordCounter struct{}

// CountTokens counts the number of words in the text.
//
// Parameters:
//   - text: The input text.
//
// Returns:
//   - int: The word count.
//   - error: Always nil.
func (r *rawWordCounter) CountTokens(text string) (int, error) {
	return countWords(text), nil
}

func (r *rawWordCounter) countRecursive(v interface{}, visited map[uintptr]bool) (int, error) {
	// Try the fast path first
	if c, handled := countWordsInValueFast(v); handled {
		return c, nil
	}

	// Optimization: Handle map[string]interface{} explicitly to avoid reflection.
	if m, ok := v.(map[string]interface{}); ok {
		return countMapStringInterface(r, m, visited)
	}

	// Optimization: Handle []interface{} explicitly to avoid reflection.
	if s, ok := v.([]interface{}); ok {
		return countSliceInterface(r, s, visited)
	}

	return countTokensReflectGeneric(r, v, visited)
}

// countWordsInValueFast handles fast-path word counting.
func countWordsInValueFast(v interface{}) (int, bool) {
	switch val := v.(type) {
	case string:
		return countWords(val), true
	case int, int64, float64, bool, nil:
		return 1, true
	case []string:
		count := 0
		for _, item := range val {
			count += countWords(item)
		}
		return count, true
	case []int:
		return len(val), true
	case []int64:
		return len(val), true
	case []float64:
		return len(val), true
	case []bool:
		return len(val), true
	case map[string]string:
		count := 0
		for key, item := range val {
			count += countWords(key)
			count += countWords(item)
		}
		return count, true
	case []byte:
		// ⚡ BOLT: Optimization for []byte to avoid expensive reflection and Sprintf fallback.
		// Randomized Selection from Top 5 High-Impact Targets
		return countWords(string(val)), true
	}
	return 0, false
}

// countTokensInValueSimpleFast handles fast-path tokenization for SimpleTokenizer.
// It returns (count, handled, error). If handled is false, the caller should fallback.
func countTokensInValueSimpleFast(st *SimpleTokenizer, v interface{}) (int, bool, error) {
	switch val := v.(type) {
	case string:
		c, err := st.CountTokens(val)
		return c, true, err
	case int:
		return simpleTokenizeInt(val), true, nil
	case int64:
		return simpleTokenizeInt64(val), true, nil
	case bool:
		return 1, true, nil
	case nil:
		return 1, true, nil
	case float64:
		// OPTIMIZATION: Check if it is an integer to avoid string allocation/formatting.
		// We can safely use simpleTokenizeInt64 for all integer-valued floats because
		// standard JSON serialization (which this estimates) avoids scientific notation
		// until 1e21, while strconv.AppendFloat uses it earlier (1e6).
		// This alignment with JSON behavior also allows us to skip expensive float formatting.
		if i := int64(val); float64(i) == val {
			return simpleTokenizeInt64(i), true, nil
		}
		// OPTIMIZATION: Use stack buffer to avoid string allocation.
		// Logic must match SimpleTokenizer.CountTokens: len(text) / 4.
		var buf [64]byte
		b := strconv.AppendFloat(buf[:0], val, 'g', -1, 64)
		count := len(b) / 4
		if count < 1 {
			count = 1
		}
		return count, true, nil
	case []string:
		count := 0
		for _, item := range val {
			c, _ := st.CountTokens(item)
			count += c
		}
		return count, true, nil
	case []int:
		count := 0
		for _, item := range val {
			count += simpleTokenizeInt(item)
		}
		return count, true, nil
	case []int64:
		count := 0
		for _, item := range val {
			count += simpleTokenizeInt64(item)
		}
		return count, true, nil
	case []bool:
		return len(val), true, nil
	case []float64:
		count := 0
		// OPTIMIZATION: Reuse stack buffer to avoid string allocation per item.
		// Logic must match SimpleTokenizer.CountTokens: len(text) / 4.
		var buf [64]byte
		for _, item := range val {
			// OPTIMIZATION: Check if it is an integer to avoid string allocation/formatting.
			// Use simpleTokenizeInt64 for all integer-valued floats to match JSON behavior
			// and improve performance significantly.
			if i := int64(item); float64(i) == item {
				count += simpleTokenizeInt64(i)
				continue
			}
			b := strconv.AppendFloat(buf[:0], item, 'g', -1, 64)
			c := len(b) / 4
			if c < 1 {
				c = 1
			}
			count += c
		}
		return count, true, nil
	case map[string]string:
		count := 0
		for key, item := range val {
			kc, _ := st.CountTokens(key)
			count += kc
			vc, _ := st.CountTokens(item)
			count += vc
		}
		return count, true, nil
	case []byte:
		// ⚡ BOLT: Optimization for []byte to avoid expensive reflection and Sprintf fallback.
		// Randomized Selection from Top 5 High-Impact Targets
		if len(val) == 0 {
			return 0, true, nil
		}
		c := len(val) / 4
		if c < 1 {
			c = 1
		}
		return c, true, nil
	}
	// Fallback to generic unhandled case
	return 0, false, nil
}

// countTokensInValueWordFast handles fast-path tokenization for WordTokenizer.
// It returns (count, handled). If handled is false, the caller should fallback.
func countTokensInValueWordFast(wt *WordTokenizer, v interface{}) (int, bool) {
	if words, handled := countWordsInValueFast(v); handled {
		count := int(float64(words) * wt.Factor)
		if count < 1 && words > 0 {
			count = 1
		}
		return count, true
	}
	// Fallback to generic unhandled case
	return 0, false
}

func countTokensInValueRecursive(t Tokenizer, v interface{}, visited map[uintptr]bool) (int, error) {
	if st, ok := t.(*SimpleTokenizer); ok {
		return countTokensInValueSimple(st, v, visited)
	}
	if wt, ok := t.(*WordTokenizer); ok {
		return countTokensInValueWord(wt, v, visited)
	}

	switch val := v.(type) {
	case string:
		return t.CountTokens(val)
	case []string:
		count := 0
		for _, item := range val {
			c, err := t.CountTokens(item)
			if err != nil {
				return 0, err
			}
			count += c
		}
		return count, nil
	case map[string]string:
		count := 0
		for key, item := range val {
			// Count the key
			kc, err := t.CountTokens(key)
			if err != nil {
				return 0, err
			}
			count += kc

			// Count the value
			vc, err := t.CountTokens(item)
			if err != nil {
				return 0, err
			}
			count += vc
		}
		return count, nil
	case int:
		return t.CountTokens(strconv.Itoa(val))
	case int64:
		return t.CountTokens(strconv.FormatInt(val, 10))
	case float64:
		return t.CountTokens(strconv.FormatFloat(val, 'g', -1, 64))
	case bool:
		if val {
			return t.CountTokens("true")
		}
		return t.CountTokens("false")
	case nil:
		return t.CountTokens("null")
	default:
		return countTokensReflect(t, val, visited)
	}
}

func countTokensInValueSimple(t *SimpleTokenizer, v interface{}, visited map[uintptr]bool) (int, error) {
	// Try the fast path first (it handles the same types)
	if c, handled, err := countTokensInValueSimpleFast(t, v); handled {
		return c, err
	}

	// Optimization: Handle map[string]interface{} explicitly to avoid reflection.
	// This is very common for JSON data.
	if m, ok := v.(map[string]interface{}); ok {
		return countMapStringInterface(t, m, visited)
	}

	// Optimization: Handle []interface{} explicitly to avoid reflection.
	// This is very common for JSON lists.
	if s, ok := v.([]interface{}); ok {
		return countSliceInterface(t, s, visited)
	}

	return countTokensReflectGeneric(t, v, visited)
}

// recursiveTokenizer is an interface for tokenizers that support efficient recursive traversal.
type recursiveTokenizer interface {
	Tokenizer
	countRecursive(v interface{}, visited map[uintptr]bool) (int, error)
}

func (t *SimpleTokenizer) countRecursive(v interface{}, visited map[uintptr]bool) (int, error) {
	return countTokensInValueSimple(t, v, visited)
}

func countTokensInValueWord(t *WordTokenizer, v interface{}, visited map[uintptr]bool) (int, error) {
	// Try the fast path first
	if c, handled := countTokensInValueWordFast(t, v); handled {
		return c, nil
	}

	r := &rawWordCounter{}
	words, err := r.countRecursive(v, visited)
	if err != nil {
		return 0, err
	}

	count := int(float64(words) * t.Factor)
	if count < 1 && words > 0 {
		count = 1
	}
	return count, nil
}

func (t *WordTokenizer) countRecursive(v interface{}, visited map[uintptr]bool) (int, error) {
	return countTokensInValueWord(t, v, visited)
}

// visitedPool reuses maps to reduce allocations in CountTokensInValue.
var visitedPool = sync.Pool{
	New: func() interface{} {
		return make(map[uintptr]bool)
	},
}

// structFieldCache stores the indices of exported fields for a given struct type.
// This avoids calling IsExported() repeatedly which involves reflection overhead.
var structFieldCache sync.Map

func getExportedFields(typ reflect.Type) []int {
	if cached, ok := structFieldCache.Load(typ); ok {
		return cached.([]int)
	}

	numFields := typ.NumField()
	var fields []int
	// Pre-allocate assuming most fields might be exported, but safe to grow.
	// If the struct is large, this is beneficial.
	fields = make([]int, 0, numFields)

	for i := 0; i < numFields; i++ {
		if typ.Field(i).IsExported() {
			fields = append(fields, i)
		}
	}

	// Store in cache
	structFieldCache.Store(typ, fields)
	return fields
}

func countTokensReflectGeneric[T recursiveTokenizer](t T, v interface{}, visited map[uintptr]bool) (int, error) {
	// Check for fmt.Stringer first to respect custom formatting
	if s, ok := v.(fmt.Stringer); ok {
		return t.CountTokens(s.String())
	}
	if e, ok := v.(error); ok {
		return t.CountTokens(e.Error())
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			// Consistent nil handling: defer to tokenizer's string logic for "null"
			return t.CountTokens("null")
		}
		ptr := val.Pointer()
		if visited[ptr] {
			return 0, fmt.Errorf("cycle detected in value")
		}
		visited[ptr] = true
		defer delete(visited, ptr)

		return t.countRecursive(val.Elem().Interface(), visited)
	case reflect.Struct:
		return countTokensReflectStruct(t, val, visited)
	case reflect.Slice:
		if !val.IsNil() {
			ptr := val.Pointer()
			if visited[ptr] {
				return 0, fmt.Errorf("cycle detected in value")
			}
			visited[ptr] = true
			defer delete(visited, ptr)
		}
		fallthrough
	case reflect.Array:
		return countTokensReflectSlice(t, val, visited)
	case reflect.Map:
		return countTokensReflectMap(t, val, visited)
	}

	// Fallback for others (channels, funcs, unhandled types)
	return t.CountTokens(fmt.Sprintf("%v", v))
}

func countTokensReflectStruct[T recursiveTokenizer](t T, val reflect.Value, visited map[uintptr]bool) (int, error) {
	count := 0
	fields := getExportedFields(val.Type())
	for _, i := range fields {
		field := val.Field(i)
		// Optimization: Avoid Interface() allocation for string fields.
		if field.Kind() == reflect.String {
			c, err := t.CountTokens(field.String())
			if err != nil {
				return 0, err
			}
			count += c
			continue
		}
		// Optimization: Avoid Interface() allocation for bool fields.
		if field.Kind() == reflect.Bool {
			s := "false"
			if field.Bool() {
				s = "true"
			}
			c, err := t.CountTokens(s)
			if err != nil {
				return 0, err
			}
			count += c
			continue
		}

		c, err := t.countRecursive(field.Interface(), visited)
		if err != nil {
			return 0, err
		}
		count += c
	}
	return count, nil
}

func countTokensReflectSlice[T recursiveTokenizer](t T, val reflect.Value, visited map[uintptr]bool) (int, error) {
	count := 0
	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		// Optimization: Avoid Interface() allocation for string elements.
		if elem.Kind() == reflect.String {
			c, err := t.CountTokens(elem.String())
			if err != nil {
				return 0, err
			}
			count += c
			continue
		}

		c, err := t.countRecursive(elem.Interface(), visited)
		if err != nil {
			return 0, err
		}
		count += c
	}
	return count, nil
}

func countTokensReflectMap[T recursiveTokenizer](t T, val reflect.Value, visited map[uintptr]bool) (int, error) {
	if !val.IsNil() {
		ptr := val.Pointer()
		if visited[ptr] {
			return 0, fmt.Errorf("cycle detected in value")
		}
		visited[ptr] = true
		defer delete(visited, ptr)
	}

	count := 0
	iter := val.MapRange()
	for iter.Next() {
		// Key
		key := iter.Key()
		if key.Kind() == reflect.String {
			kc, err := t.CountTokens(key.String())
			if err != nil {
				return 0, err
			}
			count += kc
		} else {
			kc, err := t.countRecursive(key.Interface(), visited)
			if err != nil {
				return 0, err
			}
			count += kc
		}

		// Value
		entryVal := iter.Value()
		switch entryVal.Kind() {
		case reflect.String:
			vc, err := t.CountTokens(entryVal.String())
			if err != nil {
				return 0, err
			}
			count += vc
		case reflect.Bool:
			s := "false"
			if entryVal.Bool() {
				s = "true"
			}
			vc, err := t.CountTokens(s)
			if err != nil {
				return 0, err
			}
			count += vc
		default:
			vc, err := t.countRecursive(entryVal.Interface(), visited)
			if err != nil {
				return 0, err
			}
			count += vc
		}
	}
	return count, nil
}

// countTokensReflect is the fallback for non-recursiveTokenizer implementations.
func countTokensReflect(t Tokenizer, v interface{}, visited map[uintptr]bool) (int, error) {
	// Check for fmt.Stringer first to respect custom formatting
	if s, ok := v.(fmt.Stringer); ok {
		return t.CountTokens(s.String())
	}
	if e, ok := v.(error); ok {
		return t.CountTokens(e.Error())
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return t.CountTokens("null")
		}
		ptr := val.Pointer()
		if visited[ptr] {
			return 0, fmt.Errorf("cycle detected in value")
		}
		visited[ptr] = true
		defer delete(visited, ptr)

		return countTokensInValueRecursive(t, val.Elem().Interface(), visited)
	case reflect.Struct:
		count := 0
		fields := getExportedFields(val.Type())
		for _, i := range fields {
			c, err := countTokensInValueRecursive(t, val.Field(i).Interface(), visited)
			if err != nil {
				return 0, err
			}
			count += c
		}
		return count, nil
	case reflect.Slice:
		if !val.IsNil() {
			ptr := val.Pointer()
			if visited[ptr] {
				return 0, fmt.Errorf("cycle detected in value")
			}
			visited[ptr] = true
			defer delete(visited, ptr)
		}
		fallthrough
	case reflect.Array:
		count := 0
		for i := 0; i < val.Len(); i++ {
			c, err := countTokensInValueRecursive(t, val.Index(i).Interface(), visited)
			if err != nil {
				return 0, err
			}
			count += c
		}
		return count, nil
	case reflect.Map:
		if !val.IsNil() {
			ptr := val.Pointer()
			if visited[ptr] {
				return 0, fmt.Errorf("cycle detected in value")
			}
			visited[ptr] = true
			defer delete(visited, ptr)
		}

		count := 0
		iter := val.MapRange()
		for iter.Next() {
			// Key
			kc, err := countTokensInValueRecursive(t, iter.Key().Interface(), visited)
			if err != nil {
				return 0, err
			}
			count += kc
			// Value
			vc, err := countTokensInValueRecursive(t, iter.Value().Interface(), visited)
			if err != nil {
				return 0, err
			}
			count += vc
		}
		return count, nil
	}

	// Fallback for others (channels, funcs, unhandled types)
	return t.CountTokens(fmt.Sprintf("%v", v))
}

func simpleTokenizeInt(n int) int {
	// Optimization: Fast path for common integers.
	// Integers with < 8 chars (including sign) always result in 1 token (length/4 < 2).
	if n > -1000000 && n < 10000000 {
		return 1
	}
	// Fallback to generic int64 implementation
	return simpleTokenizeInt64(int64(n))
}

func countMapStringInterface[T recursiveTokenizer](t T, m map[string]interface{}, visited map[uintptr]bool) (int, error) {
	// Cycle detection
	val := reflect.ValueOf(m)
	if !val.IsNil() {
		ptr := val.Pointer()
		if visited[ptr] {
			return 0, fmt.Errorf("cycle detected in value")
		}
		visited[ptr] = true
		defer delete(visited, ptr)
	}

	count := 0
	for key, item := range m {
		kc, err := t.CountTokens(key)
		if err != nil {
			return 0, err
		}
		count += kc
		vc, err := t.countRecursive(item, visited)
		if err != nil {
			return 0, err
		}
		count += vc
	}
	return count, nil
}

func countSliceInterface[T recursiveTokenizer](t T, s []interface{}, visited map[uintptr]bool) (int, error) {
	// Cycle detection
	val := reflect.ValueOf(s)
	if !val.IsNil() {
		ptr := val.Pointer()
		if visited[ptr] {
			return 0, fmt.Errorf("cycle detected in value")
		}
		visited[ptr] = true
		defer delete(visited, ptr)
	}

	count := 0
	for _, item := range s {
		c, err := t.countRecursive(item, visited)
		if err != nil {
			return 0, err
		}
		count += c
	}
	return count, nil
}


func simpleTokenizeInt64(n int64) int {
	// Optimization: Fast path for common integers.
	// Integers with < 8 chars (including sign) always result in 1 token (length/4 < 2).
	if n > -1000000 && n < 10000000 {
		return 1
	}

	// Calculate length using if-chain for performance (approx 4x faster than loop).
	l := 0
	if n < 0 {
		l = 1 // count the sign
		if n == -9223372036854775808 { // MinInt64
			l += 19
			return (l / 4) // 20 / 4 = 5
		}
		n = -n
	}

	// Unrolled loop for digit counting
	switch {
	case n < 10:
		l++
	case n < 100:
		l += 2
	case n < 1000:
		l += 3
	case n < 10000:
		l += 4
	case n < 100000:
		l += 5
	case n < 1000000:
		l += 6
	case n < 10000000:
		l += 7
	case n < 100000000:
		l += 8
	case n < 1000000000:
		l += 9
	case n < 10000000000:
		l += 10
	case n < 100000000000:
		l += 11
	case n < 1000000000000:
		l += 12
	case n < 10000000000000:
		l += 13
	case n < 100000000000000:
		l += 14
	case n < 1000000000000000:
		l += 15
	case n < 10000000000000000:
		l += 16
	case n < 100000000000000000:
		l += 17
	case n < 1000000000000000000:
		l += 18
	default:
		l += 19
	}

	count := l / 4
	if count < 1 {
		return 1
	}
	return count
}
