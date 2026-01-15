// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"encoding/json"
	"log/slog"
	"regexp"

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
	customPatterns []*regexp.Regexp
}

// NewRedactor creates a new Redactor from the given DLP config.
func NewRedactor(config *configv1.DLPConfig, log *slog.Logger) *Redactor {
	if config == nil || config.Enabled == nil || !*config.Enabled {
		return nil
	}

	// Separate custom patterns from default ones for optimized processing
	var customPatterns []*regexp.Regexp
	for _, p := range config.CustomPatterns {
		if r, err := regexp.Compile(p); err == nil {
			customPatterns = append(customPatterns, r)
		} else if log != nil {
			log.Warn("Invalid custom DLP pattern, ignoring", "pattern", p, "error", err)
		}
	}

	return &Redactor{
		customPatterns: customPatterns,
	}
}

// RedactJSON redacts sensitive information from a JSON byte slice.
func (r *Redactor) RedactJSON(data []byte) ([]byte, error) {
	if r == nil || len(data) == 0 {
		return data, nil
	}

	// Use streaming redaction to avoid full unmarshal/marshal cycle.
	return util.WalkJSONStrings(data, func(raw []byte) ([]byte, bool) {
		// Optimization: Check for obviously safe strings before unmarshaling.
		// If the raw string (excluding quotes) doesn't contain '@' or digits,
		// and doesn't contain escapes (which might hide them), it's likely safe.
		// We can skip unmarshal for these cases if no custom patterns are defined.

		hasAt := false
		hasDigit := false
		hasEscape := false

		// Scan the raw bytes (which includes quotes, but they are not @, digit, or backslash)
		// We can safely scan the whole slice.
		for i := 0; i < len(raw); i++ {
			c := raw[i]
			if c == '@' {
				hasAt = true
			} else if c >= '0' && c <= '9' {
				hasDigit = true
			} else if c == '\\' {
				hasEscape = true
			}
			// If we found everything, we can stop scanning
			if hasAt && hasDigit && hasEscape {
				break
			}
		}

		// If no custom patterns and no PII indicators found in raw bytes, skip unmarshal.
		// We must unmarshal if there are escapes because they might hide PII (e.g. \u0040).
		if len(r.customPatterns) == 0 {
			if !hasEscape && !hasAt && !hasDigit {
				return nil, false
			}
		}

		var s string
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
func (r *Redactor) RedactString(s string) string {
	if r == nil {
		return s
	}

	// Optimization: Scan string once to check for characteristics of PII
	// This avoids expensive regex calls for strings that are obviously safe
	hasAt := false
	hasDigit := false
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
	for _, p := range r.customPatterns {
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
