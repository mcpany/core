// Copyright 2024 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package public_api

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestNASAGIBS(t *testing.T) {
	server := integration.StartMCPANYServer(t, "nasa-gibs-test", "--config-paths", "../../examples/popular_services/nasa/config.yaml")
	defer func() {
		t.Logf("Server stdout:\n%s", server.Process.StdoutString())
		t.Logf("Server stderr:\n%s", server.Process.StderrString())
		server.CleanupFunc()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := &mcp.CallToolParams{
		Name: "nasa-gibs/-/get_tile",
		Arguments: map[string]interface{}{
			"LayerIdentifier": "MODIS_Terra_CorrectedReflectance_TrueColor",
			"Time":            "2012-07-09",
			"TileMatrixSet":   "250m",
			"TileMatrix":      "6",
			"TileRow":         "13",
			"TileCol":         "36",
		},
	}
	result, err := server.CallTool(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.Content)
}
