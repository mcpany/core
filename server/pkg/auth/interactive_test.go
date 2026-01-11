// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestInitiateOAuth(t *testing.T) {
	manager := NewManager()

	t.Run("StorageNotInitialized", func(t *testing.T) {
		_, _, err := manager.InitiateOAuth(context.Background(), "user1", "service1", "", "http://callback")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "storage not initialized")
	})

	t.Run("MissingParams", func(t *testing.T) {
		manager.SetStorage(&MockStorage{})
		_, _, err := manager.InitiateOAuth(context.Background(), "user1", "", "", "http://callback")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "either service_id or credential_id must be provided")
	})

	t.Run("CredentialID_NotFound", func(t *testing.T) {
		mockStorage := &MockStorage{
			GetCredentialFunc: func(ctx context.Context, id string) (*configv1.Credential, error) {
				return nil, nil // Not found
			},
		}
		manager.SetStorage(mockStorage)
		_, _, err := manager.InitiateOAuth(context.Background(), "user1", "", "cred1", "http://callback")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "credential cred1 not found")
	})

	t.Run("CredentialID_NoAuth", func(t *testing.T) {
		mockStorage := &MockStorage{
			GetCredentialFunc: func(ctx context.Context, id string) (*configv1.Credential, error) {
				return &configv1.Credential{Id: ptr(id)}, nil
			},
		}
		manager.SetStorage(mockStorage)
		_, _, err := manager.InitiateOAuth(context.Background(), "user1", "", "cred1", "http://callback")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "credential cred1 has no authentication config")
	})

	t.Run("CredentialID_NoOAuth2", func(t *testing.T) {
		mockStorage := &MockStorage{
			GetCredentialFunc: func(ctx context.Context, id string) (*configv1.Credential, error) {
				return &configv1.Credential{
					Id:             ptr(id),
					Authentication: &configv1.Authentication{},
				}, nil
			},
		}
		manager.SetStorage(mockStorage)
		_, _, err := manager.InitiateOAuth(context.Background(), "user1", "", "cred1", "http://callback")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "credential cred1 is not configured for OAuth2")
	})

	t.Run("CredentialID_Success", func(t *testing.T) {
		mockStorage := &MockStorage{
			GetCredentialFunc: func(ctx context.Context, id string) (*configv1.Credential, error) {
				return &configv1.Credential{
					Id: ptr(id),
					Authentication: &configv1.Authentication{
						AuthMethod: &configv1.Authentication_Oauth2{
							Oauth2: &configv1.OAuth2Auth{
								AuthorizationUrl: ptr("http://auth"),
								TokenUrl:         ptr("http://token"),
								ClientId: &configv1.SecretValue{
									Value: &configv1.SecretValue_PlainText{PlainText: "id"},
								},
							},
						},
					},
				}, nil
			},
		}
		manager.SetStorage(mockStorage)
		url, state, err := manager.InitiateOAuth(context.Background(), "user1", "", "cred1", "http://callback")
		assert.NoError(t, err)
		assert.NotEmpty(t, url)
		assert.NotEmpty(t, state)
	})

	t.Run("ServiceID_NotFound", func(t *testing.T) {
		mockStorage := &MockStorage{
			GetServiceFunc: func(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
				return nil, nil
			},
		}
		manager.SetStorage(mockStorage)
		_, _, err := manager.InitiateOAuth(context.Background(), "user1", "service1", "", "http://callback")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service service1 not found")
	})

	t.Run("ServiceID_NoUpstreamAuth", func(t *testing.T) {
		mockStorage := &MockStorage{
			GetServiceFunc: func(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
				return &configv1.UpstreamServiceConfig{Id: ptr(name)}, nil
			},
		}
		manager.SetStorage(mockStorage)
		_, _, err := manager.InitiateOAuth(context.Background(), "user1", "service1", "", "http://callback")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service service1 has no upstream auth configuration")
	})

	t.Run("ServiceID_Success", func(t *testing.T) {
		mockStorage := &MockStorage{
			GetServiceFunc: func(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
				return &configv1.UpstreamServiceConfig{
					Id: ptr(name),
					UpstreamAuth: &configv1.Authentication{
						AuthMethod: &configv1.Authentication_Oauth2{
							Oauth2: &configv1.OAuth2Auth{
								AuthorizationUrl: ptr("http://auth"),
								TokenUrl:         ptr("http://token"),
								ClientId: &configv1.SecretValue{
									Value: &configv1.SecretValue_PlainText{PlainText: "id"},
								},
							},
						},
					},
				}, nil
			},
		}
		manager.SetStorage(mockStorage)
		url, state, err := manager.InitiateOAuth(context.Background(), "user1", "service1", "", "http://callback")
		assert.NoError(t, err)
		assert.NotEmpty(t, url)
		assert.NotEmpty(t, state)
	})
}

func TestHandleOAuthCallback(t *testing.T) {
	manager := NewManager()

	t.Run("StorageNotInitialized", func(t *testing.T) {
		err := manager.HandleOAuthCallback(context.Background(), "user1", "service1", "", "code", "http://callback")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "storage not initialized")
	})

	t.Run("MissingParams", func(t *testing.T) {
		manager.SetStorage(&MockStorage{})
		err := manager.HandleOAuthCallback(context.Background(), "user1", "", "", "code", "http://callback")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "either service_id or credential_id must be provided")
	})

	t.Run("Service_Success", func(t *testing.T) {
		// Mock token endpoint
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"access_token": "access", "refresh_token": "refresh", "token_type": "Bearer", "expires_in": 3600}`))
		}))
		defer ts.Close()

		savedToken := false
		mockStorage := &MockStorage{
			GetServiceFunc: func(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
				return &configv1.UpstreamServiceConfig{
					Id: ptr(name),
					UpstreamAuth: &configv1.Authentication{
						AuthMethod: &configv1.Authentication_Oauth2{
							Oauth2: &configv1.OAuth2Auth{
								AuthorizationUrl: ptr("http://auth"),
								TokenUrl:         ptr(ts.URL),
								ClientId: &configv1.SecretValue{
									Value: &configv1.SecretValue_PlainText{PlainText: "id"},
								},
								ClientSecret: &configv1.SecretValue{
									Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
								},
							},
						},
					},
				}, nil
			},
			SaveTokenFunc: func(ctx context.Context, token *configv1.UserToken) error {
				savedToken = true
				assert.Equal(t, "user1", token.GetUserId())
				assert.Equal(t, "service1", token.GetServiceId())
				assert.Equal(t, "access", token.GetAccessToken())
				return nil
			},
		}
		manager.SetStorage(mockStorage)

		// We need to inject a context with a custom HTTP client if we want to mock the request perfectly without starting a server,
		// but since oauth2 uses context-aware client, we can register the client in context.
		// However, starting a test server (done above) is cleaner for integration style test.

		ctx := context.WithValue(context.Background(), oauth2.HTTPClient, ts.Client())

		err := manager.HandleOAuthCallback(ctx, "user1", "service1", "", "code", "http://callback")
		assert.NoError(t, err)
		assert.True(t, savedToken)
	})

	t.Run("Credential_Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"access_token": "access", "refresh_token": "refresh", "token_type": "Bearer", "expires_in": 3600}`))
		}))
		defer ts.Close()

		savedCred := false
		mockStorage := &MockStorage{
			GetCredentialFunc: func(ctx context.Context, id string) (*configv1.Credential, error) {
				return &configv1.Credential{
					Id: ptr(id),
					Authentication: &configv1.Authentication{
						AuthMethod: &configv1.Authentication_Oauth2{
							Oauth2: &configv1.OAuth2Auth{
								AuthorizationUrl: ptr("http://auth"),
								TokenUrl:         ptr(ts.URL),
								ClientId: &configv1.SecretValue{
									Value: &configv1.SecretValue_PlainText{PlainText: "id"},
								},
							},
						},
					},
				}, nil
			},
			SaveCredentialFunc: func(ctx context.Context, cred *configv1.Credential) error {
				savedCred = true
				assert.NotNil(t, cred.Token)
				assert.Equal(t, "access", cred.Token.GetAccessToken())
				return nil
			},
		}
		manager.SetStorage(mockStorage)

		ctx := context.WithValue(context.Background(), oauth2.HTTPClient, ts.Client())

		err := manager.HandleOAuthCallback(ctx, "user1", "", "cred1", "code", "http://callback")
		assert.NoError(t, err)
		assert.True(t, savedCred)
	})

	t.Run("TokenExchangeError", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer ts.Close()

		mockStorage := &MockStorage{
			GetServiceFunc: func(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
				return &configv1.UpstreamServiceConfig{
					Id: ptr(name),
					UpstreamAuth: &configv1.Authentication{
						AuthMethod: &configv1.Authentication_Oauth2{
							Oauth2: &configv1.OAuth2Auth{
								TokenUrl: ptr(ts.URL),
								ClientId: &configv1.SecretValue{
									Value: &configv1.SecretValue_PlainText{PlainText: "id"},
								},
							},
						},
					},
				}, nil
			},
		}
		manager.SetStorage(mockStorage)

		ctx := context.WithValue(context.Background(), oauth2.HTTPClient, ts.Client())
		err := manager.HandleOAuthCallback(ctx, "user1", "service1", "", "code", "http://callback")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to exchange code")
	})

	t.Run("SaveTokenError", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"access_token": "access"}`))
		}))
		defer ts.Close()

		mockStorage := &MockStorage{
			GetServiceFunc: func(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
				return &configv1.UpstreamServiceConfig{
					Id: ptr(name),
					UpstreamAuth: &configv1.Authentication{
						AuthMethod: &configv1.Authentication_Oauth2{
							Oauth2: &configv1.OAuth2Auth{
								TokenUrl: ptr(ts.URL),
								ClientId: &configv1.SecretValue{
									Value: &configv1.SecretValue_PlainText{PlainText: "id"},
								},
							},
						},
					},
				}, nil
			},
			SaveTokenFunc: func(ctx context.Context, token *configv1.UserToken) error {
				return errors.New("db error")
			},
		}
		manager.SetStorage(mockStorage)

		ctx := context.WithValue(context.Background(), oauth2.HTTPClient, ts.Client())
		err := manager.HandleOAuthCallback(ctx, "user1", "service1", "", "code", "http://callback")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save token")
	})
}

func TestResolveSecretValue(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "", resolveSecretValue(nil))
	})

	t.Run("PlainText", func(t *testing.T) {
		sv := &configv1.SecretValue{
			Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
		}
		assert.Equal(t, "secret", resolveSecretValue(sv))
	})

	t.Run("EnvVar", func(t *testing.T) {
		// Currently returns empty string as per implementation
		sv := &configv1.SecretValue{
			Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "ENV_VAR"},
		}
		assert.Equal(t, "", resolveSecretValue(sv))
	})
}

func TestPtr(t *testing.T) {
	s := "test"
	p := ptr(s)
	assert.NotNil(t, p)
	assert.Equal(t, s, *p)
}
