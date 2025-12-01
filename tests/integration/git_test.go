// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"regexp"
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Start the MCP Any server with the git example config
	serverInfo := StartMCPANYServer(t, "git-test", "--config-path", "examples/git/config.yaml")
	defer serverInfo.CleanupFunc()

	// Call the git status tool and check for a more specific pattern
	statusResult, err := serverInfo.CallTool(ctx, &mcp.CallToolParams{Name: "git.status"})
	require.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile(`On branch \w+`), statusResult.Content)

	// Call the git log tool and check for a commit hash
	logResult, err := serverInfo.CallTool(ctx, &mcp.CallToolParams{Name: "git.log"})
	require.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile(`commit [0-9a-f]{40}`), logResult.Content)

	// Call the git diff tool and check for a diff header
	diffResult, err := serverInfo.CallTool(ctx, &mcp.CallToolParams{Name: "git.diff"})
	require.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile(`diff --git a/.* b/.*`), diffResult.Content)
}
