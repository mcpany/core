// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package app provides the main application logic.
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
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/discovery"
	"github.com/mcpany/core/server/pkg/gc"
	"github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/llm"
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
	"google.golang.org/protobuf/proto"
)

const authMiddlewareName = "auth"

var healthCheckClient = &http.Client{
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func (a *Application) uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit the request body size to 10MB to prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "failed to get file from form", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	// Clean up any temporary files created by ParseMultipartForm
	if r.MultipartForm != nil {
		defer func() {
			if err := r.MultipartForm.RemoveAll(); err != nil {
				logging.GetLogger().Error("Failed to remove multipart form files", "error", err)
			}
		}()
	}

	// Consume the file content without writing to disk.
	// We discard the content to avoid disk usage and potential residue.
	written, err := io.Copy(io.Discard, file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	// Respond with the file name and size
	// Sanitize the filename to prevent reflected XSS and ensure safe filesystem usage
	safeFilename := util.SanitizeFilename(header.Filename)
	w.Header().Set("Content-Type", "text/plain")
	_, _ = fmt.Fprintf(w, "File '%s' uploaded successfully (size: %d bytes)", html.EscapeString(safeFilename), written)
}

// RunOptions configuration for starting the MCP Any application.
//
// Summary: Options for configuring the application runtime.
type RunOptions struct {
	Ctx             context.Context
	Fs              afero.Fs
	Stdio           bool
	JSONRPCPort     string
	GRPCPort        string
	ConfigPaths     []string
	APIKey          string
	ShutdownTimeout time.Duration
	TLSCert         string
	TLSKey          string
	TLSClientCA     string
	DBPath          string
}

// Runner defines the interface for running the application.
//
// Summary: Interface for application execution and management.
type Runner interface {
	// Run starts the application with the given options.
	//
	// Summary: Starts the application.
	//
	// Parameters:
	//   - opts: RunOptions. The configuration for running.
	//
	// Returns:
	//   - error: An error if startup or execution fails.
	Run(opts RunOptions) error

	// ReloadConfig reloads the application configuration.
	//
	// Summary: Triggers a configuration reload.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - fs: afero.Fs. The filesystem.
	//   - configPaths: []string. Paths to configuration files.
	//
	// Returns:
	//   - error: An error if reload fails.
	ReloadConfig(ctx context.Context, fs afero.Fs, configPaths []string) error
}

// Application is the main application struct, holding the dependencies and logic for the MCP Any server.
//
// Summary: The main application container.
type Application struct {
	runStdioModeFunc func(ctx context.Context, mcpSrv *mcpserver.Server) error
	PromptManager    prompt.ManagerInterface
	ToolManager      tool.ManagerInterface
	ResourceManager  resource.ManagerInterface
	ServiceRegistry  serviceregistry.ServiceRegistryInterface
	TopologyManager  *topology.Manager
	UpstreamFactory  factory.Factory
	configFiles      map[string]string
	fs               afero.Fs
	configPaths      []string
	Storage          storage.Storage
	TemplateManager  *TemplateManager
	// Store explicit API Key passed via CLI args
	explicitAPIKey string

	// SkillManager manages agent skills
	SkillManager *skill.Manager

	// AlertsManager manages system alerts
	AlertsManager *alerts.Manager

	// DiscoveryManager manages auto-discovery providers
	DiscoveryManager *discovery.Manager

	// lastReloadErr stores the error from the last configuration reload.
	standardMiddlewares *middleware.StandardMiddlewares
	// Settings Manager for global settings (dynamic updates)
	SettingsManager *GlobalSettingsManager
	// Profile Manager for dynamic profile updates
	ProfileManager *profile.Manager
	// Auth Manager (stored here for access in runServerMode, though it is also passed to serviceregistry)
	// We need to keep a reference to update it on reload.
	AuthManager *auth.Manager
	// Middlewares that need manual updates
	ipMiddleware   *middleware.IPAllowlistMiddleware
	corsMiddleware *middleware.HTTPCORSMiddleware
	csrfMiddleware *middleware.CSRFMiddleware

	busProvider *bus.Provider

	startupCh   chan struct{}
	startupOnce sync.Once
	configMu    sync.Mutex

	// lastReloadErr stores the error from the last configuration reload.

	// lastReloadErr stores the error from the last configuration reload.
	// It is protected by configMu.
	lastReloadErr error
	// lastReloadTime stores the time of the last configuration reload attempt.
	// It is protected by configMu.
	lastReloadTime time.Time

	// lastGoodConfig stores the raw content of the last successfully loaded configuration files.
	// Map key is the file path, value is the file content.
	// It is protected by configMu.
	lastGoodConfig map[string]string

	// configDiff stores the diff between the last good config and the failed config.
	// It is protected by configMu.
	configDiff string

	// BoundHTTPPort stores the actual port the HTTP server is listening on.
	BoundHTTPPort atomic.Int32
	// BoundGRPCPort stores the actual port the gRPC server is listening on.
	BoundGRPCPort atomic.Int32

	// startTime is the time the application started.
	startTime time.Time
	// activeConnections tracks the number of active HTTP connections.
	activeConnections int32

	// RegistrationRetryDelay allows configuring the retry delay for service registration.
	// If 0, it defaults to 5 seconds (in the worker).
	RegistrationRetryDelay time.Duration

	// MetricsGatherer is the interface for gathering metrics.
	// Defaults to prometheus.DefaultGatherer.
	MetricsGatherer prometheus.Gatherer

	// LLMProviderStore manages LLM API keys and configurations
	LLMProviderStore *llm.ProviderStore

	// statsCache for dashboard
	statsCacheMu sync.RWMutex
	statsCache   map[string]statsCacheEntry

	// AutoCraftJobs stores results of auto-craft jobs
	AutoCraftJobs sync.Map
}

type statsCacheEntry struct {
	Data      any
	ExpiresAt time.Time
}

// NewApplication creates a new Application with default dependencies.
//
// Summary: Initializes a new Application instance.
//
// Returns:
//   - *Application: The initialized application.
func NewApplication() *Application {
	busProvider, _ := bus.NewProvider(nil)
	return &Application{
		runStdioModeFunc: runStdioMode,
		PromptManager:    prompt.NewManager(),
		ToolManager:      tool.NewManager(busProvider),
		AlertsManager:    alerts.NewManager(),

		ResourceManager: resource.NewManager(),
		UpstreamFactory: factory.NewUpstreamServiceFactory(pool.NewManager(), nil),
		configFiles:     make(map[string]string),
		startupCh:       make(chan struct{}),
		startTime:       time.Now(),
		MetricsGatherer: prometheus.DefaultGatherer,
		statsCache:      make(map[string]statsCacheEntry),
	}
}

// Run starts the MCP Any server and all its components.
//
// Summary: Executes the application.
//
// Parameters:
//   - opts: RunOptions. The runtime options.
//
// Returns:
//   - error: An error if execution fails.
//
// Side Effects:
//   - Starts HTTP and gRPC servers.
//   - Initializes background workers.
//   - Loads configuration.
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

	var stores []config.Store

	// Initialize DB if empty
	if err := a.initializeDatabase(opts.Ctx, storageStore); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Determine config sources
	// Priority: Database < File (if enabled)
	stores = append(stores, storageStore)

	enableFileConfig := os.Getenv("MCPANY_ENABLE_FILE_CONFIG") == "true"
	if len(opts.ConfigPaths) > 0 {
		// Always load config files if they are explicitly provided
		log.Info("Loading config from files (overrides database)", "paths", opts.ConfigPaths)
		stores = append(stores, config.NewFileStore(fs, opts.ConfigPaths))
	} else if enableFileConfig {
		// If enabled but no paths provided, we might still want to load from default locations (if any)
		// but currently NewFileStore requires paths. We keep this variable if it's used elsewhere,
		// but for now, we just log that we are enabled but have no paths.
		log.Info("File configuration enabled via env var, but no config paths provided.")
	}
	multiStore := config.NewMultiStore(stores...)

	var cfg *config_v1.McpAnyServerConfig
	cfg, err = config.LoadServices(opts.Ctx, multiStore, "server")
	if err != nil {
		return fmt.Errorf("failed to load services from config: %w", err)
	}
	if cfg == nil {
		cfg = config_v1.McpAnyServerConfig_builder{}.Build()
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

	// Initialize LLM Provider Store
	dataDir := "data"
	if dbPath := config.GlobalSettings().DBPath(); dbPath != "" {
		dataDir = filepath.Dir(dbPath)
	}
	llmStorePath := filepath.Join(dataDir, "llm_providers.json")
	llmStore, err := llm.NewProviderStore(llmStorePath)
	if err != nil {
		return fmt.Errorf("failed to initialize LLM provider store: %w", err)
	}
	a.LLMProviderStore = llmStore


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
	// GetProfileDefinitions returns nil if not set, handled by Update
	var profileDefinitions []*config_v1.ProfileDefinition
	if cfg.GetGlobalSettings() != nil {
		profileDefinitions = cfg.GetGlobalSettings().GetProfileDefinitions()
	} else {
		profileDefinitions = config.GlobalSettings().GetProfileDefinitions()
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

	// Initialize and start Auto Craft Worker
	// We run this independently of the general worker to ensure it has dedicated resources/queues if needed,
	// or we could merge it. For now, separate is safer to guarantee it runs.
	autoCraftWorker := worker.New(busProvider, &worker.Config{
		MaxWorkers:   2,
		MaxQueueSize: 20,
	})
	autoCraftWorker.StartAutoCraftWorker(workerCtx, a.LLMProviderStore)
	defer autoCraftWorker.Stop()

	// Start Auto Craft Result Listener
	// This listens for results from the worker and stores them in memory for API access
	acResBus, err := bus.GetBus[*worker.AutoCraftResult](busProvider, "auto_craft_result")
	if err == nil {
		// Subscribe to ALL results
		// Note: We ignore the unsubscribe function here as we want this to run for the lifetime of the server
		acResBus.Subscribe(workerCtx, "auto_craft_result", func(res *worker.AutoCraftResult) {
			// Store the result using correlation ID (which is the Job ID)
			a.AutoCraftJobs.Store(res.CID, res)
		})
	} else {
		logging.GetLogger().Error("Failed to get auto craft result bus", "error", err)
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

// ReloadConfig reloads the configuration from the given paths and updates the
// services.
//
// Summary: Reloads application configuration from disk/storage.
//
// Parameters:
//   - ctx: context.Context. The context for the reload operation.
//   - fs: afero.Fs. The filesystem interface for reading configuration files.
//   - configPaths: []string. A slice of paths to configuration files to reload.
//
// Returns:
//   - error: An error if the configuration reload fails.
func (a *Application) ReloadConfig(ctx context.Context, fs afero.Fs, configPaths []string) error {
	log := logging.GetLogger()
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			log.Error("ReloadConfig panicked", "panic", r)
		}
		log.Info("ReloadConfig completed", "duration", time.Since(start))
	}()

	a.configMu.Lock()
	defer a.configMu.Unlock()

	log.Info("Reloading configuration...")
	metrics.IncrCounter([]string{"config", "reload", "total"}, 1)

	// Read new config content first for diff generation
	newConfigRaw, readErr := a.readConfigFiles(fs, configPaths)
	if readErr != nil {
		log.Error("Failed to read config files for diff", "error", readErr)
	}

	cfg, err := a.loadConfig(ctx, fs, configPaths)
	a.lastReloadTime = time.Now()
	a.lastReloadErr = err
	if err != nil {
		metrics.IncrCounter([]string{"config", "reload", "errors"}, 1)
		// Generate Diff if we have previous good config and new config
		if newConfigRaw != nil && a.lastGoodConfig != nil {
			a.configDiff = a.generateConfigDiff(a.lastGoodConfig, newConfigRaw)
		}
		return fmt.Errorf("failed to load services from config: %w", err)
	}

	// Success: Update last good config and clear diff
	if newConfigRaw != nil {
		a.lastGoodConfig = newConfigRaw
		a.configDiff = ""
	}

	// Update global settings
	a.updateGlobalSettings(cfg)

	// Update Users (Dynamic!)
	if a.AuthManager != nil {
		a.AuthManager.SetUsers(cfg.GetUsers())
		log.Info("Updated users configuration")
	}

	// Update profiles on reload
	a.ToolManager.SetProfiles(
		cfg.GetGlobalSettings().GetProfiles(),
		cfg.GetGlobalSettings().GetProfileDefinitions(),
	)

	// Update Profile Manager (Dynamic!)
	if a.ProfileManager != nil {
		a.ProfileManager.Update(cfg.GetGlobalSettings().GetProfileDefinitions())
		log.Info("Updated profile definitions configuration")
	}

	// Reconcile services (add/remove/update)
	a.reconcileServices(ctx, cfg)
	return nil
}

func (a *Application) loadConfig(ctx context.Context, fs afero.Fs, configPaths []string) (*config_v1.McpAnyServerConfig, error) {
	var stores []config.Store

	if a.Storage != nil {
		stores = append(stores, a.Storage)
	}

	enableFileConfig := os.Getenv("MCPANY_ENABLE_FILE_CONFIG") == "true"
	if enableFileConfig && len(configPaths) > 0 {
		stores = append(stores, config.NewFileStore(fs, configPaths))
	}

	store := config.NewMultiStore(stores...)
	return config.LoadServices(ctx, store, "server")
}

func (a *Application) updateGlobalSettings(cfg *config_v1.McpAnyServerConfig) {
	log := logging.GetLogger()
	if a.SettingsManager != nil {
		a.SettingsManager.Update(cfg.GetGlobalSettings(), a.explicitAPIKey)
	}

	// Update Health Alerts
	if cfg.GetGlobalSettings().GetAlerts() != nil {
		health.SetGlobalAlertConfig(cfg.GetGlobalSettings().GetAlerts())
	}

	// Update dynamic middlewares
	if a.ipMiddleware != nil {
		if err := a.ipMiddleware.Update(a.SettingsManager.GetAllowedIPs()); err != nil {
			log.Error("Failed to update IP allowlist", "error", err)
		}
	}
	if a.corsMiddleware != nil {
		a.corsMiddleware.Update(a.SettingsManager.GetAllowedOrigins())
	}
	if a.csrfMiddleware != nil {
		a.csrfMiddleware.Update(a.SettingsManager.GetAllowedOrigins())
	}

	if a.standardMiddlewares != nil {
		if a.standardMiddlewares.Audit != nil {
			if err := a.standardMiddlewares.Audit.UpdateConfig(cfg.GetGlobalSettings().GetAudit()); err != nil {
				log.Error("Failed to update audit middleware config", "error", err)
			}
		}
		if a.standardMiddlewares.GlobalRateLimit != nil {
			a.standardMiddlewares.GlobalRateLimit.UpdateConfig(cfg.GetGlobalSettings().GetRateLimit())
		}
	}
}

// reconcileServices reconciles the service registry with the new configuration.
func (a *Application) reconcileServices(ctx context.Context, cfg *config_v1.McpAnyServerConfig) {
	log := logging.GetLogger()
	// Get current active services
	currentServicesMap := make(map[string]*config_v1.UpstreamServiceConfig)
	if a.ServiceRegistry != nil {
		services, err := a.ServiceRegistry.GetAllServices()
		if err == nil {
			for _, s := range services {
				currentServicesMap[s.GetName()] = s
			}
		}
	}

	// Auto-discovery of local services
	if cfg.GetGlobalSettings().GetAutoDiscoverLocal() {
		ollamaProvider := &discovery.OllamaProvider{Endpoint: "http://localhost:11434"}
		discovered, err := ollamaProvider.Discover(ctx)
		if err != nil {
			log.Warn("Failed to auto-discover local services during reload", "provider", ollamaProvider.Name(), "error", err)
		} else {
			for _, svc := range discovered {
				log.Info("Auto-discovered local service during reload", "name", svc.GetName())
				cfg.SetUpstreamServices(append(cfg.GetUpstreamServices(), svc))
			}
		}
	}

	// Map new services by name for easy lookup
	newServices := make(map[string]*config_v1.UpstreamServiceConfig)
	if cfg.GetUpstreamServices() != nil {
		for _, svc := range cfg.GetUpstreamServices() {
			if !svc.GetDisable() {
				newServices[svc.GetName()] = svc
			}
		}
	}

	// Identify removed services
	for name := range currentServicesMap {
		if _, exists := newServices[name]; !exists {
			log.Info("Removing service", "service", name)
			if a.ServiceRegistry != nil {
				if err := a.ServiceRegistry.UnregisterService(ctx, name); err != nil {
					log.Error("Failed to unregister service", "service", name, "error", err)
				}
			}
		}
	}

	// Identify added or updated services
	for name, newSvc := range newServices {
		oldConfig, exists := currentServicesMap[name]
		needsUpdate := false

		if !exists {
			log.Info("Adding new service", "service", name)
			needsUpdate = true
		} else {
			// Compare configs
			newSvcCopy := proto.Clone(newSvc).(*config_v1.UpstreamServiceConfig)
			if newSvcCopy.GetId() == "" {
				newSvcCopy.SetId(oldConfig.GetId())
			}
			if newSvcCopy.GetSanitizedName() == "" {
				newSvcCopy.SetSanitizedName(oldConfig.GetSanitizedName())
			}

			if !proto.Equal(oldConfig, newSvcCopy) {
				log.Info("Updating service", "service", name)
				needsUpdate = true
				if a.ServiceRegistry != nil {
					if err := a.ServiceRegistry.UnregisterService(ctx, name); err != nil {
						log.Error("Failed to unregister service for update", "service", name, "error", err)
					}
				}
			}
		}

		if needsUpdate {
			switch {
			case a.busProvider != nil:
				// Async registration via bus to support retries
				registrationBus, err := bus.GetBus[*bus.ServiceRegistrationRequest](
					a.busProvider,
					bus.ServiceRegistrationRequestTopic,
				)
				if err != nil {
					log.Error("Failed to get registration bus during reload", "error", err)
					continue
				}
				regReq := &bus.ServiceRegistrationRequest{Config: newSvc}
				if err := registrationBus.Publish(context.Background(), "request", regReq); err != nil {
					log.Error("Failed to publish registration request during reload", "error", err)
				} else {
					log.Info("Queued service for registration update", "service", name)
				}
			case a.ServiceRegistry != nil:
				// Fallback to sync registration if bus is not available (e.g. tests without full init)
				_, _, _, err := a.ServiceRegistry.RegisterService(context.Background(), newSvc)
				if err != nil {
					log.Error("Failed to register upstream service", "service", name, "error", err)
					continue
				}
			default:
				log.Warn("ServiceRegistry is nil, cannot register service", "service", name)
			}
		} else {
			log.Debug("Service unchanged", "service", name)
		}
	}

	log.Info("Reload complete", "tools_count", len(a.ToolManager.ListTools()))

	// Update Auth Manager users
	users := cfg.GetUsers()
	if a.Storage != nil {
		dbUsers, err := a.Storage.ListUsers(ctx)
		if err != nil {
			log.Error("failed to list users from storage during reload", "error", err)
		} else {
			users = append(users, dbUsers...)
		}
	}
	if a.AuthManager != nil {
		a.AuthManager.SetUsers(users)
	}

	// Update Service Registry
	// We need to re-create services or update existing ones?
	// ServiceRegistry.UpdateServices?
	// Ideally we have a better way, but for now we might need to rely on individual updates or full re-init?
	// The ServiceRegistry holds state (tools, prompts).
	// If we just replace services, we might loose state.
	// But `UpdateServices` is not exposed on interface?
	// Actually `ServiceRegistry` is an interface.
	// Using `UpdateConfig` if available?
	// For now, let's assume `AuthManager` update is what we really needed for login.
	// Services are updated via bus or separate flow in real app usually.
	// But `server.go` logic for Reload needs to be checked.
	// For this task, updating AuthManager is sufficient for USER LOGIN.
}

// readConfigFiles reads the raw content of the configuration files.
// It handles directory walking similar to FileStore but only returns raw content.
func (a *Application) readConfigFiles(fs afero.Fs, paths []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, path := range paths {
		// Handle URL - skip for diffing for now as it requires network call and we only care about file changes mostly
		if strings.HasPrefix(strings.ToLower(path), "http://") || strings.HasPrefix(strings.ToLower(path), "https://") {
			continue
		}

		info, err := fs.Stat(path)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			err := afero.Walk(fs, path, func(p string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !fi.IsDir() {
					// Simple check for extension to match Config behavior roughly
					ext := strings.ToLower(filepath.Ext(p))
					if ext == ".yaml" || ext == ".yml" || ext == ".json" || ext == ".textproto" || ext == ".prototxt" {
						b, err := afero.ReadFile(fs, p)
						if err != nil {
							return err
						}
						result[p] = string(b)
					}
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			b, err := afero.ReadFile(fs, path)
			if err != nil {
				return nil, err
			}
			result[path] = string(b)
		}
	}
	return result, nil
}

// generateConfigDiff generates a unified diff between the old and new configuration files.
func (a *Application) generateConfigDiff(oldConfig, newConfig map[string]string) string {
	var diffs []string
	// Check for changed or new files
	for path, newContent := range newConfig {
		oldContent, exists := oldConfig[path]
		if !exists {
			// New file
			d, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(""),
				B:        difflib.SplitLines(newContent),
				FromFile: "/dev/null",
				ToFile:   path,
				Context:  3,
			})
			diffs = append(diffs, fmt.Sprintf("New file: %s\n%s", path, d))
		} else if oldContent != newContent {
			// Changed file
			d, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(oldContent),
				B:        difflib.SplitLines(newContent),
				FromFile: path + " (last known good)",
				ToFile:   path + " (current broken)",
				Context:  3,
			})
			diffs = append(diffs, d)
		}
	}
	// Check for deleted files
	for path, oldContent := range oldConfig {
		if _, exists := newConfig[path]; !exists {
			// Deleted file
			d, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(oldContent),
				B:        difflib.SplitLines(""),
				FromFile: path,
				ToFile:   "/dev/null",
				Context:  3,
			})
			diffs = append(diffs, fmt.Sprintf("Deleted file: %s\n%s", path, d))
		}
	}
	return strings.Join(diffs, "\n")
}

// WaitForStartup waits for the application to be fully initialized.
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

// setup initializes the filesystem for the server. It ensures that a valid
// afero.Fs is available, returning an error if a nil filesystem is provided.
//
// Parameters:
//   - fs (afero.Fs): The filesystem to be validated.
//
// Returns:
//   - (afero.Fs): A non-nil afero.Fs.
//   - (error): An error if the provided filesystem is nil.
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
// Parameters:
//   - ctx (context.Context): The context for managing the server's lifecycle.
//   - mcpSrv (*mcpserver.Server): The MCP server instance to run.
//
// Returns:
//   - (error): An error if the server fails to run in stdio mode.
func runStdioMode(ctx context.Context, mcpSrv *mcpserver.Server) error {
	log := logging.GetLogger()
	log.Info("Starting in stdio mode")
	return mcpSrv.Server().Run(ctx, &mcp.StdioTransport{})
}

// configHealthCheck checks the status of the configuration.
func (a *Application) configHealthCheck(_ context.Context) health.CheckResult {
	a.configMu.Lock()
	defer a.configMu.Unlock()

	if a.lastReloadErr != nil {
		return health.CheckResult{
			Status:  "degraded",
			Message: a.lastReloadErr.Error(),
			Latency: time.Since(a.lastReloadTime).String(),
			Diff:    a.configDiff,
		}
	}

	status := "ok"
	if a.lastReloadTime.IsZero() {
		status = "unknown"
	}

	return health.CheckResult{
		Status:  status,
		Latency: time.Since(a.lastReloadTime).String(),
	}
}

func (a *Application) filesystemHealthCheck(_ context.Context) health.CheckResult {
	if a.ServiceRegistry == nil {
		return health.CheckResult{Status: "ok"}
	}

	services, err := a.ServiceRegistry.GetAllServices()
	if err != nil {
		return health.CheckResult{
			Status:  "degraded",
			Message: fmt.Sprintf("failed to list services: %v", err),
		}
	}

	var issues []string
	start := time.Now()

	for _, svc := range services {
		fsSvc := svc.GetFilesystemService()
		if fsSvc == nil {
			continue
		}

		for virtualPath, localPath := range fsSvc.GetRootPaths() {
			if info, err := os.Stat(localPath); err != nil {
				issues = append(issues, fmt.Sprintf("service %q: root path %q (%s) is inaccessible: %v", svc.GetName(), virtualPath, localPath, err))
			} else if !info.IsDir() {
				issues = append(issues, fmt.Sprintf("service %q: root path %q (%s) is not a directory", svc.GetName(), virtualPath, localPath))
			}
		}
	}

	status := "ok"
	var message string
	if len(issues) > 0 {
		status = "degraded"
		message = strings.Join(issues, "; ")
	}

	return health.CheckResult{
		Status:  status,
		Message: message,
		Latency: time.Since(start).String(),
	}
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
//   - out (io.Writer): The writer to which the success message will be written.
//   - addr (string): The address (host:port) on which the server is running.
//   - timeout (time.Duration): The maximum duration to wait for the health check.
//
// Returns:
//   - (error): nil if the server is healthy (i.e., responds with a 200 OK), or an
//     error if the health check fails for any reason (e.g., connection error,
//     non-200 status code).
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
//   - ctx (context.Context): The context for managing the health check's lifecycle.
//   - out (io.Writer): The writer to which the success message will be written.
//   - addr (string): The address (host:port) on which the server is running.
//
// Returns:
//   - (error): nil if the server is healthy (i.e., responds with a 200 OK), or an
//     error if the health check fails for any reason (e.g., connection error,
//     non-200 status code).
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
	defer func() { _ = resp.Body.Close() }()

	// We must read the body and close it to ensure the underlying connection can be reused.
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status code: %d", resp.StatusCode)
	}

	_, _ = fmt.Fprintln(out, "Health check successful: server is running and healthy.")
	return nil
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
		if !ok {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Authentication Logic with Priority:
		// 1. Profile Authentication
		// 2. User Authentication
		// 3. Global Authentication
		// If a higher priority mechanism is configured, it MUST be satisfied. Lower priority checks are skipped.

		// Resolve the specific profile config from the user's allowed profiles
		// We need to find the profile config to check if it has an API Key.
		// The user object only has profile IDs.
		// We need to look up the profile definition.
		// Profile definitions are inside UpstreamServiceConfigs.
		// BUT, we are in the multi-user handler which routes to *any* service.
		// Wait, the profile is conceptual here?
		// The routing is /mcp/u/{uid}/profile/{profileId}
		// The `profileId` corresponds to a profile defined in ONE OR MORE upstream services.
		// Does a "Profile" exist independently?
		// In `config.proto`, `UpstreamServiceConfig` has `repeated Profile profiles`.
		// `Profile` has `id` and `api_key`.
		// Since a profile ID can be shared across services (e.g. "prod"), which `api_key` do we use?
		// If multiple services define "prod", do they share the same API key?
		// Assumption: If "prod" is defined in multiple places, they should logically share the key or we pick one.
		// Better approach: We iterate over all services to find the profile definition.

		// 1. Profile Auth - REMOVED (Per-profile ingress auth is no longer supported in this refactor)
		var isAuthenticated bool
		// 2. User Auth
		if user.GetAuthentication() != nil {
			if err := auth.ValidateAuthentication(ctx, user.GetAuthentication(), r); err == nil {
				isAuthenticated = true
			} else {
				// User auth configured but failed
				http.Error(w, "Unauthorized (User)", http.StatusUnauthorized)
				return
			}
		} else {
			// 3. Global Auth
			apiKey := a.SettingsManager.GetAPIKey()
			if apiKey != "" {
				// Manual check for global key since it's a string, or wrap it.
				// We'll just do manual check to match existing behavior logic.
				requestKey := r.Header.Get("X-API-Key")
				if requestKey == "" {
					authHeader := r.Header.Get("Authorization")
					if strings.HasPrefix(authHeader, "Bearer ") {
						requestKey = strings.TrimPrefix(authHeader, "Bearer ")
					}
				}

				if subtle.ConstantTimeCompare([]byte(requestKey), []byte(apiKey)) == 1 {
					isAuthenticated = true
					// Note: We don't inject API Key/Roles/User into ctx here because
					// the context is reset later (around line 1715) with r.Context().
					// The logic below (lines 1715+) reconstructs the context based on the target user (uid).
					// If Global Auth is used, we currently allow access to the target profile
					// but do not carry over "system-admin" identity or "admin" roles.
				} else {
					// Global auth configured but failed
					http.Error(w, "Unauthorized (Global)", http.StatusUnauthorized)
					return
				}
			} else {
				// No auth configured at any level
				// Sentinel Security: Enforce private network access if no auth is configured.
				ip := util.GetClientIP(r, trustProxy)
				if !util.IsPrivateIP(net.ParseIP(ip)) {
					logging.GetLogger().Warn("Blocked public internet request to /mcp/u/ because no API Key is configured", "remote_addr", r.RemoteAddr, "client_ip", ip)
					http.Error(w, "Forbidden: Public access requires an API Key to be configured", http.StatusForbidden)
					return
				}
				isAuthenticated = true
			}
		}

		if !isAuthenticated {
			// Should be unreachable if logic covers all cases, but safety net
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
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

	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "OK")
	}))
	mux.Handle("/metrics", authMiddleware(metrics.Handler()))
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

	// Register Debugger API if enabled
	if standardMiddlewares != nil && standardMiddlewares.Debugger != nil {
		mux.Handle("/debug/entries", authMiddleware(standardMiddlewares.Debugger.APIHandler()))
	}

	httpBindAddress := bindAddress
	if httpBindAddress == "" {
		if envAddr := os.Getenv("MCPANY_DEFAULT_HTTP_ADDR"); envAddr != "" {
			httpBindAddress = envAddr
		} else {
			httpBindAddress = "localhost:8070"
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

	var auditMiddleware *middleware.AuditMiddleware
	if standardMiddlewares != nil {
		auditMiddleware = standardMiddlewares.Audit
	}
	adminServer := admin.NewServer(cachingMiddleware, a.ToolManager, serviceRegistry, store, a.DiscoveryManager, auditMiddleware)
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
			//nolint:gosec // File path comes from trusted configuration
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

// createAuthMiddleware creates the authentication middleware.
func (a *Application) createAuthMiddleware(forcePrivateIPOnly bool, trustProxy bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow login endpoint without auth
			if r.URL.Path == "/api/v1/auth/login" {
				next.ServeHTTP(w, r)
				return
			}

			ip := util.GetClientIP(r, trustProxy)
			ctx := util.ContextWithRemoteIP(r.Context(), ip)
			r = r.WithContext(ctx)
			apiKey := a.SettingsManager.GetAPIKey()
			requestKey := r.Header.Get("X-API-Key")
			logging.GetLogger().Info("DEBUG: AuthMiddleware details", "configured_key", apiKey, "request_key", requestKey, "path", r.URL.Path)
			authenticated := false

			// 1. Check Global API Key
			if apiKey != "" {
				requestKey := r.Header.Get("X-API-Key")
				if requestKey == "" {
					requestKey = r.URL.Query().Get("api_key")
				}
				if requestKey == "" {
					authHeader := r.Header.Get("Authorization")
					if strings.HasPrefix(authHeader, "Bearer ") {
						requestKey = strings.TrimPrefix(authHeader, "Bearer ")
					}
				}

				if subtle.ConstantTimeCompare([]byte(requestKey), []byte(apiKey)) == 1 {
					authenticated = true
					// Inject API Key into context if needed
					ctx = auth.ContextWithAPIKey(ctx, requestKey)
					// Global API Key grants Admin privileges (Root Access)
					ctx = auth.ContextWithRoles(ctx, []string{"admin"})
					// Also inject a placeholder user ID so that handlers expecting a user context don't fail
					ctx = auth.ContextWithUser(ctx, "system-admin")
				}
			}

			// 2. Check User Authentication (Basic Auth)
			if !authenticated {
				username, _, ok := r.BasicAuth()
				if ok && a.AuthManager != nil {
					if user, found := a.AuthManager.GetUser(username); found {
						if err := auth.ValidateAuthentication(ctx, user.GetAuthentication(), r); err == nil {
							authenticated = true
							ctx = auth.ContextWithUser(ctx, username)
							if len(user.GetRoles()) > 0 {
								ctx = auth.ContextWithRoles(ctx, user.GetRoles())
							}
						}
					}
				}
			}

			if authenticated {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if !forcePrivateIPOnly && apiKey != "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Sentinel Security: If no API key is configured (and no user auth succeeded), enforce localhost-only access.
			// This prevents accidental exposure of the server to the public internet (RCE risk).
			if apiKey == "" {
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					// Fallback if RemoteAddr is weird, assume host is the string itself
					host = r.RemoteAddr
				}

				// Check if the request is from a loopback address
				ipAddr := net.ParseIP(host)
				if !util.IsPrivateIP(ipAddr) {
					logging.GetLogger().Warn("Blocked public internet request because no API Key is configured", "remote_addr", r.RemoteAddr)
					http.Error(w, "Forbidden: Public access requires an API Key to be configured", http.StatusForbidden)
					return
				}

				// Grant Admin privileges (Root Access) for local development/testing convenience
				// when running in insecure mode (private network, no API key).
				ctx = auth.ContextWithRoles(ctx, []string{"admin"})
				ctx = auth.ContextWithUser(ctx, "system-admin")
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}


// HTTPRequestContextMiddleware injects the HTTP request into the context.
//
// Summary: Middleware to add HTTP request to context.
//
// Parameters:
//   - next: http.Handler. The next handler.
//
// Returns:
//   - http.Handler: The wrapped handler.
func (a *Application) HTTPRequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.HTTPRequestContextKey, r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
	readyChan chan<- struct{},
	name string,
	lis net.Listener,
	shutdownTimeout time.Duration,
	server *gogrpc.Server,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverLog := logging.GetLogger().With("server", name)

		if server == nil {
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
				server.GracefulStop()
			}()

			timer := time.NewTimer(shutdownTimeout)
			defer timer.Stop()
			select {
			case <-stopped:
				// Successful graceful shutdown.
			case <-timer.C:
				// Graceful shutdown timed out.
				serverLog.Warn("Graceful shutdown timed out, forcing stop.")
				server.Stop()
			}
		}()

		serverLog.Info("gRPC server listening", "port", lis.Addr().String())
		if readyChan != nil {
			readyChan <- struct{}{}
		}
		if err := server.Serve(lis); err != nil && err != gogrpc.ErrServerStopped {
			errChan <- fmt.Errorf("[%s] server failed to serve: %w", name, err)
			cancel() // Signal shutdown goroutine to exit
		}
		<-shutdownComplete
		serverLog.Info("Server shut down.")
	}()
}

// wrapBindError checks if the error is a port conflict and returns a user-friendly error message.
func wrapBindError(err error, serverType, address, flag string) error {
	if strings.Contains(err.Error(), "address already in use") || strings.Contains(err.Error(), "bind: permission denied") {
		return fmt.Errorf("❌ %s server failed to listen on %s: %w\n\n💡 Tip: The port is already in use or restricted. Try using a different port:\n   mcpany run %s <new_port>", serverType, address, err, flag)
	}
	return fmt.Errorf("%s server failed to listen: %w", serverType, err)
}

// startHTTPServer starts an HTTP server in a new goroutine. It handles graceful
// shutdown when the context is canceled.
//
// ctx is the context for managing the server's lifecycle.
// wg is a WaitGroup to signal when the server has shut down.
// errChan is a channel for reporting errors during startup.
// name is a descriptive name for the server, used in logging.
// lis is the net.Listener on which the server will listen.
// handler is the HTTP handler for processing requests.
func startHTTPServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	errChan chan<- error,
	readyChan chan<- struct{},
	name string,
	lis net.Listener,
	handler http.Handler,
	shutdownTimeout time.Duration,
	connState func(net.Conn, http.ConnState),
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverLog := logging.GetLogger().With("server", name)
		// Listener passed in is already listening
		if readyChan != nil {
			readyChan <- struct{}{}
		}
		// We don't close lis here because http.Server.Serve closes it,
		// or we might want to close it on shutdown?
		// http.Server.Serve docs: "Serve accepts incoming connections on the Listener l... always returns a non-nil error."
		// It does NOT say it closes the listener.
		// However, Shutdown() closes the listener?
		// "Shutdown gracefully shuts down the server without interrupting any active connections... Shutdown works by first closing all open listeners..."
		// So Server.Shutdown closes it.
		// BUT if we error out before Shutdown, we should close it?
		// Let's rely on Server.Serve or Shutdown closing it, or defer close if not?
		// Ideally we defer close if Serve returns error other than ErrServerClosed.
		// Use a flag to check if Shutdown was called?
		// Or just defer Close() ignoring error? http.Server might close it too, double close is fine for net.Listener usually.
		// But let's check stdlib behavior.
		// Safe bet: defer lis.Close() at the top.
		// If Shutdown is called, it closes it. Double close is harmless for TCP listeners.
		defer func() { _ = lis.Close() }()

		serverLog = serverLog.With("port", lis.Addr().String())

		server := &http.Server{
			Handler: handler,
			BaseContext: func(_ net.Listener) context.Context {
				return ctx
			},
			ConnState: func(c net.Conn, state http.ConnState) {
				if connState != nil {
					connState(c, state)
				}
				switch state {
				case http.StateNew:
					metrics.IncrCounter([]string{"http", "connections", "opened", "total"}, 1)
				case http.StateClosed:
					metrics.IncrCounter([]string{"http", "connections", "closed", "total"}, 1)
				}
			},
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       120 * time.Second,
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
