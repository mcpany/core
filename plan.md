1. **Assessment Phase**:
    *   Verify the existence of 10 sampled files.
    *   Determine discrepancies (already done):
        *   `ui/docs/features/hitl.md` (Mentions `/approvals` or `/hitl` route. Neither exists. -> Case B)
        *   `ui/docs/features/recursive_context.md` (Mentions `/context` route. Does not exist. -> Case B)
        *   `ui/docs/features/universal_agent_bus.md` (Mentions nested sections for Universal Agent Bus. Do not exist. -> Case B)
        *   `server/docs/features/shared_kv_store.md` (Check for Blackboard/KV code. Missing. -> Case B)
        *   `server/docs/features/granular_scopes.md` (Check for token scope code. Missing. -> Case B)
        *   `server/docs/features/lazy-mcp.md` (Check for lazy loading code. Missing. -> Case B)

2. **Backend Implementations (Case B)**:
    *   Use `write_file` to create `server/pkg/middleware/shared_kv_store.go` with `SharedKVStoreConfig` and `SharedKVStoreMiddleware` structs, and an `Execute` method handling "agent_aware" isolation.
    *   Use `write_file` to create `server/pkg/middleware/shared_kv_store_test.go` to test the middleware.
    *   Use `write_file` to create `server/pkg/middleware/granular_scopes.go` with `GranularScopesConfig` and `GranularScopesMiddleware` structs, parsing tokens like "fs:read:/tmp".
    *   Use `write_file` to create `server/pkg/middleware/granular_scopes_test.go` to test scope enforcement.
    *   Use `write_file` to create `server/pkg/middleware/lazy_mcp.go` with `LazyMCPConfig` and `LazyMCPMiddleware` structs, simulating on-demand discovery.
    *   Use `write_file` to create `server/pkg/middleware/lazy_mcp_test.go` to test lazy loading.

3. **Backend Verification**:
    *   Use `list_files` in `server/pkg/middleware` to confirm the 6 new files were created.
    *   Use `run_in_bash_session` to run `cd server && make test` (or `go test ./pkg/middleware/...`) to ensure tests compile and pass.

4. **Frontend Implementations (Case B)**:
    *   Use `run_in_bash_session` with `mkdir -p` to create directories `ui/src/app/approvals` (hitl), `ui/src/app/context`, and `ui/src/app/universal-agent-bus`.
    *   Use `write_file` to create `ui/src/app/approvals/page.tsx` exporting a basic React component for "HITL Approvals".
    *   Use `write_file` to create `ui/src/app/context/page.tsx` exporting a basic React component for "Recursive Context".
    *   Use `write_file` to create `ui/src/app/universal-agent-bus/page.tsx` exporting a basic React component for "Universal Agent Bus".
    *   Use `replace_with_git_merge_diff` on `ui/src/components/app-sidebar.tsx` to add `{ title: "Approvals", url: "/approvals", icon: ShieldCheck }` and `{ title: "Universal Agent Bus", url: "/universal-agent-bus", icon: Network }` to the `platformItems` array (the `/context` route already exists in this array, but I'll add the page for it).

5. **Frontend Verification**:
    *   Use `list_files` in `ui/src/app` to verify the new page directories and files.
    *   Use `read_file` on `ui/src/components/app-sidebar.tsx` to ensure navigation links were injected correctly.
    *   Use `run_in_bash_session` to run `cd ui && npm run lint` to ensure no syntax errors.

6. **Generate Audit Report**:
    *   Use `write_file` to create `pr_description.md` formatted strictly according to Phase 4 rules.

7. **Final Verification**:
    *   Run `make test` and `make lint` from the project root to ensure all changes are correct, exit criteria are met, and no regressions were introduced.

8. **Pre-commit**:
    *   Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.

9. **Submit**:
    *   Use `submit` to submit the PR.
