# Server Roadmap

## 1. Updated Roadmap

### Status: Active Development

- **[Completed]** **Strict Service Validation**: Enhanced configuration validation to catch incomplete service definitions (e.g., gRPC without protos) at startup, preventing silent failures and infinite retry loops.
- **[Completed]** **Command Line Service Arguments**: Added support for `args` field in `command_line_service` configuration, fixing a schema gap that prevented passing arguments to commands.

## 2. Top 10 Recommended Features

These features represent the next logical steps for the product, focusing on Enterprise Readiness, Safety, and Developer Experience.

| Rank | Feature Name                                  | Why it matters                                                                                                                                                                                                     | Difficulty |
| :--- | :-------------------------------------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :--------- |
| 1    | **Tool Poisoning Mitigation**                 | **Security**: Implement integrity checks for tool definitions to prevent "Rug Pull" attacks where a tool definition changes maliciously after installation.                                                        | High       |
| 2    | **Interactive OAuth Handler**                 | **UX/Auth**: Solve the "copy-paste token" friction. Allow users to click "Login" in the UI to authenticate tools like GitHub/Google. (Backend Implemented, UI Pending)                                             | High       |
| 3    | **Interactive OAuth UI**                      | **UX**: Complete the UI implementation for the Interactive OAuth Handler to allow seamless user authentication flows.                                                                                              | Medium     |
| 4    | **Startup Health Report**                     | **UX**: Print a summary table to stdout after server startup, showing the status (Connected/Failed) of all registered upstream services to improve visibility of connectivity issues.                              | Low        |
| 5    | **Local LLM "One-Click" Connect**             | **Connectivity**: Auto-detect and connect to local inference servers (Ollama, LM Studio) to democratize AI access without cloud costs.                                                                             | Low        |
| 6    | **Tool "Dry Run" Mode**                       | **DevX**: Allow tools to define a "dry run" logic to validate inputs and permissions without executing side effects.                                                                                               | Medium     |
| 7    | **Team Configuration Sync**                   | **Collaboration**: Allow teams to synchronize `mcpany` configurations and secrets securely, ensuring consistent dev environments.                                                                                  | Medium     |
| 8    | **Smart Error Recovery**                      | **Resilience**: Use an internal LLM loop to analyze tool errors and automatically retry with corrected parameters (Self-Healing).                                                                                  | High       |
| 9    | **Canary Tool Deployment**                    | **Ops**: gradually roll out new tool versions to a subset of users or sessions to catch regressions before they impact everyone.                                                                                   | High       |
| 10   | **Compliance Reporting**                      | **Enterprise**: Automated generation of PDF/CSV reports from Audit Logs for SOC2/GDPR compliance reviews.                                                                                                          | Medium     |
| 11   | **Advanced Tiered Caching**                   | **Performance**: Implement a multi-layer cache (Memory -> Redis -> Disk) with configurable eviction policies to reduce upstream costs.                                                                             | Medium     |
| 12   | **Smart Retry Policies**                      | **Resilience**: Configurable exponential backoff and jitter for upstream connections to handle transient network failures gracefully.                                                                              | Medium     |
| 13   | **Service Dependency Graph**                  | **Observability**: Visualize dependencies between services (e.g. Service A calls Tool B in Service C) to detect circular dependencies or bottlenecks.                                                              | High       |
| 14   | **Schema Validation Playground**              | **DevX**: A dedicated UI in the dashboard to paste JSON/YAML and validate it against the server schema with real-time error highlighting.                                                                          | Low        |
| 15   | **Partial Reloads**                           | **Resilience**: When reloading config dynamically, if one service is invalid, keep the old version running instead of removing it or failing the whole reload (if possible).                                       | High       |
| 16   | **Filesystem Health Check**                   | **Observability**: Add a health check probe for filesystem roots to report status to the UI, not just logs.                                                                                                        | Low        |
| 17   | **Safe Symlink Traversal**                    | **Security**: Add configuration options to strictly control symlink traversal policies (allow/deny/internal-only).                                                                                                 | Medium     |
| 18   | **Multi-Model Advisor**                       | **Intelligence**: Orchestrate queries across multiple models (e.g. Ollama models) to synthesize insights.                                                                                                          | High       |
| 19   | **MCP Server Aggregator/Proxy**               | **Architecture**: A meta-server capability to discover, configure, and manage multiple downstream MCP servers dynamically.                                                                                         | High       |
| 20   | **Preset Service Gallery**                    | **UX**: A curated list of popular services (like `wttr.in`, `sqlite`, etc.) that can be added via CLI or UI with one click/command.                                                                                | Medium     |
| 21   | **Configuration Migration Tool**              | **DevX**: A CLI command to convert `claude_desktop_config.json` to `mcpany` config format.                                                                                                                         | Low        |
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
| 43   | **Doctor Auto-Fix**                           | **DevX**: Allow `mcpany doctor --fix` to automatically correct simple configuration errors (like typos or missing fields with defaults).                                                                           | High       |
| 44   | **Doctor Web Report**                         | **DevX**: Generate an HTML report from `mcpany doctor` for easier sharing and debugging.                                                                                                                           | Low        |
| 45   | **Upstream Latency Metrics**                  | **Observability**: Record the latency of the initial connectivity probe to help diagnose slow upstream services during startup.                                                                                    | Low        |
| 46   | **Tool Name Fuzzy Matching**                  | **UX**: Improve error messages for tool execution by suggesting similar tool names when a user makes a typo.                                                                                                       | Low        |
| 47   | **Interactive Doctor**                        | **UX**: A TUI (Text User Interface) for the doctor command that allows users to interactively retry failed checks or inspect details.                                                                              | Medium     |
| 48   | **Doctor Integration with Telemetry**         | **Observability**: Send doctor check results to telemetry (if enabled) to track fleet health during startup or health checks.                                                                                      | Low        |
| 49   | **Hard Failure Mode**                         | **Resilience**: A configuration option to strictly fail server startup (exit 1) if any service fails its connectivity probe, ensuring "fail-safe" deployments.                                                     | Low        |
| 50   | **Config Strict Mode**                        | **Ops**: Add a CLI flag to treat configuration warnings (e.g. deprecated fields) as errors to ensure clean configs.                                                                                                | Low        |
| 51   | **Duplicate Tool Detection**                  | **Safety**: Detect if two services expose tools with the same name (before sanitization) and warn about potential conflicts or shadowing.                                                                          | Low        |
| 52   | **Tool Execution Simulation**                 | **DevX**: A UI feature to "mock" tool execution with predefined outputs for testing client integrations without calling real upstreams.                                                                            | Medium     |
| 53   | **Context-Aware Suggestions**                 | **UX**: Refine the fuzzy matching logic to be context-aware, suggesting fields based on the specific message type (e.g., only suggest 'http_service' fields when inside an http_service block).                    | Medium     |
| 54   | **Interactive Config Validator**              | **DevX**: A CLI mode that walks through validation errors one by one and asks the user for the correct value interactively.                                                                                        | Medium     |
| 55   | **Config Schema Visualization**               | **UX**: A UI view to visualize the structure of the loaded configuration, highlighting inheritance or overrides.                                                                                                   | Low        |
| 56   | **Validator Plugin System**                   | **Extensibility**: Allow users to write custom validation rules (e.g. "service name must start with 'prod-'") using Rego or simple scripts.                                                                        | High       |
| 57   | **Tool Usage Analytics**                      | **Observability**: Track and visualize invocation counts, success rates, and latency per tool in the dashboard.                                                                                                    | Medium     |
| 58   | **Config Version History**                    | **Ops**: Keep a history of configuration changes and allow reverting to previous versions via UI.                                                                                                                  | High       |
| 59   | **Stdio Error Channel**                       | **DevX**: A dedicated side-channel or structured error output for stdio mode to communicate server status without interfering with JSON-RPC or stderr logging.                                                     | Medium     |
| 60   | **Log Redaction Rules**                       | **Security**: Configurable regex-based redaction for logs to prevent accidental leakage of sensitive data (API keys, PII) in stderr/files.                                                                         | Medium     |
| 61   | **Remote Schema Validation**                  | **Feature**: Allow validating schemas that use `$ref` to remote URLs by configuring a custom schema loader with HTTP support.                                                                                      | Medium     |
| 62   | **Schema Validation Caching**                 | **Performance**: Cache compiled schemas to avoid recompilation overhead during configuration reloads.                                                                                                              | Low        |
| 63   | **Config Validation Diff**                    | **Experience**: When a configuration reload fails, display a diff highlighting the changes that caused the error compared to the last known good configuration.                                                    | High       |
| 64   | **Health Webhooks**                           | **Ops**: Configure webhooks (Slack, Discord, PagerDuty) to be triggered when the system health status changes (e.g., from Healthy to Degraded).                                                                    | Medium     |
| 65   | **Service Retry Policy**                      | **Resilience**: Automatically retry connecting to failed services with exponential backoff.                                                                                                                        | Medium     |
| 66   | **Config Reload Status API**                  | **DevX**: Expose the status of the last configuration reload attempt via API to help debug silent reload failures.                                                                                                 | Low        |
| 67   | **Dynamic Profile Switching**                 | **UX**: Allow users to switch active profiles dynamically via API without restarting the server.                                                                                                                   | Medium     |
| 68   | **Config Schema Versioning**                  | **Maintenance**: Introduce `apiVersion` field in `config.yaml` to support breaking changes in configuration schema gracefully.                                                                                     | High       |
| 69   | **Connection Draining**                       | **Availability**: Utilize active connection tracking (from System Health Dashboard) to implement graceful shutdown that waits for connections to finish before exiting.                                       | Medium     |
| 70   | **Secure Defaults Enforcer**                  | **Security**: Automated "Fix-it" suggestions or enforcement of secure defaults based on security warnings visualized in the Health Dashboard.                                                                  | Medium     |

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
