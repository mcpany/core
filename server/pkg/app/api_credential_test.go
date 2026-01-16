// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		cred := &configv1.Credential{
			Name: proto.String("Test API Key"),
			Authentication: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_ApiKey{
					ApiKey: &configv1.APIKeyAuth{
						ParamName: proto.String("Authorization"),
						Value: &configv1.SecretValue{
							Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
						},
					},
				},
			},
		}
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
		cred := &configv1.Credential{
			Id:   proto.String(createdID),
			Name: proto.String("Updated Name"),
		}
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
		reqData := TestAuthRequest{
			TargetURL: upstream.URL,
			UserToken: &configv1.UserToken{
				AccessToken: proto.String("my-token"),
			},
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
		cred := &configv1.Credential{
			Id:   proto.String("cred-test"),
			Name: proto.String("Test Token"),
			Authentication: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_BearerToken{
					BearerToken: &configv1.BearerTokenAuth{
						Token: &configv1.SecretValue{
							Value: &configv1.SecretValue_PlainText{PlainText: "my-token"},
						},
					},
				},
			},
		}
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
		reqData := TestAuthRequest{
			TargetURL: upstream.URL,
			UserToken: &configv1.UserToken{
				AccessToken: proto.String("wrong-token"),
			},
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

	// 4. Test SSRF Protection (Blocked)
	t.Run("test ssrf protection blocked", func(t *testing.T) {
		// Attempt to access a private IP that should be blocked by NewSafeHTTPClient
		// 169.254.169.254 (Metadata service) should be blocked by default.
		reqData := TestAuthRequest{
			TargetURL: "http://169.254.169.254/latest/meta-data/",
		}
		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/debug/auth-test", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		app.testAuthHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp TestAuthResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		// Expect failure with specific message
		assert.Contains(t, resp.Error, "ssrf attempt blocked")
	})
}
