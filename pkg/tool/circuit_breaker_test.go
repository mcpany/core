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

package tool

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/pool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/samber/lo"
	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestHTTPTool_Execute_WithCircuitBreaker(t *testing.T) {
	var requestCount int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		if atomic.LoadInt32(&requestCount) <= 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, err := pool.New(func(ctx context.Context) (*client.HttpClientWrapper, error) {
		return &client.HttpClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 0, true)

	require.NoError(t, err)
	poolManager.Register("test-service-cb", p)

	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: lo.ToPtr("GET " + server.URL),
	}.Build()

	resilience := &configv1.ResilienceConfig{
		CircuitBreaker: &configv1.CircuitBreakerConfig{
			FailureRateThreshold: 0.6,
			OpenDuration:         durationpb.New(2 * time.Second),
			HalfOpenRequests:     2,
		},
	}

	httpTool := NewHTTPTool(mcpTool, poolManager, "test-service-cb", nil, &configv1.HttpCallDefinition{}, resilience)

	// Initial requests to trip the circuit breaker
	for i := 0; i < 3; i++ {
		_, err := httpTool.Execute(context.Background(), &ExecutionRequest{})
		require.Error(t, err)
	}

	// Circuit should be open
	_, err = httpTool.Execute(context.Background(), &ExecutionRequest{})
	require.Error(t, err)
	assert.Equal(t, gobreaker.ErrOpenState, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&requestCount), "Request should not be sent when circuit is open")

	// Wait for the circuit to enter half-open state
	time.Sleep(2 * time.Second)

	// First request in half-open state should succeed and close the circuit
	_, err = httpTool.Execute(context.Background(), &ExecutionRequest{})
	require.NoError(t, err)
	assert.Equal(t, int32(4), atomic.LoadInt32(&requestCount))

	// Circuit should be closed now, subsequent requests should succeed
	_, err = httpTool.Execute(context.Background(), &ExecutionRequest{
		ToolInputs: json.RawMessage(`{"key":"value"}`),
	})
	require.NoError(t, err)
	assert.Equal(t, int32(5), atomic.LoadInt32(&requestCount))
}
