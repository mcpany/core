/*
 * Copyright 2025 Author(s) of MCPX
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/mcpxy/mcpx/pkg/auth"
	"github.com/mcpxy/mcpx/pkg/bus"
	"github.com/mcpxy/mcpx/pkg/config"
	"github.com/mcpxy/mcpx/pkg/logging"
	"github.com/mcpxy/mcpx/pkg/mcpserver"
	"github.com/mcpxy/mcpx/pkg/middleware"
	"github.com/mcpxy/mcpx/pkg/pool"
	"github.com/mcpxy/mcpx/pkg/prompt"
	"github.com/mcpxy/mcpx/pkg/resource"
	"github.com/mcpxy/mcpx/pkg/serviceregistry"
	"github.com/mcpxy/mcpx/pkg/tool"
	"github.com/mcpxy/mcpx/pkg/upstream/factory"
	"github.com/mcpxy/mcpx/pkg/worker"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	v1 "github.com/mcpxy/mcpx/proto/api/v1"
)

// ShutdownTimeout is the duration the server will wait for graceful shutdown.
var ShutdownTimeout = 5 * time.Second

// Run starts the MCP-X server and all its components. It initializes the core
// services, loads configurations, starts background workers, and launches the
// gRPC and JSON-RPC servers.
//
// ctx is the context for the entire application.
// fs is the filesystem interface to use.
// stdio specifies whether to run in stdio mode.
// jsonrpcPort is the port for the JSON-RPC server.
// grpcPort is the port for the gRPC server.
// configPaths is a slice of paths to configuration files.
func Run(ctx context.Context, fs afero.Fs, stdio bool, jsonrpcPort, grpcPort string, configPaths []string) error {
	log := logging.GetLogger()
	fs = setup(fs)

	log.Info("Starting MCP-X Service...")

	// Core components
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager)
	toolManager := tool.NewToolManager()
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, promptManager, resourceManager, authManager)

	// New message bus and workers
	busProvider := bus.NewBusProvider()
	upstreamWorker := worker.NewUpstreamWorker(busProvider, toolManager)
	registrationWorker := worker.NewServiceRegistrationWorker(busProvider, serviceRegistry)

	// Start background workers
	upstreamWorker.Start(ctx)
	registrationWorker.Start(ctx)

	// Initialize servers with the message bus
	mcpSrv, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider)
	if err != nil {
		return fmt.Errorf("failed to create mcp server: %w", err)
	}

	toolManager.SetMCPServer(mcpSrv)

	// Load initial services from config files
	if len(configPaths) > 0 {
		store := config.NewFileStore(fs, configPaths)
		cfg, err := config.LoadServices(store)
		if err != nil {
			return fmt.Errorf("failed to load services from config: %w", err)
		}
		// Publish registration requests to the bus for each service
		registrationBus := bus.GetBus[*bus.ServiceRegistrationRequest](busProvider, "service_registration_requests")
		for _, serviceConfig := range cfg.GetUpstreamServices() {
			log.Info("Queueing service for registration from config", "service", serviceConfig.GetName())
			regReq := &bus.ServiceRegistrationRequest{Config: serviceConfig}
			// We don't need a correlation ID since we are not waiting for a response here
			registrationBus.Publish("request", regReq)
		}
	}

	mcpSrv.Server().AddReceivingMiddleware(middleware.CORSMiddleware())
	mcpSrv.Server().AddReceivingMiddleware(middleware.LoggingMiddleware(nil))
	mcpSrv.Server().AddReceivingMiddleware(middleware.AuthMiddleware(mcpSrv.AuthManager()))

	if stdio {
		return runStdioMode(ctx, mcpSrv)
	}

	return runServerMode(ctx, mcpSrv, busProvider, jsonrpcPort, grpcPort)
}

// setup initializes the filesystem for the server. It ensures that a valid
// afero.Fs is available, defaulting to the OS filesystem if nil is provided.
func setup(fs afero.Fs) afero.Fs {
	log := logging.GetLogger()
	if fs == nil {
		log.Warn("run called with nil afero.Fs, defaulting to OS filesystem. This is not recommended for new direct calls; pass afero.NewOsFs() explicitly.")
		fs = afero.NewOsFs()
	}
	return fs
}

// runStdioMode starts the server in standard I/O mode, which is useful for
// debugging and simple, single-client scenarios. It uses the standard input
// and output as the transport layer.
var runStdioMode = func(ctx context.Context, mcpSrv *mcpserver.Server) error {
	log := logging.GetLogger()
	log.Info("Starting in stdio mode")
	return mcpSrv.Server().Run(ctx, &mcp.StdioTransport{})
}

// runServerMode runs the server in the standard HTTP and gRPC server mode.
// It starts the HTTP server for JSON-RPC and the gRPC server for service
// registration, and handles graceful shutdown.
func runServerMode(ctx context.Context, mcpSrv *mcpserver.Server, bus *bus.BusProvider, jsonrpcPort, grpcPort string) error {
	errChan := make(chan error, 2)
	var wg sync.WaitGroup

	httpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return mcpSrv.Server()
	}, nil)
	startHTTPServer(ctx, &wg, errChan, "MCP-X HTTP", ":"+jsonrpcPort, httpHandler)

	if grpcPort != "" {
		startGrpcServer(ctx, &wg, errChan, "Registration", ":"+grpcPort, func(s *gogrpc.Server) {
			registrationServer, err := mcpserver.NewRegistrationServer(bus)
			if err != nil {
				errChan <- fmt.Errorf("failed to create API server: %w", err)
				return
			}
			v1.RegisterRegistrationServiceServer(s, registrationServer)
		})
	}

	select {
	case err := <-errChan:
		return fmt.Errorf("failed to start a server: %w", err)
	case <-ctx.Done():
		logging.GetLogger().Info("Received shutdown signal, shutting down gracefully...")
	}

	logging.GetLogger().Info("Waiting for HTTP and gRPC servers to shut down...")
	wg.Wait()
	logging.GetLogger().Info("All servers have shut down.")

	return nil
}

// startHTTPServer starts an HTTP server in a new goroutine. It handles graceful
// shutdown when the context is canceled.
func startHTTPServer(ctx context.Context, wg *sync.WaitGroup, errChan chan<- error, name, addr string, handler http.Handler) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverLog := logging.GetLogger().With("server", name, "port", addr)
		server := &http.Server{
			Addr:    addr,
			Handler: handler,
		}

		go func() {
			serverLog.Info("HTTP server listening")
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errChan <- fmt.Errorf("[%s] server failed: %w", name, err)
			}
		}()

		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
		defer cancel()
		serverLog.Info("Attempting to gracefully shut down server...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			serverLog.Error("Shutdown error", "error", err)
		}
		serverLog.Info("Server shut down.")
	}()
}

// startGrpcServer starts a gRPC server in a new goroutine. It handles graceful
// shutdown when the context is canceled.
func startGrpcServer(ctx context.Context, wg *sync.WaitGroup, errChan chan<- error, name, port string, register func(*gogrpc.Server)) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", port)
		if err != nil {
			errChan <- fmt.Errorf("[%s] server failed to listen: %w", name, err)
			return
		}
		serverLog := logging.GetLogger().With("server", name, "port", port)

		grpcServer := gogrpc.NewServer()
		register(grpcServer)
		reflection.Register(grpcServer)

		go func() {
			serverLog.Info("gRPC server listening")
			if err := grpcServer.Serve(lis); err != nil && err != gogrpc.ErrServerStopped {
				errChan <- fmt.Errorf("[%s] server failed to serve: %w", name, err)
			}
		}()

		<-ctx.Done()
		serverLog.Info("Attempting to gracefully shut down server...")
		grpcServer.GracefulStop()
		serverLog.Info("Server shut down.")
	}()
}
