// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"encoding/json"
	"log/slog"
	"regexp"

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

	for _, p := range config.CustomPatterns {
		if r, err := regexp.Compile(p); err == nil {
			patterns = append(patterns, r)
		} else if log != nil {
			log.Warn("Invalid custom DLP pattern, ignoring", "pattern", p, "error", err)
		}
	}

	return &Redactor{
		patterns: patterns,
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
