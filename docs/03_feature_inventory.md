# Feature Inventory: MCP Any

This is a rolling masterlist of priority features, categorized by their strategic importance and implementation status.

## Current High Priority (Universal Agent Bus)

| Feature ID | Feature Name | Description | Priority | Status |
| :--- | :--- | :--- | :--- | :--- |
| **F-001** | **Policy Firewall Engine** | Rego/CEL based hooking for tool calls to enforce Zero Trust. | P0 | In Review |
| **F-002** | **Recursive Context Protocol** | Standardized headers for Subagent inheritance (Context & Auth). | P0 | Draft |
| **F-003** | **HITL Middleware** | Human-in-the-loop suspension protocol for sensitive tool calls. | P0 | Active |
| **F-004** | **Shared KV Store** | Embedded SQLite "Blackboard" tool for shared agent state. | P1 | Upcoming |
| **F-005** | **Managed MCP Bridge** | Local proxy and policy enforcer for remote Managed MCPs (Google/Vertex). | P1 | **Proposed** |
| **F-006** | **Cross-Agent Memory** | Adapter for CLAUDE.md and other local memory standards. | P2 | **Proposed** |

## Core Gateway Features

| Feature ID | Feature Name | Description | Status |
| :--- | :--- | :--- | :--- |
| **G-001** | **HTTP Adapter** | Universal REST/JSON bridge with mapping templates. | Completed |
| **G-002** | **gRPC Adapter** | Dynamic discovery and invocation via reflection. | Completed |
| **G-003** | **Command Adapter** | Secure local CLI execution. | Completed |
| **G-004** | **Discovery Engine** | Auto-detection of local tools (Ollama, etc.). | Completed |

## Upcoming / Backlog

- **Smart Error Recovery**: LLM-driven self-healing for tool calls.
- **Canary Tool Deployment**: Gradual rollout of tool version changes.
- **Tool Execution Timeline**: Visual debugging of the tool call lifecycle.
- **Config Inheritance**: YAML `extends` capability to reduce duplication.

## Strategic Additions: [2026-02-17]
- **F-005: Managed MCP Bridge**: Necessity driven by Google Cloud's Managed MCP launch. Allows MCP Any to act as a secure, local control point for enterprise cloud MCPs.
- **F-006: Cross-Agent Memory**: Driven by Claude Code's CLAUDE.md pattern. Provides a protocol-level abstraction for shared memory files.
