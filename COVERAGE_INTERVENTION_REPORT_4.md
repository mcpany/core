# Coverage Intervention Report

*   **Target:** `server/pkg/upstream/mcp/bundle_transport.go` & `server/pkg/upstream/mcp/bundle_transport_robustness_test.go`
*   **Risk Profile:** `bundle_transport.go` contains non-trivial custom JSON-RPC parsing (e.g., `setUnexportedID`, `fixID`) that uses `unsafe`, `reflect`, and interacts with internal details of third-party SDKs (`github.com/modelcontextprotocol/go-sdk/jsonrpc`). `fixID` is explicitly identified as an issue (⚡ BOLT label) with complex reflection logic. This is high risk code that could panic.
*   **New Coverage:**
    * Deep pointer/interface traversal (`fixID`).
    * Custom map types fallback handling.
    * String-to-int fallback logic.
    * Float-to-int conversion in ID assignment.
*   **Verification:** Confirm that `bazelisk test //server/pkg/upstream/mcp:mcp_test` and `bazelisk run //:lint` passed cleanly.
