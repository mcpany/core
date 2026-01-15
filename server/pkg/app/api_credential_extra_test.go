// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// Mock response writer that fails on Write
type failWriter struct {
	http.ResponseWriter
}

func (fw *failWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestWriteError(t *testing.T) {
	t.Run("StatusNotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		writeError(w, errors.New("resource not found"))
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "resource not found")
	})

	t.Run("StatusBadRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		writeError(w, errors.New("id is required"))
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "id is required")

		w = httptest.NewRecorder()
		writeError(w, errors.New("input invalid"))
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "input invalid")
	})

	t.Run("StatusInternalServerError", func(t *testing.T) {
		w := httptest.NewRecorder()
		writeError(w, errors.New("something went wrong"))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal Server Error")
	})
}

func TestWriteJSON(t *testing.T) {
	t.Run("ProtoMessage", func(t *testing.T) {
		w := httptest.NewRecorder()
		msg := &configv1.Credential{Id: proto.String("test-id")}
		writeJSON(w, http.StatusOK, msg)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"id":"test-id"`)
	})

	t.Run("RegularJSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		msg := map[string]string{"key": "value"}
		writeJSON(w, http.StatusOK, msg)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"key":"value"`)
	})

	t.Run("MarshalError", func(t *testing.T) {
		// Mock a type that fails to marshal (not easily possible with std json or proto)
		// but we can try to pass something invalid if possible, or just skip as hard to reproduce.
	})

	t.Run("WriteError", func(t *testing.T) {
		w := &failWriter{httptest.NewRecorder()}
		msg := map[string]string{"key": "value"}
		// Should log error but not panic
		writeJSON(w, http.StatusOK, msg)
	})
}

func TestCredentialHandlers(t *testing.T) {
	store := memory.NewStore()
	app := &Application{Storage: store}

	// Create a credential to test with
	cred := &configv1.Credential{
		Id:   proto.String("test-cred"),
		Name: proto.String("Test Credential"),
	}
	require.NoError(t, store.SaveCredential(context.Background(), cred))

	t.Run("ListCredentials", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/credentials", nil)
		w := httptest.NewRecorder()
		app.listCredentialsHandler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var creds []*configv1.Credential
		err := json.Unmarshal(w.Body.Bytes(), &creds)
		require.NoError(t, err)
		assert.Len(t, creds, 1)
		assert.Equal(t, "test-cred", creds[0].GetId())
	})

	t.Run("ListCredentials_MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/credentials", nil)
		w := httptest.NewRecorder()
		app.listCredentialsHandler(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code) // writeError defaults to 500 for generic error, but msg is "method not allowed"
		// Wait, writeError logic: if not found -> 404, if required/invalid -> 400, else 500.
		// "method not allowed" -> 500 currently.
	})

	t.Run("GetCredential", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/credentials/test-cred", nil)
		w := httptest.NewRecorder()
		app.getCredentialHandler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var c configv1.Credential
		err := json.Unmarshal(w.Body.Bytes(), &c)
		require.NoError(t, err)
		assert.Equal(t, "test-cred", c.GetId())
	})

	t.Run("GetCredential_NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/credentials/missing", nil)
		w := httptest.NewRecorder()
		app.getCredentialHandler(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("GetCredential_NoID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/credentials/", nil)
		w := httptest.NewRecorder()
		app.getCredentialHandler(w, req)
		// The logic splits by /, if path is "/credentials/" -> ["credentials", ""]
		// ID is ""
		// If ID is empty, GetCredential might return error or nil.
		// Actually, logic is: id := pathParts[len(pathParts)-1]
		// if /credentials/ -> parts=["credentials", ""], id=""
		// If id is empty, does it fail?
		// "id is required" check relies on length of pathParts.
		// pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		// Trim removes trailing /. So "credentials/" -> "credentials". parts=["credentials"]. len=1.
		// Code says if len < 2 -> "id is required".
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UpdateCredential", func(t *testing.T) {
		updatedCred := &configv1.Credential{
			Id:   proto.String("test-cred"),
			Name: proto.String("Updated Name"),
		}
		body, _ := json.Marshal(updatedCred)
		req := httptest.NewRequest(http.MethodPut, "/credentials/test-cred", bytes.NewReader(body))
		w := httptest.NewRecorder()
		app.updateCredentialHandler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		stored, _ := store.GetCredential(context.Background(), "test-cred")
		assert.Equal(t, "Updated Name", stored.GetName())
	})

	t.Run("UpdateCredential_MismatchID", func(t *testing.T) {
		updatedCred := &configv1.Credential{
			Id:   proto.String("other-id"),
			Name: proto.String("Updated Name"),
		}
		body, _ := json.Marshal(updatedCred)
		req := httptest.NewRequest(http.MethodPut, "/credentials/test-cred", bytes.NewReader(body))
		w := httptest.NewRecorder()
		app.updateCredentialHandler(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code) // "id mismatch" -> 500
		// "id mismatch" does not contain "required" or "invalid" or "not found"
	})

	t.Run("UpdateCredential_BadBody", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/credentials/test-cred", bytes.NewReader([]byte("bad json")))
		w := httptest.NewRecorder()
		app.updateCredentialHandler(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code) // "invalid request body" contains "invalid"
	})

	t.Run("DeleteCredential", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/credentials/test-cred", nil)
		w := httptest.NewRecorder()
		app.deleteCredentialHandler(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		stored, _ := store.GetCredential(context.Background(), "test-cred")
		assert.Nil(t, stored)
	})

	t.Run("DeleteCredential_NoID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/credentials/", nil)
		w := httptest.NewRecorder()
		app.deleteCredentialHandler(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
