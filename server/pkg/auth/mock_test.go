// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockUpstreamAuthenticator(t *testing.T) {
	t.Run("default_behavior", func(t *testing.T) {
		mock := &MockUpstreamAuthenticator{}
		req, _ := http.NewRequest("GET", "/", nil)
		err := mock.Authenticate(req)
		assert.NoError(t, err)
	})

	t.Run("with_custom_function", func(t *testing.T) {
		expectedErr := errors.New("authentication failed")
		mock := &MockUpstreamAuthenticator{
			AuthenticateFunc: func(_ *http.Request) error {
				return expectedErr
			},
		}
		req, _ := http.NewRequest("GET", "/", nil)
		err := mock.Authenticate(req)
		assert.Equal(t, expectedErr, err)
	})
}
