// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"regexp"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
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
	customPattern  *regexp.Regexp
	customPatterns []*regexp.Regexp
}

// NewRedactor creates a new Redactor from the given DLP config.
//
// config holds the configuration settings.
// log is the log.
//
// Returns the result.
func NewRedactor(config *configv1.DLPConfig, log *slog.Logger) *Redactor {
	if config == nil || !config.GetEnabled() {
		return nil
	}

	// Separate custom patterns from default ones for optimized processing
	var validPatterns []string
	var individualPatterns []*regexp.Regexp

	for _, p := range config.GetCustomPatterns() {
		if r, err := regexp.Compile(p); err == nil {
			// Wrap in non-capturing group to isolate alternatives
			validPatterns = append(validPatterns, "(?:"+p+")")
			individualPatterns = append(individualPatterns, r)
		} else if log != nil {
			log.Warn("Invalid custom DLP pattern, ignoring", "pattern", p, "error", err)
		}
	}

	var combined *regexp.Regexp
	if len(validPatterns) > 0 {
		// ⚡ BOLT: Combined multiple regex patterns into a single pass for O(1) execution overhead relative to pattern count.
		// Randomized Selection from Top 5 High-Impact Targets
		combinedStr := strings.Join(validPatterns, "|")
		var err error
		combined, err = regexp.Compile(combinedStr)
		if err != nil {
			if log != nil {
				// This can happen if patterns contain duplicate named capture groups across different patterns
				log.Warn("Failed to compile combined DLP pattern, falling back to individual checks", "error", err)
			}
			combined = nil
		}
	}

	return &Redactor{
		customPattern:  combined,
		customPatterns: individualPatterns,
	}
}

// RedactJSON redacts sensitive information from a JSON byte slice.
//
// data is the data.
//
// Returns the result.
// Returns an error if the operation fails.
func (r *Redactor) RedactJSON(data []byte) ([]byte, error) {
	if r == nil || len(data) == 0 {
		return data, nil
	}

	// Use streaming redaction to avoid full unmarshal/marshal cycle.
	// ⚡ Bolt Optimization: Use WalkStandardJSONStrings instead of WalkJSONStrings.
	// Request/Response bodies in our system are standard JSON, so we can skip
	// the expensive comment detection logic.
	return util.WalkStandardJSONStrings(data, func(raw []byte) ([]byte, bool) {
		// ⚡ BOLT: Optimized JSON redaction by skipping unmarshal for strings without escapes.
		// Randomized Selection from Top 5 High-Impact Targets

		// Check for escapes using optimized SIMD scan
		hasEscape := bytes.IndexByte(raw, '\\') != -1

		// Optimization: Check for obviously safe strings before unmarshaling.
		// If we have no custom patterns, we can skip unmarshaling if the raw bytes
		// (including quotes) don't contain indicators of PII: '@', digits, or escapes.
		// This avoids expensive json.Unmarshal for the vast majority of safe strings.
		if r.customPattern == nil && len(r.customPatterns) == 0 {
			hasIndicator := false
			// Check for '@' and '\' first using optimized SIMD scan
			if bytes.IndexByte(raw, '@') != -1 || hasEscape {
				hasIndicator = true
			} else {
				// Check for digits '0'-'9'
				// Using bytes.IndexByte for each digit is faster than a linear scan in Go for longer strings (approx > 64 bytes)
				// because it uses SIMD instructions. Since raw can be large, this is a significant win.
				// For very short strings, the overhead is negligible.
				for c := byte('0'); c <= '9'; c++ {
					if bytes.IndexByte(raw, c) != -1 {
						hasIndicator = true
						break
					}
				}
			}

			if !hasIndicator {
				return nil, false
			}
		}

		var s string

		// If no escapes, we can safely slice the string without unmarshaling.
		// This avoids allocation and parsing overhead for the vast majority of strings.
		// Standard JSON strings are just "content" where content can only contain escaped " or \
		// Since we checked for backslash, if none, it's a raw literal.
		// We also need to be careful about unicode sequences \uXXXX but those start with backslash.
		if !hasEscape && len(raw) >= 2 {
			// Fast path: direct slice to string conversion
			// raw includes quotes, e.g. "foo"
			s = string(raw[1 : len(raw)-1])

			redacted := r.RedactString(s)
			if redacted != s {
				// If redacted, we need to marshal it back to JSON to handle potential special chars in replacement
				// although usually replacement is ***REDACTED*** which is safe.
				b, err := json.Marshal(redacted)
				if err != nil {
					return nil, false
				}
				return b, true
			}
			return nil, false
		}

		if err := json.Unmarshal(raw, &s); err != nil {
			// Should not happen for valid JSON strings
			return nil, false
		}

		redacted := r.RedactString(s)
		if redacted != s {
			b, err := json.Marshal(redacted)
			if err != nil {
				return nil, false
			}
			return b, true
		}
		return nil, false
	}), nil
}

// RedactString redacts sensitive information from a string.
//
// s is the s.
//
// Returns the result.
func (r *Redactor) RedactString(s string) string {
	if r == nil {
		return s
	}

	// Optimization: Scan string once to check for characteristics of PII
	// This avoids expensive regex calls for strings that are obviously safe
	var hasAt, hasDigit bool

	// Hybrid approach:
	// For short strings, a manual loop is faster (single pass, low overhead).
	// For longer strings, multiple strings.IndexByte calls are faster (SIMD optimized).
	// The crossover point is around 64 bytes.
	if len(s) < 64 {
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c == '@' {
				hasAt = true
			} else if c >= '0' && c <= '9' {
				hasDigit = true
			}
			if hasAt && hasDigit {
				break
			}
		}
	} else {
		hasAt = strings.IndexByte(s, '@') != -1
		for c := byte('0'); c <= '9'; c++ {
			if strings.IndexByte(s, c) != -1 {
				hasDigit = true
				break
			}
		}
	}

	res := s

	// Only run email regex if '@' is present
	if hasAt {
		res = emailRegex.ReplaceAllString(res, redactedStr)
	}

	// Only run CC and SSN regexes if digits are present
	if hasDigit {
		res = creditCardRegex.ReplaceAllString(res, redactedStr)
		res = ssnRegex.ReplaceAllString(res, redactedStr)
	}

	// Always run custom patterns as we don't know their characteristics
	if r.customPattern != nil {
		res = r.customPattern.ReplaceAllString(res, redactedStr)
	} else {
		for _, p := range r.customPatterns {
			res = p.ReplaceAllString(res, redactedStr)
		}
	}
	return res
}

// RedactStruct redacts sensitive information from a map.
//
// v is the v.
func (r *Redactor) RedactStruct(v map[string]interface{}) {
	if r == nil {
		return
	}
	for k, val := range v {
		v[k] = r.RedactValue(val)
	}
}

// RedactValue redacts sensitive information from a value.
//
// val is the val.
//
// Returns the result.
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
