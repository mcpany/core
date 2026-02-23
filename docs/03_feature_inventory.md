# Feature Inventory: MCP Any

## Strategic Priorities (P0)
*   **Policy Firewall Engine**: Rego/CEL based hooking for tool calls.
*   **HITL Middleware**: Suspension protocol for human-in-the-loop approval.
*   **Recursive Context Protocol (RCP)**: [NEW 2026-02-23] Standardized headers for subagent state inheritance.
*   **Isolated IPC (Named Pipes)**: [NEW 2026-02-23] Secure, Docker-bound inter-agent communication to replace local HTTP tunnels.

## Core Infrastructure (P1)
*   **Granular Scopes**: Capability-based token system (e.g., `fs:read:/tmp`).
*   **Shared Key-Value Store**: SQLite-backed "Blackboard" tool for agents.
*   **Tool Playground & Explorer**: Interactive UI for testing MCP tools.
*   **Live Marble Diagrams**: Real-time visualization of agent-tool flows.

## Operational Excellence (P2)
*   **Plugin Marketplace**: In-app browser for community MCP servers.
*   **Agent Black Box Player**: Replay of recorded agent sessions.
*   **Cost & Metrics Dashboard**: Token usage and P95 latency visualization.
*   **Secret Rotation Helper**: CLI tool for identifying and rotating service secrets.

## Deprecated / Re-prioritized
*   *Local HTTP Tunneling*: [Deprioritized 2026-02-23] Replaced by Isolated IPC due to security vulnerabilities identified in OpenClaw.
