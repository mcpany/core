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

package agent

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetCompletion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"choices":[{"message":{"content":"mocked completion"}}]}`)
	}))
	defer server.Close()

	originalURL := openRouterAPIURL
	openRouterAPIURL = server.URL
	defer func() { openRouterAPIURL = originalURL }()

	os.Setenv("OPENROUTER_API_KEY", "test-key")
	defer os.Unsetenv("OPENROUTER_API_KEY")

	completion, err := GetCompletion("test prompt", "test-model")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if completion != "mocked completion" {
		t.Errorf("expected completion 'mocked completion', got '%s'", completion)
	}
}
