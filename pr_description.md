# Truth Reconciliation Audit

## Executive Summary
This PR aligns the documentation, codebase, and roadmap by removing the deprecated `service_config` wrapper syntax from the `UpstreamService` documentation, fixing a divergence between the code (which rejects the wrapper natively) and the documentation (which incorrectly documented it). The 10 sampled documentation files were successfully audited against codebase reality and roadmap expectations.

## Verification Matrix
| Document Name | Status | Action Taken | Evidence |
| --- | --- | --- | --- |
| `ui/docs/features/browser_connectivity_check.md` | Verified | None | Implemented in `ui/src/components/diagnostics/connection-diagnostic.tsx` with `mode: 'no-cors'` |
| `ui/docs/features/playground.md` | Verified | None | Native file upload implemented in `ui/src/components/playground/schema-form.tsx` for base64 strings |
| `ui/docs/features/logs.md` | Verified | None | Centralized log stream implemented in `ui/src/app/logs/page.tsx` |
| `ui/docs/features/tool-diff.md` | Verified | None | Visual diffing tools correctly configured |
| `ui/docs/features/tag-based-access-control.md` | Verified | None | Tag assignments mapped to Profile Editor in UI |
| `server/docs/features/schema-validation.md` | Verified | None | Startup validation blocks invalid schema keys correctly (like `service_config` wrapper) |
| `server/docs/features/rate-limiting/README.md` | Verified | None | Token-based cost metric logic checks out |
| `server/docs/features/webhooks/README.md` | Verified | None | CloudEvents integration functional in `server/cmd/webhooks` sidecar |
| `server/docs/features/granular_scopes.md` | Verified | None | Scope bindings to roles enforce zero-trust intents |
| `server/docs/reference/configuration.md` | Drift | Removed `service_config` | Updated YAML wrapper specs in Markdown to show direct definition without wrapper |

## Remediation Log
- **Case A (Documentation Drift):** The roadmap stated the `service_config` wrapper for upstream configurations was fixed and deprecated. However, `server/docs/reference/configuration.md` still contained it in its tables and examples. This PR removes it from the documentation to match the codebase behavior (which intentionally throws a user-friendly error if it is used).

All existing tests were run to ensure these updates do not cause regressions.
