# Server Roadmap

## 1. Updated Roadmap

### Status: Active Development

### Implemented Features (Recently Completed)

- [x] **Agent Debugger & Inspector**: Middleware for traffic replay and inspection. [Docs](server/docs/features/debugger.md)
- [x] **Context Optimizer**: Middleware to prevent context bloat. [Docs](server/docs/features/context_optimizer.md)
- [x] **Diagnostic "Doctor" API**: `mcpctl` validation and health checks. [Docs](server/docs/features/mcpctl.md)
- [x] **SSO Integration**: OIDC/SAML support. [Docs](server/docs/features/sso.md)
- [x] **Audit Log Export**: Native Splunk and Datadog integration. [Docs](server/docs/features/audit_logging.md)
- [x] **Cost Attribution**: Token-based cost estimation and metrics. [Docs](server/docs/features/rate-limiting/README.md)
- [x] **Universal Connector Runtime**: Sidecar for stdio tools. [Docs](server/docs/features/connector_runtime.md)
- [x] **WASM Plugin System**: Runtime for sandboxed plugins. [Docs](server/docs/features/wasm.md)
- [x] **Hot Reload**: Dynamic configuration reloading. [Docs](server/docs/features/hot_reload.md)
- [x] **SQL Upstream**: Expose SQL databases as tools. [Docs](server/docs/features/sql_upstream.md)
- [x] **Webhooks Sidecar**: Context optimization and offloading. [Docs](server/docs/features/webhooks/sidecar.md)
- [x] **Dynamic Tool Registration**: Auto-discovery from OpenAPI/gRPC/GraphQL. [Docs](server/docs/features/dynamic_registration.md)
- [x] **Helm Chart Official Support**: K8s deployment charts. [Docs](server/docs/features/helm.md)
- [x] **Message Bus**: NATS/Kafka integration for events. [Docs](server/docs/features/message_bus.md)
- [x] **Structured Output Transformation**: JQ/JSONPath response shaping. [Docs](server/docs/features/transformation.md)

## 2. Top 10 Recommended Features

| Rank | Feature Name | Why it matters | Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **Traffic Inspector & Replay** | **Debugging**: The "Network Tab" for AI. View raw JSON-RPC messages, latency, and errors. "Replay" a tool call with modified parameters to pinpoint why a tool failed or why the agent hallucinated arguments. | Medium |
| 2 | **Context Lens & Optimizer** | **Optimization**: Solve "Context Limit Exceeded". Visual breakdown of token usage (System Prompt vs Tool Definitions vs Conversation). Strategies to auto-summarize or "forget" older turns to keep agents running. | High |
| 3 | **Prompt Studio & Evals** | **Stability**: When "tool invocation is unexpected", it's often the tool description. Edit descriptions in UI and run "Evals" (regression tests) to verify the model picks the right tool for a given query. | High |
| 4 | **Mocking & Fixtures** | **Testing**: Define static responses for tools ("Happy Path", "Error 500"). Test how your agent handles edge cases without hitting real production APIs. | Medium |
| 5 | **Session Recorder** | **Forensics**: Record full agent-server interaction sessions. Share a link to a "Broken Session" with a teammate so they can replay it and analyze the exact state that caused a failure. | Medium |
| 6 | **Unified Marketplace & Installer** | **Discovery**: "npm install" for MCP servers. Rapidly bootstrap a dev environment with standard tools (Postgres, GitHub, Slack) to test agent capabilities. | Medium |
| 7 | **Universal Client Export** | **Usability**: One-click generation of configuration files/links for Claude Desktop, Cursor, and VS Code. Removes the friction of "how do I get this URL into my editor?". | Low |
| 8 | **Interactive OAuth Handler** | **Auth**: Solves the pain of manual token management. Click "Login" in the UI to authenticate tools like GitHub/Google, handling refresh tokens automatically. | High |
| 9 | **Sandboxed Runtime (Docker/WASM)** | **Safety**: Run untrusted community tools safely. Essential for developers testing out new MCP servers from the marketplace without risking their local machine. | High |
| 10 | **Team Configuration Sync** | **Collaboration**: Share common dev tools and secrets (encrypted) with your team. Ensure everyone is testing against the same version of the backend services. | Medium |
| 11 | **Human-in-the-Loop Approval UI** | **Safety**: Intercept dangerous tool calls (e.g. `delete_db`) and require manual clicking "Approve" in the UI before execution proceeds. Essential for dev/prod safety. | Low |
| 12 | **Local LLM "One-Click" Connect** | **Connectivity**: First-class support for Ollama, LM Studio, and LocalAI. Auto-detect running local models and expose them as standardized agents/clients. | Low |
| 13 | **Tool "Dry Run" Mode** | **Debugging**: Execute tool logic in a safe mode (if supported) or validate input types without performing side effects. Great for testing complex tool arguments. | Medium |
| 14 | **Cost & Latency Budgeting** | **Ops**: Set alerts for session costs ("Warn if > $0.50") or tool latency ("Warn if > 2s"). Helps developers optimize their agent's performance and burn rate. | Medium |
| 15 | **Smart Error Recovery** | **Resilience**: Auto-suggest fixes for common JSON-RPC errors or API failures. If a tool fails with "Rate Limit", auto-retry with exponential backoff transparently. | High |
| 16 | **Documentation Generator** | **Documentation**: Auto-generate beautiful Markdown/HTML documentation for all registered tools, arguments, and return types. Keep your agent system docs always up-to-date. | Low |
| 17 | **API Key Vault** | **Security**: Secure local storage for all upstream keys, decoupled from config files. Prevents accidental git commits of secrets. | Medium |
| 18 | **Server Plugin System** | **Extensibility**: Allow developers to write middleware/plugins for MCP Any itself (e.g. "Log to Datadog", "Custom Auth") using a simple WASM or Go plugin API. | High |
| 19 | **Global "Tool Search"** | **Usability**: Cmd+K palette to search across all available tools, docs, and resources. Quickly find "Which tool allows me to search Jira?". | Low |
| 20 | **Conversation Branching** | **Testing**: In the Playground/Recorder, fork a conversation at any point to test alternate prompts or tool outputs ("What if the tool returned X instead of Y?"). | High |
| 21 | **Schema Validation Labs** | **Development**: A sandbox to test JSON Schemas against sample inputs. Verify that your robust tool definitions actually match the expected payloads. | Low |
| 22 | **Vector DB Inspector** | **Observability**: If using Compass/Semantic Caching, visualize the vector store contents. See which queries are "close" to each other in embedding space. | High |
| 23 | **Rate Limit Dashboard** | **Observability**: Visualize token buckets and quota usage for all connected services. Understand why your agent is getting throttled. | Low |
| 24 | **Multi-Tenant Profile Simulator** | **IAM**: "View As" User X. Debug permission issues by simulating how the server behaves for a specific user profile or role. | Medium |
| 25 | **No-Code Tool Builder** | **Creators**: Create simple tools (Shell command -> Tool, HTTP Request -> Tool) via a UI wizard without writing a single line of Go/Python code. | Medium |

## 3. Codebase Health

### Critical Areas

- **Rate Limiting Complexity**: The current implementation in `server/pkg/middleware/ratelimit.go` tightly couples in-memory logic with Redis commands. This structure makes unit testing difficult and prevents easy addition of new backends.
- **Filesystem Provider Monolith**: `server/pkg/upstream/filesystem/upstream.go` currently implements logic for Local, S3, and GCS backends in a single file. This violation of the Single Responsibility Principle makes the code brittle and hard to maintain.
- **Test Coverage for Cloud Providers**: There is a significant gap in testing for S3 and GCS integrations. We rely on mocks or manual testing; introducing integration tests using local emulators (like MinIO or fake-gcs-server) is critical.
- **Webhooks "Test" Code**: The `server/cmd/webhooks` directory contains code that looks like a prototype. If this is intended for production use as a sidecar, it needs to be formalized with proper configuration, logging, and error handling.
- **SDK Consolidation**: As mentioned in the recommendations, `server/pkg/client` should be moved to its own repository to allow clients to import the SDK without pulling in the entire server dependency tree.

### Recommendations

1.  **Refactor Rate Limiting**: Introduce a `RateLimiterStrategy` interface and implement distinct `LocalStrategy` and `RedisStrategy` structs.
2.  **Refactor Filesystem Upstream**: Adopt a Factory pattern to instantiate specific implementations for Local, S3, and GCS, separating their logic into distinct files.
3.  **Formalize Webhook Server**: graduate `server/cmd/webhooks` from a prototype to a fully supported sidecar component.
4.  **Standardize Configuration**: Audit configuration structures across modules to ensure consistent naming conventions and validation logic.
