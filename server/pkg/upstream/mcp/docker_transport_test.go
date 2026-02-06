package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/mcpany/core/server/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSlogWriter(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	writer := &slogWriter{log: log, level: slog.LevelInfo}

	testMessage := "Hello, world!"
	_, err := writer.Write([]byte(testMessage))
	assert.NoError(t, err)

	assert.Contains(t, buf.String(), testMessage)
}

func TestDockerConn_SessionID(t *testing.T) {
	conn := &dockerConn{}
	assert.Equal(t, "docker-transport-session", conn.SessionID())
}

type mockReadWriteCloser struct {
	bytes.Buffer
}

func (m *mockReadWriteCloser) Close() error {
	return nil
}

func (m *mockReadWriteCloser) LocalAddr() net.Addr                { return nil }
func (m *mockReadWriteCloser) RemoteAddr() net.Addr               { return nil }
func (m *mockReadWriteCloser) SetDeadline(t time.Time) error      { return nil }
func (m *mockReadWriteCloser) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockReadWriteCloser) SetWriteDeadline(t time.Time) error { return nil }

func TestDockerConn_ReadWrite(t *testing.T) {
	ctx := context.Background()
	rwc := &mockReadWriteCloser{}
	conn := &dockerConn{
		rwc:     rwc,
		decoder: json.NewDecoder(rwc),
		encoder: json.NewEncoder(rwc),
	}

	// Test Write
	testMsg := &jsonrpc.Request{
		Method: "test",
	}
	err := conn.Write(ctx, testMsg)
	assert.NoError(t, err)

	// Test Read
	readMsg, err := conn.Read(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, readMsg)

	// Test Close
	err = conn.Close()
	assert.NoError(t, err)
}

func TestDockerConn_Read_UnmarshalError(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid header", func(t *testing.T) {
		rwc := &mockReadWriteCloser{}
		conn := &dockerConn{
			rwc:     rwc,
			decoder: json.NewDecoder(rwc),
			encoder: json.NewEncoder(rwc),
		}
		invalidHeaderMsg := `{"method": 123}`
		_, _ = rwc.WriteString(invalidHeaderMsg + "\n")

		_, err := conn.Read(ctx)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to unmarshal message header")
		}
	})

	t.Run("invalid json syntax", func(t *testing.T) {
		rwc := &mockReadWriteCloser{}
		conn := &dockerConn{
			rwc:     rwc,
			decoder: json.NewDecoder(rwc),
			encoder: json.NewEncoder(rwc),
		}
		// This is syntactically invalid, and will cause `decoder.Decode` to fail.
		invalidMsg := `{"method": "test"`
		_, _ = rwc.WriteString(invalidMsg + "\n")

		_, err := conn.Read(ctx)
		assert.Error(t, err)
	})
}

func TestDockerTransport_Connect_ClientError(t *testing.T) {
	originalNewDockerClient := newDockerClient
	newDockerClient = func(_ ...client.Opt) (dockerClient, error) {
		return nil, fmt.Errorf("client error")
	}
	defer func() { newDockerClient = originalNewDockerClient }()

	stdioConfig := configv1.McpStdioConnection_builder{
		ContainerImage: proto.String("test-image"),
	}.Build()
	transport := &DockerTransport{StdioConfig: stdioConfig}
	_, err := transport.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create docker client")
}

func TestDockerTransport_Connect_ContainerCreateError(t *testing.T) {
	originalNewDockerClient := newDockerClient
	newDockerClient = func(_ ...client.Opt) (dockerClient, error) {
		return &mockDockerClient{
			ImagePullFunc: func(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte{})), nil
			},
			ContainerCreateFunc: func(_ context.Context, _ *container.Config, _ *container.HostConfig, _ *network.NetworkingConfig, _ *v1.Platform, _ string) (container.CreateResponse, error) {
				return container.CreateResponse{}, fmt.Errorf("container create error")
			},
		}, nil
	}
	defer func() { newDockerClient = originalNewDockerClient }()

	stdioConfig := configv1.McpStdioConnection_builder{
		ContainerImage: proto.String("test-image"),
	}.Build()
	transport := &DockerTransport{StdioConfig: stdioConfig}
	_, err := transport.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create container")
}

func TestDockerTransport_Connect_ContainerAttachError(t *testing.T) {
	originalNewDockerClient := newDockerClient
	newDockerClient = func(_ ...client.Opt) (dockerClient, error) {
		return &mockDockerClient{
			ImagePullFunc: func(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte{})), nil
			},
			ContainerAttachFunc: func(_ context.Context, _ string, _ container.AttachOptions) (types.HijackedResponse, error) {
				return types.HijackedResponse{}, fmt.Errorf("container attach error")
			},
		}, nil
	}
	defer func() { newDockerClient = originalNewDockerClient }()

	stdioConfig := configv1.McpStdioConnection_builder{
		ContainerImage: proto.String("test-image"),
	}.Build()
	transport := &DockerTransport{StdioConfig: stdioConfig}
	_, err := transport.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to attach to container")
}

func TestDockerTransport_Connect_ContainerStartError(t *testing.T) {
	originalNewDockerClient := newDockerClient
	newDockerClient = func(_ ...client.Opt) (dockerClient, error) {
		return &mockDockerClient{
			ImagePullFunc: func(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte{})), nil
			},
			ContainerStartFunc: func(_ context.Context, _ string, _ container.StartOptions) error {
				return fmt.Errorf("container start error")
			},
		}, nil
	}
	defer func() { newDockerClient = originalNewDockerClient }()

	stdioConfig := configv1.McpStdioConnection_builder{
		ContainerImage: proto.String("test-image"),
	}.Build()
	transport := &DockerTransport{StdioConfig: stdioConfig}
	_, err := transport.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start container")
}

func TestDockerTransport_Connect_Integration(t *testing.T) {
	// Mock Docker Client to allow testing without real Docker
	originalNewDockerClient := newDockerClient
	defer func() { newDockerClient = originalNewDockerClient }()

	newDockerClient = func(_ ...client.Opt) (dockerClient, error) {
		return &mockDockerClient{
			ImagePullFunc: func(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte{})), nil
			},
			ContainerCreateFunc: func(_ context.Context, _ *container.Config, _ *container.HostConfig, _ *network.NetworkingConfig, _ *v1.Platform, _ string) (container.CreateResponse, error) {
				return container.CreateResponse{ID: "test-container"}, nil
			},
			ContainerAttachFunc: func(_ context.Context, _ string, _ container.AttachOptions) (types.HijackedResponse, error) {
				jsonPayload := []byte(`{"jsonrpc": "2.0", "id": "1", "result": "hello"}` + "\n")
				// Create multiplexed header for StdCopy
				// Header: [STREAM_TYPE, 0, 0, 0, SIZE1, SIZE2, SIZE3, SIZE4]
				// Stream Type: 1 = stdout
				header := make([]byte, 8)
				header[0] = 1
				binary.BigEndian.PutUint32(header[4:], uint32(len(jsonPayload)))

				frame := append(header, jsonPayload...)

				return types.HijackedResponse{
					Conn:   &mockReadWriteCloser{},
					Reader: bufio.NewReader(bytes.NewReader(frame)),
				}, nil
			},
			ContainerStartFunc: func(_ context.Context, _ string, _ container.StartOptions) error {
				return nil
			},
			ContainerStopFunc: func(_ context.Context, _ string, _ container.StopOptions) error {
				return nil
			},
			ContainerRemoveFunc: func(_ context.Context, _ string, _ container.RemoveOptions) error {
				return nil
			},
		}, nil
	}

	ctx := context.Background()
	stdioConfig := configv1.McpStdioConnection_builder{
		ContainerImage: proto.String("alpine:latest"),
		Command:        proto.String("printf"),
		Args:           []string{"test"},
	}.Build()
	transport := &DockerTransport{StdioConfig: stdioConfig}

	conn, err := transport.Connect(ctx)
	require.NoError(t, err)
	require.NotNil(t, conn)

	msg, err := conn.Read(ctx)
	assert.NoError(t, err)
	require.NotNil(t, msg)

	resp, ok := msg.(*jsonrpc.Response)
	assert.True(t, ok)
	assert.Equal(t, "1", resp.ID.Raw())
	assert.Equal(t, json.RawMessage(`"hello"`), resp.Result)

	err = conn.Close()
	assert.NoError(t, err)
}

func TestDockerTransport_Connect_ImageNotFound(t *testing.T) {
	// Mock Docker Client for ImagePull failure
	originalNewDockerClient := newDockerClient
	defer func() { newDockerClient = originalNewDockerClient }()

	newDockerClient = func(_ ...client.Opt) (dockerClient, error) {
		return &mockDockerClient{
			ImagePullFunc: func(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
				return nil, fmt.Errorf("image not found")
			},
			// Even if pull fails, transport logs warning and proceeds to create container.
			// The DockerTransport logic says: "Failed to pull... will try to use local image".
			// So CreateContainer is called next. We should fail there if we want to simulate failure,
			// or if the test expects success despite pull fail (local image), we mock success.
			// The original test `assert.Error(t, err)` implies it expects failure eventually.
			// But if pull fails and we proceed, where does it fail?
			// `cli.ContainerCreate`.
			ContainerCreateFunc: func(_ context.Context, _ *container.Config, _ *container.HostConfig, _ *network.NetworkingConfig, _ *v1.Platform, _ string) (container.CreateResponse, error) {
				return container.CreateResponse{}, fmt.Errorf("image not found locally")
			},
		}, nil
	}

	ctx := context.Background()
	stdioConfig := configv1.McpStdioConnection_builder{
		ContainerImage: proto.String("this-image-does-not-exist-ever:latest"),
		Command:        proto.String("echo"),
	}.Build()
	transport := &DockerTransport{StdioConfig: stdioConfig}

	_, err := transport.Connect(ctx)
	assert.Error(t, err)
}

func TestDockerTransport_Connect_NoImage(t *testing.T) {
	ctx := context.Background()
	stdioConfig := configv1.McpStdioConnection_builder{
		Command: proto.String("echo"),
	}.Build()
	transport := &DockerTransport{StdioConfig: stdioConfig}

	_, err := transport.Connect(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container_image must be specified")
}

func TestDockerReadWriteCloser_Close_Error(t *testing.T) {
	var buf bytes.Buffer
	logging.ForTestsOnlyResetLogger()
	logging.Init(slog.LevelInfo, &buf)

	mockClient := &mockDockerClient{
		ContainerStopFunc: func(_ context.Context, _ string, _ container.StopOptions) error {
			return fmt.Errorf("stop error")
		},
		ContainerRemoveFunc: func(_ context.Context, _ string, _ container.RemoveOptions) error {
			return fmt.Errorf("remove error")
		},
		CloseFunc: func() error {
			return nil
		},
	}

	rwc := &dockerReadWriteCloser{
		WriteCloser: &mockReadWriteCloser{},
		containerID: "test-container",
		cli:         mockClient,
	}

	err := rwc.Close()
	assert.NoError(t, err)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "Failed to stop container")
	assert.Contains(t, logOutput, "stop error")
	assert.Contains(t, logOutput, "Failed to remove container")
	assert.Contains(t, logOutput, "remove error")
}
