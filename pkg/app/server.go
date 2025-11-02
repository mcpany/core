/*
 * Copyright 2025 Author(s) of MCP Any
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
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/config"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/mcpany/core/pkg/metrics"
	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream/factory"
	"github.com/mcpany/core/pkg/worker"
	v1 "github.com/mcpany/core/proto/api/v1"
	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Runner defines the interface for running the MCP Any application. It abstracts
// the application's entry point, allowing for different implementations or mocks
// for testing purposes.
type Runner interface {
	// Run starts the MCP Any application with the given context, filesystem, and
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
	Run(
		ctx context.Context,
		fs afero.Fs,
		stdio bool,
		jsonrpcPort, grpcPort string,
		configPaths []string,
		shutdownTimeout time.Duration,
	) error
}

// Application is the main application struct, holding the dependencies and
// logic for the MCP Any server. It encapsulates the components required to run
// the server, such as the stdio mode handler, and provides the main `Run`
// method that starts the application.
type Application struct {
	runStdioModeFunc func(ctx context.Context, mcpSrv *mcpserver.Server) error
}

// NewApplication creates a new Application with default dependencies.
// It initializes the application with the standard implementation of the stdio
// mode runner, making it ready to be configured and started.
//
// Returns a new instance of the Application, ready to be run.
func NewApplication() *Application {
	return &Application{
		runStdioModeFunc: runStdioMode,
	}
}

// Run starts the MCP Any server and all its components. It initializes the core
// services, loads configurations from the provided paths, starts background
// workers for handling service registration and upstream service communication,
// and launches the gRPC and JSON-RPC servers.
//
// The server's lifecycle is managed by the provided context. A graceful
// shutdown is initiated when the context is canceled.
//
// Parameters:
//   - ctx: The context for managing the application's lifecycle.
//   - fs: The filesystem interface for reading configuration files.
//   - stdio: A boolean indicating whether to run in standard I/O mode.
//   - jsonrpcPort: The port for the JSON-RPC server.
//   - grpcPort: The port for the gRPC registration server. An empty string
//     disables the gRPC server.
//   - configPaths: A slice of paths to service configuration files.
//   - shutdownTimeout: The duration to wait for a graceful shutdown before
//     forcing termination.
//
// Returns an error if any part of the startup or execution fails.
func (a *Application) Run(
	ctx context.Context,
	fs afero.Fs,
	stdio bool,
	jsonrpcPort, grpcPort string,
	configPaths []string,
	shutdownTimeout time.Duration,
) error {
	log := logging.GetLogger()
	fs, err := setup(fs)
	if err != nil {
		return fmt.Errorf("failed to setup filesystem: %w", err)
	}

	log.Info("Starting MCP Any Service...")

	// Load initial services from config files
	var cfg *config_v1.McpxServerConfig
	if len(configPaths) > 0 {
		store := config.NewFileStore(fs, configPaths)
		var err error
		cfg, err = config.LoadServices(store)
		if err != nil {
			return fmt.Errorf("failed to load services from config: %w", err)
		}
	} else {
		cfg = &config_v1.McpxServerConfig{}
	}

	busConfig := cfg.GetGlobalSettings().GetMessageBus()
	busProvider, err := bus.NewBusProvider(busConfig)
	if err != nil {
		return fmt.Errorf("failed to create bus provider: %w", err)
	}
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager)
	toolManager := tool.NewToolManager(busProvider)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(
		upstreamFactory,
		toolManager,
		promptManager,
		resourceManager,
		authManager,
	)

	// New message bus and workers
	upstreamWorker := worker.NewUpstreamWorker(busProvider, toolManager)
	registrationWorker := worker.NewServiceRegistrationWorker(busProvider, serviceRegistry)

	// Start background workers
	upstreamWorker.Start(ctx)
	registrationWorker.Start(ctx)

	// If we're using an in-memory bus, start the in-process worker
	if busConfig == nil || busConfig.GetInMemory() != nil {
		workerCfg := &worker.Config{
			MaxWorkers:   10,
			MaxQueueSize: 100,
		}
		inProcessWorker := worker.New(busProvider, workerCfg)
		inProcessWorker.Start(ctx)
		defer inProcessWorker.Stop()
	}

	// Initialize servers with the message bus
	mcpSrv, err := mcpserver.NewServer(
		ctx,
		toolManager,
		promptManager,
		resourceManager,
		authManager,
		serviceRegistry,
		busProvider,
	)
	if err != nil {
		return fmt.Errorf("failed to create mcp server: %w", err)
	}

	toolManager.SetMCPServer(mcpSrv)

	if cfg.GetUpstreamServices() != nil {
		// Publish registration requests to the bus for each service
		registrationBus := bus.GetBus[*bus.ServiceRegistrationRequest](
			busProvider,
			"service_registration_requests",
		)
		for _, serviceConfig := range cfg.GetUpstreamServices() {
			if serviceConfig.GetDisable() {
				log.Info("Skipping disabled service", "service", serviceConfig.GetName())
				continue
			}
			log.Info(
				"Queueing service for registration from config",
				"service",
				serviceConfig.GetName(),
			)
			regReq := &bus.ServiceRegistrationRequest{Config: serviceConfig}
			// We don't need a correlation ID since we are not waiting for a response here
			registrationBus.Publish(ctx, "request", regReq)
		}
	} else {
		log.Info("No services found in config, skipping service registration.")
	}

	mcpSrv.Server().AddReceivingMiddleware(middleware.CORSMiddleware())
	cachingMiddleware := middleware.NewCachingMiddleware(mcpSrv)
	mcpSrv.Server().AddReceivingMiddleware(func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if r, ok := req.(*mcp.CallToolRequest); ok {
				executionReq := &tool.ExecutionRequest{
					ToolName:   r.Params.Name,
					ToolInputs: r.Params.Arguments,
				}
				result, err := cachingMiddleware.Execute(
					ctx,
					executionReq,
					func(ctx context.Context, _ *tool.ExecutionRequest) (any, error) {
						return next(ctx, method, r)
					},
				)
				if err != nil {
					return nil, err
				}
				return result.(*mcp.CallToolResult), nil
			}
			return next(ctx, method, req)
		}
	})
	mcpSrv.Server().AddReceivingMiddleware(middleware.LoggingMiddleware(nil))
	mcpSrv.Server().AddReceivingMiddleware(middleware.AuthMiddleware(mcpSrv.AuthManager()))

	if stdio {
		return a.runStdioModeFunc(ctx, mcpSrv)
	}

	return a.runServerMode(ctx, mcpSrv, busProvider, jsonrpcPort, grpcPort, shutdownTimeout)
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
		log.Error(
			"setup called with nil afero.Fs. This is not allowed; an explicit afero.Fs must be provided.",
		)
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
// HTTP GET request to its /healthz endpoint. This is useful for monitoring and
// ensuring the server is operational.
//
// The function constructs the health check URL from the provided port and sends
// an HTTP GET request. It expects a 200 OK status code for a successful health
// check.
//
// Parameters:
//   - port: The port on which the server is running.
//
// Returns nil if the server is healthy (i.e., responds with a 200 OK), or an
// error if the health check fails for any reason (e.g., connection error,
// non-200 status code).
func HealthCheck(port string) error {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/healthz", port))
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	// We must read the body and close it to ensure the underlying connection can be reused.
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status code: %d", resp.StatusCode)
	}

	fmt.Println("Health check successful: server is running and healthy.")
	return nil
}

// runServerMode runs the server in the standard HTTP and gRPC server mode. It
// starts the HTTP server for JSON-RPC and the gRPC server for service
// registration, and handles graceful shutdown.
//
// ctx is the context for managing the server's lifecycle.
// mcpSrv is the MCP server instance.
// bus is the message bus for inter-component communication.
// jsonrpcPort is the port for the JSON-RPC server.
// grpcPort is the port for the gRPC registration server.
//
// It returns an error if any of the servers fail to start or run.
func (a *Application) runServerMode(
	ctx context.Context,
	mcpSrv *mcpserver.Server,
	bus *bus.BusProvider,
	jsonrpcPort, grpcPort string,
	shutdownTimeout time.Duration,
) error {
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
	mux.Handle("/metrics", metrics.Handler())

	startHTTPServer(ctx, &wg, errChan, "MCP Any HTTP", ":"+jsonrpcPort, mux, shutdownTimeout)

	if grpcPort != "" {
		startGrpcServer(
			ctx,
			&wg,
			errChan,
			"Registration",
			":"+grpcPort,
			shutdownTimeout,
			func(s *gogrpc.Server) {
				registrationServer, err := mcpserver.NewRegistrationServer(bus)
				if err != nil {
					errChan <- fmt.Errorf("failed to create API server: %w", err)
					return
				}
				v1.RegisterRegistrationServiceServer(s, registrationServer)
			},
		)
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
//
// ctx is the context for managing the server's lifecycle.
// wg is a WaitGroup to signal when the server has shut down.
// errChan is a channel for reporting errors during startup.
// name is a descriptive name for the server, used in logging.
// addr is the address and port on which the server will listen.
// handler is the HTTP handler for processing requests.
func startHTTPServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	errChan chan<- error,
	name, addr string,
	handler http.Handler,
	shutdownTimeout time.Duration,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverLog := logging.GetLogger().With("server", name, "port", addr)
		server := &http.Server{
			Addr:    addr,
			Handler: handler,
			BaseContext: func(_ net.Listener) context.Context {
				return ctx
			},
			ConnState: func(conn net.Conn, state http.ConnState) {
				switch state {
				case http.StateNew:
					metrics.IncrCounter([]string{"http", "connections", "opened", "total"}, 1)
				case http.StateClosed:
					metrics.IncrCounter([]string{"http", "connections", "closed", "total"}, 1)
				}
			},
		}

		shutdownComplete := make(chan struct{})
		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()
			serverLog.Info("Attempting to gracefully shut down server...")
			if err := server.Shutdown(shutdownCtx); err != nil {
				serverLog.Error("Shutdown error", "error", err)
			}
			close(shutdownComplete)
		}()

		serverLog.Info("HTTP server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("[%s] server failed: %w", name, err)
			return
		}

		<-shutdownComplete
		serverLog.Info("Server shut down.")
	}()
}

// startGrpcServer starts a gRPC server in a new goroutine. It handles graceful
// shutdown when the context is canceled.
//
// ctx is the context for managing the server's lifecycle.
// wg is a WaitGroup to signal when the server has shut down.
// errChan is a channel for reporting errors during startup.
// name is a descriptive name for the server, used in logging.
// port is the port on which the server will listen.
// register is a function that registers the gRPC services with the server.
func startGrpcServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	errChan chan<- error,
	name, port string,
	shutdownTimeout time.Duration,
	register func(*gogrpc.Server),
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", port)
		if err != nil {
			errChan <- fmt.Errorf("[%s] server failed to listen: %w", name, err)
			return
		}

		serverLog := logging.GetLogger().With("server", name, "port", port)
		grpcServer := gogrpc.NewServer(gogrpc.StatsHandler(&metrics.GrpcStatsHandler{}))
		register(grpcServer)
		reflection.Register(grpcServer)

		shutdownComplete := make(chan struct{})
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
			close(shutdownComplete)
		}()

		serverLog.Info("gRPC server listening")
		if err := grpcServer.Serve(lis); err != nil && err != gogrpc.ErrServerStopped {
			errChan <- fmt.Errorf("[%s] server failed to serve: %w", name, err)
		}
		<-shutdownComplete
		serverLog.Info("Server shut down.")
	}()
}
