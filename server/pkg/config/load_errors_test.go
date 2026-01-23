// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func strPtrHelper(s string) *string {
	return &s
}

func TestLoadServices_ReturnsConfigValidationError(t *testing.T) {
	// Create a mock store that returns a config with an invalid service
	invalidConfig := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtrHelper("bad-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtrHelper("not-a-url"),
					},
				},
			},
		},
	}

	mockStore := &MockStore{
		Config: invalidConfig,
	}

	_, err := LoadServices(context.Background(), mockStore, "server")

	assert.Error(t, err)
	var validationErr *ValidationError
	assert.True(t, errors.As(err, &validationErr), "Error should be of type ValidationError")
	assert.Contains(t, validationErr.Error(), "Configuration Validation Failed")
	assert.Contains(t, validationErr.Error(), "bad-service")
	assert.Contains(t, validationErr.Error(), "invalid http address")
}

func TestLoadServices_ReturnsConfigValidationError_OnLoadError(t *testing.T) {
	mockStore := &MockStore{
		Err: &ActionableError{
			Err:        errors.New("something went wrong"),
			Suggestion: "do something",
		},
	}

	_, err := LoadServices(context.Background(), mockStore, "server")

	assert.Error(t, err)
	var validationErr *ValidationError
	assert.True(t, errors.As(err, &validationErr), "Error should be of type ValidationError")
	assert.Contains(t, validationErr.Error(), "Configuration Loading Failed")
	assert.Contains(t, validationErr.Error(), "something went wrong")
	assert.Contains(t, validationErr.Error(), "do something")
}
