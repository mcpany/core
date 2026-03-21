// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestJSONExecutor(t *testing.T) {
	t.Parallel()
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

type failWriter struct{}

func (w *failWriter) Write(_ []byte) (n int, err error) {
	return 0, errors.New("fail write")
}

func TestJSONExecutor_Errors(t *testing.T) {
	t.Parallel()
	t.Run("Encode Error", func(t *testing.T) {
		t.Parallel()
		exec := NewJSONExecutor(&failWriter{}, &bytes.Buffer{})
		err := exec.Execute("test", nil)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Decode Error", func(t *testing.T) {
		t.Parallel()
		exec := NewJSONExecutor(&bytes.Buffer{}, bytes.NewBufferString("invalid json"))
		var result map[string]string
		err := exec.Execute("test", &result)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
