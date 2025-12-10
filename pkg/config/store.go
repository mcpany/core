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

	"github.com/spf13/afero"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"

	configv1 "github.com/mcpany/core/proto/config/v1"
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

	// Helper to fix log level
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
	return protojson.Unmarshal(jsonData, v)
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
				if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsPrivate() {
					return nil, fmt.Errorf("ssrf attempt blocked: %s", addr)
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
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
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

	defer resp.Body.Close()

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
