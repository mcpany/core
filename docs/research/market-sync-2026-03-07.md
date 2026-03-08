# Market Sync: 2026-03-07

## Overview
Today's ecosystem sync focuses on the transition from single-agent tool use to multi-agent "Swarm" orchestration, and the formalization of authorization discovery within the MCP protocol.

## Key Findings

### 1. Claude Code: Agent Teams & TeammateTool
Anthropic has officially released "Swarm mode" (Agent Teams) in Claude Code.
- **Orchestration Pattern:** A "Team Lead" agent coordinates multiple "Teammate" sessions.
- **Communication:** Teammates communicate via a specialized `TeammateTool` containing operations for spawning, task assignment, and message passing.
- **State Management:** Teammates do not inherit conversation history but share project context (CLAUDE.md) and task files on disk. This reinforces the need for MCP Any to support "Stateful Buffering" for A2A communication.

### 2. Gemini CLI v0.32.0 Update
- **Parallel Extension Loading:** Significant performance improvements in startup times by parallelizing MCP extension loading.
- **Enhanced Policy Engine:** Added support for project-level policies and, crucially, **MCP server wildcards** and tool annotation matching. This suggests MCP Any should implement more granular wildcard-based tool scoping in its Policy Firewall.
- **Generalist Agent Routing:** Improved task delegation logic, aligning with the industry shift toward specialized subagents.

### 3. MCP Specification (2025-11-25) & C# SDK 1.0
Microsoft released the official MCP C# SDK 1.0, which fully implements the latest spec changes.
- **Protected Resource Metadata (PRM):** A standardized way for servers to expose authorization requirements via a `.well-known` URL or WWW-Authentication headers.
- **Icon Metadata:** Tools, resources, and prompts can now include icon metadata and website URLs, improving the "Marketplace" and "Dashboard" UX for MCP clients.

### 4. OpenClaw v2026.2.23
- **Stability Fixes:** Addressed critical startup hangs and "cron chaos" in agent workflows.
- **Refinement:** Moving toward production-ready status with better handled action cascades.

## Strategic Implications for MCP Any
- **A2A Residency:** MCP Any must evolve its A2A Bridge to act as a "Resident" for teammate-to-teammate communication, mirroring the `TeammateTool` functionality.
- **PRM Implementation:** High priority for the server to support PRM discovery to maintain compatibility with the 2025-11-25 spec.
- **Wildcard Scoping:** The Policy Engine needs to support the wildcard patterns now appearing in the Gemini ecosystem.
