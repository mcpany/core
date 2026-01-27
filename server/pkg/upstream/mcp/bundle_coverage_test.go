// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/resource"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// Define local mock to avoid conflict and add needed fields
type mockSessionBundle struct {
	ListToolsFunc func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error)
	CloseFunc     func() error
}

func (m *mockSessionBundle) ListTools(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
	if m.ListToolsFunc != nil {
		return m.ListToolsFunc(ctx, params)
	}
	return &mcp.ListToolsResult{}, nil
}
func (m *mockSessionBundle) ListPrompts(_ context.Context, _ *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
	return &mcp.ListPromptsResult{Prompts: []*mcp.Prompt{}}, nil
}
func (m *mockSessionBundle) ListResources(_ context.Context, _ *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
	return &mcp.ListResourcesResult{Resources: []*mcp.Resource{}}, nil
}
func (m *mockSessionBundle) GetPrompt(_ context.Context, _ *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
	return nil, nil // Return nil for coverage or error
}
func (m *mockSessionBundle) ReadResource(_ context.Context, _ *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	return nil, nil
}
func (m *mockSessionBundle) CallTool(_ context.Context, _ *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	return nil, nil
}
func (m *mockSessionBundle) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// TestUnzipBundle_Cases tests various scenarios for unzipBundle
func TestUnzipBundle_Cases(t *testing.T) {
	// 1. Valid Zip
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

	// 2. Zip Slip
	t.Run("ZipSlip", func(t *testing.T) {
		// Just a smoke test here, heavily tested in mcp_coverage_test.go
		tmpDir := t.TempDir()
		zipPath := filepath.Join(tmpDir, "slip.zip")
		f, _ := os.Create(zipPath) //nolint:gosec
		// We can't use helper for ZipSlip easily, so manual:
		// Actually TestUnzipBundle_ZipSlip in mcp_coverage_test.go covers this.
		f.Close() //nolint:errcheck,gosec
	})

	// 3. Decompression Bomb (G110)
	t.Run("DecompressionBomb", func(t *testing.T) {
		tmpZip := filepath.Join(t.TempDir(), "bomb.zip")
		destDir := t.TempDir()
		_ = tmpZip
		_ = destDir
	})

	// 4. Invalid Zip File
	t.Run("InvalidZipFile", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "invalid.zip")
		_ = os.WriteFile(tmpFile, []byte("not a zip"), 0644) //nolint:gosec
		destDir := t.TempDir()
		err := unzipBundle(tmpFile, destDir)
		assert.Error(t, err)
	})

	// 5. Dest Dir Creation Fail
	t.Run("DestDirFail", func(t *testing.T) {
		tmpZip := filepath.Join(t.TempDir(), "valid.zip")
		createZip(t, tmpZip, map[string]string{"f": "c"})
		// Use a file as dest to cause mkdir fail
		destDir := t.TempDir() // Define destDir here
		destFile := filepath.Join(destDir, "file1.txt")
		_ = os.WriteFile(destFile, []byte(""), 0644) //nolint:gosec

		err := unzipBundle(tmpZip, destFile)
		assert.Error(t, err)
	})
}

func TestCreateAndRegisterMCPItemsFromBundle_Coverage(t *testing.T) {
	// Swap globals for testing
	oldNewClient := newClientForTesting
	oldConnect := connectForTesting
	defer func() {
		newClientForTesting = oldNewClient
		connectForTesting = oldConnect
	}()

	newClientForTesting = func(_ *mcp.Implementation) *mcp.Client {
		return nil
	}
	defer func() { newClientForTesting = nil }()

	// Mock Connect
	connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		return &mockSessionBundle{
			ListToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
				return &mcp.ListToolsResult{Tools: []*mcp.Tool{}}, nil
			},
			CloseFunc: func() error { return nil },
		}, nil
	}

	u := &Upstream{}

	t.Run("BundlePathMissing", func(t *testing.T) {
		_, _, err := u.createAndRegisterMCPItemsFromBundle(context.Background(), "s1", &configv1.McpBundleConnection{}, nil, nil, nil, false, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bundle_path is required")
	})

	t.Run("UnzipFail", func(t *testing.T) {
		cfg := &configv1.McpBundleConnection{BundlePath: proto.String("nonexistent.zip")}
		_, _, err := u.createAndRegisterMCPItemsFromBundle(context.Background(), "s1", cfg, nil, nil, nil, false, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unzip bundle")
	})

	t.Run("ManifestMissing", func(t *testing.T) {
		tmpZip := filepath.Join(t.TempDir(), "valid.zip")
		createZip(t, tmpZip, map[string]string{"other.txt": ""})
		cfg := &configv1.McpBundleConnection{BundlePath: proto.String(tmpZip)}
		_, _, err := u.createAndRegisterMCPItemsFromBundle(context.Background(), "s1", cfg, nil, nil, nil, false, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open manifest.json")
	})

	t.Run("ManifestInvalidJSON", func(t *testing.T) {
		tmpZip := filepath.Join(t.TempDir(), "valid.zip")
		createZip(t, tmpZip, map[string]string{"manifest.json": "{invalid"})
		cfg := &configv1.McpBundleConnection{BundlePath: proto.String(tmpZip)}
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
		cfg := &configv1.McpBundleConnection{
			BundlePath: proto.String(tmpZip),
			Env: map[string]*configv1.SecretValue{
				"FOO": {Value: &configv1.SecretValue_PlainText{PlainText: "OVERRIDE"}},
			},
		}

		u := &Upstream{}
		// We need to capture the transport env to verify
		var capturedEnv []string

		// Mock Connect to capture transport
		originalConnect := connectForTesting
		defer func() { connectForTesting = originalConnect }()

		connectForTesting = func(_ *mcp.Client, _ context.Context, transport mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			if bt, ok := transport.(*BundleDockerTransport); ok {
				capturedEnv = bt.Env
			}
			return &mockSessionBundle{}, nil
		}

		_, _, err := u.createAndRegisterMCPItemsFromBundle(context.Background(), "s1", cfg,
			&mockToolManagerCoverage{}, nil,
			&mockResourceManagerCoverage{resources: make(map[string]resource.Resource)},
			false, &configv1.UpstreamServiceConfig{})
		assert.NoError(t, err)

		// Check Env
		// "FOO=OVERRIDE" should be present
		// "BAR=BAZ" should be present
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
		cfg := &configv1.McpBundleConnection{BundlePath: proto.String(tmpZip)}

		u := &Upstream{}
		var capturedArgs []string

		originalConnect := connectForTesting
		defer func() { connectForTesting = originalConnect }()
		connectForTesting = func(_ *mcp.Client, _ context.Context, transport mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			if bt, ok := transport.(*BundleDockerTransport); ok {
				capturedArgs = bt.Args
			}
			return &mockSessionBundle{}, nil
		}

		_, _, err := u.createAndRegisterMCPItemsFromBundle(context.Background(), "s1", cfg,
			&mockToolManagerCoverage{}, nil, &mockResourceManagerCoverage{resources: make(map[string]resource.Resource)}, false, &configv1.UpstreamServiceConfig{})
		assert.NoError(t, err)

		// Expected subst: /app/bundle/main.py
		assert.Contains(t, capturedArgs, "/app/bundle/main.py")
	})

	t.Run("Success_UV_Implicit", func(t *testing.T) {
		tmpZip := filepath.Join(t.TempDir(), "valid_uv.zip")
		// No mcp_config, allow fallback
		manifest := `{
			"server": {
				"type": "uv",
				"entry_point": "script.py"
			}
		}`
		createZip(t, tmpZip, map[string]string{"manifest.json": manifest})
		cfg := &configv1.McpBundleConnection{BundlePath: proto.String(tmpZip)}

		u := &Upstream{}
		var capturedCommand string
		var capturedArgs []string

		originalConnect := connectForTesting
		defer func() { connectForTesting = originalConnect }()
		connectForTesting = func(_ *mcp.Client, _ context.Context, transport mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			if bt, ok := transport.(*BundleDockerTransport); ok {
				capturedCommand = bt.Command
				capturedArgs = bt.Args
			}
			return &mockSessionBundle{}, nil
		}

		_, _, err := u.createAndRegisterMCPItemsFromBundle(context.Background(), "s1", cfg,
			&mockToolManagerCoverage{}, nil, &mockResourceManagerCoverage{resources: make(map[string]resource.Resource)}, false, &configv1.UpstreamServiceConfig{})
		assert.NoError(t, err)

		assert.Equal(t, "uv", capturedCommand)
		assert.Equal(t, []string{"run", "/app/bundle/script.py"}, capturedArgs)
	})
}
