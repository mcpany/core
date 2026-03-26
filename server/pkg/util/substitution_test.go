// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive

import (
	"testing"
)

// TestReplaceURLPath_DoubleSubstitution ...
// Summary: TestReplaceURLPath_DoubleSubstitution
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	// Case 1: "a" injects a placeholder for "b".
	params := map[string]interface{}{
		"a": "{{b}}",
		"b": "secret",
	}

	// We disable escaping for "a" so it injects raw "{{b}}"
	noEscape := map[string]bool{
		"a": true,
	}

	path := "/{{a}}"

	// We run multiple times to catch any flaky behavior (though regex should be deterministic).
	for i := 0; i < 100; i++ {
		result := ReplaceURLPath(path, params, noEscape)
		if result == "/secret" {
			t.Fatalf("Double substitution detected! Iteration %d, result: %s", i, result)
		}
		if result != "/{{b}}" {
			t.Errorf("Expected /{{b}}, got %s", result)
		}
	}
}

// TestReplaceURLPath_Standard ...
// Summary: TestReplaceURLPath_Standard
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	params := map[string]interface{}{
		"id":   "123",
		"slug": "hello world",
	}

	path := "/users/{{id}}/posts/{{slug}}"
	result := ReplaceURLPath(path, params, nil)

	expected := "/users/123/posts/hello%20world"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestReplaceURLPath_MissingKey ...
// Summary: TestReplaceURLPath_MissingKey
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	params := map[string]interface{}{
		"id": "123",
	}

	path := "/users/{{id}}/{{missing}}"
	result := ReplaceURLPath(path, params, nil)

	expected := "/users/123/{{missing}}"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
