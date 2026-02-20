# Feature Inventory: MCP Any

This document maintains a rolling masterlist of priority features and capabilities for the MCP Any ecosystem.

## 1. Core Protocol & Adapters
*   **HTTP/REST Adapter**: Full support for OpenAPI 3.0/3.1 with parameter mapping.
*   **gRPC Adapter**: Dynamic method discovery via reflection.
*   **Command Adapter**: Safe execution of CLI tools.
*   **Filesystem Adapter**: Secure local/remote file access.

## 2. Security & Policy
*   **Policy Firewall Engine (P0)**: Rego/CEL based hooking for tool calls.
*   **Granular Scopes (P0)**: Capability-based token system.
*   **Audit Logging**: Detailed trail of all tool executions.
*   **DLP (Data Loss Prevention)**: Sensitive data redaction in logs and traces.

## 3. Advanced Middleware
*   **HITL Middleware (P0)**: Human-in-the-Loop approval flows.
*   **Recursive Context Protocol (P1)**: Context inheritance for subagents.
*   **Shared Key-Value Store (P1)**: SQLite-backed blackboard for agent memory.

## 4. Observability & Debugging
*   **Real-time Topology Graph**: Visualizer for agent-tool connections.
*   **Tool Execution Timeline**: Waterfall charts for latency debugging.
*   **Service Health History**: Uptime and error trend visualization.

## Strategic Backlog Update: 2025-02-17
*   **[New] JIT Tool Loader (P1)**: Lazy-load tool definitions based on agent intent or task classification.
*   **[New] Graph-based Orchestration Middleware (P2)**: Middleware to manage complex tool-calling dependencies across multiple agents.
*   **[Priority Shift] Policy Firewall**: Elevated to P0+ to address emerging "rogue subagent" exploit patterns.
