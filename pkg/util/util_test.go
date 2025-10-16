/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"testing"
)

func TestSanitizeOperationID(t *testing.T) {
	// Test case for the bug: a sequence of disallowed characters
	input := "test!!!name"
	expected := "test_9a7b00_name"
	output := SanitizeOperationID(input)
	if output != expected {
		t.Errorf("SanitizeOperationID(%q) = %q; want %q", input, output, expected)
	}

	// Test case with no disallowed characters
	input = "valid-name"
	expected = "valid-name"
	output = SanitizeOperationID(input)
	if output != expected {
		t.Errorf("SanitizeOperationID(%q) = %q; want %q", input, output, expected)
	}

	// Test case with a single disallowed character
	input = "test!name"
	expected = "test_0ab831_name"
	output = SanitizeOperationID(input)
	if output != expected {
		t.Errorf("SanitizeOperationID(%q) = %q; want %q", input, output, expected)
	}
}
