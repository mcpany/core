package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/client"
	healthChecker "github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/validation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// poolWithChecker wraps a pool.Pool and a health.Checker to ensure the checker
// is stopped when the pool is closed.
type poolWithChecker[T pool.ClosableClient] struct {
	pool.Pool[T]
	checker health.Checker
}

// Close stops the health checker and closes the underlying pool.
//
// Returns an error if the operation fails.
func (p *poolWithChecker[T]) Close() error {
	if p.checker != nil {
		p.checker.Stop()
	}
	return p.Pool.Close()
}

// NewGrpcPool creates a new connection pool for gRPC clients. It configures the
// pool with a factory function that establishes new gRPC connections with the
// specified address, dialer, and credentials.
//
// minSize is the initial number of connections to create.
// maxSize is the maximum number of connections the pool can hold.
// idleTimeout is the duration after which an idle connection may be closed (not currently implemented).
// address is the target address of the gRPC service.
// dialer is an optional custom dialer for creating network connections.
// creds is the per-RPC credentials to be used for authentication.
// It returns a new gRPC client pool or an error if the pool cannot be created.
func NewGrpcPool(
	minSize, maxSize int,
	idleTimeout time.Duration,
	dialer func(context.Context, string) (net.Conn, error),
	creds credentials.PerRPCCredentials,
	config *configv1.UpstreamServiceConfig,
	disableHealthCheck bool,
) (pool.Pool[*client.GrpcClientWrapper], error) {
	if config == nil {
		return nil, fmt.Errorf("service config is nil")
	}
	if config.GetGrpcService() == nil {
		return nil, fmt.Errorf("grpc service config is nil")
	}
	if config.GetGrpcService().GetAddress() == "" {
		return nil, fmt.Errorf("grpc service address is empty")
	}

	// Create a shared health checker for all clients in this pool
	checker := healthChecker.NewChecker(config)

	factory := func(_ context.Context) (*client.GrpcClientWrapper, error) {
		var transportCreds credentials.TransportCredentials
		if mtlsConfig := config.GetUpstreamAuth().GetMtls(); mtlsConfig != nil {
			if err := validation.IsSecurePath(mtlsConfig.GetClientCertPath()); err != nil {
				return nil, fmt.Errorf("invalid client certificate path: %w", err)
			}
			if err := validation.IsSecurePath(mtlsConfig.GetClientKeyPath()); err != nil {
				return nil, fmt.Errorf("invalid client key path: %w", err)
			}
			certificate, err := tls.LoadX509KeyPair(mtlsConfig.GetClientCertPath(), mtlsConfig.GetClientKeyPath())
			if err != nil {
				return nil, err
			}

			if err := validation.IsSecurePath(mtlsConfig.GetCaCertPath()); err != nil {
				return nil, fmt.Errorf("invalid CA certificate path: %w", err)
			}
			caCert, err := os.ReadFile(mtlsConfig.GetCaCertPath())
			if err != nil {
				return nil, err
			}

			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			transportCreds = credentials.NewTLS(&tls.Config{
				Certificates: []tls.Certificate{certificate},
				RootCAs:      caCertPool,
				MinVersion:   tls.VersionTLS12,
			})
		} else {
			transportCreds = insecure.NewCredentials()
		}

		opts := []grpc.DialOption{grpc.WithTransportCredentials(transportCreds)}
		if dialer != nil {
			opts = append(opts, grpc.WithContextDialer(dialer))
		}
		if creds != nil {
			opts = append(opts, grpc.WithPerRPCCredentials(creds))
		}
		addr := strings.TrimPrefix(config.GetGrpcService().GetAddress(), "grpc://")

		conn, err := grpc.NewClient(addr, opts...)
		if err != nil {
			return nil, err
		}
		return client.NewGrpcClientWrapper(conn, config, checker), nil
	}

	p, err := pool.New(factory, minSize, minSize, maxSize, idleTimeout, disableHealthCheck)
	if err != nil {
		// Ensure checker is stopped if pool creation fails
		if checker != nil {
			checker.Stop()
		}
		return nil, err
	}

	return &poolWithChecker[*client.GrpcClientWrapper]{
		Pool:    p,
		checker: checker,
	}, nil
}
