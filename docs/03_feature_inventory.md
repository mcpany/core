# Feature Inventory: MCP Any (Universal Agent Bus)

## Master Feature List

| ID | Feature Name | Status | Priority | Description |
| :--- | :--- | :--- | :--- | :--- |
| F-001 | Policy Firewall | In Design | P0 | Rego/CEL based hooking for tool calls. |
| F-002 | HITL Middleware | In Design | P0 | Human-in-the-loop approval flows. |
| F-003 | Recursive Context | Proposed | P1 | Standardized headers for Subagent inheritance. |
| F-004 | Shared KV Store | Proposed | P1 | Embedded SQLite "Blackboard" tool for agents. |
| F-005 | Named Pipe Transport | Backlog | P1 | Support for Unix Domain Sockets/Named Pipes for local isolation. |
| F-006 | Capability Handover | Backlog | P1 | Secure scope delegation for subagents. |
| F-007 | Agent Black Box Player | Backlog | P2 | Timeline-based replay of recorded agent sessions. |
| F-008 | Plugin Marketplace | Backlog | P2 | In-app browser for community MCP servers. |

---

## Evolution: 2026-02-22
- **Added:** [F-009] **Isolated Named Pipe Comms** (High Priority). Specifically targeting the "MCP-Port-Sniff" vulnerability.
- **Added:** [F-010] **Zero Trust Context Inheritance** (High Priority). Standardizing how subagents inherit and prune parent context.
- **Reprioritized:** Moved **Recursive Context** (F-003) from P1 to P0 due to widespread reports of context drift in agent swarms.
