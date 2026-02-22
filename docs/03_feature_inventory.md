# Feature Inventory

This document maintains a rolling masterlist of priority features for MCP Any.

## Current High-Priority Features

| Feature ID | Name | Description | Priority | Status |
| :--- | :--- | :--- | :--- | :--- |
| F-001 | Policy Firewall | Rego/CEL based hooking for tool calls. | P0 | In Progress |
| F-002 | HITL Middleware | Human-in-the-loop approval flows. | P0 | Planned |
| F-003 | Shared KV Store | SQLite-based blackboard for agents. | P1 | Planned |
| F-004 | Recursive Context Protocol (RCP) | Standardized header propagation for subagents. | P0 | Draft |
| F-005 | Zero Trust Named Pipes | Isolated inter-agent communication channels. | P1 | Research |
| F-006 | AgentSkills Sandbox | Secure execution for agent-generated scripts. | P1 | Research |

## Priority Shifts: 2026-02-22

*   **Promoted**: F-004 (RCP) moved to P0 due to high demand in OpenClaw/Claude Code integrations.
*   **Added**: F-005 and F-006 to address security gaps in autonomous agent execution.
