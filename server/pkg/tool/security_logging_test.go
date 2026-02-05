// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sensitiveAuthenticator adds a sensitive header
type sensitiveAuthenticator struct{}

func (a *sensitiveAuthenticator) Authenticate(r *http.Request) error {
	r.Header.Set("Authorization", "Bearer sensitive-token-123")
	r.Header.Set("X-Api-Key", "secret-api-key-456")
	return nil
}

var _ auth.UpstreamAuthenticator = &sensitiveAuthenticator{}

func TestHTTPTool_Execute_LogsSensitiveHeaders(t *testing.T) {
	// t.Parallel() removed to avoid race on global logger
	// Reset logger to ensure we can set it to DEBUG
	logging.ForTestsOnlyResetLogger()

	// Create a buffer to capture logs
	var logBuf bytes.Buffer
	logging.Init(slog.LevelDebug, &logBuf)
	defer logging.ForTestsOnlyResetLogger()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	poolManager.Register("test-service", p)

	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: lo.ToPtr("GET " + server.URL),
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", &sensitiveAuthenticator{}, &configv1.HttpCallDefinition{}, nil, nil, "")

	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
	require.NoError(t, err)

	output := logBuf.String()

	// Check if sensitive headers are redacted
	assert.Contains(t, output, "Authorization: [REDACTED]", "Authorization header should be redacted")
	assert.Contains(t, output, "X-Api-Key: [REDACTED]", "X-Api-Key header should be redacted")

	// Ensure actual secrets are NOT present
	assert.NotContains(t, output, "sensitive-token-123", "Actual token should not be present")
	assert.NotContains(t, output, "secret-api-key-456", "Actual API key should not be present")
}
