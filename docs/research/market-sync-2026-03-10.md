# Market Sync: 2026-03-10

## Ecosystem Shifts

### 1. Universal Agent Playbooks (`SKILL.md`)
- **Trend**: Anthropic (Claude Code), Cursor, Gemini CLI, and others have converged on the `SKILL.md` format for defining agent capabilities, templates, and specialized context.
- **Impact**: Agents are no longer just tool-callers; they are "skill-aware" collaborators. `SKILL.md` provides a project-local or global playbook that agents can search and invoke.
- **Gap**: MCP Any needs to bridge the gap between "Tools" (MCP) and "Skills" (SKILL.md). A skill should be able to orchestrate multiple MCP tools.

### 2. Generalist Agent Delegation (Gemini CLI v0.32.0)
- **Trend**: Google's Gemini CLI now features a "Generalist Agent" that handles task delegation and routing.
- **Insight**: Routing is moving from hardcoded logic to agentic reasoning at the gateway level.
- **Opportunity**: MCP Any's A2A protocol should adopt this "Generalist Router" pattern to allow seamless task handoffs between specialized subagents (e.g., routing a security task to a "Shannon" pen-testing subagent).

### 3. OpenClaw Security & Transport Overhaul (v2026.2.26)
- **Trend**: OpenClaw has moved to "Threadbound Agents" to prevent context mixing and "WebSocket-first" transport for reliable long-lived connections.
- **Implication**: Multi-agent swarms require strict session/thread isolation at the transport layer to prevent accidental state leakage between different user conversations.

## New "Autonomous Agent Pain Points"
- **Context Mixing**: Threadbound isolation is becoming a hard requirement for agents running on multi-user platforms (Discord/Telegram).
- **Skill Discovery Friction**: While `SKILL.md` is a great format, agents still struggle to "find" the right skill in a large repository without bloating the context window.
- **Secret Floating**: API keys still leak into conversation contexts if not handled by an external "Secret Store" that the agent can access via a secure tool.

## Strategic Summary for Today
MCP Any must evolve to support **Universal Skill Orchestration**. By ingesting `SKILL.md` files and exposing them as "Super-Tools" (composite tools), MCP Any becomes the brain that tells the agent *how* to use its tools, not just *what* tools are available.
