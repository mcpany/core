// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/pool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

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

	factory := func(_ context.Context) (*client.GrpcClientWrapper, error) {
		var transportCreds credentials.TransportCredentials
		if mtlsConfig := config.GetUpstreamAuthentication().GetMtls(); mtlsConfig != nil {
			certificate, err := tls.LoadX509KeyPair(mtlsConfig.GetClientCertPath(), mtlsConfig.GetClientKeyPath())
			if err != nil {
				return nil, err
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
		return client.NewGrpcClientWrapper(conn, config), nil
	}
	return pool.New(factory, minSize, maxSize, idleTimeout, disableHealthCheck)
}
