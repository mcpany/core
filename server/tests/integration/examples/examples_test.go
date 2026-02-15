package examples_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExampleConfigs(t *testing.T) {
	// Set dummy API key for validation to pass
	t.Setenv("GEMINI_API_KEY", "dummy-key")
	// Find the project root
	wd, err := os.Getwd()
	require.NoError(t, err)

	// Assuming we run this from anywhere in the repo, we need to find the root.
	// Common pattern: look for go.mod
	projectRoot := findProjectRoot(t, wd)
	// Change to project root so that relative paths in configs (e.g. "./examples/...") resolve correctly
	// t.Chdir(projectRoot) is better than os.Chdir as it auto-restores, but check if available in environment.
	// Assuming Go 1.14+, t.Chdir is available.
	err = os.Chdir(projectRoot)
	require.NoError(t, err)

	// Ensure stdio example binary is built, as Config validation checks for its existence
	// This makes the test robust against sharding/environment where build-examples might not have run.
	stdioBinPath := filepath.Join(projectRoot, "examples", "demo", "stdio", "my-tool-bin")
	if _, err := os.Stat(stdioBinPath); os.IsNotExist(err) {
		t.Logf("Building missing stdio example binary: %s", stdioBinPath)
		cmd := exec.Command("go", "build", "-o", stdioBinPath, filepath.Join(projectRoot, "examples", "demo", "stdio", "my-tool", "main.go"))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			t.Logf("Failed to build stdio example binary (continuing, but validation might fail): %v", err)
		}
	}

	examplesDir := filepath.Join(projectRoot, "examples")

	// Walk through examples directory
	err = filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check for config.yaml
		if !info.IsDir() && filepath.Base(path) == "config.yaml" {
			// Trim project root from path for cleaner test name
			testName := path
			if strings.HasPrefix(path, projectRoot) {
				testName = strings.TrimPrefix(path, projectRoot)
			}
			t.Run(testName, func(t *testing.T) {
				validateConfig(t, path)
			})
		}
		return nil
	})
	require.NoError(t, err)
}

func validateConfig(t *testing.T, configPath string) {
	osFs := afero.NewOsFs()

	// Set dummy values for all required environment variables found in failure logs
	// This allows the strict config validation to pass during tests
	requiredEnvVars := []string{
		"AIRTABLE_API_TOKEN",
		"FIGMA_API_TOKEN",
		"GITHUB_TOKEN",
		"GOOGLE_API_KEY",
		"GOOGLE_OAUTH_CLIENT_ID",
		"GOOGLE_OAUTH_CLIENT_SECRET",
		"GOOGLE_OAUTH_REFRESH_TOKEN",
		"IPINFO_API_TOKEN",
		"MIRO_API_TOKEN",
		"NASA_OPEN_API_KEY",
		"SLACK_API_TOKEN",
		"STRIPE_API_KEY",
		"TRELLO_API_TOKEN",
		"TRELLO_API_KEY",
		"TWILIO_ACCOUNT_SID",
		"TWILIO_API_KEY",
		"TWILIO_API_SECRET",
	}

	for _, v := range requiredEnvVars {
		t.Setenv(v, "dummy-val")
	}

	// Create a store that points to this config file
	store := config.NewFileStore(osFs, []string{configPath})

	// Load services
	// The second argument "server" matches what the CLI uses for validation context if any
	configs, err := config.LoadServices(context.Background(), store, "server")
	if err != nil {
		// Some configs might require env vars which validly fail if missing.
		// However, LoadServices typically parses the YAML/Proto.
		// If it fails due to missing env vars that are required for *parsing* (if any), that might be acceptable if we can detect it.
		// But usually configs placeholders are just strings unless they are used in a way that breaks parsing.
		// Let's see if we fail on basic loading.
		t.Fatalf("Failed to load config %s: %v", configPath, err)
	}

	// Validate
	validationErrors := config.Validate(context.Background(), configs, config.Server)
	assert.Empty(t, validationErrors, "Config validation failed for %s", configPath)
}

func findProjectRoot(t *testing.T, startDir string) string {
	dir := startDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (go.mod)")
		}
		dir = parent
	}
}
