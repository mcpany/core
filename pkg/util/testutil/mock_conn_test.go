// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type mockClientStream struct {
	grpc.ClientStream
}

func (m *mockClientStream) Header() (metadata.MD, error) {
	return nil, nil
}

func (m *mockClientStream) Trailer() metadata.MD {
	return nil
}

func (m *mockClientStream) CloseSend() error {
	return nil
}

func (m *mockClientStream) Context() context.Context {
	return context.Background()
}

func (m *mockClientStream) SendMsg(_ interface{}) error {
	return nil
}

func (m *mockClientStream) RecvMsg(_ interface{}) error {
	return nil
}

func TestMockClientConn(t *testing.T) {
	mockConn := NewMockClientConn(t)
	assert.NotNil(t, mockConn)

	t.Run("SetClient and NewStream", func(t *testing.T) {
		mockStream := &mockClientStream{}
		mockConn.SetClient("test_method", mockStream)

		stream, err := mockConn.NewStream(context.Background(), nil, "test_method")
		assert.NoError(t, err)
		assert.Equal(t, mockStream, stream)
	})

	t.Run("NewStream without client", func(t *testing.T) {
		stream, err := mockConn.NewStream(context.Background(), nil, "unknown_method")
		assert.NoError(t, err)
		assert.Nil(t, stream)
	})

	t.Run("Invoke", func(t *testing.T) {
		err := mockConn.Invoke(context.Background(), "", nil, nil)
		assert.NoError(t, err)
	})
}
