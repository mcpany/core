package webhooks

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
			name: "Convert HTML to Markdown",
			reqBody: map[string]any{
				"tool_name": "browser_action",
				"result": map[string]any{
					"content": "<h1>Hello World</h1><p>This is a test.</p>",
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

				assert.Equal(t, true, data["allowed"])
				repl := data["replacement_object"].(map[string]any)
				assert.Contains(t, repl["content"], "# Hello World")
				assert.Contains(t, repl["content"], "This is a test.")
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
func TestMarkdownHandler_Errors(t *testing.T) {
	handler := &MarkdownHandler{}

	t.Run("Invalid Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/markdown", nil)
		w := httptest.NewRecorder()
		handler.Handle(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("Invalid Body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/markdown", bytes.NewReader([]byte("invalid")))
		w := httptest.NewRecorder()
		handler.Handle(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid Data Type", func(t *testing.T) {
		event := cloudevents.NewEvent()
		event.SetID("id")
		event.SetSource("src")
		event.SetType("type")
		err := event.SetData(cloudevents.ApplicationJSON, "just a string")
		require.NoError(t, err)

		body, _ := json.Marshal(event)
		req := httptest.NewRequest(http.MethodPost, "/markdown", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/cloudevents+json")

		w := httptest.NewRecorder()
		handler.Handle(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTruncateHandler_Errors(t *testing.T) {
	handler := &TruncateHandler{}

	t.Run("Invalid Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/truncate", nil)
		w := httptest.NewRecorder()
		handler.Handle(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestPaginateHandler_Errors(t *testing.T) {
	handler := &PaginateHandler{}

	t.Run("Invalid Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/paginate", nil)
		w := httptest.NewRecorder()
		handler.Handle(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestMarkdownHandler_Complex(t *testing.T) {
	handler := &MarkdownHandler{}
	complexData := map[string]any{
		"list": []any{
			"<h1>Item 1</h1>",
			map[string]any{
				"nested": "<b>Bold</b>",
			},
		},
	}

	event := cloudevents.NewEvent()
	event.SetID("test-id")
	event.SetSource("test-source")
	event.SetType("com.mcpany.tool.call")

	reqData := map[string]any{
		"inputs": complexData,
	}
	err := event.SetData(cloudevents.ApplicationJSON, reqData)
	require.NoError(t, err)

	body, err := json.Marshal(event)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/markdown", bytes.NewReader(body))
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

	repl := respData["replacement_object"].(map[string]any)
	list := repl["list"].([]any)
	assert.Equal(t, "# Item 1", list[0])
	nested := list[1].(map[string]any)
	assert.Equal(t, "**Bold**", nested["nested"])
}

func TestTruncateHandler_InvalidParam(t *testing.T) {
	handler := &TruncateHandler{}

	event := cloudevents.NewEvent()
	event.SetID("test-id")
	event.SetSource("test-source")
	event.SetType("com.mcpany.tool.call")

	// 16 chars
	reqData := map[string]any{
		"inputs": map[string]any{"text": "long string here"},
	}
	err := event.SetData(cloudevents.ApplicationJSON, reqData)
	require.NoError(t, err)

	body, err := json.Marshal(event)
	require.NoError(t, err)

	// Invalid max_chars, should default to 100
	req := httptest.NewRequest(http.MethodPost, "/truncate?max_chars=abc", bytes.NewReader(body))
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
	repl := respData["replacement_object"].(map[string]any)
	// Should NOT be truncated
	assert.Equal(t, "long string here", repl["text"])
}

func TestPaginateHandler_InvalidParam(t *testing.T) {
	handler := &PaginateHandler{}

	event := cloudevents.NewEvent()
	event.SetID("test-id")
	event.SetSource("test-source")
	event.SetType("com.mcpany.tool.call")

	reqData := map[string]any{
		"inputs": map[string]any{"text": "long string here"},
	}
	err := event.SetData(cloudevents.ApplicationJSON, reqData)
	require.NoError(t, err)

	body, err := json.Marshal(event)
	require.NoError(t, err)

	// Invalid page_size, should default to 1000
	req := httptest.NewRequest(http.MethodPost, "/paginate?page_size=abc", bytes.NewReader(body))
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
	repl := respData["replacement_object"].(map[string]any)
	// Should NOT be paginated heavily (Page 1/1)
	expected := "Page 1/1:\nlong string here\n(Total: 16 chars)"
	assert.Equal(t, expected, repl["text"])
}

func TestMarkdownHandler_NonString(t *testing.T) {
	handler := &MarkdownHandler{}
	event := cloudevents.NewEvent()
	event.SetID("id")
	event.SetSource("src")
	event.SetType("type")

	reqData := map[string]any{
		"inputs": map[string]any{"val": 123.0}, // JSON numbers are float64 usually
	}
	err := event.SetData(cloudevents.ApplicationJSON, reqData)
	require.NoError(t, err)

	body, err := json.Marshal(event)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/markdown", bytes.NewReader(body))
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

	repl := respData["replacement_object"].(map[string]any)
	assert.Equal(t, 123.0, repl["val"])
}

func TestMarkdownHandler_NoOp(t *testing.T) {
	handler := &MarkdownHandler{}
	event := cloudevents.NewEvent()
	event.SetID("id")
	event.SetSource("src")
	event.SetType("type")

	reqData := map[string]any{
		"other": "foo",
	}
	err := event.SetData(cloudevents.ApplicationJSON, reqData)
	require.NoError(t, err)

	body, err := json.Marshal(event)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/markdown", bytes.NewReader(body))
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

	assert.True(t, respData["allowed"].(bool))
	assert.Nil(t, respData["replacement_object"])
}

func TestTruncateHandler_Panic(t *testing.T) {
	handler := &TruncateHandler{}

	event := cloudevents.NewEvent()
	event.SetID("test-id")
	event.SetSource("test-source")
	event.SetType("test-type")
	event.SetData(cloudevents.ApplicationJSON, map[string]any{
		"inputs": "some long text that needs truncating",
	})

	body, _ := json.Marshal(event)
	// Even with negative max_chars, it should now default to 1, so no panic.
	req := httptest.NewRequest(http.MethodPost, "/?max_chars=-5", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/cloudevents+json")
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	handler.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

func TestTruncateHandler_UpperBound(t *testing.T) {
	handler := &TruncateHandler{}

	event := cloudevents.NewEvent()
	event.SetID("test-id")
	event.SetSource("test-source")
	event.SetType("test-type")
	event.SetData(cloudevents.ApplicationJSON, map[string]any{
		"inputs": "some text",
	})

	body, _ := json.Marshal(event)
	// max_chars > 100000 -> Should default to 100000
	req := httptest.NewRequest(http.MethodPost, "/?max_chars=200000", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/cloudevents+json")
	w := httptest.NewRecorder()

	handler.Handle(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	// We can't easily check the internal maxChars variable, but we can assume coverage is hit if we execute this path.
}

func TestPaginateHandler_Panic(t *testing.T) {
	handler := &PaginateHandler{}

	event := cloudevents.NewEvent()
	event.SetID("test-id")
	event.SetSource("test-source")
	event.SetType("test-type")
	event.SetData(cloudevents.ApplicationJSON, map[string]any{
		"inputs": "some long text that needs paginating",
	})

	// Case 1: page_size = 0 -> Should now default to 1
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/?page_size=0", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/cloudevents+json")
	w := httptest.NewRecorder()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v", r)
			}
		}()
		handler.Handle(w, req)
	}()

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

func TestPaginateHandler_UpperBound(t *testing.T) {
	handler := &PaginateHandler{}

	event := cloudevents.NewEvent()
	event.SetID("test-id")
	event.SetSource("test-source")
	event.SetType("test-type")
	event.SetData(cloudevents.ApplicationJSON, map[string]any{
		"inputs": "some text",
	})

	body, _ := json.Marshal(event)
	// page_size > 10000 -> Should default to 10000
	req := httptest.NewRequest(http.MethodPost, "/?page_size=20000", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/cloudevents+json")
	w := httptest.NewRecorder()

	handler.Handle(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
