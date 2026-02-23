# Feature Inventory

## High Priority (P0)
- **Policy Firewall (Rego/CEL)**: Strict validation of tool calls.
- **Human-in-the-Loop (HITL) Middleware**: User approval for sensitive operations.
- **Granular Scopes & Tokenization**: Capability-based security.

## Medium Priority (P1)
- **Recursive Context Protocol**: Standardized context inheritance for subagents.
- **Shared KV Store (Blackboard)**: Shared state for agent swarms.
- **Dynamic Tool Notification System**: Real-time discovery of new capabilities.

## Low Priority (P2)
- **Team Config Sync**: Multi-user configuration management.
- **Multi-Model Advisor**: Synthesizing insights across models.

---

## Update: 2026-02-23

### Additions
- **[P0] Sandboxed Command Runner**: An isolated environment for executing shell commands to mitigate host-level exposure risks (Response to OpenClaw security gaps).
- **[P1] Moltbook Integration Bridge**: Standardized adapter for connecting MCP Any agents to agentic social networks.

### Priority Shifts
- **Recursive Context Protocol**: Promoted from P1 to **P0**. Essential for handling the complexity of the growing subagent ecosystem.
- **Shared KV Store (Blackboard)**: Promoted from P1 to **P0**. Critical for reducing hallucinations in autonomous swarms.

### Deprecations
- **Basic Auth Headers**: Moving towards **OIDC/OAuth2 Integration** as the standard for enterprise agent authentication.
