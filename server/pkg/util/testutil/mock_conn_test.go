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
