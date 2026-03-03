# Market Sync: 2026-03-02

## Ecosystem Shifts & Competitor Analysis

### OpenClaw (formerly Clawdbot/Moltbot)
- **Rapid Adoption**: Reached "meteoric" adoption levels as a self-hosted AI agent.
- **Vulnerability Alert**: A critical vulnerability was disclosed (March 2, 2026) allowing malicious websites to hijack the agent because it failed to distinguish between trusted local apps and untrusted browser-originated requests. This underscores the need for "Origin-Aware" security in MCP Any.
- **Sonnet 4.6 Integration**: Native support for Claude Sonnet 4.6, optimized for computer-use and tighter instruction following.
- **Context Scaling**: Now supports a 1M token context window, increasing the pressure on MCP gateways to handle massive tool schemas efficiently.

### Gemini CLI & Claude Code
- **Gemini CLI**: Refining tool discovery via MCP. Strict name sanitization (alphanumeric/underscore/dot/hyphen only) and automatic server-alias prefixing are becoming standard patterns.
- **Claude Code**: Shifting towards "doer" agents that manage subagents. Standardizing subagent coordination patterns is a key pain point.

### Security & Vulnerabilities
- **CVE-2026-23744 (MCPJam inspector)**: RCE due to default binding to `0.0.0.0`. This validates MCP Any's "Local-Only by Default" strategic pivot.
- **CVE-2026-1977 (mcp-vegalite-server)**: RCE via code injection in visualization parameters. Highlights the need for "Deep Input Inspection" in the Policy Firewall.
- **CVE-2026-27735 (mcp-server-git)**: Path traversal in `git_add`. Demonstrates that even standard tools need "Path-Bound Scoping" middleware.

## Autonomous Agent Pain Points
- **Origin Hijacking**: Agents listening on local ports are vulnerable to "Cross-Protocol Hijacking" from the browser.
- **Subagent State Loss**: Difficulty in passing state securely between parent and child agents without leaking full credentials.
- **Tool Sprawl**: With 1M token windows, agents still struggle with "Selection Hallucination" when presented with hundreds of tools.

## Unique Findings
- **Agent Communication Protocol (ACP)**: Emerging as a competitor/complement to MCP for A2A communication. MCP Any should consider a "Dual-Stack" gateway that bridges MCP tools and ACP messages.
