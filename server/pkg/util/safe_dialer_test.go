package util_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockIPResolver is a mock implementation of util.IPResolver.
type MockIPResolver struct {
	mock.Mock
}

func (m *MockIPResolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	args := m.Called(ctx, network, host)
	return args.Get(0).([]net.IP), args.Error(1)
}

// MockDialer is a mock implementation of util.NetDialer.
type MockDialer struct {
	mock.Mock
}

func (m *MockDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	args := m.Called(ctx, network, address)
	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}
	return val.(net.Conn), args.Error(1)
}

func TestSafeDialer_MultipleIPs_HappyEyeballs(t *testing.T) {
	// Setup
	resolver := new(MockIPResolver)
	dialer := new(MockDialer)

	safeDialer := util.NewSafeDialer()
	safeDialer.Resolver = resolver
	safeDialer.Dialer = dialer

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	host := "example.com"
	port := "80"
	addr := net.JoinHostPort(host, port)

	// IPs: first one is valid public but unreachable, second is valid public and reachable.
	ip1 := net.ParseIP("1.2.3.4")
	ip2 := net.ParseIP("5.6.7.8")
	ips := []net.IP{ip1, ip2}

	resolver.On("LookupIP", ctx, "ip", host).Return(ips, nil)

	// Mock behavior:
	// If dialing ip1, fail.
	// If dialing ip2, succeed.
	dialer.On("DialContext", ctx, "tcp", net.JoinHostPort(ip1.String(), port)).Return(nil, errors.New("connection refused"))
	dialer.On("DialContext", ctx, "tcp", net.JoinHostPort(ip2.String(), port)).Return(&net.TCPConn{}, nil)

	// Execution
	conn, err := safeDialer.DialContext(ctx, "tcp", addr)

	// Verification
	require.NoError(t, err)
	assert.NotNil(t, conn)

	resolver.AssertExpectations(t)
	// We expect calling DialContext for ip1 (fails) AND ip2 (succeeds)
	dialer.AssertCalled(t, "DialContext", ctx, "tcp", net.JoinHostPort(ip1.String(), port))
	dialer.AssertCalled(t, "DialContext", ctx, "tcp", net.JoinHostPort(ip2.String(), port))
}

func TestSafeDialer_AllIPsFail(t *testing.T) {
	// Setup
	resolver := new(MockIPResolver)
	dialer := new(MockDialer)

	safeDialer := util.NewSafeDialer()
	safeDialer.Resolver = resolver
	safeDialer.Dialer = dialer

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	host := "fail.com"
	port := "80"
	addr := net.JoinHostPort(host, port)

	// IPs: all fail
	ip1 := net.ParseIP("1.2.3.4")
	ips := []net.IP{ip1}

	resolver.On("LookupIP", ctx, "ip", host).Return(ips, nil)
	expectedErr := errors.New("timeout")
	dialer.On("DialContext", ctx, "tcp", net.JoinHostPort(ip1.String(), port)).Return(nil, expectedErr)

	// Execution
	conn, err := safeDialer.DialContext(ctx, "tcp", addr)

	// Verification
	require.Error(t, err)
	assert.Nil(t, conn)
	assert.Equal(t, expectedErr, err)
}

func TestSafeDialer_BlocksUnspecified(t *testing.T) {
	// Setup
	resolver := new(MockIPResolver)
	dialer := new(MockDialer)

	safeDialer := util.NewSafeDialer()
	safeDialer.Resolver = resolver
	safeDialer.Dialer = dialer

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	host := "unspecified"
	port := "80"
	addr := net.JoinHostPort(host, port)

	// IPs: ::
	ip := net.ParseIP("::")
	ips := []net.IP{ip}

	resolver.On("LookupIP", ctx, "ip", host).Return(ips, nil)

	// Execution
	conn, err := safeDialer.DialContext(ctx, "tcp", addr)

	// Verification
	require.Error(t, err)
	assert.Nil(t, conn)
	assert.Contains(t, err.Error(), "ssrf attempt blocked")
	// Unspecified IPs (::, 0.0.0.0) resolve to loopback, so they are now blocked by AllowLoopback check
	assert.Contains(t, err.Error(), "loopback ip")
}

func TestSafeDialer_Networks(t *testing.T) {
	// Setup
	resolver := new(MockIPResolver)
	dialer := new(MockDialer)

	safeDialer := util.NewSafeDialer()
	safeDialer.Resolver = resolver
	safeDialer.Dialer = dialer

	ctx := context.Background()
	host := "example.com"
	port := "80"
	addr := net.JoinHostPort(host, port)
	ip := net.ParseIP("1.2.3.4")

	t.Run("tcp4", func(t *testing.T) {
		resolver.On("LookupIP", ctx, "ip4", host).Return([]net.IP{ip}, nil).Once()
		dialer.On("DialContext", ctx, "tcp4", net.JoinHostPort(ip.String(), port)).Return(&net.TCPConn{}, nil).Once()

		conn, err := safeDialer.DialContext(ctx, "tcp4", addr)
		require.NoError(t, err)
		assert.NotNil(t, conn)
	})

	t.Run("tcp6", func(t *testing.T) {
		resolver.On("LookupIP", ctx, "ip6", host).Return([]net.IP{ip}, nil).Once()
		dialer.On("DialContext", ctx, "tcp6", net.JoinHostPort(ip.String(), port)).Return(&net.TCPConn{}, nil).Once()

		conn, err := safeDialer.DialContext(ctx, "tcp6", addr)
		require.NoError(t, err)
		assert.NotNil(t, conn)
	})
}
