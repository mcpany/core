# Coverage Intervention Report

* **Target:** `server/pkg/middleware/hitl.go`
* **Risk Profile:** The file handles Human-In-The-Loop approval flow logic, making it part of the critical control plane. It acts on the authorization boundary of "SensitiveTools". The tests for this middleware were "bypassing" the actual logic, substituting fake successful execution functions rather than actually asserting bus messages or timeouts. With low initial test coverage (25%) and a relatively high cyclomatic complexity (14 on `Execute`), this was an invisible risk in our system.
* **New Coverage:**
  * Added missing path testing for `TestHITLMiddleware_ApprovalGranted`: Correctly verifies that sensitive execution publishes approval requests, handles the approval response correctly, and proceeds with the middleware chain.
  * Added missing path testing for `TestHITLMiddleware_ApprovalDenied`: Asserts that when a human denies the request through the bus, the execution evaluates properly with a descriptive error.
  * Added missing path testing for `TestHITLMiddleware_Timeout`: Asserts that when no response is received, a timeout correctly rejects the request.
  * Overall test coverage increased from 25% to 87.5% for the `server/pkg/middleware/hitl.go` file.
* **Verification:** `make test` and `make lint` passed cleanly.

## Top 10 Most Critical Untested Components (High Complexity, Low Coverage)
1. `server/pkg/app/server.go` - `(*Application).runServerMode` (Complexity: 120) - Core application loop.
2. `server/pkg/upstream/http/http.go` - `(*Upstream).createAndRegisterHTTPTools` (Complexity: 88) - Upstream HTTP translation.
3. `server/pkg/tool/types.go` - `(*CommandTool).Execute` (Complexity: 66) - Core command execution.
4. `server/pkg/app/server.go` - `(*Application).Run` (Complexity: 63) - Application start entrypoint.
5. `server/pkg/tool/types.go` - `(*LocalCommandTool).Execute` (Complexity: 63) - Local command tool logic.
6. `server/pkg/util/secrets.go` - `resolveSecretImpl` (Complexity: 49) - Secrets resolution logic.
7. `server/pkg/app/server.go` - `(*Application).reconcileServices` (Complexity: 43) - Service reconciliation.
8. `server/pkg/tool/types.go` - `(*OpenAPITool).Execute` (Complexity: 40) - OpenAPI tool execution.
9. `server/pkg/middleware/registry.go` - `InitStandardMiddlewares` (Complexity: 37) - Core middleware registry configuration.
10. `server/pkg/tool/types.go` - `stripInterpreterComments` (Complexity: 36) - Command interpreter security sanitization.
