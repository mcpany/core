# Market Sync: 2026-03-05

## Ecosystem Updates

### OpenClaw: Dynamic Subagent Capability Negotiation
OpenClaw has moved beyond static capability tokens to a dynamic negotiation protocol. Subagents can now "request" specific permissions from their parent agent during execution. This reduces the initial context window overhead by not granting all permissions upfront.
*   **Impact on MCP Any**: We need to implement a "Capability Request" hook in the Policy Firewall.

### Claude Code: Resource Contention & Fragmentation
Multiple instances of Claude Code running on the same host are experiencing "MCP Lock-in" where one instance monopolizes a stdio-based MCP server.
*   **Impact on MCP Any**: MCP Any's role as a multiplexer is more critical than ever. We should explore "Fair-Share Resource Allocation" for MCP connections.

### Gemini CLI: Vertex AI Extension Bridging
Gemini CLI now supports direct bridging to Vertex AI Extensions. This introduces a new set of OAuth-heavy auth patterns that differ from standard MCP.
*   **Impact on MCP Any**: Need to expand the `auth` middleware to support delegated Vertex/GCP credentials more natively.

### Agent Swarms: The Rise of Ephemeral Tools
Swarms are increasingly generating "one-off" tools (e.g., a specific data parser) that only need to live for a single task.
*   **Impact on MCP Any**: Current "Config-Reload" approach for new tools is too slow. MCP Any needs an "Ephemeral Tool Registry" (In-Memory) for rapid tool lifecycle management.

## Autonomous Agent Pain Points
- **Context Fragmentation**: State loss when jumping between different agent frameworks (OpenClaw -> AutoGen).
- **Auth Exhaustion**: Manually configuring API keys for 20+ MCP servers.
- **Safety Latency**: The overhead of Rego/CEL policy checks is becoming a bottleneck for real-time swarms.

## Security Vulnerabilities
- **"Ghost Tooling"**: Discovery of a pattern where rogue subagents can inject "Ghost Tools" into a shared session by mimicking A2A protocol headers.
- **Action**: Strengthening the A2A Bridge with mandatory "Parent Attestation" for all new tool registrations.
