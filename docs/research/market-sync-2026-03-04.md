# Market Sync: 2026-03-04

## Ecosystem Updates

### OpenClaw (v2026.2.23 - v2026.2.26)
- **Security Hardening**: Significant focus on redacting sensitive keys (`env.*`) in config snapshots and blocking obfuscated commands.
- **"ClawJacked" Vulnerability**: A critical 0-click flaw was discovered where malicious websites could hijack local AI agents via WebSockets to localhost. OpenClaw patched this by implementing faster rate limits and better auth, but the industry-wide risk for local-bound MCP servers remains.
- **Browser Control & Sessions**: Improved reliability for browser-based automation and "sessions cleanup" for better disk-budget control.

### Claude Code & Agent Teams
- **Agent Teams**: Claude has introduced a "Teams" feature, often used with `tmux`, allowing multiple agents to collaborate on complex tasks with distinct panels for observation.
- **Multi-Agent Coordination**: The trend is moving from sub-agents to peer-to-peer "Team" structures where agents can request information from each other independently.

### Gemini & Local MCP
- Gemini CLI continues to push for local Kotlin-based MCP development, emphasizing standard `stdio` transport.

## Security & Vulnerability Trends
- **CVE-2026-27825 (mcp-atlassian)**: Critical RCE and SSRF due to missing directory confinement in attachment tools. Highlights the danger of unauthenticated MCP HTTP transports bound to `0.0.0.0`.
- **Command Injection**: Found in "HexStrike AI MCP Server" via unsanitized API arguments.
- **Zero-Trust Necessity**: The "ClawJacked" incident proves that `localhost` is not a security boundary for browser-based attacks.

## Autonomous Agent Pain Points
- **Cross-Origin Hijacking**: Agents running on developer machines are vulnerable to side-channel attacks from the browser.
- **Context Pollution**: Multi-agent teams need better ways to isolate state so specialized agents don't get overwhelmed by irrelevant parent context.
- **Permission Friction**: Users want "Set and Forget" security that doesn't require constant clicking but still prevents exfiltration.
