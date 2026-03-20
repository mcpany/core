# Coverage Intervention Report

## Phase 1: Risk-Based Discovery (The Heatmap)

Based on a scan of the repository prioritizing complex logic with low test coverage, here are the top 10 most critical untested or under-tested components:

1.  `server/pkg/app/api_audit.go` (Audit log retrieval, filtering, and export)
2.  `server/pkg/app/api_extra.go` (Resource read and prompt execution routing)
3.  `server/pkg/storage/interface.go` (Core storage abstraction interface definitions)
4.  `server/pkg/app/seeds.go` (Database seeding logic for core entities)
5.  `server/pkg/mcpserver/noop_managers.go` (No-op implementations for MCP capabilities)
6.  `server/pkg/tool/websocket.go` (WebSocket-based tool execution routing)
7.  `server/pkg/serviceregistry/mock_registry.go` (Mock service registry implementations)
8.  `server/pkg/storage/sqlite/db.go` (SQLite database connection and migration logic)
9.  `server/pkg/bus/message.go` (Event bus message structures and serialization)
10. `server/pkg/auth/oauth_test_server.go` (OAuth test server configuration and routing)

I selected `server/pkg/app/api_audit.go` and `server/pkg/app/api_extra.go` for intervention because they directly map external HTTP requests to internal core logic and represent significant security and functionality risks if edge cases (like invalid input or storage errors) are mishandled.

## Phase 4: Impact Report

* **Target:** `server/pkg/app/api_audit.go` and `server/pkg/app/api_extra.go`
* **Risk Profile:** These files handle important HTTP endpoint mappings to core logic. `api_audit.go` handles retrieval and export of audit logs, critical for security and compliance. `api_extra.go` maps MCP resources and prompt executions. While some tests existed, edge case logic like missing parameters, incorrect body, and query format validation were completely untested. Because of the direct coupling of user input to backend processes, these endpoints represented a significant security regression risk.
* **New Coverage:**
    * In `api_audit_test.go`, I added `TestHandleAuditLogs_Errors` which explicitly covers invalid query parameters (`start_time`, `end_time`, `limit`, `offset`), method not allowed, and internal server store errors using standard table-driven testing.
    * In `api_extra_test.go` (newly created), I added robust table-driven tests `TestHandleResourceRead` and `TestHandlePromptExecute` spanning HTTP method violations, missing arguments, invalid body streams, and downstream read/execution errors. Mocking was utilized via `gomock` to hermetically isolate edge case validation from live storage.
* **Verification:** `bazel test //server/pkg/app:app_test` confirms the newly added endpoint tests pass cleanly. `bazel test //server/...` verified that the rest of the legacy codebase remained completely functional, showing no regressions across all tests. All lints pass correctly.
