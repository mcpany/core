// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
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
)

// bundleMockDockerClient for robustness tests
// Defined here to avoid dependency on other test files which might not be compiled or available in all contexts.
type bundleMockDockerClient struct {
	ImagePullFunc       func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)
	ContainerCreateFunc func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
	ContainerAttachFunc func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error)
	ContainerStartFunc  func(ctx context.Context, container string, options container.StartOptions) error
	ContainerStopFunc   func(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemoveFunc func(ctx context.Context, containerID string, options container.RemoveOptions) error
	CloseFunc           func() error
}

func (m *bundleMockDockerClient) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	if m.ImagePullFunc != nil {
		return m.ImagePullFunc(ctx, ref, options)
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (m *bundleMockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	if m.ContainerCreateFunc != nil {
		return m.ContainerCreateFunc(ctx, config, hostConfig, networkingConfig, platform, containerName)
	}
	return container.CreateResponse{ID: "mock-container-id"}, nil
}

func (m *bundleMockDockerClient) ContainerAttach(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
	if m.ContainerAttachFunc != nil {
		return m.ContainerAttachFunc(ctx, container, options)
	}
	return types.HijackedResponse{}, nil
}

func (m *bundleMockDockerClient) ContainerStart(ctx context.Context, container string, options container.StartOptions) error {
	if m.ContainerStartFunc != nil {
		return m.ContainerStartFunc(ctx, container, options)
	}
	return nil
}

func (m *bundleMockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	if m.ContainerStopFunc != nil {
		return m.ContainerStopFunc(ctx, containerID, options)
	}
	return nil
}

func (m *bundleMockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	if m.ContainerRemoveFunc != nil {
		return m.ContainerRemoveFunc(ctx, containerID, options)
	}
	return nil
}

func (m *bundleMockDockerClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func TestBundleDockerConn_Read_Robustness(t *testing.T) {
	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close(); _ = c2.Close() }()

	conn := &bundleDockerConn{
		rwc:     c1,
		decoder: json.NewDecoder(c1),
		encoder: json.NewEncoder(c1),
		log:     logging.GetLogger(),
	}

	// Test 1: Invalid JSON
	go func() {
		_, _ = c2.Write([]byte("{invalid json"))
		_ = c2.Close()
	}()

	_, err := conn.Read(context.Background())
	assert.Error(t, err)

	// Test 2: Valid JSON but malformed RPC (header missing or invalid)
	c3, c4 := net.Pipe()
	defer func() { _ = c3.Close(); _ = c4.Close() }()
	conn2 := &bundleDockerConn{
		rwc:     c3,
		decoder: json.NewDecoder(c3),
		encoder: json.NewEncoder(c3),
		log:     logging.GetLogger(),
	}

	go func() {
		_, _ = c4.Write([]byte("[]"))
		_ = c4.Close()
	}()

	_, err = conn2.Read(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal message header")
}

func TestBundleDockerConn_Write_Robustness(t *testing.T) {
	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close(); _ = c2.Close() }()

	conn := &bundleDockerConn{
		rwc:     c1,
		decoder: json.NewDecoder(c1),
		encoder: json.NewEncoder(c1),
		log:     logging.GetLogger(),
	}

	req := &jsonrpc.Request{
		Method: "test",
	}
	// Use helper to set unexported ID to integer 999
	// This simulates SDK behavior where ID is wrapped in struct
	err := setUnexportedID(&req.ID, 999)
	assert.NoError(t, err)

	go func() {
		err := conn.Write(context.Background(), req)
		assert.NoError(t, err)
	}()

	var raw map[string]interface{}
	decoder := json.NewDecoder(c2)
	err = decoder.Decode(&raw)
	assert.NoError(t, err)

	// Verify ID is flattened
	assert.Equal(t, float64(999), raw["id"])
}

func TestBundleDockerTransport_Connect_Robustness(t *testing.T) {
	// No global variable swapping anymore!

	transport := &BundleDockerTransport{
		Image:   "test-image",
		Command: "test-cmd",
	}

	t.Run("ImagePull_Fail", func(t *testing.T) {
		transport.dockerClientFactory = func(_ ...client.Opt) (dockerClient, error) {
			return &bundleMockDockerClient{
				ImagePullFunc: func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
					return nil, errors.New("pull failed")
				},
				ContainerCreateFunc: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
					return container.CreateResponse{ID: "id"}, nil
				},
				ContainerAttachFunc: func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
					c1, c2 := net.Pipe()
					_ = c2 // Prevent unused warning, we just need c1 to be valid
					return types.HijackedResponse{
						Conn:   c1,
						Reader: bufio.NewReader(c1),
					}, nil
				},
			}, nil
		}
		conn, err := transport.Connect(context.Background())
		assert.NoError(t, err)
		if conn != nil {
			_ = conn.Close()
		}
	})

	t.Run("ContainerCreate_Fail", func(t *testing.T) {
		transport.dockerClientFactory = func(_ ...client.Opt) (dockerClient, error) {
			return &bundleMockDockerClient{
				ImagePullFunc: func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("")), nil
				},
				ContainerCreateFunc: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
					return container.CreateResponse{}, errors.New("create failed")
				},
			}, nil
		}
		_, err := transport.Connect(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create container")
	})

	t.Run("ContainerAttach_Fail", func(t *testing.T) {
		removedID := ""
		transport.dockerClientFactory = func(_ ...client.Opt) (dockerClient, error) {
			return &bundleMockDockerClient{
				ImagePullFunc: func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("")), nil
				},
				ContainerCreateFunc: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
					return container.CreateResponse{ID: "created-id"}, nil
				},
				ContainerAttachFunc: func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
					return types.HijackedResponse{}, errors.New("attach failed")
				},
				ContainerRemoveFunc: func(ctx context.Context, containerID string, options container.RemoveOptions) error {
					removedID = containerID
					return nil
				},
			}, nil
		}
		_, err := transport.Connect(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to attach")
		assert.Equal(t, "created-id", removedID)
	})

	t.Run("ContainerStart_Fail", func(t *testing.T) {
		removedID := ""
		transport.dockerClientFactory = func(_ ...client.Opt) (dockerClient, error) {
			return &bundleMockDockerClient{
				ImagePullFunc: func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("")), nil
				},
				ContainerCreateFunc: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
					return container.CreateResponse{ID: "created-id"}, nil
				},
				ContainerAttachFunc: func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
					c1, c2 := net.Pipe()
					_ = c2
					return types.HijackedResponse{
						Conn:   c1,
						Reader: bufio.NewReader(c1),
					}, nil
				},
				ContainerStartFunc: func(ctx context.Context, container string, options container.StartOptions) error {
					return errors.New("start failed")
				},
				ContainerRemoveFunc: func(ctx context.Context, containerID string, options container.RemoveOptions) error {
					removedID = containerID
					return nil
				},
			}, nil
		}
		_, err := transport.Connect(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start")
		assert.Equal(t, "created-id", removedID)
	})
}

func TestSetUnexportedID_Robustness(t *testing.T) {
	// 1. Success case (mimicking SDK struct)
	type ID struct {
		value interface{} // unexported
	}
	id := &ID{}
	err := setUnexportedID(id, "test")
	assert.NoError(t, err)

	// 2. Struct missing 'value' field
	type IDMissing struct {
		other int
	}
	idMissing := &IDMissing{}
	err = setUnexportedID(idMissing, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field 'value' not found")

	// 3. Passing non-pointer (should panic because Elem() is called)
	assert.Panics(t, func() {
		_ = setUnexportedID(ID{}, "test")
	}, "Should panic if not passed a pointer")

	// 4. Passing nil (should panic)
	assert.Panics(t, func() {
		_ = setUnexportedID(nil, "test")
	})

	// 5. Field 'value' exists but is specific type
	type IDInt struct {
		value int
	}
	idInt := &IDInt{}
	err = setUnexportedID(idInt, 123)
	assert.NoError(t, err)

	// 6. Type mismatch (should panic via reflect.Set)
	assert.Panics(t, func() {
		_ = setUnexportedID(idInt, "string")
	})
}

func TestFixID_Robustness(t *testing.T) {
	// Helper struct mimicking SDK ID
	type ID struct {
		value interface{}
	}

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		// Basic types
		{"nil", nil, nil},
		{"string", "123", "123"},
		{"int", 123, 123},
		{"float64_int", 123.0, 123.0},
		{"float64_dec", 123.456, 123.456},

		// Structs mimicking SDK ID (unexported field value)
		// NOTE: In the new implementation, fixID with values uses Regex fallback.
		{"struct_value_int", struct{ value int }{1}, 1},
		{"struct_value_string", struct{ value string }{"s1"}, "s1"},

		// Pointer to struct (Safe Reflection path)
		{"ptr_struct_value_int", &ID{value: 1}, 1},
		{"ptr_struct_value_string", &ID{value: "s1"}, "s1"},
		{"ptr_struct_value_string_int", &ID{value: "123"}, 123},

		// Robustness / Edge cases
		{"struct_wrong_field", struct{ id int }{1}, struct{ id int }{1}},
		{"string_looks_like_struct", "{value:123}", "{value:123}"}, // input is string, should return string
		{"map_looks_like_struct", map[string]interface{}{"value": 123}, 123},

		// Regex fallback cases
		// Note: The heuristic is aggressive. A struct with extra fields might be partially matched.
		// "value:1 extra:2" is matched by string regex because int regex fails (perhaps due to greedy matching or ordering?)
		// This documents current behavior: fixID is fragile for structs with multiple fields.
		{"struct_extra_field_int", struct {
			value int
			extra int
		}{1, 2}, "1 extra:2"},

		// The current implementation of fixID regex `value:([^}]+)` is greedy until '}'.
		// If we have {value:foo extra:2}, it matches "foo extra:2".
		// This test documents that behavior (preserving existing behavior).
		{"struct_extra_field_string", struct {
			value string
			extra int
		}{"foo", 2}, "foo extra:2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := fixID(tt.input)
			assert.Equal(t, tt.expected, res)
		})
	}
}
