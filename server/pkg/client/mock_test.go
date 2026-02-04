package client_test

import (
	"context"

	"github.com/alexliesenfeld/health"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// MockConn mocks the client.Conn interface
type MockConn struct {
	mock.Mock
	grpc.ClientConnInterface
}

func (m *MockConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConn) GetState() connectivity.State {
	args := m.Called()
	return args.Get(0).(connectivity.State)
}

// MockChecker mocks the health.Checker interface
type MockChecker struct {
	mock.Mock
}

func (m *MockChecker) Check(ctx context.Context) health.CheckerResult {
	args := m.Called(ctx)
	return args.Get(0).(health.CheckerResult)
}

func (m *MockChecker) Start() {
	m.Called()
}

func (m *MockChecker) Stop() {
	m.Called()
}

func (m *MockChecker) GetRunningPeriodicCheckCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockChecker) IsStarted() bool {
	args := m.Called()
	return args.Bool(0)
}
