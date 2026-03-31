package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/src/interop"
	"github.com/stretchr/testify/require"
)

func TestHandleInterop(t *testing.T) {
	appInstance := &Application{}

	handler := appInstance.handleInterop()

	t.Run("InvalidMethod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/interop/task", nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
	})

	t.Run("InvalidBody", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/interop/task", bytes.NewBuffer([]byte("{invalid json}")))
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("ValidRequest_CrewAI", func(t *testing.T) {
		task := interop.Task{
			ID:        "test-task",
			Framework: "CrewAI",
			Intent:    "task_delegation",
			Payload:   map[string]string{"role": "data_analyst"},
		}
		bodyBytes, _ := json.Marshal(task)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/interop/task", bytes.NewBuffer(bodyBytes))
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)

		var res interop.TaskResult
		err := json.NewDecoder(recorder.Body).Decode(&res)
		require.NoError(t, err)
		require.Equal(t, "success", res.Status)
		require.Equal(t, "data_analyst", res.Telemetry["delegated_role"])
	})

	t.Run("UnknownFramework", func(t *testing.T) {
		task := interop.Task{
			ID:        "test-task",
			Framework: "Unknown",
			Intent:    "unknown",
		}
		bodyBytes, _ := json.Marshal(task)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/interop/task", bytes.NewBuffer(bodyBytes))
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusInternalServerError, recorder.Code)
	})
}
