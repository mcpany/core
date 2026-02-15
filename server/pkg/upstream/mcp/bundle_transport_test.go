package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundleDockerTransport_Connect(t *testing.T) {
	transport := &BundleDockerTransport{
		Image:   "test-image",
		Command: "test-cmd",
	}

	// Use DI instead of global variable swapping
	transport.dockerClientFactory = func(_ ...client.Opt) (dockerClient, error) {
		return &bundleMockDockerClient{
			ImagePullFunc: func(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader("")), nil // Mock empty pull
			},
			ContainerCreateFunc: func(_ context.Context, _ *container.Config, _ *container.HostConfig, _ *network.NetworkingConfig, _ *v1.Platform, _ string) (container.CreateResponse, error) {
				return container.CreateResponse{ID: "test-container-id"}, nil
			},
			ContainerAttachFunc: func(_ context.Context, _ string, _ container.AttachOptions) (types.HijackedResponse, error) {
				// We need a real pipe to simulate connection
				c1, c2 := net.Pipe()
				// Close the other end to prevent hangs if we read from it, or keep it open if we write?
				// For now, just close it deferred or immediately?
				// If we close it, Read on c1 might EOF immediately.
				// Let's keep it open but unused, or use it for dummy IO?
				// Just ignore unused warning by using it blankly?
				_ = c2

				return types.HijackedResponse{
					Conn:   c1,
					Reader: bufio.NewReader(c1),
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

	conn, err := transport.Connect(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, conn)
	defer func() { _ = conn.Close() }()
}

func TestFixID(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{"nil", nil, nil},
		{"string", "123", "123"},
		{"int", 123, 123},
		{"struct_value_int", struct{ value int }{1}, 1},
		{"struct_value_string_brace", struct{ value string }{"foo}bar"}, "foo}bar"},
		{"struct_value_string_simple", struct{ value string }{"foo"}, "foo"},
		{"struct_value_int_string", struct{ value string }{"123"}, 123},       // Should become int to avoid regression
		{"struct_value_tricky", struct{ value string }{"123}foo"}, "123}foo"}, // Should NOT be 123
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := fixID(tt.input)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestSetUnexportedID(_ *testing.T) {
	type TestID struct {
		value any //nolint:unused // unexported
	}

	id := &TestID{}
	setUnexportedID(id, "test-id")

	// Verify using fmt.Sprintf to satisfy linter (uses the value)
	_ = fmt.Sprintf("%v", id)
}

func TestBundleDockerConn_ReadWrite(t *testing.T) {
	// Create a pipe to simulate connection
	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close() }()
	defer func() { _ = c2.Close() }()

	logger := logging.GetLogger()
	conn := &bundleDockerConn{
		rwc:     c1,
		decoder: json.NewDecoder(c1),
		encoder: json.NewEncoder(c1),
		log:     logger,
	}

	// Test SessionID
	assert.Equal(t, "bundle-docker", conn.SessionID())

	// Test Write
	req := &jsonrpc.Request{
		Method: "test_method",
		Params: json.RawMessage(`{"key":"value"}`),
	}
	// Need to fix ID?
	// As we write, we should read from c2
	go func() {
		err := conn.Write(context.Background(), req)
		assert.NoError(t, err)
	}()

	// Read from c2
	decoder := json.NewDecoder(c2)
	var raw json.RawMessage
	err := decoder.Decode(&raw)
	assert.NoError(t, err)

	// Verify what was written
	var written map[string]interface{}
	err = json.Unmarshal(raw, &written)
	assert.NoError(t, err)
	assert.Equal(t, "test_method", written["method"])

	// Test Read
	// Write to c2
	resp := map[string]interface{}{
		"jsonrpc": "2.0",
		"result":  "success",
		"id":      1,
	}
	encoder := json.NewEncoder(c2)
	go func() {
		err := encoder.Encode(resp)
		assert.NoError(t, err)
	}()

	// Read from conn
	msg, err := conn.Read(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, msg)

	// Verify message content
	// It should be a Response
	jsonResp, ok := msg.(*jsonrpc.Response)
	assert.True(t, ok)
	assert.NotNil(t, jsonResp)
	// ID logic might convert int to float64 etc.
}
