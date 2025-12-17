// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockRoundTripper is a mock implementation of http.RoundTripper for testing.
type MockRoundTripper struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

// RoundTrip executes the mock RoundTripFunc.
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func TestHTTPClientImpl_Do(t *testing.T) {
	var capturedReq *http.Request
	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
	}

	mockClient := &http.Client{
		Transport: &MockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				capturedReq = req
				return mockResp, nil
			},
		},
	}

	impl := &httpClientImpl{client: mockClient}
	req, err := http.NewRequest("GET", "http://example.com/test", nil)
	assert.NoError(t, err)

	resp, err := impl.Do(req)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	assert.NoError(t, err)
	assert.Same(t, mockResp, resp)
	assert.Same(t, req, capturedReq)
}
