// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestReproRemoteSecretValidation(t *testing.T) {
	// Mock execLookPath to ensure "ls" is always found, preventing interference from other tests
	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		return "/bin/ls", nil
	}

	// Allow loopback secrets for this test
	os.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_SECRETS")

	// Start a test server that returns "invalid-secret"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid-secret"))
	}))
	defer ts.Close()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("test-remote-secret"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_StdioConnection{
							StdioConnection: &configv1.McpStdioConnection{
								Command: proto.String("ls"),
								Env: map[string]*configv1.SecretValue{
									"TEST_REMOTE_KEY": {
										Value: &configv1.SecretValue_RemoteContent{
											RemoteContent: &configv1.RemoteContent{
												HttpUrl: proto.String(ts.URL),
											},
										},
										ValidationRegex: proto.String("^sk-[a-zA-Z0-9]{10}$"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	errs := Validate(context.Background(), config, Server)

	// Expect validation error
	assert.NotEmpty(t, errs, "Validation errors expected")
	if len(errs) > 0 {
		assert.Contains(t, errs[0].Err.Error(), "secret value does not match validation regex")
	}
}
