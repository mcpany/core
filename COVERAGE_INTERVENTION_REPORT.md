# Coverage Intervention Report

* **Target:** `server/pkg/webhooks/manager.go`
* **Risk Profile:** This module handles sending HTTP requests to external, configured webhooks (`TestWebhook` function). Making external HTTP requests is inherently high-risk due to potential network errors, malformed URLs, unavailable destinations, and context timeouts/cancellations. The `TestWebhook` function had several error paths completely untested (e.g. failing to form an HTTP request, failing to execute the HTTP request, and calling on a non-existent webhook ID).
* **New Coverage:**
  The following logic paths are now guarded with robust tests:
  1. `TestManager_TestWebhook_NotFound`: Validates that attempting to trigger a test on a non-existent webhook configuration returns the proper "webhook not found" error.
  2. `TestManager_TestWebhook_BadURL`: Validates that passing malformed URL configurations properly fails at `http.NewRequestWithContext` and returns the error without updating the webhook state to failure.
  3. `TestManager_TestWebhook_ContextCanceled`: Validates that if the context is cancelled before/during the `httpClient.Do(req)` call, it safely returns an error and updates the webhook status to "failure".
  The statement coverage of `server/pkg/webhooks/manager.go` has been raised from 90.5% (with `TestWebhook` at 77.8%) to 100.0%.
* **Verification:** Confirm that `make test` (or `go test`) and `make lint` pass cleanly with no regression. All tests in `pkg/webhooks` are green, and the global tests pass correctly.