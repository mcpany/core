// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTruncateHandler_Comprehensive(t *testing.T) {
	handler := &TruncateHandler{}

	tests := []struct {
		name     string
		input    any
		maxChars int
		expected any
	}{
		{
			name:     "UTF-8 Emoji Truncation Safe",
			input:    "Hello 🌍",
			maxChars: 7,
			expected: "Hello 🌍",
		},
		{
			name:     "UTF-8 Emoji Truncation Cut",
			input:    "Hello 🌍 World",
			maxChars: 7,
			expected: "Hello 🌍...",
		},
		{
			name:     "Asian Characters",
			input:    "こんにちは世界",
			maxChars: 5,
			expected: "こんにちは...",
		},
		{
			name:     "Exact Length",
			input:    "12345",
			maxChars: 5,
			expected: "12345",
		},
		{
			name:     "One Over Length",
			input:    "123456",
			maxChars: 5,
			expected: "12345...",
		},
		{
			name:     "Empty String",
			input:    "",
			maxChars: 5,
			expected: "",
		},
		{
			name:     "Nested Map",
			input:    map[string]any{"a": "long string here", "b": "short"},
			maxChars: 5,
			expected: map[string]any{"a": "long ...", "b": "short"},
		},
		{
			name:     "List of Strings",
			input:    []any{"long string one", "short"},
			maxChars: 5,
			expected: []any{"long ...", "short"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := cloudevents.NewEvent()
			event.SetID("test-id")
			event.SetSource("test-source")
			event.SetType("com.mcpany.tool.call")
			reqData := map[string]any{
				"inputs": tt.input,
			}
			err := event.SetData(cloudevents.ApplicationJSON, reqData)
			require.NoError(t, err)

			body, err := json.Marshal(event)
			require.NoError(t, err)

			url := fmt.Sprintf("/truncate?max_chars=%d", tt.maxChars)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/cloudevents+json")

			w := httptest.NewRecorder()
			handler.Handle(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var respEvent cloudevents.Event
			err = json.Unmarshal(w.Body.Bytes(), &respEvent)
			require.NoError(t, err)

			var respData map[string]any
			err = respEvent.DataAs(&respData)
			require.NoError(t, err)

			repl := respData["replacement_object"]
			assert.Equal(t, tt.expected, repl)
		})
	}
}

func TestPaginateHandler_Comprehensive(t *testing.T) {
	handler := &PaginateHandler{}

	tests := []struct {
		name          string
		input         any
		page          int
		pageSize      int
		checkResponse func(t *testing.T, res any)
	}{
		{
			name:     "Page 1 Standard",
			input:    "1234567890",
			page:     1,
			pageSize: 5,
			checkResponse: func(t *testing.T, res any) {
				s, ok := res.(string)
				require.True(t, ok)
				assert.Contains(t, s, "Page 1/2")
				assert.Contains(t, s, "12345")
			},
		},
		{
			name:     "Page 2 Standard",
			input:    "1234567890",
			page:     2,
			pageSize: 5,
			checkResponse: func(t *testing.T, res any) {
				s, ok := res.(string)
				require.True(t, ok)
				assert.Contains(t, s, "Page 2/2")
				assert.Contains(t, s, "67890")
			},
		},
		{
			name:     "Page Out of Bounds",
			input:    "12345",
			page:     5,
			pageSize: 5,
			checkResponse: func(t *testing.T, res any) {
				s, ok := res.(string)
				require.True(t, ok)
				assert.Contains(t, s, "Page 5 (empty)")
			},
		},
		{
			name:     "UTF-8 Pagination",
			input:    "A🌏B🌑C", // 5 runes
			page:     2,
			pageSize: 1,
			checkResponse: func(t *testing.T, res any) {
				s, ok := res.(string)
				require.True(t, ok)
				assert.Contains(t, s, "Page 2/5")
				assert.Contains(t, s, "🌏")
			},
		},
		{
			name:     "Invalid Page (Negative)",
			input:    "123",
			page:     -1,
			pageSize: 5,
			checkResponse: func(t *testing.T, res any) {
				s, ok := res.(string)
				require.True(t, ok)
				// Should default to Page 1
				assert.Contains(t, s, "Page 1/")
				assert.Contains(t, s, "123")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := cloudevents.NewEvent()
			event.SetID("test-id")
			event.SetSource("test-source")
			event.SetType("com.mcpany.tool.call")
			reqData := map[string]any{
				"inputs": tt.input,
			}
			err := event.SetData(cloudevents.ApplicationJSON, reqData)
			require.NoError(t, err)

			body, err := json.Marshal(event)
			require.NoError(t, err)

			url := fmt.Sprintf("/paginate?page_size=%d&page=%d", tt.pageSize, tt.page)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/cloudevents+json")

			w := httptest.NewRecorder()
			handler.Handle(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var respEvent cloudevents.Event
			err = json.Unmarshal(w.Body.Bytes(), &respEvent)
			require.NoError(t, err)

			var respData map[string]any
			err = respEvent.DataAs(&respData)
			require.NoError(t, err)

			repl := respData["replacement_object"]
			tt.checkResponse(t, repl)
		})
	}
}
