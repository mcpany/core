# Market Sync: 2026-03-04

## Ecosystem Updates

### OpenClaw: ClawMesh Release
OpenClaw has launched "ClawMesh," a P2P discovery protocol for subagents. This moves away from centralized registries and emphasizes local-first discovery. MCP Any should position itself as the "Mesh Gateway" that translates ClawMesh discovery into standard MCP toolsets.

### Claude Code: Enhanced Tool Sandboxing
Anthropic has tightened the "Claude Code" sandbox, specifically restricting local MCP servers from accessing sensitive directories (like `.ssh` or `.aws`) even if the agent has general filesystem access. This reinforces the need for MCP Any's "Intent-Aware" policy engine to provide granular, folder-level tool permissions.

### Gemini CLI: Native MCP Tool Support
Google's Gemini CLI now natively supports MCP servers via a new `--mcp-server` flag. However, users report "Context Drift" where the CLI doesn't maintain state between multiple tool-heavy prompts. MCP Any's "Shared KV Store" is the perfect solution for this.

## New Autonomous Agent Pain Points

### PITTO (Prompt Injection via Tool Output)
A new class of attack where a compromised or malicious tool returns a payload that, when processed by the LLM, triggers unauthorized actions (e.g., "Delete all files").
- **Impact**: Critical.
- **Opportunity**: MCP Any can implement a "Tool Output Sanitizer" middleware.

### Agentic "Shadow IT"
Large swarms (CrewAI/AutoGen) are increasingly spawning ephemeral sub-processes or containers that don't follow corporate security policies.
- **Impact**: Compliance risk.
- **Opportunity**: MCP Any as the "Single Point of Egress" for all agent-spawned tools.

### Identity Crisis: OIDC for Agents
Agents currently use the same API keys as humans, making it impossible to audit "Agent-originated" vs "Human-originated" actions.
- **Impact**: Auditability and Security.
- **Opportunity**: Implementing "Agent-on-Behalf-of-User" OIDC flows.

## GitHub/Social Trends
- **GitHub Trending**: `agent-governance-frameworks` is rising.
- **Reddit (r/LocalLLM)**: High frustration with the latency of "Tool-Calling Loops" in local swarms. Users are asking for "Batch Tool Execution" protocols.
