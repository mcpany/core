# Strategic Vision: MCP Any (Universal Agent Bus)

## Overview
MCP Any aims to be the indispensable core infrastructure layer for all AI agents, subagents, and swarms. It bridges the gap between fragmented AI tools and diverse backend services through a standardized, secure, and observable gateway.

## Core Pillars
1. **Universal Connectivity:** Seamlessly adapt REST, gRPC, CLI, and Filesystems into MCP tools.
2. **Zero Trust Security:** Enforce strict policy-based access control and isolation for autonomous agents.
3. **Observability & Control:** Provide deep visibility into agent-tool interactions and human-in-the-loop (HITL) approval flows.
4. **Standardized Context:** Simplify subagent orchestration through recursive context inheritance.

## Strategic Evolution: [2026-02-19]
### The Zero Trust Gateway Pivot
The recent security crisis in the OpenClaw ecosystem (RCE and command injection vulnerabilities) has highlighted a critical gap in the "Shadow AI" landscape. Agents are being granted broad system access without a centralized security perimeter.

**Strategic Opportunity:**
MCP Any must transition from a "Universal Adapter" to a **"Zero Trust Agent Gateway"**.

**Key Actions:**
- **Standardized Context Inheritance:** Implement a protocol for subagents to inherit security contexts and environment variables without exposing them to the LLM.
- **Shared State (Blackboard):** Provide a secure, shared KV store (SQLite-backed) for agents to coordinate without data leakage between sessions.
- **Isolated Execution Enclaves:** Move beyond simple tool calls to isolated, Docker-bound execution environments for command-line tools, preventing host-level compromise.
- **Policy-as-Code:** Adopt Rego/CEL for defining granular tool-call permissions, enabling enterprises to safely deploy autonomous agents.
