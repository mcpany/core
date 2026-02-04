// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewHTTPClientWrapper(t *testing.T) {
	httpClient := &http.Client{}
	config := &configv1.UpstreamServiceConfig{}
	checker := new(MockChecker)

	wrapper := client.NewHTTPClientWrapper(httpClient, config, checker)
	assert.NotNil(t, wrapper)
}

func TestHTTPClientWrapper_IsHealthy(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*MockChecker)
		expected      bool
	}{
		{
			name: "Healthy Checker",
			setupMock: func(ch *MockChecker) {
				ch.On("Check", mock.Anything).Return(health.CheckerResult{Status: health.StatusUp})
			},
			expected: true,
		},
		{
			name: "Unhealthy Checker",
			setupMock: func(ch *MockChecker) {
				ch.On("Check", mock.Anything).Return(health.CheckerResult{Status: health.StatusDown})
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockChecker := new(MockChecker)
			httpClient := &http.Client{}
			config := &configv1.UpstreamServiceConfig{}

			if tt.setupMock != nil {
				tt.setupMock(mockChecker)
			}

			wrapper := client.NewHTTPClientWrapper(httpClient, config, mockChecker)
			assert.Equal(t, tt.expected, wrapper.IsHealthy(context.Background()))

			mockChecker.AssertExpectations(t)
		})
	}
}

func TestHTTPClientWrapper_IsHealthy_NoChecker(t *testing.T) {
	httpClient := &http.Client{}
	config := &configv1.UpstreamServiceConfig{}

	// Pass nil as checker.
	// As before, we need a config that produces nil checker in NewChecker, or rely on passing nil.
	// But NewHTTPClientWrapper will call NewChecker if checker is nil.
	// If config is empty/default, NewChecker returns nil?
	// For HTTP service, if HealthCheck is nil, NewChecker returns nil.

	wrapper := client.NewHTTPClientWrapper(httpClient, config, nil)
	assert.True(t, wrapper.IsHealthy(context.Background()))
}

func TestHTTPClientWrapper_Close(t *testing.T) {
	httpClient := &http.Client{}
	config := &configv1.UpstreamServiceConfig{}
	checker := new(MockChecker)

	wrapper := client.NewHTTPClientWrapper(httpClient, config, checker)
	err := wrapper.Close()
	assert.NoError(t, err)
}
