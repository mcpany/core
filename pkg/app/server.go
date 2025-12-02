// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0


package app

import (
	"context"
	"fmt"
	"crypto/subtle"
	"io"
	"net"
	"net/http"
	"strings"
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
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/afero"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

var healthCheckClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

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
	ReloadConfig(fs afero.Fs, configPaths []string) error
}

// Application is the main application struct, holding the dependencies and
// logic for the MCP Any server. It encapsulates the components required to run
// the server, such as the stdio mode handler, and provides the main `Run`
// method that starts the application.
type Application struct {
	runStdioModeFunc func(ctx context.Context, mcpSrv *mcpserver.Server) error
	PromptManager    prompt.PromptManagerInterface
	ToolManager      tool.ToolManagerInterface
	ResourceManager  resource.ResourceManagerInterface
	configFiles      map[string]string
}

// NewApplication creates a new Application with default dependencies.
// It initializes the application with the standard implementation of the stdio
// mode runner, making it ready to be configured and started.
//
// Returns a new instance of the Application, ready to be run.
func NewApplication() *Application {
	busProvider, _ := bus.NewBusProvider(nil)
	return &Application{
		runStdioModeFunc: runStdioMode,
		PromptManager:    prompt.NewPromptManager(),
		ToolManager:      tool.NewToolManager(busProvider),
		ResourceManager:  resource.NewResourceManager(),
		configFiles:      make(map[string]string),
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
	var cfg *config_v1.McpAnyServerConfig
	if len(configPaths) > 0 {
		store := config.NewFileStore(fs, configPaths)
		var err error
		cfg, err = config.LoadServices(store, "server")
		if err != nil {
			return fmt.Errorf("failed to load services from config: %w", err)
		}
	} else {
		cfg = &config_v1.McpAnyServerConfig{}
	}

	busConfig := cfg.GetGlobalSettings().GetMessageBus()
	busProvider, err := bus.NewBusProvider(busConfig)
	if err != nil {
		return fmt.Errorf("failed to create bus provider: %w", err)
	}
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager)
	a.ToolManager = tool.NewToolManager(busProvider)
	a.PromptManager = prompt.NewPromptManager()
	a.ResourceManager = resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(
		upstreamFactory,
		a.ToolManager,
		a.PromptManager,
		a.ResourceManager,
		authManager,
	)

	// New message bus and workers
	upstreamWorker := worker.NewUpstreamWorker(busProvider, a.ToolManager)
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
		a.ToolManager,
		a.PromptManager,
		a.ResourceManager,
		authManager,
		serviceRegistry,
		busProvider,
	)
	if err != nil {
		return fmt.Errorf("failed to create mcp server: %w", err)
	}

	mcpSrv.SetReloadFunc(func() error {
		return a.ReloadConfig(fs, configPaths)
	})

	a.ToolManager.SetMCPServer(mcpSrv)

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
	cachingMiddleware := middleware.NewCachingMiddleware(a.ToolManager)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(a.ToolManager)
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
						return rateLimitMiddleware.Execute(
							ctx,
							executionReq,
							func(ctx context.Context, _ *tool.ExecutionRequest) (any, error) {
								return next(ctx, method, r)
							},
						)
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

	bindAddress := jsonrpcPort
	if cfg.GetGlobalSettings().GetMcpListenAddress() != "" {
		bindAddress = cfg.GetGlobalSettings().GetMcpListenAddress()
	}

	return a.runServerMode(ctx, mcpSrv, busProvider, bindAddress, grpcPort, shutdownTimeout)
}

// ReloadConfig reloads the configuration from the given paths and updates the
// services.
func (a *Application) ReloadConfig(fs afero.Fs, configPaths []string) error {
	log := logging.GetLogger()
	log.Info("Reloading configuration...")
	metrics.IncrCounter([]string{"config", "reload", "total"}, 1)
	store := config.NewFileStore(fs, configPaths)
	cfg, err := config.LoadServices(store, "server")
	if err != nil {
		metrics.IncrCounter([]string{"config", "reload", "errors"}, 1)
		return fmt.Errorf("failed to load services from config: %w", err)
	}

	// Clear existing services
	for _, serviceConfig := range cfg.GetUpstreamServices() {
		a.ToolManager.ClearToolsForService(serviceConfig.GetName())
		a.ResourceManager.ClearResourcesForService(serviceConfig.GetName())
		a.PromptManager.ClearPromptsForService(serviceConfig.GetName())
	}

	if cfg.GetUpstreamServices() != nil {
		for _, serviceConfig := range cfg.GetUpstreamServices() {
			if serviceConfig.GetDisable() {
				log.Info("Skipping disabled service", "service", serviceConfig.GetName())
				continue
			}

			// Reload tools, prompts, and resources
			upstreamFactory := factory.NewUpstreamServiceFactory(pool.NewManager())
			upstream, err := upstreamFactory.NewUpstream(serviceConfig)
			if err != nil {
				log.Error("Failed to get upstream service", "error", err)
				continue
			}
			if upstream != nil {
				_, _, _, err = upstream.Register(context.Background(), serviceConfig, a.ToolManager, a.PromptManager, a.ResourceManager, false)
				if err != nil {
					log.Error("Failed to register upstream service", "error", err)
					continue
				}
			}
		}
	}
	log.Info("Tools", "tools", a.ToolManager.ListTools())
	log.Info("Prompts", "prompts", a.PromptManager.ListPrompts())
	log.Info("Resources", "resources", a.ResourceManager.ListResources())
	return nil
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
// The function constructs the health check URL from the provided address and
// sends an HTTP GET request. It expects a 200 OK status code for a successful
// health check.
//
// Parameters:
//   - out: The writer to which the success message will be written.
//   - addr: The address (host:port) on which the server is running.
//
// Returns nil if the server is healthy (i.e., responds with a 200 OK), or an
// error if the health check fails for any reason (e.g., connection error,
// non-200 status code).
func HealthCheck(out io.Writer, addr string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return HealthCheckWithContext(ctx, out, addr)
}

// HealthCheckWithContext performs a health check against a running server by
// sending an HTTP GET request to its /healthz endpoint. This is useful for
// monitoring and ensuring the server is operational.
//
// The function constructs the health check URL from the provided address and
// sends an HTTP GET request. It expects a 200 OK status code for a successful
// health check.
//
// Parameters:
//   - ctx: The context for managing the health check's lifecycle.
//   - out: The writer to which the success message will be written.
//   - addr: The address (host:port) on which the server is running.
//
// Returns nil if the server is healthy (i.e., responds with a 200 OK), or an
// error if the health check fails for any reason (e.g., connection error,
// non-200 status code).
func HealthCheckWithContext(
	ctx context.Context,
	out io.Writer,
	addr string,
) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("http://%s/healthz", addr),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request for health check: %w", err)
	}

	resp, err := healthCheckClient.Do(req)
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

	fmt.Fprintln(out, "Health check successful: server is running and healthy.")
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
	bindAddress, grpcPort string,
	shutdownTimeout time.Duration,
) error {
	// localCtx is used to manage the lifecycle of the servers started in this function.
	// It's canceled when this function returns, ensuring that all servers are shut down.
	localCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errChan := make(chan error, 2)
	var wg sync.WaitGroup

	httpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return mcpSrv.Server()
	}, nil)

	apiKey := config.GlobalSettings().APIKey()
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if apiKey != "" {
				if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-API-Key")), []byte(apiKey)) != 1 {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}

	mux := http.NewServeMux()
	mux.Handle("/", authMiddleware(httpHandler))
	mux.Handle("/healthz", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})))
	mux.Handle("/metrics", authMiddleware(metrics.Handler()))

	if grpcPort != "" {
		gwmux := runtime.NewServeMux()
		opts := []gogrpc.DialOption{gogrpc.WithTransportCredentials(insecure.NewCredentials())}
		endpoint := grpcPort
		if strings.HasPrefix(endpoint, ":") {
			endpoint = "localhost" + endpoint
		}
		if err := v1.RegisterRegistrationServiceHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
			return fmt.Errorf("failed to register gateway: %w", err)
		}
		mux.Handle("/v1/", authMiddleware(gwmux))
	}

	httpBindAddress := bindAddress
	if httpBindAddress == "" {
		httpBindAddress = "localhost:8070"
	} else if !strings.Contains(httpBindAddress, ":") {
		httpBindAddress = ":" + httpBindAddress
	}

	startHTTPServer(localCtx, &wg, errChan, "MCP Any HTTP", httpBindAddress, mux, shutdownTimeout)

	grpcBindAddress := grpcPort
	if grpcBindAddress != "" {
		if !strings.Contains(grpcBindAddress, ":") {
			grpcBindAddress = ":" + grpcBindAddress
		}
		lis, err := net.Listen("tcp", grpcBindAddress)
		if err != nil {
			errChan <- fmt.Errorf("gRPC server failed to listen: %w", err)
		} else {
			startGrpcServer(
				localCtx,
				&wg,
				errChan,
				"Registration",
				lis,
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
	}

	var startupErr error
	select {
	case err := <-errChan:
		startupErr = fmt.Errorf("failed to start a server: %w", err)
		logging.GetLogger().Error("Server startup failed, initiating shutdown...", "error", startupErr)
		// A server failed to start, so we need to trigger a shutdown of any other
		// servers that may have started successfully.
		cancel()
	case <-localCtx.Done():
		logging.GetLogger().Info("Received shutdown signal, shutting down gracefully...")
	}

	// N.B. We wait for the servers to shut down regardless of whether there was a
	// startup error or a shutdown signal.
	logging.GetLogger().Info("Waiting for HTTP and gRPC servers to shut down...")
	wg.Wait()
	logging.GetLogger().Info("All servers have shut down.")

	return startupErr
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
		serverLog := logging.GetLogger().With("server", name)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			errChan <- fmt.Errorf("[%s] server failed to listen: %w", name, err)
			return
		}
		defer lis.Close()

		serverLog = serverLog.With("port", lis.Addr().String())

		server := &http.Server{
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

		// localCtx is used to signal the shutdown goroutine to exit.
		localCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		shutdownComplete := make(chan struct{})
		go func() {
			defer close(shutdownComplete)
			select {
			case <-ctx.Done():
				// This is the normal shutdown path.
			case <-localCtx.Done():
				// This is the shutdown path for when the server fails to start.
			}
			shutdownCtx, cancelTimeout := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancelTimeout()
			serverLog.Info("Attempting to gracefully shut down server...")
			if err := server.Shutdown(shutdownCtx); err != nil {
				serverLog.Error("Shutdown error", "error", err)
			}
		}()

		serverLog.Info("HTTP server listening")
		if err := server.Serve(lis); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("[%s] server failed: %w", name, err)
			cancel() // Signal shutdown goroutine to exit
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
// lis is the net.Listener for the server.
// register is a function that registers the gRPC services with the server.
func startGrpcServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	errChan chan<- error,
	name string,
	lis net.Listener,
	shutdownTimeout time.Duration,
	register func(*gogrpc.Server),
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverLog := logging.GetLogger().With("server", name)
		registerSafe := func() *gogrpc.Server {
			defer func() {
				if r := recover(); r != nil {
					serverLog.Error("Panic during gRPC service registration", "panic", r)
					errChan <- fmt.Errorf("[%s] panic during gRPC service registration: %v", name, r)
				}
			}()
			grpcServer := gogrpc.NewServer(gogrpc.StatsHandler(&metrics.GrpcStatsHandler{}))
			register(grpcServer)
			reflection.Register(grpcServer)
			return grpcServer
		}

		grpcServer := registerSafe()
		if grpcServer == nil {
			// A panic occurred during registration, and the error has been sent to the channel.
			// We can just return here.
			return
		}

		// localCtx is used to signal the shutdown goroutine to exit.
		localCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		shutdownComplete := make(chan struct{})
		go func() {
			defer close(shutdownComplete)
			select {
			case <-ctx.Done():
				// This is the normal shutdown path.
			case <-localCtx.Done():
				// This is the shutdown path for when the server fails to start.
			}

			serverLog.Info("Attempting to gracefully shut down server...")
			stopped := make(chan struct{})
			go func() {
				defer close(stopped)
				grpcServer.GracefulStop()
			}()

			timer := time.NewTimer(shutdownTimeout)
			defer timer.Stop()
			select {
			case <-stopped:
				// Successful graceful shutdown.
			case <-timer.C:
				// Graceful shutdown timed out.
				serverLog.Warn("Graceful shutdown timed out, forcing stop.")
				grpcServer.Stop()
			}
		}()

		serverLog.Info("gRPC server listening", "port", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil && err != gogrpc.ErrServerStopped {
			errChan <- fmt.Errorf("[%s] server failed to serve: %w", name, err)
			cancel() // Signal shutdown goroutine to exit
		}
		<-shutdownComplete
		serverLog.Info("Server shut down.")
	}()
}
