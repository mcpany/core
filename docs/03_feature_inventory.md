# Feature Inventory: MCP Any

## High-Priority Features (P0-P1)

| Feature ID | Name | Status | Category | Priority |
| :--- | :--- | :--- | :--- | :--- |
| SEC-001 | Policy Firewall Engine | Planned | Security | P0 |
| SEC-002 | HITL Middleware | Planned | Security | P0 |
| COM-001 | Recursive Context Protocol | Proposed | Connectivity | P1 |
| COM-002 | Shared Key-Value Store | Planned | Connectivity | P1 |
| SEC-003 | OpenClaw Security Sandbox | Proposed | Security | P1 |
| OBS-001 | Real-time Marble Diagrams | Planned | Observability | P1 |

## Feature Proposals: 2026-02-22

### 1. [COM-001] Recursive Context Middleware
- **Description:** Implements a header-based context propagation protocol for MCP.
- **Value:** Allows subagents to inherit goal, session, and security context from parent agents automatically.

### 2. [SEC-003] OpenClaw Security Sandbox
- **Description:** A specialized adapter that runs OpenClaw tool executions (shell, browser) in isolated Docker containers.
- **Value:** Directly addresses the primary security blocker for OpenClaw institutional adoption.
