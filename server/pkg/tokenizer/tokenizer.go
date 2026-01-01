// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package tokenizer provides interfaces and implementations for counting tokens in text.
package tokenizer

import (
	"fmt"
	"strconv"
	"strings"
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
	words := strings.Fields(text)
	count := int(float64(len(words)) * t.Factor)
	if count < 1 && len(text) > 0 {
		count = 1
	}
	return count, nil
}

// CountTokensInValue recursively counts tokens in arbitrary structures.
func CountTokensInValue(t Tokenizer, v interface{}) (int, error) {
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
