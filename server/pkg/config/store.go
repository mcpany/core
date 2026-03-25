// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package config provides the configuration engine and store for MCP Any.
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
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"
)

// Engine defines the interface for configuration unmarshaling from different file formats.
type Engine interface {
	// Unmarshal parses the given byte slice and populates the provided proto.Message.
	Unmarshal(b []byte, v proto.Message) error
}

// StructuredEngine defines an interface for engines that can unmarshal directly from a map structure.
type StructuredEngine interface {
	Engine
	// UnmarshalFromMap populates the provided proto.Message from a raw map.
	UnmarshalFromMap(m map[string]interface{}, v proto.Message, originalBytes []byte) error
}

// ConfigurableEngine defines an interface for engines that support configuration options.
type ConfigurableEngine interface {
	Engine
	// SetSkipValidation sets whether to skip schema validation.
	SetSkipValidation(skip bool)

	// SetIgnoreEnv sets whether to ignore environment variables and other external overrides.
	SetIgnoreEnv(ignore bool)
}

// NewEngine returns a configuration engine capable of unmarshaling the format indicated by the file extension.
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
	ignoreEnv      bool
}

func (e *yamlEngine) SetSkipValidation(skip bool) {
	e.skipValidation = skip
}

func (e *yamlEngine) SetIgnoreEnv(ignore bool) {
	e.ignoreEnv = ignore
}

func (e *yamlEngine) Unmarshal(b []byte, v proto.Message) error {
	var yamlMap map[string]interface{}
	if err := yaml.Unmarshal(b, &yamlMap); err != nil {
		if strings.Contains(err.Error(), "found character that cannot start any token") {
			if bytes.Contains(b, []byte("\t")) {
				// revive:disable-next-line:error-strings
				return fmt.Errorf("failed to unmarshal yaml: %w\n\nhint: yaml files cannot contain tabs, please use spaces for indentation", err)
			}
		}
		return fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return e.unmarshalInternal(yamlMap, v, b)
}

func (e *yamlEngine) UnmarshalFromMap(yamlMap map[string]interface{}, v proto.Message, originalBytes []byte) error {
	return e.unmarshalInternal(yamlMap, v, originalBytes)
}

func (e *yamlEngine) unmarshalInternal(yamlMap map[string]interface{}, v proto.Message, originalBytes []byte) error {
	if !e.ignoreEnv {
		applyEnvVarsFromSlice(yamlMap, os.Environ(), v)
		applySetOverrides(yamlMap, GlobalSettings().SetValues(), v)
	}

	if v != nil {
		fixTypes(yamlMap, v.ProtoReflect().Descriptor())
	}

	if gs, ok := yamlMap["global_settings"].(map[string]interface{}); ok {
		if ll, ok := gs["log_level"].(string); ok {
			if !strings.HasPrefix(ll, "LOG_LEVEL_") {
				gs["log_level"] = "LOG_LEVEL_" + strings.ToUpper(ll)
			}
		}
	}

	jsonData, err := json.Marshal(yamlMap)
	if err != nil {
		return fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	if err := protojson.Unmarshal(jsonData, v); err != nil {
		if originalBytes != nil {
			if matches := unknownFieldRegex.FindStringSubmatch(err.Error()); len(matches) > 1 {
				unknownField := matches[1]
				if line := findKeyLine(originalBytes, unknownField); line > 0 {
					err = fmt.Errorf("line %d: %w", line, err)
				}
			}
		}

		if strings.Contains(err.Error(), "unknown field \"mcpServers\"") {
			// revive:disable-next-line:error-strings
			return fmt.Errorf("%w\n\ndid you mean \"upstream_services\"? it looks like you might be using a Claude Desktop configuration format, MCP Any uses a different configuration structure, see documentation for details", err)
		}

		if strings.Contains(err.Error(), "unknown field \"services\"") {
			// revive:disable-next-line:error-strings
			return fmt.Errorf("%w\n\ndid you mean \"upstream_services\"? \"services\" is not a valid top-level key", err)
		}

		if strings.Contains(err.Error(), "unknown field \"service_config\"") {
			// revive:disable-next-line:error-strings
			return fmt.Errorf("%w\n\nit looks like you are using 'service_config' as a wrapper key, in MCP Any configuration, you should place the service type directly under the service definition, without a 'service_config' wrapper", err)
		}

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

func (e *textprotoEngine) Unmarshal(b []byte, v proto.Message) error {
	return prototext.Unmarshal(b, v)
}

// jsonEngine implements the Engine interface for JSON configuration files.
type jsonEngine struct{}

func (e *jsonEngine) Unmarshal(b []byte, v proto.Message) error {
	if err := protojson.Unmarshal(b, v); err != nil {
		if strings.Contains(err.Error(), "unknown field \"mcpServers\"") {
			// revive:disable-next-line:error-strings
			return fmt.Errorf("%w\n\ndid you mean \"upstream_services\"? it looks like you might be using a Claude Desktop configuration format, MCP Any uses a different configuration structure, see documentation for details", err)
		}

		if strings.Contains(err.Error(), "unknown field \"services\"") {
			// revive:disable-next-line:error-strings
			return fmt.Errorf("%w\n\ndid you mean \"upstream_services\"? \"services\" is not a valid top-level key", err)
		}

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
type Store interface {
	Load(ctx context.Context) (*configv1.McpAnyServerConfig, error)
	HasConfigSources() bool
}

// ServiceStore extends Store to provide CRUD operations for UpstreamServices.
type ServiceStore interface {
	Store
	SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error
	GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error)
	ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error)
	DeleteService(ctx context.Context, name string) error
}

var unknownFieldRegex = regexp.MustCompile(`unknown field "([^"]+)"`)

const maxExpandRecursionDepth = 100

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
	buf.Grow(len(b))

	i := 0
	for i < len(b) {
		if b[i] != '$' {
			buf.WriteByte(b[i])
			i++
			continue
		}

		if i+1 >= len(b) {
			buf.WriteByte(b[i])
			i++
			continue
		}

		if b[i+1] == '{' {
			consumed := handleBracedVar(b, i, &buf, &missingErrBuilder, &missingCount, depth)
			if consumed > 0 {
				i += consumed
				continue
			}
			buf.WriteByte(b[i])
			i++
			continue
		}

		consumed := handleSimpleVar(b, i, &buf, &missingErrBuilder, &missingCount)
		if consumed > 0 {
			i += consumed
			continue
		}

		buf.WriteByte(b[i])
		i++
	}

	if missingCount > 0 {
		// revive:disable-next-line:error-strings
		return buf.Bytes(), fmt.Errorf("missing environment variables:%s\n    -> fix: set these environment variables in your shell or .env file, or provide a default value (e.g., ${VAR:default})", missingErrBuilder.String())
	}

	return buf.Bytes(), nil
}

func handleBracedVar(b []byte, startIdx int, buf *bytes.Buffer, missingErrBuilder *strings.Builder, missingCount *int, recursionDepth int) int {
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
		return 0
	}

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
		buf.Write(b[startIdx : j+1])
		return j + 1 - startIdx
	}

	val, ok := os.LookupEnv(varName)
	if !ok && !hasDefault {
		*missingCount++
		lineNum := bytes.Count(b[:startIdx], []byte("\n")) + 1
		fmt.Fprintf(missingErrBuilder, "\n  - Line %d: variable %s is missing", lineNum, varName)
		buf.Write(b[startIdx : j+1])
		return j + 1 - startIdx
	}

	useDefault := (ok && val == "" && hasDefault) || (!ok && hasDefault)

	if useDefault {
		expanded, err := expandRecursive([]byte(defaultValue), recursionDepth+1)
		if err != nil {
			*missingCount++
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
		buf.Write(b[startIdx:j])
		return j - startIdx
	}

	val, ok := os.LookupEnv(varName)
	if !ok {
		*missingCount++
		lineNum := bytes.Count(b[:startIdx], []byte("\n")) + 1
		fmt.Fprintf(missingErrBuilder, "\n  - Line %d: variable %s is missing", lineNum, varName)
		buf.Write(b[startIdx:j])
		return j - startIdx
	}

	buf.WriteString(val)
	return j - startIdx
}

// FileStore implements the `Store` interface for loading configurations from files.
type FileStore struct {
	fs               afero.Fs
	paths            []string
	skipErrors       bool
	IgnoreMissingEnv bool
	skipValidation   bool
}

// SetSkipValidation configures whether to skip schema validation during loading.
func (s *FileStore) SetSkipValidation(skip bool) {
	s.skipValidation = skip
}

// SetIgnoreMissingEnv configures whether to ignore missing environment variables during loading.
func (s *FileStore) SetIgnoreMissingEnv(ignore bool) {
	s.IgnoreMissingEnv = ignore
}

// NewFileStore creates a new FileStore with the given filesystem and paths.
func NewFileStore(fs afero.Fs, paths []string) *FileStore {
	return &FileStore{fs: fs, paths: paths}
}

// NewFileStoreWithSkipErrors creates a new FileStore that skips malformed config files.
func NewFileStoreWithSkipErrors(fs afero.Fs, paths []string) *FileStore {
	return &FileStore{fs: fs, paths: paths, skipErrors: true}
}

// HasConfigSources returns true if the store has configuration paths configured.
func (s *FileStore) HasConfigSources() bool {
	return len(s.paths) > 0
}

// Load scans the configured paths and merges them into a single configuration.
func (s *FileStore) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	filePaths, err := s.collectFilePaths()
	if err != nil {
		return nil, fmt.Errorf("failed to collect config file paths: %w", err)
	}

	configs := make([]*configv1.McpAnyServerConfig, len(filePaths))
	g, ctx := errgroup.WithContext(ctx)

	for i, path := range filePaths {
		i, path := i, path
		g.Go(func() error {
			cfg, err := s.loadOneConfig(ctx, path)
			if err != nil {
				return err
			}
			configs[i] = cfg
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	var mergedConfig *configv1.McpAnyServerConfig
	for _, cfg := range configs {
		if cfg == nil {
			continue
		}
		if mergedConfig == nil {
			mergedConfig = cfg
		} else {
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
			}
			proto.Merge(mergedConfig, cfg)
		}
	}

	return mergedConfig, nil
}

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

	if resp.StatusCode >= 300 && resp.StatusCode <= 399 {
		return nil, fmt.Errorf("redirects are disabled for security reasons")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get config from url %s: status code %d", url, resp.StatusCode)
	}

	resp.Body = http.MaxBytesReader(nil, resp.Body, 1024*1024)
	return io.ReadAll(resp.Body)
}

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

func applyEnvVarsFromSlice(m map[string]interface{}, environ []string, v proto.Message) {
	sortedEnv := make([]string, len(environ))
	copy(sortedEnv, environ)
	sort.Strings(sortedEnv)

	for _, env := range sortedEnv {
		if !strings.HasPrefix(env, "MCPANY__") {
			continue
		}
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		trimmedKey := strings.TrimPrefix(key, "MCPANY__")
		path := strings.Split(trimmedKey, "__")

		applyPathToMap(m, path, value, v)
	}
}

func applySetOverrides(m map[string]interface{}, setValues []string, v proto.Message) {
	for _, sv := range setValues {
		parts := strings.SplitN(sv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		key = strings.ReplaceAll(key, "[", ".")
		key = strings.ReplaceAll(key, "]", ".")
		key = strings.ReplaceAll(key, "..", ".")
		key = strings.Trim(key, ".")
		path := strings.Split(key, ".")

		applyPathToMap(m, path, value, v)
	}
}

func applyPathToMap(m map[string]interface{}, path []string, value string, v proto.Message) {
	if len(path) > 0 && path[0] == "upstream" {
		path[0] = "upstream_services"
	}
	current := m
	for i, originalSection := range path {
		section := strings.ToLower(originalSection)

		if v != nil {
			md := getDescriptorAtSubpath(v.ProtoReflect().Descriptor(), path[:i])
			if md != nil {
				if fd := findField(md, section); fd != nil {
					clearOneofSiblings(current, fd)
				}
			}
		}

		if i == len(path)-1 {
			resolvedValue := resolveEnvValue(v, path, value)
			current[section] = resolvedValue
		} else {
			if next, ok := current[section].(map[string]interface{}); ok {
				current = next
			} else if slice, ok := current[section].([]interface{}); ok {
				next := make(map[string]interface{})
				for idx, val := range slice {
					next[strconv.Itoa(idx)] = val
				}
				current[section] = next
				current = next
			} else {
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
			delete(m, string(sibling.Name()))
			delete(m, sibling.JSONName())
		}
	}
}

func resolveEnvValue(root proto.Message, path []string, value string) interface{} {
	if root == nil {
		return value
	}
	md := root.ProtoReflect().Descriptor()
	var currentFd protoreflect.FieldDescriptor

	for i := 0; i < len(path); i++ {
		part := strings.ToLower(path[i])

		if currentFd != nil && currentFd.IsList() {
			if currentFd.Kind() == protoreflect.MessageKind {
				md = currentFd.Message()
				currentFd = nil
				continue
			}
			if i == len(path)-1 {
				return convertKind(currentFd.Kind(), value)
			}
			return value
		}

		fd := findField(md, part)
		if fd == nil {
			return value
		}
		currentFd = fd

		if i == len(path)-1 {
			kind := fd.Kind()

			if fd.IsList() {
				trimmed := strings.TrimSpace(value)
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

				parts := splitByCommaIgnoringBraces(value)
				var list []interface{}
				for _, part := range parts {
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

		switch {
		case fd.Kind() == protoreflect.MessageKind:
			if fd.IsList() {
				continue
			}
			md = fd.Message()
			currentFd = nil
		case fd.IsList():
			continue
		default:
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
	return value
}

func fixTypes(m map[string]interface{}, md protoreflect.MessageDescriptor) {
	for key, val := range m {
		fd := findField(md, key)
		if fd == nil {
			continue
		}

		if fd.IsList() {
			if valMap, ok := val.(map[string]interface{}); ok {
				newSlice := convertMapToSlice(valMap)
				m[key] = newSlice
				val = newSlice
			}

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
	fd := md.Fields().ByName(protoreflect.Name(name))
	if fd != nil {
		return fd
	}
	fd = md.Fields().ByJSONName(name)
	if fd != nil {
		return fd
	}
	return nil
}

// MultiStore implements the Store interface for loading configurations from multiple stores.
type MultiStore struct {
	stores []Store
}

// NewMultiStore creates a new MultiStore with the given stores.
func NewMultiStore(stores ...Store) *MultiStore {
	return &MultiStore{stores: stores}
}

// Load loads configurations from all stores and merges them into a single config.
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

func suggestFix(unknownField string, root proto.Message) string {
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
		return fmt.Sprintf("did you mean %q? (common alias)", correction)
	}

	candidates := make(map[string]struct{})
	collectFieldNames(root.ProtoReflect().Descriptor(), candidates)

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
		dist := util.LevenshteinDistance(unknownField, name)
		if dist < minDist {
			minDist = dist
			bestMatch = name
		}
	}

	limit := len(unknownField) / 2

	if len(unknownField) <= 3 {
		limit = 1
	} else if limit < 3 {
		limit = 3
	}

	if minDist <= limit {
		return fmt.Sprintf("did you mean %q?", bestMatch)
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

func unquoteCSV(s string) string {
	if len(s) >= 2 && strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		inner := s[1 : len(s)-1]
		return strings.ReplaceAll(inner, "\"\"", "\"")
	}
	return s
}

func (s *FileStore) loadOneConfig(ctx context.Context, path string) (*configv1.McpAnyServerConfig, error) {
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
		return nil, nil
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
			return nil, nil
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
			var raw map[string]interface{}
			if yaml.Unmarshal(b, &raw) == nil {
				if services, ok := raw["upstream_services"].([]interface{}); ok {
					for _, s := range services {
						if service, ok := s.(map[string]interface{}); ok {
							if name, ok := service["name"].(string); ok {
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
			return nil, nil
		}
		return nil, logErr
	}
	return cfg, nil
}
