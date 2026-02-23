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

### Deprecations / Monitoring
- *None today.*

---

## Evolution: [2026-02-24] Updates

### Proposed Additions
- **Advanced Multi-Agent Session Management**: (P0) A session-aware middleware that tracks tool state and handoffs between multiple specialized agents.
- **Unified MCP Discovery Service**: (P1) Automated discovery and registry for local and remote MCP servers (Stdio, HTTP, FastMCP).
- **Session-Bound State Persistence**: (P1) Ensuring that multi-agent "long-running" tasks maintain state across tool calls and agent switches.

### Priority Shifts
- **Policy Firewall**: Promoted to **P0** to support secure "Zero Trust" subagent isolation as ecosystems become more complex.
