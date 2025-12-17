// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
