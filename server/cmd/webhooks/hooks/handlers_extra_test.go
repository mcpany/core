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
