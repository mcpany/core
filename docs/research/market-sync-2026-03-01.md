# Market Sync: 2026-03-01

## Ecosystem Updates

### OpenClaw (v2026.2.19)
- **Runtime Containment**: Improved isolation for subagents to prevent system-wide failures.
- **Context Expansion**: Support for 1M token context windows, shifting the bottleneck from "memory" to "relevance" (discovery).
- **Multi-Agent Coordination**: Tighter integration for specialized agent teams.

### Gemini CLI (v0.31.0)
- **Project-Level Policies**: Introduced hierarchical policy management.
- **MCP Wildcards**: Support for wildcard matching in tool permissions.
- **SessionContext**: Enhanced SDK support for maintaining state across tool calls.

### Claude Code Security Incident
- **CVE-2025-59536 / CVE-2026-21852**: RCE vulnerabilities discovered in Claude Code's project hooks and MCP integrations.
- **Attack Pattern**: Malicious `mcp.json` or project hooks in cloned repositories can execute arbitrary code or exfiltrate API keys.
- **Lesson**: Automatic execution of "project-local" tools or hooks without explicit user attestation is a critical failure point.

### MCP Protocol Vulnerabilities
- **CVE-2026-27735**: Path traversal in `mcp-server-git` allowing access outside repository boundaries.
- **OWASP MCP 10**: Emerging standard for securing AI agent tool interfaces.

## Autonomous Agent Pain Points
- **Context Pollution**: Large toolsets (100+) confuse models; need for "Lazy-Discovery."
- **Governance vs. Speed**: Teams are struggling to balance autonomous execution with safety (HITL).
- **Inter-Agent State Loss**: Swarms lose context when handing off tasks between specialized nodes.

## Unique Findings for Today
- The "Malicious Hook" pattern in Claude Code suggests that MCP Any must not only secure the *tools* but also the *configuration hooks* that orchestrate them.
- "Intent-Aware" permissions are becoming the gold standard to prevent subagents from being tricked into malicious actions.
