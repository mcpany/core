package rest_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/api/rest"
	"github.com/stretchr/testify/assert"
)

func TestInteropTaskHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Method Not Allowed",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed",
		},
		{
			name:           "Bad Request - Invalid JSON",
			method:         http.MethodPost,
			body:           "{invalid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body",
		},
		{
			name:           "Success - Valid Framework",
			method:         http.MethodPost,
			body:           map[string]interface{}{"framework": "CrewAI", "intent": "task_delegation"},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "Internal Server Error - Invalid Framework",
			method:         http.MethodPost,
			body:           map[string]interface{}{"framework": "Unknown", "intent": "task_delegation"},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "no adapter registered for framework",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			if strBody, ok := tt.body.(string); ok {
				reqBody = []byte(strBody)
			} else {
				reqBody, _ = json.Marshal(tt.body)
			}

			req, _ := http.NewRequest(tt.method, "/api/v1/interop/task", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(rest.InteropTaskHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}
