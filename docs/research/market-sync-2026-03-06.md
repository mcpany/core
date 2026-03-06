# Market Sync: 2026-03-06

## Ecosystem Updates

### OpenClaw & Agent Swarms
- **OpenClaw (v2026.2.25+)**: Major security release to address a high-severity browser-origin hijack vulnerability. The framework is shifting towards stricter "Origin-Aware" connections.
- **A2A Momentum**: The Agent-to-Agent protocol is becoming the standard for delegation, with frameworks like CrewAI and AutoGen adopting it for cross-framework tasks.

### Claude Code & Gemini CLI
- **Claude Code (v2.1.30)**: Introduced sub-agent access to SDK-provided MCP tools, reducing friction for complex, nested agent tasks. Performance improvements in session loading (progressive enrichment).
- **Gemini CLI (v0.32.0)**: Enhanced the "Generalist Agent" for better task routing. The Policy Engine now supports project-level policies and tool annotation matching, aligning with MCP Any's Zero-Trust goals.

## Security & Vulnerabilities

### Browser-Origin Hijack (OpenClaw)
- A critical vulnerability was found in OpenClaw where malicious websites could hijack local agents by forging connections from the browser.
- **Mitigation**: Standardizing "Local App Attestation" and origin validation for all local tool interfaces is now a P0 priority.

### "Shadow Tool" Persistence
- Community registries continue to be a source of unverified MCP servers. The need for "Attested Tooling" (cryptographic signatures) is increasing as agents take on more sensitive tasks.

## Autonomous Agent Pain Points
- **Delegated Trust**: Agents lack a standard way to verify if a sub-agent is authorized to use a parent agent's specific tool capabilities.
- **State Fragmentation**: As agents hand off tasks (e.g., via A2A), maintaining a coherent state without excessive context bloat remains a challenge.
- **Startup Latency**: Massive tool libraries still cause slow agent initialization, reinforcing the need for "Lazy Loading" (Lazy-MCP).
