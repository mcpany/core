// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"net/http"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestAuthManager_UserManagement(t *testing.T) {
	manager := NewManager()

	users := []*configv1.User{
		{Id: ptr("user1")},
		{Id: ptr("user2")},
	}

	manager.SetUsers(users)

	t.Run("GetUser_Existing", func(t *testing.T) {
		u, ok := manager.GetUser("user1")
		assert.True(t, ok)
		assert.Equal(t, "user1", u.GetId())
	})

	t.Run("GetUser_NonExisting", func(t *testing.T) {
		u, ok := manager.GetUser("user3")
		assert.False(t, ok)
		assert.Nil(t, u)
	})
}

func TestAuthManager_StorageManagement(t *testing.T) {
	manager := NewManager()
	mockStorage := &MockStorage{}
	manager.SetStorage(mockStorage)
}

func TestValidateAuthentication_Extra(t *testing.T) {
	ctx := context.Background()
	req, _ := http.NewRequest("GET", "/", nil)

	t.Run("NilConfig", func(t *testing.T) {
		err := ValidateAuthentication(ctx, nil, req)
		assert.NoError(t, err)
	})

	t.Run("APIKey_Success", func(t *testing.T) {
		loc := configv1.APIKeyAuth_HEADER
		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_ApiKey{
				ApiKey: &configv1.APIKeyAuth{
					ParamName:         ptr("X-API-Key"),
					In:                &loc,
					VerificationValue: ptr("secret"),
				},
			},
		}
		req.Header.Set("X-API-Key", "secret")
		err := ValidateAuthentication(ctx, config, req)
		assert.NoError(t, err)
	})

	t.Run("APIKey_Failure", func(t *testing.T) {
		loc := configv1.APIKeyAuth_HEADER
		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_ApiKey{
				ApiKey: &configv1.APIKeyAuth{
					ParamName:         ptr("X-API-Key"),
					In:                &loc,
					VerificationValue: ptr("secret"),
				},
			},
		}
		req.Header.Set("X-API-Key", "wrong")
		err := ValidateAuthentication(ctx, config, req)
		assert.Error(t, err)
	})

	t.Run("APIKey_InvalidConfig", func(t *testing.T) {
		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_ApiKey{
				ApiKey: &configv1.APIKeyAuth{
					// Missing params
				},
			},
		}
		err := ValidateAuthentication(ctx, config, req)
		assert.Error(t, err)
	})

	t.Run("BasicAuth_Failure", func(t *testing.T) {
		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					Username:     ptr("user"),
					PasswordHash: ptr("invalid_hash"),
				},
			},
		}
		req.SetBasicAuth("user", "pass")
		err := ValidateAuthentication(ctx, config, req)
		assert.Error(t, err)
	})

	t.Run("BasicAuth_InvalidConfig", func(t *testing.T) {
		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					// Missing hash
				},
			},
		}
		err := ValidateAuthentication(ctx, config, req)
		assert.Error(t, err)
	})

	t.Run("TrustedHeader_Success", func(t *testing.T) {
		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_TrustedHeader{
				TrustedHeader: &configv1.TrustedHeaderAuth{
					HeaderName: ptr("X-User"),
				},
			},
		}
		req.Header.Set("X-User", "user")
		err := ValidateAuthentication(ctx, config, req)
		assert.NoError(t, err)
	})

	t.Run("TrustedHeader_InvalidConfig", func(t *testing.T) {
		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_TrustedHeader{
				TrustedHeader: &configv1.TrustedHeaderAuth{
					// Missing header name
				},
			},
		}
		err := ValidateAuthentication(ctx, config, req)
		assert.Error(t, err)
	})

	t.Run("OAuth2_MissingIssuer", func(t *testing.T) {
		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_Oauth2{
				Oauth2: &configv1.OAuth2Auth{
					// Missing IssuerURL
				},
			},
		}
		err := ValidateAuthentication(ctx, config, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing issuer_url")
	})

	t.Run("OIDC_MissingIssuer", func(t *testing.T) {
		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_Oidc{
				Oidc: &configv1.OIDCAuth{
					// Missing Issuer
				},
			},
		}
		err := ValidateAuthentication(ctx, config, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing issuer")
	})
}
