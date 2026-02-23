# Strategic Vision: MCP Any

## Mission Statement
MCP Any aims to be the indispensable core infrastructure layer for all AI agents, subagents, and swarms. It provides a universal adapter and gateway that standardizes how agents interact with tools, manage context, and enforce security policies.

## Core Pillars
1. **Universal Connectivity**: Support any MCP server, any LLM, and any agent framework.
2. **Zero Trust Security**: Granular, capability-based access control for all tool calls.
3. **Context Persistence**: Shared state and context inheritance across agent swarms and execution environments.

---

## Strategic Evolution: [2026-02-23]
### Focus: Standardized Context Inheritance & Multi-Env Bridging
**Context**: Today's research highlights a major gap in how subagents inherit parent context and how agents bridge the gap between cloud sandboxes (e.g., Anthropic's) and local tools.
**Strategic Pivot**:
- **Environment Bridging**: MCP Any will act as a "secure proxy" that synchronizes state between sandboxed environments and local execution.
- **Context Inheritance Protocol**: Implementing a recursive header standard that allows subagents to automatically inherit "intent-scoped" context without bloating the LLM window.
- **Zero-Knowledge Context**: Ensuring subagents only receive the minimal state required for their specific task, following the principle of least privilege.

---

## Strategic Evolution: [2026-02-24]
### Focus: Hardened Egress & Path Validation (The SSRF Shield)
**Context**: Recent critical vulnerabilities in OpenClaw (SSRF, Path Traversal) have exposed a massive structural weakness in how agent gateways handle tool parameters.
**Strategic Pivot**:
- **Zero-Trust Egress Shield**: MCP Any will implement mandatory egress filtering for all tool calls, blocking access to internal metadata services (e.g., 169.254.169.254) and private networks by default.
- **Path Sanitization Middleware**: To mitigate path traversal, MCP Any will enforce strict boundary checks for all file-based tools, ensuring agents cannot escape configured root directories.
- **Attestation-Based Tool Invocation**: Transitioning towards a model where every tool call must be accompanied by a cryptographically signed "Intent Token" to prevent unauthorized tool hijacking.
