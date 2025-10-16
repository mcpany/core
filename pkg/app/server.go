/*
 * Copyright 2025 Author(s) of MCP-XY
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

	"github.com/mcpxy/core/pkg/auth"
	"github.com/mcpxy/core/pkg/bus"
	"github.com/mcpxy/core/pkg/config"
	"github.com/mcpxy/core/pkg/logging"
	"github.com/mcpxy/core/pkg/mcpserver"
	"github.com/mcpxy/core/pkg/middleware"
	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/serviceregistry"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/upstream/factory"
	"github.com/mcpxy/core/pkg/worker"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	v1 "github.com/mcpxy/core/proto/api/v1"
)

// Runner defines the interface for running the MCP-XY application.
type Runner interface {
	// Run starts the MCP-XY application with the given context, filesystem, and
	// configuration. It is the primary entry point for the server.
	//
	// ctx is the context that controls the application's lifecycle.
	// fs is the filesystem interface for reading configurations.
	// stdio specifies whether to run in standard I/O mode.
	// jsonrpcPort is the port for the JSON-RPC server.
	// grpcPort is the port for the gRPC registration server.
	// configPaths is a slice of paths to configuration files.
	//
	// It returns an error if the application fails to start or run.
	Run(ctx context.Context, fs afero.Fs, stdio bool, jsonrpcPort, grpcPort string, configPaths []string) error
}

// Application is the main application struct, holding the dependencies and logic
// for the MCP-XY server. It encapsulates the components required to run the
// server, such as the stdio mode handler.
type Application struct {
	runStdioModeFunc func(ctx context.Context, mcpSrv *mcpserver.Server) error
	shutdownTimeout  time.Duration
}

// NewApplication creates a new Application with default dependencies. It
// initializes the application with the standard implementation of the stdio mode
// runner.
//
// Returns a new instance of the Application.
func NewApplication() *Application {
	return &Application{
		runStdioModeFunc: runStdioMode,
		shutdownTimeout:  5 * time.Second,
	}
}

// Run starts the MCP-XY server and all its components. It initializes the core
// services, loads configurations, starts background workers, and launches the
// gRPC and JSON-RPC servers.
//
// ctx is the context for managing the application's lifecycle.
// fs is the filesystem for reading configuration files.
// stdio determines if the server should run in stdio mode.
// jsonrpcPort specifies the port for the JSON-RPC server.
// grpcPort specifies the port for the gRPC registration server.
// configPaths provides a list of paths to service configuration files.
//
// It returns an error if any part of the startup or execution fails.
func (a *Application) Run(ctx context.Context, fs afero.Fs, stdio bool, jsonrpcPort, grpcPort string, configPaths []string) error {
	log := logging.GetLogger()
	fs, err := setup(fs)
	if err != nil {
		return fmt.Errorf("failed to setup filesystem: %w", err)
	}

	log.Info("Starting MCP-XY Service...")

	// Core components
	busProvider := bus.NewBusProvider()
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager)
	toolManager := tool.NewToolManager(busProvider)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, promptManager, resourceManager, authManager)

	// New message bus and workers
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
		return a.runStdioModeFunc(ctx, mcpSrv)
	}

	return runServerMode(ctx, mcpSrv, busProvider, jsonrpcPort, grpcPort, a.shutdownTimeout)
}

// setup initializes the filesystem for the server. It ensures that a valid
// afero.Fs is available, returning an error if a nil filesystem is provided.
//
// fs is the filesystem to be validated.
//
// It returns a non-nil afero.Fs or an error if the provided filesystem is nil.
func setup(fs afero.Fs) (afero.Fs, error) {
	log := logging.GetLogger()
	if fs == nil {
		log.Error("setup called with nil afero.Fs. This is not allowed; an explicit afero.Fs must be provided.")
		return nil, fmt.Errorf("filesystem not provided")
	}
	return fs, nil
}

// runStdioMode starts the server in standard I/O mode, which is useful for
// debugging and simple, single-client scenarios. It uses the standard input
// and output as the transport layer.
//
// ctx is the context for managing the server's lifecycle.
// mcpSrv is the MCP server instance to run.
//
// It returns an error if the server fails to run in stdio mode.
func runStdioMode(ctx context.Context, mcpSrv *mcpserver.Server) error {
	log := logging.GetLogger()
	log.Info("Starting in stdio mode")
	return mcpSrv.Server().Run(ctx, &mcp.StdioTransport{})
}

// HealthCheck performs a health check against a running server by sending an
// HTTP GET request to its /healthz endpoint.
//
// port specifies the port on which the server is running.
//
// It returns nil if the server is healthy, or an error if the health check
// fails.
func HealthCheck(port string) error {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/healthz", port))
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status code: %d", resp.StatusCode)
	}

	fmt.Println("Health check successful: server is running and healthy.")
	return nil
}

// runServerMode runs the server in the standard HTTP and gRPC server mode. It
// starts the HTTP server for JSON-RPC and the gRPC server for service
// registration, and handles graceful shutdown.
func runServerMode(ctx context.Context, mcpSrv *mcpserver.Server, bus *bus.BusProvider, jsonrpcPort, grpcPort string, shutdownTimeout time.Duration) error {
	errChan := make(chan error, 2)
	var wg sync.WaitGroup

	httpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return mcpSrv.Server()
	}, nil)

	mux := http.NewServeMux()
	mux.Handle("/", httpHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	startHTTPServer(ctx, &wg, errChan, "MCP-XY HTTP", ":"+jsonrpcPort, mux, shutdownTimeout)

	if grpcPort != "" {
		startGrpcServer(ctx, &wg, errChan, "Registration", ":"+grpcPort, func(s *gogrpc.Server) {
			registrationServer, err := mcpserver.NewRegistrationServer(bus)
			if err != nil {
				errChan <- fmt.Errorf("failed to create API server: %w", err)
				return
			}
			v1.RegisterRegistrationServiceServer(s, registrationServer)
		}, shutdownTimeout)
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
func startHTTPServer(ctx context.Context, wg *sync.WaitGroup, errChan chan<- error, name, addr string, handler http.Handler, shutdownTimeout time.Duration) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverLog := logging.GetLogger().With("server", name, "port", addr)
		server := &http.Server{
			Addr:    addr,
			Handler: handler,
		}

		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()
			serverLog.Info("Attempting to gracefully shut down server...")
			if err := server.Shutdown(shutdownCtx); err != nil {
				serverLog.Error("Shutdown error", "error", err)
			}
		}()

		serverLog.Info("HTTP server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("[%s] server failed: %w", name, err)
		}
		serverLog.Info("Server shut down.")
	}()
}

// startGrpcServer starts a gRPC server in a new goroutine. It handles graceful
// shutdown when the context is canceled.
func startGrpcServer(ctx context.Context, wg *sync.WaitGroup, errChan chan<- error, name, port string, register func(*gogrpc.Server), shutdownTimeout time.Duration) {
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
			<-ctx.Done()
			serverLog.Info("Attempting to gracefully shut down server...")

			stopped := make(chan struct{})
			go func() {
				grpcServer.GracefulStop()
				close(stopped)
			}()

			select {
			case <-time.After(shutdownTimeout):
				serverLog.Warn("Graceful shutdown timed out, forcing stop.")
				grpcServer.Stop()
			case <-stopped:
				serverLog.Info("Server gracefully stopped.")
			}
			shutdownMutex.Unlock()
		}()

		serverLog.Info("gRPC server listening")
		if err := grpcServer.Serve(lis); err != nil && err != gogrpc.ErrServerStopped {
			errChan <- fmt.Errorf("[%s] server failed to serve: %w", name, err)
		}
		serverLog.Info("Server shut down.")
	}()
}
