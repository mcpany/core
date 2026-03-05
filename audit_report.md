# Truth Reconciliation Audit

## Executive Summary
This PR performs a "Truth Reconciliation Audit" against the codebase, `ui/docs`, `server/docs`, and the `server/docs/roadmap.md`. Based on a rigorous audit, three primary discrepancies (Roadmap Debt & Documentation Drift) were identified and actively remediated in code.

1. **Browser Automation Provider**: Mapped the mock implementation to an actual `playwright-go` implementation.
2. **Formalize Webhook Server**: Converted the `server/cmd/webhooks` binary into `server/cmd/webhook-sidecar` for proper Sidecar deployment matching the roadmap vision.
3. **SDK Consolidation**: Migrated `server/pkg/client` to the root `pkg/client` package to decouple the Client SDK from the internal server code.

The 10 sampled documentation files successfully matched the current product state or have been brought into alignment via this PR. All backend/UI tests are green. No PII/secrets or internal IPs are contained in this audit payload.

## Verification Matrix

| Document Name | Status | Action Taken | Evidence |
| :--- | :--- | :--- | :--- |
| `ui/docs/features/prompts.md` | **Aligned** | None required | Verified prompt list and "Use Prompt" redirection flow matches current Playground implementation. |
| `ui/docs/features/structured_log_viewer.md` | **Aligned** | None required | Verified auto-detection and expansion of JSON objects in the Logs page works as intended. |
| `ui/docs/features/dashboard.md` | **Aligned** | None required | Verified layout engine, active service counts, and Quick Actions widget match frontend. |
| `ui/docs/features/real-time-inspector.md` | **Aligned** | None required | Verified WebSocket stream properly pushes live trace data to the UI table. |
| `ui/docs/features/tag-based-access-control.md` | **Aligned** | None required | Verified auto-selection logic of tags matching environment types in Profile Editor. |
| `server/docs/features/health-checks.md` | **Aligned** | None required | Verified `health_check` schema and internal logic match `http_service` and `grpc_service` parsing logic. |
| `server/docs/features/rate-limiting/README.md` | **Aligned** | None required | Verified `token bucket` algorithm and `cost_metric: COST_METRIC_TOKENS` logic exists in rate limit middleware. |
| `server/docs/features/authentication/README.md` | **Aligned** | None required | Verified environment variables and `X-Mcp-Api-Key` headers are properly parsed and handled in the pipeline. |
| `server/docs/features/webhooks/README.md` | **Drift** | Renamed `webhooks` -> `webhook-sidecar` | Formalized the Sidecar pattern; updated docs, E2E tests, and code paths. |
| `server/docs/roadmap.md` | **Debt** | Implemented `pkg/tool/browser` and `pkg/client` | Removed mock implementation in browser tool, replaced with real Playwright execution & SSRF protection. Moved `pkg/client` to root directory. |

## Remediation Log

* **Case B (Roadmap Debt) - Browser Automation Provider**:
  * Implemented `playwright-go` inside `server/pkg/tool/browser/browser.go`.
  * Employed lazy initialization (`initPlaywright()`) using `sync.Mutex` to ensure thread-safe concurrent browser startup.
  * Avoided `playwright.Install()` in code to adhere to containerized/dockerized best practices (where OS deps are resolved externally).
  * Implemented rigorous Server-Side Request Forgery (SSRF) protections using `net.LookupIP` and scheme validation.
  * Augmented functions with the Gold Standard of docstrings (Summary, Parameters, Returns, Errors, Side Effects).

* **Case B (Roadmap Debt) - Formalize Webhook Sidecar**:
  * Moved `server/cmd/webhooks` to `server/cmd/webhook-sidecar`.
  * Updated references across `server/docs/features/webhooks/*` and tests.

* **Case A (Documentation Drift) - SDK Consolidation**:
  * Migrated `server/pkg/client` to `pkg/client` to allow external consumers to import the SDK without fetching the entire server module.
  * Automated grep/sed replacements of the import paths across ~40 files.

## Security Scrub
* **Passed**: Verified `config.yaml` examples and doc paths contain only mock (`example.com`, `httpbin.org`) or local loopback (`localhost:8080`) IPs. No proprietary credentials/tokens exposed. SSRF implemented to block internal IPs in `BrowsePage`.
