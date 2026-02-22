# Feature Inventory: MCP Any

This is a master list of all features, ranging from active to proposed.

| ID | Feature Name | Status | Priority | Description |
|---|---|---|---|---|
| F-001 | Policy Firewall | Active | P0 | Rego/CEL based hooking for tool calls. |
| F-002 | HITL Middleware | Active | P0 | Human-in-the-Loop approval flows. |
| F-003 | Shared KV Store | Active | P1 | SQLite "Blackboard" for agents. |
| F-004 | Recursive Context | Active | P1 | Standardized headers for Subagent inheritance. |

## New Feature Proposals: 2026-02-22

| ID | Feature Name | Status | Priority | Description |
|---|---|---|---|---|
| F-005 | OpenClaw Security Sandbox | Proposed | P0 | Isolated Docker-bound or gVisor-based environment for executing local CLI tools to prevent unauthorized host access. |
| F-006 | Cross-Platform Skill Adapter | Proposed | P1 | Bi-directional translation between Claude Code "Skills" and Gemini CLI "Extensions" using MCP as the common format. |
| F-007 | Inter-Agent "Clink" Protocol | Proposed | P1 | Standardized MCP notification and request routing for inter-agent communication (e.g., Agent A calling Agent B). |
| F-008 | Zero Trust Local FS Adapter | Proposed | P0 | A filesystem adapter that uses "named pipes" or isolated volumes instead of direct host mounts. |
