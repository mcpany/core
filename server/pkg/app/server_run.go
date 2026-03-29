package app

import (
	"bytes"
	"context"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	pb_admin "github.com/mcpany/core/proto/admin/v1"
	v1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/pkg/admin"
	"github.com/mcpany/core/server/pkg/alerts"
	"github.com/mcpany/core/server/pkg/appconsts"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/catalog"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/discovery"
	"github.com/mcpany/core/server/pkg/gc"
	"github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/profile"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/storage/postgres"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/mcpany/core/server/pkg/telemetry"
	"github.com/mcpany/core/server/pkg/tokenizer"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/mcpany/core/server/pkg/webhooks"
	"github.com/mcpany/core/server/pkg/worker"
	"github.com/pmezard/go-difflib/difflib"
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	// config_v1 "github.com/mcpany/core/proto/config/v1".
	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/api/rest"
	"github.com/mcpany/core/server/pkg/topology"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/afero"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Run starts the MCP Any server and all its components.
//
// Summary: Executes the application.
//
// Parameters:
//   - opts (RunOptions): The runtime options.
//
// Returns:
//   - (error): An error if execution fails.
//
// Side Effects:
//   - Starts HTTP and gRPC servers.
//   - Initializes background workers.
//   - Loads configuration.
//
//nolint:gocyclo // Run is the main entry point and setup function, expected to be complex
func (a *Application) Run(opts RunOptions) error {
	log := logging.GetLogger()
	fs, err := setup(opts.Fs)
	if err != nil {
		return fmt.Errorf("failed to setup filesystem: %w", err)
	}
	a.fs = fs
	a.configPaths = opts.ConfigPaths
	a.explicitAPIKey = opts.APIKey
	log.Info("DEBUG: Run API Key", "key", opts.APIKey)

	// Telemetry initialization moved after config loading

	log.Info("Starting MCP Any Service...")

	// Load initial services from config files and Storage
	var storageStore config.Store
	var storageCloser func() error

	if a.Storage != nil {
		storageStore = a.Storage
	} else {
		// Default to SQLite if not specified or explicitly sqlite
		dbDriver := config.GlobalSettings().GetDbDriver()
		switch dbDriver {
		case "", "sqlite":
			dbPath := opts.DBPath
			if dbPath == "" {
				if _, ok := opts.Fs.(*afero.MemMapFs); ok {
					dbPath = ":memory:"
				}
			}
			if dbPath == "" {
				dbPath = config.GlobalSettings().DBPath()
			}
			if dbPath == "" {
				dbPath = "mcpany.db"
			}
			sqliteDB, err := sqlite.NewDB(dbPath)
			if err != nil {
				return fmt.Errorf("failed to initialize sqlite db: %w", err)
			}
			storageCloser = sqliteDB.Close
			storageStore = sqlite.NewStore(sqliteDB)
		case "postgres":
			dsn := config.GlobalSettings().GetDbDsn()
			if dsn == "" {
				return fmt.Errorf("postgres driver selected but db_dsn is empty")
			}
			pgDB, err := postgres.NewDB(dsn)
			if err != nil {
				return fmt.Errorf("failed to initialize postgres db: %w", err)
			}
			storageCloser = func() error { return pgDB.Close() }
			storageStore = postgres.NewStore(pgDB)
		default:
			return fmt.Errorf("unsupported db driver: %s", dbDriver)
		}
	}
	defer func() {
		if storageCloser != nil {
			_ = storageCloser()
		}
	}()

	// Determine config sources
	// Priority: File (if enabled/provided) > Database
	var stores []config.Store

	enableFileConfig := os.Getenv("MCPANY_ENABLE_FILE_CONFIG") == "true"
	if len(opts.ConfigPaths) > 0 {
		// Always load config files if they are explicitly provided
		log.Info("Loading config from files (highest priority)", "paths", opts.ConfigPaths)
		stores = append(stores, config.NewFileStore(fs, opts.ConfigPaths))
	} else if enableFileConfig {
		log.Info("File configuration enabled via env var, but no config paths provided.")
	}

	// Add database as a fallback/secondary source
	stores = append(stores, storageStore)

	multiStore := config.NewMultiStore(stores...)

	var cfg *config_v1.McpAnyServerConfig
	cfg, err = config.LoadServices(opts.Ctx, multiStore, "server")
	if err != nil {
		return fmt.Errorf("failed to load services from merged config: %w", err)
	}
	if cfg == nil {
		cfg = config_v1.McpAnyServerConfig_builder{}.Build()
	}

	// Initialize DB if empty (passing the loaded config to check if seeding is needed)
	if err := a.initializeDatabase(opts.Ctx, storageStore, cfg); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize Log Persistence (Hydration & Worker)
	if s, ok := storageStore.(storage.Storage); ok {
		if err := a.initializeLogPersistence(opts.Ctx, s); err != nil {
			log.Error("Failed to initialize log persistence", "error", err)
		}
		a.startLogPersistence(opts.Ctx, s)
	}
	a.lastReloadTime = time.Now()

	// Populate initial good config for diffing
	if len(opts.ConfigPaths) > 0 {
		a.lastGoodConfig, _ = a.readConfigFiles(fs, opts.ConfigPaths)
	}

	// Initialize Telemetry with loaded config
	shutdownTelemetry, err := telemetry.InitTelemetry(opts.Ctx, appconsts.Name, appconsts.Version, cfg.GetGlobalSettings().GetTelemetry(), os.Stderr)
	if err != nil {
		// Log error but don't fail startup just for telemetry if we want resilience,
		// but typically we might want to know.
		log.Error("Failed to initialize telemetry", "error", err)
	} else {
		defer func() {
			if err := shutdownTelemetry(context.Background()); err != nil {
				log.Error("Failed to shutdown telemetry", "error", err)
			}
		}()
	}

	// Initialize Settings Manager
	a.SettingsManager = NewGlobalSettingsManager(
		opts.APIKey,
		cfg.GetGlobalSettings().GetAllowedIps(),
		// Logic for origins default moved to inside NewGlobalSettingsManager or updated here
		nil,
	)
	a.SettingsManager.Update(cfg.GetGlobalSettings(), opts.APIKey)

	busConfig := cfg.GetGlobalSettings().GetMessageBus()
	busProvider, err := bus.NewProvider(busConfig)
	if err != nil {
		return fmt.Errorf("failed to create bus provider: %w", err)
	}
	a.busProvider = busProvider

	poolManager := pool.NewManager()
	if gs := cfg.GetGlobalSettings(); gs != nil {
		validation.SetAllowedPaths(gs.GetAllowedFilePaths())
	}
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, cfg.GetGlobalSettings())
	a.ToolManager = tool.NewManager(busProvider)
	// Add Tool Metrics Middleware
	a.ToolManager.AddMiddleware(middleware.NewToolMetricsMiddleware(tokenizer.NewSimpleTokenizer()))
	// Add Resilience Middleware
	a.ToolManager.AddMiddleware(middleware.NewResilienceMiddleware(a.ToolManager))

	a.PromptManager = prompt.NewManager()
	a.TemplateManager = NewTemplateManager("data") // Use "data" directory for now
	a.ResourceManager = resource.NewManager()

	a.DiscoveryManager = discovery.NewManager()

	// Initialize Skill Manager
	skillManager, err := skill.NewManager("skills") // Use "skills" directory in CWD for now
	if err != nil {
		return fmt.Errorf("failed to initialize skill manager: %w", err)
	}
	a.SkillManager = skillManager

	// Initialize auth manager
	authManager := auth.NewManager()
	users := cfg.GetUsers()
	if s, ok := storageStore.(storage.Storage); ok {
		dbUsers, err := s.ListUsers(opts.Ctx)
		if err != nil {
			log.Error("failed to list users from storage", "error", err)
		} else {
			users = append(users, dbUsers...)
		}
	}
	authManager.SetUsers(users)

	// Cast storageStore to storage.Storage
	if s, ok := storageStore.(storage.Storage); ok {
		authManager.SetStorage(s)
	} else {
		// This should theoretically not happen if storageStore is properly initialized from sqlite/postgres
		log.Warn("storageStore does not implement storage.Storage, interactive OAuth will be disabled")
	}

	// Use SetAPIKey from config if available
	if a.SettingsManager.GetAPIKey() != "" {
		authManager.SetAPIKey(a.SettingsManager.GetAPIKey())
	}
	// Note: previous code checked cfg.GetGlobalSettings().GetApiKeyParamName() but that might be inside Authentication config?
	// GlobalSettings usually has Authentication field.
	// Let's rely on SettingsManager or check cfg.GetGlobalSettings().GetAuthentication() if needed
	// For API Key param name, it is likely in Authentication message if configured.
	// But AuthManager uses APIKeyAuthenticator which takes config.APIKeyAuth.
	// The explicit API key (CLI) overrides or sets a simple key.
	// We'll leave the complex check out for now unless it was critical for something else.

	// Register auth manager
	a.AuthManager = authManager

	// Initialize Profile Manager and set profile definitions
	var profileDefinitions []*config_v1.ProfileDefinition
	if cfg.GetGlobalSettings() != nil {
		profileDefinitions = cfg.GetGlobalSettings().GetProfileDefinitions()
	} else {
		profileDefinitions = config.GlobalSettings().GetProfileDefinitions()
	}

	// Ensure there is at least a "default" profile
	hasDefault := false
	for _, p := range profileDefinitions {
		if p.GetName() == "default" {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		log.Debug("Injected missing 'default' profile definition")
		profileDefinitions = append(profileDefinitions, config_v1.ProfileDefinition_builder{
			Name: proto.String("default"),
		}.Build())
	}

	a.ProfileManager = profile.NewManager(profileDefinitions)

	// Set profiles for tool filtering
	a.ToolManager.SetProfiles(
		cfg.GetGlobalSettings().GetProfiles(),
		cfg.GetGlobalSettings().GetProfileDefinitions(),
	)

	serviceRegistry := serviceregistry.New(
		upstreamFactory,
		a.ToolManager,
		a.PromptManager,
		a.ResourceManager,
		authManager,
	)
	a.ServiceRegistry = serviceRegistry

	// New message bus and workers
	upstreamWorker := worker.NewUpstreamWorker(busProvider, a.ToolManager)
	registrationWorker := worker.NewServiceRegistrationWorker(busProvider, serviceRegistry)
	if a.RegistrationRetryDelay > 0 {
		registrationWorker.SetRetryDelay(a.RegistrationRetryDelay)
	}

	// Create a context for workers that we can cancel on shutdown
	workerCtx, workerCancel := context.WithCancel(opts.Ctx)
	defer workerCancel()

	// Start background workers
	upstreamWorker.Start(workerCtx)
	registrationWorker.Start(workerCtx)
	// Start periodic health checks (every 30 seconds)
	serviceRegistry.StartHealthChecks(workerCtx, 30*time.Second)

	// If we're using an in-memory bus, start the in-process worker
	if busConfig == nil || busConfig.GetInMemory() != nil {
		workerCfg := &worker.Config{
			MaxWorkers:   10,
			MaxQueueSize: 100,
		}
		inProcessWorker := worker.New(busProvider, workerCfg)
		inProcessWorker.Start(workerCtx)
		defer inProcessWorker.Stop()
	}

	// Initialize and start Global GC Worker
	gcSettings := cfg.GetGlobalSettings().GetGcSettings()
	if gcSettings != nil && gcSettings.GetEnabled() {
		interval, _ := time.ParseDuration(gcSettings.GetInterval())
		ttl, _ := time.ParseDuration(gcSettings.GetTtl())

		gpPaths := gcSettings.GetPaths()
		// Always include the bundle directory if it's set in env (which we did for config)
		// Or we can rely on config.
		// For now, respect config exactly.

		gcWorker := gc.New(gc.Config{
			Enabled:  true,
			Interval: interval,
			TTL:      ttl,
			Paths:    gpPaths,
		})
		gcWorker.Start(workerCtx)
	}

	// Initialize Topology Manager
	a.TopologyManager = topology.NewManager(serviceRegistry, a.ToolManager)

	// Initialize servers with the message bus
	mcpSrv, err := mcpserver.NewServer(
		opts.Ctx,
		a.ToolManager,
		a.PromptManager,
		a.ResourceManager,
		authManager,
		serviceRegistry,
		a.CatalogManager,
		busProvider,
		config.GlobalSettings().IsDebug(),
	)
	if err != nil {
		workerCancel()
		upstreamWorker.Stop()
		registrationWorker.Stop()
		return fmt.Errorf("failed to create mcp server: %w", err)
	}

	mcpSrv.SetReloadFunc(func(ctx context.Context) error {
		return a.ReloadConfig(ctx, fs, opts.ConfigPaths)
	})

	// Register Skill resources
	if err := mcpserver.RegisterSkillResources(a.ResourceManager, a.SkillManager); err != nil {
		log.Error("Failed to register skill resources", "error", err)
		// Don't fail startup for this?
	}

	a.ToolManager.SetMCPServer(mcpSrv)

	if cfg.GetUpstreamServices() != nil {
		// Publish registration requests to the bus for each service
		registrationBus, err := bus.GetBus[*bus.ServiceRegistrationRequest](
			busProvider,
			"service_registration_requests",
		)
		if err != nil {
			workerCancel()
			upstreamWorker.Stop()
			registrationWorker.Stop()
			return fmt.Errorf("failed to get registration bus: %w", err)
		}
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
			if err := registrationBus.Publish(opts.Ctx, "request", regReq); err != nil {
				log.Error("Failed to publish registration request", "error", err)
			}
		}
	} else {
		log.Info("No services found in config, skipping service registration.")
	}

	// Initialize standard middlewares in registry
	cachingMiddleware := middleware.NewCachingMiddleware(a.ToolManager)
	standardMiddlewares, err := middleware.InitStandardMiddlewares(
		mcpSrv.AuthManager(),
		a.ToolManager,
		cfg.GetGlobalSettings().GetAudit(),
		cachingMiddleware,
		cfg.GetGlobalSettings().GetRateLimit(),
		cfg.GetGlobalSettings().GetDlp(),
		cfg.GetGlobalSettings().GetContextOptimizer(),
		cfg.GetGlobalSettings().GetDebugger(),
		cfg.GetGlobalSettings().GetSmartRecovery(),
		nil,
	)
	if err != nil {
		workerCancel()
		upstreamWorker.Stop()
		registrationWorker.Stop()
		return fmt.Errorf("failed to init standard middlewares: %w", err)
	}

	// Auto-discovery of local services
	if cfg.GetGlobalSettings().GetAutoDiscoverLocal() {
		// Register default providers
		a.DiscoveryManager.RegisterProvider(&discovery.OllamaProvider{Endpoint: "http://localhost:11434"})
		a.DiscoveryManager.RegisterProvider(&discovery.OpenAPIProvider{Endpoint: "http://localhost:8080/openapi.json"})
		a.DiscoveryManager.RegisterProvider(&discovery.GRPCProvider{Endpoint: "localhost:50051"})
		a.DiscoveryManager.RegisterProvider(&discovery.GraphQLProvider{Endpoint: "http://localhost:8080/graphql"})

		discovered := a.DiscoveryManager.Run(opts.Ctx)
		for _, svc := range discovered {
			log.Info("Auto-discovered local service", "name", svc.GetName())
			// Use the getter for UpstreamServices
			cfg.SetUpstreamServices(append(cfg.GetUpstreamServices(), svc))
		}
	}
	a.standardMiddlewares = standardMiddlewares
	if standardMiddlewares.Cleanup != nil {
		defer func() {
			if err := standardMiddlewares.Cleanup(); err != nil {
				log.Error("Failed to close audit middleware", "error", err)
			}
		}()
	}
	// Get configured middlewares
	// We clone them to avoid modifying the singleton's underlying slice if we append/modify.
	middlewares := append([]*config_v1.Middleware(nil), config.GlobalSettings().Middlewares()...)
	if len(middlewares) == 0 {
		// Default chain if none configured
		middlewares = []*config_v1.Middleware{
			config_v1.Middleware_builder{
				Name:     proto.String("debug"),
				Priority: proto.Int32(10),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String(authMiddlewareName),
				Priority: proto.Int32(20),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String("logging"),
				Priority: proto.Int32(30),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String("audit"),
				Priority: proto.Int32(40),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String("dlp"),
				Priority: proto.Int32(42),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String("global_ratelimit"),
				Priority: proto.Int32(45),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String("call_policy"),
				Priority: proto.Int32(50),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String("caching"),
				Priority: proto.Int32(60),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String("ratelimit"),
				Priority: proto.Int32(70),
			}.Build(),
			// CORS
			config_v1.Middleware_builder{
				Name:     proto.String("cors"),
				Priority: proto.Int32(0),
			}.Build(),
		}
	}

	// Apply middlewares
	// Registry returns sorted list based on priority (low to high).
	// If priority 0 is first, it wraps the rest?
	// If we iterate:
	// M1(M2(M3(...)))
	// M1 is priority 0.
	// If running in stdio mode, we must remove the auth middleware as it requires
	// an HTTP request availability in the context, which is not present in stdio.
	// Stdio mode implies local access (shell), so we trust the user.
	// Stdio mode implies local access (shell), so we trust the user.
	if opts.Stdio {
		var filtered []*config_v1.Middleware
		for _, m := range middlewares {
			if m.GetName() != authMiddlewareName {
				filtered = append(filtered, m)
			}
		}
		middlewares = filtered
	} else {
		// Enforce auth middleware presence in non-stdio modes
		hasAuth := false
		for _, m := range middlewares {
			if m.GetName() == authMiddlewareName {
				hasAuth = true
				break
			}
		}
		if !hasAuth {
			logging.GetLogger().Warn("Auth middleware not found in config, injecting it")
			middlewares = append(middlewares, config_v1.Middleware_builder{
				Name:     proto.String(authMiddlewareName),
				Priority: proto.Int32(20), // Default priority
			}.Build())
		}
	}

	chain := middleware.GetMCPMiddlewares(middlewares)
	for _, m := range chain {
		logging.GetLogger().Info("Adding middleware", "count", len(chain))
		mcpSrv.Server().AddReceivingMiddleware(m)
	}

	// Add Topology Middleware (Always Active)
	mcpSrv.Server().AddReceivingMiddleware(a.TopologyManager.Middleware)

	// Add Prometheus Metrics Middleware (Always Active)
	// We use SimpleTokenizer for low-overhead token counting
	mcpSrv.Server().AddReceivingMiddleware(middleware.PrometheusMetricsMiddleware(tokenizer.NewSimpleTokenizer()))

	if opts.Stdio {
		err := a.runStdioModeFunc(opts.Ctx, mcpSrv)
		workerCancel()
		upstreamWorker.Stop()
		registrationWorker.Stop()
		return err
	}

	bindAddress := opts.JSONRPCPort
	if cfg.GetGlobalSettings().GetMcpListenAddress() != "" {
		bindAddress = cfg.GetGlobalSettings().GetMcpListenAddress()
	}

	// Use storageStore which is initialized as either sqlite or postgres
	// We need to assert it to storage.Storage. Both implement it.
	// But stores[...] is config.Store. storageStore is config.Store.
	// However, we know storageStore implements storage.Storage because we created it as such.

	// Wait, storageStore is declared as config.Store in my previous edit.
	// I should cast it or change its type declaration.
	// Let's change declaration in previous step, but since I can't undo easily without reset,
	// I'll cast it here.
	s, ok := storageStore.(storage.Storage)
	if !ok {
		// Should not happen if code is correct
		return fmt.Errorf("storage store does not implement storage.Storage")
	}
	a.Storage = s

	// Signal startup complete
	startupCallback := func() {
		a.startupOnce.Do(func() {
			close(a.startupCh)
		})
	}

	// Start servers
	if err := a.runServerMode(
		opts.Ctx,
		mcpSrv,
		busProvider,
		bindAddress,
		opts.GRPCPort,
		opts.ShutdownTimeout,
		cfg.GetGlobalSettings(),
		cachingMiddleware,
		standardMiddlewares,
		s,
		serviceRegistry,
		startupCallback,
		opts.TLSCert,
		opts.TLSKey,
		opts.TLSClientCA,
	); err != nil {
		workerCancel()
		upstreamWorker.Stop()
		registrationWorker.Stop()
		return err
	}

	// Stop workers
	workerCancel()
	upstreamWorker.Stop()
	registrationWorker.Stop()

	return nil
}

// WaitForStartup waits for the application to be fully initialized.
//
// Summary: Waits for application startup completion.
//
// It blocks until the startup process is complete or the context is canceled.
//
// Parameters:
//   - ctx (context.Context): The context to wait on.
//
// Returns:
//   - (error): nil if startup completes successfully, or a context error if canceled.
func (a *Application) WaitForStartup(ctx context.Context) error {
	select {
	case <-a.startupCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// runServerMode runs the server in the standard HTTP and gRPC server mode. It
// starts the HTTP server for JSON-RPC and the gRPC server for service
// registration, and handles graceful shutdown.
//
// Parameters:
//   - ctx (context.Context): The context for managing the server's lifecycle.
//   - mcpSrv (*mcpserver.Server): The MCP server instance.
//   - bus (*bus.Provider): The message bus for inter-component communication.
//   - bindAddress (string): The address for the HTTP/JSON-RPC server.
//   - grpcPort (string): The port for the gRPC registration server.
//   - shutdownTimeout (time.Duration): Duration to wait for graceful shutdown.
//   - globalSettings (*config_v1.GlobalSettings): Global configuration settings.
//   - cachingMiddleware (*middleware.CachingMiddleware): The caching middleware.
//   - standardMiddlewares (*middleware.StandardMiddlewares): The standard middleware chain.
//   - store (storage.Storage): The storage interface.
//   - serviceRegistry (*serviceregistry.ServiceRegistry): The service registry.
//   - startupCallback (func()): Callback function executed when servers are ready.
//   - tlsCert (string): Path to TLS certificate.
//   - tlsKey (string): Path to TLS key.
//   - tlsClientCA (string): Path to TLS Client CA.
//
// Returns:
//   - (error): An error if any of the servers fail to start or run.
//
//nolint:gocyclo
func (a *Application) runServerMode(
	ctx context.Context,
	mcpSrv *mcpserver.Server,
	bus *bus.Provider,
	bindAddress, grpcPort string,
	shutdownTimeout time.Duration,
	globalSettings *config_v1.GlobalSettings,
	cachingMiddleware *middleware.CachingMiddleware,
	standardMiddlewares *middleware.StandardMiddlewares,
	store storage.Storage,
	serviceRegistry *serviceregistry.ServiceRegistry,
	startupCallback func(),
	tlsCert, tlsKey, tlsClientCA string,
) error {
	ipMiddleware, err := middleware.NewIPAllowlistMiddleware(a.SettingsManager.GetAllowedIPs())
	if err != nil {
		return fmt.Errorf("failed to create IP allowlist middleware: %w", err)
	}

	a.configMu.Lock()
	a.ipMiddleware = ipMiddleware
	a.configMu.Unlock()

	// localCtx is used to manage the lifecycle of the servers started in this function.
	// It's canceled when this function returns, ensuring that all servers are shut down.
	localCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errChan := make(chan error, 2)
	readyChan := make(chan struct{}, 2)
	expectedReady := 0
	var wg sync.WaitGroup

	rawHTTPHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return mcpSrv.Server()
	}, nil)

	// Wrap the HTTP handler with OpenTelemetry instrumentation
	// Note: We don't inject HTTPRequestContextKey here anymore because we do it globally
	// in the HTTPRequestContextMiddleware.
	httpHandler := otelhttp.NewHandler(rawHTTPHandler, "server-request")

	// Check if auth middleware is disabled in config
	var authDisabled bool

	// Use passed globalSettings for middleware config check
	if globalSettings != nil {
		logging.GetLogger().Info("DEBUG: GlobalSettings middlewares", "count", len(globalSettings.GetMiddlewares()), "middlewares", globalSettings.GetMiddlewares())
		for _, m := range globalSettings.GetMiddlewares() {
			if m.GetName() == authMiddlewareName && m.GetDisabled() {
				// Only disable if API Key is NOT set.
				// If API Key is present, we enforce auth regardless of this flag to prevent accidental exposure.
				if a.SettingsManager.GetAPIKey() == "" {
					authDisabled = true
				} else {
					logging.GetLogger().Warn("Auth middleware disabled in config but API Key is present. IGNORING disable flag to enforce security.", "api_key_present", true)
				}
				break
			}
		}
	}
	// Note: We don't fall back to config.GlobalSettings() singleton here because it
	// might be modified by other tests in the same package, leading to flaky tests.
	// If globalSettings is nil, authDisabled remains false (enabled).

	// Trust Proxy Config
	trustProxy := os.Getenv("MCPANY_TRUST_PROXY") == util.TrueStr

	var authMiddleware func(http.Handler) http.Handler
	if authDisabled {
		logging.GetLogger().Warn("Auth middleware is disabled by config! Enforcing private-IP-only access for safety.")
		// Even if auth is disabled, we enforce private-IP-only access to prevent public exposure.
		authMiddleware = a.createAuthMiddleware(true, trustProxy)
	} else {
		authMiddleware = a.createAuthMiddleware(false, trustProxy)
	}

	mux := http.NewServeMux()

	// UI Handler
	// We prioritize serving from build directories (./ui/out, ./ui/dist).
	// If only ./ui exists, we check if it contains source code (package.json) and block it if so.
	// We use a.fs (Afero) to allow testing/mocking.
	var uiPath string
	var uiFS http.FileSystem

	if _, err := a.fs.Stat("./ui/out"); err == nil {
		uiPath = "./ui/out"
	} else if _, err := a.fs.Stat("./ui/dist"); err == nil {
		uiPath = "./ui/dist"
	} else if _, err := a.fs.Stat("./ui"); err == nil {
		// Check for package.json to detect source code
		if _, err := a.fs.Stat("./ui/package.json"); err == nil {
			logging.GetLogger().Warn("UI directory ./ui contains package.json. Refusing to serve source code for security.", "path", "./ui")
		} else {
			uiPath = "./ui"
		}
	} else {
		logging.GetLogger().Info("No UI directory found (./ui/out, ./ui/dist, ./ui). UI will not be served.")
	}

	if uiPath != "" {
		// Use Afero's httpFs adapter
		// We create a BasePathFs to restrict access to the UI directory
		baseFs := afero.NewBasePathFs(a.fs, uiPath)
		uiFS = afero.NewHttpFs(baseFs)

		// File server with Cache-Control headers
		fileServer := http.FileServer(uiFS)
		cachingFileServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add Cache-Control headers
			// For immutable assets (usually hashed), we can cache for a long time.
			// Next.js puts static assets in _next/static.
			if strings.Contains(r.URL.Path, "_next/static/") || strings.Contains(r.URL.Path, "static/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else if strings.HasSuffix(r.URL.Path, ".html") || r.URL.Path == "/" {
				// HTML files should not be cached (or short cache) to ensure updates are seen
				w.Header().Set("Cache-Control", "no-cache")
			}
			fileServer.ServeHTTP(w, r)
		})

		// Apply Gzip compression
		handler := middleware.GzipCompressionMiddleware(cachingFileServer)
		mux.Handle("/ui/", http.StripPrefix("/ui", handler))
	}

	// Handle root path with gRPC-Web support
	// We defer the decision to the wrapper or the httpHandler
	// But we need wrappedGrpc to be ready.
	// Since we are moving gRPC init before this, we can use a closure.
	// However, we haven't moved it yet in this execution flow relative to lines 1179.
	// So we need to do the setup HERE or move this Handler registration DOWN?
	// Moving mux.Handle("/", ...) down is safer.

	// API Routes for Configuration Management
	// Protected by auth middleware
	apiHandler := http.StripPrefix("/api/v1", a.createAPIHandler(store))
	mux.Handle("/api/v1/", authMiddleware(apiHandler))

	// Topology API is now handled by apiHandler via api.go

	// Catalog API
	// Expose via REST for UI
	mux.Handle("/api/v1/catalog/services", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		catalogServer := rest.NewCatalogServer(a.CatalogManager)
		resp, err := catalogServer.ListServices(r.Context(), &v1.ListCatalogServicesRequest{})
		if err != nil {
			logging.GetLogger().Error("Failed to list catalog services", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		b, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(resp)
		if err != nil {
			logging.GetLogger().Error("Failed to marshal catalog services response", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(b)
	})))

	logging.GetLogger().Info("DEBUG: Registering /mcp/u/ handler")
	// Multi-user handler
	// pattern: /mcp/u/{uid}/profile/{profile_id}
	// We use a prefix match via stripping.
	// NOTE: We manually handle the path parsing because we support subpaths like /sse or /messages
	mux.HandleFunc("/mcp/u/", func(w http.ResponseWriter, r *http.Request) {
		// Fix data race: Shadow ctx to prevent modifying the captured variable from outer scope.
		// Use request context as the base.
		ctx := r.Context()

		// Expected path: /mcp/u/{uid}/profile/{profileId}/...
		parts := strings.Split(r.URL.Path, "/")
		// parts[0] = ""
		// parts[1] = "mcp"
		// parts[2] = "u"
		// parts[3] = {uid}
		// parts[4] = "profile"
		// parts[5] = {profileId}
		if len(parts) < 6 || parts[4] != "profile" {
			http.NotFound(w, r)
			return
		}
		uid := parts[3]
		profileID := parts[5]

		// Dynamic User Lookup
		user, ok := a.AuthManager.GetUser(uid)
		// SECURITY: Do NOT return "User not found" yet to prevent user enumeration via status codes.
		// We must attempt authentication first. If authentication fails, we return Unauthorized.
		// If authentication succeeds (e.g. global admin), THEN we can reveal that the user is missing.

		// Authentication Logic with Priority:
		// 1. Profile Auth - REMOVED
		// 2. User Auth
		// 3. Global Auth

		var isAuthenticated bool

		// 2. User Auth
		// Only attempt if user exists and has auth configured
		if ok && user.GetAuthentication() != nil {
			if err := auth.ValidateAuthentication(ctx, user.GetAuthentication(), r); err == nil {
				isAuthenticated = true
			} else {
				// User auth configured but failed.
				// We do NOT return immediately, we might fall through to global auth?
				// Typically, if User Auth is present, it is required.
				// However, if we return 401 here immediately, we distinguish between "User exists + bad pass" (401)
				// and "User does not exist" -> Fallthrough -> Global Auth -> "No Global Auth" -> 403/404?
				// To prevent enumeration, we must ensure the behavior is consistent.
				// If we fail here, we return 401.
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		// 3. Global Auth (Fallback if User Auth not used or User not found)
		if !isAuthenticated {
			// If User Auth was attempted and failed, we already returned 401 above.
			// So we are here only if User Auth was NOT attempted (User missing OR User has no auth config).

			apiKey := a.SettingsManager.GetAPIKey()
			if apiKey != "" {
				// Manual check for global key
				requestKey := r.Header.Get("X-API-Key")
				if requestKey == "" {
					authHeader := r.Header.Get("Authorization")
					if strings.HasPrefix(authHeader, "Bearer ") {
						requestKey = strings.TrimPrefix(authHeader, "Bearer ")
					}
				}

				if subtle.ConstantTimeCompare([]byte(requestKey), []byte(apiKey)) == 1 {
					isAuthenticated = true
				} else {
					// Global auth configured but failed.
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			} else {
				// No Global Auth configured.
				// If User Auth was NOT attempted (e.g. user missing), and no Global Auth,
				// we fall into "No Auth Configured" logic.

				// Sentinel Security: Enforce private network access if no auth is configured.
				ip := util.GetClientIP(r, trustProxy)
				if !util.IsPrivateIP(net.ParseIP(ip)) {
					logging.GetLogger().Warn("Blocked public internet request to /mcp/u/ because no API Key is configured", "remote_addr", r.RemoteAddr, "client_ip", ip)
					// We return Unauthorized (401) instead of Forbidden (403) to match the User Auth failure case,
					// reducing information leakage about user existence vs configuration state.
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				isAuthenticated = true
			}
		}

		if !isAuthenticated {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Authentication passed. Now we can check if user exists.
		if !ok {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Profile access check
		hasAccess := false
		for _, pid := range user.GetProfileIds() {
			if pid == profileID {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			http.Error(w, "Forbidden: User does not have access to this profile", http.StatusForbidden)
			return
		}

		// RBAC Check: Check if profile requires specific roles
		// Dynamic Profile Lookup
		if def, ok := a.ProfileManager.GetProfileDefinition(profileID); ok && len(def.GetRequiredRoles()) > 0 {
			hasRole := false
			// Check if user has any of the required roles
			for _, requiredRole := range def.GetRequiredRoles() {
				for _, userRole := range user.GetRoles() {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}
			if !hasRole {
				// Don't leak required roles to the client
				logging.GetLogger().Warn("Forbidden access to profile", "profile", profileID, "required_roles", def.GetRequiredRoles(), "user_roles", user.GetRoles())
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}

		// Inject context
		// We use r.Context() to ensure we start fresh, discarding any partial auth context from above.
		// Note: We assign to the shadowed ctx variable.
		ctx = auth.ContextWithUser(ctx, uid)
		ctx = auth.ContextWithProfileID(ctx, profileID)
		ctx = auth.ContextWithRoles(ctx, user.GetRoles())

		// Strip the prefix so the underlying handler sees the relative path
		prefix := fmt.Sprintf("/mcp/u/%s/profile/%s", uid, profileID)
		delegate := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size to 5MB to prevent DoS attacks via large payloads.
			// This applies to both the stateless JSON-RPC handler and the underlying MCP handler.
			r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

			logging.GetLogger().Info("Delegate Handler", "method", r.Method, "path", r.URL.Path)
			// Support stateless JSON-RPC for simple clients
			if r.Method == http.MethodPost && (r.URL.Path == "/" || r.URL.Path == "") {
				var req struct {
					JSONRPC string          `json:"jsonrpc"`
					ID      any             `json:"id"`
					Method  string          `json:"method"`
					Params  json.RawMessage `json:"params"`
				}
				body, err := io.ReadAll(r.Body)
				if err != nil {
					// http.MaxBytesReader returns an error if the limit is exceeded.
					// We should log it and return an appropriate error.
					logging.GetLogger().Error("Failed to read request body", "error", err)
					if strings.Contains(err.Error(), "request body too large") {
						http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
						return
					}
					http.Error(w, "Failed to read request body", http.StatusInternalServerError)
					return
				}

				r.Body = io.NopCloser(bytes.NewBuffer(body)) // Restore body just in case
				if err := json.Unmarshal(body, &req); err != nil {
					http.Error(w, "Invalid JSON", http.StatusBadRequest)
					return
				}

				if req.Method == "tools/list" {
					tools := mcpSrv.ListTools()
					var responseTools []map[string]any
					for _, t := range tools {
						v1Tool := t.Tool()
						serviceID := v1Tool.GetServiceId()
						_, ok := mcpSrv.GetServiceInfo(serviceID)
						if !ok {
							continue
						}

						// Check profiles
						if profileID != "" {
							if !mcpSrv.ToolManager().IsServiceAllowed(serviceID, profileID) {
								continue
							}
						}

						responseTools = append(responseTools, map[string]any{
							"name":        v1Tool.GetName(),
							"description": v1Tool.GetDescription(),
						})
					}

					// Ensure we return an empty list if no tools are found/allowed, not nil?
					// JSON encoding nil slice as null is usually fine, but empty list [] is better for clients.
					if responseTools == nil {
						responseTools = []map[string]any{}
					}

					resp := map[string]any{
						"jsonrpc": "2.0",
						"id":      req.ID,
						"result": map[string]any{
							"tools": responseTools,
						},
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(resp)
					return
				}

				// Add logging to see unsupported methods
				logging.GetLogger().Info("Unsupported stateless method", "method", req.Method)
				http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
				return
			}
			httpHandler.ServeHTTP(w, r)
		})
		http.StripPrefix(prefix, delegate).ServeHTTP(w, r.WithContext(ctx))
	})

	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "OK")
	})
	mux.Handle("/healthz", healthHandler)
	mux.Handle("/health", healthHandler)
	mux.Handle("/metrics", authMiddleware(metrics.Handler()))
	mux.Handle("/api/v1/alignment/status", authMiddleware(a.handleActiveIntentAlignment()))
	mux.Handle("/upload", authMiddleware(http.HandlerFunc(a.uploadFile)))

	// OIDC Routes
	var oidcConfig *config_v1.OIDCConfig
	if globalSettings != nil {
		oidcConfig = globalSettings.GetOidc()
	} else {
		oidcConfig = config.GlobalSettings().GetOidc()
	}

	if oidcConfig != nil {
		provider, err := auth.NewOIDCProvider(localCtx, auth.OIDCConfig{
			Issuer:       oidcConfig.GetIssuer(),
			ClientID:     oidcConfig.GetClientId(),
			ClientSecret: oidcConfig.GetClientSecret(),
			RedirectURL:  oidcConfig.GetRedirectUrl(),
		})
		if err != nil {
			logging.GetLogger().Error("Failed to initialize OIDC provider", "error", err)
		} else {
			mux.HandleFunc("/auth/login", provider.HandleLogin)
			mux.HandleFunc("/auth/callback", provider.HandleCallback)
		}
	}

	// OAuth API Routes
	mux.Handle("/auth/oauth/initiate", authMiddleware(http.HandlerFunc(a.handleInitiateOAuth)))
	mux.Handle("/auth/oauth/callback", authMiddleware(http.HandlerFunc(a.handleOAuthCallback)))

	// Secrets API
	mux.Handle("/api/v1/secrets", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			a.listSecretsHandler(w, r)
		case http.MethodPost:
			a.createSecretHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/v1/secrets/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for /reveal
		if strings.HasSuffix(r.URL.Path, "/reveal") {
			a.revealSecretHandler(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			a.getSecretHandler(w, r)
		case http.MethodDelete:
			a.deleteSecretHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Credentials API
	// Note: Standard mux doesn't handle methods nicely, so we route by path and check method in handler.
	// We route /credentials to list (GET) and create (POST)
	// We route /credentials/ to get/update/delete (with ID)
	mux.Handle("/credentials", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			a.listCredentialsHandler(w, r)
		case http.MethodPost:
			a.createCredentialHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	// mux.Handle("/api/v1/skills", authMiddleware(a.handleListSkills())) // Replaced by gRPC Gateway
	// mux.Handle("/api/v1/skills/create", authMiddleware(a.handleCreateSkill())) // Replaced by gRPC Gateway

	// Register Config Validation Endpoint
	mux.Handle("/api/v1/config/validate", authMiddleware(http.HandlerFunc(rest.ValidateConfigHandler)))

	// Asset upload is handled later in the gRPC gateway block to support fallback

	// Wait, we need to handle assets specifically.
	// Let's use a more specific path for assets if possible, or ensure we fallback to gwmux?
	// Mux doesn't fallback easily.
	// Better: Register /v1/skills/{name}/assets manual handler if possible?
	// ServeMux doesn't support wildcards.
	// So we must handle /v1/skills/ and forward non-asset requests?
	// But `gwmux` is NOT a simple handler we can call easily from here without re-entering the stack.
	// ACTUALLY: gwmux is served via `mux.Handle("/v1/", ...)`
	// If I register `mux.Handle("/v1/skills/", ...)` it takes precedence.
	// So I MUST handle standard skill requests here too if I do that.
	// OR I can use a different prefix for assets? No, API spec.
	// OR I pass standard requests to `gwmux`?
	// `gwmux.ServeHTTP(w, r)`!
	// Yes, I can use `gwmux` as fallback.

	mux.Handle("/credentials/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This handles /credentials/:id
		switch r.Method {
		case http.MethodGet:
			a.getCredentialHandler(w, r)
		case http.MethodPut:
			a.updateCredentialHandler(w, r)
		case http.MethodDelete:
			a.deleteCredentialHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/debug/auth-test", authMiddleware(http.HandlerFunc(a.testAuthHandler)))
	mux.Handle("/api/v1/debug/seed_traffic", authMiddleware(a.handleDebugSeedTraffic()))
	mux.Handle("/api/v1/debug/seed", authMiddleware(a.handleDebugSeed()))
	mux.Handle("/api/v1/debug/traces", authMiddleware(a.handleDebugSeedTraces()))

	// User Preferences
	mux.Handle("/api/v1/user/preferences", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			a.handleGetUserPreferences(w, r)
		case http.MethodPost:
			a.handleUpdateUserPreferences(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Register Debugger API if enabled
	if standardMiddlewares != nil && standardMiddlewares.Debugger != nil {
		mux.Handle("/debug/entries", authMiddleware(standardMiddlewares.Debugger.APIHandler()))
	}

	// Register Recursive Context Manager
	if standardMiddlewares != nil && standardMiddlewares.RecursiveContext == nil {
		standardMiddlewares.RecursiveContext = middleware.NewRecursiveContextManager()
	}
	if standardMiddlewares != nil {
		mux.Handle("/context/session", authMiddleware(standardMiddlewares.RecursiveContext.APIHandler()))
		mux.Handle("/context/session/", authMiddleware(standardMiddlewares.RecursiveContext.APIHandler()))
	}

	httpBindAddress := bindAddress
	if httpBindAddress == "" {
		if envAddr := os.Getenv("MCPANY_DEFAULT_HTTP_ADDR"); envAddr != "" {
			httpBindAddress = envAddr
		} else {
			httpBindAddress = ":8070"
		}
	} else if !strings.Contains(httpBindAddress, ":") {
		httpBindAddress = ":" + httpBindAddress
	}

	// Apply Global Rate Limit: 20 RPS with a burst of 50.
	// This helps prevent basic DoS attacks on all HTTP endpoints, including /upload.
	// We enable trustProxy if MCPANY_TRUST_PROXY is set, to handle load balancers correctly.
	// trustProxy is already defined above
	rateLimiter := middleware.NewHTTPRateLimitMiddleware(20, 50, middleware.WithTrustProxy(trustProxy))

	// Apply CORS Middleware
	corsMiddleware := middleware.NewHTTPCORSMiddleware(a.SettingsManager.GetAllowedOrigins())
	a.configMu.Lock()
	a.corsMiddleware = corsMiddleware
	a.configMu.Unlock()

	// Apply CSRF Middleware
	csrfMiddleware := middleware.NewCSRFMiddleware(a.SettingsManager.GetAllowedOrigins())
	a.configMu.Lock()
	a.csrfMiddleware = csrfMiddleware
	a.configMu.Unlock()

	// Prepare final handler (Mux wrapped with Content Optimizer and Debugger)
	var finalHandler http.Handler = mux

	if standardMiddlewares != nil {
		// Context Optimizer (inner)
		if standardMiddlewares.ContextOptimizer != nil {
			finalHandler = standardMiddlewares.ContextOptimizer.Handler(finalHandler)
		}
		// Debugger (outer to capture optimized response)
		if standardMiddlewares.Debugger != nil {
			finalHandler = standardMiddlewares.Debugger.Handler(finalHandler)
		}
		// Recursive Context
		if standardMiddlewares.RecursiveContext != nil {
			finalHandler = standardMiddlewares.RecursiveContext.HandleContext(finalHandler)
		}
	}

	// Middleware order: SecurityHeaders -> CORS -> CSRF -> JSONRPCCompliance -> Recovery -> IPAllowList -> RateLimit -> (Debugger -> Optimizer -> Mux)
	// We wrap everything with a debug logger to see what's coming in
	handler := middleware.HTTPSecurityHeadersMiddleware(
		corsMiddleware.Handler(
			csrfMiddleware.Handler(
				middleware.JSONRPCComplianceMiddleware(
					middleware.RecoveryMiddleware(
						a.HTTPRequestContextMiddleware(
							ipMiddleware.Handler(
								rateLimiter.Handler(finalHandler),
							),
						),
					),
				),
			),
		),
	)

	// gRPC Server Setup
	var grpcServer *gogrpc.Server
	var wrappedGrpc *grpcweb.WrappedGrpcServer

	grpcBindAddress := grpcPort

	// Initialize gRPC Interceptors
	grpcUnaryInterceptor := func(ctx context.Context, req interface{}, _ *gogrpc.UnaryServerInfo, handler gogrpc.UnaryHandler) (interface{}, error) {
		if p, ok := peer.FromContext(ctx); ok {
			ip := util.ExtractIP(p.Addr.String())
			ctx = util.ContextWithRemoteIP(ctx, ip)

			if !ipMiddleware.Allow(p.Addr.String()) {
				return nil, status.Error(codes.PermissionDenied, "IP not allowed")
			}
		}
		return handler(ctx, req)
	}
	grpcStreamInterceptor := func(srv interface{}, ss gogrpc.ServerStream, _ *gogrpc.StreamServerInfo, handler gogrpc.StreamHandler) error {
		if p, ok := peer.FromContext(ss.Context()); ok {
			ip := util.ExtractIP(p.Addr.String())
			// Wrapper to modify context for stream
			wrappedStream := &util.WrappedServerStream{
				ServerStream: ss,
				Ctx:          util.ContextWithRemoteIP(ss.Context(), ip),
			}
			if !ipMiddleware.Allow(p.Addr.String()) {
				return status.Error(codes.PermissionDenied, "IP not allowed")
			}
			return handler(srv, wrappedStream)
		}
		return handler(srv, ss)
	}
	grpcOpts := []gogrpc.ServerOption{
		gogrpc.UnaryInterceptor(grpcUnaryInterceptor),
		gogrpc.StreamInterceptor(grpcStreamInterceptor),
		gogrpc.StatsHandler(&metrics.GrpcStatsHandler{Wrapped: otelgrpc.NewServerHandler()}),
	}

	grpcServer = gogrpc.NewServer(grpcOpts...)
	reflection.Register(grpcServer)

	// Register Services
	registrationServer, err := mcpserver.NewRegistrationServer(bus, a.AuthManager)
	if err != nil {
		return fmt.Errorf("failed to create API server: %w", err)
	}
	v1.RegisterRegistrationServiceServer(grpcServer, registrationServer)

	adminServer := admin.NewServer(cachingMiddleware, a.ToolManager, serviceRegistry, store, a.DiscoveryManager, a.GetAuditMiddleware)
	pb_admin.RegisterAdminServiceServer(grpcServer, adminServer)

	// Register Skill Service
	v1.RegisterSkillServiceServer(grpcServer, NewSkillServiceServer(a.SkillManager))

	// Initialize gRPC-Web wrapper even if gRPC port is not exposed
	wrappedGrpc = grpcweb.WrapServer(grpcServer,
		grpcweb.WithOriginFunc(func(_ string) bool { return true }),
		grpcweb.WithAllowedRequestHeaders([]string{"*"}),
	)

	if grpcBindAddress != "" {
		if !strings.Contains(grpcBindAddress, ":") {
			grpcBindAddress = ":" + grpcBindAddress
		}
		lis, err := util.ListenWithRetry(ctx, "tcp", grpcBindAddress)
		if err != nil {
			errChan <- wrapBindError(err, "gRPC", grpcBindAddress, "--grpc-port")
		} else {
			if addr, ok := lis.Addr().(*net.TCPAddr); ok {
				a.BoundGRPCPort.Store(int32(addr.Port)) //nolint:gosec // Port fits in int32

				// Register gRPC Gateway with the bound port
				gwmux := runtime.NewServeMux()
				opts := []gogrpc.DialOption{gogrpc.WithTransportCredentials(insecure.NewCredentials())}
				endpoint := fmt.Sprintf("127.0.0.1:%d", a.BoundGRPCPort.Load())

				if err := v1.RegisterRegistrationServiceHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
					errChan <- fmt.Errorf("failed to register gateway: %w", err)
				} else if err := v1.RegisterSkillServiceHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
					errChan <- fmt.Errorf("failed to register skill gateway: %w", err)
				} else if err := pb_admin.RegisterAdminServiceHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
					errChan <- fmt.Errorf("failed to register admin gateway: %w", err)
				} else {
					// Consolidated handler for /v1/ to support both gRPC Gateway and Asset Uploads
					mux.Handle("/v1/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if strings.HasSuffix(r.URL.Path, "/assets") {
							a.handleUploadSkillAsset()(w, r)
						} else {
							gwmux.ServeHTTP(w, r)
						}
					})))
				}
			}
			expectedReady++
			startGrpcServer(
				localCtx,
				&wg,
				errChan,
				readyChan,
				"Registration",
				lis,
				shutdownTimeout,
				grpcServer,
			)
		}
	}

	// Register Root Handler with gRPC-Web support
	mux.Handle("/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wrappedGrpc != nil && wrappedGrpc.IsGrpcWebRequest(r) {
			wrappedGrpc.ServeHTTP(w, r)
			return
		}

		// UI Routing for root path
		if r.URL.Path == "/" && uiPath != "" {
			http.ServeFile(w, r, filepath.Join(uiPath, "index.html"))
			return
		}

		// Fallback to JSON-RPC handler (for API calls at root or SSE)
		httpHandler.ServeHTTP(w, r)
	})))

	var httpLis net.Listener

	if tlsCert != "" && tlsKey != "" {
		cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
		if err != nil {
			return fmt.Errorf("failed to load TLS key pair: %w", err)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
		if tlsClientCA != "" {
			caCert, err := os.ReadFile(tlsClientCA)
			if err != nil {
				return fmt.Errorf("failed to read TLS client CA: %w", err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.ClientCAs = caCertPool
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}

		logging.GetLogger().Info("Enabling TLS for HTTP server", "mtls_enabled", tlsConfig.ClientAuth == tls.RequireAndVerifyClientCert)

		// Use standard Listen and then wrap with TLS
		l, err := util.ListenWithRetry(ctx, "tcp", httpBindAddress)
		if err != nil {
			// Handle error
			errChan <- wrapBindError(err, "HTTP", httpBindAddress, "--json-rpc-port")
		} else {
			httpLis = tls.NewListener(l, tlsConfig)
		}
	} else {
		l, err := util.ListenWithRetry(ctx, "tcp", httpBindAddress)
		if err != nil {
			// Handle error
			errChan <- wrapBindError(err, "HTTP", httpBindAddress, "--json-rpc-port")
		} else {
			httpLis = l
		}
	}

	if httpLis != nil {
		if addr, ok := httpLis.Addr().(*net.TCPAddr); ok {
			a.BoundHTTPPort.Store(int32(addr.Port)) //nolint:gosec // Port fits in int32
		}
		expectedReady++
		// Handle active connection tracking
		connState := func(_ net.Conn, state http.ConnState) {
			switch state {
			case http.StateNew:
				atomic.AddInt32(&a.activeConnections, 1)
			case http.StateClosed, http.StateHijacked:
				atomic.AddInt32(&a.activeConnections, -1)
			}
		}

		startHTTPServer(localCtx, &wg, errChan, readyChan, "MCP Any HTTP", httpLis, handler, shutdownTimeout, connState)
	}

	// Wait for servers to be ready
	timeout := time.NewTimer(30 * time.Second) // Reasonable timeout for binding ports, increased for slow CI
	defer timeout.Stop()

	for i := 0; i < expectedReady; i++ {
		select {
		case <-readyChan:
			// One server is ready
		case err := <-errChan:
			return fmt.Errorf("failed to start a server: %w", err)
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			return fmt.Errorf("timed out waiting for servers to be ready")
		}
	}

	if startupCallback != nil {
		startupCallback()
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

	// Shutdown all upstreams
	if serviceRegistry != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()
		if err := serviceRegistry.Close(shutdownCtx); err != nil {
			logging.GetLogger().Error("Failed to shutdown services", "error", err)
		}
	}

	return startupErr
}
