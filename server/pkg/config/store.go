// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

// Engine defines the interface for configuration unmarshaling from different
// file formats. Implementations of this interface are responsible for parsing a
// byte slice and populating a protobuf message.
type Engine interface {
	// Unmarshal parses the given byte slice and populates the provided
	// proto.Message.
	Unmarshal(b []byte, v proto.Message) error
}

// NewEngine returns a configuration engine capable of unmarshaling the format
// indicated by the file extension of the given path. It supports `.json`,
// `.yaml`, `.yml`, and `.textproto` file formats.
//
// Parameters:
//   - path: The file path used to determine the configuration format.
//
// Returns an `Engine` implementation for the corresponding file format, or an
// error if the format is not supported.
func NewEngine(path string) (Engine, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return &jsonEngine{}, nil
	case ".yaml", ".yml":
		return &yamlEngine{}, nil
	case ".textproto", ".prototxt", ".pb", ".pb.txt":
		return &textprotoEngine{}, nil
	default:
		return nil, fmt.Errorf("unsupported config file extension '%s' for file %s", ext, path)
	}
}

// yamlEngine implements the Engine interface for YAML configuration files.
type yamlEngine struct{}

// Unmarshal parses a YAML byte slice into a `proto.Message`. It achieves this
// by first unmarshaling the YAML into a generic map, then marshaling that map
// to JSON, and finally unmarshaling the JSON into the target protobuf message.
// This two-step process is a common pattern for converting YAML to a protobuf
// message.
func (e *yamlEngine) Unmarshal(b []byte, v proto.Message) error {
	// First, unmarshal YAML into a generic map.
	var yamlMap map[string]interface{}
	if err := yaml.Unmarshal(b, &yamlMap); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Apply environment variable overrides: MCPANY__SECTION__KEY -> section.key
	// This allows overriding any configuration value using environment variables.
	applyEnvVars(yamlMap)

	// Helper to fix log level if it was set via env vars or file without prefix
	if gs, ok := yamlMap["global_settings"].(map[string]interface{}); ok {
		if ll, ok := gs["log_level"].(string); ok {
			if !strings.HasPrefix(ll, "LOG_LEVEL_") {
				gs["log_level"] = "LOG_LEVEL_" + ll
			}
		}
	}

	// Then, marshal the map to JSON. This is a common way to convert YAML to JSON.
	jsonData, err := json.Marshal(yamlMap)
	if err != nil {
		return fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	// Finally, unmarshal the JSON into the protobuf message.
	if err := protojson.Unmarshal(jsonData, v); err != nil {
		return err
	}
	// Debug logging to inspect unmarshaled user

	// Validate the unmarshaled message against the schema
	// We marshal back to JSON to ensure canonical format (camelCase) matches the schema generated from Go structs.
	canonicalJSON, err := protojson.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal proto for validation: %w", err)
	}
	var canonicalMap map[string]interface{}
	if err := json.Unmarshal(canonicalJSON, &canonicalMap); err != nil {
		return fmt.Errorf("failed to unmarshal canonical json: %w", err)
	}

	if err := ValidateConfigAgainstSchema(canonicalMap); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}

// textprotoEngine implements the Engine interface for textproto configuration
// files.
type textprotoEngine struct{}

// Unmarshal parses a textproto byte slice into a `proto.Message`.
func (e *textprotoEngine) Unmarshal(b []byte, v proto.Message) error {
	return prototext.Unmarshal(b, v)
}

// jsonEngine implements the Engine interface for JSON configuration files.
type jsonEngine struct{}

// Unmarshal parses a JSON byte slice into a `proto.Message`.
func (e *jsonEngine) Unmarshal(b []byte, v proto.Message) error {
	return protojson.Unmarshal(b, v)
}

// Store defines the interface for loading MCP-X server configurations.
// Implementations of this interface provide a way to retrieve the complete
// server configuration from a source, such as a file or a remote service.
type Store interface {
	// Load retrieves and returns the McpAnyServerConfig.
	Load() (*configv1.McpAnyServerConfig, error)
}

// ServiceStore extends Store to provide CRUD operations for UpstreamServices.
type ServiceStore interface {
	Store
	// SaveService saves or updates a service configuration.
	//
	// Parameters:
	//   - service: The service configuration to save.
	//
	// Returns an error if the operation fails.
	SaveService(service *configv1.UpstreamServiceConfig) error

	// GetService retrieves a service configuration by name.
	//
	// Parameters:
	//   - name: The name of the service to retrieve.
	//
	// Returns the service configuration, or an error if not found or the operation fails.
	GetService(name string) (*configv1.UpstreamServiceConfig, error)

	// ListServices retrieves all stored service configurations.
	//
	// Returns a slice of service configurations, or an error if the operation fails.
	ListServices() ([]*configv1.UpstreamServiceConfig, error)

	// DeleteService removes a service configuration by name.
	//
	// Parameters:
	//   - name: The name of the service to delete.
	//
	// Returns an error if the operation fails.
	DeleteService(name string) error
}

var envVarRegex = regexp.MustCompile(`\${([^{}]+)}`)

func expand(b []byte) []byte {
	return envVarRegex.ReplaceAllFunc(b, func(match []byte) []byte {
		s := string(match[2 : len(match)-1])
		parts := strings.SplitN(s, ":", 2)
		varName := parts[0]
		val, ok := os.LookupEnv(varName)
		if ok && val != "" {
			return []byte(val)
		}
		if len(parts) > 1 {
			return []byte(parts[1])
		}
		return match
	})
}

// FileStore implements the `Store` interface for loading configurations from one
// or more files or directories on a filesystem. It supports multiple file
// formats (JSON, YAML, and textproto) and merges the configurations into a
// single `McpAnyServerConfig`.
type FileStore struct {
	fs    afero.Fs
	paths []string
}

// NewFileStore creates a new FileStore with the given filesystem and a list of
// paths to load configurations from.
//
// Parameters:
//   - fs: The filesystem interface to use for file operations.
//   - paths: A slice of file or directory paths to scan for configuration
//     files.
//
// Returns a new instance of `FileStore`.
func NewFileStore(fs afero.Fs, paths []string) *FileStore {
	return &FileStore{fs: fs, paths: paths}
}

// Load scans the configured paths for supported configuration files (JSON,
// YAML, and textproto), reads them, unmarshals their contents, and merges them
// into a single `McpAnyServerConfig`.
//
// The files are processed in alphabetical order, and configurations from later
// files are merged into earlier ones. This allows for a cascading configuration
// setup where base configurations can be overridden by more specific ones.
//
// Returns the merged `McpAnyServerConfig` or an error if any part of the process
// fails.
func (s *FileStore) Load() (*configv1.McpAnyServerConfig, error) {
	var mergedConfig *configv1.McpAnyServerConfig

	filePaths, err := s.collectFilePaths()
	if err != nil {
		return nil, fmt.Errorf("failed to collect config file paths: %w", err)
	}

	for _, path := range filePaths {
		var b []byte
		var err error
		if isURL(path) {
			b, err = readURL(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read config from URL %s: %w", path, err)
			}
		} else {
			b, err = afero.ReadFile(s.fs, path)
			if err != nil {
				return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
			}
		}

		if len(b) == 0 {
			continue
		}

		b = expand(b)

		engine, err := NewEngine(path)
		if err != nil {
			return nil, err
		}

		cfg := &configv1.McpAnyServerConfig{}
		if err := engine.Unmarshal(b, cfg); err != nil {
			if strings.Contains(err.Error(), "is already set") {
				// Find the service name
				var raw map[string]interface{}
				if yaml.Unmarshal(b, &raw) == nil {
					if services, ok := raw["upstream_services"].([]interface{}); ok {
						for _, s := range services {
							if service, ok := s.(map[string]interface{}); ok {
								if name, ok := service["name"].(string); ok {
									// Heuristic: if the raw service definition has more than one service type key, it's the culprit
									keys := 0
									serviceKeys := []string{"http_service", "grpc_service", "openapi_service", "command_line_service", "websocket_service", "webrtc_service", "graphql_service", "mcp_service"}
									for _, k := range serviceKeys {
										if _, ok := service[k]; ok {
											keys++
										}
									}
									if keys > 1 {
										return nil, fmt.Errorf("failed to unmarshal config from %s: service %q has multiple service types defined", path, name)
									}
								}
							}
						}
					}
				}
			}
			return nil, fmt.Errorf("failed to unmarshal config from %s: %w", path, err)
		}

		if mergedConfig == nil {
			mergedConfig = cfg
		} else {
			proto.Merge(mergedConfig, cfg)
		}
	}

	return mergedConfig, nil
}

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			ips, err := net.LookupIP(host)
			if err != nil {
				return nil, err
			}

			var dialAddr string
			for _, ip := range ips {
				if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() {
					continue
				}
				// Use the first valid IP address for the connection.
				if dialAddr == "" {
					dialAddr = net.JoinHostPort(ip.String(), port)
				}
			}

			if dialAddr == "" {
				return nil, fmt.Errorf("no valid IP address found for host: %s", host)
			}

			return (&net.Dialer{}).DialContext(ctx, network, dialAddr)
		},
	},
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func readURL(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for url %s: %w", url, err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get config from url %s: %w", url, err)
	}

	defer func() { _ = resp.Body.Close() }()

	// Since redirects are disabled, a redirect attempt will result in a 3xx status code.
	if resp.StatusCode >= 300 && resp.StatusCode <= 399 {
		return nil, fmt.Errorf("redirects are disabled for security reasons")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get config from url %s: status code %d", url, resp.StatusCode)
	}

	// Limit the size of the response to 1MB to prevent DoS attacks.
	resp.Body = http.MaxBytesReader(nil, resp.Body, 1024*1024)
	return io.ReadAll(resp.Body)
}

// collectFilePaths recursively scans the configured paths and returns a list of valid config files.
func (s *FileStore) collectFilePaths() ([]string, error) {
	var files []string
	for _, path := range s.paths {
		if isURL(path) {
			files = append(files, path)
			continue
		}
		info, err := s.fs.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to stat path %s: %w", path, err)
		}

		if info.IsDir() {
			err := afero.Walk(s.fs, path, func(p string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !fi.IsDir() {
					if _, err := NewEngine(p); err == nil {
						files = append(files, p)
					}
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("failed to walk directory %s: %w", path, err)
			}
		} else {
			if _, err := NewEngine(path); err == nil {
				files = append(files, path)
			}
		}
	}
	sort.Strings(files)
	return files, nil
}

func isURL(path string) bool {
	return strings.HasPrefix(strings.ToLower(path), "http://") || strings.HasPrefix(strings.ToLower(path), "https://")
}

// applyEnvVars iterates over environment variables and applies those starting with "MCPANY__"
// to the configuration map. It supports nested structure via "__" separator.
// Example: MCPANY__GLOBAL_SETTINGS__MCP_LISTEN_ADDRESS -> global_settings.mcp_listen_address
func applyEnvVars(m map[string]interface{}) {
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "MCPANY__") {
			continue
		}
		// Split into key and value
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		// Remove prefix and split by double underscore
		trimmedKey := strings.TrimPrefix(key, "MCPANY__")
		path := strings.Split(trimmedKey, "__")

		// Walk the map and create nested maps as needed
		current := m
		for i, originalSection := range path {
			section := strings.ToLower(originalSection) // Normalize to snake_case/lowercase
			if i == len(path)-1 {
				// We are at the leaf, set the value.
				// We overwrite whatever is there.
				current[section] = value
			} else {
				// We need to go deeper
				if next, ok := current[section].(map[string]interface{}); ok {
					current = next
				} else {
					// Create new map if it doesn't exist or isn't a map
					next := make(map[string]interface{})
					current[section] = next
					current = next
				}
			}
		}
	}
}

// MultiStore implements the Store interface for loading configurations from multiple stores.
// It merges the configurations in the order the stores are provided.
type MultiStore struct {
	stores []Store
}

// NewMultiStore creates a new MultiStore with the given stores.
func NewMultiStore(stores ...Store) *MultiStore {
	return &MultiStore{stores: stores}
}

// Load loads configurations from all stores and merges them into a single config.
func (ms *MultiStore) Load() (*configv1.McpAnyServerConfig, error) {
	mergedConfig := &configv1.McpAnyServerConfig{}
	for _, s := range ms.stores {
		cfg, err := s.Load()
		if err != nil {
			return nil, err
		}
		if cfg != nil {
			proto.Merge(mergedConfig, cfg)
		}
	}
	return mergedConfig, nil
}
