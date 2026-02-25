// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"context"
	"testing"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// MockConn mocks the client.Conn interface (which includes grpc.ClientConnInterface).
type MockConn struct {
	mock.Mock
}

func (m *MockConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	callArgs := m.Called(ctx, method, args, reply, opts)
	return callArgs.Error(0)
}

func (m *MockConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	callArgs := m.Called(ctx, desc, method, opts)
	return callArgs.Get(0).(grpc.ClientStream), callArgs.Error(1)
}

func (m *MockConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConn) GetState() connectivity.State {
	args := m.Called()
	return args.Get(0).(connectivity.State)
}

// MockChecker mocks the health.Checker interface.
type MockChecker struct {
	mock.Mock
}

func (m *MockChecker) Start() {
	m.Called()
}

func (m *MockChecker) Stop() {
	m.Called()
}

func (m *MockChecker) Check(ctx context.Context) health.CheckerResult {
	args := m.Called(ctx)
	return args.Get(0).(health.CheckerResult)
}

func (m *MockChecker) GetRunningPeriodicCheckCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockChecker) IsStarted() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestNewGrpcClientWrapper(t *testing.T) {
	mockConn := new(MockConn)
	mockChecker := new(MockChecker)
	config := &configv1.UpstreamServiceConfig{}

	wrapper := client.NewGrpcClientWrapper(mockConn, config, mockChecker)

	assert.NotNil(t, wrapper)
}

func TestGrpcClientWrapper_IsHealthy_Shutdown(t *testing.T) {
	mockConn := new(MockConn)
	mockChecker := new(MockChecker)
	config := &configv1.UpstreamServiceConfig{}

	wrapper := client.NewGrpcClientWrapper(mockConn, config, mockChecker)

	// Mock GetState to return Shutdown
	mockConn.On("GetState").Return(connectivity.Shutdown)

	isHealthy := wrapper.IsHealthy(context.Background())
	assert.False(t, isHealthy)
	mockConn.AssertExpectations(t)
}

func TestGrpcClientWrapper_IsHealthy_Bufnet(t *testing.T) {
	mockConn := new(MockConn)
	mockChecker := new(MockChecker)

	grpcSvc := &configv1.GrpcUpstreamService{}
	grpcSvc.SetAddress("bufnet")

	config := &configv1.UpstreamServiceConfig{}
	config.SetGrpcService(grpcSvc)

	wrapper := client.NewGrpcClientWrapper(mockConn, config, mockChecker)

	// Mock GetState to return Ready (anything but Shutdown)
	mockConn.On("GetState").Return(connectivity.Ready)

	isHealthy := wrapper.IsHealthy(context.Background())
	assert.True(t, isHealthy)
	mockConn.AssertExpectations(t)
	// Checker should NOT be called
	mockChecker.AssertNotCalled(t, "Check")
}

func TestGrpcClientWrapper_IsHealthy_NoChecker(t *testing.T) {
	mockConn := new(MockConn)
	config := &configv1.UpstreamServiceConfig{}

	// If we pass nil, NewGrpcClientWrapper calls health.NewChecker(config).
	// If config is empty/minimal, health.NewChecker returns nil.
	wrapper := client.NewGrpcClientWrapper(mockConn, config, nil)

	mockConn.On("GetState").Return(connectivity.Ready)

	isHealthy := wrapper.IsHealthy(context.Background())
	assert.True(t, isHealthy)
	mockConn.AssertExpectations(t)
}

func TestGrpcClientWrapper_IsHealthy_CheckerUp(t *testing.T) {
	mockConn := new(MockConn)
	mockChecker := new(MockChecker)

	// Need valid config so it doesn't default to bufnet check logic failure (though bufnet check requires specific address)
	grpcSvc := &configv1.GrpcUpstreamService{}
	grpcSvc.SetAddress("localhost:50051")
	config := &configv1.UpstreamServiceConfig{}
	config.SetGrpcService(grpcSvc)

	wrapper := client.NewGrpcClientWrapper(mockConn, config, mockChecker)

	mockConn.On("GetState").Return(connectivity.Ready)
	mockChecker.On("Check", mock.Anything).Return(health.CheckerResult{Status: health.StatusUp})

	isHealthy := wrapper.IsHealthy(context.Background())
	assert.True(t, isHealthy)
	mockConn.AssertExpectations(t)
	mockChecker.AssertExpectations(t)
}

func TestGrpcClientWrapper_IsHealthy_CheckerDown(t *testing.T) {
	mockConn := new(MockConn)
	mockChecker := new(MockChecker)

	grpcSvc := &configv1.GrpcUpstreamService{}
	grpcSvc.SetAddress("localhost:50051")
	config := &configv1.UpstreamServiceConfig{}
	config.SetGrpcService(grpcSvc)

	wrapper := client.NewGrpcClientWrapper(mockConn, config, mockChecker)

	mockConn.On("GetState").Return(connectivity.Ready)
	mockChecker.On("Check", mock.Anything).Return(health.CheckerResult{Status: health.StatusDown})

	isHealthy := wrapper.IsHealthy(context.Background())
	assert.False(t, isHealthy)
	mockConn.AssertExpectations(t)
	mockChecker.AssertExpectations(t)
}

func TestGrpcClientWrapper_Close(t *testing.T) {
	mockConn := new(MockConn)
	mockChecker := new(MockChecker)
	config := &configv1.UpstreamServiceConfig{}

	wrapper := client.NewGrpcClientWrapper(mockConn, config, mockChecker)

	mockConn.On("Close").Return(nil)

	err := wrapper.Close()
	assert.NoError(t, err)
	mockConn.AssertExpectations(t)
}
