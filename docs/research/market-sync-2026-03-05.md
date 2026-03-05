# Market Sync: 2026-03-05

## Ecosystem Shifts & Competitor Analysis

### 1. OpenClaw Security Crisis (WebSocket Hijacking)
- **Insight**: A critical vulnerability was disclosed (and patched in v2026.2.25) that allowed malicious websites to hijack local OpenClaw agents. By opening a WebSocket connection to `localhost` from a browser tab, attackers could execute arbitrary commands or exfiltrate data.
- **Impact for MCP Any**: This confirms that "Safe-by-Default" must include strict **Origin Verification** for all local listeners. We cannot assume `localhost` is a trusted perimeter if the user is running a web browser.
- **Action**: Elevate "Origin-Aware Middleware" to a P0 priority.

### 2. Claude Code: MCP Tool Search GA
- **Insight**: Anthropic has officially moved "MCP Tool Search" out of beta. It now implements a "Deferred Search" pattern: if tool schemas exceed 10% of the context window, they are automatically replaced by a search tool.
- **Impact for MCP Any**: This validates our **Lazy-MCP** design. The "10% Threshold" is a useful benchmark for our own implementation.
- **Action**: Align `design-on-demand-discovery.md` with the deferred search pattern.

### 3. Emergence of Agentic Identity Governance
- **Insight**: Industry reports (Dark Reading, SecurityWeek) are highlighting the "Governance Gap" for non-human identities. Agents are increasingly treated as first-class employees with elevated permissions.
- **Impact for MCP Any**: MCP Any should not just be a tool adapter but an **Identity Gateway** that manages the credentials and "Agentic Persona" of the subagents it orchestrates.
- **Action**: Propose "Agentic Identity Governance" as a new strategic pillar.

## Autonomous Agent Pain Points
- **Context Pollution**: Still the #1 complaint for developers with large toolsets (50+ tools).
- **Silent Hacking**: Fear of agents performing "side-effects" (deleting files/emails) without explicit consent, as seen in recent OpenClaw community reports.
- **Shadow AI**: IT departments are struggling to track which agents are running locally on developer machines.

## Security Vulnerabilities
- **Cross-Origin Hijacking**: The "Local-Only" assumption is broken by browser-based attacks.
- **Token Leakage**: Agents inadvertently sharing environment variables or API keys during multi-agent handoffs.
