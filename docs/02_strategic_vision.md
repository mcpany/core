# Strategic Vision: MCP Any (Universal Agent Bus)

## Context
MCP Any is evolving from a simple tool gateway to a Universal Agent Bus. This document tracks the strategic shifts and architectural north stars.

## Core Pillars
1. **Universal Connectivity**: Support every protocol (HTTP, gRPC, CLI, etc.).
2. **Zero Trust execution**: Secure agent interactions via sandboxing and granular policies.
3. **Seamless State & Context**: Standardize how agents share information and maintain session continuity.

## Strategic Evolution: 2026-02-22

### From Gateway to Bus
The market is rapidly shifting from single-agent tool access to multi-agent swarms. MCP Any must transition from a "Passive Gateway" to an "Active Agent Bus."

### Key Patterns for Adoption
*   **Recursive Context Inheritance**: Standardizing how `auth` and `trace` headers flow between agents. This prevents the "Context Cliff" where a subagent loses the original user's intent or permissions.
*   **Zero Trust Inter-Agent Comms**: Moving away from local port exposure. We will prioritize **Isolated Docker-bound Named Pipes**. This ensures that even if a subagent is compromised, it cannot scan the host network or access unauthorized local services.
*   **Agent-as-a-Tool (AaaT)**: Treating complex agent workflows as standardized MCP tools, enabling true recursive agentic architectures.
