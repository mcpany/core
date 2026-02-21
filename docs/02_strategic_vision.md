# Strategic Vision: MCP Any

## Overarching Goal
To become the universal, configuration-driven interface for all AI interactions with external tools and systems.

## Strategic Evolution: 2026-02-21
**Finding:** The rise of multi-agent CLI tools (Claude Code, AutoGen swarms) highlights a major gap in **context inheritance**. Currently, each subagent or tool call often starts from a blank slate or requires manual passing of context (auth tokens, project root, user-imposed restrictions).

**Strategy Update:**
*   **The Universal Agent Bus:** MCP Any will evolve from a simple adapter into a "Context Bus". It will not just proxy calls, but manage a shared "Context Layer" that can be inherited by any agent connected to the gateway.
*   **Zero Trust for Subagents:** We will implement a "Permission Inheritance" model. If a parent agent is restricted from certain paths, all subagents routed through MCP Any automatically inherit these restrictions.
*   **Protocol Neutrality:** While MCP is our primary interface, we will prepare for A2A (Agent-to-Agent) standards to ensure MCP Any remains the core infrastructure even as orchestration patterns shift.
