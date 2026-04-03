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

// Redactor represents the public Redactor entity.
//
// Summary: Defines the structured data model representing a .
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Redactor struct {
	customPattern  *regexp.Regexp
	customPatterns []*regexp.Regexp
}

// NewRedactor serves as a public interface for interacting with NewRedactor.
//
// Summary: Constructs and returns an initialized redactor ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// RedactJSON serves as a public interface for interacting with RedactJSON.
//
// Summary: Redact the json appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r *Redactor) RedactJSON(data []byte) ([]byte, error) {
	if r == nil || len(data) == 0 {
		return data, nil
	}

	// Use streaming redaction to avoid full unmarshal/marshal cycle.
	// We use WalkJSONStrings to correctly handle comments (JSONC), ensuring we don't
	// accidentally redact strings inside comments or corrupt structure.
	return util.WalkJSONStrings(data, func(raw []byte) ([]byte, bool) {
		// Optimization: Check for obviously safe strings before unmarshaling.
		// If we have no custom patterns, we can skip unmarshaling if the raw bytes
		// (including quotes) don't contain indicators of PII: '@', digits, or escapes.
		// This avoids expensive json.Unmarshal for the vast majority of safe strings.
		if r.customPattern == nil && len(r.customPatterns) == 0 {
			hasIndicator := false
			// Check for '@' and '\' first using optimized SIMD scan
			if bytes.IndexByte(raw, '@') != -1 || bytes.IndexByte(raw, '\\') != -1 {
				hasIndicator = true
			} else if bytes.ContainsAny(raw, "0123456789") {
				// ⚡ BOLT: Replaced sequential bytes.IndexByte calls (up to 10 passes) with a single string.ContainsAny check.
				// Randomized Selection from Top 5 High-Impact Targets
				hasIndicator = true
			}

			if !hasIndicator {
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

// RedactString serves as a public interface for interacting with RedactString.
//
// Summary: Redact the string appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
		// ⚡ BOLT: Replaced sequential strings.IndexByte calls (up to 10 passes) with a single string.ContainsAny check.
		// Randomized Selection from Top 5 High-Impact Targets
		hasDigit = strings.ContainsAny(s, "0123456789")
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

// RedactStruct serves as a public interface for interacting with RedactStruct.
//
// Summary: Redact the struct appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r *Redactor) RedactStruct(v map[string]interface{}) {
	if r == nil {
		return
	}
	for k, val := range v {
		v[k] = r.RedactValue(val)
	}
}

// RedactValue serves as a public interface for interacting with RedactValue.
//
// Summary: Redact the value appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
