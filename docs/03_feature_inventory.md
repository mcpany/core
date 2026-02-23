# Feature Inventory: Universal Agent Bus

## Priority Features (P0 - P1)

| Feature | Priority | Category | Status | Description |
| :--- | :--- | :--- | :--- | :--- |
| **Policy Firewall** | P0 | Security | Proposed | Rego/CEL based hooking for tool calls to enforce Zero Trust. |
| **HITL Middleware** | P0 | Safety | Proposed | Human-in-the-Loop suspension protocol for user approval flows. |
| **Recursive Context** | P1 | Comms | Proposed | Standardize headers for Subagent inheritance (Recursive Context Protocol). |
| **Shared KV Store** | P1 | State | Proposed | Embedded SQLite "Blackboard" tool for agent coordination. |
| **Lazy Tool Discovery** | P1 | Perf | New | Support for on-demand tool listing to reduce context window usage. |
| **Tool Namespacing** | P1 | Architecture | New | Automatic prefixing (`server__tool`) for collision resolution. |
| **Policy-Based Scopes** | P1 | Security | Proposed | Capability-based token system (e.g., `fs:read:/tmp`). |

## Upcoming Features (P2 - P3)

| Feature | Priority | Category | Status | Description |
| :--- | :--- | :--- | :--- | :--- |
| **Tool Playground** | P2 | DevX | Planned | UI for testing tools with auto-generated forms. |
| **Live Marble Diagrams** | P2 | Observability | Planned | Reactive visualization of agent flows and dependencies. |
| **Interactive Debugger** | P2 | Debug | Planned | Breakpoints and variable inspection for tool calls. |
| **Team Config Sync** | P2 | Collab | Planned | Secure synchronization of configurations across teams. |
| **Smart Error Recovery**| P2 | Resilience | Planned | LLM-based self-healing for tool execution errors. |

## Deprecated / Re-prioritized
*   *None at this time.*

## Strategic Alignment: 2026-02-23
*   **Lazy Tool Discovery** added to address context exhaustion issues identified in Claude Code.
*   **Tool Namespacing** added to align with Gemini CLI's conflict resolution strategy.
*   Elevated **Policy Firewall** to P0 due to the rise of autonomous agents with local execution.
