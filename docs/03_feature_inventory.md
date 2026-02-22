# Feature Inventory: MCP Any

This document maintains a rolling masterlist of priority features for MCP Any.

## Priority 0 (Critical)
*   **Policy Firewall Engine:** Rego/CEL based hooking for tool calls.
*   **HITL Middleware:** Suspension protocol for user approval flows.
*   **Zero Trust Tool Execution:** Isolated execution environments for command-based tools.
*   **[NEW] Secure Sandbox Adapter:** Docker-bound named pipes for isolated CLI/FS tool execution.

## Priority 1 (High)
*   **Recursive Context Protocol:** Standardized headers for Subagent inheritance.
*   **Shared Key-Value Store:** Embedded SQLite "Blackboard" tool for agents.
*   **Tool Playground & Explorer:** Auto-generated forms and history visualization.
*   **[NEW] Just-in-Time Tool Loading:** Dynamic tool registration/unregistration to optimize context window.

## Priority 2 (Medium)
*   **Plugin Marketplace:** Discover and install community MCP servers.
*   **Cost & Metrics Dashboard:** Token usage and performance tracking.
*   **Service Health History:** Historical availability trends.
*   **[NEW] Dynamic Discovery Notifications:** Server-sent events for tool availability changes in the swarm.

## Completed
*   Interactive Doctor Resilience
*   Pre-flight Command Validation
*   Actionable Configuration Errors
*   Async Tool Loading
*   Smart Retry Policies
