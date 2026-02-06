// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"al.essio.dev/pkg/shellescape"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// dockerClient is an interface that abstracts the Docker client methods used by DockerTransport.
// It is used for testing purposes to allow mocking of the Docker client.
//
// Summary: Abstract interface for Docker client operations.
type dockerClient interface {
	ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
	ContainerAttach(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error)
	ContainerStart(ctx context.Context, container string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	Close() error
}

// newDockerClient creates a new Docker client instance.
//
// Summary: Factory function for creating a Docker client.
//
// Parameters:
//   - ops: ...client.Opt. Optional configuration options for the client.
//
// Returns:
//   - dockerClient: The created Docker client interface.
//   - error: An error if client creation fails.
var newDockerClient = func(ops ...client.Opt) (dockerClient, error) {
	return client.NewClientWithOpts(ops...)
}

// DockerTransport implements the mcp.Transport interface to connect to a service
// running inside a Docker container. It manages the container lifecycle.
//
// Summary: MCP Transport implementation for Docker containers.
type DockerTransport struct {
	// StdioConfig holds the configuration for running the container in stdio mode.
	StdioConfig *configv1.McpStdioConnection
}

// Connect establishes a connection to the service within the Docker container.
//
// Summary: Starts the container and establishes a connection.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//
// Returns:
//   - mcp.Connection: The established connection.
//   - error: An error if the operation fails.
func (t *DockerTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	log := logging.GetLogger()
	cli, err := newDockerClient(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	// Ensure the client is closed if connection setup fails
	success := false
	defer func() {
		if !success {
			_ = cli.Close()
		}
	}()

	img := t.StdioConfig.GetContainerImage()
	if img == "" {
		return nil, fmt.Errorf("container_image must be specified for docker transport")
	}

	reader, err := cli.ImagePull(ctx, img, image.PullOptions{})
	if err != nil {
		log.Warn("Failed to pull docker image, will try to use local image if available", "image", img, "error", err)
	} else {
		_, _ = io.Copy(io.Discard, reader)
		log.Info("Successfully pulled docker image", "image", img)
	}

	setupCmds := t.StdioConfig.GetSetupCommands()

	// Sentinel Security: Disable setup_commands by default as they allow arbitrary command execution.
	if len(setupCmds) > 0 {
		if os.Getenv("MCP_ALLOW_UNSAFE_SETUP_COMMANDS") != "true" {
			return nil, fmt.Errorf("setup_commands are disabled by default for security reasons. Set MCP_ALLOW_UNSAFE_SETUP_COMMANDS=true to enable them if you trust the configuration")
		}
		log.Warn("Using setup_commands in DockerTransport is dangerous and allows Command Injection if config is untrusted.", "setup_commands", "HIDDEN")
	}

	// Allocate slice with capacity for setup commands + 1 main command
	scriptCommands := make([]string, 0, len(setupCmds)+1)

	// Redirect stdout of setup commands to stderr to avoid polluting the JSON-RPC channel
	for _, cmd := range setupCmds {
		scriptCommands = append(scriptCommands, fmt.Sprintf("(%s) >&2", cmd))
	}

	// Add the main command. `exec` is used to replace the shell process with the main command.
	mainCommandParts := []string{"exec", shellescape.Quote(t.StdioConfig.GetCommand())}
	for _, arg := range t.StdioConfig.GetArgs() {
		mainCommandParts = append(mainCommandParts, shellescape.Quote(arg))
	}
	scriptCommands = append(scriptCommands, strings.Join(mainCommandParts, " "))
	script := strings.Join(scriptCommands, " && ")

	// Prepare environment variables
	resolvedEnv, err := util.ResolveSecretMap(ctx, t.StdioConfig.GetEnv(), nil)
	if err != nil {
		return nil, err
	}

	envVars := make([]string, 0, len(resolvedEnv))
	for k, v := range resolvedEnv {
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        img,
		Cmd:          []string{"/bin/sh", "-c", script},
		WorkingDir:   t.StdioConfig.GetWorkingDirectory(),
		Env:          envVars,
		Tty:          false,
		OpenStdin:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	}, nil, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	log.Info("Container created", "id", resp.ID, "env", envVars)

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
	log.Info("Attached to container", "id", resp.ID)

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		_ = cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true})
		return nil, fmt.Errorf("failed to start container: %w", err)
	}
	log.Info("Container started", "id", resp.ID)

	stdoutReader, stdoutWriter := io.Pipe()

	// Capture stderr for better error reporting on early failure
	stderrCapture := &tailBuffer{limit: 4096}

	go func() {
		defer func() { _ = stdoutWriter.Close() }()
		logWriter := &slogWriter{log: log, level: slog.LevelError}
		// Write stderr to both capture buffer and log
		multiStderr := io.MultiWriter(logWriter, stderrCapture)
		_, err := stdcopy.StdCopy(stdoutWriter, multiStderr, hijackedResp.Reader)
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
	success = true
	return &dockerConn{
		rwc:           rwc,
		decoder:       json.NewDecoder(rwc),
		encoder:       json.NewEncoder(rwc),
		stderrCapture: stderrCapture,
	}, nil
}

// dockerConn provides a concrete implementation of the mcp.Connection interface,
// tailored for communication with a service running in a Docker container.
//
// Summary: Connection implementation for Docker transport.
type dockerConn struct {
	rwc           io.ReadWriteCloser
	decoder       *json.Decoder
	encoder       *json.Encoder
	stderrCapture *tailBuffer
}

// Read decodes a single JSON-RPC message from the container's output stream.
//
// Summary: Reads and decodes the next JSON-RPC message.
//
// Parameters:
//   - _: context.Context. The context (unused in this implementation as Read is blocking).
//
// Returns:
//   - jsonrpc.Message: The decoded message (Request or Response).
//   - error: An error if reading or decoding fails.
func (c *dockerConn) Read(_ context.Context) (jsonrpc.Message, error) {
	var raw json.RawMessage
	if err := c.decoder.Decode(&raw); err != nil {
		if err == io.EOF {
			// If we hit EOF, check if there was any stderr output captured.
			// This usually indicates the container exited early (e.g. wrong command or config).
			stderr := c.stderrCapture.String()
			if stderr != "" {
				return nil, fmt.Errorf("connection closed. Stderr: %s", stderr)
			}
		}
		return nil, err
	}

	var header struct {
		Method string `json:"method"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message header: %w", err)
	}

	var msg jsonrpc.Message
	isRequest := header.Method != ""

	if isRequest {
		req := &jsonrpc.Request{}
		if err := json.Unmarshal(raw, req); err != nil {
			// Alternative: Unmarshal into a temporary struct that matches Request/Response but with Any ID.
			type requestAnyID struct {
				Method string          `json:"method"`
				Params json.RawMessage `json:"params,omitempty"`
				ID     any             `json:"id,omitempty"`
			}
			var rAny requestAnyID
			if err2 := json.Unmarshal(raw, &rAny); err2 != nil {
				return nil, fmt.Errorf("failed to unmarshal request: %w (and %v)", err2, err)
			}
			req = &jsonrpc.Request{
				Method: rAny.Method,
				Params: rAny.Params,
			}
			if err := setUnexportedID(&req.ID, rAny.ID); err != nil {
				logging.GetLogger().Error("Failed to set unexported ID on request", "error", err)
			}
			msg = req
		} else {
			msg = req
		}
	} else {
		resp := &jsonrpc.Response{}
		if err := json.Unmarshal(raw, resp); err != nil {
			// Use alias struct
			type responseAnyID struct {
				Result json.RawMessage `json:"result,omitempty"`
				Error  *transportError `json:"error,omitempty"`
				ID     any             `json:"id,omitempty"`
			}
			var rAny responseAnyID
			if err2 := json.Unmarshal(raw, &rAny); err2 != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w (and %v)", err2, err)
			}
			resp = &jsonrpc.Response{
				Result: rAny.Result,
			}
			if rAny.Error != nil {
				resp.Error = rAny.Error
			}
			if err := setUnexportedID(&resp.ID, rAny.ID); err != nil {
				logging.GetLogger().Error("Failed to set unexported ID on response", "error", err)
			}
			msg = resp
		} else {
			msg = resp
		}
	}

	return msg, nil
}

// Write encodes and sends a JSON-RPC message to the container's input stream.
//
// Summary: Encodes and writes a JSON-RPC message.
//
// Parameters:
//   - _: context.Context. The context (unused).
//   - msg: jsonrpc.Message. The message to send.
//
// Returns:
//   - error: An error if encoding or writing fails.
func (c *dockerConn) Write(_ context.Context, msg jsonrpc.Message) error {
	var method string
	var params any
	var result any
	var errorObj any
	var id any

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

	return c.encoder.Encode(wire)
}

// Close terminates the connection by closing the underlying ReadWriteCloser.
//
// Summary: Closes the connection.
//
// Returns:
//   - error: An error if closing fails.
func (c *dockerConn) Close() error {
	return c.rwc.Close()
}

// SessionID returns a static identifier for the Docker transport session.
//
// Summary: Returns the session ID.
//
// Returns:
//   - string: "docker-transport-session".
func (c *dockerConn) SessionID() string {
	return "docker-transport-session"
}

// dockerReadWriteCloser combines an io.Reader and an io.WriteCloser and manages container cleanup.
//
// Summary: Helper struct for managing container I/O and lifecycle.
type dockerReadWriteCloser struct {
	io.Reader
	io.WriteCloser
	containerID string
	cli         dockerClient
}

// Close closes the underlying connection and removes the associated Docker container.
//
// Summary: Closes the connection and cleans up the container.
//
// Returns:
//   - error: An error if closing or cleanup fails.
func (c *dockerReadWriteCloser) Close() error {
	err := c.WriteCloser.Close()

	ctx := context.Background()
	timeout := 10
	stopOptions := container.StopOptions{Timeout: &timeout}
	if stopErr := c.cli.ContainerStop(ctx, c.containerID, stopOptions); stopErr != nil {
		logging.GetLogger().Error("Failed to stop container", "containerID", c.containerID, "error", stopErr)
	}

	if rmErr := c.cli.ContainerRemove(ctx, c.containerID, container.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); rmErr != nil {
		logging.GetLogger().Error("Failed to remove container", "containerID", c.containerID, "error", rmErr)
	}

	_ = c.cli.Close()
	return err
}

// slogWriter implements the io.Writer interface, allowing it to be used as a
// destination for log output. It writes each line of the input to a slog.Logger.
//
// Summary: Adapts io.Writer to slog.Logger.
type slogWriter struct {
	log   *slog.Logger
	level slog.Level
}

// Write takes a byte slice, scans it for lines, and logs each line
// individually using the configured slog.Logger and level.
//
// Summary: Writes data to the logger.
//
// Parameters:
//   - p: []byte. The data to write.
//
// Returns:
//   - int: The number of bytes written.
//   - error: Always nil.
func (s *slogWriter) Write(p []byte) (n int, err error) {
	scanner := bufio.NewScanner(strings.NewReader(string(p)))
	for scanner.Scan() {
		s.log.Log(context.Background(), s.level, scanner.Text())
	}
	return len(p), nil
}

// tailBuffer is a thread-safe buffer that keeps the last N bytes written to it.
//
// Summary: A circular-like buffer that keeps the tail of the data.
type tailBuffer struct {
	buf   []byte
	limit int
	mu    sync.Mutex
}

// Write writes data to the buffer, maintaining the size limit.
//
// Summary: Writes data, dropping old data if limit exceeded.
//
// Parameters:
//   - p: []byte. The data to write.
//
// Returns:
//   - int: The number of bytes written.
//   - error: Always nil.
func (b *tailBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buf = append(b.buf, p...)
	if len(b.buf) > b.limit {
		// Keep the last 'limit' bytes
		b.buf = b.buf[len(b.buf)-b.limit:]
	}
	return len(p), nil
}

// String returns the buffered data as a string.
//
// Summary: Returns the buffer content as string.
//
// Returns:
//   - string: The buffered data.
func (b *tailBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.buf)
}
