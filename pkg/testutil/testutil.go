// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/pool"
	"github.com/stretchr/testify/require"
)

func NewTestPoolManager(t *testing.T) *pool.Manager {
	t.Helper()
	pm := pool.NewManager()
	httpPool, err := pool.New(
		func(ctx context.Context) (*client.HttpClientWrapper, error) {
			return &client.HttpClientWrapper{Client: &http.Client{Timeout: 5 * time.Second}}, nil
		},
		1,
		10,
		int(1*time.Minute),
		false,
	)
	require.NoError(t, err)
	pm.Register("test-service", httpPool)
	return pm
}

type MockAuthenticator struct {
	AuthenticateFunc func(req *http.Request) error
}

func (m *MockAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}
