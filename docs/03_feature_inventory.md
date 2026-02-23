# Feature Inventory: MCP Any

## Current Backlog (P0/P1)
- **Policy Firewall**: Rego/CEL based hooking for tool calls.
- **HITL Middleware**: Suspension protocol for user approval flows.
- **Recursive Context Protocol**: Standardized headers for subagent inheritance.
- **Shared KV Store**: Embedded SQLite "Blackboard" tool for agents.

---

## Evolution: [2026-02-23] Updates

### Proposed Additions
- **Environment Bridging Middleware**: (P1) Bridge between cloud-sandboxed agents (e.g., Claude Code Sandbox) and local MCP Any tools. Enables seamless state transfer.
- **Machine-Checkable Security Contracts**: (P1) Declarative security models for tools that can be verified by automated agents (inspired by OpenClaw).
- **Zero-Trust Subagent Scoping**: (P0) Capability-based tokens that restrict subagents to a specific "intent-scope" of a parent's permissions.

### Priority Shifts
- **Recursive Context Protocol**: Promoted from **P1** to **P0**. Essential for modern agent swarms to prevent state loss.
- **Shared KV Store**: Promoted from **P1** to **P0**. Critical for coordinating multi-agent actions in complex workflows.

---

## Evolution: [2026-02-24] Updates

### Proposed Additions
- **Zero-Trust Egress Shield**: (P0) Mandatory egress filtering for tool calls (anti-SSRF).
- **Path Sanitization Middleware**: (P0) Strict directory boundary enforcement for filesystem tools.
- **Signed Intent Tokens**: (P1) Cryptographic attestation for tool invocations to prevent hijacking.

### Priority Shifts
- **Machine-Checkable Security Contracts**: Promoted from **P1** to **P0**. Essential for automated auditing of tool safety in light of recent exploits.

### Deprecations / Monitoring
- *None today.*
