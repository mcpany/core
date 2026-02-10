// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleAuthTest_Detailed(t *testing.T) {
	// Setup
	store := memory.NewStore()
	am := auth.NewManager()
	am.SetStorage(store)
	app := &Application{AuthManager: am, Storage: store}

	// Helper to reset mocks
	resetMocks := func() {
		execLookPath = func(file string) (string, error) {
			return "/bin/" + file, nil
		}
		makeHTTPClient = func(timeout time.Duration) *http.Client {
			return &http.Client{Timeout: timeout}
		}
	}
	defer resetMocks()

	// 1. HTTP Tests
	t.Run("HTTP_Success_APIKey_Header", func(t *testing.T) {
		resetMocks()
		// Save Credential
		credID := "cred-api-key"
		loc := configv1.APIKeyAuth_HEADER
		sv := &configv1.SecretValue{}
		sv.SetPlainText("mykey")

		cred := configv1.Credential_builder{
			Id: proto.String(credID),
			Authentication: configv1.Authentication_builder{
				ApiKey: configv1.APIKeyAuth_builder{
					In:                &loc,
					ParamName:         proto.String("X-API-Key"),
					VerificationValue: proto.String("mykey"),
					Value:             sv,
				}.Build(),
			}.Build(),
		}.Build()
		require.NoError(t, store.SaveCredential(context.Background(), cred))

		// Mock HTTP Client
		makeHTTPClient = func(timeout time.Duration) *http.Client {
			return &http.Client{
				Timeout: timeout,
				Transport: &mockTransport{
					RoundTripFunc: func(req *http.Request) (*http.Response, error) {
						// Verify URL and Headers
						if req.URL.String() != "https://example.com/health" {
							return &http.Response{StatusCode: http.StatusNotFound}, nil
						}
						if req.Header.Get("X-API-Key") != "mykey" {
							return &http.Response{StatusCode: http.StatusUnauthorized}, nil
						}
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader("OK")),
						}, nil
					},
				},
			}
		}

		reqBody := AuthTestRequest{
			CredentialID: credID,
			ServiceType:  "HTTP",
			ServiceConfig: map[string]any{
				"http_service": map[string]any{
					"address": "https://example.com/health",
				},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAuthTest().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp AuthTestResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "successful")
	})

	t.Run("HTTP_Success_BearerToken", func(t *testing.T) {
		resetMocks()
		credID := "cred-bearer"
		sv := &configv1.SecretValue{}
		sv.SetPlainText("mytoken")

		cred := configv1.Credential_builder{
			Id: proto.String(credID),
			Authentication: configv1.Authentication_builder{
				BearerToken: configv1.BearerTokenAuth_builder{
					Token: sv,
				}.Build(),
			}.Build(),
		}.Build()
		require.NoError(t, store.SaveCredential(context.Background(), cred))

		makeHTTPClient = func(timeout time.Duration) *http.Client {
			return &http.Client{
				Timeout: timeout,
				Transport: &mockTransport{
					RoundTripFunc: func(req *http.Request) (*http.Response, error) {
						if req.Header.Get("Authorization") != "Bearer mytoken" {
							return &http.Response{StatusCode: http.StatusUnauthorized}, nil
						}
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader("OK")),
						}, nil
					},
				},
			}
		}

		reqBody := AuthTestRequest{
			CredentialID: credID,
			ServiceType:  "HTTP",
			ServiceConfig: map[string]any{
				"http_service": map[string]any{
					"address": "https://example.com",
				},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAuthTest().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp AuthTestResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
	})

	t.Run("HTTP_Error_404", func(t *testing.T) {
		resetMocks()
		makeHTTPClient = func(timeout time.Duration) *http.Client {
			return &http.Client{
				Timeout: timeout,
				Transport: &mockTransport{
					RoundTripFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusNotFound,
							Body:       io.NopCloser(strings.NewReader("Not Found")),
						}, nil
					},
				},
			}
		}

		reqBody := AuthTestRequest{
			ServiceType: "HTTP",
			ServiceConfig: map[string]any{
				"http_service": map[string]any{"address": "https://example.com"},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAuthTest().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code) // Endpoint always returns 200, success field indicates status
		var resp AuthTestResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "404")
	})

	t.Run("HTTP_Network_Error", func(t *testing.T) {
		resetMocks()
		makeHTTPClient = func(timeout time.Duration) *http.Client {
			return &http.Client{
				Timeout: timeout,
				Transport: &mockTransport{
					RoundTripFunc: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("network unreachable")
					},
				},
			}
		}

		reqBody := AuthTestRequest{
			ServiceType: "HTTP",
			ServiceConfig: map[string]any{
				"http_service": map[string]any{"address": "https://example.com"},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAuthTest().ServeHTTP(w, req)

		var resp AuthTestResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "network unreachable")
	})

	t.Run("HTTP_Unsafe_URL", func(t *testing.T) {
		resetMocks()
		reqBody := AuthTestRequest{
			ServiceType: "HTTP",
			ServiceConfig: map[string]any{
				"http_service": map[string]any{
					"address": "http://169.254.169.254/metadata", // Metadata service (unsafe)
				},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAuthTest().ServeHTTP(w, req)

		var resp AuthTestResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "unsafe url")
	})

	// 2. Command Tests
	t.Run("Command_Success", func(t *testing.T) {
		resetMocks()
		// Mock LookPath to succeed
		execLookPath = func(file string) (string, error) {
			if file == "python3" {
				return "/usr/bin/python3", nil
			}
			return "", errors.New("not found")
		}

		reqBody := AuthTestRequest{
			ServiceType: "CMD",
			ServiceConfig: map[string]any{
				"command_line_service": map[string]any{
					"command": "python3 script.py",
				},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAuthTest().ServeHTTP(w, req)

		var resp AuthTestResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
	})

	t.Run("Command_Failure_NotFound", func(t *testing.T) {
		resetMocks()
		// Mock LookPath to fail
		execLookPath = func(file string) (string, error) {
			return "", errors.New("executable file not found in $PATH")
		}

		reqBody := AuthTestRequest{
			ServiceType: "CMD",
			ServiceConfig: map[string]any{
				"command_line_service": map[string]any{
					"command": "nonexistent_cmd",
				},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAuthTest().ServeHTTP(w, req)

		var resp AuthTestResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "not found")
	})

	t.Run("Command_Empty", func(t *testing.T) {
		resetMocks()
		reqBody := AuthTestRequest{
			ServiceType: "CMD",
			ServiceConfig: map[string]any{
				"command_line_service": map[string]any{
					"command": "   ",
				},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAuthTest().ServeHTTP(w, req)

		var resp AuthTestResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "empty command")
	})

	// 3. Edge Cases
	t.Run("Credential_NotFound", func(t *testing.T) {
		resetMocks()
		reqBody := AuthTestRequest{
			CredentialID: "unknown-id",
			ServiceType:  "HTTP",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAuthTest().ServeHTTP(w, req)

		var resp AuthTestResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "Credential not found")
	})

	t.Run("Invalid_Method", func(t *testing.T) {
		resetMocks()
		req := httptest.NewRequest(http.MethodGet, "/auth/test", nil)
		w := httptest.NewRecorder()
		app.handleAuthTest().ServeHTTP(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("Invalid_JSON", func(t *testing.T) {
		resetMocks()
		req := httptest.NewRequest(http.MethodPost, "/auth/test", strings.NewReader("invalid-json"))
		w := httptest.NewRecorder()
		app.handleAuthTest().ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Unknown_ServiceType_With_Credential", func(t *testing.T) {
		resetMocks()
		credID := "cred-generic"
		cred := configv1.Credential_builder{
			Id: proto.String(credID),
		}.Build()
		require.NoError(t, store.SaveCredential(context.Background(), cred))

		reqBody := AuthTestRequest{
			CredentialID: credID,
			ServiceType:  "UNKNOWN",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAuthTest().ServeHTTP(w, req)

		var resp AuthTestResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "successful")
	})
}
