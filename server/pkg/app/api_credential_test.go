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
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func setupTestApp() *Application {
	app := NewApplication()
	// Initialize minimal dependencies
	app.Storage = memory.NewStore()
	app.AuthManager = auth.NewManager()
	app.AuthManager.SetStorage(app.Storage)
	app.ServiceRegistry = serviceregistry.New(nil, nil, nil, nil, app.AuthManager)

	return app
}

func TestCredentialCRUD(t *testing.T) {
	app := setupTestApp()
	ctx := context.Background()

	// 1. List (Empty)
	t.Run("list empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/credentials", nil)
		rr := httptest.NewRecorder()
		app.listCredentialsHandler(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.JSONEq(t, "[]", rr.Body.String())
	})

	// 2. Create
	var createdID string
	t.Run("create credential", func(t *testing.T) {
		cred := configv1.Credential_builder{
			Name: proto.String("Test API Key"),
			Authentication: configv1.Authentication_builder{
				ApiKey: configv1.APIKeyAuth_builder{
					ParamName: proto.String("Authorization"),
					Value: configv1.SecretValue_builder{
						PlainText: proto.String("secret"),
					}.Build(),
				}.Build(),
			}.Build(),
		}.Build()
		body, _ := protojson.Marshal(cred)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/credentials", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		app.createCredentialHandler(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var resp configv1.Credential
		err := protojson.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Test API Key", resp.GetName())
		assert.NotEmpty(t, resp.GetId())
		createdID = resp.GetId()
	})

	// 3. Get
	t.Run("get credential", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/credentials/"+createdID, nil)
		rr := httptest.NewRecorder()
		app.getCredentialHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp configv1.Credential
		err := protojson.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, createdID, resp.GetId())
	})

	// 4. Update
	t.Run("update credential", func(t *testing.T) {
		cred := configv1.Credential_builder{
			Id:   proto.String(createdID),
			Name: proto.String("Updated Name"),
		}.Build()
		body, _ := protojson.Marshal(cred)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/credentials/"+createdID, bytes.NewReader(body))
		rr := httptest.NewRecorder()
		app.updateCredentialHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Verify
		saved, err := app.Storage.GetCredential(ctx, createdID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", saved.GetName())
	})

	// 5. Delete
	t.Run("delete credential", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/credentials/"+createdID, nil)
		rr := httptest.NewRecorder()
		app.deleteCredentialHandler(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify gone
		saved, err := app.Storage.GetCredential(ctx, createdID)
		require.NoError(t, err)
		assert.Nil(t, saved)
	})
}

func TestAuthTestEndpoint(t *testing.T) {
	// Allow loopback for this test since httptest.NewServer uses 127.0.0.1
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

	app := setupTestApp()
	ctx := context.Background()

	// Create a mock upstream server
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "Bearer my-token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer upstream.Close()

	// 1. Test with Inline Auth
	t.Run("test with inline user token", func(t *testing.T) {
		userToken := configv1.UserToken_builder{
			AccessToken: proto.String("my-token"),
		}.Build()
		tokenBody, _ := protojson.Marshal(userToken)
		var tokenJSON json.RawMessage = tokenBody

		reqData := map[string]any{
			"target_url": upstream.URL,
			"user_token": tokenJSON,
		}
		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/debug/auth-test", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		app.testAuthHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp TestAuthResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.Status)
		assert.Equal(t, "success", resp.Body)
	})

	// 2. Test with Saved Credential
	t.Run("test with saved credential", func(t *testing.T) {
		// Create credential first
		cred := configv1.Credential_builder{
			Id:   proto.String("cred-test"),
			Name: proto.String("Test Token"),
			Authentication: configv1.Authentication_builder{
				BearerToken: configv1.BearerTokenAuth_builder{
					Token: configv1.SecretValue_builder{
						PlainText: proto.String("my-token"),
					}.Build(),
				}.Build(),
			}.Build(),
		}.Build()
		err := app.Storage.SaveCredential(ctx, cred)
		require.NoError(t, err)

		reqData := TestAuthRequest{
			TargetURL:    upstream.URL,
			CredentialID: "cred-test",
		}
		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/debug/auth-test", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		app.testAuthHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp TestAuthResponse
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.Status)
	})

	// 3. Test Failure
	t.Run("test auth failure", func(t *testing.T) {
		userToken := configv1.UserToken_builder{
			AccessToken: proto.String("wrong-token"),
		}.Build()
		tokenBody, _ := protojson.Marshal(userToken)
		var tokenJSON json.RawMessage = tokenBody

		reqData := map[string]any{
			"target_url": upstream.URL,
			"user_token": tokenJSON,
		}
		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/debug/auth-test", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		app.testAuthHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp TestAuthResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 401, resp.Status)
	})
}

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
		msg := configv1.Credential_builder{Id: proto.String("test-id")}.Build()
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
	cred := configv1.Credential_builder{
		Id:   proto.String("test-cred"),
		Name: proto.String("Test Credential"),
	}.Build()
	require.NoError(t, store.SaveCredential(context.Background(), cred))

	t.Run("ListCredentials", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/credentials", nil)
		w := httptest.NewRecorder()
		app.listCredentialsHandler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var raw []json.RawMessage
		err := json.Unmarshal(w.Body.Bytes(), &raw)
		require.NoError(t, err)

		creds := make([]*configv1.Credential, len(raw))
		for i, r := range raw {
			creds[i] = &configv1.Credential{}
			err = protojson.Unmarshal(r, creds[i])
			require.NoError(t, err)
		}
		assert.Len(t, creds, 1)
		assert.Equal(t, "test-cred", creds[0].GetId())
	})

	t.Run("ListCredentials_MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/credentials", nil)
		w := httptest.NewRecorder()
		app.listCredentialsHandler(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("GetCredential", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/credentials/test-cred", nil)
		w := httptest.NewRecorder()
		app.getCredentialHandler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var c configv1.Credential
		err := protojson.Unmarshal(w.Body.Bytes(), &c)
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
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UpdateCredential", func(t *testing.T) {
		updatedCred := configv1.Credential_builder{
			Id:   proto.String("test-cred"),
			Name: proto.String("Updated Name"),
		}.Build()
		body, _ := protojson.Marshal(updatedCred)
		req := httptest.NewRequest(http.MethodPut, "/credentials/test-cred", bytes.NewReader(body))
		w := httptest.NewRecorder()
		app.updateCredentialHandler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		stored, _ := store.GetCredential(context.Background(), "test-cred")
		assert.Equal(t, "Updated Name", stored.GetName())
	})

	t.Run("UpdateCredential_MismatchID", func(t *testing.T) {
		updatedCred := configv1.Credential_builder{
			Id:   proto.String("other-id"),
			Name: proto.String("Updated Name"),
		}.Build()
		body, _ := protojson.Marshal(updatedCred)
		req := httptest.NewRequest(http.MethodPut, "/credentials/test-cred", bytes.NewReader(body))
		w := httptest.NewRecorder()
		app.updateCredentialHandler(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("UpdateCredential_BadBody", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/credentials/test-cred", bytes.NewReader([]byte("bad json")))
		w := httptest.NewRecorder()
		app.updateCredentialHandler(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
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

func TestHandleCredentials_Security_Redaction(t *testing.T) {
	app := NewApplication()
	store := memory.NewStore()
	app.Storage = store

	// Create credential with sensitive data
	cred := configv1.Credential_builder{
		Id:   proto.String("cred1"),
		Name: proto.String("Test Cred"),
		Authentication: configv1.Authentication_builder{
			BearerToken: configv1.BearerTokenAuth_builder{
				Token: configv1.SecretValue_builder{
					PlainText: proto.String("my-secret-token"),
				}.Build(),
			}.Build(),
		}.Build(),
		Token: configv1.UserToken_builder{
			AccessToken:  proto.String("access-token-123"),
			RefreshToken: proto.String("refresh-token-456"),
		}.Build(),
	}.Build()
	require.NoError(t, store.SaveCredential(context.Background(), cred))

	t.Run("ListCredentials_ShouldNotLeakSecrets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/credentials", nil)
		w := httptest.NewRecorder()
		app.listCredentialsHandler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		body := w.Body.String()

		assert.NotContains(t, body, "my-secret-token", "Bearer token should be redacted")
		assert.NotContains(t, body, "access-token-123", "Access token should be redacted")
		assert.NotContains(t, body, "refresh-token-456", "Refresh token should be redacted")
	})
}

func TestAuthTestEndpoint_SSRF(t *testing.T) {
	// Ensure env vars are unset for this test to enforce strict SSRF protection
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "false")
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "false")
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "false")

	app := setupTestApp()

	// Create a mock upstream server (on 127.0.0.1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer upstream.Close()

	t.Run("should block 127.0.0.1 access by default", func(t *testing.T) {
		reqData := TestAuthRequest{
			TargetURL: upstream.URL, // This is 127.0.0.1
		}
		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/debug/auth-test", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		app.testAuthHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp TestAuthResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		// We expect the request to have failed due to SSRF protection.
		if resp.Status == 200 && resp.Body == "success" {
			t.Logf("VULNERABILITY CONFIRMED: Successfully accessed 127.0.0.1: %s", upstream.URL)
			t.Fail()
		} else {
			assert.NotEmpty(t, resp.Error, "Expected an error message due to blocked connection")
			assert.Contains(t, strings.ToLower(resp.Error), "blocked", "Error message should mention 'blocked'")
		}
	})
}
