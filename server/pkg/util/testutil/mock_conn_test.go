// Copyright 2025 Author(s) of MCP Any
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

// Header ...
// Summary: Header
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}

// Trailer ...
// Summary: Trailer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// CloseSend ...
// Summary: CloseSend
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// Context ...
// Summary: Context
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return context.Background()
}

// SendMsg ...
// Summary: SendMsg
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// RecvMsg ...
// Summary: RecvMsg
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// TestMockClientConn ...
// Summary: TestMockClientConn
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
