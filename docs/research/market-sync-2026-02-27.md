# Market Sync: 2026-02-27

## Ecosystem Updates

### OpenClaw: Critical WebSocket Hijacking Vulnerability
- **Finding**: Oasis Security discovered a critical vulnerability chain in OpenClaw (reported 2026-02-26). Malicious websites can open a WebSocket connection to `localhost` on the OpenClaw gateway port. Since browsers do not block WebSocket connections to localhost, an attacker can silently take full control of a developer's AI agent.
- **Impact**: Zero-interaction hijacking of local agents. This highlights a massive "localhost security" gap in the agentic ecosystem.
- **Action for MCP Any**: Must implement strict WebSocket Origin validation and mandatory authentication for all local connections, even on `localhost`.

### Gemini CLI: v0.30.0 Release
- **Finding**: Google released Gemini CLI v0.30.0 (2026-02-25) featuring a new Policy Engine, "Strict Seatbelt" profiles, and deprecation of `--allowed-tools` in favor of a declarative policy system.
- **Trend**: Shift from command-line flags to robust, intent-based policy files.
- **Action for MCP Any**: Align our "Policy Firewall" with Gemini's declarative approach to ensure interoperability.

### Claude Code: Agent Teams
- **Finding**: Anthropic introduced "Agent Teams" for Claude Code. This allows parallel execution where a lead agent coordinates multiple teammate agents, each with its own context window.
- **Trend**: Multi-agent orchestration is moving from sequential delegation to parallel swarm execution.
- **Action for MCP Any**: Expand the `A2A Interop Bridge` to support parallel dispatching and multi-agent state merging.

### General Agentic Trends
- **"The Year of the Defender"**: Industry experts (Palo Alto Networks, Proofpoint) are labeling 2026 as the year where AI-driven defenses must counter "agent hijacking" and "indirect prompt injection" (IPI).
- **Agentic Web Attacks**: Agents consuming untrusted web content are the new primary attack vector.

## Unique Findings Summary
Today's unique finding is the **Localhost WebSocket Hijacking** pattern. While most security focus has been on remote attacks, the local execution boundary is extremely porous due to browser WebSocket behavior. MCP Any has a strategic opportunity to provide the "Secure Local Gateway" that standardizes protection against these local-host attacks.
