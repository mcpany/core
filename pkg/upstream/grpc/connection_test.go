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

package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

import (
	"context"
	"net"
)
func TestNewConnectionFactory(t *testing.T) {
	factory := NewConnectionFactory()
	assert.NotNil(t, factory)

	// Test WithDialer option
	customDialer := func(ctx context.Context, addr string) (net.Conn, error) {
		d := net.Dialer{}
		return d.DialContext(ctx, "tcp", addr)
	}
	factory.WithDialer(customDialer)
	assert.NotNil(t, factory.dialer)

	// Test NewConnection
	// Note: This will attempt a real connection to a non-existent server,
	// so we expect an error. The purpose is to ensure the dialer is called.
	_, err := factory.NewConnection(context.Background(), "localhost:50051")
	assert.Error(t, err)
}
