// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"fmt"
)

// CanonicalJSON marshals the input into a canonical JSON string.
// It ensures that map keys are sorted, which allows the output to be used as a deterministic cache key.
func CanonicalJSON(v any) (string, error) {
	// encoding/json sorts map keys by default.
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to canonical json: %w", err)
	}
	return string(b), nil
}
