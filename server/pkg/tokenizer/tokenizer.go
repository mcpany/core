// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package tokenizer provides interfaces and implementations for counting tokens in text.
package tokenizer

import (
	"fmt"
	"reflect"
	"strconv"
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
	visited := make(map[uintptr]bool)
	return countTokensInValueRecursive(t, v, visited)
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
		return countTokensReflect(t, val, visited)
	}
}

func countTokensInValueSimple(t *SimpleTokenizer, v interface{}, visited map[uintptr]bool) (int, error) {
	switch val := v.(type) {
	case string:
		return t.CountTokens(val)
	case int:
		return simpleTokenizeInt(val), nil
	case int64:
		return simpleTokenizeInt64(val), nil
	case bool:
		return 1, nil // "true" (4 chars) or "false" (5 chars) / 4 >= 1. Both result in 1 token.
	case nil:
		return 1, nil // "null" (4 chars) / 4 = 1
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
	case float64:
		return t.CountTokens(strconv.FormatFloat(val, 'g', -1, 64))
	default:
		return countTokensReflect(t, val, visited)
	}
}

func countTokensInValueWord(t *WordTokenizer, v interface{}, visited map[uintptr]bool) (int, error) {
	// Optimization: For primitive types, WordTokenizer effectively returns
	// max(1, int(1.0 * t.Factor)) assuming standard string representation
	// (no spaces).
	// Note: t.Factor defaults to 1.3. int(1.3) = 1.
	// If t.Factor is 2.0, int(2.0) = 2.
	// We assume standard primitives don't have spaces.
	primitiveCount := int(t.Factor)
	if primitiveCount < 1 {
		primitiveCount = 1
	}

	switch val := v.(type) {
	case string:
		return t.CountTokens(val)
	case int, int64, float64, bool, nil:
		return primitiveCount, nil
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
	default:
		return countTokensReflect(t, val, visited)
	}
}

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
		// We shouldn't defer delete(visited, ptr) because we want to detect shared references in a DAG too?
		// No, usually for token counting, we might count shared references multiple times if they appear multiple times.
		// But cycles must be avoided.
		// If we want to allow DAGs but prevent cycles, we should remove from visited after return.
		// However, for strict cycle detection in serialization-like traversal, removing after return is correct.
		// If we leave it, we treat it as "already counted" which deduplicates shared nodes.
		// Deduplication might be desired or not.
		// "CountTokensInValue" usually implies counting the full expanded structure.
		// So we should unmark it when leaving the branch.
		defer delete(visited, ptr)

		return countTokensInValueRecursive(t, val.Elem().Interface(), visited)
	case reflect.Struct:
		count := 0
		for i := 0; i < val.NumField(); i++ {
			if val.Type().Field(i).IsExported() {
				c, err := countTokensInValueRecursive(t, val.Field(i).Interface(), visited)
				if err != nil {
					return 0, err
				}
				count += c
			}
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
