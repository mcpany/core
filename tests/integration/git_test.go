// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitUpstreamService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Ensure git repo has enough commits for git diff HEAD~1 HEAD to work
	ensureEnoughCommits(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Start the MCP Any server with the git example config
	serverInfo := StartMCPANYServer(t, "git-test", "--config-path", "examples/git/config.yaml")
	defer serverInfo.CleanupFunc()

	// Connect using MCP SDK Client to ensure proper session initialization
	client := mcp.NewClient(&mcp.Implementation{Name: "git-test-client", Version: "1.0.0"}, nil)
	transport := &mcp.StreamableClientTransport{Endpoint: serverInfo.HTTPEndpoint}

	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err)
	defer session.Close()

	// Call the git status tool and check for a more specific pattern
	statusResult, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "git.status"})
	require.NoError(t, err)
	statusText := getToolOutput(t, statusResult)
	// Support both "On branch ..." and "HEAD detached ..."
	assert.Regexp(t, regexp.MustCompile(`(On branch \w+|HEAD detached .*)`), statusText)

	// Call the git log tool and check for a commit hash
	logResult, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "git.log"})
	require.NoError(t, err)
	logText := getToolOutput(t, logResult)
	assert.Regexp(t, regexp.MustCompile(`commit [0-9a-f]{40}`), logText)

	// Call the git diff tool and check for a diff header
	diffResult, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "git.diff"})
	require.NoError(t, err)
	diffText := getToolOutput(t, diffResult)
	// With the dummy commit, we should have a diff
	assert.Regexp(t, regexp.MustCompile(`diff --git a/.* b/.*`), diffText)
}

func ensureEnoughCommits(t *testing.T) {
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	output, err := cmd.Output()
	if err == nil {
		count := strings.TrimSpace(string(output))
		// If 0 or 1 commit, we need more history for HEAD~1 to work
		if count == "1" || count == "0" {
			t.Log("Adding dummy commit to ensure enough history for git diff tests")

			// Configure git user if not present (might be needed in some CI envs)
			exec.Command("git", "config", "user.email", "test@example.com").Run()
			exec.Command("git", "config", "user.name", "Test User").Run()

			// Create a dummy file
			exec.Command("touch", "dummy_test_file_auto").Run()
			exec.Command("git", "add", "dummy_test_file_auto").Run()
			exec.Command("git", "commit", "-m", "Auto dummy commit").Run()
		}
	} else {
		t.Logf("Failed to check git commit count: %v", err)
	}
}

func getToolOutput(t *testing.T, result *mcp.CallToolResult) string {
    text := getTextContent(t, result)
    var resultMap map[string]interface{}
    // The output might be JSON string if it comes from CommandTool.
    if err := json.Unmarshal([]byte(text), &resultMap); err == nil {
        if output, ok := resultMap["combined_output"].(string); ok {
            return output
        }
        if stdout, ok := resultMap["stdout"].(string); ok {
             return stdout
        }
    }
    return text
}

func getTextContent(t *testing.T, result *mcp.CallToolResult) string {
	if len(result.Content) == 0 {
		return ""
	}
	content := result.Content[0]
	if textContent, ok := content.(*mcp.TextContent); ok {
		return textContent.Text
	}
	// Fallback or failure
	t.Logf("Unexpected content type: %T", content)
	return fmt.Sprint(content)
}
