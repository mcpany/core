# Server Roadmap

## 1. Completed Features

- **Interactive Doctor Resilience**
  - **Description**: Enhanced `doctor` command to gracefully handle missing environment variables in configuration files, allowing it to report specific missing variables and proceed with other checks instead of aborting.
- **Pre-flight Command Validation**
  - **Description**: Validates that the executable exists for command-based services before attempting to run it, providing a clear error message if it's missing.
- **Actionable Configuration Errors**
  - **Description**: Improved configuration loading and validation to provide "Actionable Errors" with specific "Fix" suggestions for common issues like missing environment variables, missing files, and invalid paths.
- **Environment Variable Fuzzy Matching**
  - **Description**: Enhances "Actionable Errors" by suggesting similar environment variables when a configured variable is missing, helping users catch typos (e.g., "Did you mean 'API_KEY'?").
- **RegEx Environment Variable Validation**
  - **Description**: Validating the format of environment variables using regex (e.g., ensuring an API key matches a pattern) in addition to existence checks.
- **Async Tool Loading**
  - **Description**: Ensure server waits for initial roots/tools to be loaded before accepting requests to prevent race conditions on startup.
- **Preset Service Gallery**
  - **Description**: A curated list of popular services (like `wttr.in`, `sqlite`, etc.) that can be added via CLI or UI. Implemented via example configurations in `server/examples/popular_services`.
- **HTTP Upstream Env Validation**
  - **Description**: Extend required environment variable validation to HTTP connections (e.g. for `http_address` or auth headers).
- **Tool Poisoning Mitigation**
  - **Description**: Integrity checks for tool definitions to prevent "Rug Pull" attacks. Implemented via SHA256 hashing of tool definitions.
- **Local LLM "One-Click" Connect**
  - **Description**: Auto-detection and template-based connection to local inference servers. Supports Ollama via `/api/tags` discovery.
- **Tool "Dry Run" Mode**
  - **Description**: Allows tools to validate inputs and return a preview of side effects without executing them. Supported in the common tool execution lifecycle.
- **Smart Retry Policies**
  - **Description**: Configurable exponential backoff and jitter for upstream connections, integrated with circuit breakers.
- **Service Dependency Graph**
  - **Description**: Visual topology of the MCP ecosystem, visualizing clients, services, tools, and their relationships with real-time metrics.
- **Runtime Health Visibility**
  - **Description**: Exposed real-time service health status (`last_error`) and tool counts in the API, enabling the UI to show error badges for failing upstreams instantly.
- **Port Conflict Hints**
  - **Description**: Detects "Address already in use" errors during server startup and suggests using `--json-rpc-port` or `--grpc-port` flags to resolve the conflict.
- **Whitespace URL Validation**
  - **Description**: Detects and warns about hidden whitespace in URL configurations (HTTP, WebSocket, OpenAPI, etc.) which often occurs when copying from external sources, providing actionable fixes.
- **gRPC Health Checks**
  - **Description**: Implements `CheckHealth` for gRPC upstreams using the standard gRPC Health Checking Protocol to detect service availability.
- **Context Optimizer Middleware**
  - **Description**: Automatically truncates large text outputs in JSON responses to prevent "Context Bloat" and reduce token usage.
- **Smart Error Recovery**
  - **Description**: Uses an internal LLM loop to analyze tool errors and automatically retry with corrected parameters (Self-Healing).
- **Service Health History**
  - **Description**: Stores historical health check results (in-memory) to visualize availability trends in the dashboard.

## 2. Updated Roadmap

### Status: Active Development

## 2. Top 10 Recommended Features

These features represent the next logical steps for the product, focusing on Enterprise Readiness, Safety, and Developer Experience.

| Rank | Feature Name                | Why it matters                                                                                                                         | Difficulty |
| :--- | :-------------------------- | :------------------------------------------------------------------------------------------------------------------------------------- | :--------- |
| 1    | **Team Configuration Sync** | **Collaboration**: Allow teams to synchronize `mcpany` configurations and secrets securely, ensuring consistent dev environments.      | Medium     |
| 2    | **Persistent Health History**| **Observability**: Persist health check history to database (SQLite/Postgres) to retain trends across server restarts.                 | Medium     |
| 3    | **Tool Execution Timeline** | **Debugging**: A visual waterfall chart of tool execution stages (hooks, middleware, upstream call) to debug latency bottlenecks.      | High       |
| 3    | **Canary Tool Deployment**  | **Ops**: gradually roll out new tool versions to a subset of users or sessions to catch regressions before they impact everyone.       | High       |
| 4    | **Compliance Reporting**    | **Enterprise**: Automated generation of PDF/CSV reports from Audit Logs for SOC2/GDPR compliance reviews.                              | Medium     |
| 5    | **Advanced Tiered Caching** | **Performance**: Implement a multi-layer cache (Memory -> Redis -> Disk) with configurable eviction policies to reduce upstream costs. | Medium     |

| 14 | **Partial Reloads** | **Resilience**: When reloading config dynamically, if one service is invalid, keep the old version running instead of removing it or failing the whole reload (if possible). | High |
| 15 | **Filesystem Health Check** | **Observability**: Add a health check probe for filesystem roots to report status to the UI, not just logs. | Low |
| 16 | **Safe Symlink Traversal** | **Security**: Add configuration options to strictly control symlink traversal policies (allow/deny/internal-only). | Medium |
| 17 | **Multi-Model Advisor** | **Intelligence**: Orchestrate queries across multiple models (e.g. Ollama models) to synthesize insights. | High |
| 18 | **MCP Server Aggregator/Proxy** | **Architecture**: A meta-server capability to discover, configure, and manage multiple downstream MCP servers dynamically. | High |
| 20 | **Configuration Migration Tool** | **DevX**: A CLI command to convert `claude_desktop_config.json` to `mcpany` config format. | Low |
| 22 | **Dynamic Tool Pruning** | **Performance/Cost**: Feature to filter visible tools based on the current user's role or context to reduce LLM context window usage and costs. | High |
| 23 | **Config Schema Migration** | **Maintenance**: Automated tool to upgrade configuration files when the schema evolves (e.g. `v1alpha` to `v1`). | Medium |

| 26 | **Linter Git Hook** | **DevX**: Provide a pre-commit hook script that automatically runs `mcpany lint` on staged configuration files to prevent committing insecure configs. | Low |
| 27 | **Secret Rotation Helper** | **Ops**: A CLI tool to help rotate secrets by identifying which services are using a specific secret key/path and validating the new secret against the upstream. | Medium |
| 28 | **Structured Logging for Config Errors** | **DevX**: Output configuration errors in a structured JSON format to allow the UI or IDEs to pinpoint the exact location of the error. | Low |
| 33 | **Interactive Config Validator** | **DevX**: A CLI mode that walks through validation errors one by one and asks the user for the correct value interactively. | Medium |
| 34 | **Secret Validation Pre-flight** | **Security**: Validate all secrets (files, env vars, remote) before starting the server to ensure all credentials are accessible. | Low |
| 29 | **Automatic Config Fixer** | **DevX**: An interactive CLI tool that detects common configuration errors (like legacy formats) and offers to fix them automatically. | Medium |
| 30 | **Windows Filesystem Locking Fix** | **Compatibility**: Handle EPERM errors gracefully on Windows when renaming files, ensuring cross-platform stability. | Medium |
| 32 | **Interactive Config Generator** | **DevX**: `mcpany init` wizard that asks questions and generates a valid `config.yaml` with best practices (secure defaults, comments). | Low |

| 34 | **Configuration Diffing API** | **Observability**: An API endpoint to compare the currently active configuration with the previous version or the file on disk, helping users understand what changed during a reload. | Medium |
| 35 | **Automatic WebSocket Reconnection Strategy** | **Resilience**: Allow users to configure retry backoff and max attempts for WS connections to handle transient network drops. | Medium |
| 36 | **WebSocket Message Inspector** | **Debugging**: A UI tool to capture and view raw WS frames (text/binary) for debugging protocol issues. | Medium |
| 37 | **Config Diffing UI** | **UX**: A visual configuration comparison tool in the management dashboard. | Medium |
| 39 | **Config Snapshot/Restore** | **Ops**: Ability to save current runtime configuration state to a file (snapshot) and restore it later, useful for backing up verified working configs. | Medium |
| 40 | **Config Inheritance** | **DevX**: Allow `config.yaml` to extend/import other configuration files (e.g. `extends: base.yaml`) to reduce duplication across environments. | High |
| 43 | **Doctor Auto-Fix** | **DevX**: Allow `mcpany doctor --fix` to automatically correct simple configuration errors (like typos or missing fields with defaults). | High |
| 44 | **Doctor Web Report** | **DevX**: Generate an HTML report from `mcpany doctor` for easier sharing and debugging. | Low |
| 45 | **Upstream Latency Metrics** | **Observability**: Record the latency of the initial connectivity probe to help diagnose slow upstream services during startup. | Low |

| 47 | **Interactive Doctor** | **UX**: A TUI (Text User Interface) for the doctor command that allows users to interactively retry failed checks or inspect details. | Medium |
| 48 | **Doctor Integration with Telemetry** | **Observability**: Send doctor check results to telemetry (if enabled) to track fleet health during startup or health checks. | Low |
| 41 | **Hard Failure Mode** | **Resilience**: A configuration option to strictly fail server startup (exit 1) if any service fails its connectivity probe, ensuring "fail-safe" deployments. | Low |
| 41 | **Tool Name Fuzzy Matching** | **UX**: Improve error messages for tool execution by suggesting similar tool names when a user makes a typo. | Low |
| 42 | **Config Strict Mode** | **Ops**: Add a CLI flag to treat configuration warnings (e.g. deprecated fields) as errors to ensure clean configs. | Low |
| 43 | **Context-Aware Suggestions** | **UX**: Refine the fuzzy matching logic to be context-aware, suggesting fields based on the specific message type (e.g., only suggest 'http_service' fields when inside an http_service block). | Medium |
| 43 | **Config Schema Visualization** | **UX**: A UI view to visualize the structure of the loaded configuration, highlighting inheritance or overrides. | Low |
| 44 | **Validator Plugin System** | **Extensibility**: Allow users to write custom validation rules (e.g. "service name must start with 'prod-'") using Rego or simple scripts. | High |

| 44 | **Config Version History** | **Ops**: Keep a history of configuration changes and allow reverting to previous versions via UI. | High |
| 43 | **Stdio Error Channel** | **DevX**: A dedicated side-channel or structured error output for stdio mode to communicate server status without interfering with JSON-RPC or stderr logging. | Medium |
| 44 | **Log Redaction Rules** | **Security**: Configurable regex-based redaction for logs to prevent accidental leakage of sensitive data (API keys, PII) in stderr/files. | Medium |
| 45 | **Remote Schema Validation** | **Feature**: Allow validating schemas that use `$ref` to remote URLs by configuring a custom schema loader with HTTP support. | Medium |
| 46 | **Schema Validation Caching** | **Performance**: Cache compiled schemas to avoid recompilation overhead during configuration reloads. | Low |
| 46 | **Health Webhooks** | **Ops**: Configure webhooks (Slack, Discord, PagerDuty) to be triggered when the system health status changes (e.g., from Healthy to Degraded). | Medium |
| 47 | **Config Validation Dry Run** | **DevX**: Allow users to upload a config to a "dry run" endpoint to see if it would pass validation without applying it. | Medium |
| 47 | **Metrics Persistence** | **Observability**: Store historical metrics (latency, error rates) in SQLite/Postgres for long-term trending and analysis. | High |

| 50 | **Duplicate Tool Detection** | **Safety**: Detect if two services expose tools with the same name (before sanitization) and warn about potential conflicts or shadowing. | Low |
| 51 | **Tool Execution Simulation** | **DevX**: A UI feature to "mock" tool execution with predefined outputs for testing client integrations without calling real upstreams. | Medium |
| 62 | **Config Validation Diff** | **Experience**: When a configuration reload fails, display a diff highlighting the changes that caused the error compared to the last known good configuration. | High |
| 64 | **Service Retry Policy** | **Resilience**: Automatically retry connecting to failed services with exponential backoff. | Medium |
| 65 | **Config Reload Status API** | **DevX**: Expose the status of the last configuration reload attempt via API to help debug silent reload failures. | Low |
| 66 | **Dynamic Profile Switching** | **UX**: Allow users to switch active profiles dynamically via API without restarting the server. | Medium |
| 67 | **Config Schema Versioning** | **Maintenance**: Introduce `apiVersion` field in `config.yaml` to support breaking changes in configuration schema gracefully. | High |
| 68 | **Connection Draining** | **Availability**: Utilize active connection tracking (from System Health Dashboard) to implement graceful shutdown that waits for connections to finish before exiting. | Medium |
| 69 | **Secure Defaults Enforcer** | **Security**: Automated "Fix-it" suggestions or enforcement of secure defaults based on security warnings visualized in the Health Dashboard. | Medium |
| 70 | **Interactive Doctor Fixer** | **DevX**: Extend the doctor command to automatically fix common configuration issues (e.g. creating missing files, updating schema versions). | High |
| 71 | **Config Validation Webhook** | **Ops**: A pre-commit or CI webhook that runs `mcpany config validate` on changed files to prevent bad config from being merged. | Medium |
| 70 | **Tool Activity Feed** | **UX**: A dedicated UI component to show the tool execution history (structured), separate from raw logs, providing clear visibility into tool usage and performance. | Medium |
| 70 | **User Preference Storage** | **UX/Backend**: API to store and retrieve user-specific UI preferences (layout, theme, etc.) in the database. | Low |
| 71 | **Top Tools API Extensions** | **Observability**: Enhance the top tools API to support time ranges (last 1h, 24h) using historical metrics if available. | Medium |
| 72 | **Config Hot-Reload Validation** | **Resilience**: Validate configuration changes before applying them during a hot-reload to prevent breaking the running server with a bad config. | High |
| 76 | **Config Auto-Format API** | **DevX**: API endpoint to format uploaded config (JSON/YAML) according to standard style. | Low |
| 77 | **Service Dependency Alerts** | **Ops**: Alert if a service dependency (e.g. database) is down for more than X minutes. | Medium |
| 78 | **Tool Execution Timeout Configuration** | **Resilience**: Allow configuring timeouts per-tool or per-service to prevent hanging tools. | Medium |
| 79 | **Secret Versioning Support** | **Security**: Allow referencing specific versions of secrets in configuration (e.g. `secret:my-secret:v1`). | Medium |
| 73 | **Docker Secret Native Support** | **Ops**: Native support for reading Docker secrets (files in `/run/secrets`) and substituting them into configuration without needing environment variable mapping. | Medium |
| 75 | **Health Check Flap Damping** | **Resilience**: Configurable retries and thresholds for health checks to prevent services from flapping between Healthy and Unhealthy states due to transient network issues. | Medium |
| 74 | **Environment Variable Wizard** | **DevX**: A UI helper to identify used environment variables in a config and prompt the user to fill them if missing during startup/testing. | Low |
| 75 | **Global Redaction Policy** | **Security**: Centralized configuration to define patterns (regex) for redaction across all logs, error messages, and traces. | Medium |
| 74 | **Tool Search & Filter API** | **UX/DevX**: A dedicated API to search tools by name/description/tags with fuzzy matching, to power UI search bars and "did you mean" hints in the frontend. | Low |
| 75 | **Tool Execution Trace ID** | **Observability**: Propagate a trace ID through the tool execution flow (hooks, middleware, execution) to aid in debugging complex tool chains. | Medium |
| 76 | **Auto-Discovery Status API** | **Observability**: Expose the status of auto-discovery providers (Last run, Error, Success) via API to the UI, so users know why local tools (like Ollama) are missing. | Low |
| 77 | **Configurable Discovery Providers** | **Configuration**: Allow defining discovery providers in `config.yaml` (e.g. `discovery: { ollama: { url: "http://host:11434" } }`) instead of hardcoded defaults. | Medium |
| 76 | **Config Schema Validation with Line Numbers**| **DevX**: Extend line number reporting to schema validation errors (e.g., missing required fields, type mismatches) by mapping schema errors back to YAML AST nodes. | Medium |
| 77 | **YAML AST Caching** | **Performance**: Cache parsed YAML ASTs to avoid re-parsing for multiple error lookups during configuration loading. | Low |
| 78 | **Strict Config Validation on Reload** | **Resilience**: Extend strict configuration validation to dynamic reloads, ensuring that invalid configurations are rejected with a detailed diff and error report before any changes are applied. | High |
| 79 | **Conflict-Free Port Allocation** | **DevX**: Add a `--random-port` flag that automatically finds an available port if the default is taken, useful for automated testing. | Low |
| 80 | **Secret Format Validation for Known Services** | **Security**: Heuristic validation for common secret formats (e.g. OpenAI `sk-`, GitHub `ghp_`) to catch invalid keys early. | Low |
| 81 | **Interactive Env Var Fixer** | **DevX**: A CLI tool that detects validation errors like hidden whitespace and offers to interactively fix the .env file. | Medium |
| 78 | **Upstream Connectivity Debugger** | **DevX**: CLI tool to debug connectivity issues with upstreams (like `curl` but with MCP auth/headers injected from config). | Medium |
| 79 | **Configuration Template Generator** | **DevX**: CLI command to generate a scaffold `config.yaml` based on a list of desired services (e.g. `mcpany config init --services github,postgres`). | Low |

## 3. Codebase Health

### Critical Areas (Refactoring Needed)

*None at this time.*

### Warning Areas

1.  **UI Component Duplication**: Some UI components in `ui/src/components` seem to have overlapping responsibilities (e.g., multiple "detail" views). A UI component audit is recommended.
2.  **Test Coverage gaps**: While core logic is tested, cloud providers (S3/GCS) and some new UI features lack comprehensive integration tests.

### Healthy Areas

- **Core Middleware Pipeline**: The middleware architecture is robust and extensible.
- **Protocol Implementation**: `server/pkg/mcpserver` cleanly separates protocol details from business logic.
- **Documentation**: The project has excellent documentation coverage for most features.
