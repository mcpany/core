// Copyright 2024 Author(s)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/.LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/stats"
)

func TestGrpcStatsHandler_TagRPC(t *testing.T) {
	handler := &GrpcStatsHandler{}
	ctx := context.Background()
	newCtx := handler.TagRPC(ctx, &stats.RPCTagInfo{})
	assert.Equal(t, ctx, newCtx)
}

func TestGrpcStatsHandler_HandleRPC(t *testing.T) {
	handler := &GrpcStatsHandler{}
	// This is a no-op, so we just call it to ensure it doesn't panic
	handler.HandleRPC(context.Background(), &stats.Begin{})
}

func TestGrpcStatsHandler_TagConn(t *testing.T) {
	handler := &GrpcStatsHandler{}
	ctx := context.Background()
	newCtx := handler.TagConn(ctx, &stats.ConnTagInfo{})
	assert.Equal(t, ctx, newCtx)
}

func TestGrpcStatsHandler_HandleConn(t *testing.T) {
	// Initialize a new metrics instance for this test
	m, handler := newTestMetrics(t)
	// Temporarily replace the global metrics instance
	originalMetrics := GlobalMetrics
	GlobalMetrics = m
	defer func() {
		GlobalMetrics = originalMetrics
	}()

	statsHandler := &GrpcStatsHandler{}

	// Test ConnBegin
	statsHandler.HandleConn(context.Background(), &stats.ConnBegin{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	handler.ServeHTTP(rr, req)
	body := rr.Body.String()
	assert.Contains(t, body, "mcpany_grpc_connections_opened_total 1")

	// Test ConnEnd
	statsHandler.HandleConn(context.Background(), &stats.ConnEnd{})
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/metrics", nil)
	handler.ServeHTTP(rr, req)
	body = rr.Body.String()
	assert.Contains(t, body, "mcpany_grpc_connections_closed_total 1")
}
