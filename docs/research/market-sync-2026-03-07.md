# Market Sync: 2026-03-07

## Ecosystem Shifts & Competitor Updates

### OpenClaw
- **Security Incident**: A critical vulnerability (patched 2026-03-02) was discovered that allowed malicious websites to hijack local agents via unauthenticated local endpoints. The root cause was a lack of origin verification for incoming connections.
- **v2026.3.2 Update**: "ACP subagents" (Agent Control Protocol) are now enabled by default. This marks a shift towards standardized subagent delegation and shared task context within the OpenClaw ecosystem.

### Gemini CLI (v0.32.0)
- **Generalist Agent**: Introduced improved task routing and delegation logic.
- **Policy Engine Hardening**: Now supports project-level policies and MCP server wildcards, allowing for more flexible but secure tool access control.

### Claude Code
- **MCP Tool Search GA**: Dynamic tool discovery is now the standard for handling large MCP servers (50+ tools).
- **Context Pollution Mitigation**: Automatic deferral of tool descriptions when they exceed 10% of the context window.

### Agent Swarms (Strands/Watsonx)
- **Shared Working Memory**: Emerging patterns for "Swarm" orchestration where multiple specialized agents operate on a shared state (Blackboard pattern) without a central supervisor.

## Autonomous Agent Pain Points
- **Origin Hijacking**: As agents run locally with powerful tools, the boundary between the "browser" (untrusted) and "local environment" (trusted) is becoming the primary attack vector.
- **Context Management in Swarms**: Efficiently sharing state between 5+ agents without hitting token limits or losing "intent" during handoffs.

## Security Vulnerabilities
- **Cross-Origin Tool Execution**: The OpenClaw exploit proves that "Local-Only" binding is insufficient if the server doesn't validate the `Origin` or `Referer` headers of incoming JSON-RPC requests.

## Unique Findings
- The industry is rapidly converging on a "Delegation-First" architecture. MCP Any must not just bridge tools, but provide the secure "Switchboard" for these delegations.
