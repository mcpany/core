# Evolution: [2026-10-28] Policy Firewall Engine (CEL-based)

## Introduction

As agent ecosystems become more complex, relying solely on basic authentication or coarse-grained capability limits is insufficient. We need to evaluate the context of individual tool calls (parameters, intent, caller identity) against a centralized "Policy Firewall."

This document proposes the integration of the **Policy Firewall Engine**, a critical security service that implements Common Expression Language (CEL) based hooking for tool calls.

## Core Logic

The Policy Firewall sits transparently as a Middleware between the incoming MCP Gateway (HTTP/WebSocket/Stdio) and the actual tool execution.

1. **Interception**: Every `CallToolRequest` is intercepted.
2. **Context Enrichment**: The firewall gathers context, including the user, tool name, tool parameters, and any active session tokens.
3. **Policy Evaluation**: The firewall evaluates the context against configured CEL (Common Expression Language) policies.
4. **Enforcement**: If the evaluation returns `true`, the call proceeds. If `false`, the firewall blocks the request and returns an explicit access denied error to the caller, preventing upstream execution.

## Flow Diagram

```mermaid
graph TD
    A[Client Application (LLM/Agent)] --> B(MCP Any Adapter / Gateway)
    B --> C{Policy Firewall Middleware}

    C -->|Extract Context & Args| D[CEL Evaluation Engine]
    D --> E{Policy Pass?}

    E -->|No| F[Return PermissionDenied Error]
    F --> A

    E -->|Yes| G[Tool/Service Adapter]
    G --> H((Upstream Capability))
    H --> G
    G --> B
    B --> A
```

## Implementation Details

We will implement this as a new middleware `policy_firewall.go` in `server/pkg/middleware`.

We will leverage `github.com/google/cel-go` (Common Expression Language) to parse and evaluate the security policies. The server will dynamically load these policies from its configuration and compile them at startup for sub-millisecond evaluation latency.

### Example Policy
```cel
// Ensure the user is only reading files from /tmp
request.params.path.startsWith("/tmp/") && request.tool == "fs_read"
```

## Security Posture
This elevates our "Zero Trust" model by allowing administrators to author machine-checkable security contracts that enforce fine-grained access without writing custom Go code.
