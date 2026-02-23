# MCP Any: Feature Inventory

This document tracks the current and proposed features for MCP Any, aligned with our Strategic Vision.

## Current High-Priority Backlog

| Feature | Status | Priority | Description |
| :--- | :--- | :--- | :--- |
| **Policy Firewall Engine** | In Design | P0 | Rego/CEL based hooking for tool calls. |
| **HITL Middleware** | In Design | P0 | Suspension protocol for user approval flows. |
| **Recursive Context Protocol** | In Design | P1 | Standardized headers for subagent inheritance. |
| **Shared KV Store** | Backlog | P1 | Embedded SQLite "Blackboard" tool for agents. |

## Proposed Additions: 2026-02-23

### Zero Trust Tool Sandboxing (P0)
- **Context**: Market shifts in OpenClaw show a need for isolated tool execution.
- **Description**: Automatically wrap Command-line and Filesystem tools in ephemeral Docker containers or WASM sandboxes.

### Cross-Agent Shared State (Blackboard) (P1)
- **Context**: Need to reduce context bloat in swarms.
- **Description**: A dedicated MCP tool that provides a temporary, scoped key-value store for agents in the same session.

### Global Tool Discovery Registry (P2)
- **Context**: Claude Code latency issues with 50+ local servers.
- **Description**: An optimized, cached index of all available tools across multiple MCP Any instances.
