# Market Sync: 2026-02-27

## Ecosystem Updates

### Claude Code: Selective Context Inheritance
- **Insight**: Anthropic released a major update to Claude Code introducing "Selective Context Inheritance." This allows parent agents to specify exactly which environment variables, file handles, and session metadata are inherited by child agents/subagents.
- **Impact**: Reduces context bloat and improves security by following the principle of least privilege.
- **MCP Any Opportunity**: Implement a "Selective Inheritance Guard" middleware that enforces these boundaries for all tool-triggered subagents.

### Gemini CLI: Native MCP-over-WebSockets
- **Insight**: Google's Gemini CLI now natively supports MCP-over-WebSockets, optimizing for long-lived connections between cloud-based models and local tool servers.
- **Impact**: Shift away from pure Stdio for local-to-cloud scenarios where persistence is key.
- **MCP Any Opportunity**: Accelerate the "Unified Transport Layer" to prioritize high-performance WebSocket bridging.

### OpenClaw: Swarm-Sync Protocol
- **Insight**: OpenClaw announced "Swarm-Sync," a protocol for real-time state synchronization between agents using a shared vector-based blackboard.
- **Impact**: Multi-agent coordination is moving beyond simple message passing to shared semantic memory.
- **MCP Any Opportunity**: Evolve the "Shared KV Store" into a "Shared Vector Blackboard" that supports semantic search across agent sessions.

## Autonomous Agent Pain Points
- **Federation Latency**: As tool sets grow and become federated, agents are struggling with "Latency Hallucination"—assuming a tool will respond quickly when it's actually on a slow remote node.
- **Context Fragmentation**: Subagents losing the "higher intent" of the parent agent due to aggressive context pruning.

## Security Vulnerabilities
- **A2A Identity Smuggling**: A new exploit pattern where a compromised subagent in one framework (e.g., AutoGen) uses a shared session token to impersonate a high-privilege agent in another framework (e.g., CrewAI) via the A2A bridge.
- **Websocket Hijacking**: Unencrypted MCP-over-WS connections in local environments being intercepted by malicious local processes.
