// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/spf13/afero"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
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
		// Enhance error message for common YAML mistakes
		if strings.Contains(err.Error(), "found character that cannot start any token") {
			if bytes.Contains(b, []byte("\t")) {
				// revive:disable-next-line:error-strings // This error message is user facing and needs to be descriptive
				//nolint:staticcheck // This error message is user facing and needs to be descriptive
				return fmt.Errorf("failed to unmarshal YAML: %w\n\nHint: YAML files cannot contain tabs. Please use spaces for indentation.", err)
			}
		}
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Apply environment variable overrides: MCPANY__SECTION__KEY -> section.key
	// This allows overriding any configuration value using environment variables.
	applyEnvVars(yamlMap, v)

	// Helper to fix log level if it was set via env vars or file without prefix
	if gs, ok := yamlMap["global_settings"].(map[string]interface{}); ok {
		if ll, ok := gs["log_level"].(string); ok {
			if !strings.HasPrefix(ll, "LOG_LEVEL_") {
				gs["log_level"] = "LOG_LEVEL_" + strings.ToUpper(ll)
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
		// Detect if the user is using Claude Desktop config format
		if strings.Contains(err.Error(), "unknown field \"mcpServers\"") {
			// revive:disable-next-line:error-strings // This error message is user facing and needs to be descriptive
			//nolint:staticcheck // This error message is user facing and needs to be descriptive
			return fmt.Errorf("%w\n\nDid you mean \"upstream_services\"? It looks like you might be using a Claude Desktop configuration format. MCP Any uses a different configuration structure. See documentation for details.", err)
		}

		// Detect if the user is using "services" which is a common alias for "upstream_services"
		if strings.Contains(err.Error(), "unknown field \"services\"") {
			// revive:disable-next-line:error-strings // This error message is user facing and needs to be descriptive
			//nolint:staticcheck // This error message is user facing and needs to be descriptive
			return fmt.Errorf("%w\n\nDid you mean \"upstream_services\"? \"services\" is not a valid top-level key.", err)
		}

		// Detect invalid use of service_config wrapper (common mistake due to old docs)
		if strings.Contains(err.Error(), "unknown field \"service_config\"") {
			// revive:disable-next-line:error-strings // This error message is user facing and needs to be descriptive
			//nolint:staticcheck // This error message is user facing and needs to be descriptive
			return fmt.Errorf("%w\n\nIt looks like you are using 'service_config' as a wrapper key. In MCP Any configuration, you should place the service type (e.g., 'http_service', 'grpc_service') directly under the service definition, without a 'service_config' wrapper.", err)
		}

		// Check for unknown fields and suggest fuzzy matches
		if strings.Contains(err.Error(), "unknown field") {
			matches := unknownFieldRegex.FindStringSubmatch(err.Error())
			if len(matches) > 1 {
				unknownField := matches[1]
				suggestion := suggestFix(unknownField, v)
				if suggestion != "" {
					return fmt.Errorf("%w\n\n%s", err, suggestion)
				}
			}
		}
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
//
// b is the b.
// v is the v.
//
// Returns an error if the operation fails.
func (e *textprotoEngine) Unmarshal(b []byte, v proto.Message) error {
	return prototext.Unmarshal(b, v)
}

// jsonEngine implements the Engine interface for JSON configuration files.
type jsonEngine struct{}

// Unmarshal parses a JSON byte slice into a `proto.Message`.
//
// b is the b.
// v is the v.
//
// Returns an error if the operation fails.
func (e *jsonEngine) Unmarshal(b []byte, v proto.Message) error {
	if err := protojson.Unmarshal(b, v); err != nil {
		// Detect if the user is using Claude Desktop config format
		if strings.Contains(err.Error(), "unknown field \"mcpServers\"") {
			// revive:disable-next-line:error-strings // This error message is user facing and needs to be descriptive
			//nolint:staticcheck // This error message is user facing and needs to be descriptive
			return fmt.Errorf("%w\n\nDid you mean \"upstream_services\"? It looks like you might be using a Claude Desktop configuration format. MCP Any uses a different configuration structure. See documentation for details.", err)
		}

		// Detect if the user is using "services" which is a common alias for "upstream_services"
		if strings.Contains(err.Error(), "unknown field \"services\"") {
			// revive:disable-next-line:error-strings // This error message is user facing and needs to be descriptive
			//nolint:staticcheck // This error message is user facing and needs to be descriptive
			return fmt.Errorf("%w\n\nDid you mean \"upstream_services\"? \"services\" is not a valid top-level key.", err)
		}

		// Check for unknown fields and suggest fuzzy matches
		if strings.Contains(err.Error(), "unknown field") {
			matches := unknownFieldRegex.FindStringSubmatch(err.Error())
			if len(matches) > 1 {
				unknownField := matches[1]
				suggestion := suggestFix(unknownField, v)
				if suggestion != "" {
					return fmt.Errorf("%w\n\n%s", err, suggestion)
				}
			}
		}
		return err
	}
	return nil
}

// Store defines the interface for loading MCP-X server configurations.
// Implementations of this interface provide a way to retrieve the complete
// server configuration from a source, such as a file or a remote service.
type Store interface {
	// Load retrieves and returns the McpAnyServerConfig.
	//
	// ctx is the context for the request.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	Load(ctx context.Context) (*configv1.McpAnyServerConfig, error)

	// HasConfigSources returns true if the store has configuration sources (e.g., file paths) configured.
	//
	// Returns true if successful.
	HasConfigSources() bool
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
	SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error

	// GetService retrieves a service configuration by name.
	//
	// Parameters:
	//   - name: The name of the service to retrieve.
	//
	// Returns the service configuration, or an error if not found or the operation fails.
	GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error)

	// ListServices retrieves all stored service configurations.
	//
	// Returns a slice of service configurations, or an error if the operation fails.
	ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error)

	// DeleteService removes a service configuration by name.
	//
	// Parameters:
	//   - name: The name of the service to delete.
	//
	// Returns an error if the operation fails.
	DeleteService(ctx context.Context, name string) error
}

var envVarRegex = regexp.MustCompile(`\${([^{}]+)}`)
var unknownFieldRegex = regexp.MustCompile(`unknown field "([^"]+)"`)

// expand replaces ${VAR} or ${VAR:default} with environment variables.
// If a variable is missing and no default is provided, it returns an error.
func expand(b []byte) ([]byte, error) {
	var missingVars []string
	expanded := envVarRegex.ReplaceAllFunc(b, func(match []byte) []byte {
		s := string(match[2 : len(match)-1])
		parts := strings.SplitN(s, ":", 2)
		varName := parts[0]
		val, ok := os.LookupEnv(varName)
		if ok {
			if val == "" && len(parts) > 1 {
				return []byte(parts[1])
			}
			return []byte(val)
		}
		if len(parts) > 1 {
			return []byte(parts[1])
		}
		missingVars = append(missingVars, varName)
		return match
	})

	if len(missingVars) > 0 {
		return nil, fmt.Errorf("missing environment variables: %v", missingVars)
	}

	return expanded, nil
}

// FileStore implements the `Store` interface for loading configurations from one
// or more files or directories on a filesystem. It supports multiple file
// formats (JSON, YAML, and textproto) and merges the configurations into a
// single `McpAnyServerConfig`.
type FileStore struct {
	fs         afero.Fs
	paths      []string
	skipErrors bool
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

// NewFileStoreWithSkipErrors creates a new FileStore that skips malformed config files
// instead of returning an error.
func NewFileStoreWithSkipErrors(fs afero.Fs, paths []string) *FileStore {
	return &FileStore{fs: fs, paths: paths, skipErrors: true}
}

// HasConfigSources returns true if the store has configuration paths configured.
//
// Returns true if successful.
func (s *FileStore) HasConfigSources() bool {
	return len(s.paths) > 0
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
func (s *FileStore) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	var mergedConfig *configv1.McpAnyServerConfig

	filePaths, err := s.collectFilePaths()
	if err != nil {
		return nil, fmt.Errorf("failed to collect config file paths: %w", err)
	}

	for _, path := range filePaths {
		var b []byte
		var err error
		if isURL(path) {
			b, err = readURL(ctx, path)
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

		b, err = expand(b)
		if err != nil {
			return nil, fmt.Errorf("failed to expand environment variables in %s: %w", path, err)
		}

		engine, err := NewEngine(path)
		if err != nil {
			if s.skipErrors {
				logging.GetLogger().Error("Failed to determine config engine, skipping file", "path", path, "error", err)
				continue
			}
			return nil, err
		}

		cfg := &configv1.McpAnyServerConfig{}
		if err := engine.Unmarshal(b, cfg); err != nil {
			logErr := fmt.Errorf("failed to unmarshal config from %s: %w", path, err)
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
										logErr = fmt.Errorf("failed to unmarshal config from %s: service %q has multiple service types defined", path, name)
									}
								}
							}
						}
					}
				}
			}
			if s.skipErrors {
				logging.GetLogger().Error("Failed to parse config file, skipping", "path", path, "error", logErr)
				continue
			}
			return nil, logErr
		}

		if mergedConfig == nil {
			mergedConfig = cfg
		} else {
			proto.Merge(mergedConfig, cfg)
		}
	}

	return mergedConfig, nil
}

// httpClient is a safe http client for loading configurations.
// It uses SafeDialer to prevent SSRF by blocking access to private and link-local IPs.
// It also disables redirects.
var httpClient = func() *http.Client {
	client := util.NewSafeHTTPClient()
	client.Timeout = 5 * time.Second
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return client
}()

func readURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
// Example: MCPANY__GLOBAL_SETTINGS__MCP_LISTEN_ADDRESS -> global_settings.mcp_listen_address.
func applyEnvVars(m map[string]interface{}, v proto.Message) {
	applyEnvVarsFromSlice(m, os.Environ(), v)
	if v != nil {
		fixTypes(m, v.ProtoReflect().Descriptor())
	}
}

// applyEnvVarsFromSlice is the logic for applyEnvVars, separated for testing.
func applyEnvVarsFromSlice(m map[string]interface{}, environ []string, v proto.Message) {
	// Sort the environment variables to ensure deterministic application order.
	// We make a copy to avoid modifying the input slice.
	sortedEnv := make([]string, len(environ))
	copy(sortedEnv, environ)
	sort.Strings(sortedEnv)

	for _, env := range sortedEnv {
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
				resolvedValue := resolveEnvValue(v, path, value)
				current[section] = resolvedValue
			} else {
				// We need to go deeper
				if next, ok := current[section].(map[string]interface{}); ok {
					current = next
				} else if slice, ok := current[section].([]interface{}); ok {
					// It is a slice, convert to map keyed by index to support merging
					next := make(map[string]interface{})
					for idx, val := range slice {
						next[strconv.Itoa(idx)] = val
					}
					current[section] = next
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

// resolveEnvValue attempts to determine the correct type for the environment variable
// by traversing the protobuf message descriptor.
func resolveEnvValue(root proto.Message, path []string, value string) interface{} {
	if root == nil {
		return value
	}
	md := root.ProtoReflect().Descriptor()
	var currentFd protoreflect.FieldDescriptor

	for i := 0; i < len(path); i++ {
		part := strings.ToLower(path[i])

		if currentFd != nil && currentFd.IsList() {
			// We are inside a list. 'part' should be an index.
			// If element is Message, we switch 'md' to that message.
			// If element is Scalar, we are done (or expecting this to be leaf).

			if currentFd.Kind() == protoreflect.MessageKind {
				md = currentFd.Message()
				currentFd = nil // Reset, we are now inside the message
				continue
			}
			// Scalar list. 'part' is the index.
			// If this is the last part, we are setting the value of this element.
			if i == len(path)-1 {
				// Convert value based on currentFd.Kind()
				return convertKind(currentFd.Kind(), value)
			}
			// If not last part, mismatch? Scalar list doesn't have fields.
			return value
		}

		fd := findField(md, part)
		if fd == nil {
			// Can't resolve, return string
			return value
		}
		currentFd = fd

		if i == len(path)-1 {
			// We found the leaf field. Check its kind.
			kind := fd.Kind()

			if fd.IsList() {
				// Repeated field: split by comma
				r := csv.NewReader(strings.NewReader(value))
				r.TrimLeadingSpace = true
				parts, err := r.Read()
				if err != nil {
					// Fallback to simple split if CSV parsing fails
					parts = strings.Split(value, ",")
				}
				var list []interface{}
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if kind == protoreflect.BoolKind {
						b, err := strconv.ParseBool(part)
						if err == nil {
							list = append(list, b)
						} else {
							list = append(list, part)
						}
					} else {
						list = append(list, part)
					}
				}
				return list
			}

			if kind == protoreflect.BoolKind {
				b, err := strconv.ParseBool(value)
				if err == nil {
					return b
				}
			}
			return convertKind(fd.Kind(), value)
		}

		// Navigate deeper
		switch {
		case fd.Kind() == protoreflect.MessageKind:
			if fd.IsList() {
				// Next iteration will handle index
				continue
			}
			md = fd.Message()
			currentFd = nil
		case fd.IsList():
			// Scalar list. Next iteration will handle index.
			continue
		default:
			// Path continues but field is not a message? mismatch.
			return value
		}
	}
	return value
}

func convertKind(kind protoreflect.Kind, value string) interface{} {
	if kind == protoreflect.BoolKind {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	// For numbers, we can also convert, but strings are accepted by protojson for numbers.
	// However, if we want strict typing for ints/floats, we could do it here.
	// protojson is usually permissive with strings for numbers.
	return value
}

// fixTypes traverses the map and converts maps to slices where appropriate based on the protobuf schema.
// This is necessary because environment variables might create maps for list fields (using indices as keys),
// but protojson expects slices.
func fixTypes(m map[string]interface{}, md protoreflect.MessageDescriptor) {
	for key, val := range m {
		fd := findField(md, key)
		if fd == nil {
			continue
		}

		if fd.IsList() {
			// If it's a map, convert to slice
			if valMap, ok := val.(map[string]interface{}); ok {
				newSlice := convertMapToSlice(valMap)
				m[key] = newSlice
				// Re-assign val for recursive step
				val = newSlice
			}

			// If it is a slice (either originally or converted), recurse if it contains messages
			if valSlice, ok := val.([]interface{}); ok {
				if fd.Kind() == protoreflect.MessageKind {
					msgDesc := fd.Message()
					for _, item := range valSlice {
						if itemMap, ok := item.(map[string]interface{}); ok {
							fixTypes(itemMap, msgDesc)
						}
					}
				}
			}
		} else if fd.Kind() == protoreflect.MessageKind && !fd.IsMap() {
			// Recurse
			if valMap, ok := val.(map[string]interface{}); ok {
				fixTypes(valMap, fd.Message())
			}
		}
	}
}

func convertMapToSlice(m map[string]interface{}) []interface{} {
	type entry struct {
		idx int
		val interface{}
	}
	var entries []entry
	for k, v := range m {
		idx, err := strconv.Atoi(k)
		if err == nil {
			entries = append(entries, entry{idx, v})
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].idx < entries[j].idx })

	res := make([]interface{}, len(entries))
	for i, e := range entries {
		res[i] = e.val
	}
	return res
}

func findField(md protoreflect.MessageDescriptor, name string) protoreflect.FieldDescriptor {
	// Try ByName (snake_case usually)
	fd := md.Fields().ByName(protoreflect.Name(name))
	if fd != nil {
		return fd
	}
	// Try ByJSONName (camelCase)
	fd = md.Fields().ByJSONName(name)
	if fd != nil {
		return fd
	}
	return nil
}

// MultiStore implements the Store interface for loading configurations from multiple stores.
// It merges the configurations in the order the stores are provided.
type MultiStore struct {
	stores []Store
}

// NewMultiStore creates a new MultiStore with the given stores.
//
// stores is the stores.
//
// Returns the result.
func NewMultiStore(stores ...Store) *MultiStore {
	return &MultiStore{stores: stores}
}

// Load loads configurations from all stores and merges them into a single config.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns an error if the operation fails.
func (ms *MultiStore) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	mergedConfig := &configv1.McpAnyServerConfig{}
	for _, s := range ms.stores {
		cfg, err := s.Load(ctx)
		if err != nil {
			return nil, err
		}
		if cfg != nil {
			proto.Merge(mergedConfig, cfg)
		}
	}
	return mergedConfig, nil
}

// levenshtein calculates the Levenshtein distance between two strings.
func levenshtein(s, t string) int {
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
	}
	for i := 0; i <= len(s); i++ {
		d[i][0] = i
	}
	for j := 0; j <= len(t); j++ {
		d[0][j] = j
	}
	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				minVal := d[i-1][j]
				if d[i][j-1] < minVal {
					minVal = d[i][j-1]
				}
				if d[i-1][j-1] < minVal {
					minVal = d[i-1][j-1]
				}
				d[i][j] = minVal + 1
			}
		}
	}
	return d[len(s)][len(t)]
}

// suggestFix finds the closest matching field name in the proto message.
func suggestFix(unknownField string, root proto.Message) string {
	candidates := make(map[string]struct{})
	collectFieldNames(root.ProtoReflect().Descriptor(), candidates)

	// Explicitly add fields from common nested configuration objects to the candidates.
	// We avoid full recursion to prevent suggesting fields from obscure/irrelevant parts of the schema
	// (like "services" from Collection which confuses users when they mean "upstream_services").
	commonMessages := []proto.Message{
		&configv1.GlobalSettings{},
		&configv1.UpstreamServiceConfig{},
		&configv1.HttpUpstreamService{},
		&configv1.GrpcUpstreamService{},
		&configv1.McpUpstreamService{},
		&configv1.OpenapiUpstreamService{},
		&configv1.CommandLineUpstreamService{},
		&configv1.SqlUpstreamService{},
		&configv1.Authentication{},
	}

	for _, msg := range commonMessages {
		collectFieldNames(msg.ProtoReflect().Descriptor(), candidates)
	}

	bestMatch := ""
	minDist := 100

	for name := range candidates {
		dist := levenshtein(unknownField, name)
		if dist < minDist {
			minDist = dist
			bestMatch = name
		}
	}

	// Only suggest if it's reasonably close.
	// We allow up to 3 edits for short strings, or more for longer strings (50% rule).
	limit := len(unknownField) / 2
	if limit < 3 {
		limit = 3
	}

	if minDist <= limit {
		return fmt.Sprintf("Did you mean %q?", bestMatch)
	}
	return ""
}

func collectFieldNames(md protoreflect.MessageDescriptor, candidates map[string]struct{}) {
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		candidates[string(fd.Name())] = struct{}{}
		candidates[fd.JSONName()] = struct{}{}
	}
}

// HasConfigSources returns true if any of the underlying stores have configuration sources.
//
// Returns true if successful.
func (ms *MultiStore) HasConfigSources() bool {
	for _, s := range ms.stores {
		if s.HasConfigSources() {
			return true
		}
	}
	return false
}
