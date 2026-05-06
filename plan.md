1.  **Sync & Understand:** We identified that CVE-2026-32922 demands Privilege-Constrained Token Rotation (PCTR). This ensures that any rotated token has scopes that are a strict subset of the original token.
2.  **Evolve:** Update `docs/research/market-sync-2026-07-25-pctr.md` to include a 'Core Logic' section and a Mermaid diagram mapping the flow from the Gateway to the Adapters.
3.  **Implement PCTR Middleware:** Implement `PCTRMiddleware` in `server/pkg/middleware/pctr.go` with Go code that accepts old token scopes and requested new token scopes (or simulates the validation logic for subset enforcement).
4.  **Register API:** Update `server/pkg/app/api.go` to mount the `PCTRMiddleware` API handler under `/auth/rotate`.
5.  **Verify Registration:** Verify the API registration in `server/pkg/app/api.go` using `read_file` or `git diff`.
6.  **Write Tests & Update BUILD.bazel:** Write tests in `server/pkg/middleware/pctr_test.go` confirming that a token rotation request with a strict subset of scopes succeeds, and a request with escalated scopes fails. Add the new files to `BUILD.bazel`.
7.  **Verify BUILD.bazel:** Use `read_file` to inspect the updated `BUILD.bazel` file to ensure the new files are correctly included.
8.  **Run specific test:** Verify tests run cleanly with `bazelisk test //server/pkg/middleware/...`
9.  **Full Test Suite:** Run `bazelisk test //...` to ensure all tests pass and no regressions are introduced. Fix any failures.
10. **Pre-commit:** Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.
11. **Submit PR:** Submit the change with a technical PR description.
