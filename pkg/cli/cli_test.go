
// Copyright 2024 Author of MCP any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// synchronizedBuffer is a thread-safe buffer.
type synchronizedBuffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func (b *synchronizedBuffer) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Read(p)
}

func (b *synchronizedBuffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}

func (b *synchronizedBuffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}

func TestNewJSONExecutor(t *testing.T) {
	executor := NewJSONExecutor(nil, nil)
	assert.NotNil(t, executor)
}

func TestJSONExecutor_Execute(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		// Pipe for the process's stdin
		procInR, procInW := io.Pipe()
		// Pipe for the process's stdout
		procOutR, procOutW := io.Pipe()

		// The executor writes to the process's stdin and reads from the process's stdout
		executor := NewJSONExecutor(procInW, procOutR)

		// Goroutine to simulate the external process
		go func() {
			defer procInR.Close()
			defer procOutW.Close()

			// Read request from stdin
			var req map[string]interface{}
			err := json.NewDecoder(procInR).Decode(&req)
			if err != nil {
				// Can't call t.Fatal from a goroutine, so panic is the simplest way to fail the test
				panic(err)
			}
			assert.Equal(t, "request_value", req["request_key"])

			// Write response to stdout
			respData := map[string]interface{}{"response_key": "response_value"}
			json.NewEncoder(procOutW).Encode(respData)
		}()

		req := map[string]interface{}{"request_key": "request_value"}
		var res map[string]interface{}
		err := executor.Execute(req, &res)

		require.NoError(t, err)
		assert.Equal(t, "response_value", res["response_key"])
	})

	t.Run("invalid json response", func(t *testing.T) {
		var in, out synchronizedBuffer
		executor := NewJSONExecutor(&out, &in)

		// Write invalid JSON to the input buffer, which the executor reads from
		go func() {
			// First, the executor will write the request to `out`, so we need to handle that.
			// We can just ignore it for this test.
			// Then the executor will try to read the response from `in`.
			in.Write([]byte(`{"key": "value"`)) // Malformed JSON
		}()

		time.Sleep(10 * time.Millisecond) // Give the goroutine time to write

		var res map[string]interface{}
		err := executor.Execute(map[string]interface{}{}, &res)
		assert.Error(t, err)
	})

	t.Run("empty input", func(t *testing.T) {
		var in, out synchronizedBuffer
		executor := NewJSONExecutor(&out, &in)

		time.Sleep(10 * time.Millisecond) // Give the goroutine time to write

		var res map[string]interface{}
		err := executor.Execute(map[string]interface{}{}, &res)

		// This should result in an EOF error because the input stream is empty
		assert.Error(t, err)
	})
}
