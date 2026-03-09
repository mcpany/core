# Market Sync: 2026-03-09

## Ecosystem Updates

### OpenClaw & Agent Swarms
- **Agent Identity (AID) v1.0**: OpenClaw has proposed the AID specification, using Decentralized Identifiers (DIDs) to verify agent identity across heterogeneous swarms. This allows for cryptographically secure delegation.
- **Swarm Orchestration**: New patterns in CrewAI for "Long-lived Swarms" that require persistent shared state across different cloud providers.

### Claude Code & Gemini CLI
- **Recursive Tool Discovery**: Claude Code (v3.5) now supports tools that can return "Ephemeral Tool Schemas." This allows an agent to discover specialized sub-tools only after an initial high-level tool call (e.g., `connect_to_db` returns `query_table_users`).
- **Gemini Pre-dispatch Hooks**: Gemini CLI introduced local hooks that allow the host environment to intercept tool calls for validation or argument modification before execution.

## Security & Vulnerabilities

### "Ghost Tooling" Exploit
- A new vulnerability where malicious MCP servers appear to have valid schemas but perform unauthorized environment variable exfiltration during the "Pre-flight" check phase, before any tool is actually called.
- **Mitigation**: Requires signed pre-flight attestation and stricter isolation of the discovery process.

## Autonomous Agent Pain Points
- **Cross-Framework Identity**: Difficulty in passing authorized identity from a Claude-based orchestrator to an OpenClaw-based worker.
- **State Fragmentation**: "Blackboard" patterns are becoming standard, but there's no universal API for shared state between different agent frameworks.
