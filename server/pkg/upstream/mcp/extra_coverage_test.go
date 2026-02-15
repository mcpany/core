package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// Reuse mocks if possible, or define coverage specific ones
// We need access to newDockerClient, so we are in the same package `mcp`.

func TestDockerTransport_Connect_Success_Mock(t *testing.T) {
	originalNewDockerClient := newDockerClient
	defer func() { newDockerClient = originalNewDockerClient }()

	newDockerClient = func(_ ...client.Opt) (dockerClient, error) {
		return &mockDockerClient{
			ImagePullFunc: func(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader("")), nil
			},
			ContainerCreateFunc: func(_ context.Context, _ *container.Config, _ *container.HostConfig, _ *network.NetworkingConfig, _ *v1.Platform, _ string) (container.CreateResponse, error) {
				return container.CreateResponse{ID: "test-container-id"}, nil
			},
			ContainerAttachFunc: func(_ context.Context, _ string, _ container.AttachOptions) (types.HijackedResponse, error) {
				c1, c2 := net.Pipe()
				// We need to simulate the docker daemon side on c2
				go func() {
					defer c2.Close() //nolint:errcheck
					// Write some stdout/stderr to satisfy stdcopy
					// Stdcopy protocol: [STREAM_TYPE, 0, 0, 0, SIZE, ... PAYLOAD]
					// STREAM_TYPE: 1=stdout, 2=stderr
					// Let's just keep it open or write simple header
				}()
				return types.HijackedResponse{Conn: c1, Reader: bufio.NewReader(c1)}, nil
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
			CloseFunc: func() error { return nil },
		}, nil
	}

	stdioConfig := configv1.McpStdioConnection_builder{
		ContainerImage: proto.String("test-image"),
		Command:        proto.String("echo"),
		Args:           []string{"hello"},
		Env: map[string]*configv1.SecretValue{
			"foo": configv1.SecretValue_builder{
				PlainText: proto.String("bar"),
			}.Build(),
		},
	}.Build()

	transport := &DockerTransport{StdioConfig: stdioConfig}
	conn, err := transport.Connect(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	// Test Close
	err = conn.Close()
	assert.NoError(t, err)
}

func TestBundleDockerConn_Read_Fallback_Coverage(t *testing.T) {
	// Test the fallback logic in Read where standard unmarshal fails
	rwc := &mockRWCExtra{}
	conn := &bundleDockerConn{
		decoder: json.NewDecoder(rwc),
		encoder: json.NewEncoder(rwc),
		log:     logging.GetLogger(),
		rwc:     rwc,
	}

	t.Run("Request_ID_As_Struct", func(t *testing.T) {
		// Simulate a request where ID is {"value": 1} which might fail standard unmarshal
		// if SDK Request expects something else, OR if we rely on fallback
		// construct raw message
		raw := `{"jsonrpc": "2.0", "method": "foo", "id": {"value": 123}}`
		rwc.readBuf = bytes.NewBufferString(raw)
		conn.decoder = json.NewDecoder(rwc.readBuf)

		msg, err := conn.Read(context.Background())
		assert.NoError(t, err)
		req, ok := msg.(*jsonrpc.Request)
		assert.True(t, ok)
		_ = req // Use req
		// Check if ID was fixed
		// fixID parses "value:123" string representation if Unmarshal failed

		// Test `setUnexportedID` with complex ID via internal helper access?
		// We can't access setUnexportedID directly as it is unexported in bundle_transport.go
		// But it is used in Read.

		// Test `setUnexportedID` via reflection if we really want to be sure it works,
		// but checking `req.ID` (if exported) or side effects is better.
		// SDK jsonrpc.Request ID is NOT exported usually?
		// `jsonrpc.Request` struct definition:
		// type Request struct { ..., ID ID, ... }
		// type ID struct { value interface{} } // unexported value.
		// So we can't check ID easily.
	})

	t.Run("Fallback_Unmarshal_Request", func(t *testing.T) {
		// Try to force fallback by sending invalid type for a field that is strict in Request but strict in AnyID too?
		// We found that [1,2] for ID might work if AnyID treats ID as any.

		raw := `{"jsonrpc": "2.0", "method": "foo", "id": [1,2]}`
		rwc.readBuf = bytes.NewBufferString(raw)
		conn.decoder = json.NewDecoder(rwc.readBuf)

		msg, err := conn.Read(context.Background())
		assert.NoError(t, err)
		req, ok := msg.(*jsonrpc.Request)
		assert.True(t, ok)
		assert.Equal(t, "foo", req.Method)
	})
}

// mockRWCExtra for Read/Write
type mockRWCExtra struct {
	readBuf *bytes.Buffer
}

func (m *mockRWCExtra) Write(p []byte) (n int, err error) { return len(p), nil }
func (m *mockRWCExtra) Read(p []byte) (n int, err error) {
	if m.readBuf == nil {
		return 0, io.EOF
	}
	return m.readBuf.Read(p)
}

func (m *mockRWCExtra) Close() error { return nil }

type mockToolManagerExtra struct {
	tool.ManagerInterface
	tools map[string]tool.Tool
}

func (m *mockToolManagerExtra) GetTool(name string) (tool.Tool, bool) {
	if m.tools == nil {
		return nil, false
	}
	t, ok := m.tools[name]
	return t, ok
}

func (m *mockToolManagerExtra) ListServices() []*tool.ServiceInfo {
	return nil
}

func (m *mockToolManagerExtra) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

func TestRegister_DynamicResource_Success(t *testing.T) {
	u := &Upstream{}

	// Config with 1 tool and 1 dynamic resource pointing to it
	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("my-tool"),
		CallId: proto.String("call-1"),
	}.Build()

	resourceDef := configv1.ResourceDefinition_builder{
		Name: proto.String("my-resource"),
		Uri:  proto.String("test://my-resource"),
		Dynamic: configv1.DynamicResource_builder{
			McpCall: configv1.MCPCallDefinition_builder{
				Id: proto.String("call-1"),
			}.Build(),
		}.Build(),
	}.Build()

	mcpService := configv1.McpUpstreamService_builder{
		Tools:     []*configv1.ToolDefinition{toolDef},
		Resources: []*configv1.ResourceDefinition{resourceDef},
	}.Build()

	// Mock managers
	tm := &mockToolManagerExtra{
		tools: map[string]tool.Tool{
			"s1.my-tool": &tool.MCPTool{},
		},
	}
	rm := &mockResourceManagerCoverage{
		resources: make(map[string]resource.Resource),
	}

	u.registerDynamicResources("s1", mcpService, tm, rm, nil)

	// Check if resource added
	assert.Len(t, rm.resources, 1)
	assert.Contains(t, rm.resources, "test://my-resource")
}

func TestUpstream_Register_Stdio_Success(t *testing.T) {
	u := &Upstream{}
	ctx := context.Background()
	tm := &mockToolManagerCoverage{}
	pm := &mockPromptManagerCoverage{}
	rm := &mockResourceManagerCoverage{}

	config := configv1.UpstreamServiceConfig_builder{
		Name:             proto.String("test-stdio-success"),
		AutoDiscoverTool: proto.Bool(true),
		McpService: configv1.McpUpstreamService_builder{
			StdioConnection: configv1.McpStdioConnection_builder{
				Command: proto.String("echo"),
			}.Build(),
		}.Build(),
	}.Build()

	// Mock Connect logic
	originalConnect := connectForTesting
	defer func() { connectForTesting = originalConnect }()

	mockCS := &mockSessionCoverage{
		tools: &mcp.ListToolsResult{
			Tools: []*mcp.Tool{
				{Name: "tool1", Description: "desc1"},
			},
		},
	}
	SetConnectForTesting(func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		return mockCS, nil
	})

	_, tools, _, err := u.Register(ctx, config, tm, pm, rm, false)
	assert.NoError(t, err)
	assert.Len(t, tools, 1)
	assert.Equal(t, "tool1", tools[0].GetName())
}
