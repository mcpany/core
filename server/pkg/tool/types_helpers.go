// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/validation"
)

// checkURLsInString extracts URLs from the input string and validates that they are safe.
// It returns an error if any extracted URL is unsafe (e.g. points to loopback or private network).
func checkURLsInString(s string) error {
	matches := urlPattern.FindAllString(s, -1)
	for _, match := range matches {
		if err := validation.IsSafeTarget(match); err != nil {
			// If IsSafeTarget returns an error, it might be due to invalid URL format OR unsafe target.
			// We only want to block if it's explicitly unsafe (loopback, private IP, etc.).
			// Generic parse errors might be due to the regex matching something that isn't really a URL,
			// or the URL being malformed but harmless.
			// However, relying on error message text is brittle.
			// But IsSafeTarget returns "invalid URL" if parsing fails.
			// It returns "unsupported scheme" if IsSafeURL was used (but we use IsSafeTarget which allows any scheme with host).
			// It returns "URL is missing host" if host is empty.
			// It returns "failed to resolve host" if DNS fails.
			// It returns "host ... resolves to unsafe IP ..." if unsafe.
			// It returns "loopback/private/link-local ... not allowed" if IP is unsafe.

			// We treat "invalid URL" and "failed to resolve host" as warnings but not blocks,
			// unless we are in strict mode. Here, we aim to prevent SSRF to KNOWN bad targets.
			// Blocking on DNS failure might cause flakiness if internet is down or DNS is flaky.
			// Blocking on "invalid URL" might block valid text.

			// So we check for specific security errors.
			errMsg := err.Error()
			if strings.Contains(errMsg, "is not allowed") || strings.Contains(errMsg, "resolves to unsafe IP") {
				return fmt.Errorf("unsafe url target in argument: %q: %w", match, err)
			}
		}
	}
	return nil
}
