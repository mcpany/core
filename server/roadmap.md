# Server Roadmap

## 1. Updated Roadmap

### Status: Active Development

### Implemented Features (Recently Completed)

- [x] **Strict JSON Schema Validation**: Enhanced configuration validator to strictly validate `input_schema` and `output_schema` against JSON Schema Draft 2020-12 using `jsonschema` library, preventing runtime errors due to invalid tool definitions. (Friction Fighter)
- [x] **Strict Config Parsing**: Startup now fails immediately on YAML/JSON syntax errors in configuration files (previously skipped silently).
- [x] **Upstream Service Diagnostics**: Improved error reporting in the UI for failed upstream connections (e.g., connection refused, auth failure).
- [x] **Config Environment Variable Validation**: Strict validation for missing environment variables in configuration files. [PR](#)
- [x] **Agent Debugger & Inspector**: Middleware for traffic replay and inspection. [Docs](server/docs/features/debugger.md)
- [x] **Context Optimizer**: Middleware to prevent context bloat. [Docs](server/docs/features/context_optimizer.md)
- [x] **Data Loss Prevention (DLP)**: Redaction of PII in inputs/outputs. [Docs](server/docs/features/dlp.md)
- [x] **Diagnostic "Doctor" API**: `mcpctl` validation and health checks. [Docs](server/docs/features/mcpctl.md)
  - _Update_: CLI implementation completed. Now supports connection verification and detailed health reporting.
- [x] **SSO Integration**: OIDC/SAML support. [Docs](server/docs/features/sso.md)
- [x] **Audit Log Export**: Native Splunk and Datadog integration. [Docs](server/docs/features/audit_logging.md)
- [x] **Cost Attribution**: Token-based cost estimation and metrics. [Docs](server/docs/features/rate-limiting/README.md)
- [x] **Distributed Tracing**: OpenTelemetry support. [Docs](server/docs/features/monitoring/tracing)
- [x] **Universal Connector Runtime**: Sidecar for stdio tools. [Docs](server/docs/features/connector_runtime.md)
- [x] **WASM Plugin System**: Runtime for sandboxed plugins. [Docs](server/docs/features/wasm.md)
- [x] **Hot Reload**: Dynamic configuration reloading. [Docs](server/docs/features/hot_reload.md)
- [x] **SQL Upstream**: Expose SQL databases as tools. [Docs](server/docs/features/sql_upstream.md)
- [x] **Webhooks Sidecar**: Context optimization and offloading. [Docs](server/docs/features/webhooks/sidecar.md)
- [x] **Dynamic Tool Registration**: Auto-discovery from OpenAPI/gRPC/GraphQL. [Docs](server/docs/features/dynamic_registration.md)
- [x] **Helm Chart Official Support**: K8s deployment charts. [Docs](server/docs/features/helm.md)
- [x] **Message Bus**: NATS/Kafka integration for events. [Docs](server/docs/features/message_bus.md)
- [x] **Structured Output Transformation**: JQ/JSONPath response shaping. [Docs](server/docs/features/transformation.md)
- [x] **Documentation Generator**: Auto-generate beautiful Markdown/HTML documentation. [Docs](server/docs/features/documentation_generation.md)
- [x] **Secrets Management Integration**: First-class integration with HashiCorp Vault or AWS Secrets Manager. [Docs](server/docs/reference/configuration.md#upstreamauthentication-outgoing)
- [x] **Resilient Configuration Loading**: Better error reporting for invalid configs and keeping services in the list with error state.
- [x] **Pre-flight Config Validation**: Added pre-flight checks for command existence and working directory validity to prevent "silent failures" at runtime.
- [x] **Strict Startup Validation**: Server now fails fast with descriptive errors if the initial configuration contains invalid services, preventing "zombie services" (Friction Fighter).
- [x] **Service Name Validation**: Enforce non-empty service names in configuration to prevent silent failures. (Friction Fighter)
- [x] **Filesystem Path Validation**: Proactive validation of configured `root_paths` for Filesystem services to warn about missing directories on startup.
- [x] **Enhanced Secret Validation**: Strict startup validation for referenced environment variables and secret files. Server will now fail fast if a configured secret (env var or file) does not exist.
- [x] **Robust Transport Error Reporting**: Improved error messages for command-based and Docker transports. Now captures and surfaces `stderr` when a process exits unexpectedly, aiding quick debugging of configuration or runtime errors.
- [x] **Proactive Schema Sanitization**: Automatically fixes common schema issues (like missing `type: object`) in tool definitions to ensure compatibility with strict MCP clients (e.g. Claude Code).
- [x] **Smart Config Error Messages**: Detect and guide users migrating from Claude Desktop configuration format (`mcpServers` vs `upstream_services`). (Friction Fighter)
- [x] **Relative Command Resolution**: Fixed an issue where relative commands in `stdio_connection` failed validation even if they existed in the specified `working_directory`. (Friction Fighter)
- [x] **Empty Configuration Guard**: Detect and error out if the user provides configuration sources but the resulting configuration is empty, preventing confusing "silent" startups with zero services. (Friction Fighter)
- [x] **Environment Variable Management**: Integrated `.env` file support using `godotenv` to automatically load environment variables at startup, reducing friction for local development and configuration. (Friction Fighter)
- [x] **Config Linting Tool**: Implemented `mcpany lint config.yaml` to detect security issues (plain text secrets, shell injection) and best practice violations before runtime. (Friction Fighter)
- [x] **Smart Stdio Argument Validation**: Detects and validates script files in `stdio_connection` arguments for common interpreters (Python, Node, etc.) while correctly handling remote URLs and module execution flags. (Friction Fighter)
- [x] **JSON Config Helper**: Added helpful error messages for JSON configuration files when users mistakenly use Claude Desktop format (`mcpServers`). (Friction Fighter)
- [x] **Live Logs Stream**: Fixed the WebSocket connection for the live logs dashboard by correcting the frontend URL path. Users can now see real-time server logs in the UI. (Experience Crafter)
- [x] **Intelligent Config Diagnostics**: Added fuzzy matching suggestions for "unknown field" errors (e.g. "Did you mean 'address'?" for 'target_address') and helpful hints for YAML syntax errors (like tabs vs spaces). (Friction Fighter)
- [x] **Smart Config Error Messages - Address**: Fixed an issue where the validator confusingly referred to 'target_address' instead of 'address', aligning error messages with the actual configuration schema. (Friction Fighter)
- [x] **Configuration Health Check**: Added a Doctor check that tracks the status of dynamic configuration reloads (success/failure, timestamp) and exposes it via `/doctor` API to aid in debugging silent reload failures. (Experience Crafter)
- [x] **Frontend Config Status Banner**: Add a visual indicator (banner/toast) in the UI Management Dashboard that queries the new `/doctor` configuration check and warns the user if the server is running with a stale or invalid configuration. (Experience Crafter)
- [x] **WebSocket Diagnostic Probe**: Implemented a diagnostic probe in the UI to validate WebSocket service configuration and test connectivity (including browser-side check). (Experience Crafter)
- [x] **Required Environment Variable Validation**: Added explicit `validation.required_env` support for Stdio connections to fail fast with descriptive errors if required environment variables are missing. (Friction Fighter)
- [x] **JSON Schema for Config**: Auto-generate and publish a JSON Schema for the `config.yaml` to enable intellisense/validation in IDEs like VS Code. (Friction Fighter)
- [x] **Config Schema Validation via CLI**: `mcpany config check config.yaml` that validates against the full JSON schema (including types and enums) using `jsonschema` library, providing line-number precise errors. (Friction Fighter)
- [x] **Strict Auth Config Validation**: Added startup validation for OAuth2, OIDC, and TrustedHeader configuration to ensure required fields (like `client_id`, `issuer`) are present. (Friction Fighter)
- [x] **Validation CLI Command**: Enhanced `mcpany config validate` with `--check-connection` flag to run deep checks, including connecting to upstream services to verify connectivity. Also added `mcpany doctor` for comprehensive system health diagnosis. (Friction Fighter)
- [x] **Upstream Connectivity Probe**: Implemented a startup connection check for HTTP upstream services. The server now logs a warning if the upstream service is unreachable during registration, preventing silent failures. (Friction Fighter)
- [x] **Actionable Config Errors**: Enhanced configuration validator to provide specific, actionable "Fix" suggestions for common errors like missing environment variables, files, or invalid URL schemes. (Friction Fighter)
- [x] **Alias & Suggestion Improvement**: Added direct support for "services" -> "upstream_services" alias suggestion in configuration errors, and improved fuzzy matching to prioritize fields from common configuration objects while avoiding misleading suggestions from irrelevant schemas. (Friction Fighter)
- [x] **Enhanced Service Visibility**: Added real-time health status, error reporting, and tool count metrics to the Service Management dashboard table. (Experience Crafter)
- [x] **Strict Startup Mode**: Implemented `--strict` flag for the `run` command. When enabled, the server performs pre-flight connectivity checks ("doctor" checks) on all configured upstream services and fails to start if any are unreachable, preventing silent failures. (Friction Fighter)
- [x] **Stdio Logging Visibility**: Fixed an issue where logs were discarded in stdio mode (used by clients like Claude Desktop). Logs now correctly output to stderr, ensuring "silent failures" are visible and debuggable. (Friction Fighter)
- [x] **OAuth2 Token Reachability Check**: Reliability: During validation (or via `doctor` command), attempt to connect to the OAuth2 token URL to verify network connectivity, going beyond simple syntax checks. (Friction Fighter)

## 2. Top 10 Recommended Features

These features represent the next logical steps for the product, focusing on Enterprise Readiness, Safety, and Developer Experience.

| Rank | Feature Name                                  | Why it matters                                                                                                                                                                                                     | Difficulty |
| :--- | :-------------------------------------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :--------- |
| 1    | **Tool Poisoning Mitigation**                 | **Security**: Implement integrity checks for tool definitions to prevent "Rug Pull" attacks where a tool definition changes maliciously after installation.                                                        | High       |
| 2    | **Interactive OAuth Handler**                 | **UX/Auth**: Solve the "copy-paste token" friction. Allow users to click "Login" in the UI to authenticate tools like GitHub/Google. (Backend Implemented, UI Pending)                                             | High       |
| 3    | **Interactive OAuth UI**                      | **UX**: Complete the UI implementation for the Interactive OAuth Handler to allow seamless user authentication flows.                                                                                              | Medium     |
| 4    | **Local LLM "One-Click" Connect**             | **Connectivity**: Auto-detect and connect to local inference servers (Ollama, LM Studio) to democratize AI access without cloud costs.                                                                             | Low        |
| 5    | **Tool "Dry Run" Mode**                       | **DevX**: Allow tools to define a "dry run" logic to validate inputs and permissions without executing side effects.                                                                                               | Medium     |
| 6    | **Team Configuration Sync**                   | **Collaboration**: Allow teams to synchronize `mcpany` configurations and secrets securely, ensuring consistent dev environments.                                                                                  | Medium     |
| 7    | **Smart Error Recovery**                      | **Resilience**: Use an internal LLM loop to analyze tool errors and automatically retry with corrected parameters (Self-Healing).                                                                                  | High       |
| 8    | **Canary Tool Deployment**                    | **Ops**: gradually roll out new tool versions to a subset of users or sessions to catch regressions before they impact everyone.                                                                                   | High       |
| 9    | **Compliance Reporting**                      | **Enterprise**: Automated generation of PDF/CSV reports from Audit Logs for SOC2/GDPR compliance reviews.                                                                                                          | Medium     |
| 10   | **Advanced Tiered Caching**                   | **Performance**: Implement a multi-layer cache (Memory -> Redis -> Disk) with configurable eviction policies to reduce upstream costs.                                                                             | Medium     |
| 11   | **Smart Retry Policies**                      | **Resilience**: Configurable exponential backoff and jitter for upstream connections to handle transient network failures gracefully.                                                                              | Medium     |
| 12   | **Service Dependency Graph**                  | **Observability**: Visualize dependencies between services (e.g. Service A calls Tool B in Service C) to detect circular dependencies or bottlenecks.                                                              | High       |
| 13   | **Schema Validation Playground**              | **DevX**: A dedicated UI in the dashboard to paste JSON/YAML and validate it against the server schema with real-time error highlighting.                                                                          | Low        |
| 14   | **Partial Reloads**                           | **Resilience**: When reloading config dynamically, if one service is invalid, keep the old version running instead of removing it or failing the whole reload (if possible).                                       | High       |
| 15   | **Filesystem Health Check**                   | **Observability**: Add a health check probe for filesystem roots to report status to the UI, not just logs.                                                                                                        | Low        |
| 16   | **Safe Symlink Traversal**                    | **Security**: Add configuration options to strictly control symlink traversal policies (allow/deny/internal-only).                                                                                                 | Medium     |
| 17   | **Multi-Model Advisor**                       | **Intelligence**: Orchestrate queries across multiple models (e.g. Ollama models) to synthesize insights.                                                                                                          | High       |
| 18   | **MCP Server Aggregator/Proxy**               | **Architecture**: A meta-server capability to discover, configure, and manage multiple downstream MCP servers dynamically.                                                                                         | High       |
| 19   | **Preset Service Gallery**                    | **UX**: A curated list of popular services (like `wttr.in`, `sqlite`, etc.) that can be added via CLI or UI with one click/command.                                                                                | Medium     |
| 20   | **Configuration Migration Tool**              | **DevX**: A CLI command to convert `claude_desktop_config.json` to `mcpany` config format.                                                                                                                         | Low        |
| 21   | **Auth Doctor**                               | **UX**: Diagnostic tool to verify that configured authentication methods match upstream expectations (e.g., detecting if a Bearer token is used where Basic Auth is expected).                                     | Medium     |
| 22   | **Dynamic Tool Pruning**                      | **Performance/Cost**: Feature to filter visible tools based on the current user's role or context to reduce LLM context window usage and costs.                                                                    | High       |
| 23   | **Config Schema Migration**                   | **Maintenance**: Automated tool to upgrade configuration files when the schema evolves (e.g. `v1alpha` to `v1`).                                                                                                   | Medium     |
| 25   | **Environment Variable Linter**               | **DevX**: A tool/check to scan config files and verify that all referenced environment variables (e.g. `${API_KEY}`) are actually set in the current shell/env.                                                    | Low        |
| 26   | **Linter Git Hook**                           | **DevX**: Provide a pre-commit hook script that automatically runs `mcpany lint` on staged configuration files to prevent committing insecure configs.                                                             | Low        |
| 27   | **Secret Rotation Helper**                    | **Ops**: A CLI tool to help rotate secrets by identifying which services are using a specific secret key/path and validating the new secret against the upstream.                                                  | Medium     |
| 28   | **Structured Logging for Config Errors**      | **DevX**: Output configuration errors in a structured JSON format to allow the UI or IDEs to pinpoint the exact location of the error.                                                                             | Low        |
| 29   | **Automatic Config Fixer**                    | **DevX**: An interactive CLI tool that detects common configuration errors (like legacy formats) and offers to fix them automatically.                                                                             | Medium     |
| 30   | **Windows Filesystem Locking Fix**            | **Compatibility**: Handle EPERM errors gracefully on Windows when renaming files, ensuring cross-platform stability.                                                                                               | Medium     |
| 31   | **Async Tool Loading**                        | **Reliability**: Ensure server waits for initial roots/tools to be loaded before accepting requests to prevent race conditions on startup.                                                                         | Medium     |
| 32   | **Interactive Config Generator**              | **DevX**: `mcpany init` wizard that asks questions and generates a valid `config.yaml` with best practices (secure defaults, comments).                                                                            | Low        |
| 33   | **Frontend Config Status Banner**             | **UX**: Add a visual indicator (banner/toast) in the UI Management Dashboard that queries the new `/doctor` configuration check and warns the user if the server is running with a stale or invalid configuration. | Medium     |
| 34   | **Configuration Diffing API**                 | **Observability**: An API endpoint to compare the currently active configuration with the previous version or the file on disk, helping users understand what changed during a reload.                             | Medium     |
| 35   | **Automatic WebSocket Reconnection Strategy** | **Resilience**: Allow users to configure retry backoff and max attempts for WS connections to handle transient network drops.                                                                                      | Medium     |
| 36   | **WebSocket Message Inspector**               | **Debugging**: A UI tool to capture and view raw WS frames (text/binary) for debugging protocol issues.                                                                                                            | Medium     |
| 37   | **RegEx Environment Variable Validation**     | **Security**: Allow validating the _format_ of environment variables using regex (e.g., ensuring an API key matches a pattern) in addition to existence checks.                                                    | Low        |
| 38   | **HTTP Upstream Env Validation**              | **Consistency**: Extend required environment variable validation to HTTP connections (e.g. for `http_address` or auth headers).                                                                                    | Low        |
| 39   | **Config Snapshot/Restore**                   | **Ops**: Ability to save current runtime configuration state to a file (snapshot) and restore it later, useful for backing up verified working configs.                                                            | Medium     |
| 40   | **Config Inheritance**                        | **DevX**: Allow `config.yaml` to extend/import other configuration files (e.g. `extends: base.yaml`) to reduce duplication across environments.                                                                    | High       |
| 42   | **OIDC Auto-Discovery**                       | **UX**: Automatically configure OIDC endpoints (auth URL, token URL, keys) by fetching the `/.well-known/openid-configuration` from the issuer URL, simplifying configuration.                                     | Medium     |
| 43   | **Doctor Auto-Fix**                           | **DevX**: Allow `mcpany doctor --fix` to automatically correct simple configuration errors (like typos or missing fields with defaults).                                                                           | High       |
| 44   | **Doctor Web Report**                         | **DevX**: Generate an HTML report from `mcpany doctor` for easier sharing and debugging.                                                                                                                           | Low        |
| 45   | **Upstream Latency Metrics**                  | **Observability**: Record the latency of the initial connectivity probe to help diagnose slow upstream services during startup.                                                                                    | Low        |
| 46   | **Tool Name Fuzzy Matching**                  | **UX**: Improve error messages for tool execution by suggesting similar tool names when a user makes a typo.                                                                                                       | Low        |
| 47   | **Interactive Doctor**                        | **UX**: A TUI (Text User Interface) for the doctor command that allows users to interactively retry failed checks or inspect details.                                                                              | Medium     |
| 48   | **Doctor Integration with Telemetry**         | **Observability**: Send doctor check results to telemetry (if enabled) to track fleet health during startup or health checks.                                                                                      | Low        |
| 41   | **Doctor Auto-Fix**                           | **DevX**: Allow `mcpany doctor --fix` to automatically correct simple configuration errors (like typos or missing fields with defaults).                                                                           | High       |
| 42   | **Doctor Web Report**                         | **DevX**: Generate an HTML report from `mcpany doctor` for easier sharing and debugging.                                                                                                                           | Low        |
| 41   | **Hard Failure Mode**                         | **Resilience**: A configuration option to strictly fail server startup (exit 1) if any service fails its connectivity probe, ensuring "fail-safe" deployments.                                                     | Low        |
| 42   | **Upstream Latency Metrics**                  | **Observability**: Record the latency of the initial connectivity probe to help diagnose slow upstream services during startup.                                                                                    | Low        |
| 41   | **Tool Name Fuzzy Matching**                  | **UX**: Improve error messages for tool execution by suggesting similar tool names when a user makes a typo.                                                                                                       | Low        |
| 42   | **Config Strict Mode**                        | **Ops**: Add a CLI flag to treat configuration warnings (e.g. deprecated fields) as errors to ensure clean configs.                                                                                                | Low        |
| 43   | **Context-Aware Suggestions**                 | **UX**: Refine the fuzzy matching logic to be context-aware, suggesting fields based on the specific message type (e.g., only suggest 'http_service' fields when inside an http_service block).                    | Medium     |
| 44   | **Interactive Config Validator**              | **DevX**: A CLI mode that walks through validation errors one by one and asks the user for the correct value interactively.                                                                                        | Medium     |
| 43   | **Config Schema Visualization**               | **UX**: A UI view to visualize the structure of the loaded configuration, highlighting inheritance or overrides.                                                                                                   | Low        |
| 44   | **Validator Plugin System**                   | **Extensibility**: Allow users to write custom validation rules (e.g. "service name must start with 'prod-'") using Rego or simple scripts.                                                                        | High       |
| 43   | **Tool Usage Analytics**                      | **Observability**: Track and visualize invocation counts, success rates, and latency per tool in the dashboard.                                                                                                    | Medium     |
| 44   | **Config Version History**                    | **Ops**: Keep a history of configuration changes and allow reverting to previous versions via UI.                                                                                                                  | High       |
| 43   | **Stdio Error Channel**                       | **DevX**: A dedicated side-channel or structured error output for stdio mode to communicate server status without interfering with JSON-RPC or stderr logging.                                                     | Medium     |
| 44   | **Log Redaction Rules**                       | **Security**: Configurable regex-based redaction for logs to prevent accidental leakage of sensitive data (API keys, PII) in stderr/files.                                                                         | Medium     |
| 45   | **Remote Schema Validation**                  | **Feature**: Allow validating schemas that use `$ref` to remote URLs by configuring a custom schema loader with HTTP support.                                                                                      | Medium     |
| 46   | **Schema Validation Caching**                 | **Performance**: Cache compiled schemas to avoid recompilation overhead during configuration reloads.                                                                                                              | Low        |
| 45   | **Config Validation Diff**                    | **Experience**: When a configuration reload fails, display a diff highlighting the changes that caused the error compared to the last known good configuration.                                                    | High       |
| 46   | **Health Webhooks**                           | **Ops**: Configure webhooks (Slack, Discord, PagerDuty) to be triggered when the system health status changes (e.g., from Healthy to Degraded).                                                                    | Medium     |
| 47   | **Metrics Persistence**                       | **Observability**: Store historical metrics (latency, error rates) in SQLite/Postgres for long-term trending and analysis.                                                                                         | High       |
| 48   | **Alerting System**                           | **Ops**: Configurable alerts based on metric thresholds (e.g. Error Rate > 5%) to trigger webhooks.                                                                                                                | Medium     |

## 3. Codebase Health

### Critical Areas (Refactoring Needed)

1.  **Rate Limiting Complexity (`server/pkg/middleware/ratelimit.go`)**
    - **Status**: ✅ Refactored (Strategy Pattern Implemented).
    - **Issue**: The current implementation tightly couples business logic with Redis/Memory backend logic.
    - **Risk**: High cognitive load, difficult to test, hard to add new backends (e.g., Postgres).
    - **Recommendation**: Extract a `RateLimitStrategy` interface.

2.  **Filesystem Upstream Monolith (`server/pkg/upstream/filesystem`)**
    - **Status**: ✅ Refactored (Provider Pattern Implemented).
    - **Issue**: The logic for Local, S3, and GCS providers is often intertwined or shares too much structural code in a way that violates SRP.
    - **Risk**: Changes to S3 support might break Local file support.
    - **Recommendation**: Strictly separate providers into distinct packages/structs implementing a common interface.

3.  **Webhook "Prototype" Status (`server/cmd/webhooks`)**
    - **Issue**: The code appears experimental and lacks the robustness of the main server (error handling, config validation).
    - **Risk**: If used in production, it could become a single point of failure.
    - **Recommendation**: Graduate this to `server/pkg/sidecar/webhooks` with full test coverage.

### Warning Areas

1.  **UI Component Duplication**: Some UI components in `ui/src/components` seem to have overlapping responsibilities (e.g., multiple "detail" views). A UI component audit is recommended.
2.  **Test Coverage gaps**: While core logic is tested, cloud providers (S3/GCS) and some new UI features lack comprehensive integration tests.

### Healthy Areas

- **Core Middleware Pipeline**: The middleware architecture is robust and extensible.
- **Protocol Implementation**: `server/pkg/mcpserver` cleanly separates protocol details from business logic.
- **Documentation**: The project has excellent documentation coverage for most features.
