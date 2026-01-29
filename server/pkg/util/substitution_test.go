package util //nolint:revive

import (
	"testing"
)

func TestReplaceURLPath_DoubleSubstitution(t *testing.T) {
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

func TestReplaceURLPath_Standard(t *testing.T) {
	t.Parallel()
	params := map[string]interface{}{
		"id": "123",
		"slug": "hello world",
	}

	path := "/users/{{id}}/posts/{{slug}}"
	result := ReplaceURLPath(path, params, nil)

	expected := "/users/123/posts/hello%20world"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestReplaceURLPath_MissingKey(t *testing.T) {
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
