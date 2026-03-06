# Market Sync: 2026-03-06

## Ecosystem Updates

### OpenClaw & "ClawJacked" Vulnerability
- **Finding**: A critical vulnerability (CVE-2026-2256) was disclosed in OpenClaw, allowing malicious websites to hijack local agents by exploiting improper origin validation and unsanitized input in the `Shell` tool.
- **Impact**: Highlights the failure of "Localhost-is-Safe" assumptions. Agents must transition to a Zero-Trust model even for local inter-process communication.
- **Trend**: Shift towards **Origin-Aware Tool Execution** and **Cryptographic Attestation** for all tool calls.

### Claude Code: Discovery & Control
- **Finding**: Claude Code has implemented "Tool Search" for on-demand discovery and a `claude agents` command for lifecycle management. It also introduced a `SIMPLE` mode to strip all agentic capabilities for high-security environments.
- **Trend**: Moving away from static tool lists to **Intent-Based Discovery**.

### Gemini CLI: Configuration Persistence
- **Finding**: Gemini CLI now supports persistent MCP server configuration via `settings.json`, moving away from transient CLI-only setup.
- **Trend**: Standardizing configuration formats for agent toolkits.

### The "8,000 Exposed Servers" Crisis
- **Finding**: Recent scans found over 8,000 MCP servers exposed to the public internet, many with no authentication, vulnerable to prompt injection and data exfiltration.
- **Trend**: Urgent need for **Safe-by-Default Infrastructure** where remote access is disabled and requires explicit, MFA-backed enabling.

## Strategic Gaps Identified
1. **Agent Identity**: Most systems still use shared API keys. We need **Independent Agent Identities** with unique, short-lived certificates for tool access.
2. **Origin Validation**: MCP servers must verify the *caller's origin* (e.g., specific binary or signed request) to prevent "Browser-to-Agent" hijacking.
3. **Intent-Scoped Discovery**: Instead of agents seeing all tools, they should only see tools that match the *user's verified intent*.

## User Pain Points
- **Security Anxiety**: Developers are afraid to run local agents that have shell access after the OpenClaw exploit.
- **Tool Bloat**: Too many MCP servers making the LLM "confused" and wasting tokens.
- **Shadow AI**: IT departments lack visibility into what autonomous agents are doing in the background.
