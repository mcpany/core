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
	CountTokens(text string) (int, error)
}

// SimpleTokenizer implements a character-based heuristic.
// Logic: ~4 characters per token.
type SimpleTokenizer struct{}

// NewSimpleTokenizer creates a new SimpleTokenizer.
func NewSimpleTokenizer() *SimpleTokenizer {
	return &SimpleTokenizer{}
}

// CountTokens counts tokens in text using the simple heuristic.
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
func NewWordTokenizer() *WordTokenizer {
	return &WordTokenizer{Factor: 1.3}
}

// CountTokens counts tokens in text using the word-based heuristic.
func (t *WordTokenizer) CountTokens(text string) (int, error) {
	if len(text) == 0 {
		return 0, nil
	}

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

	count := int(float64(wordCount) * t.Factor)
	if count < 1 && len(text) > 0 {
		count = 1
	}
	return count, nil
}

// CountTokensInValue recursively counts tokens in arbitrary structures.
func CountTokensInValue(t Tokenizer, v interface{}) (int, error) {
	// OPTIMIZATION: Handle common primitive types and simple collections
	// without allocating the 'visited' map. This significantly improves performance
	// for simple inputs (strings, ints, etc.) which are common in metrics and logging.

	if st, ok := t.(*SimpleTokenizer); ok && st != nil {
		if c, handled, err := countTokensInValueSimpleFast(st, v); handled {
			return c, err
		}
	} else if wt, ok := t.(*WordTokenizer); ok && wt != nil {
		if c, handled, err := countTokensInValueWordFast(wt, v); handled {
			return c, err
		}
	} else {
		// Generic fallback for other Tokenizer implementations
		// We can still handle string safely
		if str, ok := v.(string); ok {
			return t.CountTokens(str)
		}
	}

	visited := make(map[uintptr]bool)
	return countTokensInValueRecursive(t, v, visited)
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
		c, err := st.CountTokens(strconv.FormatFloat(val, 'g', -1, 64))
		return c, true, err
	case []string:
		count := 0
		for _, item := range val {
			c, err := st.CountTokens(item)
			if err != nil {
				return 0, true, err
			}
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
		for _, item := range val {
			c, err := st.CountTokens(strconv.FormatFloat(item, 'g', -1, 64))
			if err != nil {
				return 0, true, err
			}
			count += c
		}
		return count, true, nil
	case map[string]string:
		count := 0
		for key, item := range val {
			kc, err := st.CountTokens(key)
			if err != nil {
				return 0, true, err
			}
			count += kc
			vc, err := st.CountTokens(item)
			if err != nil {
				return 0, true, err
			}
			count += vc
		}
		return count, true, nil
	}
	return 0, false, nil
}

// countTokensInValueWordFast handles fast-path tokenization for WordTokenizer.
// It returns (count, handled, error). If handled is false, the caller should fallback.
func countTokensInValueWordFast(wt *WordTokenizer, v interface{}) (int, bool, error) {
	// Calculate primitive count for WordTokenizer
	primitiveCount := int(wt.Factor)
	if primitiveCount < 1 {
		primitiveCount = 1
	}

	switch val := v.(type) {
	case string:
		c, err := wt.CountTokens(val)
		return c, true, err
	case int, int64, float64, bool, nil:
		return primitiveCount, true, nil
	case []string:
		count := 0
		for _, item := range val {
			c, err := wt.CountTokens(item)
			if err != nil {
				return 0, true, err
			}
			count += c
		}
		return count, true, nil
	case []int:
		return len(val) * primitiveCount, true, nil
	case []int64:
		return len(val) * primitiveCount, true, nil
	case []float64:
		return len(val) * primitiveCount, true, nil
	case []bool:
		return len(val) * primitiveCount, true, nil
	case map[string]string:
		count := 0
		for key, item := range val {
			kc, err := wt.CountTokens(key)
			if err != nil {
				return 0, true, err
			}
			count += kc
			vc, err := wt.CountTokens(item)
			if err != nil {
				return 0, true, err
			}
			count += vc
		}
		return count, true, nil
	}
	return 0, false, nil
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
	case []interface{}:
		count := 0
		for _, item := range val {
			c, err := countTokensInValueRecursive(t, item, visited)
			if err != nil {
				return 0, err
			}
			count += c
		}
		return count, nil
	case map[string]interface{}:
		count := 0
		for key, item := range val {
			// Count the key
			kc, err := t.CountTokens(key)
			if err != nil {
				return 0, err
			}
			count += kc

			// Count the value
			vc, err := countTokensInValueRecursive(t, item, visited)
			if err != nil {
				return 0, err
			}
			count += vc
		}
		return count, nil
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
		return countTokensReflect(t, val, visited, nil)
	}
}

func countTokensInValueSimple(t *SimpleTokenizer, v interface{}, visited map[uintptr]bool) (int, error) {
	// Try the fast path first (it handles the same types)
	if c, handled, err := countTokensInValueSimpleFast(t, v); handled {
		return c, err
	}

	switch val := v.(type) {
	case []interface{}:
		count := 0
		for _, item := range val {
			c, err := countTokensInValueSimple(t, item, visited)
			if err != nil {
				return 0, err
			}
			count += c
		}
		return count, nil
	case map[string]interface{}:
		count := 0
		for key, item := range val {
			// Count the key
			kc, err := t.CountTokens(key)
			if err != nil {
				return 0, err
			}
			count += kc

			// Count the value
			vc, err := countTokensInValueSimple(t, item, visited)
			if err != nil {
				return 0, err
			}
			count += vc
		}
		return count, nil
	default:
		return countTokensReflectGeneric(t, val, visited)
	}
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
	if c, handled, err := countTokensInValueWordFast(t, v); handled {
		return c, err
	}

	switch val := v.(type) {
	case []interface{}:
		count := 0
		for _, item := range val {
			c, err := countTokensInValueWord(t, item, visited)
			if err != nil {
				return 0, err
			}
			count += c
		}
		return count, nil
	case map[string]interface{}:
		count := 0
		for key, item := range val {
			// Count the key
			kc, err := t.CountTokens(key)
			if err != nil {
				return 0, err
			}
			count += kc

			// Count the value
			vc, err := countTokensInValueWord(t, item, visited)
			if err != nil {
				return 0, err
			}
			count += vc
		}
		return count, nil
	default:
		return countTokensReflectGeneric(t, val, visited)
	}
}

func (t *WordTokenizer) countRecursive(v interface{}, visited map[uintptr]bool) (int, error) {
	return countTokensInValueWord(t, v, visited)
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
		count := 0
		fields := getExportedFields(val.Type())
		for _, i := range fields {
			c, err := t.countRecursive(val.Field(i).Interface(), visited)
			if err != nil {
				return 0, err
			}
			count += c
		}
		return count, nil
	case reflect.Slice, reflect.Array:
		count := 0
		for i := 0; i < val.Len(); i++ {
			c, err := t.countRecursive(val.Index(i).Interface(), visited)
			if err != nil {
				return 0, err
			}
			count += c
		}
		return count, nil
	case reflect.Map:
		count := 0
		iter := val.MapRange()
		for iter.Next() {
			// Key
			kc, err := t.countRecursive(iter.Key().Interface(), visited)
			if err != nil {
				return 0, err
			}
			count += kc
			// Value
			vc, err := t.countRecursive(iter.Value().Interface(), visited)
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

// countTokensReflect is the fallback for non-recursiveTokenizer implementations.
func countTokensReflect(t Tokenizer, v interface{}, visited map[uintptr]bool, _ func(interface{}, map[uintptr]bool) (int, error)) (int, error) {
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
	case reflect.Slice, reflect.Array:
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
	l := 0
	if n == 0 {
		l = 1
	} else {
		if n < 0 {
			l = 1 // count the sign
			// Handle MinInt special case where -n overflows
			// For int64 (usually int is int64), MinInt is -9223372036854775808
			// which has 19 digits.
			// We can just divide by 10 once to make it safe to negate,
			// or process negative numbers.
		}

		for n != 0 {
			l++
			n /= 10
		}
	}

	count := l / 4
	if count < 1 {
		return 1
	}
	return count
}

func simpleTokenizeInt64(n int64) int {
	l := 0
	if n == 0 {
		l = 1
	} else {
		if n < 0 {
			l = 1 // count the sign
		}

		for n != 0 {
			l++
			n /= 10
		}
	}

	count := l / 4
	if count < 1 {
		return 1
	}
	return count
}
