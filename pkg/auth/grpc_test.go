/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package auth

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPerRPCCredentials(t *testing.T) {
	t.Run("NilAuthenticator", func(t *testing.T) {
		creds := NewPerRPCCredentials(nil)
		assert.Nil(t, creds)
	})

	t.Run("ValidAuthenticator", func(t *testing.T) {
		mockAuth := &MockUpstreamAuthenticator{}
		creds := NewPerRPCCredentials(mockAuth)
		require.NotNil(t, creds)
		_, ok := creds.(*PerRPCCredentials)
		assert.True(t, ok, "Should be of type *PerRPCCredentials")
	})
}

func TestPerRPCCredentials_GetRequestMetadata(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockAuth := &MockUpstreamAuthenticator{
			AuthenticateFunc: func(r *http.Request) error {
				r.Header.Set("Authorization", "Bearer my-secret-token")
				r.Header.Set("X-Another-Header", "value1,value2")
				return nil
			},
		}
		creds := NewPerRPCCredentials(mockAuth)
		require.NotNil(t, creds)

		metadata, err := creds.GetRequestMetadata(ctx)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		assert.Equal(t, "Bearer my-secret-token", metadata["authorization"])
		assert.Equal(t, "value1,value2", metadata["x-another-header"])
	})

	t.Run("NilAuthenticator", func(t *testing.T) {
		creds := &PerRPCCredentials{authenticator: nil}
		metadata, err := creds.GetRequestMetadata(ctx)
		assert.NoError(t, err)
		assert.Nil(t, metadata)
	})

	t.Run("AuthenticatorError", func(t *testing.T) {
		expectedErr := fmt.Errorf("authentication failed")
		mockAuth := &MockUpstreamAuthenticator{
			AuthenticateFunc: func(r *http.Request) error {
				return expectedErr
			},
		}
		creds := NewPerRPCCredentials(mockAuth)
		require.NotNil(t, creds)

		_, err := creds.GetRequestMetadata(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to apply upstream authenticator for grpc")
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestPerRPCCredentials_RequireTransportSecurity(t *testing.T) {
	creds := &PerRPCCredentials{}
	assert.False(t, creds.RequireTransportSecurity(), "RequireTransportSecurity should return false")
}
