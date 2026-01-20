// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net/url"
)

// RedactURL returns a string representation of the URL with sensitive query parameters and user credentials redacted.
func RedactURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	// copy the url to avoid modifying the original
	redacted := *u

	// Redact User Password if present
	if redacted.User != nil {
		if _, hasPass := redacted.User.Password(); hasPass {
			redacted.User = url.UserPassword(redacted.User.Username(), redactedPlaceholder)
		}
	}

	q := redacted.Query()
	for k := range q {
		if IsSensitiveKey(k) {
			q.Set(k, redactedPlaceholder)
		}
	}
	redacted.RawQuery = q.Encode()
	return redacted.String()
}

// RedactURLString parses the raw URL string and redacts sensitive query parameters and credentials.
// If parsing fails, it returns the original string.
func RedactURLString(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return RedactURL(u)
}
