# Coverage Intervention Report

**Target:** `server/pkg/config/secrets.go`

**Risk Profile:**
This module handles the extraction (stripping) and hydration (injection) of secrets such as API keys, OAuth tokens, Basic Auth passwords, and environment variables across multiple critical configurations (Upstream Services, Profiles, Collections, Authentication). A misconfiguration or logic error in secret management would lead to high-impact security leaks or authorization failures in production. Given its complex interactions with recursive nested Protocol Buffer structs, it was identified as untested "Dark Matter" despite carrying extreme risk.

**New Coverage:**
I implemented robust table-driven testing in `server/pkg/config/secrets_test.go` covering the following core logic paths:
- `StripSecretsFromService`
- `StripSecretsFromProfile`
- `StripSecretsFromCollection`
- `StripSecretsFromAuth`
- `HydrateSecretsInService`

The tests assert that sensitive data fields are actively cleared out during redaction while preserving the rest of the configuration struct. They also verify that values retrieved from the external secret store are safely injected back into target config structs like `HTTPService` params, `Auth`, and `CommandLineService` environment variables. Edge cases including `nil` inputs are gracefully handled.

**Verification:**
`make test` and `make lint` passed cleanly. Bazel execution (`bazelisk test //server/pkg/config/...`) reported success, confirming we adhere to the "Do No Harm" principle and do not break the legacy test suite.
