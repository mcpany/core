package mcp

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestUpstream_Shutdown(t *testing.T) {
	u := NewUpstream(nil)
	err := u.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestTransportError(t *testing.T) {
	e := &transportError{
		Code:    123,
		Message: "test error",
		Data:    "some data",
	}
	assert.Equal(t, "test error", e.Error())
}

func TestInferImage(t *testing.T) {
	tests := []struct {
		serverType string
		expected   string
	}{
		{"node", "node:18-alpine"},
		{"python", "python:3.11-slim"},
		{"uv", "ghcr.io/astral-sh/uv:python3.11-bookworm-slim"},
		{"binary", "debian:bookworm-slim"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.serverType, func(t *testing.T) {
			got := inferImage(tt.serverType)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestFixID is already in bundle_transport_test.go

type mockRWC struct {
	written []byte
}

func (m *mockRWC) Write(p []byte) (n int, err error) {
	m.written = append(m.written, p...)
	return len(p), nil
}

func (m *mockRWC) Read(_ []byte) (n int, err error) {
	return 0, io.EOF
}

func (m *mockRWC) Close() error {
	return nil
}

func TestBundleDockerConn_Write(t *testing.T) {
	rwc := &mockRWC{}
	conn := &bundleDockerConn{
		encoder: json.NewEncoder(rwc),
		log:     logging.GetLogger(),
		rwc:     rwc,
	}

	// Test Request
	// Avoid Unmarshal error by creating struct manually and setting ID via helper
	req := &jsonrpc.Request{
		Method: "method",
		Params: json.RawMessage(`{"foo":"bar"}`),
	}
	setUnexportedID(&req.ID, "123")

	err := conn.Write(context.Background(), req)
	assert.NoError(t, err)
	// Verify output - ID is parsed as int by fixID because it looks like a number
	assert.Contains(t, string(rwc.written), `"method":"method"`)
	assert.Contains(t, string(rwc.written), `"id":123`)

	// Reset
	rwc.written = nil

	// Test Response
	resp := &jsonrpc.Response{
		Result: json.RawMessage(`{"result":"ok"}`),
	}
	// Use string "456" to align with Request success
	setUnexportedID(&resp.ID, "456")

	err = conn.Write(context.Background(), resp)
	assert.NoError(t, err)
	assert.Contains(t, string(rwc.written), `"result":{"result":"ok"}`)
	assert.Contains(t, string(rwc.written), `"id":456`)
}

func TestUnzipBundle(t *testing.T) {
	// Create a zip file
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")
	zipFile, err := os.Create(zipPath) //nolint:gosec
	assert.NoError(t, err)

	w := zip.NewWriter(zipFile)

	// File 1
	f, err := w.Create("file1.txt")
	assert.NoError(t, err)
	_, err = f.Write([]byte("content1"))
	assert.NoError(t, err)

	// Dir/File 2
	f, err = w.Create("dir/file2.txt")
	assert.NoError(t, err)
	_, err = f.Write([]byte("content2"))
	assert.NoError(t, err)

	assert.NoError(t, w.Close())
	assert.NoError(t, zipFile.Close())

	// Unzip
	destDir := filepath.Join(tmpDir, "extracted")
	err = unzipBundle(zipPath, destDir)
	assert.NoError(t, err)

	// Verify
	c1, err := os.ReadFile(filepath.Join(destDir, "file1.txt")) //nolint:gosec
	assert.NoError(t, err)
	assert.Equal(t, "content1", string(c1))

	c2, err := os.ReadFile(filepath.Join(destDir, "dir", "file2.txt")) //nolint:gosec
	assert.NoError(t, err)
	assert.Equal(t, "content2", string(c2))
}

func TestBundleDockerConn_Read_Error(t *testing.T) {
	rwc := &mockRWC{}
	conn := &bundleDockerConn{
		decoder: json.NewDecoder(rwc), // Empty reader, will return EOF
		log:     logging.GetLogger(),
		rwc:     rwc,
	}

	_, err := conn.Read(context.Background())
	assert.Error(t, err) // EOF
}

func TestAuthenticatedRoundTripper_Coverage(t *testing.T) {
	// Test AuthenticatedRoundTripper
	mockRT := &mockRTCoverage{
		resp: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("")),
		},
	}
	art := &authenticatedRoundTripper{
		base: mockRT,
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, err := art.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	if err == nil {
		defer resp.Body.Close() //nolint:errcheck
	}

	// Test with authenticator
	art.authenticator = &mockAuthCoverage{}
	resp, err = art.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, "true", req.Header.Get("Authenticated"))
	resp.Body.Close() //nolint:errcheck,gosec

	// Test authenticator error
	art.authenticator = &mockAuthCoverage{fail: true}
	resp, err = art.RoundTrip(req)
	if err == nil {
		resp.Body.Close() //nolint:errcheck,gosec
	}
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to authenticate")
}

func TestUnzipBundle_ZipSlip(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "slip.zip")
	zipFile, err := os.Create(zipPath) //nolint:gosec
	assert.NoError(t, err)

	w := zip.NewWriter(zipFile)
	// Create a file with ".." in name
	f, err := w.Create("../evil.txt")
	assert.NoError(t, err)
	_, err = f.Write([]byte("evil"))
	assert.NoError(t, err)

	assert.NoError(t, w.Close())
	assert.NoError(t, zipFile.Close())

	destDir := filepath.Join(tmpDir, "extracted")
	err = unzipBundle(zipPath, destDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "illegal file path")
}

func TestUpstream_Register_Bundle_Error(t *testing.T) {
	u := NewUpstream(nil)
	ctx := context.Background()
	tm := &mockToolManagerCoverage{}
	pm := &mockPromptManagerCoverage{}
	rm := &mockResourceManagerCoverage{}

	// Case 1: Empty Bundle Path
	configEmpty := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-bundle-empty"),
		McpService: configv1.McpUpstreamService_builder{
			BundleConnection: configv1.McpBundleConnection_builder{
				BundlePath: proto.String(""),
			}.Build(),
		}.Build(),
	}.Build()

	_, _, _, err := u.Register(ctx, configEmpty, tm, pm, rm, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bundle_path is required")

	// Case 2: Unzip Fail (File not found)
	configMissing := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-bundle-missing"),
		McpService: configv1.McpUpstreamService_builder{
			BundleConnection: configv1.McpBundleConnection_builder{
				BundlePath: proto.String("/non/existent/path.zip"),
			}.Build(),
		}.Build(),
	}.Build()

	_, _, _, err = u.Register(ctx, configMissing, tm, pm, rm, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unzip bundle")

	// Case 3: Manifest Missing
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "no_manifest.zip")
	createZip(t, zipPath, map[string]string{"foo.txt": "bar"})

	configNoManifest := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-bundle-no-manifest"),
		McpService: configv1.McpUpstreamService_builder{
			BundleConnection: configv1.McpBundleConnection_builder{
				BundlePath: proto.String(zipPath),
			}.Build(),
		}.Build(),
	}.Build()

	_, _, _, err = u.Register(ctx, configNoManifest, tm, pm, rm, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open manifest.json")

	// Case 4: Manifest Invalid JSON
	zipPathInvalid := filepath.Join(tmpDir, "invalid_manifest.zip")
	createZip(t, zipPathInvalid, map[string]string{"manifest.json": "{invalid"})

	configInvalidManifest := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-bundle-invalid-manifest"),
		McpService: configv1.McpUpstreamService_builder{
			BundleConnection: configv1.McpBundleConnection_builder{
				BundlePath: proto.String(zipPathInvalid),
			}.Build(),
		}.Build(),
	}.Build()

	_, _, _, err = u.Register(ctx, configInvalidManifest, tm, pm, rm, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode manifest.json")
}

func TestUpstream_Register_Bundle_Success(t *testing.T) {
	u := NewUpstream(nil)
	ctx := context.Background()
	tm := &mockToolManagerCoverage{}
	pm := &mockPromptManagerCoverage{}
	rm := &mockResourceManagerCoverage{}

	// Setup valid bundle
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "valid_bundle.zip")
	manifest := `{
		"manifest_version": "1.0",
		"name": "test-bundle",
		"version": "1.0.0",
		"description": "Test Bundle",
		"server": {
			"type": "node",
			"entry_point": "index.js",
			"mcp_config": {
				"command": "node",
				"args": ["index.js"],
				"env": {"FOO": "BAR"}
			}
		},
		"tools": [
			{"name": "tool1", "description": "d1", "inputSchema": {"type": "object"}}
		],
		"resources": [
			{"uri": "resource://r1", "name": "r1", "description": "d1"}
		],
		"prompts": [
			{"name": "prompt1", "description": "d1"}
		]
	}`
	createZip(t, zipPath, map[string]string{"manifest.json": manifest})

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-bundle-success"),
		McpService: configv1.McpUpstreamService_builder{
			BundleConnection: configv1.McpBundleConnection_builder{
				BundlePath: proto.String(zipPath),
			}.Build(),
		}.Build(),
	}.Build()

	// Mock Connect logic
	originalConnect := connectForTesting
	defer func() { connectForTesting = originalConnect }()

	mockCS := &mockSessionCoverage{
		tools: &mcp.ListToolsResult{
			Tools: []*mcp.Tool{
				{Name: "tool1", InputSchema: map[string]interface{}{"type": "object"}},
			},
		},
	}
	SetConnectForTesting(func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		return mockCS, nil
	})

	_, _, _, err := u.Register(ctx, config, tm, pm, rm, false)
	assert.NoError(t, err)
}

func TestUpstream_Register_Bundle_Variants(t *testing.T) {
	u := NewUpstream(nil)
	ctx := context.Background()

	variants := []struct {
		name     string
		manifest string
	}{
		{
			name: "python",
			manifest: `{
				"manifest_version": "1.0",
				"name": "python-bundle",
				"version": "1.0.0",
				"description": "d",
				"server": {
					"type": "python",
					"entry_point": "main.py"
				}
			}`,
		},
		{
			name: "uv",
			manifest: `{
				"manifest_version": "1.0",
				"name": "uv-bundle",
				"version": "1.0",
				"description": "d",
				"server": { "type": "uv", "entry_point": "script.py" }
			}`,
		},
	}

	for _, v := range variants {
		t.Run(v.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			zipPath := filepath.Join(tmpDir, "bundle.zip")
			createZip(t, zipPath, map[string]string{"manifest.json": v.manifest})

			config := configv1.UpstreamServiceConfig_builder{
				Name: proto.String("test-" + v.name),
				McpService: configv1.McpUpstreamService_builder{
					BundleConnection: configv1.McpBundleConnection_builder{
						BundlePath: proto.String(zipPath),
					}.Build(),
				}.Build(),
			}.Build()

			// Mock Connect
			originalConnect := connectForTesting
			defer func() { connectForTesting = originalConnect }()
			SetConnectForTesting(func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
				return &mockSessionCoverage{
					tools: &mcp.ListToolsResult{},
				}, nil
			})

			_, _, _, err := u.Register(ctx, config, &mockToolManagerCoverage{}, &mockPromptManagerCoverage{}, &mockResourceManagerCoverage{}, false)
			assert.NoError(t, err)
		})
	}
}

func TestUpstream_Register_Bundle_RealClient(t *testing.T) {
	// Test the path where mcp.NewClient is called (not mocked)
	u := NewUpstream(nil)
	ctx := context.Background()

	// Reset any global mocks if they leak (though they shouldn't if I use Set correctly)
	// But u.Register uses global variables `newClientForTesting`.
	// I need to ensure `newClientForTesting` is nil.
	// But `SetNewClientForTesting` sets the global.
	// I need to unset it.
	originalNewClient := newClientForTesting
	newClientForTesting = nil // Force nil
	defer func() { newClientForTesting = originalNewClient }()

	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "bundle_real.zip")
	manifest := `{
		"manifest_version": "1.0",
		"name": "real-bundle",
		"version": "1.0",
		"server": { "type": "node", "mcp_config": { "command": "node" } }
	}`
	createZip(t, zipPath, map[string]string{"manifest.json": manifest})

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-real"),
		McpService: configv1.McpUpstreamService_builder{
			BundleConnection: configv1.McpBundleConnection_builder{
				BundlePath: proto.String(zipPath),
			}.Build(),
		}.Build(),
	}.Build()

	// Mock Connect ONLY. Real Client will be passed.
	originalConnect := connectForTesting
	defer func() { connectForTesting = originalConnect }()
	SetConnectForTesting(func(client *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		assert.NotNil(t, client)
		return &mockSessionCoverage{
			tools: &mcp.ListToolsResult{},
		}, nil
	})

	_, _, _, err := u.Register(ctx, config, &mockToolManagerCoverage{}, &mockPromptManagerCoverage{}, &mockResourceManagerCoverage{}, false)
	assert.NoError(t, err)
}

func TestUpstream_Register_Bundle_UnknownType(t *testing.T) {
	u := NewUpstream(nil)
	ctx := context.Background()

	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "bundle_unknown.zip")
	manifest := `{
		"manifest_version": "1.0", "name": "unknown", "version": "1.0",
		"server": { "type": "unknown_xyz" }
	}`
	createZip(t, zipPath, map[string]string{"manifest.json": manifest})

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-unknown"),
		McpService: configv1.McpUpstreamService_builder{
			BundleConnection: configv1.McpBundleConnection_builder{
				BundlePath: proto.String(zipPath),
			}.Build(),
		}.Build(),
	}.Build()

	_, _, _, err := u.Register(ctx, config, &mockToolManagerCoverage{}, &mockPromptManagerCoverage{}, &mockResourceManagerCoverage{}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to infer container image")
}

func TestDockerTransport_Connect_Errors(t *testing.T) {
	originalNewDockerClient := newDockerClient
	defer func() { newDockerClient = originalNewDockerClient }()

	t.Run("CreateError", func(t *testing.T) {
		newDockerClient = func(_ ...client.Opt) (dockerClient, error) {
			return &mockDockerClient{
				ImagePullFunc: func(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("")), nil
				},
				ContainerCreateFunc: func(_ context.Context, _ *container.Config, _ *container.HostConfig, _ *network.NetworkingConfig, _ *v1.Platform, _ string) (container.CreateResponse, error) {
					return container.CreateResponse{}, fmt.Errorf("create failed")
				},
			}, nil
		}
		transport := &BundleDockerTransport{Image: "test", Command: "test"}
		_, err := transport.Connect(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create container")
	})

	t.Run("AttachError", func(t *testing.T) {
		newDockerClient = func(_ ...client.Opt) (dockerClient, error) {
			return &mockDockerClient{
				ImagePullFunc: func(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("")), nil
				},
				ContainerCreateFunc: func(_ context.Context, _ *container.Config, _ *container.HostConfig, _ *network.NetworkingConfig, _ *v1.Platform, _ string) (container.CreateResponse, error) {
					return container.CreateResponse{ID: "id"}, nil
				},
				ContainerAttachFunc: func(_ context.Context, _ string, _ container.AttachOptions) (types.HijackedResponse, error) {
					return types.HijackedResponse{}, fmt.Errorf("attach failed")
				},
				ContainerRemoveFunc: func(_ context.Context, _ string, _ container.RemoveOptions) error { return nil },
			}, nil
		}
		transport := &BundleDockerTransport{Image: "test", Command: "test"}
		_, err := transport.Connect(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to attach")
	})

	t.Run("StartError", func(t *testing.T) {
		newDockerClient = func(_ ...client.Opt) (dockerClient, error) {
			return &mockDockerClient{
				ImagePullFunc: func(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("")), nil
				},
				ContainerCreateFunc: func(_ context.Context, _ *container.Config, _ *container.HostConfig, _ *network.NetworkingConfig, _ *v1.Platform, _ string) (container.CreateResponse, error) {
					return container.CreateResponse{ID: "id"}, nil
				},
				ContainerAttachFunc: func(_ context.Context, _ string, _ container.AttachOptions) (types.HijackedResponse, error) {
					c1, c2 := net.Pipe()
					_ = c2 // Ignore
					return types.HijackedResponse{Conn: c1, Reader: bufio.NewReader(c1)}, nil
				},
				ContainerStartFunc: func(_ context.Context, _ string, _ container.StartOptions) error {
					return fmt.Errorf("start failed")
				},
				ContainerRemoveFunc: func(_ context.Context, _ string, _ container.RemoveOptions) error { return nil },
			}, nil
		}
		transport := &BundleDockerTransport{Image: "test", Command: "test"}
		_, err := transport.Connect(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start")
	})

	t.Run("ImagePullError", func(t *testing.T) {
		// Image pull error should be logged but continue (warn)
		newDockerClient = func(_ ...client.Opt) (dockerClient, error) {
			return &mockDockerClient{
				ImagePullFunc: func(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
					return nil, fmt.Errorf("pull failed")
				},
				ContainerCreateFunc: func(_ context.Context, _ *container.Config, _ *container.HostConfig, _ *network.NetworkingConfig, _ *v1.Platform, _ string) (container.CreateResponse, error) {
					return container.CreateResponse{ID: "id"}, nil
				},
				ContainerAttachFunc: func(_ context.Context, _ string, _ container.AttachOptions) (types.HijackedResponse, error) {
					c1, c2 := net.Pipe()
					_ = c2
					return types.HijackedResponse{Conn: c1, Reader: bufio.NewReader(c1)}, nil
				},
				ContainerStartFunc: func(_ context.Context, _ string, _ container.StartOptions) error { return nil },
			}, nil
		}
		transport := &BundleDockerTransport{Image: "test", Command: "test"}
		conn, err := transport.Connect(context.Background())
		assert.NoError(t, err)
		conn.Close() //nolint:errcheck,gosec
	})
}

func TestRegister_DynamicResource_EdgeCases(t *testing.T) {
	u := NewUpstream(nil)
	ctx := context.Background()
	tm := &mockToolManagerCoverage{}
	pm := &mockPromptManagerCoverage{}
	rm := &mockResourceManagerCoverage{}

	// Case: Dynamic Resource referencing non-existent tool call
	// Config has resource but NO matching tool call ID in any tool
	resDef := configv1.ResourceDefinition_builder{
		Name: proto.String("orphan-dynamic"),
		Uri:  proto.String("test://orphan"),
		Dynamic: configv1.DynamicResource_builder{
			McpCall: configv1.MCPCallDefinition_builder{
				Id: proto.String("missing-call-id"),
			}.Build(),
		}.Build(),
	}.Build()

	mcpService := configv1.McpUpstreamService_builder{
		StdioConnection: configv1.McpStdioConnection_builder{Command: proto.String("echo")}.Build(),
		Resources:       []*configv1.ResourceDefinition{resDef},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:       proto.String("test-orphan"),
		McpService: mcpService,
	}.Build()

	originalConnect := connectForTesting
	defer func() { connectForTesting = originalConnect }()
	SetConnectForTesting(func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		return &mockSessionCoverage{}, nil
	})

	_, _, _, err := u.Register(ctx, config, tm, pm, rm, false)
	assert.NoError(t, err)
	// Verify resource NOT registered
	_, ok := rm.GetResource("test://orphan")
	assert.False(t, ok, "orphan resource should not be registered")

	// Case: Dynamic Resource with nil Call
	// Not easily constructable via builders if they enforce fields?
	// But if we have McpCall field in DynamicResource builder...
	// We can try to build one without McpCall?
	// Let's assume builder validates or we skip it.
}

func createZip(t *testing.T, path string, files map[string]string) {
	f, err := os.Create(path) //nolint:gosec
	assert.NoError(t, err)
	defer f.Close() //nolint:errcheck
	w := zip.NewWriter(f)
	defer w.Close() //nolint:errcheck

	for name, content := range files {
		wf, err := w.Create(name)
		assert.NoError(t, err)
		_, err = wf.Write([]byte(content))
		assert.NoError(t, err)
	}
}

type mockRTCoverage struct {
	resp *http.Response
	err  error
}

func (m *mockRTCoverage) RoundTrip(_ *http.Request) (*http.Response, error) {
	return m.resp, m.err
}

type mockAuthCoverage struct {
	fail bool
}

func (m *mockAuthCoverage) Authenticate(req *http.Request) error {
	if m.fail {
		return context.DeadlineExceeded // some error
	}
	req.Header.Set("Authenticated", "true")
	return nil
}

func TestStreamableHTTP_RoundTrip_Coverage(t *testing.T) {
	tr := &StreamableHTTP{
		Client: &http.Client{
			Transport: &mockRTCoverage{
				resp: &http.Response{StatusCode: 202},
			},
		},
	}
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, err := tr.RoundTrip(req)
	assert.NoError(t, err)
	if err == nil {
		defer resp.Body.Close() //nolint:errcheck
	}
	assert.Equal(t, 202, resp.StatusCode)

	// Test Default Client
	trDefault := &StreamableHTTP{}
	_ = trDefault // Use it
	reqInvalid, _ := http.NewRequest("GET", "http://invalid.local", nil)
	resp, err = trDefault.RoundTrip(reqInvalid) // Should not panic
	if err == nil {
		defer resp.Body.Close() //nolint:errcheck
	}
}

type mockToolManagerCoverage struct {
	tool.ManagerInterface
}

func (m *mockToolManagerCoverage) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}
func (m *mockToolManagerCoverage) AddTool(_ tool.Tool) error                    { return nil }
func (m *mockToolManagerCoverage) GetTool(_ string) (tool.Tool, bool)           { return nil, false }
func (m *mockToolManagerCoverage) ListServices() []*tool.ServiceInfo            { return nil }

type mockPromptManagerCoverage struct {
	prompt.ManagerInterface
}
type mockResourceManagerCoverage struct {
	resource.ManagerInterface
	resources map[string]resource.Resource
}

func (m *mockResourceManagerCoverage) AddResource(r resource.Resource) {
	if m.resources == nil {
		m.resources = make(map[string]resource.Resource)
	}
	m.resources[r.Resource().URI] = r
}

func (m *mockResourceManagerCoverage) GetResource(uri string) (resource.Resource, bool) {
	if m.resources == nil {
		return nil, false
	}
	r, ok := m.resources[uri]
	return r, ok
}

func TestUpstream_Register_Coverage(t *testing.T) {
	// Mocks
	tm := &mockToolManagerCoverage{}
	pm := &mockPromptManagerCoverage{}
	rm := &mockResourceManagerCoverage{}

	u := NewUpstream(nil)
	ctx := context.Background()

	// Case 1: Nil McpService -> Error "mcp service config is nil"
	// Builders might not allow nil McpService easily if they enforce validation?
	// But we can build empty config?
	svcConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()

	// If Builders ensure usage, we might not be able to produce invalid config easily.
	// But Upstream Register checks for nil.
	// Let's try to pass the config.
	_, _, _, err := u.Register(ctx, svcConfig, tm, pm, rm, false)
	assert.Error(t, err)
	// Register checks `if config.McpService == nil`.
	// The builder might leave it nil if not set.

	// Case 2: Stdio Connect Error
	svcConfigStdio := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-stdio"),
		McpService: configv1.McpUpstreamService_builder{
			StdioConnection: configv1.McpStdioConnection_builder{
				Command: proto.String("echo"),
			}.Build(),
		}.Build(),
	}.Build()

	// Hook connect
	originalConnect := connectForTesting
	defer func() { connectForTesting = originalConnect }()

	SetConnectForTesting(func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		return nil, fmt.Errorf("connect failed")
	})

	_, _, _, err = u.Register(ctx, svcConfigStdio, tm, pm, rm, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connect failed")

	// Case 3: Http Connect Error
	// Case 3: Http Connect Error
	svcConfigHTTP := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-http"),
		McpService: configv1.McpUpstreamService_builder{
			HttpConnection: configv1.McpStreamableHttpConnection_builder{
				HttpAddress: proto.String("http://127.0.0.1"),
			}.Build(),
		}.Build(),
	}

	// Register
	_, _, _, err = u.Register(context.Background(), svcConfigHTTP.Build(), tm, pm, rm, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connect failed")

	// Case 4: Bundle Connect Error
	// Bundle requires bundle path
	// We can point to a dummy zip or just ensure it tries to connect
	// But bundle connect involves unzip + docker connect.
	// If Unzip fails, it returns error early.
	// If unzip succeeds, it calls Connect.

	// Let's rely on Stdio and Http error paths for now as they share significant logic in processMCPItems too if we let them succeed connect but fail list tools.

	// Case 5: Connect Success, ListTools Fail
	SetConnectForTesting(func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		return &mockSessionCoverage{listToolsErr: fmt.Errorf("list tools failed")}, nil
	})

	_, _, _, err = u.Register(ctx, svcConfigStdio, tm, pm, rm, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list tools failed")
}

func TestBundleSlogWriter(t *testing.T) {
	logger := logging.GetLogger()
	w := &bundleSlogWriter{
		log:   logger,
		level: slog.LevelInfo,
	}
	n, err := w.Write([]byte("test log"))
	assert.NoError(t, err)
	assert.Equal(t, 8, n)
}

func TestBundleDockerConn_Read_Malformed(t *testing.T) {
	rwc := &mockRWC{}
	conn := &bundleDockerConn{
		decoder: json.NewDecoder(rwc),
		log:     logging.GetLogger(),
		rwc:     rwc,
	}

	// Case: Invalid JSON (Decode fails)
	conn.decoder = json.NewDecoder(io.MultiReader(strings.NewReader(`{invalid`), rwc))
	_, err := conn.Read(context.Background())
	assert.Error(t, err)

	// Case: Valid JSON but unexpected type (e.g. integer instead of object)
	conn.decoder = json.NewDecoder(strings.NewReader(`123`))
	_, err = conn.Read(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal message header")
}

type mockSessionCoverage struct {
	listToolsErr error
	tools        *mcp.ListToolsResult
}

func (m *mockSessionCoverage) ListTools(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
	if m.tools != nil {
		return m.tools, m.listToolsErr
	}
	if m.listToolsErr != nil {
		return nil, m.listToolsErr
	}
	return &mcp.ListToolsResult{}, nil
}
func (m *mockSessionCoverage) ListPrompts(_ context.Context, _ *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
	return &mcp.ListPromptsResult{}, nil
}
func (m *mockSessionCoverage) ListResources(_ context.Context, _ *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
	return &mcp.ListResourcesResult{}, nil
}
func (m *mockSessionCoverage) GetPrompt(_ context.Context, _ *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
	return nil, nil
}
func (m *mockSessionCoverage) ReadResource(_ context.Context, _ *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	return nil, nil
}
func (m *mockSessionCoverage) CallTool(_ context.Context, _ *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	return nil, nil
}
func (m *mockSessionCoverage) Close() error {
	return nil
}

func TestUnzipBundle_InvalidFile(t *testing.T) {
	// Create a dummy non-zip file
	tmpFile, err := os.CreateTemp("", "not-a-zip")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name()) //nolint:errcheck
	_, _ = tmpFile.WriteString("this is not a zip file")
	_ = tmpFile.Close()

	destDir, err := os.MkdirTemp("", "unzip-dest-fail")
	assert.NoError(t, err)
	defer os.RemoveAll(destDir) //nolint:errcheck

	err = unzipBundle(tmpFile.Name(), destDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "zip: not a valid zip file")
}
