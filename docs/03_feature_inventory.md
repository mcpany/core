# Feature Inventory: MCP Any

This is a rolling masterlist of priority features for MCP Any.

## Priority Features (Groomed: 2025-05-22)

### High Priority (P0)
*   **[NEW] Rego Policy Firewall**: Integrate OPA for fine-grained tool call authorization.
*   **HITL Middleware**: User approval flow for sensitive operations.
*   **Standardized Recursive Context**: Protocol-level inheritance of session headers/credentials.

### Medium Priority (P1)
*   **[NEW] Shared Agent Blackboard**: Scoped KV store tool for cross-agent state synchronization.
*   **Tool Execution Simulation**: Mocking tool calls for development.
*   **Dynamic Tool Pruning**: Filter visible tools based on session context to save tokens.

### Low Priority (P2)
*   **Config Inheritance**: `extends: base.yaml` support for configuration files.
*   **Plugin Marketplace**: Browser for community MCP servers.
*   **Agent Black Box Player**: Replay of recorded agent sessions.

## Recently Added/Deprecated
*   **[2025-05-22] Added**: Rego Policy Firewall (Strategic Shift to Zero Trust).
*   **[2025-05-22] Added**: Shared Agent Blackboard (Supporting Decentralized Swarms).
*   **[2025-05-22] Re-prioritized**: Recursive Context (Moved to P0 due to OpenClaw trends).
