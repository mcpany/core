// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"context"
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

// Engine defines the interface for configuration unmarshaling from different file formats.
//
// Summary: Abstraction for parsing configuration files into protobuf messages.
type Engine interface {
	// Unmarshal parses the given byte slice and populates the provided proto.Message.
	//
	// Summary: Parses bytes into a protobuf message.
	//
	// Parameters:
	//   - b: []byte. The raw bytes to parse.
	//   - v: proto.Message. The destination protobuf message.
	//
	// Returns:
	//   - error: An error if parsing fails.
	Unmarshal(b []byte, v proto.Message) error
}

// StructuredEngine defines an interface for engines that can unmarshal directly from a map structure.
//
// Summary: Abstraction for parsing configurations from map structures, avoiding double parsing.
type StructuredEngine interface {
	Engine
	// UnmarshalFromMap populates the provided proto.Message from a raw map.
	//
	// Summary: Parses a map into a protobuf message.
	//
	// Parameters:
	//   - m: map[string]interface{}. The raw map data.
	//   - v: proto.Message. The destination protobuf message.
	//   - originalBytes: []byte. Optional original bytes for error reporting (line numbers).
	//
	// Returns:
	//   - error: An error if parsing fails.
	UnmarshalFromMap(m map[string]interface{}, v proto.Message, originalBytes []byte) error
}

// ConfigurableEngine defines an interface for engines that support configuration options.
//
// Summary: Interface for engines that can be configured (e.g. skip validation).
type ConfigurableEngine interface {
	Engine
	// SetSkipValidation sets whether to skip schema validation.
	//
	// Summary: Configures the engine to skip schema validation.
	//
	// Parameters:
	//   - skip: bool. True to skip validation.
	SetSkipValidation(skip bool)
}

// NewEngine returns a configuration engine capable of unmarshaling the format indicated by the file extension.
//
// Summary: Factory function to create the appropriate Engine for a given file path.
//
// Parameters:
//   - path: string. The file path used to determine the configuration format.
//
// Returns:
//   - Engine: An initialized Engine implementation.
//   - error: An error if the file extension is not supported.
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
type yamlEngine struct {
	skipValidation bool
}

// SetSkipValidation sets whether to skip schema validation.
//
// Summary: sets whether to skip schema validation.
//
// Parameters:
//   - skip: bool. The skip.
//
// Returns:
//   None.
func (e *yamlEngine) SetSkipValidation(skip bool) {
	e.skipValidation = skip
}

// Unmarshal parses a YAML byte slice into a `proto.Message`.
//
// Summary: parses a YAML byte slice into a `proto.Message`.
//
// Parameters:
//   - b: []byte. The b.
//   - v: proto.Message. The v.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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

	return e.unmarshalInternal(yamlMap, v, b)
}

// UnmarshalFromMap populates the provided proto.Message from a raw map.
//
// Summary: populates the provided proto.Message from a raw map.
//
// Parameters:
//   - yamlMap: map[string]interface{}. The yamlMap.
//   - v: proto.Message. The v.
//   - originalBytes: []byte. The originalBytes.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (e *yamlEngine) UnmarshalFromMap(yamlMap map[string]interface{}, v proto.Message, originalBytes []byte) error {
	return e.unmarshalInternal(yamlMap, v, originalBytes)
}

func (e *yamlEngine) unmarshalInternal(yamlMap map[string]interface{}, v proto.Message, originalBytes []byte) error {
	// Apply environment variable overrides: MCPANY__SECTION__KEY -> section.key
	// This allows overriding any configuration value using environment variables.
	applyEnvVarsFromSlice(yamlMap, os.Environ(), v)

	// Apply --set overrides: section.key=value or section[idx].key=value
	applySetOverrides(yamlMap, GlobalSettings().SetValues(), v)

	if v != nil {
		fixTypes(yamlMap, v.ProtoReflect().Descriptor())
	}

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
		// Attempt to find the line number in the original YAML
		if originalBytes != nil {
			if matches := unknownFieldRegex.FindStringSubmatch(err.Error()); len(matches) > 1 {
				unknownField := matches[1]
				if line := findKeyLine(originalBytes, unknownField); line > 0 {
					err = fmt.Errorf("line %d: %w", line, err)
				}
			}
		}

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

	if !e.skipValidation {
		if err := ValidateConfigAgainstSchema(canonicalMap); err != nil {
			return fmt.Errorf("schema validation failed: %w", err)
		}
	}

	return nil
}

// textprotoEngine implements the Engine interface for textproto configuration files.
type textprotoEngine struct{}

// Unmarshal parses a textproto byte slice into a `proto.Message`.
//
// Summary: parses a textproto byte slice into a `proto.Message`.
//
// Parameters:
//   - b: []byte. The b.
//   - v: proto.Message. The v.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (e *textprotoEngine) Unmarshal(b []byte, v proto.Message) error {
	return prototext.Unmarshal(b, v)
}

// jsonEngine implements the Engine interface for JSON configuration files.
type jsonEngine struct{}

// Unmarshal parses a JSON byte slice into a `proto.Message`.
//
// Summary: parses a JSON byte slice into a `proto.Message`.
//
// Parameters:
//   - b: []byte. The b.
//   - v: proto.Message. The v.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
//
// Summary: Abstraction for configuration sources.
type Store interface {
	// Load retrieves and returns the McpAnyServerConfig.
	//
	// Summary: Loads the complete server configuration.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//
	// Returns:
	//   - *configv1.McpAnyServerConfig: The loaded configuration.
	//   - error: An error if loading fails.
	Load(ctx context.Context) (*configv1.McpAnyServerConfig, error)

	// HasConfigSources returns true if the store has configuration sources (e.g., file paths) configured.
	//
	// Summary: Checks if the store has any configured sources.
	//
	// Returns:
	//   - bool: True if sources are configured, false otherwise.
	HasConfigSources() bool
}

// ServiceStore extends Store to provide CRUD operations for UpstreamServices.
//
// Summary: Interface for stores that support managing individual services.
type ServiceStore interface {
	Store
	// SaveService saves or updates a service configuration.
	//
	// Summary: Persists a service configuration.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - service: *configv1.UpstreamServiceConfig. The service configuration to save.
	//
	// Returns:
	//   - error: An error if the operation fails.
	SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error

	// GetService retrieves a service configuration by name.
	//
	// Summary: Retrieves a service configuration.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - name: string. The name of the service to retrieve.
	//
	// Returns:
	//   - *configv1.UpstreamServiceConfig: The service configuration.
	//   - error: An error if the service is not found or the operation fails.
	GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error)

	// ListServices retrieves all stored service configurations.
	//
	// Summary: Lists all services.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//
	// Returns:
	//   - []*configv1.UpstreamServiceConfig: A slice of service configurations.
	//   - error: An error if the operation fails.
	ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error)

	// DeleteService removes a service configuration by name.
	//
	// Summary: Deletes a service configuration.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - name: string. The name of the service to delete.
	//
	// Returns:
	//   - error: An error if the operation fails.
	DeleteService(ctx context.Context, name string) error
}

var unknownFieldRegex = regexp.MustCompile(`unknown field "([^"]+)"`)

const maxExpandRecursionDepth = 100

// expand replaces ${VAR}, $VAR, or ${VAR:default} with environment variables.
//
// Summary: Expands environment variables in a byte slice.
//
// Parameters:
//   - b: []byte. The bytes containing variable references.
//
// Returns:
//   - []byte: The expanded bytes.
//   - error: An error if expansion fails or recursion limit is exceeded.
func expand(b []byte) ([]byte, error) {
	return expandRecursive(b, 0)
}

func expandRecursive(b []byte, depth int) ([]byte, error) {
	if depth > maxExpandRecursionDepth {
		return nil, fmt.Errorf("environment variable expansion recursion depth exceeded (max %d)", maxExpandRecursionDepth)
	}

	var missingErrBuilder strings.Builder
	missingCount := 0

	var buf bytes.Buffer
	// Estimate capacity
	buf.Grow(len(b))

	i := 0
	for i < len(b) {
		// Look for '$'
		if b[i] != '$' {
			buf.WriteByte(b[i])
			i++
			continue
		}

		// Found '$', check next char
		if i+1 >= len(b) {
			// Trailing '$', just write it
			buf.WriteByte(b[i])
			i++
			continue
		}

		// Case 1: ${...}
		if b[i+1] == '{' {
			consumed := handleBracedVar(b, i, &buf, &missingErrBuilder, &missingCount, depth)
			if consumed > 0 {
				i += consumed
				continue
			}
			// If not consumed (e.g. unclosed brace), treat as literal
			buf.WriteByte(b[i])
			i++
			continue
		}

		// Case 2: $VAR (alphanumeric + _)
		consumed := handleSimpleVar(b, i, &buf, &missingErrBuilder, &missingCount)
		if consumed > 0 {
			i += consumed
			continue
		}

		// Not a variable
		buf.WriteByte(b[i])
		i++
	}

	if missingCount > 0 {
		// revive:disable-next-line:error-strings // This error message is user facing and needs to be descriptive
		//nolint:staticcheck // This error message is user facing and needs to be descriptive
		return buf.Bytes(), fmt.Errorf("missing environment variables:%s\n    -> Fix: Set these environment variables in your shell or .env file, or provide a default value (e.g., ${VAR:default}).", missingErrBuilder.String())
	}

	return buf.Bytes(), nil
}

func handleBracedVar(b []byte, startIdx int, buf *bytes.Buffer, missingErrBuilder *strings.Builder, missingCount *int, recursionDepth int) int {
	// Find matching '}' accounting for nesting
	innerStart := startIdx + 2
	depth := 1
	j := innerStart
	for j < len(b) {
		if b[j] == '{' {
			depth++
		} else if b[j] == '}' {
			depth--
			if depth == 0 {
				break
			}
		}
		j++
	}

	if depth > 0 {
		// Unclosed brace, treat as literal
		return 0
	}

	// Content inside ${...}
	content := string(b[innerStart:j])
	parts := strings.SplitN(content, ":", 2)
	varName := parts[0]
	var hasDefault bool
	var defaultValue string

	if len(parts) > 1 {
		hasDefault = true
		defaultValue = parts[1]
	}

	if !util.IsEnvVarAllowed(varName) {
		*missingCount++
		lineNum := bytes.Count(b[:startIdx], []byte("\n")) + 1
		fmt.Fprintf(missingErrBuilder, "\n  - Line %d: variable %s is restricted", lineNum, varName)
		// Write the original string to preserve structure
		buf.Write(b[startIdx : j+1])
		return j + 1 - startIdx
	}

	val, ok := os.LookupEnv(varName)
	if !ok && !hasDefault {
		*missingCount++
		lineNum := bytes.Count(b[:startIdx], []byte("\n")) + 1
		fmt.Fprintf(missingErrBuilder, "\n  - Line %d: variable %s is missing", lineNum, varName)
		// Write the original string to preserve structure
		buf.Write(b[startIdx : j+1])
		return j + 1 - startIdx
	}

	useDefault := (ok && val == "" && hasDefault) || (!ok && hasDefault)

	if useDefault {
		expanded, err := expandRecursive([]byte(defaultValue), recursionDepth+1)
		if err != nil {
			*missingCount++
			// Clean up error message from recursive call
			errMsg := err.Error()
			prefix := "missing environment variables:"
			errMsg = strings.TrimPrefix(errMsg, prefix)
			fmt.Fprintf(missingErrBuilder, "\n  - In default value for %s:%s", varName, errMsg)
		}
		buf.Write(expanded)
	} else {
		buf.WriteString(val)
	}

	return j + 1 - startIdx
}

func handleSimpleVar(b []byte, startIdx int, buf *bytes.Buffer, missingErrBuilder *strings.Builder, missingCount *int) int {
	// Scan for variable name
	// First char must be [a-zA-Z_]
	if startIdx+1 >= len(b) {
		return 0
	}
	first := b[startIdx+1]
	isFirstValid := (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_'

	if !isFirstValid {
		return 0
	}

	j := startIdx + 1
	for j < len(b) {
		c := b[j]
		isAlphaNum := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
		if !isAlphaNum {
			break
		}
		j++
	}

	varName := string(b[startIdx+1 : j])
	if !util.IsEnvVarAllowed(varName) {
		*missingCount++
		lineNum := bytes.Count(b[:startIdx], []byte("\n")) + 1
		fmt.Fprintf(missingErrBuilder, "\n  - Line %d: variable %s is restricted", lineNum, varName)
		// Write the original string to preserve structure
		buf.Write(b[startIdx:j])
		return j - startIdx
	}

	val, ok := os.LookupEnv(varName)
	if !ok {
		*missingCount++
		lineNum := bytes.Count(b[:startIdx], []byte("\n")) + 1
		fmt.Fprintf(missingErrBuilder, "\n  - Line %d: variable %s is missing", lineNum, varName)
		// Write the original string to preserve structure
		buf.Write(b[startIdx:j])
		return j - startIdx
	}

	buf.WriteString(val)
	return j - startIdx
}

// FileStore implements the `Store` interface for loading configurations from files.
//
// Summary: Loads configurations from the filesystem.
type FileStore struct {
	fs               afero.Fs
	paths            []string
	skipErrors       bool
	IgnoreMissingEnv bool
	skipValidation   bool
}

// SetSkipValidation configures whether to skip schema validation during loading.
//
// Summary: Sets the skip validation flag.
//
// Parameters:
//   - skip: bool. True to skip validation.
func (s *FileStore) SetSkipValidation(skip bool) {
	s.skipValidation = skip
}

// SetIgnoreMissingEnv configures whether to ignore missing environment variables during loading.
//
// Summary: Sets the ignore missing environment variables flag.
//
// Parameters:
//   - ignore: bool. True to ignore missing environment variables.
func (s *FileStore) SetIgnoreMissingEnv(ignore bool) {
	s.IgnoreMissingEnv = ignore
}

// NewFileStore creates a new FileStore with the given filesystem and paths.
//
// Summary: Initializes a new FileStore.
//
// Parameters:
//   - fs: afero.Fs. The filesystem to use.
//   - paths: []string. The list of paths to scan.
//
// Returns:
//   - *FileStore: A new instance of FileStore.
func NewFileStore(fs afero.Fs, paths []string) *FileStore {
	return &FileStore{fs: fs, paths: paths}
}

// NewFileStoreWithSkipErrors creates a new FileStore that skips malformed config files.
//
// Summary: Initializes a new FileStore that tolerates errors in config files.
//
// Parameters:
//   - fs: afero.Fs. The filesystem to use.
//   - paths: []string. The list of paths to scan.
//
// Returns:
//   - *FileStore: A new instance of FileStore.
func NewFileStoreWithSkipErrors(fs afero.Fs, paths []string) *FileStore {
	return &FileStore{fs: fs, paths: paths, skipErrors: true}
}

// HasConfigSources returns true if the store has configuration paths configured.
//
// Summary: Checks if the store has any configured paths.
//
// Returns:
//   - bool: True if paths are configured, false otherwise.
func (s *FileStore) HasConfigSources() bool {
	return len(s.paths) > 0
}

// Load scans the configured paths and merges them into a single configuration.
//
// Summary: Loads and merges configurations from all configured paths.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//
// Returns:
//   - *configv1.McpAnyServerConfig: The merged configuration.
//   - error: An error if loading or merging fails.
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
			if !s.IgnoreMissingEnv {
				return nil, WrapActionableError(fmt.Sprintf("failed to expand environment variables in %s", path), err)
			}
			logging.GetLogger().Warn("Missing environment variables in config, proceeding with unexpanded values", "path", path, "error", err)
		}

		engine, err := NewEngine(path)
		if err != nil {
			if s.skipErrors {
				logging.GetLogger().Error("Failed to determine config engine, skipping file", "path", path, "error", err)
				continue
			}
			return nil, err
		}

		if configurable, ok := engine.(ConfigurableEngine); ok {
			configurable.SetSkipValidation(s.skipValidation)
		}

		cfg := configv1.McpAnyServerConfig_builder{}.Build()
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
			// Apply Merge Strategy if defined
			if ms := cfg.GetMergeStrategy(); ms != nil {
				if ms.GetUpstreamServiceList() == "replace" {
					mergedConfig.SetUpstreamServices(nil)
				}

				if ms.GetProfileList() == "replace" {
					if gs := mergedConfig.GetGlobalSettings(); gs != nil {
						gs.SetProfiles(nil)
						gs.SetProfileDefinitions(nil)
					}
				}

				// Handle other lists if needed (e.g., users, collections) based on requirements
				// For now, adhering to the documented strategies.
			}
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

		applyPathToMap(m, path, value, v)
	}
}

// applySetOverrides applies configuration overrides from the --set flag.
// It supports dot notation and bracket notation for indices.
// Example: upstream_services[0].http_service.address=http://localhost:8080
func applySetOverrides(m map[string]interface{}, setValues []string, v proto.Message) {
	for _, sv := range setValues {
		parts := strings.SplitN(sv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		// Normalize key: replace [idx] with .idx.
		key = strings.ReplaceAll(key, "[", ".")
		key = strings.ReplaceAll(key, "]", ".")
		key = strings.ReplaceAll(key, "..", ".")
		key = strings.Trim(key, ".")
		path := strings.Split(key, ".")

		applyPathToMap(m, path, value, v)
	}
}

// applyPathToMap walks the map and sets the value at the given path.
func applyPathToMap(m map[string]interface{}, path []string, value string, v proto.Message) {
	if len(path) > 0 && path[0] == "upstream" {
		path[0] = "upstream_services"
	}
	current := m
	for i, originalSection := range path {
		section := strings.ToLower(originalSection) // Normalize to snake_case/lowercase

		// Clear oneof siblings if this field is part of a oneof.
		// This prevents "oneof is already set" errors when overriding.
		if v != nil {
			md := getDescriptorAtSubpath(v.ProtoReflect().Descriptor(), path[:i])
			if md != nil {
				if fd := findField(md, section); fd != nil {
					clearOneofSiblings(current, fd)
				}
			}
		}

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

func getDescriptorAtSubpath(md protoreflect.MessageDescriptor, path []string) protoreflect.MessageDescriptor {
	current := md
	for _, part := range path {
		if _, err := strconv.Atoi(part); err == nil {
			// If part is a number, we assume it's a list index.
			// The current 'md' should already be the message type of the list elements.
			continue
		}
		fd := findField(current, part)
		if fd == nil {
			return nil
		}
		if fd.IsList() || fd.Kind() == protoreflect.MessageKind {
			if fd.Kind() == protoreflect.MessageKind {
				current = fd.Message()
			}
			continue
		}
		return nil
	}
	return current
}

func clearOneofSiblings(m map[string]interface{}, fd protoreflect.FieldDescriptor) {
	oo := fd.ContainingOneof()
	if oo == nil {
		return
	}
	for i := 0; i < oo.Fields().Len(); i++ {
		sibling := oo.Fields().Get(i)
		if sibling.FullName() != fd.FullName() {
			fmt.Fprintf(os.Stderr, "DEBUG: Clearing oneof sibling %s (Name: %s, JSONName: %s) because %s is being set\n", sibling.FullName(), sibling.Name(), sibling.JSONName(), fd.Name())
			delete(m, string(sibling.Name()))
			delete(m, sibling.JSONName())
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
				trimmed := strings.TrimSpace(value)
				// Special handling for JSON arrays/objects in environment variables.
				// This allows setting repeated fields (especially messages) using valid JSON syntax.
				if strings.HasPrefix(trimmed, "[") {
					var jsonList []interface{}
					if json.Unmarshal([]byte(value), &jsonList) == nil {
						return jsonList
					}
				}
				if strings.HasPrefix(trimmed, "{") {
					var jsonObj map[string]interface{}
					if json.Unmarshal([]byte(value), &jsonObj) == nil {
						return []interface{}{jsonObj}
					}
				}

				// Repeated field: split by comma, respecting JSON structure and quotes
				parts := splitByCommaIgnoringBraces(value)
				var list []interface{}
				for _, part := range parts {
					// If it looks like a quoted CSV value, unquote it.
					part = unquoteCSV(part)

					switch kind {
					case protoreflect.BoolKind:
						b, err := strconv.ParseBool(part)
						if err == nil {
							list = append(list, b)
						} else {
							list = append(list, part)
						}
					case protoreflect.MessageKind:
						// For repeated messages, try to unmarshal each part as JSON
						var msgMap map[string]interface{}
						if json.Unmarshal([]byte(part), &msgMap) == nil {
							list = append(list, msgMap)
						} else {
							list = append(list, part)
						}
					default:
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
//
// Summary: Combines multiple stores into a single logical store.
type MultiStore struct {
	stores []Store
}

// NewMultiStore creates a new MultiStore with the given stores.
//
// Summary: Initializes a new MultiStore.
//
// Parameters:
//   - stores: ...Store. The stores to aggregate.
//
// Returns:
//   - *MultiStore: A new instance of MultiStore.
func NewMultiStore(stores ...Store) *MultiStore {
	return &MultiStore{stores: stores}
}

// Load loads configurations from all stores and merges them into a single config.
//
// Summary: Loads and merges configurations from all underlying stores.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//
// Returns:
//   - *configv1.McpAnyServerConfig: The merged configuration.
//   - error: An error if loading fails.
func (ms *MultiStore) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	mergedConfig := configv1.McpAnyServerConfig_builder{}.Build()
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
	// Check common aliases first for immediate feedback
	aliases := map[string]string{
		"url":       "address",
		"uri":       "address",
		"endpoint":  "address",
		"endpoints": "address",
		"host":      "address",
		"cmd":       "command",
		"args":      "arguments",
	}
	if correction, ok := aliases[strings.ToLower(unknownField)]; ok {
		return fmt.Sprintf("Did you mean %q? (Common alias)", correction)
	}

	candidates := make(map[string]struct{})
	collectFieldNames(root.ProtoReflect().Descriptor(), candidates)

	// Explicitly add fields from common nested configuration objects to the candidates.
	// We avoid full recursion to prevent suggesting fields from obscure/irrelevant parts of the schema
	// (like "services" from Collection which confuses users when they mean "upstream_services").
	commonMessages := []proto.Message{
		configv1.GlobalSettings_builder{}.Build(),
		configv1.UpstreamServiceConfig_builder{}.Build(),
		configv1.HttpUpstreamService_builder{}.Build(),
		configv1.GrpcUpstreamService_builder{}.Build(),
		configv1.McpUpstreamService_builder{}.Build(),
		configv1.OpenapiUpstreamService_builder{}.Build(),
		configv1.CommandLineUpstreamService_builder{}.Build(),
		configv1.SqlUpstreamService_builder{}.Build(),
		configv1.Authentication_builder{}.Build(),
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
	limit := len(unknownField) / 2

	// For short strings, be strict to avoid garbage suggestions (e.g. xyz -> env)
	if len(unknownField) <= 3 {
		limit = 1
	} else if limit < 3 {
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
// Summary: Checks if any underlying store has configuration sources.
//
// Returns:
//   - bool: True if at least one store has sources, false otherwise.
func (ms *MultiStore) HasConfigSources() bool {
	for _, s := range ms.stores {
		if s.HasConfigSources() {
			return true
		}
	}
	return false
}

func findKeyLine(b []byte, key string) int {
	var node yaml.Node
	if err := yaml.Unmarshal(b, &node); err != nil {
		return 0
	}
	return findKeyInNode(&node, key)
}

func findKeyInNode(node *yaml.Node, key string) int {
	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			if line := findKeyInNode(child, key); line > 0 {
				return line
			}
		}
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]

			if keyNode.Value == key {
				return keyNode.Line
			}

			if line := findKeyInNode(valNode, key); line > 0 {
				return line
			}
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			if line := findKeyInNode(child, key); line > 0 {
				return line
			}
		}
	}
	return 0
}

// splitByCommaIgnoringBraces splits a string by comma, but ignores commas inside
// braces {}, brackets [], and double quotes "".
func splitByCommaIgnoringBraces(s string) []string {
	var parts []string
	var current strings.Builder
	depth := 0
	quote := false
	escape := false

	for _, r := range s {
		if escape {
			escape = false
			current.WriteRune(r)
			continue
		}

		if r == '\\' {
			escape = true
			current.WriteRune(r)
			continue
		}

		if r == '"' {
			if quote {
				quote = false
			} else {
				// Only start quote if inside braces OR at start of field
				// (ignoring leading whitespace handled by TrimSpace logic on output, but here we need to check current buffer)
				isStartOfField := strings.TrimSpace(current.String()) == ""
				if depth > 0 || isStartOfField {
					quote = true
				}
			}
		}

		if !quote {
			switch r {
			case '{', '[':
				depth++
			case '}', ']':
				depth--
			}
		}

		if r == ',' && depth == 0 && !quote {
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}
	return parts
}

// unquoteCSV removes surrounding double quotes and unescapes paired double quotes.
// Example: "foo" -> foo
// Example: "foo""bar" -> foo"bar.
func unquoteCSV(s string) string {
	if len(s) >= 2 && strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		inner := s[1 : len(s)-1]
		return strings.ReplaceAll(inner, "\"\"", "\"")
	}
	return s
}
