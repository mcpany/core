package client

import (
	"context"
	"errors"
	"testing"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type mockConn struct {
	state connectivity.State
	err   error
}

func (m *mockConn) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	return nil
}

func (m *mockConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}

func (m *mockConn) Close() error {
	return m.err
}

func (m *mockConn) GetState() connectivity.State {
	return m.state
}

type mockChecker struct {
	status health.AvailabilityStatus
}

func (m *mockChecker) Check(ctx context.Context) health.CheckerResult {
	return health.CheckerResult{Status: m.status}
}

func (m *mockChecker) Start() {}
func (m *mockChecker) Stop() {}

func TestGrpcClientWrapper(t *testing.T) {
	conn := &mockConn{state: connectivity.Ready}
	config := &configv1.UpstreamServiceConfig{}

	checker := &mockChecker{status: health.StatusUp}

	wrapper := NewGrpcClientWrapper(conn, config, checker)

	assert.True(t, wrapper.IsHealthy(context.Background()))
	assert.NoError(t, wrapper.Close())
}

func TestGrpcClientWrapper_Shutdown(t *testing.T) {
	conn := &mockConn{state: connectivity.Shutdown}
	config := &configv1.UpstreamServiceConfig{}

	wrapper := NewGrpcClientWrapper(conn, config, nil)
	assert.False(t, wrapper.IsHealthy(context.Background()))
}

func TestGrpcClientWrapper_Bufnet(t *testing.T) {
	conn := &mockConn{state: connectivity.Ready}
	config := &configv1.UpstreamServiceConfig{
		Service: &configv1.UpstreamServiceConfig_GrpcService{
			GrpcService: &configv1.GrpcServiceDefinition{
				Address: "bufnet",
			},
		},
	}

	wrapper := NewGrpcClientWrapper(conn, config, nil)
	assert.True(t, wrapper.IsHealthy(context.Background()))
}

func TestGrpcClientWrapper_NoChecker(t *testing.T) {
	conn := &mockConn{state: connectivity.Ready}
	config := &configv1.UpstreamServiceConfig{}

	// Create with nil checker, NewGrpcClientWrapper will create one unless it's nil
	wrapper := &GrpcClientWrapper{
	    Conn: conn,
	    config: config,
	    checker: nil,
	}
	assert.True(t, wrapper.IsHealthy(context.Background()))
}

func TestGrpcClientWrapper_CheckerDown(t *testing.T) {
	conn := &mockConn{state: connectivity.Ready}
	config := &configv1.UpstreamServiceConfig{}
	checker := &mockChecker{status: health.StatusDown}

	wrapper := NewGrpcClientWrapper(conn, config, checker)
	assert.False(t, wrapper.IsHealthy(context.Background()))
}

func TestGrpcClientWrapper_CloseError(t *testing.T) {
	expectedErr := errors.New("close error")
	conn := &mockConn{state: connectivity.Ready, err: expectedErr}
	config := &configv1.UpstreamServiceConfig{}
	checker := &mockChecker{status: health.StatusUp}

	wrapper := NewGrpcClientWrapper(conn, config, checker)
	err := wrapper.Close()
	assert.Equal(t, expectedErr, err)
}
