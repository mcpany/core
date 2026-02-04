package client_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/protobuf/proto"
)


func TestNewGrpcClientWrapper(t *testing.T) {
	mockConn := new(MockConn)
	config := &configv1.UpstreamServiceConfig{}
	checker := new(MockChecker)

	wrapper := client.NewGrpcClientWrapper(mockConn, config, checker)
	assert.NotNil(t, wrapper)
}

func TestGrpcClientWrapper_IsHealthy(t *testing.T) {
	tests := []struct {
		name          string
		connState     connectivity.State
		address       string
		checkerResult health.AvailabilityStatus
		setupMock     func(*MockConn, *MockChecker)
		expected      bool
	}{
		{
			name:      "Connection Shutdown",
			connState: connectivity.Shutdown,
			setupMock: func(c *MockConn, ch *MockChecker) {
				c.On("GetState").Return(connectivity.Shutdown)
			},
			expected: false,
		},
		{
			name:      "Bufnet Address",
			connState: connectivity.Ready,
			address:   "bufnet",
			setupMock: func(c *MockConn, ch *MockChecker) {
				c.On("GetState").Return(connectivity.Ready)
			},
			expected: true,
		},
		{
			name:      "Healthy Checker",
			connState: connectivity.Ready,
			address:   "localhost:50051",
			setupMock: func(c *MockConn, ch *MockChecker) {
				c.On("GetState").Return(connectivity.Ready)
				ch.On("Check", mock.Anything).Return(health.CheckerResult{Status: health.StatusUp})
			},
			expected: true,
		},
		{
			name:      "Unhealthy Checker",
			connState: connectivity.Ready,
			address:   "localhost:50051",
			setupMock: func(c *MockConn, ch *MockChecker) {
				c.On("GetState").Return(connectivity.Ready)
				ch.On("Check", mock.Anything).Return(health.CheckerResult{Status: health.StatusDown})
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := new(MockConn)
			mockChecker := new(MockChecker)

			// Construct config using the builder pattern or direct struct if possible.
			// The error message suggested fields were unknown, which implies I used wrong names or struct types.
			// Using the builder pattern as seen in health_test.go

			config := configv1.UpstreamServiceConfig_builder{
				GrpcService: configv1.GrpcUpstreamService_builder{
					Address: proto.String(tt.address),
				}.Build(),
			}.Build()

			if tt.setupMock != nil {
				tt.setupMock(mockConn, mockChecker)
			}

			wrapper := client.NewGrpcClientWrapper(mockConn, config, mockChecker)
			assert.Equal(t, tt.expected, wrapper.IsHealthy(context.Background()))

			mockConn.AssertExpectations(t)
			mockChecker.AssertExpectations(t)
		})
	}
}

func TestGrpcClientWrapper_IsHealthy_NoChecker(t *testing.T) {
	mockConn := new(MockConn)
	mockConn.On("GetState").Return(connectivity.Ready)

	config := configv1.UpstreamServiceConfig_builder{
		GrpcService: configv1.GrpcUpstreamService_builder{
			Address: proto.String("localhost:50051"),
		}.Build(),
	}.Build()

	// NewGrpcClientWrapper handles nil checker by creating one.
	// To test the "checker == nil" path in IsHealthy, we need NewGrpcClientWrapper to NOT create a checker.
	// NewChecker returns nil if config is invalid or missing health check config.
	// If we provide a config that yields a nil checker, we can test that path.
	// For GrpcService, if HealthCheck is nil, NewChecker returns nil?
	// Let's check health.go:
	// case configv1.UpstreamServiceConfig_GrpcService_case:
	//		if uc.GetGrpcService().GetHealthCheck() == nil {
	//			return nil
	//		}

	// So if we don't set HealthCheck, checker will be nil.

	wrapper := client.NewGrpcClientWrapper(mockConn, config, nil)
	assert.True(t, wrapper.IsHealthy(context.Background()))
}

func TestGrpcClientWrapper_Close(t *testing.T) {
	mockConn := new(MockConn)
	mockConn.On("Close").Return(errors.New("close error"))

	config := &configv1.UpstreamServiceConfig{}
	checker := new(MockChecker)

	wrapper := client.NewGrpcClientWrapper(mockConn, config, checker)
	err := wrapper.Close()

	assert.EqualError(t, err, "close error")
	mockConn.AssertExpectations(t)
}
