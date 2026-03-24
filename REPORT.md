* **Target:** `server/pkg/alerts/manager.go`
* **Risk Profile:** The `alerts` package lacked any unit tests for the critical Webhook notification functionality (`SetWebhookURL`, `CreateAlert`, `UpdateAlert`). Given that alerts represent system incidents, failures in webhook delivery pose a high operational risk as operators would not be notified of critical issues. The cyclomatic complexity of `manager.go` combined with low test coverage made it a prime candidate.
* **New Coverage:**
  * Guards the happy path of alert creation triggering a webhook call with correct JSON payload.
  * Guards the happy path of alert status updates triggering a webhook call.
  * Guards the edge case where updating a non-existent alert does not trigger a webhook call.
* **Verification:** Confirmed that `bazel test //server/pkg/alerts/...` passes cleanly and the broader test suite `bazel test //...` introduces no regressions.
