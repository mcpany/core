/*
 * Copyright 2025 Author(s) of MCP-XY
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

package grpc

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func startMockGrpcServer(t *testing.T) *bufconn.Listener {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	t.Cleanup(func() {
		s.Stop()
	})
	return lis
}

func TestGrpcPool_New(t *testing.T) {
	lis := startMockGrpcServer(t)
	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return lis.Dial()
	}

	p, err := NewGrpcPool(1, 5, 100, "bufnet", dialer, nil)
	require.NoError(t, err)
	assert.NotNil(t, p)
	defer p.Close()

	assert.Equal(t, 1, p.Len())

	client, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.True(t, client.IsHealthy())

	p.Put(client)
	assert.Equal(t, 1, p.Len())
}
