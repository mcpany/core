# Coverage Intervention Report

*   **Target:** `server/pkg/upstream/mcp/streamable_http.go` and `server/pkg/upstream/mcp/bundle_local_transport.go`
*   **Risk Profile:** These components manage upstream MCP health checks, tool/prompt/resource definitions, and local bundle connections via exec. Both had segments with 0% coverage on critical initialization or reporting pathways, which poses a severe risk in correctly routing and tracking upstream availability or connection state.
*   **New Coverage:**
    * `streamable_http.go`: Covered `CheckHealth` (both no checker and status reporting code paths) and `Definition` for translating prompt arguments into expected MCP property schemas.
    * `bundle_local_transport.go`: Covered `Connect`, ensuring we generate correct `StdioTransport` components when running local commands.
*   **Verification:** Verified that `go test ./pkg/upstream/mcp/...` passes cleanly without regression. Coverage in the mcp upstream module increased from 81.7% to 84.5%. Linting passes cleanly.
