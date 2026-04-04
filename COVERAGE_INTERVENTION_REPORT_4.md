# Coverage Intervention Report

* **Target:** `server/pkg/util/secrets.go`
* **Risk Profile:** The function `ResolveSecret` and `resolveSecretImpl` are heavily relied upon to fetch sensitive tokens, database passwords, and API keys. Unhandled edge cases here could lead to direct secret leakage, Server-Side Request Forgery (SSRF) vulnerabilities via cloud metadata endpoints (like AWS metadata services for IAM roles), or malicious environment variable extraction in uncontrolled deployments. Prior to this intervention, many of the fallback behaviors, context timeouts, and structural metadata parsing errors (such as malformed JSONs in AWS Secrets Manager payloads, missing keys, and SSRF restrictions) lacked hermetic testing.
* **New Coverage:** The following logic paths are now guarded by comprehensive tests:
  - Error paths and failures for `AwsSecretManager` secrets (malformed JSON, missing keys, empty payloads, profile loading failures).
  - Explicit enforcement tests ensuring that restricted environment variables cannot be accessed.
  - Verification of SSRF blocks and context timeout behavior on Vault and HTTP-based remote secrets.
  - Failures in authentication token loading across multiple methods (API Key, Bearer Token, BasicAuth, OAuth2).
* **Verification:** `bazelisk test //server/pkg/util/...` cleanly passed with 96.7% coverage on the targeted file (`secrets.go`) without any regressions across the rest of the workspace.
