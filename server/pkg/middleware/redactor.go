// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"encoding/json"
	"log/slog"
	"regexp"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	// Common PII patterns.
	emailRegex      = regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`)
	creditCardRegex = regexp.MustCompile(`(?:\d{4}[-\s]?){3}\d{4}`)
	ssnRegex        = regexp.MustCompile(`\d{3}-\d{2}-\d{4}`)

	redactedStr = "***REDACTED***"
)

// Redactor handles redaction of sensitive data based on configuration.
type Redactor struct {
	patterns []*regexp.Regexp
	combined *regexp.Regexp
}

// NewRedactor creates a new Redactor from the given DLP config.
func NewRedactor(config *configv1.DLPConfig, log *slog.Logger) *Redactor {
	if config == nil || config.Enabled == nil || !*config.Enabled {
		return nil
	}

	patterns := []*regexp.Regexp{
		emailRegex,
		creditCardRegex,
		ssnRegex,
	}

	patternStrings := []string{
		emailRegex.String(),
		creditCardRegex.String(),
		ssnRegex.String(),
	}

	for _, p := range config.CustomPatterns {
		if r, err := regexp.Compile(p); err == nil {
			patterns = append(patterns, r)
			patternStrings = append(patternStrings, p)
		} else if log != nil {
			log.Warn("Invalid custom DLP pattern, ignoring", "pattern", p, "error", err)
		}
	}

	// Optimization: Combine all patterns into a single regex for better performance
	// and reduced memory allocations, especially for non-matching strings (common case).
	// We wrap each pattern in a non-capturing group (?:...) and join with |.
	validPatternStrings := make([]string, 0, len(patternStrings))
	for _, p := range patternStrings {
		// Ensure pattern is valid before combining (already checked above, but safe to double check implicitly)
		validPatternStrings = append(validPatternStrings, "(?:"+p+")")
	}

	var combined *regexp.Regexp
	if len(validPatternStrings) > 0 {
		joined := strings.Join(validPatternStrings, "|")
		c, err := regexp.Compile(joined)
		if err == nil {
			combined = c
		} else if log != nil {
			// This might happen if the combined regex is too large or complex
			log.Warn("Failed to compile combined DLP regex, falling back to sequential patterns", "error", err)
		}
	}

	return &Redactor{
		patterns: patterns,
		combined: combined,
	}
}

// RedactJSON redacts sensitive information from a JSON byte slice.
func (r *Redactor) RedactJSON(data []byte) ([]byte, error) {
	if r == nil || len(data) == 0 {
		return data, nil
	}

	var args map[string]interface{}
	// Try unmarshaling as map
	if err := json.Unmarshal(data, &args); err == nil {
		r.RedactStruct(args)
		return json.Marshal(args)
	}

	// Fallback: try unmarshaling as generic interface
	var genericArgs interface{}
	if err := json.Unmarshal(data, &genericArgs); err == nil {
		redacted := r.RedactValue(genericArgs)
		return json.Marshal(redacted)
	}

	return data, nil
}

// RedactString redacts sensitive information from a string.
func (r *Redactor) RedactString(s string) string {
	if r == nil {
		return s
	}

	// Optimization: Use combined regex if available
	if r.combined != nil {
		// Optimization: Check for match first to avoid allocation in ReplaceAllString
		// if there are no matches (which is the most common case).
		// MatchString is boolean and allocation-free (mostly).
		if !r.combined.MatchString(s) {
			return s
		}
		return r.combined.ReplaceAllString(s, redactedStr)
	}

	// Fallback to sequential patterns
	res := s
	for _, p := range r.patterns {
		res = p.ReplaceAllString(res, redactedStr)
	}
	return res
}

// RedactStruct redacts sensitive information from a map.
func (r *Redactor) RedactStruct(v map[string]interface{}) {
	if r == nil {
		return
	}
	for k, val := range v {
		v[k] = r.RedactValue(val)
	}
}

// RedactValue redacts sensitive information from a value.
func (r *Redactor) RedactValue(val interface{}) interface{} {
	if r == nil {
		return val
	}
	switch v := val.(type) {
	case string:
		return r.RedactString(v)
	case map[string]interface{}:
		r.RedactStruct(v)
		return v
	case []interface{}:
		for i, item := range v {
			v[i] = r.RedactValue(item)
		}
		return v
	case *structpb.Value:
		return val
	default:
		return val
	}
}
