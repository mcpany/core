# Market Sync: 2026-03-01

## 1. Ecosystem Updates

### Anthropic: Claude Code Security & Human-in-the-Loop (HITL)
*   **Claude Opus 4.6 Launch**: Demonstrated finding 500+ zero-day vulnerabilities in production OSS.
*   **Architecture Shift**: Anthropic is signaling that "human-approval architecture" is the standard for consequential AI agent execution. They are moving away from purely autonomous execution towards a governed model where high-risk actions require explicit human verification.
*   **MCP Tool Search**: Claude Code now uses dynamic tool discovery (MCP Tool Search) to prevent context pollution, loading only what is needed.

### Google: Gemini CLI & SDK (v0.31.0)
*   **Policy Engine Maturity**: Gemini CLI v0.31.0 introduced project-level policies, MCP server wildcards, and tool annotation matching. This aligns with the push for granular, capability-based security.
*   **SessionContext**: The new SDK introduces `SessionContext` for tool calls, reinforcing the need for stateful, session-aware agent interactions.
*   **Experimental Agents**: Introduction of an experimental browser agent, expanding the scope of local execution.

### OpenClaw: The "ClawHavoc" Crisis
*   **Supply Chain Poisoning**: OpenClaw (formerly Clawdbot) is facing a massive security crisis. The "ClawHavoc" campaign poisoned the skill marketplace with 1,184+ malicious packages (e.g., `solana-wallet-tracker`) that install AMOS malware to exfiltrate credentials.
*   **Public Exposure**: Over 42,000 exposed instances found on the public internet, highlighting the dangers of "Ease of Use" without "Safe-by-Default" configurations.

## 2. Autonomous Agent Pain Points
*   **Marketplace Trust**: The "ClawHavoc" incident has shattered trust in unverified community tool registries.
*   **Context Pollution**: As MCP servers grow (50+ tools), upfront loading is becoming unsustainable.
*   **A2A Coordination**: High-latency and state loss in multi-agent swarms remain a bottleneck for complex workflows.

## 3. Security Vulnerabilities
*   **Command Injection**: CVE-2026-0755 identified in `gemini-mcp-tool` due to improper validation in `execAsync`.
*   **Local Port Exposure**: Rogue subagents in OpenClaw have been observed exploiting local HTTP tunnels to gain unauthorized host-level file access.
