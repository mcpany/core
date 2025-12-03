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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONExecutor(t *testing.T) {
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
}

type errWriter struct{}

func (e *errWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrClosedPipe
}

func TestJSONExecutor_Execute_EncodeError(t *testing.T) {
	out := &errWriter{}
	in := &bytes.Buffer{}
	executor := NewJSONExecutor(out, in)
	err := executor.Execute(nil, nil)
	assert.Error(t, err)
}

func TestJSONExecutor_Execute_DecodeError(t *testing.T) {
	out := &bytes.Buffer{}
	in := &bytes.Buffer{}
	_, err := in.WriteString("invalid json")
	assert.NoError(t, err)
	executor := NewJSONExecutor(out, in)
	err = executor.Execute(nil, nil)
	assert.Error(t, err)
}
