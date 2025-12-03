/*
 * Copyright 2025 Author(s) of MCP Any
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

package cli

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestJSONExecutor(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		in := &bytes.Buffer{}
		out := &bytes.Buffer{}

		executor := NewJSONExecutor(in, out)

		// Test data
		requestData := map[string]string{"hello": "world"}
		responseData := map[string]string{"foo": "bar"}

		// Encode the response data to the output buffer
		if err := json.NewEncoder(out).Encode(responseData); err != nil {
			t.Fatalf("failed to encode response data: %v", err)
		}

		// Execute the command
		var resultData map[string]string
		if err := executor.Execute(requestData, &resultData); err != nil {
			t.Fatalf("failed to execute command: %v", err)
		}

		// Check the input buffer for the correct data
		var decodedRequestData map[string]string
		if err := json.NewDecoder(in).Decode(&decodedRequestData); err != nil {
			t.Fatalf("failed to decode request data: %v", err)
		}

		if decodedRequestData["hello"] != "world" {
			t.Errorf("unexpected request data: got %v, want %v", decodedRequestData, requestData)
		}

		// Check the result data for the correct data
		if resultData["foo"] != "bar" {
			t.Errorf("unexpected result data: got %v, want %v", resultData, responseData)
		}
	})

	t.Run("encoding error", func(t *testing.T) {
		in := &bytes.Buffer{}
		out := &bytes.Buffer{}

		executor := NewJSONExecutor(in, out)

		// Invalid data that cannot be marshaled to JSON
		requestData := make(chan int)

		var resultData map[string]string
		err := executor.Execute(requestData, &resultData)
		if err == nil {
			t.Fatal("expected an error, but got nil")
		}
	})

	t.Run("decoding error", func(t *testing.T) {
		in := &bytes.Buffer{}
		out := &bytes.Buffer{}

		executor := NewJSONExecutor(in, out)

		requestData := map[string]string{"hello": "world"}

		// Write invalid JSON to the output buffer
		out.WriteString("invalid json")

		var resultData map[string]string
		err := executor.Execute(requestData, &resultData)
		if err == nil {
			t.Fatal("expected an error, but got nil")
		}
	})
}
