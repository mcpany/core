# Market Sync: 2026-03-09

## Ecosystem Updates

### OpenClaw Security Overhaul (v2026.2.26)
*   **Vulnerability**: A critical flaw allowed malicious websites to hijack OpenClaw agents by exploiting the local gateway's failure to distinguish between trusted local requests and cross-origin requests from a browser.
*   **Fix**: Introduced mandatory origin validation and transition to WebSocket-first transport.
*   **Context Isolation**: Added "Threadbound agents" to prevent cross-conversation context leakage.

### Claude Code: MCP Tool Search GA
*   **On-Demand Discovery**: Anthropic has moved "MCP Tool Search" to General Availability.
*   **Efficiency**: Claude now dynamically searches for tools when descriptions exceed a certain context threshold, significantly reducing token bloat for agents with large toolsets.

### Gemini CLI (v0.32.0)
*   **Policy Engine Enhancements**: Support for project-level policies, MCP server wildcards, and tool annotation matching.
*   **Generalist Agent**: Introduced a new routing layer to improve task delegation between specialized subagents.

### A2A (Agent-to-Agent) Protocol
*   **Momentum**: Growing industry support for the A2A protocol as the standard for vendor-agnostic agent collaboration. Gartner predicts 40% of enterprise apps will feature task-specific agents by 2026.

## Strategic Implications for MCP Any
1.  **Origin-Aware Security**: MCP Any must implement strict `Origin` and `Host` header validation for its local listeners to prevent "OpenClaw-style" hijacking.
2.  **Universal Search Tool**: As "Tool Search" becomes the standard (Claude), MCP Any should provide a unified `tools/search` interface that works for all clients, even those without native search support.
3.  **A2A Residency**: MCP Any should accelerate its role as a stateful buffer for A2A messages to support the growing multi-agent ecosystem.
