/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law of or agreed to in writing, software
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
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestNewConnectionFactory(t *testing.T) {
	factory := NewConnectionFactory()
	assert.NotNil(t, factory, "NewConnectionFactory should not return nil")
}

func TestWithDialer(t *testing.T) {
	factory := NewConnectionFactory()
	customDialer := func(context.Context, string) (net.Conn, error) {
		return nil, nil
	}
	factory.WithDialer(customDialer)
	assert.NotNil(t, factory.dialer, "WithDialer should set the dialer")
}

func TestNewConnection(t *testing.T) {
	// Test with a real gRPC server
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer s.Stop()

	factory := NewConnectionFactory()
	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	factory.WithDialer(dialer)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := factory.NewConnection(ctx, "bufnet")
	assert.NoError(t, err, "NewConnection should not return an error")
	assert.NotNil(t, conn, "NewConnection should return a non-nil connection")
	_ = conn.Close()
}
