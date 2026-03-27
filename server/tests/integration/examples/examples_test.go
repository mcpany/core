package examples

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestExampleConfigs(t *testing.T) {
	// Set dummy API key for validation to pass
	t.Setenv("GEMINI_API_KEY", "dummy-key")
	projectRoot, rootFetchErr := sourceProjectRoot()
	require.NoError(t, rootFetchErr)
	runtimeRoot := filepath.Join(t.TempDir(), "server")
	examplesDir := filepath.Join(runtimeRoot, "examples")
	require.NoError(t, copyDir(filepath.Join(projectRoot, "examples"), examplesDir))

	// Ensure stdio example binary is built, as Config validation checks for its existence
	// This makes the test robust against sharding/environment where build-examples might not have run.
	stdioBinPath := filepath.Join(runtimeRoot, "examples", "demo", "stdio", "my-tool-bin")
	if _, statErr := os.Stat(stdioBinPath); os.IsNotExist(statErr) {
		t.Logf("Building missing stdio example binary: %s", stdioBinPath)
		buildCmd := exec.Command("go", "build", "-o", stdioBinPath, filepath.Join(runtimeRoot, "examples", "demo", "stdio", "my-tool", "main.go"))
		buildCmd.Env = append(os.Environ(), "GOCACHE="+filepath.Join(t.TempDir(), "gocache"))
		buildCmd.Dir = runtimeRoot
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if buildExecErr := buildCmd.Run(); buildExecErr != nil {
			t.Logf("Failed to build stdio example binary (continuing, but validation might fail): %v", buildExecErr)
		}
	}

	// Walk through examples directory
	walkingErr := filepath.Walk(examplesDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// Check for config.yaml
		if !info.IsDir() && filepath.Base(path) == "config.yaml" {
			// Trim project root from path for cleaner test name
			relPath, relPathErr := filepath.Rel(runtimeRoot, path)
			require.NoError(t, relPathErr)

			t.Run(relPath, func(t *testing.T) {
				validateConfig(t, path)
			})
		}
		return nil
	})
	require.NoError(t, walkingErr)
}

func validateConfig(t *testing.T, configPath string) {
	osFs := afero.NewOsFs()

	// Create a store that points to this config file
	store := config.NewFileStore(osFs, []string{configPath})

	// Load services
	// The second argument "server" matches what the CLI uses for validation context if any
	configs, loadErr := config.LoadServices(context.Background(), store, "server")
	if loadErr != nil {
		// Some configs might require env vars which validly fail if missing.
		// However, LoadServices typically parses the YAML/Proto.
		// If it fails due to missing env vars that are required for *parsing* (if any), that might be acceptable if we can detect it.
		// But usually configs placeholders are just strings unless they are used in a way that breaks parsing.
		t.Errorf("failed to load services from %s: %v", configPath, loadErr)
		return
	}

	// Validate services
	if validateErr := configs.Validate(); validateErr != nil {
		t.Errorf("validation failed for %s: %v", configPath, validateErr)
	}
}

func sourceProjectRoot() (string, error) {
	const workspace = "mcp-any"
	for _, base := range []string{os.Getenv("TEST_SRCDIR"), os.Getenv("RUNFILES_DIR")} {
		if base == "" {
			continue
		}
		candidate := filepath.Join(base, workspace, "server")
		if _, statErr := os.Stat(filepath.Join(candidate, "examples")); statErr == nil {
			return candidate, nil
		}
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", os.ErrNotExist
	}
	// From server/tests/integration/examples/examples_test.go -> server/
	return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(file)))), nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, copyWalkErr error) error {
		if copyWalkErr != nil {
			return copyWalkErr
		}
		relPath, relErr := filepath.Rel(src, path)
		if relErr != nil {
			return relErr
		}
		target := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		in, openInErr := os.Open(path)
		if openInErr != nil {
			return openInErr
		}
		defer in.Close()
		out, openOutErr := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if openOutErr != nil {
			return openOutErr
		}
		defer out.Close()
		_, copyErr := io.Copy(out, in)
		return copyErr
	})
}
