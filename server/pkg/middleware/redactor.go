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
	var customPatterns []*regexp.Regexp
	for _, p := range config.GetCustomPatterns() {
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
		// Optimization: Check for obviously safe strings before unmarshaling.
		// If we have no custom patterns, we can skip unmarshaling if the raw bytes
		// (including quotes) don't contain indicators of PII: '@', digits, or escapes.
		// This avoids expensive json.Unmarshal for the vast majority of safe strings.
		if len(r.customPatterns) == 0 {
			hasIndicator := false
			// Check for '@' and '\' first using optimized SIMD scan
			if bytes.IndexByte(raw, '@') != -1 || bytes.IndexByte(raw, '\\') != -1 {
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
	for _, p := range r.customPatterns {
		res = p.ReplaceAllString(res, redactedStr)
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
