# Market Sync: 2026-03-11

## Ecosystem Shifts

### 1. Claude Code: Tool Search & SKILL.md Standard
- **Context Pollution Solved**: Claude Code officially shipped "MCP Tool Search," moving from upfront tool schema pushing to on-demand discovery. This reduces initial token overhead by ~85%, enabling sessions with hundreds of MCP servers.
- **SKILL.md Portability**: The `SKILL.md` format has emerged as the "Docker Compose for Agents," allowing specialized playbooks to be shared across Claude Code, Gemini CLI, and Cursor.

### 2. Gemini CLI: Policy Engine & Generalist Agents
- **Project-Level Policies**: Gemini CLI v0.32.0 introduced project-level policy enforcement and parallel extension loading.
- **Generalist Agent Routing**: Improved delegation logic for routing tasks between specialized subagents.

### 3. OpenClaw: Multi-Agent Refinement Risks
- **Routing Exploits**: A new exploit pattern has been identified in OpenClaw's subagent routing where local HTTP tunneling for inter-agent communication is being intercepted or spoofed, leading to unauthorized host-level file access.

## Autonomous Agent Pain Points
- **Discovery Friction**: Even with Tool Search, agents still struggle with "Intent-to-Tool" mapping when tool names are ambiguous.
- **Insecure Defaults**: The "8,000 Exposed Servers" crisis highlights that developers are still deploying MCP servers bound to `0.0.0.0` without authentication.
- **Context Fragmentation**: State loss during handoffs between different agent frameworks (e.g., moving a task from Claude Code to a specialized OpenClaw swarm).

## Security Vulnerabilities
- **Shadow MCP Servers**: Unvetted MCP servers being "vibe-coded" and deployed rapidly are creating a massive, unmanaged non-human identity (NHI) attack surface.
- **RCE via Hook Ingestion**: Continued risk of agents ingesting malicious hooks from collaborator-provided project configs.
