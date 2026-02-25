# Coverage Intervention Report

## Target
**File:** `server/pkg/tool/websocket.go`

## Risk Profile
**Selection Reason:** High Risk / Core Logic.
The `WebsocketTool` handles real-time communication with upstream services. It is responsible for:
1.  **Connection Management:** Managing WebSocket connections via a pool.
2.  **Secret Injection:** Resolving and injecting sensitive secrets (API keys) into parameters.
3.  **Data Transformation:** Transforming input and output data using templates and JQ queries.

**Prior State:**
-   **Coverage:** Partial. While basic execution was tested, specific edge cases like **secret resolution** and integration with the environment were not fully covered or were relying on mock implementations that didn't exercise the full `ResolveSecret` path in context.
-   **Complexity:** Medium-High (Async communication, error handling, dynamic configuration).

## New Coverage
**Implemented Defenses:**
Enhanced `server/pkg/tool/websocket_tool_test.go` to include a dedicated test case for Secret Resolution.

**Guarded Paths:**
1.  **Secret Resolution:** Verifies that environment variables referenced in `WebsocketParameterMapping` are correctly resolved and injected into the WebSocket message payload.
2.  **Concurrency Safety:** Modified the test suite to safely handle environment variable manipulation (`t.Setenv`) by ensuring conflicting tests do not run in parallel.

## Verification
-   **New Tests:** `go test -v ./server/pkg/tool/ -run TestWebsocketTool_Execute` passed successfully.
-   **Regression Testing:** `go test -v ./server/pkg/tool/...` passed successfully, ensuring no negative impact on other tool implementations.
