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

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"

	configv1 "github.com/mcpxy/core/proto/config/v1"
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
// indicated by the file extension of the given path. It supports .json, .yaml,
// .yml, and .textproto.
//
// path is the file path used to determine the configuration format.
func NewEngine(path string) (Engine, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return &jsonEngine{}, nil
	case ".yaml", ".yml":
		return &yamlEngine{}, nil
	case ".textproto":
		return &textprotoEngine{}, nil
	default:
		return nil, fmt.Errorf("unsupported config file extension '%s' for file %s", ext, path)
	}
}

// yamlEngine implements the Engine interface for YAML configuration files.
type yamlEngine struct{}

// Unmarshal parses a YAML byte slice into a proto.Message. It achieves this by
// first unmarshaling the YAML into a generic map, then marshaling that map to
// JSON, and finally unmarshaling the JSON into the target protobuf message.
func (e *yamlEngine) Unmarshal(b []byte, v proto.Message) error {
	// First, unmarshal YAML into a generic map.
	var yamlMap map[string]interface{}
	if err := yaml.Unmarshal(b, &yamlMap); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Then, marshal the map to JSON. This is a common way to convert YAML to JSON.
	jsonData, err := json.Marshal(yamlMap)
	if err != nil {
		return fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	// Finally, unmarshal the JSON into the protobuf message.
	return protojson.Unmarshal(jsonData, v)
}

// textprotoEngine implements the Engine interface for textproto configuration files.
type textprotoEngine struct{}

// Unmarshal parses a textproto byte slice into a proto.Message.
func (e *textprotoEngine) Unmarshal(b []byte, v proto.Message) error {
	return prototext.Unmarshal(b, v)
}

// jsonEngine implements the Engine interface for JSON configuration files.
type jsonEngine struct{}

// Unmarshal parses a JSON byte slice into a proto.Message.
func (e *jsonEngine) Unmarshal(b []byte, v proto.Message) error {
	return protojson.Unmarshal(b, v)
}

// Store defines the interface for loading MCP-X server configurations.
// Implementations of this interface provide a way to retrieve the complete
// server configuration from a source, such as a file or a remote service.
type Store interface {
	// Load retrieves and returns the McpxServerConfig.
	Load() (*configv1.McpxServerConfig, error)
}

// FileStore implements the Store interface for loading configurations from one or
// more files or directories on a filesystem.
type FileStore struct {
	fs    afero.Fs
	paths []string
}

// NewFileStore creates a new FileStore with the given filesystem and a list of
// paths to load configurations from.
//
// fs is the filesystem interface to use for file operations.
// paths is a slice of file or directory paths to scan for configuration files.
func NewFileStore(fs afero.Fs, paths []string) *FileStore {
	return &FileStore{fs: fs, paths: paths}
}

// Load scans the configured paths for supported configuration files, reads them,
// unmarshals their contents, and merges them into a single McpxServerConfig.
// The files are processed in alphabetical order, and later configurations are
// merged into earlier ones.
func (s *FileStore) Load() (*configv1.McpxServerConfig, error) {
	var mergedConfig *configv1.McpxServerConfig

	filePaths, err := s.collectFilePaths()
	if err != nil {
		return nil, fmt.Errorf("failed to collect config file paths: %w", err)
	}

	for _, path := range filePaths {
		b, err := afero.ReadFile(s.fs, path)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
		}

		if len(b) == 0 {
			continue
		}

		engine, err := NewEngine(path)
		if err != nil {
			return nil, err
		}

		cfg := &configv1.McpxServerConfig{}
		if err := engine.Unmarshal(b, cfg); err != nil {
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

// collectFilePaths recursively scans the configured paths and returns a list of valid config files.
func (s *FileStore) collectFilePaths() ([]string, error) {
	var files []string
	for _, path := range s.paths {
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
