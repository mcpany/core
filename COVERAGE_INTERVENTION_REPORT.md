# Coverage Intervention Report

## Phase 1: Risk-Based Discovery
Top 10 most critical untested components:
1. `pkg/storage/interface.go` (665 lines) - Core storage interface definition.
2. `pkg/app/seeds.go` (451 lines) - Seed configurations for upstream templates.
3. `pkg/mcpserver/noop_managers.go` (407 lines) - Fallback mock managers.
4. `pkg/tool/websocket.go` (251 lines) - WebSocket tool implementation.
5. `pkg/serviceregistry/mock_registry.go` (208 lines) - Mock registry for testing.
6. `pkg/bus/message.go` (181 lines) - Core message bus payload types.
7. `pkg/upstream/filesystem/provider/zip.go` (171 lines) - ZIP file system provider.
8. `pkg/storage/sqlite/db.go` (163 lines) - SQLite database implementation.
9. `pkg/app/api_hitl.go` (162 lines) - Human-in-the-Loop approval workflows.
10. `pkg/client/grpc_client_wrapper.go` (154 lines) - gRPC client wrapper logic.

## Phase 4: Impact Report

*   **Target:** `server/pkg/app/api_hitl.go`
*   **Risk Profile:** High. This module orchestrates the Human-in-the-Loop (HITL) approval workflows for sensitive tools. As a critical component of the platform's security and operational guardrails, its lack of testing presented a significant risk. The file handles both HTTP endpoints for user interaction and concurrent state management (`globalHITLState`) coupled with asynchronous event bus communications, making it highly susceptible to subtle race conditions or logic errors.
*   **New Coverage:**
    *   **Happy Paths:** Successfully querying pending approvals (`GET /hitl/approvals`), approving a request (`POST /hitl/approvals/{id}` with `action=approved`), and rejecting a request (`POST /hitl/approvals/{id}` with `action=rejected`).
    *   **Edge Cases:** Rejecting incorrect HTTP methods, handling malformed request bodies, and validating asynchronous message ingestion from the `hitl.requests` event bus.
    *   **Mechanisms:** Mocks were employed for the `busProvider` to ensure hermetic testing of both ingress and egress messaging pathways.
*   **Verification:** Confirm that tests pass cleanly via Bazel (`./bazelisk test //server/pkg/app:app_test --test_filter=TestMountHITL` and full suite).
