package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
)

func TestHandleSecretDetail(t *testing.T) {
	// Initialize a memory store
	store := memory.NewStore()

	app := &Application{
		configPaths: []string{}, // No actual config paths needed for this mock
	}

	// Create a secret to test with
	secret := &configv1.Secret{}
	secret.SetId("test-secret-id")
	secret.SetName("Test Secret")
	secret.SetValue("super-secret-value")

	err := store.SaveSecret(context.Background(), secret)
	require.NoError(t, err)

	handler := app.handleSecretDetail(store)

	tests := []struct {
		name           string
		method         string
		url            string
		body           interface{}
		expectedStatus int
		verifyResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "Get existing secret",
			method:         http.MethodGet,
			url:            "/secrets/test-secret-id",
			expectedStatus: http.StatusOK,
			verifyResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				var resp struct {
					Id    string `json:"id"`
					Name  string `json:"name"`
					Value string `json:"value"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, "test-secret-id", resp.Id)
				assert.Equal(t, "Test Secret", resp.Name)
				assert.Equal(t, "[REDACTED]", resp.Value)
			},
		},
		{
			name:           "Get non-existent secret",
			method:         http.MethodGet,
			url:            "/secrets/non-existent-id",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Missing ID",
			method:         http.MethodGet,
			url:            "/secrets/",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Method not allowed",
			method:         http.MethodPatch,
			url:            "/secrets/test-secret-id",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Delete existing secret",
			method:         http.MethodDelete,
			url:            "/secrets/test-secret-id",
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var bodyBytes []byte
			if tc.body != nil {
				bodyBytes, _ = json.Marshal(tc.body)
			}
			req := httptest.NewRequest(tc.method, tc.url, bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.verifyResponse != nil {
				tc.verifyResponse(t, w)
			}
		})
	}
}

func TestHandleSecretDetail_Put(t *testing.T) {
	store := memory.NewStore()
	app := &Application{
		configPaths: []string{},
	}
	handler := app.handleSecretDetail(store)

	tests := []struct {
		name           string
		url            string
		body           interface{}
		expectedStatus int
		verifyState    func(t *testing.T, store *memory.Store)
	}{
		{
			name: "Create new secret",
			url:  "/secrets/new-secret",
			body: map[string]string{
				"id":    "new-secret",
				"name":  "New Secret",
				"value": "super-secret-value",
			},
			expectedStatus: http.StatusOK,
			verifyState: func(t *testing.T, store *memory.Store) {
				secret, err := store.GetSecret(context.Background(), "new-secret")
				require.NoError(t, err)
				require.NotNil(t, secret)
				assert.Equal(t, "new-secret", secret.GetId())
				assert.Equal(t, "New Secret", secret.GetName())
				assert.Equal(t, "super-secret-value", secret.GetValue())
			},
		},
		{
			name: "Update secret (missing name uses ID)",
			url:  "/secrets/update-secret",
			body: map[string]string{
				"id":    "update-secret",
				"value": "new-value",
			},
			expectedStatus: http.StatusOK,
			verifyState: func(t *testing.T, store *memory.Store) {
				secret, err := store.GetSecret(context.Background(), "update-secret")
				require.NoError(t, err)
				require.NotNil(t, secret)
				assert.Equal(t, "update-secret", secret.GetId())
				assert.Equal(t, "update-secret", secret.GetName())
				assert.Equal(t, "new-value", secret.GetValue())
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPut, tc.url, bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.verifyState != nil {
				tc.verifyState(t, store)
			}
		})
	}
}

func TestHandleSecretReveal(t *testing.T) {
	store := memory.NewStore()
	app := &Application{
		configPaths: []string{},
	}

	secret := &configv1.Secret{}
	secret.SetId("reveal-secret")
	secret.SetValue("hidden-value")

	err := store.SaveSecret(context.Background(), secret)
	require.NoError(t, err)

	handler := app.handleSecretDetail(store)

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		verifyResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "Reveal existing secret",
			method:         http.MethodPost,
			url:            "/secrets/reveal-secret/reveal",
			expectedStatus: http.StatusOK,
			verifyResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, "hidden-value", resp["value"])
			},
		},
		{
			name:           "Reveal non-existent secret",
			method:         http.MethodPost,
			url:            "/secrets/non-existent/reveal",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Missing ID for reveal",
			method:         http.MethodPost,
			url:            "/secrets//reveal",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Method not allowed for reveal",
			method:         http.MethodGet,
			url:            "/secrets/reveal-secret/reveal",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.verifyResponse != nil {
				tc.verifyResponse(t, w)
			}
		})
	}
}
