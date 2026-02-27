# Market Sync: 2026-02-27

## Ecosystem Updates

### OpenClaw: Security-First Infrastructure
- **Insight**: OpenClaw (v2.x) emphasizes that agent skills are executable code and must be treated as server-side infrastructure. It introduced a deep security audit command (`openclaw security audit --deep`) to identify exposed gateway processes and unauthorized tool access.
- **Impact**: MCP Any must integrate automated security auditing for all connected MCP servers to match this "Infrastructure-as-Agent" safety standard.

### Claude Code: GA of Tool Search & Context Compaction
- **Insight**: Anthropic has moved Tool Search, Web Fetch, and Memory tools to General Availability. Notably, they implemented "Context Compaction" triggered at 50k tokens to maintain reasoning performance in long-running sessions.
- **Impact**: MCP Any's On-Demand Discovery must now benchmark against GA-level search performance and consider context compaction as a first-class middleware.

### Gemini CLI: SessionContext & Seatbelt Policies
- **Insight**: Gemini CLI (v0.30.0) introduced `SessionContext` for SDK tool calls and a new `--policy` flag that supports "Strict Seatbelt Profiles." This allows users to enforce rigid safety boundaries per session.
- **Impact**: MCP Any needs to adopt a standardized `SessionContext` propagation model to ensure that policy enforcement is consistent across disparate agent frameworks.

## Autonomous Agent Pain Points
- **Silent Tool Exposure**: OpenClaw users are reporting accidental exposure of local files and secrets when enabling community "Skills" without proper auditing.
- **Context Bloat vs. Reasoning Quality**: Despite Claude's compaction, agents still struggle with "distraction" when too many irrelevant tool schemas are present.
- **Policy Inconsistency**: Users find it difficult to maintain the same "Seatbelt" safety profile when switching between Gemini CLI and other MCP-native agents.

## Security Vulnerabilities
- **Skill Injection**: Malicious OpenClaw skills can act as trojans, opening reverse shells via the agent gateway.
- **Session Leakage**: Without proper `SessionContext` isolation, sensitive data from one agent task can leak into the context of a subsequent, unrelated subagent task.
