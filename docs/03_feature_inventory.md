# Feature Inventory: MCP Any

This is the living masterlist of features for MCP Any, categorized by strategic priority.

## P0: Critical Infrastructure (Foundation)
- **Multi-Protocol Adapters**: Support for HTTP/REST, gRPC, and Command-line tools.
- **Service Registry**: Dynamic loading and lifecycle management of upstream services.
- **Audit Logging**: Comprehensive recording of all tool calls and responses.
- **DLP (Data Loss Prevention)**: Sensitive data redaction in logs and responses.

## P1: Enterprise & Safety (Evolving)
- **Policy Engine (Rego/CEL)**: Fine-grained access control for tool execution.
- **Human-in-the-Loop (HITL)**: Middleware for manual approval of sensitive actions.
- **Context Optimizer**: Token management and response truncation.
- **Service Health Dashboard**: Real-time monitoring and diagnostics.

## Proposed Features: [2026-02-23]

### [P0] Recursive Context Propagation
- **Problem**: Subagents lose parent context and authentication state.
- **Solution**: A protocol extension allowing parent agents to delegate a subset of their capabilities and context to children via standard MCP headers.

### [P0] Zero Trust Tool Sandboxing (WASM/Docker)
- **Problem**: Command-line upstreams can be abused for host-level attacks.
- **Solution**: Execute all CLI and script-based tools within isolated WASM runtimes or ephemeral Docker containers.

### [P1] Shared Key-Value "Blackboard" Store
- **Problem**: Swarms lack a common memory space for sharing intermediate results.
- **Solution**: An embedded SQLite-backed tool that agents can use to `get`/`set`/`list` shared state variables within a session.

### [P0] Metadata Injection Firewall
- **Problem**: Agents can be tricked into calling the wrong tool or using wrong schemas via prompt injection.
- **Solution**: Middleware that validates tool call metadata against strict, pre-defined schemas and integrity hashes (Tool Poisoning Mitigation).
