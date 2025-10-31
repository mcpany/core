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

package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/mcpany/core/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DockerTransport implements the mcp.Transport interface to connect to a service
// running inside a Docker container. It manages the container lifecycle.
type DockerTransport struct {
	StdioConfig *configv1.McpStdioConnection
}

// Connect establishes a connection to the service within the Docker container.
func (t *DockerTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	log := logging.GetLogger()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

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

	var scriptCommands []string
	scriptCommands = append(scriptCommands, t.StdioConfig.GetSetupCommands()...)
	mainCommandParts := []string{"exec", t.StdioConfig.GetCommand()}
	mainCommandParts = append(mainCommandParts, t.StdioConfig.GetArgs()...)
	scriptCommands = append(scriptCommands, strings.Join(mainCommandParts, " "))
	script := strings.Join(scriptCommands, " && ")

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        img,
		Cmd:          []string{"/bin/sh", "-c", script},
		WorkingDir:   t.StdioConfig.GetWorkingDirectory(),
		Tty:          false,
		OpenStdin:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	}, nil, nil, nil, "")
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

	go func() {
		defer stdoutWriter.Close()
		logWriter := &slogWriter{log: log, level: slog.LevelError}
		_, err := stdcopy.StdCopy(stdoutWriter, logWriter, hijackedResp.Reader)
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
	return &dockerConn{
		rwc: rwc,
		dec: json.NewDecoder(rwc),
		enc: json.NewEncoder(rwc),
	}, nil
}

// dockerConn provides a concrete implementation of the mcp.Connection interface,
// tailored for communication with a service running in a Docker container.
type dockerConn struct {
	rwc io.ReadWriteCloser
	dec *json.Decoder
	enc *json.Encoder
}

// Read decodes a single JSON-RPC message from the container's output stream.
func (c *dockerConn) Read(_ context.Context) (jsonrpc.Message, error) {
	var raw json.RawMessage
	if err := c.dec.Decode(&raw); err != nil {
		return nil, err
	}

	var header struct {
		Method string `json:"method"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message header: %w", err)
	}

	var msg jsonrpc.Message
	if header.Method != "" {
		msg = &jsonrpc.Request{}
	} else {
		msg = &jsonrpc.Response{}
	}

	if err := json.Unmarshal(raw, msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return msg, nil
}

// Write encodes and sends a JSON-RPC message to the container's input stream.
func (c *dockerConn) Write(_ context.Context, msg jsonrpc.Message) error {
	return c.enc.Encode(msg)
}

// Close terminates the connection by closing the underlying ReadWriteCloser.
func (c *dockerConn) Close() error {
	return c.rwc.Close()
}

// SessionID returns a static identifier for the Docker transport session.
func (c *dockerConn) SessionID() string {
	return "docker-transport-session"
}

// dockerReadWriteCloser combines an io.Reader and an io.WriteCloser and manages container cleanup.
type dockerReadWriteCloser struct {
	io.Reader
	io.WriteCloser
	containerID string
	cli         *client.Client
}

// Close closes the underlying connection and removes the associated Docker container.
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

	c.cli.Close()
	return err
}

// slogWriter implements the io.Writer interface, allowing it to be used as a
// destination for log output. It writes each line of the input to a slog.Logger.
type slogWriter struct {
	log   *slog.Logger
	level slog.Level
}

// Write takes a byte slice, scans it for lines, and logs each line
// individually using the configured slog.Logger and level.
func (s *slogWriter) Write(p []byte) (n int, err error) {
	scanner := bufio.NewScanner(strings.NewReader(string(p)))
	for scanner.Scan() {
		s.log.Log(context.Background(), s.level, scanner.Text())
	}
	return len(p), nil
}
