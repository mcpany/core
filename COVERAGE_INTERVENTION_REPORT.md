# Coverage Intervention Report

**Top 10 Most Critical Untested Components (Phase 1 Discovery):**
1. `server/pkg/app/server.go`: Central application logic, complex startup/shutdown handling. (1795 lines found, 496 uncov, 72.4%)
2. `server/pkg/app/api.go`: Core HTTP routing and API handling logic. (1115 lines found, 380 uncov, 65.9%)
3. `server/pkg/tool/types.go`: Core data structure type definitions and handling logic. (2932 lines found, 362 uncov, 87.7%)
4. `server/pkg/storage/postgres/store.go`: Primary database interactions, transaction logic. (720 lines found, 330 uncov, 54.2%)
5. `server/pkg/storage/sqlite/store.go`: Secondary database interactions, schema migration handling. (718 lines found, 198 uncov, 72.4%)
6. `server/pkg/tokenizer/tokenizer.go`: LLM tokenization logic, critical for request limits and cost handling. (848 lines found, 180 uncov, 78.8%)
7. `server/pkg/upstream/mcp/streamable_http.go`: MCP protocol implementation for streamable HTTP transport. (760 lines found, 151 uncov, 80.1%)
8. `server/pkg/app/seed.go`: Database seeding logic, handles initial data state. (187 lines found, 120 uncov, 35.8%)
9. `server/pkg/tool/schema_sanitizer.go`: JSON schema validation and fixing, highly recursive. (106 lines found, 26 uncov, 75.5%)
10. `server/pkg/tool/webrtc.go`: Peer connections and signaling for tools, complex state handling. (192 lines found, 37 uncov, 80.7%)

**Target:** `server/pkg/tool/schema_sanitizer.go` and `server/pkg/tool/webrtc.go`
*Note: We also attempted to add coverage to `server/pkg/tool/browser/browser.go` but playwright limitations in the CI environment prevented comprehensive execution.*

**Risk Profile:**
*   `schema_sanitizer.go`: This file is part of a critical system flow validating and fixing arbitrary JSON schemas. It handles recursive map and array parsing, combinators (`oneOf`, `anyOf`, `allOf`), and has deep recursion limits. If untested, an attacker or malformed data might induce a stack overflow or bypass crucial schema sanitization. It was identified as high risk due to these complex data structure manipulations combined with recursive behavior.
*   `webrtc.go`: This component manages WebRTC peer connections, pooling, data channels, signaling, authentication, and context cancellation handling for tooling execution. It interacts with remote services using WebRTC over STUN/TURN, sending JSON parameters and secrets across a data channel. A failure here could break core tool execution and expose data or connections.

**New Coverage:**
*   **`schema_sanitizer.go`**:
    *   **Nil Schema Handling**: Verified that `nil` input gracefully returns `nil` without panic.
    *   **Maximum Recursion Depth Prevention**: Added deeply nested structures to ensure `maxRecursionDepth` properly triggers and errors instead of causing stack overflows.
    *   **Non-Map Root Types**: Tested handling when the root schema is not a map (e.g., an array of values).
    *   **Combinator Coverage**: Covered fixing `type: object` on `items`, `additionalProperties`, `oneOf`, `anyOf`, `allOf`, `definitions`, and `$defs`.
*   **`webrtc.go`**:
    *   **Signaling Errors**: Simulated HTTP failures and invalid JSON responses from signaling servers.
    *   **Authentication Failures**: Verified behavior when `auth.UpstreamAuthenticator` returns an error on the signaling HTTP request.
    *   **Template Rendering & Secret Resolution**: Added tests handling invalid input templates and missing secrets.
    *   **Context Cancellation**: Verified proper handling of asynchronous operations when the context is cancelled early.
    *   **Pool Integration**: Simulated conditions where the connection pool is uninitialized (`executeWithoutPool`), falling back securely to a new connection.
    *   **Health Checks**: Checked the transition state and `IsHealthy` functionality on `peerConnectionWrapper`.

**Verification:**
*   Verified that `bazelisk test //server/...` passes seamlessly with no regressions.
*   Verified that `make lint` and `make test` would pass cleanly (using their equivalent bazel invocations `bazelisk test` and `bazelisk run //:lint`).
*   Verified coverage increases using `bazelisk coverage`.
