package util

import (
	"reflect"
	"testing"
)

func TestRedactMap_Coverage(t *testing.T) {
	input := map[string]interface{}{
		"simple": "value",
		"nested": map[string]interface{}{
			"secret": "value",
			"clean":  "value",
		},
		"list": []interface{}{
			map[string]interface{}{"api_key": "123"},
			"string",
			[]interface{}{"nested_list"},
		},
		"clean": map[string]interface{}{"a": "b"},
	}

	expected := map[string]interface{}{
		"simple": "value",
		"nested": map[string]interface{}{
			"secret": "[REDACTED]",
			"clean":  "value",
		},
		"list": []interface{}{
			map[string]interface{}{"api_key": "[REDACTED]"},
			"string",
			[]interface{}{"nested_list"},
		},
		"clean": map[string]interface{}{"a": "b"},
	}

	result := RedactMap(input)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("RedactMap mismatch.\nGot: %#v\nWant: %#v", result, expected)
	}
}

func TestIsSensitiveKey_Coverage_MatchFoldRest(t *testing.T) {
	// "api_key" is sensitive.
	// "api.key" should fail match on '_' vs '.'.
	// This exercises the `else { return false }` branch in `matchFoldRest`
	// when the character is not a-z but mismatches.
	if IsSensitiveKey("api.key") {
		t.Error("api.key should not be sensitive")
	}
}
