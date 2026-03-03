# Market Sync: 2026-03-03

## Ecosystem Shifts & Findings

### 1. OpenClaw "Zero-Click" WebSocket Vulnerability
**Source:** Cyber Press, Penligent.
- **Incident**: A critical 0-click flaw was discovered in OpenClaw (formerly Clawdbot/MoltBot) where malicious websites could silently hijack local developer AI agents.
- **Mechanism**: The vulnerability resides in the core gateway—a local WebSocket server. Malicious sites can initiate connections to `localhost` ports used by the agent, bypassing standard web security boundaries because the local server lacked strict Origin validation and authentication for initial handshakes.
- **Pain Point**: This highlights the extreme risk of "Local-by-Default" gateways that don't implement robust Cross-Origin Resource Sharing (CORS) or WebSocket Origin enforcement.

### 2. Gemini CLI & SDK Evolutionary Updates
**Source:** Gemini CLI Changelog (v0.30.0, v0.31.0).
- **SessionContext**: Introduced `SessionContext` for SDK tool calls, formalizing how state is passed through tool execution chains.
- **Policy Engine Maturity**: Gemini CLI moved from simple `--allowed-tools` flags to a sophisticated Policy Engine supporting project-level policies and tool annotation matching.
- **Browser Agent Integration**: Introduction of an experimental browser agent signals a move towards agents that can directly interact with the web, increasing the importance of secure tool execution boundaries.

### 3. Agent Swarm Pain Points
- **Context Pollution**: As agents use more tools (100+), the "context window tax" is becoming a primary bottleneck.
- **Security vs. Autonomy**: Developers are struggling to balance "autonomous action" with "security seatbelts," leading to a demand for "Intent-Aware" permissions rather than static capability-based ones.

## Summary for MCP Any
Today's findings emphasize that MCP Any must not only be "Safe-by-Default" regarding IP bindings but must also implement **Strict Origin Enforcement** for WebSockets and **Intent-Aware Policy** filtering to prevent the types of exploits seen in the OpenClaw incident.
