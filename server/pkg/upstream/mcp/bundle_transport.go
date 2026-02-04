// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	idValueIntRegex = regexp.MustCompile(`value:(\d+)`)
	idValueStrRegex = regexp.MustCompile(`value:([^}]+)`)
)

// transportError implements the error interface for JSON-RPC errors.
type transportError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error returns the error message.
//
// Returns the result.
func (e *transportError) Error() string {
	return e.Message
}

// BundleDockerTransport implements the mcp.Transport interface to connect to a service
// running inside a Docker container from a bundle. It supports mounts and environment variables.
type BundleDockerTransport struct {
	Image      string
	Command    string
	Args       []string
	Env        []string
	Mounts     []mount.Mount
	WorkingDir string

	// dockerClientFactory allows injecting a custom docker client for testing.
	// If nil, newDockerClient is used.
	dockerClientFactory func(ops ...client.Opt) (dockerClient, error)
}

// Connect establishes a connection to the service within the Docker container.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns an error if the operation fails.
func (t *BundleDockerTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	log := logging.GetLogger()

	factory := t.dockerClientFactory
	if factory == nil {
		factory = newDockerClient
	}

	cli, err := factory(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	if t.Image == "" {
		return nil, fmt.Errorf("image must be specified for bundle docker transport")
	}

	reader, err := cli.ImagePull(ctx, t.Image, image.PullOptions{})
	if err != nil {
		log.Warn("Failed to pull docker image, will try to use local image if available", "image", t.Image, "error", err)
	} else {
		_, _ = io.Copy(io.Discard, reader)
		log.Info("Successfully pulled docker image", "image", t.Image)
	}

	// Construct the shell command (similar to DockerTransport)
	// We use /bin/sh -c to run the command to handle complex args or shell features if needed,
	// but mostly to stay consistent with DockerTransport logic.
	// However, if Command is simple, we might not need sh -c.
	// But let's stick to the pattern to ensure we can chain commands if needed (though here we just run one).
	// Actually, if we just want to run the command, we can set Cmd directly.
	// But existing DockerTransport uses:
	// mainCommandParts := []string{"exec", command} ... script := ... /bin/sh -c script
	// This is robust for stdio piping.

	var cmd []string
	if t.Command != "" {
		cmd = append([]string{t.Command}, t.Args...)
	}

	containerConfig := &container.Config{
		Image:        t.Image,
		Cmd:          cmd,
		Env:          t.Env,
		WorkingDir:   t.WorkingDir,
		Tty:          false, // Must be false for MCP stdio
		OpenStdin:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	}

	hostConfig := &container.HostConfig{
		Mounts: t.Mounts,
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	hijackedResp, err := cli.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		_ = cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true})
		return nil, fmt.Errorf("failed to attach to container: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		_ = cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true})
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	stdoutReader, stdoutWriter := io.Pipe()

	// Capture Stderr for logging
	go func() {
		defer func() { _ = stdoutWriter.Close() }()
		logWriterWithLevel := &bundleSlogWriter{log: log, level: slog.LevelError}
		_, err := stdcopy.StdCopy(stdoutWriter, logWriterWithLevel, hijackedResp.Reader)
		if err != nil && err != io.EOF {
			log.Error("Error demultiplexing docker stream", "error", err)
		}
	}()

	rwc := &dockerReadWriteCloser{
		Reader:      stdoutReader,
		WriteCloser: hijackedResp.Conn,
		containerID: resp.ID,
		cli:         cli,
	}
	return &bundleDockerConn{
		rwc:     rwc,
		decoder: json.NewDecoder(rwc),
		encoder: json.NewEncoder(rwc),
		log:     log,
	}, nil
}

type bundleDockerConn struct {
	rwc     io.ReadWriteCloser
	decoder *json.Decoder
	encoder *json.Encoder
	log     *slog.Logger
}

// Read reads a JSON-RPC message from the connection.
//
// _ is an unused parameter.
//
// Returns the result.
// Returns an error if the operation fails.
func (c *bundleDockerConn) Read(_ context.Context) (jsonrpc.Message, error) {
	// ⚡ BOLT: Optimized JSON parsing to avoid triple-unmarshaling.
	// Randomized Selection from Top 5 High-Impact Targets
	var rawMap map[string]json.RawMessage
	if err := c.decoder.Decode(&rawMap); err != nil {
		c.log.Error("Failed to decode message", "error", err)
		return nil, err
	}

	// Efficiently determine if it's a request by checking for "method"
	var method string
	if m, ok := rawMap["method"]; ok {
		if err := json.Unmarshal(m, &method); err != nil {
			return nil, fmt.Errorf("failed to unmarshal method: %w", err)
		}
	}

	isRequest := method != ""
	if isRequest {
		c.log.Info("Read request", "method", method)
	} else {
		c.log.Info("Read response")
	}

	var msg jsonrpc.Message

	if isRequest {
		req := &jsonrpc.Request{
			Method: method,
		}
		if p, ok := rawMap["params"]; ok {
			req.Params = p
		}
		if idRaw, ok := rawMap["id"]; ok {
			var idVal any
			if err := json.Unmarshal(idRaw, &idVal); err != nil {
				return nil, fmt.Errorf("failed to unmarshal id: %w", err)
			}
			if err := setUnexportedID(&req.ID, idVal); err != nil {
				c.log.Error("Failed to set unexported ID on request", "error", err)
			}
		}
		msg = req
	} else {
		resp := &jsonrpc.Response{}
		if r, ok := rawMap["result"]; ok {
			resp.Result = r
		}
		if e, ok := rawMap["error"]; ok {
			var transErr transportError
			if err := json.Unmarshal(e, &transErr); err != nil {
				return nil, fmt.Errorf("failed to unmarshal error: %w", err)
			}
			resp.Error = &transErr
		}
		if idRaw, ok := rawMap["id"]; ok {
			var idVal any
			if err := json.Unmarshal(idRaw, &idVal); err != nil {
				return nil, fmt.Errorf("failed to unmarshal id: %w", err)
			}
			if err := setUnexportedID(&resp.ID, idVal); err != nil {
				c.log.Error("Failed to set unexported ID on response", "error", err)
			}
		}
		msg = resp
	}

	return msg, nil
}

func setUnexportedID(idPtr interface{}, val interface{}) error {
	if val == nil {
		return nil // ID{value: nil} is default
	}
	// jsonrpc2.ID struct has 'value' field.
	// Check if val is number (float64 from json) -> convert to int if possible?
	// jsonrpc2.ID value field is interface{}.

	// Ensure val is int64 if it looks like int (for consistency with SDK which uses int64)
	// JSON unmarshals integer as float64.
	if f, ok := val.(float64); ok {
		if float64(int64(f)) == f {
			val = int64(f)
		}
	}

	v := reflect.ValueOf(idPtr).Elem()
	f := v.FieldByName("value")
	if !f.IsValid() {
		// This suggests the SDK internal structure has changed.
		return fmt.Errorf("field 'value' not found in jsonrpc.ID struct")
	}

	// Safety check: ensure the field is addressable before unsafe operation
	if !f.CanAddr() {
		return fmt.Errorf("field 'value' is not addressable")
	}

	f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem() //nolint:gosec // Need unsafe to access unexported field
	f.Set(reflect.ValueOf(val))
	return nil
}

// Write writes a JSON-RPC message to the connection.
//
// _ is an unused parameter.
// msg is the msg.
//
// Returns an error if the operation fails.
func (c *bundleDockerConn) Write(_ context.Context, msg jsonrpc.Message) error {
	// Workaround: jsonrpc.ID in the SDK marshals to {} because of unexported fields.
	// We extract the value manually and send an intermediate struct.

	var method string
	var params any
	var result any
	var errorObj any
	var id any

	// Use reflection or type assertion to get fields?
	// jsonrpc.Message is interface { GetID() interface{}; ... }? No, it's just a marker context?
	// We cast to specific types.

	if req, ok := msg.(*jsonrpc.Request); ok {
		method = req.Method
		params = req.Params
		id = fixID(req.ID)
	} else if resp, ok := msg.(*jsonrpc.Response); ok {
		result = resp.Result
		errorObj = resp.Error
		id = fixID(resp.ID)
	}

	wire := map[string]any{
		"jsonrpc": "2.0",
	}
	if method != "" {
		wire["method"] = method
	}
	if params != nil {
		wire["params"] = params
	}
	if id != nil {
		wire["id"] = id
	}
	if result != nil {
		wire["result"] = result
	}
	if errorObj != nil {
		wire["error"] = errorObj
	}

	c.log.Debug("Writing wire message", "wire", wire)
	return c.encoder.Encode(wire)
}

func fixID(id interface{}) interface{} {
	if id == nil {
		return nil
	}
	// Check if it's already simple type
	switch v := id.(type) {
	case string, int, int64, float64:
		return v
	}

	// If it's the broken struct, print it and parse
	// This is fragile, but needed until SDK exports ID or provides a way to marshal it.
	s := fmt.Sprintf("%+v", id)
	// Parse string value, handling potential closing braces in the content
	// Format is {value:<content>}
	if strings.HasPrefix(s, "{value:") && strings.HasSuffix(s, "}") {
		content := s[7 : len(s)-1]
		// Try to maintain integer type if possible to avoid regressions
		if i, err := strconv.Atoi(content); err == nil {
			return i
		}
		return content
	}

	// Expect {value:1}
	matches := idValueIntRegex.FindStringSubmatch(s)
	if len(matches) > 1 {
		if i, err := strconv.Atoi(matches[1]); err == nil {
			return i
		}
	}

	matchesStr := idValueStrRegex.FindStringSubmatch(s)
	if len(matchesStr) > 1 {
		return matchesStr[1]
	}

	// If parsing fails, return the ID as is and hope for the best (it might be marshaled as {})
	return id
}

// Close closes the connection.
//
// Returns an error if the operation fails.
func (c *bundleDockerConn) Close() error {
	return c.rwc.Close()
}

// SessionID returns the session ID of the connection.
//
// Returns the result.
func (c *bundleDockerConn) SessionID() string {
	return "bundle-docker"
}

// bundleSlogWriter duplicates slogWriter from docker_transport.go.
type bundleSlogWriter struct {
	log   *slog.Logger
	level slog.Level
}

// Write writes the log message to the logger.
//
// p is the p.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *bundleSlogWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	s.log.Log(context.Background(), s.level, msg)
	return len(p), nil
}
