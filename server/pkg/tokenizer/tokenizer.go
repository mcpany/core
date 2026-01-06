// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package tokenizer provides interfaces and implementations for counting tokens in text.
package tokenizer

import (
	"fmt"
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
			} else {
				// c <= ' '
				// unicode.IsSpace for ASCII includes '\t', '\n', '\v', '\f', '\r', ' '.
				if c == ' ' || (c >= '\t' && c <= '\r') {
					inWord = false
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
	if st, ok := t.(*SimpleTokenizer); ok {
		return countTokensInValueSimple(st, v)
	}
	if wt, ok := t.(*WordTokenizer); ok {
		return countTokensInValueWord(wt, v)
	}


	switch val := v.(type) {
	case string:
		return t.CountTokens(val)
	case []interface{}:
		count := 0
		for _, item := range val {
			c, err := CountTokensInValue(t, item)
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
			vc, err := CountTokensInValue(t, item)
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
		// Convert to string representation
		return t.CountTokens(fmt.Sprintf("%v", val))
	}
}

func countTokensInValueSimple(t *SimpleTokenizer, v interface{}) (int, error) {
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
			c, err := countTokensInValueSimple(t, item)
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
			vc, err := countTokensInValueSimple(t, item)
			if err != nil {
				return 0, err
			}
			count += vc
		}
		return count, nil
	case float64:
		return t.CountTokens(strconv.FormatFloat(val, 'g', -1, 64))
	default:
		// Convert to string representation
		return t.CountTokens(fmt.Sprintf("%v", val))
	}
}

func countTokensInValueWord(t *WordTokenizer, v interface{}) (int, error) {
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
			c, err := countTokensInValueWord(t, item)
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
			vc, err := countTokensInValueWord(t, item)
			if err != nil {
				return 0, err
			}
			count += vc
		}
		return count, nil
	default:
		// Convert to string representation
		return t.CountTokens(fmt.Sprintf("%v", val))
	}
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
