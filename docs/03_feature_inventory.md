# Feature Inventory: MCP Any

This is the rolling masterlist of priority features for MCP Any.

## Priority Features

| Feature ID | Name | Priority | Status | Description |
| :--- | :--- | :--- | :--- | :--- |
| F-001 | Policy Firewall Engine | P0 | In Progress | Rego/CEL based hooking for tool calls. |
| F-002 | HITL Middleware | P0 | Planned | Human-in-the-loop approval flows for dangerous tools. |
| F-003 | Shared KV Store | P1 | Planned | Embedded SQLite "Blackboard" for agents. |
| F-004 | Recursive Context | P1 | Planned | Standardized headers for subagent inheritance. |

## New Additions: [2026-02-22]

*   **F-005: Isolated Subagent Comms (Named Pipes):** [P0] Implementation of isolated, non-TCP communication channels for inter-agent routing to mitigate local port exposure exploits.
*   **F-006: Zero Trust Tool Sandbox:** [P0] A containerized execution environment for "Command" upstreams to prevent host-level access by autonomous agents.
*   **F-007: Agent Discovery Notification:** [P1] Implementation of the MCP notification protocol to allow agents to discover new tools and other agents dynamically.
