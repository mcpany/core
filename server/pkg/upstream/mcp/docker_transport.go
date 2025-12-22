package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// dockerClient is an interface that abstracts the Docker client methods used by DockerTransport.
// It is used for testing purposes to allow mocking of the Docker client.
type dockerClient interface {
	ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
	ContainerAttach(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error)
	ContainerStart(ctx context.Context, container string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	Close() error
}

var newDockerClient = func(ops ...client.Opt) (dockerClient, error) {
	return client.NewClientWithOpts(ops...)
}

// DockerTransport implements the mcp.Transport interface to connect to a service
// running inside a Docker container. It manages the container lifecycle.
type DockerTransport struct {
	StdioConfig *configv1.McpStdioConnection
}

// Connect establishes a connection to the service within the Docker container.
func (t *DockerTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	log := logging.GetLogger()
	cli, err := newDockerClient(client.FromEnv, client.WithAPIVersionNegotiation())
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

	go func() {
		defer func() { _ = stdoutWriter.Close() }()
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
		rwc:     rwc,
		decoder: json.NewDecoder(rwc),
		encoder: json.NewEncoder(rwc),
	}, nil
}

// dockerConn provides a concrete implementation of the mcp.Connection interface,
// tailored for communication with a service running in a Docker container.
type dockerConn struct {
	rwc     io.ReadWriteCloser
	decoder *json.Decoder
	encoder *json.Encoder
}

// Read decodes a single JSON-RPC message from the container's output stream.
func (c *dockerConn) Read(_ context.Context) (jsonrpc.Message, error) {
	var raw json.RawMessage
	if err := c.decoder.Decode(&raw); err != nil {
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
	return c.encoder.Encode(msg)
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
	cli         dockerClient
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

	_ = c.cli.Close()
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
