# Market Sync: 2026-03-03

## Ecosystem Updates

### Gemini CLI & SDK (v0.31.0)
- **Gemini 3.1 Pro Support**: New model support with improved reasoning.
- **Policy Engine Enhancements**: Introduction of tool annotation matching and project-level policies. This shifts tool governance from simple "allow/deny" lists to metadata-driven intent matching.
- **SessionContext**: The new SDK formally introduces `SessionContext` for tool calls, reinforcing the need for MCP Any to handle session-bound state.

### Claude Code
- **MCP Tool Search (Dynamic Loading)**: Claude now implements a "defer_loading" strategy when tool schemas exceed 10% of the context window. It uses a specialized search tool to discover relevant MCP tools on-demand.
- **Standardized Environment Variables**: Heavy reliance on `ENABLE_TOOL_SEARCH` for governing context bloat.

### OpenClaw
- **ClawHub Registry**: A minimal, automatic skill registry that allows agents to pull in new capabilities without manual configuration.
- **A2A Coordination (`sessions_*` tools)**: OpenClaw has formalized agent-to-agent messaging and history fetching, effectively turning other agent sessions into "tools."

## Autonomous Agent Pain Points & Vulnerabilities
- **Machine-Speed Decisions**: Security architectures are struggling to keep up with agents that outnumber humans 82:1. The "Year of the Defender" (2026) focus is on visibility into machine-speed decisions.
- **Shadow AI Discovery**: Enterprises are struggling to detect unauthorized local AI agents (OpenClaw, etc.) running on employee machines, creating a demand for "Agent Governance" layers.

## Unique Findings for MCP Any
- **The "Context Threshold" Pattern**: There is a clear market convergence on "Lazy Loading" tools. MCP Any should implement a universal version of this that works across all models, not just Claude.
- **Annotation-Driven Security**: Aligning with the Gemini update, MCP Any's Policy Firewall should support filtering based on tool annotations/metadata (e.g., `danger_level: high`).
