# Feature Inventory: MCP Any

This is the rolling masterlist of priority features for MCP Any.

## Priority Features (Active Development)

| ID | Feature Name | Description | Status | Priority |
| :--- | :--- | :--- | :--- | :--- |
| FEAT-001 | Policy Firewall Engine | Rego/CEL based tool call validation. | In Development | P0 |
| FEAT-002 | Recursive Context Protocol | Standardized context inheritance for subagents. | Proposed | P1 |
| FEAT-003 | Shared KV "Blackboard" | Embedded SQLite store for inter-agent state. | Proposed | P1 |
| FEAT-004 | Zero Trust Subagent Sandbox | Isolated execution environment for untrusted subagents. | New | P0 |
| FEAT-005 | Automated Policy Generator | LLM-assisted generation of security policies based on usage. | New | P2 |
| FEAT-006 | HITL Middleware | Human-in-the-loop approval protocol for sensitive actions. | In Review | P0 |

## Strategic Updates: 2026-02-22

- **Added FEAT-004 (Zero Trust Subagent Sandbox):** Direct response to OpenClaw's local execution vulnerabilities.
- **Added FEAT-005 (Automated Policy Generator):** Lowers the barrier to entry for secure configuration.
- **Upgraded FEAT-002 (Recursive Context Protocol) to P1:** Critical for supporting Claude Code/Gemini CLI agent swarms.
