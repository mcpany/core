# Feature Inventory

## Current Priority Backlog

| Feature | Category | Priority | Status |
| :--- | :--- | :--- | :--- |
| **Policy Firewall Engine** | Security | P0 | Draft Design |
| **HITL Middleware** | Safety | P0 | In Roadmap |
| **Recursive Context Protocol** | Comms | P1 | Proposed |
| **Shared Key-Value Store** | State | P1 | In Roadmap |
| **OpenClaw Skill Adapter** | Adapter | P1 | Proposed |
| **Zero Trust OpenClaw Bridge**| Security | P0 | New (2026-02-22) |

## Newly Proposed Features: [2026-02-22]

### 1. Zero Trust OpenClaw Bridge
*   **Description**: A specialized adapter that wraps OpenClaw skills in a secure sandbox, enforcing Rego policies for any shell or filesystem operation triggered by the agent.
*   **Impact**: Enables institutional use of OpenClaw by mitigating its primary security vulnerabilities.

### 2. Recursive Context Protocol (RCP)
*   **Description**: Implementation of standardized headers (`X-MCP-Context-ID`, `X-MCP-Parent-Agent`) to allow subagents to automatically inherit parent context, authentication, and policy constraints.
*   **Impact**: Solves the "Context Bloat" and inheritance pain points in multi-agent swarms.
