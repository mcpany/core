# Feature Inventory: MCP Any

This is the rolling masterlist of priority features for MCP Any, reflecting current development status and strategic alignment.

## P0: Critical Infrastructure (The Universal Agent Bus)

| Feature ID | Feature Name | Description | Status |
| :--- | :--- | :--- | :--- |
| CORE-001 | **Policy Firewall** | Rego/CEL based hooking for tool calls and zero-trust security. | In Progress |
| CORE-002 | **Recursive Context Protocol** | Standardized headers for context inheritance between agents. | **Proposed** |
| CORE-003 | **HITL Middleware** | Suspension protocol for human-in-the-loop approval flows. | In Progress |
| CORE-004 | **Shared Key-Value Store** | Embedded "Blackboard" tool for inter-agent coordination. | **Proposed** |

## P1: Developer Experience & Safety

| Feature ID | Feature Name | Description | Status |
| :--- | :--- | :--- | :--- |
| DX-001 | **Tool Playground** | Interactive UI to test tool calls and visualize results. | In Progress |
| DX-002 | **Context Optimizer** | Automatic truncation of large outputs to save token usage. | **Completed** |
| SEC-001 | **Isolated Shell Sandbox** | Secure, containerized execution environment for command tools. | **Proposed** |
| SEC-002 | **Log Redaction** | Regex-based redaction of PII and secrets in logs. | Proposed |

## P2: Ecosystem & Scaling

| Feature ID | Feature Name | Description | Status |
| :--- | :--- | :--- | :--- |
| ECO-001 | **Plugin Marketplace** | In-app browser for community MCP servers. | Backlog |
| OPS-001 | **Compliance Reporting** | Automated SOC2/GDPR audit log generation. | Backlog |

## Recent Updates: [2026-02-22]
- **Added**: CORE-002 (Recursive Context Protocol) - Identified as critical for agent swarm coordination.
- **Added**: CORE-004 (Shared Key-Value Store) - To support inter-agent memory.
- **Added**: SEC-001 (Isolated Shell Sandbox) - Mitigates shell access risks found in OpenClaw-like deployments.
- **Priority Shift**: CORE-001 (Policy Firewall) moved to top of P0 to address security concerns.
