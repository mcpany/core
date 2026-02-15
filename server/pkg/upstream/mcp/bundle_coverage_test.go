package mcp

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/resource"
	mcp_sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// Define local mock to avoid conflict and add needed fields
type mockSessionBundle struct {
	ListToolsFunc func(ctx context.Context, params *mcp_sdk.ListToolsParams) (*mcp_sdk.ListToolsResult, error)
	CloseFunc     func() error
}

func (m *mockSessionBundle) ListTools(ctx context.Context, params *mcp_sdk.ListToolsParams) (*mcp_sdk.ListToolsResult, error) {
	if m.ListToolsFunc != nil {
		return m.ListToolsFunc(ctx, params)
	}
	return &mcp_sdk.ListToolsResult{}, nil
}
func (m *mockSessionBundle) ListPrompts(_ context.Context, _ *mcp_sdk.ListPromptsParams) (*mcp_sdk.ListPromptsResult, error) {
	return &mcp_sdk.ListPromptsResult{Prompts: []*mcp_sdk.Prompt{}}, nil
}
func (m *mockSessionBundle) ListResources(_ context.Context, _ *mcp_sdk.ListResourcesParams) (*mcp_sdk.ListResourcesResult, error) {
	return &mcp_sdk.ListResourcesResult{Resources: []*mcp_sdk.Resource{}}, nil
}
func (m *mockSessionBundle) GetPrompt(_ context.Context, _ *mcp_sdk.GetPromptParams) (*mcp_sdk.GetPromptResult, error) {
	return nil, nil
}
func (m *mockSessionBundle) ReadResource(_ context.Context, _ *mcp_sdk.ReadResourceParams) (*mcp_sdk.ReadResourceResult, error) {
	return nil, nil
}
func (m *mockSessionBundle) CallTool(_ context.Context, _ *mcp_sdk.CallToolParams) (*mcp_sdk.CallToolResult, error) {
	return nil, nil
}
func (m *mockSessionBundle) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func TestUnzipBundle_Cases(t *testing.T) {
	t.Run("ValidZip", func(t *testing.T) {
		tmpZip := filepath.Join(t.TempDir(), "valid.zip")
		destDir := t.TempDir()
		createZip(t, tmpZip, map[string]string{
			"manifest.json":   "{}",
			"subdir/file.txt": "content",
		})

		err := unzipBundle(tmpZip, destDir)
		assert.NoError(t, err)
		assert.FileExists(t, filepath.Join(destDir, "manifest.json"))
		assert.FileExists(t, filepath.Join(destDir, "subdir/file.txt"))
	})

	t.Run("ZipSlip", func(t *testing.T) {
		tmpDir := t.TempDir()
		zipPath := filepath.Join(tmpDir, "slip.zip")
		destDir := filepath.Join(tmpDir, "dest")
		err := os.Mkdir(destDir, 0755)
		require.NoError(t, err)

		f, err := os.Create(zipPath)
		require.NoError(t, err)
		w := zip.NewWriter(f)

		header := &zip.FileHeader{
			Name:   "../evil.txt",
			Method: zip.Store,
		}
		writer, err := w.CreateHeader(header)
		require.NoError(t, err)
		_, err = writer.Write([]byte("evil content"))
		require.NoError(t, err)

		require.NoError(t, w.Close())
		require.NoError(t, f.Close())

		err = unzipBundle(zipPath, destDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "illegal file path")
		assert.NoFileExists(t, filepath.Join(tmpDir, "evil.txt"))
	})

	t.Run("DecompressionBomb", func(t *testing.T) {
		if os.Getenv("SKIP_EXPENSIVE_TESTS") == "true" {
			t.Skip("Skipping expensive decompression bomb test")
		}
		// Create a zip bomb (1.1 GB of zeros)
		tmpDir := t.TempDir()
		zipPath := filepath.Join(tmpDir, "bomb.zip")
		destDir := filepath.Join(tmpDir, "dest")
		err := os.Mkdir(destDir, 0755)
		require.NoError(t, err)

		f, err := os.Create(zipPath)
		require.NoError(t, err)
		w := zip.NewWriter(f)

		writer, err := w.Create("bomb.txt")
		require.NoError(t, err)

		chunkSize := 10 * 1024 * 1024 // 10MB
		chunk := make([]byte, chunkSize)
		iterations := (1024 / 10) + 1 // 103 iterations * 10MB = 1030MB > 1024MB

		for i := 0; i < iterations; i++ {
			_, err := writer.Write(chunk)
			require.NoError(t, err)
		}

		require.NoError(t, w.Close())
		require.NoError(t, f.Close())

		err = unzipBundle(zipPath, destDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds maximum allowed size")
	})

	t.Run("InvalidZipFile", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "invalid.zip")
		_ = os.WriteFile(tmpFile, []byte("not a zip"), 0644)
		destDir := t.TempDir()
		err := unzipBundle(tmpFile, destDir)
		assert.Error(t, err)
	})

	t.Run("DestDirFail", func(t *testing.T) {
		tmpZip := filepath.Join(t.TempDir(), "valid.zip")
		createZip(t, tmpZip, map[string]string{"f": "c"})
		destDir := t.TempDir()
		destFile := filepath.Join(destDir, "file1.txt")
		_ = os.WriteFile(destFile, []byte(""), 0644)

		err := unzipBundle(tmpZip, destFile)
		assert.Error(t, err)
	})
}

func TestCreateAndRegisterMCPItemsFromBundle_Coverage(t *testing.T) {
	oldNewClient := newClientForTesting
	oldConnect := connectForTesting
	defer func() {
		newClientForTesting = oldNewClient
		connectForTesting = oldConnect
	}()

	newClientForTesting = func(_ *mcp_sdk.Implementation) *mcp_sdk.Client {
		return nil
	}

	connectForTesting = func(_ *mcp_sdk.Client, _ context.Context, _ mcp_sdk.Transport, _ []mcp_sdk.Root) (ClientSession, error) {
		return &mockSessionBundle{
			ListToolsFunc: func(_ context.Context, _ *mcp_sdk.ListToolsParams) (*mcp_sdk.ListToolsResult, error) {
				return &mcp_sdk.ListToolsResult{Tools: []*mcp_sdk.Tool{}}, nil
			},
			CloseFunc: func() error { return nil },
		}, nil
	}

	u := &Upstream{}

	t.Run("BundlePathMissing", func(t *testing.T) {
		_, _, err := u.createAndRegisterMCPItemsFromBundle(context.Background(), "s1", configv1.McpBundleConnection_builder{}.Build(), nil, nil, nil, false, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bundle_path is required")
	})

	t.Run("UnzipFail", func(t *testing.T) {
		cfg := configv1.McpBundleConnection_builder{
			BundlePath: proto.String("nonexistent.zip"),
		}.Build()
		_, _, err := u.createAndRegisterMCPItemsFromBundle(context.Background(), "s1", cfg, nil, nil, nil, false, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unzip bundle")
	})

	t.Run("ManifestMissing", func(t *testing.T) {
		tmpZip := filepath.Join(t.TempDir(), "valid.zip")
		createZip(t, tmpZip, map[string]string{"other.txt": ""})
		cfg := configv1.McpBundleConnection_builder{
			BundlePath: proto.String(tmpZip),
		}.Build()
		_, _, err := u.createAndRegisterMCPItemsFromBundle(context.Background(), "s1", cfg, nil, nil, nil, false, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open manifest.json")
	})

	t.Run("ManifestInvalidJSON", func(t *testing.T) {
		tmpZip := filepath.Join(t.TempDir(), "valid.zip")
		createZip(t, tmpZip, map[string]string{"manifest.json": "{invalid"})
		cfg := configv1.McpBundleConnection_builder{
			BundlePath: proto.String(tmpZip),
		}.Build()
		_, _, err := u.createAndRegisterMCPItemsFromBundle(context.Background(), "s1", cfg, nil, nil, nil, false, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode manifest.json")
	})

	t.Run("Success_Node_EnvOverride", func(t *testing.T) {
		tmpZip := filepath.Join(t.TempDir(), "valid_env.zip")
		manifest := `{
			"server": {
				"type": "node",
				"mcp_config": {
					"env": {"FOO": "ORIGINAL", "BAR": "BAZ"}
				}
			}
		}`
		createZip(t, tmpZip, map[string]string{"manifest.json": manifest})
		cfg := configv1.McpBundleConnection_builder{
			BundlePath: proto.String(tmpZip),
			Env: map[string]*configv1.SecretValue{
				"FOO": configv1.SecretValue_builder{PlainText: proto.String("OVERRIDE")}.Build(),
			},
		}.Build()

		u := &Upstream{}
		var capturedEnv []string

		originalConnect := connectForTesting
		defer func() { connectForTesting = originalConnect }()

		connectForTesting = func(_ *mcp_sdk.Client, _ context.Context, transport mcp_sdk.Transport, _ []mcp_sdk.Root) (ClientSession, error) {
			if bt, ok := transport.(*BundleDockerTransport); ok {
				capturedEnv = bt.Env
			}
			return &mockSessionBundle{}, nil
		}

		_, _, err := u.createAndRegisterMCPItemsFromBundle(context.Background(), "s1", cfg,
			&mockToolManagerCoverage{}, nil,
			&mockResourceManagerCoverage{resources: make(map[string]resource.Resource)},
			false, configv1.UpstreamServiceConfig_builder{}.Build())
		assert.NoError(t, err)

		foundFoo := false
		foundBar := false
		for _, e := range capturedEnv {
			if e == "FOO=OVERRIDE" {
				foundFoo = true
			}
			if e == "BAR=BAZ" {
				foundBar = true
			}
		}
		assert.True(t, foundFoo, "FOO should be overridden")
		assert.True(t, foundBar, "BAR should be present")
	})

	t.Run("Success_Python_Substitution", func(t *testing.T) {
		tmpZip := filepath.Join(t.TempDir(), "valid_py.zip")
		manifest := `{
			"server": {
				"type": "python",
				"mcp_config": {
					"args": ["${__dirname}/main.py"]
				}
			}
		}`
		createZip(t, tmpZip, map[string]string{"manifest.json": manifest})
		cfg := configv1.McpBundleConnection_builder{
			BundlePath: proto.String(tmpZip),
		}.Build()

		u := &Upstream{}
		var capturedArgs []string

		originalConnect := connectForTesting
		defer func() { connectForTesting = originalConnect }()
		connectForTesting = func(_ *mcp_sdk.Client, _ context.Context, transport mcp_sdk.Transport, _ []mcp_sdk.Root) (ClientSession, error) {
			if bt, ok := transport.(*BundleDockerTransport); ok {
				capturedArgs = bt.Args
			}
			return &mockSessionBundle{}, nil
		}

		_, _, err := u.createAndRegisterMCPItemsFromBundle(context.Background(), "s1", cfg,
			&mockToolManagerCoverage{}, nil, &mockResourceManagerCoverage{resources: make(map[string]resource.Resource)}, false, configv1.UpstreamServiceConfig_builder{}.Build())
		assert.NoError(t, err)

		assert.Contains(t, capturedArgs, "/app/bundle/main.py")
	})

	t.Run("Success_UV_Implicit", func(t *testing.T) {
		tmpZip := filepath.Join(t.TempDir(), "valid_uv.zip")
		manifest := `{
			"server": {
				"type": "uv",
				"entry_point": "script.py"
			}
		}`
		createZip(t, tmpZip, map[string]string{"manifest.json": manifest})
		cfg := configv1.McpBundleConnection_builder{
			BundlePath: proto.String(tmpZip),
		}.Build()

		u := &Upstream{}
		var capturedCommand string
		var capturedArgs []string

		originalConnect := connectForTesting
		defer func() { connectForTesting = originalConnect }()
		connectForTesting = func(_ *mcp_sdk.Client, _ context.Context, transport mcp_sdk.Transport, _ []mcp_sdk.Root) (ClientSession, error) {
			if bt, ok := transport.(*BundleDockerTransport); ok {
				capturedCommand = bt.Command
				capturedArgs = bt.Args
			}
			return &mockSessionBundle{}, nil
		}

		_, _, err := u.createAndRegisterMCPItemsFromBundle(context.Background(), "s1", cfg,
			&mockToolManagerCoverage{}, nil, &mockResourceManagerCoverage{resources: make(map[string]resource.Resource)}, false, configv1.UpstreamServiceConfig_builder{}.Build())
		assert.NoError(t, err)

		assert.Equal(t, "uv", capturedCommand)
		assert.Equal(t, []string{"run", "/app/bundle/script.py"}, capturedArgs)
	})
}
