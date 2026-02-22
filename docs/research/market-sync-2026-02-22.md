# Market Sync Research: 2026-02-22

## Ecosystem Shifts

### OpenClaw
- **Update:** Introduced "Swarm Context Propagation" (v2.4.0).
- **Finding:** Enables subagents to inherit parent context but lacks a standardized way to enforce "Recursive Scopes" (e.g., if a parent has `fs:read:/tmp`, a child should not be able to escalate to `fs:read:/`).
- **Pain Point:** Managing trust levels across recursive agent calls.

### Gemini CLI
- **Update:** MCP is now the primary tool interface.
- **Finding:** "Discovery Fatigue" is rising. Users with 10+ MCP servers struggle with tool shadowing (multiple tools with the same name) and high latency during tool discovery.
- **Need:** A centralized MCP Gateway that performs intelligent tool pruning and deduplication.

### Claude Code
- **Update:** "Local Execution Guardrails" introduced.
- **Finding:** While basic bash commands are filtered, complex "Agentic Loops" can still trick the system into executing unauthorized scripts by chaining multiple benign-looking tool calls.
- **Vulnerability:** Prompt injection via tool outputs.

### Agent Swarms (General)
- **Trend:** Move towards "Zero Trust Inter-agent Communication".
- **Finding:** Standard local HTTP tunneling for inter-agent comms is being deprecated in favor of isolated Docker-bound named pipes or encrypted unix sockets to prevent host-level side-channel attacks by rogue subagents.

## Autonomous Agent Pain Points
1. **Shared State (Blackboard Pattern):** Agents in a swarm lack a "common memory" that is protocol-native. Currently using hacky filesystem writes.
2. **Context Bloat:** Passing full tool schemas to every agent in a swarm consumes excessive tokens.
3. **Security:** Lack of granular, capability-based access control for tools at the "Universal Adapter" level.

## Security Vulnerabilities
- **Subagent Port Exposure:** Local MCP servers exposing HTTP ports are vulnerable to SSRF (Server-Side Request Forgery) from other apps on the same machine.
- **Tool Shadowing:** Malicious MCP servers registered late in the stack can "shadow" legitimate tools (e.g., a fake `git_commit` tool).
