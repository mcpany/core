// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"net/http"
)

type MockHttpClientWrapper struct {
	Client  *http.Client
	CloseFn func() error
}

func (m *MockHttpClientWrapper) Do(req *http.Request) (*http.Response, error) {
	if m.Client != nil {
		return m.Client.Do(req)
	}
	return nil, nil
}

func (m *MockHttpClientWrapper) Close() error {
	if m.CloseFn != nil {
		return m.CloseFn()
	}
	return nil
}

func (m *MockHttpClientWrapper) IsHealthy(ctx context.Context) bool {
	return true
}
