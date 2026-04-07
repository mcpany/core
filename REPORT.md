# Coverage Intervention Impact Report

* **Target:** `server/pkg/upstream/http/http.go`
* **Risk Profile:** This code handles the auto-discovery, configuration parsing, validation, and registration of HTTP tools and prompts with their respective managers. The logic is complex (especially `createAndRegisterHTTPTools` and `createAndRegisterPrompts`), dealing with protobuf schema conversions, url parsing, double-slash normalization logic, call policy evaluation, and authenticator setup. A regression here would silently or explicitly prevent critical capabilities from being registered and exposed to LLMs.
* **New Coverage:**
  - `createAndRegisterHTTPTools`: Now guards the happy path (auto-discovery of tools configured via `HttpCall` structs), handling of base URL parsing errors, missing tool definitions mapping, and skipping disabled tools.
  - `createAndRegisterPrompts`: Now guards the happy path (successfully registering a prompt into the `ManagerInterface`) and the disabled prompt edge case (skipping registration if the `Disable` flag is true). Tested actual registration outcome by verifying the internal state of the injected `prompt.Manager`.
* **Verification:** Confirmed that `bazelisk test //server/pkg/upstream/http:http_test` and `bazelisk test //server/...` passed cleanly, ensuring no regressions.
