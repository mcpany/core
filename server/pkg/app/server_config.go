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

// ReloadConfig reloads the configuration from the given paths and updates the
// services.
//
// Summary: Reloads application configuration from disk/storage.
//
// Parameters:
//   - ctx (context.Context): The context for the reload operation.
//   - fs (afero.Fs): The filesystem interface for reading configuration files.
//   - configPaths ([]string): A slice of paths to configuration files to reload.
//
// Returns:
//   - (error): An error if the configuration reload fails.
//
// Side Effects:
//   - Reads configuration files.
//   - Updates global settings, user auth, profiles, and service registry.
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

	// Update log level
	if cfg.GetGlobalSettings().GetLogLevel() != 0 {
		newLevel := logging.ToSlogLevel(cfg.GetGlobalSettings().GetLogLevel())
		logging.SetLevel(newLevel)
		log.Info("Updated log level", "level", newLevel)
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
//
//nolint:gocyclo // complexity is fine here
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

	// Helper to deduplicate tools by name after proto.Merge appends them
	dedupTools := func(svc *config_v1.UpstreamServiceConfig) {
		if cmd := svc.GetCommandLineService(); cmd != nil {
			seen := make(map[string]bool)
			var deduplicated []*config_v1.ToolDefinition
			tools := cmd.GetTools()
			// Iterate backwards so the latter tools (from database override) take precedence
			for i := len(tools) - 1; i >= 0; i-- {
				t := tools[i]
				if !seen[t.GetName()] {
					seen[t.GetName()] = true
					// Prepend to keep original order with overrides
					deduplicated = append([]*config_v1.ToolDefinition{t}, deduplicated...)
				}
			}
			cmd.SetTools(deduplicated)
		}

		if http := svc.GetHttpService(); http != nil {
			seen := make(map[string]bool)
			var deduplicated []*config_v1.ToolDefinition
			tools := http.GetTools()
			for i := len(tools) - 1; i >= 0; i-- {
				t := tools[i]
				if !seen[t.GetName()] {
					seen[t.GetName()] = true
					deduplicated = append([]*config_v1.ToolDefinition{t}, deduplicated...)
				}
			}
			http.SetTools(deduplicated)
		}
	}
	if cfg.GetUpstreamServices() != nil {
		for _, svc := range cfg.GetUpstreamServices() {
			if !svc.GetDisable() {
				if existing, ok := newServices[svc.GetName()]; ok {
					proto.Merge(existing, svc)
					dedupTools(existing)
				} else {
					s := proto.Clone(svc).(*config_v1.UpstreamServiceConfig)
					dedupTools(s)
					newServices[svc.GetName()] = s
				}
			}
		}
	}
	if cfg.GetCollections() != nil {
		for _, collection := range cfg.GetCollections() {
			for _, svc := range collection.GetServices() {
				if svc.GetName() == "" {
					continue
				}
				if !svc.GetDisable() {
					if existing, ok := newServices[svc.GetName()]; ok {
						proto.Merge(existing, svc)
						dedupTools(existing)
					} else {
						s := proto.Clone(svc).(*config_v1.UpstreamServiceConfig)
						dedupTools(s)
						newServices[svc.GetName()] = s
					}
				}
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
