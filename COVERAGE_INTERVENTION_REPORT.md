# Coverage Intervention Report

## Phase 1: Risk-Based Discovery (Top 10 Most Critical Untested Components)

Based on cyclomatic complexity and absence of test coverage, the following 10 components were identified as the highest risk "Dark Matter" areas:

1. `server/pkg/storage/interface.go` (Complexity: 119) - Interface definition, no logical paths to test directly.
2. `server/pkg/tool/websocket.go` (Complexity: 49) - Core network communication and transformation logic.
3. `server/tools/license/remove.go` (Complexity: 47) - Tooling script.
4. `server/tools/check_doc.go` (Complexity: 36) - Tooling script.
5. `server/examples/upstream_service_demo/webrtc/server/main.go` (Complexity: 33) - Example code.
6. `server/examples/upstream_service_demo/webrtc/client/main.go` (Complexity: 30) - Example code.
7. `server/pkg/mcpserver/noop_managers.go` (Complexity: 30) - Trivial logic returning nils.
8. `server/pkg/serviceregistry/mock_registry.go` (Complexity: 26) - Mock logic.
9. `server/pkg/tool/mock_tool.go` (Complexity: 23) - Mock logic.
10. `server/pkg/tool/base.go` (Complexity: 22) - Core utility functions for tools.

## Phase 4: Impact Report

* **Target:** `server/pkg/tool/websocket.go`
* **Risk Profile:** This file was selected because it is high risk. It ranks second overall in complexity among untested paths (Complexity: 49) and handles core execution logic, data transformation, authorization checks, and network flows for WebSocket connections.
* **New Coverage:** Added hermetic, table-driven tests mimicking the existing testing style in `server/pkg/tool/websocket_tool_test.go`. The new paths guarded include:
  * **Secret Resolution Failure:** Verified that if a parameter relies on a secret (e.g., an unset environment variable), the tool immediately traps the error during execution and bubbles it up securely (`secret_resolution_error`).
  * **Input Transformation Render Error:** Verified that invalid template variables during rendering (e.g. referencing an unknown function) securely halt the execution without crashing the application (`input_transformer_render_error`).
* **Verification:** Confirmed that `./bazelisk test //...` and `make lint` passed cleanly, ensuring no regressions and adherence to quality standards.
