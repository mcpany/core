# Feature Inventory: MCP Any

## High Priority (P0)
- **Policy Firewall Engine:** Rego/CEL based hooking for tool calls with Zero Trust isolation.
- **HITL Middleware:** Human-in-the-loop approval flow for sensitive actions.
- **Shared Key-Value Store:** Embedded SQLite "Blackboard" for inter-agent state sharing.

## Medium Priority (P1)
- **Recursive Context Protocol:** Standardized headers for subagent context inheritance.
- **Zero-Trust Sandbox for Local Execution:** Docker-bound execution environment for command-based tools.
- **Autonomous Tool Synthesis (AgentSkills compatibility):** Interface for agents to dynamically register synthesized tools.
- **Dynamic Tool Pruning:** Context-aware schema truncation to reduce token usage.

## Low Priority (P2)
- **Team Configuration Sync:** Secure sharing of `mcpany` configurations.
- **Smart Error Recovery:** LLM-driven self-healing for tool errors.
- **Plugin Marketplace:** In-app discovery of community MCP servers.

## Recent Changes: [2026-02-22]
- **Added:** "Autonomous Tool Synthesis" (P1) - Essential for compatibility with OpenClaw's AgentSkills.
- **Added:** "Zero-Trust Sandbox for Local Execution" (P1) - Critical for Claude Code security requirements.
- **Upgraded:** "Recursive Context Protocol" from P2 to P1 due to increased demand for agent swarms.
