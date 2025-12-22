package hooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownHandler(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        map[string]any
		expectedStatus int
		verifyResponse func(t *testing.T, respBytes []byte)
	}{
		{
			name: "Valid HTML to Markdown (PreCall)",
			reqBody: map[string]any{
				"tool_name": "get_html",
				"inputs": map[string]any{
					"content": "<h1>Hello</h1>",
				},
			},
			expectedStatus: http.StatusOK,
			verifyResponse: func(t *testing.T, respBytes []byte) {
				event := cloudevents.NewEvent()
				err := json.Unmarshal(respBytes, &event)
				require.NoError(t, err)

				var data map[string]any
				err = event.DataAs(&data)
				require.NoError(t, err)

				assert.True(t, data["allowed"].(bool))
				repl := data["replacement_object"].(map[string]any)
				assert.Equal(t, "# Hello", repl["content"])
			},
		},
		{
			name: "Valid HTML to Markdown (PostCall)",
			reqBody: map[string]any{
				"tool_name": "get_html",
				"result": map[string]any{
					"content": "<h1>Hello</h1>",
				},
			},
			expectedStatus: http.StatusOK,
			verifyResponse: func(t *testing.T, respBytes []byte) {
				event := cloudevents.NewEvent()
				err := json.Unmarshal(respBytes, &event)
				require.NoError(t, err)

				var data map[string]any
				err = event.DataAs(&data)
				require.NoError(t, err)

				assert.True(t, data["allowed"].(bool))
				repl := data["replacement_object"].(map[string]any)
				assert.Equal(t, "# Hello", repl["content"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &MarkdownHandler{}

			event := cloudevents.NewEvent()
			event.SetID("test-id")
			event.SetSource("test-source")
			event.SetType("com.mcpany.tool.call")
			if err := event.SetData(cloudevents.ApplicationJSON, tt.reqBody); err != nil {
				t.Fatalf("Failed to set data: %v", err)
			}

			body, _ := json.Marshal(event)
			req := httptest.NewRequest(http.MethodPost, "/markdown", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/cloudevents+json")

			w := httptest.NewRecorder()
			handler.Handle(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.verifyResponse != nil {
				tt.verifyResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestTruncateHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		reqBody        map[string]any
		expectedStatus int
		verifyResponse func(t *testing.T, respBytes []byte)
	}{
		{
			name:        "Truncate Long String",
			queryParams: "?max_chars=5",
			reqBody: map[string]any{
				"tool_name": "echo",
				"result": map[string]any{
					"text": "Hello World",
				},
			},
			expectedStatus: http.StatusOK,
			verifyResponse: func(t *testing.T, respBytes []byte) {
				event := cloudevents.NewEvent()
				err := json.Unmarshal(respBytes, &event)
				require.NoError(t, err)

				var data map[string]any
				err = event.DataAs(&data)
				require.NoError(t, err)

				repl := data["replacement_object"].(map[string]any)
				assert.Equal(t, "Hello...", repl["text"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &TruncateHandler{}

			event := cloudevents.NewEvent()
			event.SetID("test-id")
			event.SetSource("test-source")
			event.SetType("com.mcpany.tool.call")
			if err := event.SetData(cloudevents.ApplicationJSON, tt.reqBody); err != nil {
				t.Fatalf("Failed to set data: %v", err)
			}

			body, _ := json.Marshal(event)
			req := httptest.NewRequest(http.MethodPost, "/truncate"+tt.queryParams, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/cloudevents+json")

			w := httptest.NewRecorder()
			handler.Handle(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.verifyResponse != nil {
				tt.verifyResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestPaginateHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		reqBody        map[string]any
		expectedStatus int
		verifyResponse func(t *testing.T, respBytes []byte)
	}{
		{
			name:        "Paginate Long String",
			queryParams: "?page_size=5",
			reqBody: map[string]any{
				"tool_name": "echo",
				"result": map[string]any{
					"text": "Hello World",
				},
			},
			expectedStatus: http.StatusOK,
			verifyResponse: func(t *testing.T, respBytes []byte) {
				event := cloudevents.NewEvent()
				err := json.Unmarshal(respBytes, &event)
				require.NoError(t, err)

				var data map[string]any
				err = event.DataAs(&data)
				require.NoError(t, err)

				repl := data["replacement_object"].(map[string]any)
				assert.Equal(t, "Page 1/3:\nHello\n(Total: 11 chars)", repl["text"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &PaginateHandler{}

			event := cloudevents.NewEvent()
			event.SetID("test-id")
			event.SetSource("test-source")
			event.SetType("com.mcpany.tool.call")
			if err := event.SetData(cloudevents.ApplicationJSON, tt.reqBody); err != nil {
				t.Fatalf("Failed to set data: %v", err)
			}

			body, _ := json.Marshal(event)
			req := httptest.NewRequest(http.MethodPost, "/paginate"+tt.queryParams, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/cloudevents+json")

			w := httptest.NewRecorder()
			handler.Handle(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.verifyResponse != nil {
				tt.verifyResponse(t, w.Body.Bytes())
			}
		})
	}
}
