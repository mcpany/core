* **Target:** `server/pkg/alerts/manager.go` (`CreateAlert` and `UpdateAlert` functions)

### Phase 1: Risk-Based Discovery (The Heatmap)

Based on a test coverage check (`go test -cover ./...`), there were several untested high-risk areas identified, out of which the Top 10 High-Risk Untested Components include:

1. `server/pkg/alerts/manager.go` (35% coverage on specific methods like `CreateAlert` and `UpdateAlert` regarding the webhook firing mechanisms).
2. `server/pkg/audit/postgres.go` (11.8% coverage).
3. `server/pkg/audit/datadog.go` (50.0% coverage).
4. `server/pkg/audit/splunk.go` (50.0% coverage).
5. `server/pkg/audit/webhook.go` (50.0% coverage).
6. `server/pkg/auth/oauth.go` (59.3% coverage).
7. `server/pkg/config/store.go` (Various un-covered loading methods).
8. `server/pkg/auth/oidc.go` (71.4% coverage).
9. `server/pkg/tool/management.go` (Some subroutines with un-covered logic paths).
10. `server/pkg/mcpserver/noop_managers.go` (0% coverage).

We selected `server/pkg/alerts/manager.go` because the alerting manager is central to incident notification and response within the system. Prior to this intervention, the asynchronous delivery path for webhooks triggered upon alert creation or updates had zero test coverage. Because this is an external network call (`http.DefaultClient.Do`), it presents a high risk for edge-case errors, such as connection failures, HTTP 500s from the destination server, or invalid URLs. Without coverage, the failure of a webhook could go unnoticed, or a malformed alerting payload could crash the system logic unexpectedly.

* **New Coverage:**
  * **Happy Path:** Validates that `CreateAlert` triggers a POST request to a configured URL containing a correctly marshaled alert payload with the correct `application/json` content type. Validates that `UpdateAlert` similarly triggers an update webhook payload.
  * **Edge Cases:** Validates that if the webhook target returns an HTTP 500 internal server error, the process correctly skips without crashing the application. Validates that an un-routable / un-resolvable webhook URL correctly completes its path without crashing the alert generator routine.
  * The test coverage for the `server/pkg/alerts` package increased from ~71.3% to ~91.9%.

* **Verification:** `make lint` passes successfully. `go test -cover -race ./pkg/alerts/...` passes with green assertions on all new coverage and checks.
