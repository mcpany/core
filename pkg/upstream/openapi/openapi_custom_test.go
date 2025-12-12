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
